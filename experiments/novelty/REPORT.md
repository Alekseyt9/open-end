# Stage 5: novelty and stagnation detector

Date: 2026-09-10. Heuristics tested on 16 independent worlds: 8 seeds × 2 modes, 100,000 ticks each, telemetry every 1000 ticks. World size 32×32; matter and chemical transport intervals 4. `ecology` uses standard mutations; `ecology-no-mutation` uses the same configuration with mutations disabled.

Sixteen processes, each with `GOMAXPROCS=1`. The batch completed in **51.96 s**, including the runner and result output. This is the duration of this batch, not a separate CPU/GPU benchmark. Analysis used Go; physics and the Warp backend were unchanged in Stage 5.

## Results

The detector analyzes four consecutive blocks within the window. Every result records all parameters: at least 10,000 ticks, 10% plateau tolerance, 90% monoculture threshold, 75% repetition threshold, and four quantization bins per octave of `log2(1+rate)`.

| Mode | Window | Stagnating | Persistent development | Ambiguous |
|---|---|---:|---:|---:|
| With mutations, seeds 1–8 | 80,000–100,000 | 0 | 0 | 8 |
| Without mutations, seeds 1–8 | 80,000–100,000 | 5 | 0 | 3 |
| With mutations, seeds 1–8 | 50,000–100,000 | 0 | 0 | 8 |
| Without mutations, seeds 1–8 | 50,000–100,000 | 7 | 0 | 1 |

The short window detects stagnation in control seeds 3–7; the long window detects it in all control seeds except 7. Controls have no new genomes and retain monoculture and structural plateaus. Ambiguous control cases arise when action rates cross behavioral quantization boundaries. A longer window smooths some fluctuations, but the conclusion depends on observation scale.

Mutation worlds keep producing new genomes, but that alone does not prove behavioral or adaptive novelty. The examined windows provide insufficient confirmation of persistent development under the chosen strict heuristic. `mixed` means neither proven stagnation nor proven novelty.

## False-signal check

The initial persistence check required only a new matching hash in the final two blocks. It labeled ecology seed 4 as developing over 20,000 ticks and seed 6 over 50,000. Inspecting raw rates showed that crossing a rounding boundary could create this signal without a sufficient behavioral change.

The final detector requires each of the last two blocks to differ from each of the first two by at least one full logarithmic bin on at least one feature. A regression test covers nearly identical rates in adjacent bins. Reanalysis gives the former candidates `mixed` status. Plateau and quantization parameters were not tuned to the control seeds.

Synthetic recordings test joint plateaus, persistent behavioral shifts, transient spikes, structural growth, distribution changes at constant maximum size, neutral genomes born and lost between frames, extinction, short and sparse history, rule change/rollback, gaps, and mixed sessions. Determinism, input immutability, and normalization across interval lengths were verified. CLI JSON/text output and invalid configurations were checked.

## Reproduction

From the repository root:

```powershell
./scripts/experiments.ps1 -Workers 16 -Seeds (1..8) -Cases ecology,ecology-no-mutation -Ticks 100000 -Every 1000 -OutputDirectory data/novelty-stage5-replay
go run ./cmd/summarize -input data/novelty-stage5-replay -window 20000 -detect -format json
go run ./cmd/summarize -input data/novelty-stage5-replay -window 50000 -detect -format json
go test ./...
go vet ./...
go build ./...
```

Archive files:

- `manifest.json` — parameters, duration, and physical executable SHA-256;
- `simulation-summary.json` — final physical hashes of all worlds;
- `detection-20k.json`, `detection-50k.json` — compact results with reasons, thresholds, raw rates, and behavioral hashes;
- `source-hashes.json` — JSONL and detector-source SHA-256 values.

Full telemetry, snapshots, and summaries are stored locally under `data/novelty-stage5` and excluded from Git. As in Stage 4, the observation-session identity differs from the physical snapshot hash.

This experiment checks the heuristic's operation and window sensitivity; it does not estimate accuracy against labeled examples of adaptive novelty. Behavioral features are world-aggregated, and structural features describe component sizes. Usefulness tests, per-lineage profiles, and a long-term archive remain future work.
