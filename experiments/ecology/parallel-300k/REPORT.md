# Minimal ecology: 48 runs on 16 compute threads

Date: 2026-09-10. Completed 48 runs: three configurations × seeds 1–16 × 300,000 ticks, for **14.4 million ticks** in total. A pool of 16 processes used `GOMAXPROCS=1` each; the Ryzen 7 5700X has 8 physical cores and 16 logical threads. The full series, including compilation and saving results, took **365.1 s**. Development checks ran concurrently, so this is the duration of an actual experiment series, not an isolated benchmark.

Physical rules and snapshot format are unchanged: `ecology-2`, format 2. Parallelism is across worlds and does not alter event resolution within a world.

## Persistence criterion

Metrics were recorded every 10,000 ticks. The primary analysis window is ticks **270,000–300,000**. A genome qualifies only if:

- At least 5 particles carry this code at each of the four window boundaries.
- The genome performs at least 5 new copies in each of the three intervals.
- The genome appears in every frame and cumulative counters never decrease.

This criterion excludes old non-replicating remnants and transient single mutants. It measures presence and new offspring over the selected horizon; it does not prove indefinite persistence or adaptation.

Behavioral groups are defined from counter increments over the window: bonds per copy, moves per instruction, and the fraction of each chemical reaction. Thresholds are documented in the [simulator reference](../../../docs/reference.md) and `observer.Persistent`. These groups describe observations; no such roles exist in the physics.

## Results

All worlds use 32×32 grids. Matter transport is enabled at an interval of 4 ticks. In ordinary ecology, chemical transport also occurs every 4 ticks, with 1% mutations on COPY/COPYMEM. Controls change exactly one of these parameters.

| Configuration | Final particle range | Persistent genomes per world | Mean persistent genomes | Worlds with ≥2 behavioral groups |
| --- | ---: | ---: | ---: | ---: |
| Ecology | 655–684 | 3–10 | 6.44 | 15/16 |
| No chemical transport | 652–683 | 2–5 | 2.81 | 16/16 |
| No mutations | 564–585 | 1 | 1.00 | 0/16 |

In ordinary ecology, each world performed **5098–14,205 copies** in the final 10,000 ticks. Multiple distinct reproducing genomes exist for all 16 seeds; in 15 seeds they also differ under the behavioral grouping used here.

The no-mutation control retains exactly one genome in all 16 worlds. This is consistent with copying errors generating the new genetic variants in the main series.

Extending the window to **100,000 ticks** (200,000–300,000, ten intervals) leaves **3–8 persistent genomes** in ordinary ecology, with multiple behavioral groups in the same 15 of 16 worlds. The result survives this longer observation window.

## What these results do not prove

Multiple persistent strategies also emerge without chemical transport. Their diversity alone therefore cannot establish necessary metabolite exchange between lineages. The higher mean number of persistent genomes with transport is an observation from this series; its mechanism has not been established separately.

The analysis found no simultaneously persistent lineages dominated by opposite reactions at the 90% threshold in any of the 48 runs. The absence of such candidates does not rule out subtler interactions. However, the stronger criterion from the later Stage B—emergent, lasting interdependence between metabolic lineages—**is not yet supported**.

The next substantive test of this criterion is to assay extracted genomes separately and in mixtures under controlled resources, then selectively disable the proposed exchange. Predefined species need not be added to force a positive result.

## Checks and reproduction

`go test ./...`, `go vet ./...`, and compilation passed. The runner integration check compared the same 16 seeds over 3000 ticks using one process and 16 processes: **all 16 full state SHA-256 hashes matched**. This check is also available through `scripts/test-experiments.ps1`.

```powershell
./scripts/experiments.ps1 -Workers 16 -Seeds (1..16) -Cases ecology,ecology-no-mutation,ecology-no-chemical-diffusion -Ticks 300000 -Every 10000 -OutputDirectory data/stage2-run
go run ./cmd/analyze -input data/stage2-run > data/stage2-run/analysis.json
go run ./cmd/analyze -input data/stage2-run -windows 10 > data/stage2-run/analysis-100k.json
```

Saved artifacts from this series:

- [manifest.json](manifest.json): parameters, status, duration, revision, and the binary hash. `SourceDirty=true` reflects runner development during the series; physical rules did not change.
- [summary.json](summary.json): results and SHA-256 hashes for all 48 worlds.
- [analysis.json](analysis.json): genomes, raw action increments, and groups over the final 30,000 ticks.
- [analysis-100k.json](analysis-100k.json): validation over a 100,000-tick window.

Large snapshots and raw JSONL remain in `data/stage2-workers16/`, which is excluded from Git. Paths in the summary refer to files from that original series; to repeat the analysis on another computer, first run the reproduction commands.
