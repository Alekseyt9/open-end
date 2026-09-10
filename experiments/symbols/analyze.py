"""Recompute Stage 15 evidence from local artifacts, using only the standard library."""
import argparse
import copy
import hashlib
import json
from pathlib import Path


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def encoded_hash(value):
    # These snapshots contain only ASCII keys/identifiers, integers and booleans.
    return hashlib.sha256((json.dumps(value, separators=(",", ":"), ensure_ascii=False) + "\n").encode()).hexdigest()


def normalized(world):
    value = copy.deepcopy(world)
    value["config"].pop("symbols")
    value.pop("symbols")
    return value


def frames(path):
    return [json.loads(line) for line in path.read_text(encoding="utf-8-sig").splitlines()]


def counts(state):
    return {k: v for k, v in state.items() if k != "actors"}


def check_counts(state):
    for key, value in counts(state).items():
        assert value == sum(a[key] for a in state.get("actors", {}).values())
    assert state["pair_reads"] <= state["nonempty_reads"] <= state["reads"]
    assert state["foreign_reads"] <= state["nonempty_reads"]
    assert state["context_lookups"] <= state["lookups"]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--batch", type=Path, default=Path("data/symbols-stage15"))
    parser.add_argument("--assay", type=Path, default=Path("data/symbol-assay-stage15"))
    parser.add_argument("--out", type=Path, default=Path("experiments/symbols/evidence.json"))
    args = parser.parse_args()
    manifest, assay_manifest = read(args.batch / "manifest.json"), read(args.assay / "manifest.json")
    assert manifest["Status"] == assay_manifest["status"] == "complete"
    rows, trials = read(args.batch / "summary.json"), read(args.assay / "results.json")
    assert len(rows) == manifest["Total"] == manifest["Completed"] == 32
    assert len(trials) == assay_manifest["total"] == assay_manifest["completed"] == 24
    sources, samples, evidence_rows = {}, {}, []
    for row in rows:
        path = args.batch / row["Snapshot"]
        assert digest(path) == row["StateSHA256"]
        snap = read(path)
        assert encoded_hash(snap) == row["StateSHA256"]
        w = snap["world"]
        check_counts(w["symbols"])
        fs = frames(args.batch / row["Metrics"])
        assert len(fs) == 101 and fs[0]["tick"] == 0 and fs[-1]["tick"] == 100000
        assert fs[-1]["symbols"]["state"] == w["symbols"]
        for f in fs:
            check_counts(f["symbols"]["state"])
        key = (row["Case"], row["Seed"])
        assert key not in sources
        sources[key], samples[key] = snap, fs
        evidence_rows.append(dict(case=row["Case"], seed=row["Seed"], snapshot_sha256=row["StateSHA256"],
                                  metrics_sha256=digest(args.batch / row["Metrics"]),
                                  counts=counts(w["symbols"]), words=fs[-1]["symbols"]["words"],
                                  pairs=fs[-1]["symbols"]["pairs"], entities=len(w["particles"]),
                                  readers=[dict(genome=h, **a, code=w["genomes"][h]["code"])
                                           for h, a in w["symbols"].get("actors", {}).items() if a["nonempty_reads"]]))
        if row["Case"] == "symbols-no-mutation":
            assert not any(counts(w["symbols"]).values())
    comparisons = []
    for seed in range(1, 9):
        baseline = sources[("symbols", seed)]["world"]
        for case in ("symbols-scrambled", "symbols-unreadable"):
            checkpoints_equal = True
            for a, b in zip(samples[("symbols", seed)], samples[(case, seed)]):
                a, b = copy.deepcopy(a), copy.deepcopy(b)
                for f in (a, b):
                    f.pop("symbols")
                    f["telemetry"].pop("session_initial_world_sha256")
                checkpoints_equal &= a == b
            comparisons.append(dict(seed=seed, arm=case, other_endpoint_fields_equal=
                                    normalized(baseline) == normalized(sources[(case, seed)]["world"]),
                                    other_sampled_metrics_equal=checkpoints_equal))
    endpoints, continuation_rows = {}, []
    for trial in trials:
        path = args.assay / trial["snapshot"]
        assert digest(path) == trial["final_sha256"]
        seed, mode = trial["seed"], trial["variant"]
        source = sources[("symbols", seed)]
        assert trial["source_sha256"] == encoded_hash(source)
        initial = copy.deepcopy(source)
        initial["world"]["config"]["symbols"] = mode
        assert trial["initial_sha256"] == encoded_hash(initial)
        end = read(path)["world"]
        check_counts(end["symbols"])
        fs = frames(args.assay / trial["metrics"])
        assert fs[0]["symbols"]["state"] == source["world"]["symbols"]
        assert fs[-1]["symbols"]["state"] == end["symbols"]
        assert fs[0]["tick"] == 100000 and fs[-1]["tick"] == 120000
        delta = {k: v - source["world"]["symbols"][k] for k, v in counts(end["symbols"]).items()}
        assert all(trial["summary"]["symbols"][k] == v for k, v in delta.items())
        endpoints[(seed, mode)] = end
        continuation_rows.append(dict(seed=seed, arm=mode, source_sha256=trial["source_sha256"],
                                      initial_sha256=trial["initial_sha256"], final_sha256=trial["final_sha256"],
                                      metrics_sha256=digest(args.assay / trial["metrics"]), counts=delta))
    paired = [dict(seed=seed, arm=mode, other_endpoint_fields_equal=
                   normalized(endpoints[(seed, "persistent")]) == normalized(endpoints[(seed, mode)]))
              for seed in range(1, 9) for mode in ("scrambled", "unreadable")]
    evidence = dict(version=1, batch_manifest=manifest, assay_manifest=assay_manifest,
                    normalization="Remove config.symbols and world.symbols only; retain words, memories, both RNGs and all other fields.",
                    batch=evidence_rows, development_comparisons=comparisons,
                    continuations=continuation_rows, paired_comparisons=paired)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(dict(batch=len(rows), continuations=len(trials),
                          development_equal=sum(x["other_endpoint_fields_equal"] for x in comparisons),
                          sampled_equal=sum(x["other_sampled_metrics_equal"] for x in comparisons),
                          continuation_equal=sum(x["other_endpoint_fields_equal"] for x in paired))))


if __name__ == "__main__":
    main()
