"""Verify matched sources, resource ledgers, ancestry metadata and aggregation."""
import argparse
import copy
import hashlib
import json
from pathlib import Path


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def encoded(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode()


def snapshot_hash(value):
    return hashlib.sha256(encoded(value) + b"\n").hexdigest()


def physical_audit(w):
    a = w["accounting"]
    energy = sum(c["energy"] + c.get("signal", 0) + 8*c["chemical"][0] + 4*c["chemical"][1] for c in w["cells"])
    energy += sum(p["energy"] for p in w["particles"].values())
    matter = len(w["particles"]) + sum(c["matter"] + c.get("terrain", 0) for c in w["cells"])
    assert energy == a["initial_energy"] + a["injected"] - a["dissipated"]
    assert matter == a["initial_matter"]
    assert sum(sum(c["chemical"]) for c in w["cells"]) == a["initial_chemical"]
    width, height = w["config"]["width"], w["config"]["height"]
    for edge in w["relations"].values():
        p, q = (w["particles"][str(edge[k])] for k in ("a", "b"))
        x, y = p["position"] % width, p["position"] // width
        assert q["position"] in [((y-1) % height)*width+x, y*width+(x+1) % width,
                                  ((y+1) % height)*width+x, y*width+(x-1) % width]


def census(w):
    ids = {edge[k] for edge in w["relations"].values() for k in ("a", "b")}
    width, height = w["config"]["width"], w["config"]["height"]
    result = dict(initial_linked_particles=len(ids), initial_linked_with_move_code=0,
                  initial_linked_with_unbind_code=0, initial_linked_with_free_matter_neighbor=0)
    for id_ in ids:
        p = w["particles"][str(id_)]
        code = p["code"] or []
        result["initial_linked_with_move_code"] += any(i["op"] == 2 for i in code)
        result["initial_linked_with_unbind_code"] += any(i["op"] == 16 for i in code)
        x, y = p["position"] % width, p["position"] // width
        ns = [((y-1) % height)*width+x, y*width+(x+1) % width, ((y+1) % height)*width+x, y*width+(x-1) % width]
        result["initial_linked_with_free_matter_neighbor"] += any(w["cells"][i]["occupant"] == 0 and w["cells"][i]["matter"] > 0 for i in ns)
    return result


def verify_run(path):
    manifest = read(path / "manifest.json")
    assert manifest["status"] == "complete" and manifest["total"] == manifest["completed"] == 16
    root = Path(manifest["input"])
    sources = {r["Seed"]: r for r in read(root / "summary.json") if r["Case"] == manifest["case"]}
    trials = read(path / "results.json")
    assert len(trials) == 16 and len({(r["seed"], r["variant"]) for r in trials}) == 16
    reference_path = Path("data/collectives-stage12") if manifest["case"] == "environment" else Path("data/symbol-assay-stage15-verified")
    reference_mode = "intact" if manifest["case"] == "environment" else "persistent"
    refs = {r["seed"]: r["final_sha256"] for r in read(reference_path / "results.json") if r["variant"] == reference_mode}
    summary = []
    for r in trials:
        source_path = root / sources[r["seed"]]["Snapshot"]
        source = read(source_path)
        assert digest(source_path) == r["source_sha256"] == snapshot_hash(source)
        initial = copy.deepcopy(source)
        if r["variant"] == "yielding":
            initial["format"] = 8
            initial["world"]["config"]["bond_motion"] = "yielding"
        else:
            assert r["variant"] == "intact"
        assert snapshot_hash(initial) == r["initial_sha256"]
        snap_path = path / r["snapshot"]
        assert digest(snap_path) == r["final_sha256"]
        final = read(snap_path)["world"]
        physical_audit(final)
        if r["variant"] == "intact":
            assert r["final_sha256"] == refs[r["seed"]]
        d, s = r["diagnostics"], r["summary"]
        for key, value in census(source["world"]).items():
            assert d[key] == value
        assert d["bond_break_energy"] == 2*d["motion_broken_bonds"]
        assert d["successful_yield_moves"] <= d["motion_broken_bonds"] <= 4*d["successful_yield_moves"]
        assert d["tick_start_allocation_no_space"] + d["tick_start_allocation_no_matter"] <= d["tick_start_linked_allocate_intents"]
        assert d["single_founder_group_ticks"] + d["multiple_founder_group_ticks"] == d["founder_free_descendant_group_ticks"]
        assert d["parent_absent_group_ticks"] <= d["founder_free_descendant_group_ticks"]
        assert d["qualified_clonal_group_ticks"] <= d["single_founder_group_ticks"]
        if r["variant"] == "intact":
            assert d["successful_yield_moves"] == d["bond_break_energy"] == 0
        c = s["collectives"]["end"]
        assert d["successful_linked_copies"] == s["collectives"]["copies_by_linked_particles"]
        cohorts = {x["id"]: x for x in c["cohorts"]}
        seen = set()
        for candidate in d["clonal_daughter_candidates"]:
            key = (candidate["cohort"], tuple(candidate["members"]))
            assert key not in seen
            seen.add(key)
            co = cohorts[candidate["cohort"]]
            assert candidate["members"] == sorted(set(candidate["members"])) and len(candidate["members"]) >= 2
            assert not set(candidate["members"]) & set(co["founders"])
            assert candidate["founder"] in co["founders"]
            assert co["tagged_tick"] <= candidate["since_tick"] <= candidate["qualified_tick"] <= candidate["last_observed_tick"] <= s["to_tick"]
            assert candidate["qualified_tick"] - candidate["since_tick"] >= manifest["group_age"]
            assert sum(g["count"] for g in candidate["genomes_at_qualification"]) == len(candidate["members"])
            assert all(g["hash"] in final["genomes"] for g in candidate["genomes_at_qualification"])
        if d["skipped_clonal_group_ticks"] == 0:
            assert d["max_concurrent_qualified_clonal_groups"] <= len(seen)
        assert d["qualified_clonal_group_ticks"] >= len(seen)
        fs = [json.loads(line) for line in (path / r["metrics"]).read_text().splitlines()]
        assert fs[0]["tick"] == source["world"]["tick"] and fs[-1]["tick"] == final["tick"]
        assert fs[-1]["telemetry"]["collectives"] == c
        summary.append(dict(seed=r["seed"], arm=r["variant"], source_sha256=r["source_sha256"],
                            initial_sha256=r["initial_sha256"], final_sha256=r["final_sha256"],
                            metrics_sha256=digest(path / r["metrics"]), population=s["population"]["end"],
                            stable_group_ticks=s["collectives"]["stable_group_ticks"],
                            strict_daughters=len(c["daughter_candidates"]), productive_strict=c["productive_daughter_candidates"],
                            observer_skipped=c["skipped_cohorts_or_candidates"], diagnostics=d))
    return dict(manifest=manifest, historical_intact_hashes_match=True, trials=summary)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--runs", type=Path, nargs="+", default=[Path("data/multicell-motion-environment-verified"), Path("data/multicell-motion-symbols-verified")])
    parser.add_argument("--out", type=Path, default=Path("experiments/multicell-motion/evidence.json"))
    args = parser.parse_args()
    result = dict(version=1, verification_scope="Hashes, initial census, resource ledgers, counter identities, reported ancestry constraints and timestamps; persistence across every intermediate tick is covered by Go observation and tests, not independently replayed in Python.", batches=[verify_run(p) for p in args.runs])
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    for b in result["batches"]:
        for arm in ("intact", "yielding"):
            rs = [r for r in b["trials"] if r["arm"] == arm]
            print(b["manifest"]["case"], arm, "clonal_sets", sum(len(r["diagnostics"]["clonal_daughter_candidates"]) for r in rs),
                  "stable_group_ticks", sum(r["stable_group_ticks"] for r in rs),
                  "max_concurrent_in_one_world", max(r["diagnostics"]["max_concurrent_qualified_clonal_groups"] for r in rs))


if __name__ == "__main__":
    main()
