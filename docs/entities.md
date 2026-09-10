# Automatic entity discovery (Stage 13)

`cmd/discover` proposes candidate boundaries from physical bonds, resource transfers, and successful copying ancestry. It records a hierarchy of strict containment and preserves overlapping alternatives. These are descriptions for subsequent research, not new objects inside the simulator or declarations of higher-level individuality.

The implementation runs entirely in Go. It replays existing experiments with a read-only observer and requires exact agreement with their final snapshot hashes. No physical rules, seed programs, fitness functions, snapshots, or UI are changed. [The first experiment](../experiments/entities/REPORT.md) analyzes all 32 Stage 12 continuations.

## Run

```powershell
go run ./cmd/discover -input data/collectives-stage12 -source data/environment-stage11 -out data/entity-discovery -workers 16 -every 1000 -group-age 100
```

`-input` is a completed `group-assay` directory. `-source` is the original completed batch containing the source snapshots identified by the assay. Replay duration and treatments come from the assay, so the observer cannot silently choose a different physical horizon. `-every` controls discovery checkpoints; `-group-age` controls the persistence threshold. `-record-limit` defaults to 65,536 unique directed particle pairs and 65,536 acting particles per interval.

The command validates batch completion, source metadata and hashes, treatment initialization, and saved final snapshots before creating its output directory. Each replay then verifies its final physical hash against the original. Failures are recorded in the manifest; a divergent replay cannot be published as complete. Existing output directories are rejected. Sixteen independent workers use GOMAXPROCS=16, with unchanged deterministic event order inside each world.

The original Stage 12 telemetry aggregates ordinary interaction edges by genome. Those edges cannot establish which specific particles exchanged resources. Replay recovers particle-ID events from the actual simulator execution; it does not infer missing events from genome similarity or final state.

## Candidate boundaries

| Source | Candidate rule | What it does not establish |
|---|---|---|
| `bonds` | Fully programmed connected components of at least two explicitly bonded particles | Obligate cooperation or a shared life cycle |
| `reciprocal_transfer` | Strongly connected components of positive directed TRANSFER edges among particles still programmed and alive at the checkpoint | Mutual benefit, synchronous interaction, or causal coordination |
| `copy_ancestry` | At least two living programmed particles descended from one coded particle present at replay start | A physical boundary, local proximity, or collective reproduction |

Every coded particle at replay start receives a distinct ancestry root. Successful COPY propagates its actor's root to the copied target; allocation-parent metadata is not used as a substitute. Roots remain meaningful after their original particle dies, but dead members leave the current candidate. Common ancestry before replay start is not reconstructed.

This first detector covers programmed particles. Persistent terrain artifacts and nonprogrammed scaffolds are outside its candidate vocabulary.

TRANSFER direction follows the actual donor and recipient. TAKE is recorded separately as resource flow from victim to taker and cannot create a reciprocal-transfer candidate. One-way provisioning of offspring also cannot create reciprocal connectivity by itself. Reciprocal paths may involve more than two particles. The graph aggregates an interval; it does not enforce temporal ordering along a path or require all edges to coexist at one instant.

Candidate member IDs are sorted. Identical member sets from different sources are merged into one candidate carrying all applicable `boundary_sources`. Its `id` is SHA-256 of the JSON member-ID array, scoped by the containing world/report. This identity tracks a member set, not a species, phenotype, or continuous life history.

## Hierarchy and persistence

`coded_particle_ids` lists the complete micro level, including particles absent from every compound candidate. A candidate's `members` links it to these leaves. `direct_subset_candidates` lists maximal proper candidate subsets: transitive containment links are omitted. `structural_depth` is 1 for a candidate without another candidate inside it, 2 for one containing such a candidate, and so on; individual particle leaves are depth 0.

This is an inclusion DAG, not a forced partition or a lineage tree. A candidate can have several parents. Overlapping sets without containment retain both alternatives and increment `overlapping_noncontained_candidates`. Containment depth alone does not establish a biological hierarchy: a scattered ancestry family can contain a physical group, or a physical group can contain a copying family.

