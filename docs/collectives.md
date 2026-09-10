# Proto-multicellularity (Stage 12)

Stage 12 adds conservative observation of linked groups and matched interventions. Existing particles already have explicit bonds, local energy transfer, and environmental signals. No organism class, group fitness, special birth action, or predefined division of labor is introduced. Individual programs continue to acquire resources, bind, copy, and die under ordinary rules.

The [first 32-world experiment](../experiments/collectives/REPORT.md) found a substantial survival effect of bonds in several evolved populations, but no qualifying daughter group. The mechanisms and measurement harness are implemented; the research criterion of whole-group reproduction remains open. Interface work is deferred.

A [follow-up on bond motion and clonal ancestry](multicell-motion.md) now records one-founder clonal daughters separately. It found persistent clonal descendant groups excluded by the original multi-founder criterion, including mixed-genome candidates preserved as physical witnesses. This expands observation without changing the historical definition or establishing inherited division of labor.

## Recording a world

```powershell
go run ./cmd/sim -width 32 -height 32 -ecology -copy-model evolving -environment coupled -matter-diffusion 4 -chemical-diffusion 4 -groups -group-age 100 -ticks 20000 -every 1000 -metrics data/groups.jsonl -save data/groups.snapshot.json
go run ./cmd/summarize -input data/groups.jsonl -window 20000 -format json
```

`-groups` requires `-metrics`. `-group-age` must be positive and defaults to 100 ticks. Tracking is opt-in and Go-only. It reads the bond graph after every tick and actual successful COPY and TRANSFER events; it consumes no RNG and changes no physical state. Reporting every 1,000 ticks does not reduce the frequency of continuity observation.

A physical snapshot does not serialize observer history. Resume with `-load ... -groups` to start a new session. Existing council continuations automatically preserve the recorded minimum age and treatment configuration, while resetting the observer session. Histories from different sessions cannot be concatenated as one observation. The archive separates observation protocols by minimum group age as well as its existing configuration and sampling checks.

## Operational definitions

| Measurement | Definition |
|---|---|
| Linked group | A connected component of at least two particles joined by explicit bonds; mere spatial proximity is insufficient |
| Membership age | Time since that exact sorted set of particle IDs was first observed without a missing tick boundary; changes to members reset age |
| Stable group ticks | Number of group/tick observations whose membership age has reached the configured minimum |
| Reference group | A fully programmed linked component in the unchanged initial source, shared across intervention arms |
| Reference members alive | Original reference particle IDs still present, regardless of their current links |
| Reference groups together | Reference groups whose original members are all still in one connected component; additional members are allowed |
| Founder cohort | A fully programmed group reaching the age threshold with no already tagged member; each founder receives a distinct ancestry tag |
| Linked copies | Successful COPY actions whose actor is linked at the time of copying |
| Bonded transfer | Energy actually transferred by TRANSFER between directly bonded particles |

Topology may change without resetting membership age if the same component remains connected at tick boundaries. A disconnect and reconnect wholly within one tick is not observed as a separate membership episode. Reference-member survival and retained reference groups describe a fixed initial set. New groups and their activity are recorded separately.

Founding tags live only in the observer. A successful COPY passes the actor's cohort and original-founder tag to its new programmed target. ALLOCATE parent IDs and shared genome hashes do not establish this ancestry. Descendants remain associated with that cohort; mixed or already tagged groups are not assigned a new cohort. This conservative policy avoids repeatedly counting the same ancestry as new group origins, but can miss later mergers and other forms of collective heredity.

Per-cohort `tagged_lineage_activity` records copies, energy acquired through ABSORB or positive net CONVERT effects, and directly bonded transfers, grouped by the acting genome. The counts include tagged descendants wherever they live. They are activity descriptions, not proof of functional specialization or mutual dependence.

## Daughter-group candidates

A daughter candidate must satisfy all of these conditions at consecutive observed tick boundaries for at least the minimum age:

1. It is a fully programmed linked group containing no original founder particle.
2. Every member descends through successful COPY events from the same founder cohort.
3. At least two distinct original founding particles contribute descendant lineages. These can have identical genomes.
4. A separate surviving parental component retains at least two original founders.

`all_founder_lineages` distinguishes candidates representing every founder from those representing only a subset. The `founder_lineages` count refers to original particle ancestry, not genome diversity.

