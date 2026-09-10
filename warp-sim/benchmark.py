"""Run the Warp solver and compare every physical field with production Go."""
from __future__ import annotations

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import platform
import statistics
import subprocess
import time

import numpy as np
import warp as wp

from solver import Config, State, simulate

ROOT = Path(__file__).resolve().parents[1]
STATS = ["initial_energy", "injected", "dissipated", "initial_matter", "allocations",
         "copies", "deaths", "instructions", "absorbed", "transferred", "initial_chemical",
         "converted0", "converted1", "charged", "taken", "failed_space", "failed_matter",
         "failed_reserve", "failed_limit", "failed_absorb", "failed_reaction", "instruction_starved"]
PARTICLE = ["position", "energy", "length", "ip", "flag", "id", "parent", "target", "created", "generation"]
RESOURCES = ["X", "Y", "Z", "field_energy", "energy"]
ARRAYS = ["cells", "particles", "code", "memory", "bonds", "order", "free", "events",
          "control", "meta", "stats", "modules", "changes", "usage"]


class Batch:
    """Input must first pass Go's World.Validate (the CLI enforces this).

    This experimental backend tracks physics and global accounting. It does not
    build the Go observer's historical genome/origin ledgers or a Go snapshot.
    """
    def __init__(self, worlds: list[dict], device: str):
        if not worlds:
            raise ValueError("empty batch")
        if any(w["config"].get("collective_ablation") for w in worlds):
            raise ValueError("Warp does not support collective ablations; use Go")
        if any(w["config"].get("environment") for w in worlds):
            raise ValueError("Warp does not support environment engineering; use the Go simulator")
        if any(w["config"].get("copy_model") for w in worlds):
            raise ValueError("Warp does not support encoded copy policies; use the Go simulator")
        self.initial = copy.deepcopy(worlds)
        self.device = device
        self.count = len(worlds)
        config = worlds[0]["config"]
        signature = {k: v for k, v in config.items() if k != "seed"}
        for world in worlds:
            if {k: v for k, v in world["config"].items() if k != "seed"} != signature:
                raise ValueError("all worlds in one batch must share physical configuration")
            if world["next_id"] >= 2**62 or world["tick"] >= 2**62:
                raise ValueError("Warp prototype requires tick/ID below 2**62")
            for p in world["particles"].values():
                if any(not -(2**30) <= i[k] < 2**30 for i in p["code"] or [] for k in ("a", "b")):
                    raise ValueError("Warp instruction arguments must fit signed 31 bits")
        self.config = Config()
        for attr, key in {"width":"width", "height":"height", "entities":"max_entities",
                          "code_limit":"max_code", "cell_capacity":"cell_capacity",
                          "energy_capacity":"energy_capacity", "inflow":"inflow",
                          "maintenance":"maintenance", "mutation":"mutation_ppm",
                          "matter_diffusion":"matter_diffusion", "chemical_diffusion":"chemical_diffusion",
                          "ecology":"ecology"}.items():
            setattr(self.config, attr, int(config[key]))
        n, slots, length = self.count, config["max_entities"], config["max_code"]
        # Explicit allocation estimate prevents accidental device OOM at huge batches.
        estimate = n * (config["width"] * config["height"] * 6 * 8 + slots * (length * 3 * 8 + 300))
        if estimate > 6 * 1024**3:
            raise ValueError(f"batch needs about {estimate / 1024**3:.1f} GiB; use smaller batches")
        modules = [None]
        by_hash = {"builtin": 0}
        for world in worlds:
            rs = world.get("rule_state")
            if rs:
                candidates = [rs["active"], *(rs["history"] or []), *(c.get("module") for c in rs["pending"] or [])]
                for m in candidates:
                    if m and m["sha256"] not in by_hash:
                        by_hash[m["sha256"]] = len(modules)
                        modules.append(m)
        self.usage_keys = [m["sha256"] + "/" + r["name"] for m in modules[1:] for r in m["bytecode"]]
        table = np.zeros((len(modules)*16, 15, n), dtype=np.int64)
        for mi, m in enumerate(modules[1:], 1):
            for r in m["bytecode"]:
                rid = mi*16+r["id"]
                table[rid, :4, :] = np.array([1, r["energy_cost"], r["max_batch"], 2*len(r["code"])+5])[:, None]
                for inst in r["code"]:
                    if inst["op"] == 0:
                        table[rid, 4+inst["resource"], :] = inst["amount"]
                        table[rid, 9+inst["resource"], :] -= inst["amount"]
                    else:
                        table[rid, 9+inst["resource"], :] += inst["amount"]
                table[rid, 14, :] = self.usage_keys.index(m["sha256"]+"/"+r["name"])
        change_count = max(1, max(len((w.get("rule_state") or {}).get("pending") or []) for w in worlds))
        host = {
            "cells": np.zeros((6, config["width"]*config["height"], n), np.int64),
            "particles": np.zeros((10, slots, n), np.int64),
            "code": np.zeros((slots*length, 3, n), np.int64),
            "memory": np.zeros((16, slots, n), np.int64),
            "bonds": np.zeros((4, slots, n), np.int32),
            "order": np.zeros((slots, n), np.int32),
            "free": np.zeros((slots, n), np.int32),
            "events": np.zeros((5, slots, n), np.int64),
            "control": np.zeros((4, n), np.int32),
            "meta": np.zeros((4, n), np.uint64),
            "stats": np.zeros((23, n), np.int64),
            "modules": table,
            "changes": np.full((change_count, 2, n), np.iinfo(np.uint64).max, np.uint64),
            "usage": np.zeros((max(1, len(self.usage_keys)), n), np.int64),
        }
        for wi, world in enumerate(worlds):
            particles = sorted(world["particles"].values(), key=lambda p: p["id"])
            ids = {p["id"]: i for i, p in enumerate(particles)}
            for pos, cell in enumerate(world["cells"]):
                host["cells"][:, pos, wi] = [cell["energy"], cell["matter"], ids[cell["occupant"]]+1 if cell["occupant"] else 0, *cell["chemical"]]
            for slot, p in enumerate(particles):
                code = p["code"] or []
                host["particles"][:, slot, wi] = [p["position"], p["energy"], len(code), p["ip"], p["flag"], p["id"], p["parent"], p["target"], p["created"], p["generation"]]
                host["memory"][:, slot, wi] = p["memory"]+p["initial_memory"]
                for j, inst in enumerate(code):
                    host["code"][slot*length+j, :, wi] = [inst["op"], inst["a"], inst["b"]]
            for relation in world["relations"].values():
                a, b = ids[relation["a"]], ids[relation["b"]]
                for p, q in ((a, b), (b, a)):
                    empty = np.flatnonzero(host["bonds"][:, p, wi] == 0)[0]
                    host["bonds"][empty, p, wi] = q+1
            count = len(particles)
            host["order"][:count, wi] = np.arange(count)
            host["free"][:slots-count, wi] = np.arange(slots-1, count-1, -1)
            host["control"][:2, wi] = [count, slots-count]
            host["meta"][:, wi] = [world["tick"], world["rng"]["state"], world["transport_rng"]["state"], world["next_id"]]
            for i, key in enumerate(STATS):
                host["stats"][i, wi] = world["accounting"]["converted"][i-11] if i in (11, 12) else world["accounting"][key]
            rs = world.get("rule_state")
            if rs:
                active = rs["active"]
                history = list(rs["history"] or [])
                host["control"][2, wi] = by_hash[active["sha256"] if active else "builtin"]
                for ci, change in enumerate(rs["pending"] or []):
                    if change.get("rollback"):
                        active = history.pop()
                    else:
                        history.append(active)
                        active = change["module"]
                    host["changes"][ci, :, wi] = [change["tick"], by_hash[active["sha256"] if active else "builtin"]]
                host["stats"][22, wi] = rs["instructions"]
                for ui, key in enumerate(self.usage_keys):
                    host["usage"][ui, wi] = (rs["usage"] or {}).get(key, 0)
        self.state = State()
        for name, arr in host.items():
            dtype = wp.uint64 if arr.dtype == np.uint64 else wp.int32 if arr.dtype == np.int32 else wp.int64
            setattr(self.state, name, wp.array(arr, dtype=dtype, device=device))

    def run(self, ticks: int, chunk: int = 8, lanes: int = 32):
        if ticks < 0 or chunk < 1 or lanes not in (1, 32):
            raise ValueError("invalid tick count/chunk")
        remaining = ticks
        while remaining:
            n = min(chunk, remaining)
            wp.launch(simulate, dim=self.count*lanes, inputs=[self.state, self.config, n, lanes], device=self.device, block_dim=32)
            # Bound queued work on a Windows display GPU; includes sync overhead.
            wp.synchronize_device(self.device)
            remaining -= n

    def project(self) -> list[dict]:
        a = {name: getattr(self.state, name).numpy() for name in ARRAYS if name not in ("events", "free", "modules", "changes")}
        output = []
        for wi in range(self.count):
            p = a["particles"][:, :, wi]
            particles = []
            relations = set()
            for slot in a["order"][:a["control"][0, wi], wi]:
                row = {k: int(p[i, slot]) for i, k in enumerate(PARTICLE) if k != "length"}
                row["flag"] = bool(row["flag"])
                length = int(p[2, slot])
                row["code"] = [{"op": int(x[0]), "a": int(x[1]), "b": int(x[2])} for x in a["code"][slot*self.config.code_limit:slot*self.config.code_limit+length, :, wi]] or None
                row["memory"] = a["memory"][:8, slot, wi].tolist()
                row["initial_memory"] = a["memory"][8:, slot, wi].tolist()
                particles.append(row)
                for q in a["bonds"][:, slot, wi]:
                    if q:
                        relations.add(tuple(sorted((row["id"], int(p[5, q-1])))))
            cells = []
            for cell in a["cells"][:, :, wi].T:
                cells.append({"energy": int(cell[0]), "matter": int(cell[1]), "occupant": int(p[5, cell[2]-1]) if cell[2] else 0, "chemical": cell[3:].tolist()})
            stats = {key: int(a["stats"][i, wi]) for i, key in enumerate(STATS) if i not in (11, 12)}
            stats["converted"] = a["stats"][11:13, wi].tolist()
            row = {"tick": int(a["meta"][0, wi]), "rng": int(a["meta"][1, wi]), "transport_rng": int(a["meta"][2, wi]), "next_id": int(a["meta"][3, wi]), "cells": cells, "particles": particles, "relations": [{"a": x, "b": y} for x, y in sorted(relations)], "accounting": stats}
            rs = copy.deepcopy(self.initial[wi].get("rule_state"))
            if rs:
                pending = list(rs["pending"] or [])
                for change in pending[:int(a["control"][3, wi])]:
                    rs["history"] = rs["history"] or []
                    if change.get("rollback"):
                        rs["active"] = rs["history"].pop()
                    else:
                        rs["history"].append(rs["active"])
                        rs["active"] = change["module"]
                    rs["events"] = rs["events"] or []
                    rs["events"].append({"tick": change["tick"], "action": "rollback" if change.get("rollback") else "load", "sha256": rs["active"]["sha256"] if rs["active"] else "builtin"})
                if pending:
                    rs["pending"] = pending[int(a["control"][3, wi]):]
                rs["instructions"] = int(a["stats"][22, wi])
                for ui, key in enumerate(self.usage_keys):
                    value = int(a["usage"][ui, wi])
                    if value:
                        rs["usage"] = rs["usage"] or {}
                        rs["usage"][key] = value
                row["rule_state"] = rs
            validate_physical(row, self.initial[wi]["config"])
            output.append(row)
        return output


