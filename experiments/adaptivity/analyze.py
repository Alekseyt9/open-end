"""Independently verify retention traces, scores, branch points and allocation."""
import argparse
import copy
import hashlib
import json
import math
from pathlib import Path


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def canonical_hash(value):
    return hashlib.sha256((json.dumps(value, separators=(",", ":"), ensure_ascii=False) + "\n").encode()).hexdigest()


def initial(source, challenge, mode):
    snap = copy.deepcopy(source)
    w = snap["world"]
    c = w["config"]
    if challenge == "dim":
        c["inflow"] //= 2
    elif challenge == "pulse":
        c["inflow"] = 0
    else:
        width, height = c["width"], c["height"]
        dx, dy = width // 2, 0
        if challenge == "mix-dim":
            dx, dy = width // 3, height // 2
            c["inflow"] = c["inflow"] * 3 // 4
        before = copy.deepcopy(w["cells"])
        for i, cell in enumerate(w["cells"]):
            j = ((i // width + dy) % height) * width + (i % width + dx) % width
            cell["energy"], cell["chemical"] = before[j]["energy"], before[j]["chemical"]
    if mode == "memory-reset":
        for p in w["particles"].values():
            p["memory"] = [0] * 8
    return snap


def score(probes, split, threshold):
    selected = {p["challenge"] for p in probes if p["split"] == split}
    robustness, perception, memory, productive = [], [], [], False
    for challenge in sorted(selected):
        arms = {p["mode"]: p for p in probes if p["challenge"] == challenge}
        a, b, c = (arms[k] for k in ("intact", "frozen-perception", "memory-reset"))
        value = a["mean_capped_population_retention"]
        robustness.append(value)
        perception.append(value - b["mean_capped_population_retention"])
        memory.append(value - c["mean_capped_population_retention"])
        productive |= a["copies"] > 0
    r, p, m = (sum(v) / len(v) for v in (robustness, perception, memory))
    info = (max(0, p) + max(0, m)) / 2
    return dict(robustness=r, perception_benefit=p, memory_benefit=m, information_benefit=info,
                adaptive_proxy_0_100=100 * r * info if productive and info >= threshold else 0, productive=productive)


def select(rows, slots, seed):
    result = {}
    for r in sorted(rows, key=lambda r: (-r["selection"]["adaptive_proxy_0_100"], r["source_sha256"])):
        if len(result) >= slots // 2 or r["selection"]["adaptive_proxy_0_100"] <= 0:
            break
        result[r["source_sha256"]] = "adaptive_proxy"
    while len(result) < slots - (1 if slots > 1 else 0):
        def distance(row):
            refs = [r for r in rows if r["source_sha256"] in result]
            if not refs:
                return row["functional_diversity"]
            return min(sum((a - b) ** 2 for a, b in zip(row["behavior_descriptor"], r["behavior_descriptor"])) for r in refs)
        winner = min((r for r in rows if r["source_sha256"] not in result), key=lambda r: (-distance(r), r["source_sha256"]))
        result[winner["source_sha256"]] = "behavioral_diversity"
    if len(result) < slots:
        winner = min((r for r in rows if r["source_sha256"] not in result),
                     key=lambda r: hashlib.sha256(f'{seed}/{r["source_sha256"]}'.encode()).hexdigest())
        result[winner["source_sha256"]] = "exploration"
    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, default=Path("data/adaptivity-round1-verified"))
    parser.add_argument("--symbol-assay", type=Path, default=Path("data/symbol-assay-stage15-verified"))
    parser.add_argument("--out", type=Path, default=Path("experiments/adaptivity/evidence.json"))
    args = parser.parse_args()
    manifest = read(args.input / "manifest.json")
    assert manifest["status"] == "complete"
    rows = read(args.input / "evaluation.json")
    assert len(rows) * 12 == manifest["probes"]
    threshold = manifest["config"]["minimum_mean_benefit"]
    differences = []
    for r in rows:
        path = Path(manifest["input"]) / r["source_snapshot"]
        assert digest(path) == r["source_sha256"]
        source = read(path)
        assert len(r["probes"]) == 12
        keys = {(p["challenge"], p["mode"]) for p in r["probes"]}
        assert len(keys) == 12
        n = sum(bool(p["code"]) for p in source["world"]["particles"].values())
        for p in r["probes"]:
            assert p["source_sha256"] == r["source_sha256"]
            assert p["initial_sha256"] == canonical_hash(initial(source, p["challenge"], p["mode"]))
            assert p["initial_executable"] == n and len(p["executable_trace"]) == p["ticks"]
            assert all(isinstance(x, int) and 0 <= x <= source["world"]["config"]["max_entities"] for x in p["executable_trace"])
            assert p["executable_trace"][-1] == p["final_executable"]
            retention = sum(min(1, x / n) for x in p["executable_trace"]) / p["ticks"]
            differences.append(abs(retention - p["mean_capped_population_retention"]))
        for split in ("selection", "validation"):
            expected = score(r["probes"], split, threshold)
            for key, value in expected.items():
                assert math.isclose(value, r[split][key], abs_tol=1e-12)
    expected = select(rows, manifest["selected_slots"], manifest["selection_seed"])
    for r in rows:
        assert r["selected"] == (r["source_sha256"] in expected)
        assert r.get("selection_reason") == expected.get(r["source_sha256"])
    controls = {r["seed"]: r["final_sha256"] for r in read(args.symbol_assay / "results.json") if r["variant"] == "persistent"}
    final = read(args.input / "summary.json")
    assert len(final) == manifest["completed"] == manifest["total"] == manifest["selected_slots"]
    assert {r["Seed"] for r in final} == {r["seed"] for r in rows if r["selected"]}
    for r in final:
        assert digest(args.input / r["Snapshot"]) == r["StateSHA256"] == controls[r["Seed"]]
        fs = [json.loads(line) for line in (args.input / r["Metrics"]).read_text().splitlines()]
        candidate = next(x for x in rows if x["seed"] == r["Seed"])
        source_world = read(Path(manifest["input"]) / candidate["source_snapshot"])["world"]
        # Telemetry hashes the world alone; snapshot hashes include its envelope.
        assert fs[0]["telemetry"]["session_initial_world_sha256"] == canonical_hash(source_world)
        assert fs[-1]["tick"] == r["Tick"]
    compact = copy.deepcopy(rows)
    for r in compact:
        for p in r["probes"]:
            trace = p.pop("executable_trace")
            p["trace_sha256"] = canonical_hash(trace)
    evidence = dict(version=1, manifest=manifest, evaluation_sha256=digest(args.input / "evaluation.json"),
                    max_retention_recalculation_difference=max(differences), normal_continuations_match_symbol_controls=True,
                    candidates=compact, continuations=final)
    assert max(differences) < 1e-12
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(dict(probes=manifest["probes"], selected=[r["Seed"] for r in final],
                          max_retention_difference=max(differences), independently_verified_selection=True)))


if __name__ == "__main__":
    main()