A productive candidate has performed at least one successful COPY while its exact membership component exists during the eligible episode, including its qualifying period. This confirms individual copying activity within a candidate, not the production of another whole group. The same daughter member-ID set is credited once even if it disappears and rejoins. `since_tick` and `qualified_tick` describe its first qualified episode; `last_observed_tick` is its latest qualifying observation and can follow a gap. Their difference must not be interpreted as uninterrupted lifetime after qualification.

Each session retains at most 1,024 founder cohorts and 1,024 daughter candidates. `skipped_cohorts_or_candidates` exposes attempted records beyond these limits. A zero candidate count is qualified by the age threshold, tagging policy, observation horizon, and any skipped records. It is not a proof that collective reproduction is impossible.

## Matched interventions

```powershell
go run ./cmd/group-assay -input data/environment-stage11 -case environment -out data/group-assay -ticks 20000 -every 1000 -workers 16 -group-age 100
go run ./cmd/summarize -input data/group-assay -window 20000 -format json
go run ./cmd/council prepare -input data/group-assay -out data/group-round -window 20000
```

The input is a completed batch with a summary and validated snapshot hashes. The selected case must have unique seeds, enabled environmental engineering, and no previous collective ablation. Each source is cloned into four treatments:

| Treatment | Physical change |
|---|---|
| `intact` | No configuration or state change |
| `bonds` | Remove initial bonds and prevent subsequent BIND effects |
| `sharing` | Prevent TRANSFER effects only when actor and target are directly bonded |
| `signal-reading` | Return zero for SENSE channels 7 and 12–15 |

Instruction costs remain payable in disabled operations. Bond removal is a recorded experimental intervention, not a simulated free action. Sharing removal preserves ALLOCATE provisioning, unbonded transfers to offspring, and TAKE. Signal-reading removal preserves emission, its energy cost, existing fields, transport, decay, and other sensing channels. These interventions isolate specific mechanisms; they do not remove all possible communication, cooperation, or resource dependence.

Reference groups always come from the unmodified source, including the bond-removal arm. The source hash and reference-world digest match across arms. Physical initial hashes differ where the treatment configuration or initial bonds differ. The reference digest is SHA-256 of compact JSON World state without the snapshot envelope or trailing newline, not `kernel.Hash`.

Parallelism is across independent worlds. `-workers 16` uses 16 worker goroutines and GOMAXPROCS=16; event ordering within a world remains deterministic. Existing output directories are rejected. The manifest becomes complete only after all snapshots, JSONL files, summaries, and trial records are written. `results.json` carries source/initial/final hashes and summaries; `summary.json` supports the existing summarizer and council workflow. Cases include the treatment suffix so downstream comparisons do not silently mix treatments.

## Persistence and validation

`Config.collective_ablation` is omitted for normal worlds. Supported nonempty values are the three ablation names above; signal-reading removal requires engineering. Ablation worlds use snapshot format **6**, including their original copying, environment, and DSL state. Unmodified observed worlds retain format 2, 3, 4, or 5 as applicable. Loading preserves the intervention and rejects incompatible format/configuration combinations. Bonds in a bond-disabled snapshot are invalid.

Telemetry stores version 1 `collectives` frames inside the existing event telemetry. Window summaries report differences in cumulative counters; their `end` block retains the full session history and final reference survival. A window's new-candidate count and the session's total candidate count are different quantities. Validation checks timestamps, sorted membership, cohort ancestry metadata, candidate founder exclusion and coverage, monotonic counters, immutable recorded identities, and consistent observation sessions. Historical group events cannot be reconstructed from a final snapshot alone.

Tests include hand-constructed positive and negative daughter examples, one-founder rejection, absent-parent rejection, membership-age reset, snapshot replay, observer neutrality, instruction costs, conservation, matched references, unchanged unbonded provisioning, worker equivalence, and continued council sessions. The synthetic positive fixture validates the observer; it is not evolutionary evidence.

## Following stages

[Stage 13](entities.md) now proposes bond, reciprocal-transfer, and copying-ancestry boundaries, preserves containment and overlap, and records subsequent-interval activity. Coordination and information closure remain unmeasured. Stage 14 compares macro descriptions against particle-level descriptions using prediction and interventions. Neither stage should declare an organism solely because it is large, connected, long-lived, or busy copying.
