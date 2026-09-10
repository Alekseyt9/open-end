"""Run validated Go snapshots or fixture batches through the Warp solver.

Output is a physical-state report, not a Go snapshot: historical observer
ledgers are deliberately not synthesized. Batch.run supports repeated advances.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time

import warp as wp
from benchmark import Batch, ROOT, physical_hash
from solver import simulate


def main():
    p = argparse.ArgumentParser(description=__doc__)
    source = p.add_mutually_exclusive_group(required=True)
    source.add_argument("--snapshots", nargs="+", type=Path)
    source.add_argument("--fixture", type=Path)
    p.add_argument("--ticks", type=int, required=True)
    p.add_argument("--output", type=Path, required=True)
    p.add_argument("--device", default="cuda:0")
    p.add_argument("--chunk", type=int, default=8)
    p.add_argument("--lanes", type=int, choices=(1, 32), default=32)
    args = p.parse_args()
    if args.ticks < 0 or args.chunk < 1:
        p.error("ticks must be nonnegative and chunk positive")
    if args.output.exists():
        p.error("output already exists")
    worlds, sources = [], []
    with tempfile.TemporaryDirectory(prefix="open-end-warp-") as tmp:
        exe = Path(tmp) / ("reference.exe" if os.name == "nt" else "reference")
        result = Path(tmp) / "validated.json"
        subprocess.run(["go", "build", "-o", str(exe), "./cmd/warp-reference"], cwd=ROOT, check=True)
        for path in args.snapshots or [args.fixture]:
            path = path.resolve()
            subprocess.run([str(exe), "-snapshot" if args.snapshots else "-input", str(path), "-ticks", "0", "-output", str(result)], cwd=ROOT, check=True)
            worlds.extend(json.loads(result.read_text(encoding="utf-8"))["initial"])
            sources.append({"path": str(path), "sha256": hashlib.sha256(path.read_bytes()).hexdigest()})
    wp.init()
    started = time.perf_counter()
    batch = Batch(worlds, args.device)
    wp.launch(simulate, dim=args.lanes, inputs=[batch.state, batch.config, 0, args.lanes], device=args.device, block_dim=32)
    wp.synchronize_device(args.device)
    prepared = time.perf_counter()
    batch.run(args.ticks, args.chunk, args.lanes)
    finished = time.perf_counter()
    physical = batch.project()
    output = {"format": "warp-physical-report-1", "go_snapshot_compatible": False, "warp": wp.__version__,
              "device": wp.get_device(args.device).name, "sources": sources, "additional_ticks": args.ticks,
              "prepare_and_compile_seconds": prepared-started, "solve_seconds": finished-prepared,
              "physical_sha256": physical_hash(physical), "worlds": physical}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("x", encoding="utf-8") as f:
        json.dump(output, f, separators=(",", ":"))
        f.write("\n")
    print(json.dumps({k:v for k,v in output.items() if k != "worlds"}, indent=2))


if __name__ == "__main__":
    main()
