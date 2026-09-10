# Stage 9: novelty archive and Pareto selection

Date: 2026-09-10. Stage 9 was validated on the existing Stage 8 tree and on controlled selection fixtures. The implementation provides an immutable archive, a frontier over measured objectives, retention of behavior cells, and explicit application of a multi-branch continuation set.

## Archive → continue → refresh

The starting tree contained five cohorts and 80 stored world states. Its two tips at tick 140,000 were b000004 (baseline direction) and b000005 (the earlier solar-y-recycle direction).

The first archive decision recommended **both b000004 and b000005**. The recommendation was applied, then both cohorts were continued for another 20,000 ticks using 16 workers. This created 32 new world continuations in **19.01 seconds**, including trial execution and child dossier publication; archive preparation was outside that recorded run duration.

A refreshed archive recommended **both b000006 and b000007**, now at tick 160,000. The resulting tree has **7 cohorts, 112 preserved world states, and three generations**. The original five nodes and all 80 earlier physical hashes are unchanged.

The first immutable decision remains available alongside the second. Repeating an archive refresh on identical evidence and settings reuses the same decision rather than creating another revision.

## Final measured objectives

Window: ticks 150,000–160,000. Each cohort contains the same 16 worlds: ecology and ecology-no-mutation, seeds 1–8. Reported diversity and structural scale are means across all 16 worlds, including the no-mutation controls.

| Cohort | Direction | Novelty distance | Effective diversity | Bonded scale proxy | Productive block fraction | Pareto tip |
|---|---|---:|---:|---:|---:|---|
| b000006 | Baseline continuation | 0.116410 | 4.933881 | 0.548613 | 1.000 | Yes |
| b000007 | Proposal continuation | 0.185965 | 3.934983 | 0.423197 | 1.000 | Yes |

The baseline direction has higher diversity and structural scale. The proposal direction is farther from its compatible archive references under the chosen common-action descriptor. Neither dominates the other on all objectives, so both remain.

This does not reverse the earlier conclusion that the proposal had lower diversity and copying. Descriptive novelty is a separate research objective; it does not establish improvement, adaptation, or open-ended evolution.

Both current tips cover the two viable behavior cells in their comparison context. Historical representatives remain preserved. The older root uses a different 50,000-tick analysis window and is not Pareto-ranked against the 10,000-tick tips.

## Controlled checks

The tests include a rare-cell candidate with lower diversity, structural scale, and novelty than another comparable tip. It is excluded from the Pareto front but retained because its behavior cell would otherwise disappear from the recommended set. Another weaker candidate from that already covered cell is not recommended. This establishes the Stage 9 retention behavior independently of the real run's tradeoff.

Other checks cover:

- equality and tradeoffs under strict multi-objective dominance;
- deterministic results after permuting candidate order;
- zero novelty for an identical behavior descriptor;
- deduplicated novelty references, preventing repeated trajectories from changing reference density;
- reactivation of a historical cell representative when no current tip covers it;
- separation of incompatible contexts and final horizons;
- rejection of insufficient stable history and invalid policy settings;
- an empty recommendation preserving the previous saved selection;
- read-only preview versus explicit application;
- unchanged physical snapshots after archive operations;
- immutable previous decisions, idempotent refresh, and detection of edited decisions;
- stale archive detection after tree growth.

All Go tests, `go vet ./...`, and `go build ./...` passed.

## Interface validation

The offline HTML explorer shows seven archive records, a matched-horizon chart, all objective values, retention reasons, and novelty-reference IDs. Browser checks covered:

- loading recommendations into the continuation command builder;
- changing chart axes and inspecting point evidence;
- generation of custom archive and apply commands;
- rejection of invalid settings in the command builder;
- stale recommendations disabled after new tree evidence appears;
- desktop and narrow viewport layouts, without page JavaScript errors.

The UI prepares commands; it does not run simulations automatically.

## Artifacts and reproduction

- [decision-before.json](decision-before.json): first five-cohort archive decision.
- [decision-after.json](decision-after.json): refreshed seven-cohort decision and references.
- [continuation.json](continuation.json): the 16-worker continuation parameters and duration.
- [tree-compact.json](tree-compact.json): final ancestry, selection, metrics, and physical hashes.

Full snapshots, dossiers, JSONL, and immutable archive revisions remain locally in `data/branching-stage8/`. These compact artifacts alone cannot resume a simulation.

Starting from a reproduced [Stage 8 tree](../branching/REPORT.md):

```powershell
go run ./cmd/council tree archive -tree data/branching-repeat -apply
go run ./cmd/council tree grow -tree data/branching-repeat -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council tree archive -tree data/branching-repeat -apply
go run ./cmd/council tree export -tree data/branching-repeat
```

Causal structure, hierarchy depth, new information processing, and adaptive value remain unavailable. The implemented policy uses the measured proxies described in the [archive reference](../../docs/archive.md).
