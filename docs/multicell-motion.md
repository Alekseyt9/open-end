# Collective reproduction: bond motion and ancestry diagnostics

The current goal is increasing organization toward multicellularity: collective persistence, differentiation, coordination and inherited organization. This experiment addresses a concrete mechanical constraint without assigning organism classes or preprogramming reproduction of a group.

In the original physics, every bonded particle is immobile. Existing MOVE instructions have no effect until a program executes UNBIND or bonds disappear through death. The original group observer also requires descendants from at least two founding particles, so it deliberately excludes a clonal bud whose cells descend from one member of a mature parent group.

## Physical intervention

`-bond-motion yielding` permits a bonded particle's existing MOVE instruction to break its own bonds and move into an empty neighboring cell. The default empty setting retains the original immobility rule.

| Condition | Effect |
|---|---|
| Ordinary MOVE instruction attempt | Pay the existing 1-unit instruction cost |
| Free destination and sufficient remaining energy | Pay 2 additional energy units per distinct bond, remove those bonds, then move |
| Occupied destination, no reachable empty neighbor, or insufficient rupture energy | Keep bonds and position; no rupture energy is charged |

The particle must retain positive energy after rupture work; normal end-of-tick maintenance still applies. Negative MOVE directions retain the existing random-start search over four neighbors. In yielding mode this search is reached only after sufficient rupture energy and an available destination are established. A blocked attempt therefore does not consume a random direction or perturb later mutation draws. A successful move removes only the actor's bonds, so a larger component may fragment; it does not move an entire group or create a daughter automatically. Two-wide toroidal grids count repeated neighbor IDs only once.

All rupture work dissipates energy through the existing accounting ledger. Matter, code, memory, other particles' energy and the mutation repertoire are unchanged at the intervention boundary. The observer records ordinary UNBIND events and a separate successful-motion diagnostic. No new opcode or seeded behavior is introduced. Rebinding requires ordinary BIND execution.

This intervention may make fission possible, or it may destroy useful groups faster than they can reproduce. Neither outcome is assumed in advance. Yielding is opt-in; it does not replace the default.

## Matched assay

```powershell
go run ./cmd/multicell-assay -input data/environment-stage11 -case environment -out data/my-multicell-run -workers 16 -ticks 20000 -every 1000 -group-age 100
```

Each selected source is copied into `intact` and `yielding` arms. Source hash, configuration, horizon and seed are validated; sources with an existing collective ablation or bond-motion treatment are rejected. The two arms have identical initial particles, bonds, fields, programs, memory, ledgers and RNGs; only the yielding configuration differs. Independent worlds use up to 16 workers. The source batch and output directories remain separate.

Results include ordinary snapshots and JSONL, a batch summary for existing tools, and `results.json` with additional diagnostics. Group reference membership is taken from the same unchanged source in both arms. Existing strict daughter tracking and its minimum age remain enabled.

```powershell
go run ./cmd/council prepare -input data/my-multicell-run -out data/my-multicell-dossier -window 20000
```

The optional `bond_motion` configuration uses snapshot format **8** and persists on load and subsequent council/tree continuations. Formats 2–7 retain their original hashes and semantics. Relabeling a yielding snapshot as an older format is rejected. This work targets Go; Warp maintenance and UI work remain paused.

## Growth and separation diagnostics

The initial census records all linked particles, how many contain MOVE or UNBIND in their code, and how many have at least one empty neighbor containing matter. Code presence does not imply instruction execution.

Per-step diagnostics inspect programs immediately before kernel inflow and resolution. They count linked MOVE and ALLOCATE intents, lack of space in the requested direction, lack of matter when space exists, and insufficient initial allocation reserve. Resource transport and earlier instruction resolutions may subsequently change availability. These are opportunity diagnostics, **not exact causes of failed ALLOCATE resolution**. Low reserve can overlap the spatial categories.

Actual successful allocations and copies are counted when their actor is linked at resolution. Successful yielding moves separately record number of ruptured bonds and extra dissipated energy. The identity `bond_break_energy == 2 * motion_broken_bonds` must hold. These counters remain outside physical snapshots and cannot alter replay.

## Clonal branches, separately from the original criterion

The original Stage 12 criterion still requires a founder-free connected daughter containing descendants from at least two founders, plus a separate parental component retaining at least two original founders. Its definition and historical records are unchanged.

The additional observer follows the same actual-COPY founder tags but also inspects branches with one founder lineage. A **clonal daughter candidate** must satisfy all of the following at every consecutive tick boundary for at least `group-age` ticks:

1. It is a fully programmed connected group of at least two particles.
2. It contains no original founding particle.
3. All members descend through successful COPY events from one founder of the same mature parent cohort.
4. A separate connected component retains at least two original founders of that cohort.
5. Its exact member-ID set remains unchanged during the qualifying episode.

An interruption of membership, ancestry eligibility or parental presence resets qualification age. A cohort/member-set identity is credited once; later requalification only updates its last qualifying observation. The first `since_tick` and latest `last_observed_tick` can span a gap and must not be treated as uninterrupted lifetime. At most 1,024 identities are credited, with excess eligible group-ticks counted explicitly.

The diagnostic funnel also records founder-free descendant-group ticks, one- versus multiple-founder ancestry and missing-parent ticks. These categories expose which condition prevents qualification. Original founder-tagging limits and rules still apply: e.g. a founder dying before maturity is never tagged. Parent-destroying fission, mergers of cohorts and unicellular dispersal are not exhaustively covered.

Qualification records include the contributing founder and genome counts at qualification. Diagnostics also record qualified clonal group-ticks and the maximum number qualified simultaneously. Distinct qualified member sets can overlap across time and must not be treated as independent births. `experiments/multicell-motion/witnesses.py` replays intact sources to preserve mixed-genome qualification snapshots and checks their actual connected components and surviving parental founders.

Neither a clonal candidate nor a strict multi-founder candidate alone proves inherited topology, division of labor or a repeating multicellular life cycle. Useful follow-up checks include reproduction over additional generations, persistence of complementary functions, and matched disruption of those functions. The [experiment report](../experiments/multicell-motion/REPORT.md) separates mechanical effects from these stronger criteria.
