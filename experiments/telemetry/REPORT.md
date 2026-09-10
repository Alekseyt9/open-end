# Stage 4: event telemetry

2026-09-10. **The Stage 4 criterion is implemented:** `cmd/summarize` automatically describes the last N ticks from validated logs. World execution and observation are separate; physical hashes and reproducibility are preserved.

## Validation on actual worlds

16 seeds (1–16), 16 workers, 32×32 grids, standard 1% mutations, and matter and chemical transport every 4 ticks. Each world ran for 20,000 ticks, with records every 1000 ticks. The full validation series took 14.99 seconds, including the runner and writing results; this is not a separate telemetry overhead benchmark. The analysis window is ticks 10,000–20,000, with 11 frames per world.

| Seed | Final particles | Final genomes | Effective diversity | New genomes in window | Mean completed lifetime, ticks | Largest structure in sampled frames |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 666 | 14 | 2.77 | 75 | 346.13 | 379 |
| 2 | 632 | 31 | 4.23 | 99 | 470.50 | 16 |
| 3 | 661 | 28 | 3.41 | 71 | 548.67 | 85 |
| 4 | 651 | 29 | 2.48 | 76 | 496.97 | 1 |
| 5 | 620 | 37 | 4.47 | 109 | 395.37 | 3 |
| 6 | 615 | 38 | 4.55 | 112 | 403.55 | 1 |
| 7 | 628 | 17 | 4.55 | 96 | 377.94 | 22 |
| 8 | 643 | 22 | 1.59 | 118 | 278.91 | 1 |
| 9 | 669 | 31 | 6.13 | 96 | 370.82 | 1 |
| 10 | 636 | 30 | 2.63 | 95 | 448.02 | 1 |
| 11 | 620 | 29 | 1.64 | 70 | 360.48 | 1 |
| 12 | 628 | 29 | 7.05 | 145 | 297.52 | 205 |
| 13 | 634 | 24 | 2.28 | 115 | 279.45 | 3 |
| 14 | 601 | 27 | 3.11 | 113 | 382.25 | 2 |
| 15 | 636 | 22 | 3.22 | 119 | 330.59 | 2 |
| 16 | 621 | 18 | 1.64 | 84 | 391.33 | 1 |

The new-genome count includes all successful copies with previously unseen code, including variants that disappear before the next frame. It does not imply that these changes are adaptive. Completed lifetimes and the ages of living particles are different quantities; a component of size 1 indicates the absence of a bonded structure in the corresponding frames.

## Example explanation: seed 1

The command detected population growth from **636 → 666**, alongside a decline in effective diversity from **5.08 → 2.77**. The window contains 9287 code copies and 9269 particle deaths. Copying code is not equivalent to allocating a new particle, so copies minus deaths need not equal the population change.

The final dominant genome is `05b26b88…`: 472 particles, or 71.2% of executable particles; during the window this lineage performed 4831 copies, 5240 moves, and 4832 bonding actions. At the end, 41 bonded structures contain 490 particles; the largest component observed during the window has 379 particles. One final component contains multiple genomes.

The mean full lifetime of particles that died during the window is 346.13 ticks; final living-particle ages have P50=6170 and P90=16,269. Sparse snapshots of living particles alone would therefore substantially distort the description of completed lifetimes.

Energy inflow was 16,640,000 and dissipation was 16,639,092; stored energy increased by 908. TRANSFER moved 339,472 energy units, and ALLOCATE assigned 111,588 to startup reserves. There were 181,046 failures due to lack of space and 29,363 due to lack of matter. These are direct counters of instruction failure reasons, not a causal explanation of the overall population change.

English translations of the complete automatically generated explanations for all 16 worlds: [window-summary.txt](window-summary.txt). Machine-readable results with graphs, histograms, discoveries, and genome activity: [window-summary.json](window-summary.json).

## Correctness checks

- Tests produce the same `kernel.Hash` for identical worlds with observation enabled or disabled; changing frame frequency or resuming from a snapshot does not change the trajectory.
- Final hashes for all 16 worlds match the previous validation series after the observation format was refined.
- Completed lifetimes include particles born and killed between frames; living particles are recorded separately as censored observations.
- TRANSFER/TAKE/ALLOCATE flows and COPY counts reconcile with global counters; for every genome, initial population + code acquisitions − deaths equals final population.
- Missing frames, mixed world identities, legacy metrics without events, missing edges, and incorrect death counts are rejected.
- The DSL scenario with a change at tick 500 and rollback at tick 1000 retained the previous physical SHA-256 `fae07fd94bd6a8b6418637227ada4fff6e16a06daffbbc72dd03bb898f4ac6a2`. The 500–2000 window summary records both changes and usage of each module, although the initial and final modules are identical. [DSL summary](dsl-summary.json).
- All Go tests, `go vet`, compilation, and six Warp differential tests pass. The new event telemetry is currently implemented in Go; Warp checks confirm that physical semantics are preserved.

## Reproduction

```powershell
./scripts/experiments.ps1 -Workers 16 -Seeds (1..16) -Cases ecology -Ticks 20000 -Every 1000 -OutputDirectory data/telemetry-repeat
go run ./cmd/summarize -input data/telemetry-repeat -window 10000
go run ./cmd/summarize -input data/telemetry-repeat -window 10000 -format json > data/telemetry-repeat/window-summary.json
```

[manifest.json](manifest.json) contains parameters, series duration, and the binary SHA-256. [summary.json](summary.json) contains the final worlds' physical hashes. Raw JSONL and snapshots for this series are stored locally in `data/telemetry-stage4-verified/`, outside Git.

The interaction graph represents genomes and direct actions; anonymous chemical fields do not attribute exchange to individual lineages. Coexistence, bonds, and energy transfer are not treated as proof of obligate interdependence. Novelty and stagnation detection remain the next stage.
