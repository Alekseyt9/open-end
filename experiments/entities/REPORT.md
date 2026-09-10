# Automatic entity discovery: verified replay of Stage 12

The detector found persistent physical boundaries and nested copying-ancestry sets. It found **no reciprocal-transfer candidate** in any observed interval and did not establish higher-level individuality. The containment hierarchy is descriptive; it does not demonstrate a collective life cycle or causal coordination.

## Experiment

All **32 Stage 12 continuations** were replayed from their original Stage 11 snapshots: eight seeds under intact, bond removal, bonded-sharing removal, and signal-reading removal. Each continuation covers ticks 100,000–120,000, with discovery every 1,000 ticks and minimum boundary age 100. The per-interval edge and actor limits were 65,536 each.

All **32 final physical hashes matched** the original assay, and source snapshots were preserved. The final verified batch took **24.7829294 seconds** with 16 workers and GOMAXPROCS=16. Timing begins after source/final-snapshot validation and includes replay, discovery, and report writing; it excludes Go compilation and initial validation. Earlier development runs are not pooled into a benchmark estimate.

There are 21 checkpoints per world, including the initial state, and 20 nonoverlapping discovery intervals. The last checkpoint has no subsequent interval. All activity intervals are complete; no records hit the activity limits. Candidate definitions and limitations are in [the protocol](../../docs/entities.md).

## Intact worlds

Counts refer to the final checkpoint, except maximum depth, which covers the whole replay. Boundary hypotheses are deduplicated when they contain exactly the same particle IDs. Ancestry-only candidates are copied families without an identical bond or reciprocal-flow boundary; they may be spatially scattered. Nested candidates contain a proper candidate subset.

| Seed | Final candidates | Persistent bond boundaries | Ancestry-only candidates | Nested candidates at end | Maximum structural depth |
|---:|---:|---:|---:|---:|---:|
| 1 | 72 | 2 | 70 | 0 | 1 |
| 2 | 156 | 45 | 77 | 38 | 2 |
| 3 | 84 | 0 | 84 | 0 | 1 |
| 4 | 58 | 1 | 57 | 1 | 2 |
| 5 | 84 | 7 | 61 | 14 | 2 |
| 6 | 88 | 3 | 84 | 1 | 2 |
| 7 | 89 | 25 | 62 | 8 | 2 |
| 8 | 64 | 0 | 64 | 0 | 1 |

Across intact worlds there are **695 final candidate sets**, including **83 persistent bond boundaries**. Removing bonds leaves 468 final candidate sets, all ancestry-only, with maximum depth 1. Thus many detected sets are simply copying families; their existence cannot be used as evidence for a new organism.

Maximum observed depth is 2 in intact seeds 2, 4, 5, 6, and 7. Particle leaves are depth 0, compound candidate sets are depth 1, and a set containing another candidate can reach depth 2. This is containment between boundary hypotheses, not two proven levels of biological individuality. Noncontained overlaps are preserved instead of forcing one global partition.

For example, at tick 101000 in seed 2, a candidate with 3 members and sources `bonds` contains 1 direct candidate subset(s). The [compact evidence](evidence.json) includes its exact member IDs, subset identities, and activity in the separate following interval. This example illustrates the inclusion algorithm only.

## Resource exchange and reproduction

Across all 32 worlds and all checkpoints, there are **zero reciprocal-transfer components** with two or more surviving coded particles. The earlier experiment did record substantial energy sharing. Here, however, the directed positive TRANSFER graph contains no qualifying return paths within any 1,000-tick interval. One-way provisioning and TAKE are explicitly excluded from evidence of reciprocal transfer.

Because no reciprocal-flow candidate was discovered, the count eligible for subsequent-interval flow validation is zero. That is an absence of eligible candidates, not a measured 0% success rate. Synthetic tests verify positive cycles, prospective persistence, and failure of a return path that leaves the proposed boundary. They are software validation, not evolved phenomena.

Intact and signal-reading arms have identical complete discovery frames for all eight seeds, consistent with their identical physical histories in Stage 12. All arms still have zero qualifying daughter-group candidates. Information closure, obligatory specialization, collective heredity, and causal coordination remain unmeasured or unproven.

The number of final intact bond boundaries with membership age at least 10, 100, and 1,000 ticks is respectively **114**, **83**, and **52**. This threshold check reuses recorded exact bond ages; it does not rerun different physical worlds. Reciprocal-flow persistence has the coarser discovery resolution and cannot be inferred at 100-tick resolution from these reports.

## Reproduce and inspect

```powershell
go run ./cmd/discover -input data/collectives-stage12 -source data/environment-stage11 -out data/entities-stage13-repeat -workers 16 -every 1000 -group-age 100
```

The verified output is in `data/entities-stage13-verified`. Run output is ignored by Git; [evidence.json](evidence.json) retains report hashes, verified physical hashes, per-world summaries, the measured manifest, threshold counts, and the nested example. Reports are analysis JSON, not resumable snapshots or ordinary metrics JSONL. The CLI validates the preserved inputs and requires a new output directory.

Verification covers deterministic report hashes with 1 and 16 workers, observer neutrality across sampling intervals and record caps, all four physical treatments, reciprocal cycles, predation exclusions, dead members, COPY actor ancestry, interval truncation, overlap, and transitive containment reduction. An independent pass checked every emitted member-set hash, micro membership, containment edge, overlap count, hierarchy depth, next-interval boundary, survivor count, and report-file hash.

## Next stage

Stage 14 should compare predictive descriptions at different scales and test proposed boundaries under controlled interventions. Existing boundary candidates supply hypotheses, not a causal-emergence result. Discovery and evaluation data must remain separate, and particle-level or matched alternative groupings are needed as controls. Interface work remains deferred.
