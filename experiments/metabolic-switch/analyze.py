"""Audit the switching-cost screen and matched chemical witness continuations."""
import argparse
import copy
import hashlib
import json
import math
from pathlib import Path
import runpy

h = runpy.run_path("experiments/multicell-motion/analyze.py")
physical_audit, snapshot_hash = h["physical_audit"], h["snapshot_hash"]
def read(p): return json.loads(Path(p).read_text(encoding="utf-8-sig"))
def digest(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def prepared(source,cost):
    w=copy.deepcopy(source)
    if cost:
        w["format"]=9; w["world"]["config"]["metabolic_switch_cost"]=cost
    return w
def audit_switch(w,cost):
    physical_audit(w)
    assert w["config"].get("metabolic_switch_cost",0)==cost
    a=w["accounting"]; events=a.get("metabolic_switches",0); starved=a.get("metabolic_switch_starved",0); spent=a.get("metabolic_switch_energy",0)
    assert 0<=spent<=a["dissipated"] and events+starved<=a["instructions"]
    assert events*cost+starved<=spent<=(events+starved)*cost
    for p in w["particles"].values():
        assert p.get("metabolic_state",0) in ((0,1,2) if cost else (0,))
def kind(u):
    if sum(u)<32: return "low-activity"
    if u[0]*10>=sum(u)*9: return "reaction0"
    if u[1]*10>=sum(u)*9: return "reaction1"
    return "generalist"
def screen(path):
    m=read(path/"manifest.json"); rows=read(path/"results.json")
    m["results_sha256"]=digest(path/"results.json")
    assert m["status"]=="complete" and m["total"]==m["completed"]==len(rows)==32
    root=Path(m["input"]); source_meta={x["Seed"]:x for x in read(root/"summary.json") if x["Case"]==m["case"]}
    sources={seed:read(root/x["Snapshot"]) for seed,x in source_meta.items()}
    for seed,s in sources.items(): assert snapshot_hash(s)==source_meta[seed]["StateSHA256"]==digest(root/source_meta[seed]["Snapshot"])
    refs={r["seed"]:r["final_sha256"] for r in read(Path(f"data/multicell-motion-{m['case']}-verified/results.json")) if r["variant"]=="intact"}
    result=[];seen=set()
    for r in rows:
        cost=int(r["variant"]); key=(r["seed"],cost);assert key not in seen;seen.add(key)
        s=sources[r["seed"]]; assert r["source_sha256"]==snapshot_hash(s)
        assert r["initial_sha256"]==snapshot_hash(prepared(s,cost))
        final_path=path/r["snapshot"]; assert digest(final_path)==r["final_sha256"]
        w=read(final_path)["world"];audit_switch(w,cost)
        assert w["tick"]==s["world"]["tick"]+m["ticks"]
        if cost==0: assert r["final_sha256"]==refs[r["seed"]]
        previous={};from_tick=s["world"]["tick"]; totals=[0,0]; switches=heat=starved=0
        for frame in r["metabolic_windows"]:
            assert frame["From"]==from_tick and 0<frame["To"]-frame["From"]<=r["window_ticks"];from_tick=frame["To"]
            active={};special=[0,0];persist=[0,0];general=0
            assert len({p["id"] for p in frame["active_profiles"]})==len(frame["active_profiles"])
            for p in frame["active_profiles"]:
                assert p["class"]==kind(p["reaction_units"])!="low-activity"
                if not p["alive_at_window_end"]: continue
                active[p["id"]]=p["class"]
                if p["class"]=="generalist":general+=1;continue
                i=int(p["class"][-1]);special[i]+=1;persist[i]+=previous.get(p["id"])==p["class"]
            assert special==frame["living_specialists"] and persist==frame["same_cell_specialists_in_consecutive_windows"] and general==frame["living_generalists"]
            for group in frame["complementary_bonded_groups"]:
                assert group==sorted(set(group)) and {active.get(i) for i in group}>={"reaction0","reaction1"}
            for i in (0,1):
                assert sum(p["reaction_units"][i] for p in frame["active_profiles"])<=frame["reaction_units"][i]
                totals[i]+=frame["reaction_units"][i]
            switches+=frame["switches"];heat+=frame["switch_energy"];starved+=frame["switch_starved"]
            previous=active
        a=w["accounting"];sa=s["world"]["accounting"]
        assert totals==[a["converted"][i]-sa["converted"][i] for i in (0,1)] and from_tick==w["tick"]
        assert switches==a.get("metabolic_switches",0) and heat==a.get("metabolic_switch_energy",0) and starved==a.get("metabolic_switch_starved",0)
        last=r["metabolic_windows"][-1]
        assert last["population"]==len(w["particles"])
        for p in last["active_profiles"]:
            assert p["alive_at_window_end"]==(str(p["id"]) in w["particles"])
        result.append({"case":m["case"],"seed":r["seed"],"cost":cost,"source_sha256":r["source_sha256"],"final_sha256":r["final_sha256"],"snapshot":str(final_path),"copies":a["copies"]-sa["copies"],"switches":switches,"switch_energy":heat,"switch_starved":starved,
                       "windows":[{k:v for k,v in x.items() if k!="active_profiles"} for x in r["metabolic_windows"]],"final_living_specialist_profiles":[p for p in last["active_profiles"] if p["alive_at_window_end"] and p["class"]!="generalist"]})
    assert seen=={(seed,c) for seed in source_meta for c in (0,1,2,4)}
    return m,result
def witnesses(path):
    m=read(path/"manifest.json");rows=read(path/"results.json");assert m["status"]=="complete" and m["completed"]==m["total"]==len(rows)==56
    m["results_sha256"]=digest(path/"results.json")
    root=Path(m["input"]); sources={x["snapshot"]:read(root/x["snapshot"]) for x in read(root/"witnesses.json")}
    refs={r["witness"]:r["final_sha256"] for r in read("data/multicell-mechanisms-v2/results.json") if r["protocol"]["mode"]=="intact"}
    result=[];seen=set()
    for r in rows:
        p=r["protocol"];cost=p.get("metabolic_switch_cost",0);key=(r["witness"],cost);assert key not in seen;seen.add(key)
        source=sources[r["witness"]];assert p["version"]==3 and p["mode"]=="intact" and r["ordinary_control_verified"]
        assert snapshot_hash(source)==r["source_sha256"] and snapshot_hash(prepared(source,cost))==r["initial_sha256"]
        f=path/r["released_snapshot"];assert digest(f)==r["final_sha256"]
        w=read(f)["world"];audit_switch(w,cost)
        if cost==0: assert r["final_sha256"]==refs[r["witness"]]
        c=r["chemistry"];matrix=c["consumed_y_by_producer_consumer"];n=len(c["roles"])
        for j in range(n+2):
            assert math.isclose(c["initial_y_by_producer"][j]+c["produced_y_by_producer"][j],sum(matrix[j])+c["remaining_y_by_producer"][j],abs_tol=1e-6,rel_tol=1e-10)
        for j, role in enumerate(c["roles"]):
            assert sum(role["lineage_reaction_units"])*4==r["roles"][j]["converted_energy"]
            assert math.isclose(sum(row[j+1] for row in matrix),role["lineage_reaction_units"][1],abs_tol=1e-6)
        peer=sum(matrix[a+2][b+1] for a in range(n) for b in range(n) if a!=b)
        result.append({"witness":r["witness"],"cost":cost,"final_sha256":r["final_sha256"],"peer_y_units":peer,"total_y_consumed":sum(q["lineage_reaction_units"][1] for q in c["roles"]),"cell_ticks":sum(x["lineage_cell_ticks"] for x in r["roles"]),"copies":sum(x["successful_copies"] for x in r["roles"]),"chemistry":c})
    assert seen=={(name,cost) for name in sources for cost in (0,1,2,4)}
    return m,result
def main():
    parser=argparse.ArgumentParser()
    parser.add_argument("--environment",type=Path,default=Path("data/metabolic-switch-environment"))
    parser.add_argument("--symbols",type=Path,default=Path("data/metabolic-switch-symbols"))
    parser.add_argument("--witnesses",type=Path,default=Path("data/metabolic-switch-witnesses"))
    parser.add_argument("--output",type=Path,default=Path("experiments/metabolic-switch/evidence.json"));args=parser.parse_args()
    manifests=[];screens=[]
    for path in (args.environment,args.symbols):
        m,rs=screen(path);manifests.append(m);screens.extend(rs)
    wm,wr=witnesses(args.witnesses)
    aggregate=[]
    for case in ("environment","symbols"):
        for cost in (0,1,2,4):
            rs=[r for r in screens if r["case"]==case and r["cost"]==cost];last=[r["windows"][-1] for r in rs]
            aggregate.append({"case":case,"cost":cost,"population":sum(w["population"] for w in last),"copies":sum(r["copies"] for r in rs),"specialists":[sum(w["living_specialists"][j] for w in last) for j in (0,1)],"persistent_specialists":[sum(w["same_cell_specialists_in_consecutive_windows"][j] for w in last) for j in (0,1)],"generalists":sum(w["living_generalists"] for w in last),"complementary_groups":sum(len(w["complementary_bonded_groups"]) for w in last)})
    chem=[]
    for cost in (0,1,2,4):
        rs=[r for r in wr if r["cost"]==cost];peer=sum(r["peer_y_units"] for r in rs);total=sum(r["total_y_consumed"] for r in rs)
        chem.append({"cost":cost,"peer_y_units":peer,"total_y_consumed":total,"peer_percent":100*peer/total if total else None,"cell_ticks":sum(r["cell_ticks"] for r in rs),"copies":sum(r["copies"] for r in rs)})
    out={"screen_manifests":manifests,"witness_manifest":wm,"screen_aggregates":aggregate,"chemical_aggregates":chem,"screens":screens,"witnesses":wr,"limits":"Window activity labels are descriptive, not evidence of inherited functional organization. Chemical provenance follows a proportional mixing convention. Related witness histories are not independent replicates; this is a finite screening experiment."}
    args.output.parent.mkdir(parents=True,exist_ok=True);args.output.write_text(json.dumps(out,indent=2)+"\n",encoding="utf-8")
    print(json.dumps(aggregate,indent=2));print(json.dumps(chem,indent=2))
if __name__=="__main__":main()
