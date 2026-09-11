# Cell contributions: offspring provisioning without demonstrated division of labor

**The preserved mixed-genome groups do not yet demonstrate useful division of labor.** In the tested continuations, all observed energy transfer involving their tagged lineages went from a copying parent to its own copied offspring. No transfer crossed between the original cell lineages, and no targeted signal operation was executed. Bonds mattered to persistence, but their movement constraints are a competing explanation.

## Experiment

Fourteen full qualification snapshots were taken from the preceding [clonal-group experiment](../multicell-motion/REPORT.md): one from environment seed 5, twelve from symbols seed 3, and one from symbols seed 6. They contain 46 original-member occurrences but only 24 distinct original particle identities across three source worlds. Several qualification states overlap in membership and history. **They are related witnesses, not 14 independent evolutionary replicates.**

Each snapshot was continued for 5,000 ticks. Five arms were shared across all witnesses: intact, no peer sharing, no tagged sharing, no signals and no internal bonds. An additional acquisition-disabled arm was run for each original member's COPY lineage. This gives **116 trials**, using **16 workers**, completed in **41.81 seconds** including endpoint output and ordinary control replay. Per-cell statistics and connectivity were accumulated every tick; frames were saved every 100 ticks.

The verified inputs are `data/multicell-roles-final`. Earlier `multicell-roles-v1` and `multicell-roles-verified` runs were implementation pilots and are excluded. [Evidence and matched comparisons](evidence.json) record the final manifest, hashes, baseline cell contributions, source programs and all trial comparisons.

## Direct evidence

- All 14 intact endpoints exactly match ordinary Go replay without the new observer or intervention hook.
- Intact tagged lineages performed **392 successful COPY operations**. These are cell-copy events, not group births.
- They transferred **17,472 energy units** in total. Every unit was attributed to an actual COPY parent supplying its own copied child during the assay. No energy transfer crossed between two original cell lineages.
- `no-peer-sharing` produced **zero suppressed operations and 14/14 identical endpoint hashes**. Its lack of effect reflects an unexercised interaction channel, not demonstrated robustness to losing useful cooperation.
- `no-signals` likewise produced **zero suppressed/blinded operations and 14/14 identical endpoints**. This covers EMIT, TOKEN, LISTEN and signal-channel SENSE for tagged cells. It does not test chemical or terrain-mediated coordination.
- The source groups contained **eight genome programs**. All had instructions for both baseline chemical reactions. One truncated variant lacked COPY; another replaced its reproduction-branch JUMP with TARGET. These are concrete alternatives to interpreting genetic differences as complementary roles. Code presence alone does not establish executed metabolic specialization.

## Matched effects

The first five rows below compare the same 14 witnesses. Totals deliberately retain the related observations; they describe this assay rather than independent replicate estimates.

| Treatment | Changed full endpoints | Tagged COPY events | Tagged cell-ticks | All original members connected: ticks |
|---|---:|---:|---:|---:|
| Intact | — | 392 | 227,608 | 7,766 |
| No peer sharing | 0/14 | 392 | 227,608 | 7,766 |
| No signals | 0/14 | 392 | 227,608 | 7,766 |
| No tagged sharing, including copied children | 13/14 | 769 | 257,386 | 18,936 |
| No internal bonds | 14/14 | 150 | 87,718 | 0 |

Preventing tagged sharing retains energy in copying parents. In these windows, aggregate copies and tagged cell-ticks increased. The intervention also removes post-COPY child provisioning and changes later competition; it is not a clean test of cooperation between different original members. It reduced the exploratory stable founder-free-component counter from 462 to zero component-ticks, illustrating why more copies alone cannot establish organizational progress.

Removing bonds reduced tagged cell-ticks by **61.5%** and copies by **61.7%**. This establishes a consequential bond treatment in these worlds. It does not separate adhesion from preventing MOVE, nor demonstrate coordinated function. The original-members-together statistic necessarily collapses after their links are removed and is not an independent success criterion for this treatment.

The 46 acquisition-disabled arms suppressed 393 paid acquisition operations. Original-cell survival outside the disabled lineage decreased in 39 comparisons and increased in seven. These differences include altered local chemistry, competition, deaths and diverging execution/RNG histories. They cannot be interpreted as 39 demonstrations of partner dependence. The synthetic positive-control test does detect a deliberately constructed energy donor/recipient dependency; that validates the tool, not the evolved witnesses.

Original whole-group continuity in intact runs ranged from **28 to 3,250 additional ticks**. Several new mutant participants were short-lived. Persistent related cell lineages and genetic variation are established observations; repeating reproduction of a useful multicellular organization remains open.

## Verification and reproducibility

The independent Python audit checks complete matched arms, source and endpoint hashes, reconstructed treatment-start hashes, final resource/chemistry conservation and local bonds, acquisition identities, flow bounds, source programs and protocol metadata. Every endpoint also passed Go world validation. Python does not replay every intermediate COPY ancestry event or every connectivity boundary.

Go tests cover a known donor-dependent pair, peer versus own-offspring targeting, actual COPY tagging, lineage persistence after founder death, unchanged ordinary controls, deterministic protocol replay, selective sensory blinding, starvation and instruction costs, skipped effect RNG, invalid members, altered input hashes, new output directories, and equality with one versus 16 workers. `go test ./...`, `go vet ./...` and `go build ./...` pass.

```powershell
go run ./cmd/role-assay -input data/multicell-mixed-witnesses -out data/my-role-assay -workers 16 -ticks 5000 -every 100
python experiments/multicell-roles/analyze.py --input data/my-role-assay
```

Each result includes an external intervention protocol. Saved endpoints are explicitly named **released snapshots**: resuming one with `sim` runs ordinary physics. Repeating a treatment requires replay from its original witness with the recorded protocol. [Full protocol and measurement definitions](../../docs/multicell-roles.md).

## Research decision

Keep multicellular organization as the target. Retain the targeted assay as a reusable evaluation tool. Do not promote these witnesses on genome diversity, copying rate or bond persistence alone.

Next, distinguish the mobility benefit of bonds from functional integration, measure chemical contributions at the reaction and spatial levels, and test candidate complementary functions with matched local interventions. Only then use repeated daughter generations to assess whether that organization is transmitted. No new physical mechanism is adopted by this experiment; the default solver remains unchanged.