Two persistence measures remain separate:

- `bond_membership_age` uses Stage 12's exact member-set tracking at every tick boundary. `persistent_bond_boundary` requires that age to reach `-group-age`.
- Reciprocal-flow persistence requires the same exact component at at least two consecutive discovery checkpoints and an observed checkpoint span reaching `-group-age`. Missing or incomplete intervals reset this episode. With `-every 1000`, a 100-tick minimum still requires at least a 1,000-tick checkpoint span for flow persistence.

Copying ancestry is a separate hypothesis about membership, with no automatic persistence or individuality label. `previously_qualified_daughter_membership` connects a current member set to a previously qualified Stage 12 daughter record. The [original daughter criteria](collectives.md) still apply; copying by any member does not by itself demonstrate whole-group reproduction.

## Discovery and subsequent evidence

`discovery_interval_activity` attributes interval events retrospectively to the member IDs present in a discovered candidate at the checkpoint. It counts internal, incoming, and outgoing TRANSFER energy; equivalent TAKE flows; copies by members; and energy acquired through ABSORB or positive net CONVERT effects. Events can precede the candidate's final topology. These are activity counts, not a claim that the exact boundary existed throughout the interval.

For each candidate, the next interval freezes its member IDs prospectively. The subsequent checkpoint adds:

- `next_interval_activity`, including events involving frozen members that die during the interval;
- the interval boundaries and `next_interval_complete`;
- `next_interval_members_alive`, without adding newborn descendants to the frozen set;
- `next_interval_reciprocal_connectivity` for previously discovered reciprocal-flow candidates, requiring a return path entirely inside the original member set and every member still programmed and alive at the next checkpoint.

An outside particle cannot complete a candidate's internal return path. Positive internal transfer in the next interval is weaker evidence than retained reciprocal connectivity; the index reports both separately. Repeated candidate observations are not independent experimental replicates. The final frame has no subsequent interval: its missing `next_interval_activity` and timestamps mean unavailable evidence, not failed persistence.

`transfer_retention` is internal TRANSFER energy divided by internal plus incoming plus outgoing TRANSFER energy. It is null when that denominator is zero. It excludes field acquisition, chemistry, energy costs, and TAKE; even a value of 1 is not energetic autonomy or information closure. No score for information closure, causal coordination, or higher-level individuality is assigned.

## Output and resource bounds

| File | Contents |
|---|---|
| `manifest.json` | Configuration, source paths, completion, worker count, and measured wall time |
| `index.json` | Per-world report hashes, verified final hashes, endpoint counts, maximum depth, and subsequent-interval evidence counts |
| `seed-N-TREATMENT.json` | Source/initial/final hashes, complete checkpoint candidates, micro IDs, provenance labels, activity, and interpretation limits |

These are analysis artifacts, not loadable physical snapshots or ordinary metrics JSONL. `cmd/discover` does not change the snapshot format and does not require Warp. Use the preserved original snapshots for physical continuation. The stage currently provides CLI and JSON outputs; UI work remains deferred.

Activity maps reset after each interval. Reaching either record limit marks the entire interval incomplete and reports dropped event records. The detector then emits no reciprocal-flow candidates or activity ratios from that interval. Bond and copying-ancestry candidates remain available because they use independent data. The existing Stage 12 cohort/candidate caps and skipped count are also exposed. Full report size grows with the number of checkpoints and candidate memberships.

## Verification and next stage

Tests cover reciprocal cycles, one-way transfer, predation, dead endpoints, external return paths, containment reduction, noncontained overlap, prospective activity, independent intervals, ancestry attribution to actual COPY actors, truncation, deterministic report hashes across worker counts, observer neutrality, and failed replay publication.

[Stage 14](causal.md) now evaluates future survival with whole-seed exclusion, particle and group predictors, randomized boundary controls, and local bond interventions. Boundaries are selected from prior discovery evidence and evaluated on new continuations. The inclusion hierarchy and persistence counts alone do not answer that causal question; the first forecast experiment found no macro advantage over the local particle model.
