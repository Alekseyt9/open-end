# Collective reproduction: observation gap and bond-rupture experiment

**Clonal daughter-group candidates already occur under the original mechanics.** The previous multi-founder criterion deliberately excluded them. Making all bonded MOVE actions capable of rupture reduced persistent groups and clonal candidates in both tested cohorts; it is not adopted as the default.

This is evidence of persistent clonal descendant groups, not yet proof of inherited division of labor or a repeating multicellular life cycle.

## Protocol

Two source cohorts, each with seeds 1–8 at tick 100,000:

- `environment`: earlier Stage 11 worlds, with coupled engineering and evolving copying.
- `symbols`: later Stage 15 worlds, additionally offering symbolic primitives.

Each source was continued for 20,000 ticks under `intact` and `yielding` bond motion. All other initial fields and both RNGs were identical. The ordinary seed, mutation repertoire and instruction encodings were unchanged. Qualification age was 100 consecutive ticks; observation ran every tick and reporting every 1,000 ticks.

The verified experiment contains **32 continuations**, using **16 workers** per cohort. Recorded batch durations were **30.90 s** and **30.58 s**. These include cloning, telemetry and output; they are not isolated solver benchmarks. Earlier pilot directories are excluded from the evidence below.

## Initial growth and separation constraints

| Source cohort | Linked particles | With MOVE in code | With UNBIND in code | With a free matter-bearing neighbor |
|---|---:|---:|---:|---:|
| Environment | 1,039 | 1,035 | 0 | 97 |
| Symbols | 1,032 | 995 | 0 | 78 |

The census includes linked unprogrammed particles, whereas the earlier reference-cohort metric includes only fully programmed components. This explains the difference from earlier reference-member totals. Code presence does not imply execution, and this census makes no claim about extinct historical genomes.

In intact continuations, 495,935 of 579,215 linked ALLOCATE intents in the environment cohort and 585,804 of 652,214 in the symbols cohort had no empty destination in their allowed neighborhood at the pre-inflow tick boundary. These are approximately **85.6%** and **89.8%**. They show strong spatial crowding; they are not exact failure-cause counts because inflow and earlier action resolutions can change availability.

## Candidate groups and intervention effects

| Cohort | Motion | Qualified clonal member sets | Maximum simultaneous qualified clonal groups in one world | Stable group-ticks | Mixed-genome clonal member sets |
|---|---|---:|---:|---:|---:|
| Environment | Intact | 194 | 8 | 1,489,460 | 1 |
| Environment | Yielding | 0 | 0 | 128,266 | 0 |
| Symbols | Intact | 334 | 10 | 2,223,867 | 13 |
| Symbols | Yielding | 51 | 6 | 962,628 | 0 |

All 32 continuations still had **zero candidates under the original multi-founder criterion**. Neither observer limit was reached.

The **528 intact clonal candidates are distinct qualified cohort/member-ID sets, not 528 independent births**. A growing or changing descendant group can qualify with more than one membership set at different times. The maximum simultaneous counts above avoid interpreting all such episodes as coexisting daughters. The candidates had 2–6 particles; all their members descended through actual COPY events from one founder of a mature parental cohort, excluded original founders, and coexisted with a separate component retaining at least two original founders.

| Cohort / seed | Intact clonal sets | Intact maximum simultaneous | Yielding clonal sets | Yielding maximum simultaneous |
|---|---:|---:|---:|---:|
| Environment / 2 | 95 | 5 | 0 | 0 |
| Environment / 5 | 27 | 3 | 0 | 0 |
| Environment / 7 | 72 | 8 | 0 | 0 |
| Symbols / 3 | 162 | 7 | 24 | 6 |
| Symbols / 5 | 21 | 3 | 21 | 3 |
| Symbols / 6 | 151 | 10 | 6 | 2 |

All other seeds had zero clonal candidates. Symbols seed 5 performed no yielding movement; its unchanged result is a useful unexercised-treatment control.

Yielding produced 24,158 successful rupturing moves in the environment cohort and 24,125 in the symbols cohort. Their additional rupture work dissipated 48,820 and 48,794 energy units respectively. Stable group-ticks decreased by approximately **91.4%** and **56.7%**. These are descriptive matched-cohort differences, not independent group-level significance estimates. Automatic rupture on MOVE is too disruptive to promote as the next default under these conditions.

## Preserved mixed-genome witnesses

Fourteen intact candidates contained more than one genome at qualification: one from environment seed 5, twelve from symbols seed 3, and one from symbols seed 6. Different genomes do not establish different useful roles.

The original Go worlds were independently replayed to each candidate's qualification tick, using up to 16 workers. All **14 full witness snapshots** were preserved in **9.12 s** at `data/multicell-mixed-witnesses`. The verifier checked that each candidate was exactly a connected component, its cells were programmed, its genome counts matched the observation, it contained no original founder, and a separate component retained at least two original founders. These witnesses come from three source worlds and must not be treated as 14 independent experimental replicates.

The snapshots and [witness index](witnesses.json) provide concrete starting points for the next functional experiments: identify whether different cells perform complementary work, disrupt those contributions in matched copies, and test whether the organization recurs in descendants. Witnesses are ordinary resumable physical worlds, not fabricated colonies or modified seed programs.

## Verification and reproduction

All 16 intact endpoint hashes match the previously stored Stage 12 and Stage 15 control endpoints exactly. The new diagnostics therefore did not change baseline physics. Independent verification checks all source/initial/final hashes, initial code/frontier census, energy/matter/chemistry accounting, bond locality, rupture-cost identities, cohort membership constraints, candidate timestamps and genome counts. The qualification-time witnesses additionally check physical connected components through independent ordinary Go replays. Python does not independently reconstruct every founder-tag event or every intermediate persistence tick; those paths are covered by the Go observer and positive/negative tests.

Go tests cover successful and blocked motion, energy costs, multiple bonds, duplicate neighbors on narrow tori, conservation, snapshot format 8 replay, downgrade rejection, original-control equality, real one-founder COPY ancestry, parent-presence/maturity resets, candidate deduplication, worker equivalence and council/tree continuation metadata. An older one-founder negative fixture was corrected to copy along an adjacent path, ensuring it tests ancestry rather than a failed COPY.

```powershell
go run ./cmd/multicell-assay -input data/environment-stage11 -case environment -out data/my-multicell-environment -workers 16 -ticks 20000 -every 1000 -group-age 100
go run ./cmd/multicell-assay -input data/symbols-stage15 -case symbols -out data/my-multicell-symbols -workers 16 -ticks 20000 -every 1000 -group-age 100
python experiments/multicell-motion/analyze.py
python experiments/multicell-motion/witnesses.py --out data/my-multicell-witnesses --workers 16
```

Recorded inputs for the verifier are `data/multicell-motion-environment-verified` and `data/multicell-motion-symbols-verified`. [Evidence](evidence.json) contains manifests, hashes, per-seed diagnostics and all clonal records. [Mechanics and definitions](../../docs/multicell-motion.md).

**Decision:** retain fixed bonds as the default, keep yielding as an explicit experimental treatment, and prioritize functional tests of the preserved mixed-genome clonal groups. Controlled detachment and inherited organization remain research targets. UI development and Warp maintenance remain paused.