def validate_physical(row, config):
    cells, particles, stats = row["cells"], row["particles"], row["accounting"]
    energy = sum(c["energy"] + 8*c["chemical"][0] + 4*c["chemical"][1] for c in cells) + sum(p["energy"] for p in particles)
    matter = sum(c["matter"] for c in cells) + len(particles)
    chemical = sum(sum(c["chemical"]) for c in cells)
    if energy != stats["initial_energy"] + stats["injected"] - stats["dissipated"] or matter != stats["initial_matter"] or chemical != stats["initial_chemical"]:
        raise AssertionError("Warp resource conservation failed")
    ids = {p["id"] for p in particles}
    if len(ids) != len(particles) or len(particles) > config["max_entities"]:
        raise AssertionError("Warp particle ID/capacity invariant failed")
    for c in cells:
        if not 0 <= c["energy"] <= config["cell_capacity"] or c["matter"] < 0 or min(c["chemical"]) < 0 or c["occupant"] and c["occupant"] not in ids:
            raise AssertionError("Warp cell invariant failed")
    for p in particles:
        if not 0 < p["energy"] <= config["energy_capacity"] or cells[p["position"]]["occupant"] != p["id"] or p["id"] >= row["next_id"]:
            raise AssertionError("Warp particle invariant failed")
        if p["code"] and not 0 <= p["ip"] < len(p["code"]):
            raise AssertionError("Warp instruction pointer invariant failed")


