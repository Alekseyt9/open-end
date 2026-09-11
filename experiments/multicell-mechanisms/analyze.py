"""Verify mobility controls and proportional Y provenance (run at repo root)."""
import argparse
from collections import Counter
import copy
import hashlib
import json
import math
from pathlib import Path
import runpy

helpers = runpy.run_path("experiments/multicell-motion/analyze.py")
physical_audit, snapshot_hash = helpers["physical_audit"], helpers["snapshot_hash"]


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def functional_hash(snapshot):
    # Exactly these two state fields describe edges or count successful BIND.
    # Everything else, including both RNGs and all resources/ancestry, is kept.
    w = copy.deepcopy(snapshot["world"])
    w.pop("relations")
    for genome in w["genomes"].values():
        genome.pop("binds", None)
    return hashlib.sha256(json.dumps(w, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, default=Path("data/multicell-mechanisms-v2"))
    parser.add_argument("--reference", type=Path, default=Path("data/multicell-roles-final/results.json"))
    parser.add_argument("--output", type=Path, default=Path("experiments/multicell-mechanisms/evidence.json"))
    args = parser.parse_args()
    manifest = read(args.input / "manifest.json")
    rows = read(args.input / "results.json")
    assert manifest["status"] == "complete" and manifest["total"] == manifest["completed"] == len(rows)
    assert manifest["protocol_version"] == 2 and manifest["suite"] == "mechanisms"
    witnesses = {w["snapshot"]: w for w in read(Path(manifest["input"]) / "witnesses.json")}
    sources = {k: read(Path(manifest["input"]) / k) for k in witnesses}
    for name, w in witnesses.items():
        assert digest(Path(manifest["input"]) / name) == snapshot_hash(sources[name]) == w["snapshot_sha256"]
    reference = {(r["witness"], r["protocol"]["mode"]): r for r in read(args.reference) if r["protocol"]["mode"] in ("intact", "no-bonds")}
    controls = {r["witness"]: r for r in rows if r["protocol"]["mode"] == "intact"}
    seen, projections = set(), {}
    evidence, baseline = [], []
    for r in rows:
        p, c = r["protocol"], r["chemistry"]
        key = (r["witness"], p["mode"], p.get("donor", 0))
        assert key not in seen; seen.add(key)
        source = sources[r["witness"]]
        assert p["version"] == 2 and p["members"] == witnesses[r["witness"]]["candidate"]["members"]
        assert r["source_sha256"] == witnesses[r["witness"]]["snapshot_sha256"]
        initial = copy.deepcopy(source)
        removed = 0
        if p["mode"] in ("no-bonds", "no-bonds-anchored"):
            edges = initial["world"]["relations"]
            for k, e in list(edges.items()):
                if e["a"] in p["members"] and e["b"] in p["members"]:
                    del edges[k]; removed += 1
        assert removed == r["initial_removed_bonds"] and snapshot_hash(initial) == r["initial_sha256"]
        path = args.input / r["released_snapshot"]
        assert digest(path) == r["final_sha256"]
        final = read(path); w = final["world"]
        physical_audit(w)
        assert w["config"] == source["world"]["config"]
        assert w["tick"] == r["final_tick"] == r["start_tick"]+p["ticks"]
        if p["mode"] in ("intact", "no-bonds"):
            assert r["final_sha256"] == reference[(r["witness"], p["mode"])]["final_sha256"]
        if p["mode"] == "intact":
            assert r["ordinary_control_verified"]
        projections[key] = functional_hash(final)
        n = len(p["members"])
        assert len(c["roles"]) == n and len(c["consumed_y_by_producer_consumer"]) == n+2
        assert len(c["initial_y_by_producer"]) == len(c["produced_y_by_producer"]) == len(c["remaining_y_by_producer"]) == n+2
        assert c["maximum_endpoint_cell_mass_error"] < 1e-7
        assert sum(c["initial_y_by_producer"]) == sum(x["chemical"][1] for x in source["world"]["cells"])
        matrix = c["consumed_y_by_producer_consumer"]
        for j in range(n+2):
            assert len(matrix[j]) == n+1 and all(math.isfinite(x) and x >= 0 for x in matrix[j])
            assert math.isclose(c["initial_y_by_producer"][j]+c["produced_y_by_producer"][j], sum(matrix[j])+c["remaining_y_by_producer"][j], abs_tol=1e-6, rel_tol=1e-10)
        assert math.isclose(sum(c["remaining_y_by_producer"]), sum(x["chemical"][1] for x in w["cells"]), abs_tol=1e-6)
        assert math.isclose(sum(c["produced_y_by_producer"]), w["accounting"]["converted"][0]-source["world"]["accounting"]["converted"][0], abs_tol=1e-6)
        assert math.isclose(sum(map(sum, matrix)), w["accounting"]["converted"][1]-source["world"]["accounting"]["converted"][1], abs_tol=1e-6, rel_tol=1e-10)
        for j, role in enumerate(c["roles"]):
            assert role["founder"] == p["members"][j] == r["roles"][j]["founder"]
            units = role["lineage_reaction_units"]
            assert r["roles"][j]["converted_energy"] == 4*sum(units)
            assert all(0 <= a <= b for a, b in zip(role["original_cell_reaction_units"], units))
            assert c["produced_y_by_producer"][j+2] == units[0]
            assert math.isclose(sum(row[j+1] for row in matrix), units[1], abs_tol=1e-6)
            if role["founder"] == p.get("donor") and p["mode"].startswith("no-reaction"):
                assert units[int(p["mode"][-1])] == 0
        cross = sum(matrix[a+2][b+1] for a in range(n) for b in range(n) if a != b)
        own = sum(matrix[a+2][a+1] for a in range(n))
        outside = sum(matrix[1][1:]); unknown = sum(matrix[0][1:])
        if p["mode"] == "intact":
            baseline.append({"witness":r["witness"], "case":r["case"], "seed":r["seed"], "members":p["members"], "chemistry":c,
                             "peer_y_units":cross, "own_y_units":own, "outside_y_units":outside, "initial_unknown_y_units":unknown})
        def total(row, field, exclude=0):
            return sum(x[field] for x in row["roles"] if x["founder"] != exclude)
        control = controls[r["witness"]]; donor = p.get("donor", 0)
        evidence.append({"witness":r["witness"], "case":r["case"], "seed":r["seed"], "mode":p["mode"], "donor":donor,
                         "final_sha256":r["final_sha256"], "functional_sha256":projections[key],
                         "copies":total(r,"successful_copies"), "cell_ticks":total(r,"lineage_cell_ticks"),
                         "original_survival_ticks":total(r,"founder_alive_ticks"),
                         "other_original_survival_delta":total(r,"founder_alive_ticks",donor)-total(control,"founder_alive_ticks",donor),
                         "other_copy_delta":total(r,"successful_copies",donor)-total(control,"successful_copies",donor),
                         "peer_y_units":cross, "suppressed_actions":sum(sum(x["suppressed_or_blinded_by_opcode"].values()) for x in r["roles"])})
    expected = {(name, mode, 0) for name in witnesses for mode in ("intact","no-bonds","anchored","no-bonds-anchored")}
    expected |= {(name, mode, id_) for name, w in witnesses.items() for mode in ("no-reaction0","no-reaction1") for id_ in w["candidate"]["members"]}
    assert seen == expected
    mobility = [{"witness":name,"anchored_equal_without_bond_fields":projections[(name,"anchored",0)] == projections[(name,"no-bonds-anchored",0)]} for name in witnesses]
    by_key = {(r["witness"], r["protocol"]["mode"], r["protocol"].get("donor",0)): r for r in rows}
    for pair in mobility:
        name = pair["witness"]
        pair["anchored_chemistry_equal"] = by_key[(name,"anchored",0)]["chemistry"] == by_key[(name,"no-bonds-anchored",0)]["chemistry"]
    aggregates = {}
    for mode in sorted({r["mode"] for r in evidence}):
        selected = [r for r in evidence if r["mode"] == mode]
        aggregates[mode] = {"trials":len(selected)}
        for k in ("copies","cell_ticks","original_survival_ticks","suppressed_actions","peer_y_units"):
            aggregates[mode][k] = sum(r[k] for r in selected)
        aggregates[mode]["other_survival_delta_signs"] = dict(Counter("negative" if r["other_original_survival_delta"] < 0 else "positive" if r["other_original_survival_delta"] > 0 else "zero" for r in selected))
    output = {"manifest":manifest, "results_sha256":digest(args.input/"results.json"), "reference_results_sha256":digest(args.reference),
              "baseline":baseline, "mobility_controls":mobility, "aggregates":aggregates, "comparisons":evidence,
              "verification_scope":"Matched source/initial/final hashes, historical controls, full physical ledgers, tracer balances, reaction and acquisition identities. Projection removes only bond graph and genome BIND counters. Python does not independently replay fractional transport, intermediate ancestry, or every original-cell position."}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(output, indent=2)+"\n", encoding="utf-8")
    print(json.dumps(aggregates,indent=2))
    print("Matched anchored endpoints excluding bond fields:", sum(r["anchored_equal_without_bond_fields"] for r in mobility),"/",len(mobility))
    print("Baseline Y:", {key:sum(b[key] for b in baseline) for key in ("peer_y_units","own_y_units","outside_y_units","initial_unknown_y_units")})


if __name__ == "__main__":
    main()
