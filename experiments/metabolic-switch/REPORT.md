# Switching costs reduce population and chemical coupling without inducing complementary roles

**The first switching-cost screen does not support adopting the mechanism as a default.** Across the tested costs and 20,000-tick continuations, no living reaction-1 specialists or complementary bonded groups were recorded at any metabolic window boundary. Population fell while copying increased. In the separate witness study, chemical consumption from other selected cell lineages decreased.

## Implementation and protocol

The optional physical mechanism charges energy when a cell changes its prepared reaction type. First preparation is free; descendants begin unprepared. Ordinary instruction costs remain. Preparation persists in **snapshot format 9**, together with the cost and switching-energy ledger. A zero cost preserves legacy state and hashes. [Full semantics and commands](../../docs/metabolic-switch.md).

The population screen uses two existing cohorts at tick 100,000: eight Stage 11 environment worlds and eight Stage 15 symbol worlds. Every source was continued under costs **0, 1, 2 and 4**, for **64 runs of 20,000 ticks**. These were followed by **56 runs of 5,000 ticks** from the fourteen previously preserved mixed-genome witnesses, using the same four costs and chemical provenance measurement.

All batches used **16 workers**. Recorded durations were **39.83 s**, **36.98 s** and **30.53 s**, including observation and output; witness runs additionally include ordinary control replays. These are batch timings, not isolated solver benchmarks.

The sixteen source worlds are matched within their cohorts; overlapping seeds/cohort histories should not be treated as sixteen fully independent evolutionary histories. The fourteen witnesses are particularly dependent: they come from three source worlds and share some members and history. The witness experiment is a short-term matched test of existing groups, separate from the population evolutionary screen.

## Evolutionary screen

The table reports totals across eight worlds in each cohort. Specialist counts use the final 5,000-tick window: at least 32 successful reaction units and at least 90% in one reaction. Persistence means the same living particle met that definition in both final windows.

| Cohort | Cost | Final particles | COPY events over 20,000 ticks | Living reaction-0 specialists | Persistent reaction-0 specialists |
|---|---:|---:|---:|---:|---:|
| Environment | 0 | 5,353 | 187,741 | 7 | 6 |
| Environment | 1 | 4,983 | 212,254 | 3 | 2 |
| Environment | 2 | 4,659 | 215,725 | 1 | 1 |
| Environment | 4 | 4,090 | 243,725 | 3 | 0 |
| Symbols | 0 | 5,349 | 158,408 | 13 | 12 |
| Symbols | 1 | 4,985 | 186,173 | 0 | 0 |
| Symbols | 2 | 4,686 | 197,394 | 1 | 1 |
| Symbols | 4 | 4,095 | 244,813 | 1 | 0 |

All **256 window observations** had zero living reaction-1 specialists and zero complementary bonded groups. This does not exclude brief specialization between boundaries or behavior below the activity threshold. It does establish that this screen did not produce the sought persistent producer/consumer organization by its declared criteria.

At cost 4, combined final population was **23.52% lower** than zero cost, while COPY events were **41.14% higher**. More copying with fewer surviving particles indicates greater replacement/turnover in this comparison, not increased organizational complexity. These counts are cell-copy events, not births of groups.

## Chemical contributions of preserved witnesses

The tracer attributes Y to its latest producer using proportional mixing. The peer column denotes another selected original-cell COPY lineage, not necessarily membership in the same group at consumption time.

| Global switching cost | Peer-produced Y consumed | All Y consumed by tagged lineages | Peer fraction | Tagged cell-ticks | Tagged COPY events |
|---|---:|---:|---:|---:|---:|
| 0 | 783.49 | 60,014 | 1.31% | 227,608 | 392 |
| 1 | 150.17 | 33,976 | 0.44% | 116,937 | 197 |
| 2 | 143.91 | 25,124 | 0.57% | 77,917 | 191 |
| 4 | 124.88 | 47,580 | 0.26% | 118,959 | 489 |

There is no monotonic gain in cell-lineage persistence or chemical cooperation. Absolute peer contribution is below the zero-cost value for every nonzero treatment; the fraction also decreases. Increased copying at cost 4 again does not establish collective improvement. Tracer fractions are accounting under the well-mixed convention, not identities of individually observed molecules.

The treatment changes all cells, including outsiders. These comparisons cannot isolate a selected group's response from the surrounding ecological response. Short-term loss of existing lineages also does not settle whether a different organization could evolve over a much longer horizon.

## Verification and evidence

- All sixteen zero-cost population endpoints exactly match the earlier saved 120,000-tick controls.
- All fourteen zero-cost witness endpoints exactly match the previous mechanisms experiment; all 56 witness endpoints match ordinary replay under the same physical cost.
- All 120 endpoints pass Go world validation and independent resource/chemistry accounting. Switching heat is included in dissipation, including partial payments before switching starvation.
- Window reaction totals match physical reaction ledgers. The Python audit reconstructs classifications and persistence from the saved profiles, checks snapshot/configuration hashes, and verifies witness tracer balances and conversion-energy identities.
- Tests cover preparation costs, no-substrate attempts, invalid operands, newborn reset, starvation, nonzero-state snapshot replay, format downgrade rejection, DSL incompatibility, source immutability, zero-cost equivalence, activity/purity thresholds, interruptions of specialist persistence, and one-versus-16-worker equality. Full tests, vet and build pass.

[Compact evidence](evidence.json) contains manifests, result-file hashes, per-world windows, final living specialist profiles and witness chemical matrices. Full profiles, metrics and snapshots remain in the reproducible ignored output directories: `data/metabolic-switch-environment`, `data/metabolic-switch-symbols`, and `data/metabolic-switch-witnesses`. The independent audit checks reported window classifications but does not reconstruct every intermediate per-cell reaction event from scratch.

```powershell
go run ./cmd/switch-assay -input data/environment-stage11 -case environment -out data/my-switch-environment -workers 16 -ticks 20000 -every 1000 -group-age 100
go run ./cmd/switch-assay -input data/symbols-stage15 -case symbols -out data/my-switch-symbols -workers 16 -ticks 20000 -every 1000 -group-age 100
go run ./cmd/role-assay -suite switching -input data/multicell-mixed-witnesses -out data/my-switch-witnesses -workers 16 -ticks 5000 -every 100
python experiments/metabolic-switch/analyze.py --environment data/my-switch-environment --symbols data/my-switch-symbols --witnesses data/my-switch-witnesses
```

The audit defaults to the three recorded directories named above; its input flags allow alternative output names. Ordinary `sim -load` resumes the stored switching physics; only the external observation protocol must be replayed from its source.

## Decision and next hypothesis

Keep cost zero as the default and retain positive costs as explicit experimental options. The results do not justify escalating the penalty merely to force specialization.

The next hypothesis concerns access to the intermediate product: consecutive local reactions can recapture their own Y before it becomes available to neighbors. Compare transport timing and local resource availability, with matched zero-cost controls, before combining those conditions with a switching cost. This is a proposed explanation to test, not a causal conclusion of the current study. Continue to require useful complementary activity and eventual inheritance of organization rather than more copying, longer genomes or larger stationary clusters.