def compare(actual, expected, path="worlds"):
    if type(actual) is not type(expected):
        raise AssertionError(f"{path}: types {type(actual).__name__} != {type(expected).__name__}")
    if isinstance(actual, dict):
        if actual.keys() != expected.keys():
            raise AssertionError(f"{path}: keys {actual.keys() ^ expected.keys()}")
        for key in actual:
            compare(actual[key], expected[key], f"{path}.{key}")
    elif isinstance(actual, list):
        if len(actual) != len(expected):
            raise AssertionError(f"{path}: lengths {len(actual)} != {len(expected)}")
        for i, (a, b) in enumerate(zip(actual, expected)):
            compare(a, b, f"{path}[{i}]")
    elif actual != expected:
        raise AssertionError(f"{path}: Warp {actual!r} != Go {expected!r}")


def physical_hash(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--worlds", default="1,16,256")
    parser.add_argument("--size", type=int, default=32)
    parser.add_argument("--ticks", type=int, default=1000)
    parser.add_argument("--warmup", type=int, default=300)
    parser.add_argument("--repeats", type=int, default=3)
    parser.add_argument("--workers", type=int, default=16)
    parser.add_argument("--scenario", choices=["baseline", "ecology", "mutation", "dsl", "assay", "odd"], default="ecology")
    parser.add_argument("--device", default="cuda:0")
    parser.add_argument("--chunk", type=int, default=8)
    parser.add_argument("--lanes", type=int, choices=(1, 32), default=32, help="32 isolates divergent worlds into separate CUDA warps")
    parser.add_argument("--output", required=True, type=Path, help="new directory")
    args = parser.parse_args()
    counts = [int(x) for x in args.worlds.split(",")]
    if min(counts) < 1 or len(set(counts)) != len(counts) or args.repeats < 1 or args.ticks < 1 or args.warmup < 0 or args.chunk < 1 or not 1 <= args.workers <= 256:
        parser.error("invalid benchmark limits")
    directory = args.output.resolve()
    directory.mkdir(parents=True, exist_ok=False)
    exe = directory / ("reference.exe" if os.name == "nt" else "reference")
    subprocess.run(["go", "build", "-o", str(exe), "./cmd/warp-reference"], cwd=ROOT, check=True)
    wp.init()
    device = wp.get_device(args.device)
    manifest = {"status": "running", "warp": wp.__version__, "python": platform.python_version(), "device": device.name,
                "scenario": args.scenario, "size": args.size, "ticks": args.ticks, "warmup": args.warmup,
                "repeats": args.repeats, "workers": args.workers, "chunk": args.chunk, "lanes": args.lanes,
                "parity_scope": "all physical state, both RNGs, global accounting and DSL state; excludes historical genome/origin ledgers",
                "go_binary_sha256": hashlib.sha256(exe.read_bytes()).hexdigest(),
                "sources_sha256": {name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest() for name in ("warp-sim/solver.py", "warp-sim/benchmark.py", "cmd/warp-reference/main.go")},
                "cpu_processor": platform.processor(), "rows": []}
    report = directory / "summary.json"
    def save():
        report.write_text(json.dumps(manifest, indent=2)+"\n", encoding="utf-8")
    save()
    try:
        for count in counts:
            fixture_path = directory / f"initial-{count}.json"
            subprocess.run([str(exe), "-worlds", str(count), "-size", str(args.size), "-scenario", args.scenario, "-warmup", str(args.warmup), "-ticks", "0", "-workers", str(args.workers), "-output", str(fixture_path)], cwd=ROOT, check=True)
            initial = json.loads(fixture_path.read_text(encoding="utf-8"))["initial"]
            manifest["physical_config"] = initial[0]["config"]
            t0 = time.perf_counter()
            warm = Batch(initial[:1], args.device)
            wp.launch(simulate, dim=args.lanes, inputs=[warm.state, warm.config, 0, args.lanes], device=args.device, block_dim=32)
            wp.synchronize_device(args.device)
            compile_seconds = time.perf_counter()-t0
            del warm
            cpu = {}
            expected = None
            for workers in sorted({1, args.workers}):
                timings = []
                for repeat in range(args.repeats):
                    path = directory / f"go-{count}-{workers}-{repeat}.json"
                    subprocess.run([str(exe), "-input", str(fixture_path), "-ticks", str(args.ticks), "-workers", str(workers), "-output", str(path)], cwd=ROOT, check=True)
                    ref = json.loads(path.read_text(encoding="utf-8"))
                    manifest["go_version"] = ref["go"]
                    timings.append(ref["seconds"])
                    if expected is None:
                        expected = ref["final"]
                    else:
                        compare(ref["final"], expected)
                cpu[str(workers)] = {"samples_seconds": timings, "median_seconds": statistics.median(timings)}
            gpu, end_to_end = [], []
            for repeat in range(args.repeats):
                t0 = time.perf_counter()
                batch = Batch(initial, args.device)
                wp.synchronize_device(args.device)
                t1 = time.perf_counter()
                batch.run(args.ticks, args.chunk, args.lanes)
                elapsed = time.perf_counter()-t1
                actual = batch.project()
                end_to_end.append(time.perf_counter()-t0)
                compare(actual, expected)
                gpu.append(elapsed)
                del batch
            row = {"worlds": count, "go": cpu, "warp_seconds": gpu, "warp_median_seconds": statistics.median(gpu),
                   "warp_end_to_end_seconds": end_to_end, "compile_and_warm_init_seconds": compile_seconds,
                   "speedup_vs_go_1": cpu["1"]["median_seconds"]/statistics.median(gpu),
                   "speedup_vs_go_workers": cpu[str(args.workers)]["median_seconds"]/statistics.median(gpu),
                   "parity": "exact", "physical_sha256": physical_hash(expected), "initial_sha256": physical_hash(initial)}
            manifest["rows"].append(row)
            save()
            print(json.dumps(row), flush=True)
        manifest["status"] = "complete"
        save()
    except BaseException as exc:
        manifest["status"] = "failed"
        manifest["error"] = str(exc)
        save()
        raise


if __name__ == "__main__":
    main()
