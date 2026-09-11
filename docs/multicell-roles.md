# Cell contributions in clonal groups

`cmd/role-assay` continues preserved mixed-genome witnesses with targeted interventions. It measures what each original cell and its actual-COPY descendants contribute, then compares their survival, copying and energy flows. The [first experiment](../experiments/multicell-roles/REPORT.md) contains 116 continuations from 14 related qualification states.

```powershell
go run ./cmd/role-assay -input data/multicell-mixed-witnesses -out data/my-role-assay -workers 16 -ticks 5000 -every 100
python experiments/multicell-roles/analyze.py --input data/my-role-assay
```

Input must be a completed witness directory containing `manifest.json`, `witnesses.json` and full snapshots, as produced by [the witness replay script](../experiments/multicell-motion/witnesses.py). The command verifies hashes, qualification tick, seed, exact connected membership and genome composition. Output directories must be new. All intact endpoints are also computed by ordinary, unobserved Go replay and must match exactly.

## Interventions

Every original group member starts a separate diagnostic cell lineage. A successful COPY propagates its tag; ALLOCATE alone does not. Existing genome identities never determine targeting, so a mutated descendant retains the intervention of its copying ancestor. Lineages can persist after original cells die. Tags do not label unrelated surrounding cells.

| Mode | Targeted effect |
|---|---|
| `intact` | Observe ordinary physics. |
| `no-peer-sharing` | Suppress TRANSFER between two different tagged cell lineages. Preserve transfers to the actor's own lineage and to outsiders. |
| `no-sharing` | Suppress TRANSFER between any two tagged cells, including copied offspring. The initial 12-unit ALLOCATE reserve is retained. |
| `no-signals` | Suppress tagged EMIT and TOKEN effects. For tagged SENSE of signal channels and LISTEN, write the ordinary result of an empty input, including zero word metadata. Other sensing and LOOKUP remain ordinary. |
| `no-bonds` | Remove initial bonds within the selected group and suppress later BIND between tagged cells. Bonds to outsiders, UNBIND and ordinary movement remain available. |
| `no-acquisition` | One arm per original cell: suppress ABSORB and CONVERT for that cell and its COPY lineage. Transfers, taking energy, allocation and other instructions remain available. |

The experimental hook runs **after ordinary instruction cost and instruction-pointer advancement**, before the selected physical effect. Starved instructions never reach the hook. A suppressed effect does not consume its extra resource costs or its RNG draws. Subsequent RNG trajectories can diverge after interventions change movement or reproduction. These are deterministic matched counterfactuals, not isolated estimates of a single cell's causal contribution independent of its environment.

The explicit `kernel.StepWithIntervention` API separates treatment callbacks from observation. `Step` and `StepObserved` retain ordinary behavior. Neither callbacks nor lineage tags are serialized. Each result records protocol version 1, source/initial/final hashes, members, mode, donor, horizon and sampling interval. The manifest also records the executable hash.

**Endpoint snapshots are released physical states.** Loading a `*-released.json` snapshot with `sim` resumes ordinary physics, without the intervention or its tags. To repeat a treatment, start from the original witness and replay the recorded protocol. The assay does not claim that an endpoint snapshot alone resumes its experimental conditions.

## Measurements

For each original cell lineage the result records its initial genome, successful copies and allocations, acquired energy split into absorption and conversion, paid opcode counts, suppressed/blinded opcode counts, surviving tagged cells, tagged cell-ticks, and original-cell survival ticks. A paid opcode is not necessarily a successful physical action; suppression counts include attempts whose ordinary effect could have been zero. Acquisition excludes transferred or taken energy, which have their own flow records.

TRANSFER and TAKE events record resource-flow direction between original-cell lineages. Founder ID zero denotes an untagged participant; an uncoded target is additionally identified in the event kind. Flow records distinguish direct bonds and **observed COPY-parent-to-child** energy. The latter requires a COPY event within this assay, rather than assuming allocation parenthood equals genome parenthood. The reported initial genome identifies a lineage's starting cell, not every mutated descendant's current genotype.

Every tick, connectivity is computed on the full bond graph. Original-member connected-pair ticks and ticks with all original members alive and connected are accumulated. A component can include extra cells: this metric does not require unchanged membership. It can also contain outsiders, so it does not assert exclusive internal cohesion.

A separate exploratory counter records component-ticks after an exact founder-free, fully tagged connected member set has persisted for 100 ticks. Components containing outsiders or an original selected cell do not qualify. This counter **does not require a surviving parent or transmitted organization** and is not the clonal-daughter criterion of the preceding experiment. Growing member sets restart the age; it must not be interpreted as a birth count or a multicellular-life-cycle score.

## Interpretation and next tests

Mixed genomes, different copying rates, effects of bond removal and effects of disabling a cell's metabolism each have multiple explanations. In particular, removing bonds changes mobility; disabling acquisition changes local chemical processing, competition, death and later resolution history. A decline in other cells after that intervention is not by itself proof of useful specialization.

The strongest direct checks in this assay concern energy exchange and the implemented signal channels. They do not exhaust coordination through chemical intermediates, terrain or spatial exclusion. Future work should separate the movement consequences of bond loss from adhesion effects, resolve metabolic activity by reaction and location, and test whether complementary functions recur across successive group generations. It should not rank worlds by copying count or group size alone.
