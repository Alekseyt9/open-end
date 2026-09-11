# Bond benefits explained by immobility; limited chemical coupling between generalists

**Holding movement fixed removes the measured non-bond-state effect of removing internal bonds in all 14 tested witnesses.** A passive chemical tracer also detects consumption of products from other cell lineages, but every observed original cell executes both reactions. These findings distinguish persistence and material coupling from useful division of labor.

## Protocol and scope

The same 14 mixed-genome qualification snapshots from three source worlds were continued for 5,000 ticks: environment seed 5, symbols seed 3 and symbols seed 6. Several snapshots overlap in particles and history: there are 46 original-member occurrences but only 24 distinct original particle identities. They are related observations, not 14 independent evolutionary replicates.

Four mobility arms were run per snapshot, plus two reaction-disabled arms per original member. **148 trials on 16 workers completed in 68.66 seconds**, including output and ordinary control replay. All sources use baseline ecology, without DSL rules. Frames were saved every 100 ticks, while lineage, chemistry and connectivity observations ran every tick.

The recorded batch is `data/multicell-mechanisms-v2`. [Evidence](evidence.json) includes the manifest, input/result hashes, chemical provenance matrices, mobility projections and all matched trial summaries. [Mechanics and reproduction instructions](../../docs/multicell-mechanisms.md).

## What bonds accomplish in these worlds

| Treatment | COPY events by tagged lineages | Tagged cell-ticks | Original-cell survival ticks |
|---|---:|---:|---:|
| Intact | 392 | 227,608 | 120,973 |
| Internal bonds removed | 150 | 87,718 | 32,986 |
| All tagged movement suppressed; bonds retained | 446 | 257,551 | 117,672 |
| All tagged movement suppressed; internal bonds removed | 446 | 257,551 | 117,672 |

For **14/14 anchored pairs**, endpoint states are identical after removing only the relation graph and genome BIND counters from the comparison. All resources, particles, positions, executable programs, ancestry, both RNG states and other counters remain included. Traced chemical consumption also agrees. Full snapshot hashes differ because the relations themselves differ.

Thus the earlier bond-removal penalty does not establish an additional benefit from collective coordination. In these conditions, restricting movement accounts for the measured difference. Anchoring all tagged descendants is stronger than ordinary bond immobility, so the anchored arms are controls for each other rather than substitutes for the intact world. This conclusion is limited to these snapshots, mechanics and horizon.

## Chemical contributions

In intact continuations, tagged lineages executed **60,141 units of reaction 0** and **60,014 units of reaction 1**. The proportional Y tracer attributes the latter consumption as follows:

| Most recent Y producer | Y units consumed by tagged cells | Share |
|---|---:|---:|
| Same original-cell COPY lineage | 55,661.27 | 92.75% |
| Another original-cell COPY lineage from the selected witness | 783.49 | 1.31% |
| Untagged cells outside these lineages | 3,552.07 | 5.92% |
| Initial Y of unknown provenance | 17.17 | 0.03% |

Percentages are rounded independently. These totals include related witnesses and must not be interpreted as independent sample estimates. The second category means different lineages selected at the start; it does not guarantee that producer and consumer still belong to one connected group when consumption occurs.

All 14 intact witnesses show positive traced consumption from another selected lineage. However, **all 46 original-cell observations execute both reactions**. Their reaction-0 fraction ranges from 0.375 to 0.536; the more extreme values occur in short-lived cells with small reaction counts. This is compatible with generalist metabolism using a shared chemical field, without establishing complementary producer/consumer roles.

The tracer distributes labels proportionally when Y mixes and is consumed. Its fractional units are accounting estimates under that well-mixed convention, not observed identities of individual molecules. The previous report's absence of cross-lineage TRANSFER instructions remains correct; indirect chemical coupling is a different pathway that the new observer measures.

## Reaction-specific disruptions

Each of the 46 original-member occurrences was tested once with its lineage's reaction 0 disabled and once with reaction 1 disabled. Normal instruction costs remain; other metabolism and interactions are available.

| Disabled reaction in one lineage | Arms | Other original cells' summed survival decreases / increases / unchanged |
|---|---:|---|
| Reaction 0: X to Y | 46 | 38 / 7 / 1 |
| Reaction 1: Y to Z | 46 | 39 / 7 / 0 |

These are exploratory whole-continuation effects. Disabling either half of a generalist's metabolism changes its survival, local chemistry, competition, reproduction and later RNG trajectory. The many negative partner-survival differences are not separate demonstrations of useful specialization. Reaction-0 provenance from the disabled lineage is zero, as required; this validates the scope of the intervention without proving a group-level benefit.

## Verification

- All 14 intact endpoints match ordinary unobserved Go replay.
- All 28 intact and no-bonds endpoints match the corresponding saved endpoints of the preceding role assay exactly.
- All 148 endpoints pass world validation, physical resource/chemistry accounting and bond-locality checks.
- All tracer producer balances and reaction/acquisition identities pass. The largest endpoint cell-mass error is `2.22e-16` Y units; the largest producer-balance error is `5.82e-10` Y units.
- Tests detect a deliberately constructed chemical producer/consumer dependency, check proportional mixing and unknown initial Y, distinguish anchoring from bonds, verify actual resolved chemistry callbacks, and compare outputs with one versus 16 workers. The positive control validates measurement; it is not an evolved result.
- `go test ./...`, `go vet ./...` and `go build ./...` pass.

The Python audit verifies artifacts, endpoint projections and accounting identities; it does not independently replay every intermediate transport event or lineage tag. The full external protocol and tracer state are not embedded in endpoint snapshots. Saved `*-released.json` files resume ordinary physics; rerunning an intervention requires its original witness and protocol.

```powershell
go run ./cmd/role-assay -suite mechanisms -input data/multicell-mixed-witnesses -out data/my-mechanisms -workers 16 -ticks 5000 -every 100
python experiments/multicell-mechanisms/analyze.py --input data/my-mechanisms
```

## Next experiment

The next hypothesis is that cheap alternation of both metabolic reactions favors generalists and gives little advantage to division of labor. Prepare an optional, resource-funded reaction-switching cost with a zero-cost control, then compare evolving worlds for persistent specialization, useful cross-lineage chemical contributions and reproduction transmitting those functions. Do not assign cell types or reward group size directly. This hypothesis is motivated by the measured reaction profiles; it has not yet been tested, and no new physical mechanism is adopted here.
