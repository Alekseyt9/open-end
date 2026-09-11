"""Audit role-assay artifacts and compute matched, descriptive comparisons.

Run from the repository root. No group-level independence or significance is
assumed. This checks artifacts/ledgers, not every intermediate ancestry event.
"""
import argparse
from collections import Counter
import copy
import hashlib
import json
from pathlib import Path
import runpy

audit_helpers = runpy.run_path("experiments/multicell-motion/analyze.py")
physical_audit = audit_helpers["physical_audit"]
snapshot_hash = audit_helpers["snapshot_hash"]


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, default=Path("data/multicell-roles-final"))
    parser.add_argument("--output", type=Path, default=Path("experiments/multicell-roles/evidence.json"))
    args = parser.parse_args()
    manifest = read(args.input / "manifest.json")
    rows = read(args.input / "results.json")
    assert manifest["status"] == "complete" and manifest["total"] == manifest["completed"] == len(rows)
    witnesses = {w["snapshot"]: w for w in read(Path(manifest["input"]) / "witnesses.json")}
    controls = {r["witness"]: r for r in rows if r["protocol"]["mode"] == "intact"}
    assert len(controls) == len(witnesses)
    seen = set()
    comparisons, baseline = [], []
    initial_programs = {}
    for r in rows:
        p = r["protocol"]
        key = (r["witness"], p["mode"], p.get("donor", 0))
        assert key not in seen
        seen.add(key)
        witness = witnesses[r["witness"]]
        src_path = Path(manifest["input"]) / r["witness"]
        src = read(src_path)
        assert digest(src_path) == snapshot_hash(src) == r["source_sha256"] == witness["snapshot_sha256"]
        assert p["version"] == 1 and p["members"] == witness["candidate"]["members"]
        assert r["start_tick"] == src["world"]["tick"] and r["final_tick"] == r["start_tick"] + p["ticks"]
        initial = copy.deepcopy(src)
        removed = 0
        if p["mode"] == "no-bonds":
            edges = initial["world"]["relations"]
            for k, edge in list(edges.items()):
                if edge["a"] in p["members"] and edge["b"] in p["members"]:
                    del edges[k]
                    removed += 1
        assert removed == r["initial_removed_bonds"] and snapshot_hash(initial) == r["initial_sha256"]
        final_path = args.input / r["released_snapshot"]
        assert digest(final_path) == r["final_sha256"]
        final = read(final_path)["world"]
        physical_audit(final)
        assert final["tick"] == r["final_tick"] and final["config"] == src["world"]["config"]
        assert [x["founder"] for x in r["roles"]] == p["members"]
        assert r["frames"][0]["tick"] == r["start_tick"] and r["frames"][-1]["tick"] == r["final_tick"]
        assert [x["lineage_alive"] for x in r["roles"]] == r["frames"][-1]["alive_by_founder"]
        for role in r["roles"]:
            assert role["initial_genome"] == src["world"]["particles"][str(role["founder"])]["genome"]
            assert role["acquired_energy"] == role["absorbed_energy"] + role["converted_energy"]
            assert 0 <= role["founder_alive_ticks"] <= p["ticks"] and role["founder_alive_ticks"] <= role["lineage_cell_ticks"]
            for op, count in role["suppressed_or_blinded_by_opcode"].items():
                assert 0 <= count <= role["paid_instructions_by_opcode"][op]
            if p["mode"] == "no-acquisition" and role["founder"] == p["donor"]:
                assert role["acquired_energy"] == 0
        cross = sum(f["energy"] for f in r["flows"] if f["kind"] == "transfer" and f["source_founder"] and f["target_founder"] and f["source_founder"] != f["target_founder"])
        for f in r["flows"]:
            assert 0 <= f["bonded_energy"] <= f["energy"] and 0 <= f["observed_copy_parent_to_child_energy"] <= f["energy"]
        if p["mode"] == "no-peer-sharing":
            assert cross == 0
        if p["mode"] == "no-sharing":
            assert not any(f["kind"] == "transfer" and f["source_founder"] and f["target_founder"] for f in r["flows"])
        control = controls[r["witness"]]
        if p["mode"] == "intact":
            assert r["ordinary_control_verified"] and not any(x["suppressed_or_blinded_by_opcode"] for x in r["roles"])
            for role in r["roles"]:
                code = src["world"]["particles"][str(role["founder"])]["code"]
                prior = initial_programs.setdefault(role["initial_genome"], code)
                assert prior == code
            baseline.append({"witness": r["witness"], "case": r["case"], "seed": r["seed"], "source_sha256": r["source_sha256"],
                             "members": p["members"], "roles": r["roles"], "flows": r["flows"],
                             "together_ticks": r["all_original_members_connected_ticks"],
                             "stable_descendant_component_ticks": r["stable_founder_free_component_ticks"],
                             "cross_lineage_transfer_energy": cross})
        donor = p.get("donor", 0)
        def total(row, metric, exclude=0):
            return sum(x[metric] for x in row["roles"] if x["founder"] != exclude)
        comparisons.append({"witness": r["witness"], "case": r["case"], "seed": r["seed"], "mode": p["mode"], "donor": donor,
                            "final_sha256": r["final_sha256"], "changed_endpoint": r["final_sha256"] != control["final_sha256"],
                            "suppressed_actions": sum(sum(x["suppressed_or_blinded_by_opcode"].values()) for x in r["roles"]),
                            "copies": total(r, "successful_copies"), "copy_delta": total(r, "successful_copies")-total(control, "successful_copies"),
                            "lineage_cell_ticks": total(r, "lineage_cell_ticks"),
                            "lineage_cell_tick_delta": total(r, "lineage_cell_ticks")-total(control, "lineage_cell_ticks"),
                            "original_survival_tick_delta": total(r, "founder_alive_ticks")-total(control, "founder_alive_ticks"),
                            "other_original_survival_tick_delta": total(r, "founder_alive_ticks", donor)-total(control, "founder_alive_ticks", donor),
                            "together_ticks": r["all_original_members_connected_ticks"],
                            "stable_descendant_component_ticks": r["stable_founder_free_component_ticks"]})
    expected = {(name, mode, 0) for name in witnesses for mode in ["intact", "no-sharing", "no-peer-sharing", "no-signals", "no-bonds"]}
    expected |= {(name, "no-acquisition", id_) for name, w in witnesses.items() for id_ in w["candidate"]["members"]}
    assert seen == expected
    aggregates = {}
    for mode in sorted({r["mode"] for r in comparisons}):
        selected = [r for r in comparisons if r["mode"] == mode]
        aggregates[mode] = {"trials": len(selected), "changed_endpoints": sum(r["changed_endpoint"] for r in selected)}
        for k in ["suppressed_actions", "copies", "copy_delta", "lineage_cell_ticks", "lineage_cell_tick_delta", "original_survival_tick_delta", "together_ticks", "stable_descendant_component_ticks"]:
            aggregates[mode][k] = sum(r[k] for r in selected)
        aggregates[mode]["other_original_survival_delta_signs"] = dict(Counter("negative" if r["other_original_survival_tick_delta"] < 0 else "positive" if r["other_original_survival_tick_delta"] > 0 else "zero" for r in selected))
    output = {"manifest": manifest, "results_sha256": digest(args.input / "results.json"), "source_worlds": sorted({(r["case"], r["seed"]) for r in rows}),
              "baseline": baseline, "initial_programs": initial_programs, "aggregates": aggregates, "comparisons": comparisons,
              "verification_scope": "Source, treatment-start and endpoint hashes; endpoint resource ledgers and local bonds; complete matched arms; telemetry identities. Ordinary Go replay independently verifies intact endpoints. Intermediate lineage tags and causal interpretation are not independently replayed by Python."}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(aggregates, indent=2))
    print(f"Verified {len(rows)} trials from {len(witnesses)} related witnesses: {args.output}")


if __name__ == "__main__":
    main()
