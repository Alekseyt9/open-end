# Stage 14: predictive descriptions and local bond interventions

Local bonds affected survival, but the group-level predictor did not outperform the particle predictor. This experiment provides evidence for a specified local mechanism, **not a positive causal-emergence result or a new level of individuality**.

## Design and execution

The targets were the **83 persistent bond groups** selected at tick 120,000 in the intact Stage 13 worlds. Six of eight seeds contained eligible groups; seeds 3 and 8 supplied no forecast targets. Selection used only prior discovery evidence. The original source snapshots and discovery-file hashes were verified and preserved.

Each intact source continued for five 1,000-tick forecast intervals. Three random partitions reassigned the same target-particle pool into groups with the same initial sizes. This yielded **1,508 eligible group/interval samples**, including **368 real-group samples**. All observations from an evaluation seed were excluded from fitting. Each seed receives equal weight in the reported mean loss. The ridge coefficient was fixed at 0.01; there was no hyperparameter search on evaluation outcomes.

Each real group also received a one-time local cut of all its internal bonds and, where available, a cut of an equal number of outside bonds. Both began from the unchanged original source, with intact first-horizon outcomes as controls. Ordinary BIND remained enabled. **80 outside controls** were available; three were explicitly unavailable because too few outside bonds existed.

The final verified batch took **8.6399935 seconds on 16 workers**. It recorded 174 tasks: eight baseline continuations, 83 local-cut tasks, and 83 outside-cut tasks. Three unavailable outside cuts performed no simulation, so 171 continuations executed, totaling 203,000 world ticks. Time starts after input validation and includes continuation, feature recording, fitting, and artifact writing; Go compilation and initial validation are excluded.

## Prediction on unseen seeds

The target is the fraction of currently alive original group members surviving the next 1,000 ticks. New descendants are excluded. MSE is the primary squared-error criterion; MAE is shown as an alternative loss.

| Model | Seed-mean MSE | Seed-mean MAE |
|---|---:|---:|
| persistence | 0.017288 | 0.033572 |
| constant | 0.016659 | 0.062564 |
| micro | 0.012424 | 0.051180 |
| macro | 0.013069 | 0.051554 |
| micro-context | 0.012203 | 0.051010 |

The macro MSE is about **5.2% higher** than micro MSE. Adding group context to individual predictions reduces micro MSE by about **1.8%**, a small descriptive difference across these six eligible seeds. The macro model improves over a learned constant but does not surpass the local particle model. There is no demonstrated predictive advantage of the coarser representation.

The always-survive predictor has the lowest MAE, despite worse MSE. Survival is common, so conclusions about the best predictor depend on the loss. Neither the small context gain nor any single fold is treated as a statistically established general advantage. Repeated groups and intervals within a seed are dependent.

These model families are deliberately limited: micro uses 12 local features, macro uses 19 aggregate/context features, and micro-context uses 31. The micro predictor does not observe the complete program, memory, or RNG state. Feature counts, objectives, and fixed regularization are explicit in [the protocol](../../docs/causal.md); this is not a comparison with an optimal predictor of the full microscopic state.

## Random grouping controls

| Partition | Micro MSE | Macro MSE | Micro-context MSE |
|---|---:|---:|---:|
| bonds | 0.012424 | 0.013069 | 0.012203 |
| random-1 | 0.011051 | 0.012849 | 0.011504 |
| random-2 | 0.009491 | 0.011090 | 0.010304 |
| random-3 | 0.008508 | 0.009722 | 0.008902 |

The macro model is also worse than the particle model on all three random partitions. Lower absolute errors for some randomized targets are not evidence of better entity boundaries: regrouping changes the target survival fractions and their variance. The controls preserve initial member marginals and group sizes, but eligibility and sizes can diverge after deaths. Small pools can yield unchanged groups; exact counts are retained in the evidence.

## Local causal effects

Effects below are differences in surviving-original-member fraction, in percentage points. Means are taken over target groups within each seed. Outside means use only available matched controls and therefore may cover fewer groups.

| Seed | Groups | Local cut harmed | Local cut benefited | Outside controls | Local − intact, pp | Outside − intact, pp |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 2 | 0 | 0 | 1 | 0.00 | 0.00 |
| 2 | 45 | 15 | 0 | 45 | -17.35 | -0.56 |
| 4 | 1 | 0 | 0 | 0 | 0.00 | unavailable |
| 5 | 7 | 5 | 0 | 7 | -42.38 | 0.00 |
| 6 | 3 | 2 | 0 | 2 | -33.33 | 0.00 |
| 7 | 25 | 12 | 1 | 25 | -27.47 | 1.33 |

The local cut reduced survival in **34/83 groups**, increased it in **1/83**, and left it unchanged in **48/83**. Effects vary strongly across seeds. Mean local-minus-intact survival, averaging the six seed means equally, is **-20.09 percentage points**. This is descriptive for the selected groups and horizon; it is not a population-level estimate from 83 independent replicates.

The outside cuts generally produced smaller target-survival changes. They match the number of removed links but not location, endpoint programs, or local resource conditions. Their results support a location-specific role for internal bonds under this pulse intervention; they do not isolate every possible mechanism. Rebinding remains possible, and outcomes are conditional on the original evolved worlds.

The intervention result and forecast result are compatible: a mechanism can affect survival without its aggregate description becoming a superior predictive model. Collective reproduction, causal autonomy, and higher-level individuality are not established by this experiment.

## Reproduction and verification

```powershell
go run ./cmd/causal -input data/entities-stage13-verified -snapshots data/collectives-stage12 -out data/causal-stage14-repeat -workers 16 -horizon 1000 -blocks 5 -permutations 3 -ridge 0.01
```

The complete output is in `data/causal-stage14-verified`, excluded from Git. [Compact evidence](evidence.json) preserves provenance, scores and folds, per-seed effects, target memberships, and all 83 matched intervention records. Source snapshots plus listed relation removals reproduce cut initial states; full final cut snapshots are not stored.

Independent verification checked every artifact hash, all cut initial-state hashes reconstructed from the preserved source, matched cut counts and endpoint membership, random-partition marginals, and all survival contrasts. NumPy independently refitted every held-out predictor and recomputed every prediction and loss; maximum prediction difference from Go was 2.97e-13. Software tests also verify exclusion of held-out labels, neutral baseline observation, unavailable controls, source preservation, ridge fitting, and identical output hashes with one and 16 workers. The full Go tests, vet, and build pass.

## Following work

Stage 14's analysis harness is implemented, with no positive causal-emergence finding in this configuration. More mechanisms, horizons, and predeclared predictive representations could change that result. Stage 15 explores a symbolic layer; existing signals do not yet establish symbols or context-sensitive meaning. Interface work remains deferred.
