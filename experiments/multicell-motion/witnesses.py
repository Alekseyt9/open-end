"""Replay intact sources and preserve mixed-genome clonal qualification states."""
import argparse
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")


def components(w):
    adj = {}
    for e in w["relations"].values():
        adj.setdefault(e["a"], set()).add(e["b"])
        adj.setdefault(e["b"], set()).add(e["a"])
    lookup = {}
    for id_ in sorted(adj):
        if id_ in lookup:
            continue
        visited, todo = {id_}, [id_]
        while todo:
            for other in adj[todo.pop()] - visited:
                visited.add(other)
                todo.append(other)
        members = tuple(sorted(visited))
        for member in members:
            lookup[member] = members
    return lookup


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--runs", type=Path, nargs="+", default=[Path("data/multicell-motion-environment-verified"), Path("data/multicell-motion-symbols-verified")])
    parser.add_argument("--out", type=Path, default=Path("data/multicell-mixed-witnesses"))
    parser.add_argument("--workers", type=int, default=16)
    args = parser.parse_args()
    if not 1 <= args.workers <= 256:
        parser.error("workers must be in 1..256")
    jobs = []
    for path in args.runs:
        manifest = read(path / "manifest.json")
        assert manifest["status"] == "complete"
        root = Path(manifest["input"])
        rows = {r["Seed"]: r for r in read(root / "summary.json") if r["Case"] == manifest["case"]}
        for trial in read(path / "results.json"):
            if trial["variant"] != "intact":
                continue
            source = root / rows[trial["seed"]]["Snapshot"]
            assert digest(source) == trial["source_sha256"]
            cohorts = {c["id"]: c for c in trial["summary"]["collectives"]["end"]["cohorts"]}
            for candidate in trial["diagnostics"]["clonal_daughter_candidates"]:
                if len(candidate["genomes_at_qualification"]) < 2:
                    continue
                assert trial["summary"]["from_tick"] <= candidate["qualified_tick"] <= trial["summary"]["to_tick"]
                jobs.append(dict(case=manifest["case"], seed=trial["seed"], source=source,
                                 source_sha256=trial["source_sha256"], from_tick=trial["summary"]["from_tick"],
                                 candidate=candidate, cohort=cohorts[candidate["cohort"]]))
    args.out.mkdir(parents=True, exist_ok=False)
    start = time.perf_counter()
    manifest = dict(status="running", workers=args.workers, total=len(jobs), completed=0)
    write(args.out / "manifest.json", manifest)
    binary = (args.out / ("sim.exe" if os.name == "nt" else "sim")).resolve()
    try:
        subprocess.run(["go", "build", "-o", str(binary), "./cmd/sim"], check=True)
        manifest["binary_sha256"] = digest(binary)

        def run(job):
            c = job["candidate"]
            name = f'{job["case"]}-seed{job["seed"]}-t{c["qualified_tick"]}-cohort{c["cohort"]}-founder{c["founder"]}.json'
            ticks = c["qualified_tick"] - job["from_tick"]
            path = (args.out / name).resolve()
            output = subprocess.run([str(binary), "-load", str(job["source"].resolve()), "-ticks", str(ticks),
                                     "-every", str(max(1, ticks)), "-save", str(path)], check=True,
                                    capture_output=True, text=True, env=dict(os.environ, GOMAXPROCS="1"),
                                    creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
            sha = output.stdout.strip().split("state_sha256=")[-1]
            assert sha == digest(path)
            w = read(path)["world"]
            assert w["tick"] == c["qualified_tick"]
            groups = components(w)
            assert groups[c["members"][0]] == tuple(c["members"])
            actual = Counter(w["particles"][str(id_)]["genome"] for id_ in c["members"])
            assert actual == {g["hash"]: g["count"] for g in c["genomes_at_qualification"]}
            assert all(w["particles"][str(id_)]["code"] for id_ in c["members"])
            founders = set(job["cohort"]["founders"])
            assert not founders.intersection(c["members"]) and c["founder"] in founders
            parents = Counter(groups[id_] for id_ in founders if id_ in groups)
            assert max(parents.values(), default=0) >= 2
            assert digest(job["source"]) == job["source_sha256"]
            return dict(case=job["case"], seed=job["seed"], source_sha256=job["source_sha256"],
                        snapshot=name, snapshot_sha256=sha, candidate=c,
                        parent_founders=[id_ for id_ in sorted(founders) if id_ in groups and parents[groups[id_]] >= 2])

        names = [(j["case"], j["seed"], j["candidate"]["qualified_tick"], j["candidate"]["cohort"], j["candidate"]["founder"]) for j in jobs]
        assert len(names) == len(set(names)), "witness filenames would collide"
        results = []
        with ThreadPoolExecutor(max_workers=args.workers) as executor:
            for future in as_completed([executor.submit(run, job) for job in jobs]):
                results.append(future.result())
        results.sort(key=lambda r: (r["case"], r["seed"], r["candidate"]["qualified_tick"]))
        write(args.out / "witnesses.json", results)
        manifest.update(status="complete", completed=len(results), seconds=time.perf_counter()-start)
        write(args.out / "manifest.json", manifest)
        print(f'Preserved and checked {len(results)} mixed-genome witnesses in {manifest["seconds"]:.2f} seconds.')
    except BaseException as error:
        manifest.update(status="failed", error=str(error))
        write(args.out / "manifest.json", manifest)
        raise


if __name__ == "__main__":
    main()
