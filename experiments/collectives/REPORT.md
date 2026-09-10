# Proto-multicellularity: first matched ablation experiment

The observation and intervention machinery is implemented. Bonds strongly improved survival of the same initial group members in several tested worlds. No daughter-group candidate met the declared criterion. Whole-group reproduction and the Stage 12 research criterion remain unproven.

## Design

Eight unchanged `environment` snapshots from Stage 11 at tick 100,000 were each continued for 20,000 ticks under four treatments: intact, no bonds, no directly bonded resource sharing, and no signal reading. Seeds were 1–8. Worlds retain their terrain, chemistry, evolving copying, particles, RNG states, and ancestry. Only the declared treatment differs at initialization. The source snapshots are preserved.

The assay ran 32 independent worlds with 16 workers and GOMAXPROCS=16. Recorded assay wall time was **24.943739 seconds**, measured after source loading/validation; it includes simulation, group tracking, and output generation, and excludes Go compilation. This is one measured batch, not a benchmark distribution. Reporting occurs every 1,000 ticks; group continuity is observed at every tick boundary. Minimum unchanged membership age is 100 ticks. Both historical tracking caps are 1,024; no cohort or candidate was skipped.

Reference groups are fully programmed components with at least two explicitly bonded particles in the unmodified source. All treatments follow exactly the same reference member IDs, including the treatment where their bonds are removed. These references contain **126 groups and 1,022 members**. Cohort ancestry starts only after a group survives 100 observed ticks; it follows actual successful COPY actors, not allocation parents or genome similarity.

See [the protocol](../../docs/collectives.md) for exact exclusions and candidate criteria. No collective entity class, special reproductive operation, or fitness reward was added. The seed program remains unchanged.

## Results

The table compares the same reference individuals in each treatment. Copy counts cover the entire world, not just reference groups. A zero-member source supplies no test of initial group-member survival.

| Seed | Initial reference members | Alive: intact | Alive: no bonds | Alive: no sharing | World copies: intact | World copies: no bonds | World copies: no sharing | Bonded energy transferred: intact |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 5 | 5 | 5 | 5 | 30666 | 30295 | 30666 | 0 |
| 2 | 472 | 343 | 5 | 370 | 21157 | 21558 | 26884 | 216926 |
| 3 | 0 | 0 | 0 | 0 | 20193 | 20193 | 20193 | 0 |
| 4 | 2 | 2 | 0 | 2 | 32387 | 32303 | 32387 | 0 |
| 5 | 445 | 364 | 1 | 364 | 26734 | 23194 | 26734 | 0 |
| 6 | 11 | 6 | 0 | 9 | 16538 | 16603 | 16577 | 4080 |
| 7 | 87 | 27 | 6 | 30 | 24027 | 24937 | 23958 | 48 |
| 8 | 0 | 0 | 0 | 0 | 16039 | 16039 | 16039 | 0 |

Across the eight worlds, **747/1,022** reference members survived with intact mechanisms, compared with **17/1,022** without bonds and **780/1,022** without bonded sharing. These are pooled descriptive counts; members within a world are not independent replicates. Five of the six worlds containing reference groups had more surviving reference members with bonds than without them; seed 1 tied. This intervention tests removing bonds in evolved populations. It does not independently compare isolated group and singleton populations in identical habitats or establish why bonds help.

Resource sharing was active in intact seeds 2, 6, and 7. Removing it increased survival of reference members in all three. For example, seed 2 produced 21,157 world copies intact versus 26,884 without sharing. Transfer activity therefore cannot be treated as evidence of mutual benefit. Removing bonds reduced world copying in three seeds, increased it in three, and left two unchanged; reference survival and total copying answer different questions.

All eight signal-reading treatments ended with exactly the same complete World state as their intact counterparts after removing the recorded treatment configuration. The comparison includes RNGs, particles, fields, ancestry, and counters. No effect of signal information was detected in this continuation window; this does not prove signals could never matter. Seeds 3 and 8 also matched intact under both other treatments after removing treatment configuration.

**All 32 trials recorded zero daughter-group candidates and zero productive daughter-group candidates.** Linked individuals did copy successfully, and some tagged cohorts had living descendants. That alone does not meet the criterion: new linked descendants must represent at least two distinct founding particles, persist separately, and coexist with a surviving parental component. The detector can miss other collective life cycles and groups outside its conservative tagging policy. No claim of obligatory specialization, group heredity, information closure, or multicellular individuality follows from these results.

## Reproduction and evidence

```powershell
go run ./cmd/group-assay -input data/environment-stage11 -case environment -out data/collectives-stage12-repeat -ticks 20000 -every 1000 -workers 16 -group-age 100
go run ./cmd/summarize -input data/collectives-stage12-repeat -window 20000 -format json
go run ./cmd/council prepare -input data/collectives-stage12-repeat -out data/council-stage12-repeat -window 20000
```

The original run is in `data/collectives-stage12`; its dossier is in `data/council-stage12`. Run directories are ignored by Git. [Compact evidence](evidence.json) retains source/initial/final snapshot hashes, metrics-file hashes, the measured manifest, per-trial counts, and physical equality checks. Every saved final snapshot hash was verified, and recomputing summaries from JSONL reproduced all recorded collective summaries.

Validation covers observer neutrality, deterministic replay of format 6, conservation and instruction costs under ablation, unbonded child provisioning, signal-channel isolation, matched reference membership, synthetic positive/negative daughter events, frame ownership, telemetry validation, and one-worker versus 16-worker equivalence. The existing council continuation harness preserves the observation age and ablation configuration in a fresh session. The Go test suite, vet, build, and nine Warp differential/rejection tests pass. Warp explicitly rejects these unsupported treatments.

## Next research step

Stage 13 can use these observations to propose entity boundaries and compare linked groups with resource-flow and ancestry groupings. Persistence, coordinated activity, and reproduction should remain separate descriptors. Controlled separation and mechanism-specific interventions are needed before assigning a higher-level individual. Stage 14 then asks whether such macro descriptions improve prediction under interventions. Interface work is deferred.
