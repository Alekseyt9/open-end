# Stage 11: local environmental engineering

Mutation produced programs that build local terrain and emit signals. Terrain measurably changed resource transport and incoming energy, and matched feedback interventions changed subsequent dynamics. **New niches and adaptive engineering were not established.** The strongest builders were single surviving individuals with no produced copies during the final 10,000-tick observation window.

## Protocol and artifacts

The initial batch ran 24 worlds: seeds 1–8 for coupled terrain, inert terrain, and coupled terrain with zero mutation. All used 32×32 ecology, matter/chemical diffusion every four ticks, evolving copying, and the unchanged generalist seed. Mutation-enabled cases used a 1% base rate. Each world ran for 100,000 ticks, with metrics every 1,000 ticks. The reported window is 90,000–100,000.

Sixteen workers completed the batch in **80.82 seconds**, including build and output collection. A second experiment continued each of the eight coupled snapshots for 10,000 ticks in both feedback modes, again with 16 workers. Those 16 matched continuations took **5.10 seconds**. The timings measure these workloads and are not GPU comparisons.

The [compact evidence](evidence.json) records the original batch manifest and binary identity, per-world final hashes, window action counts, actor production, and all intervention source/initial/final hashes. Full local snapshots and JSONL are under `data/environment-stage11` and `data/environment-assay-stage11`. The branch explorer is `data/branching-stage11/index.html`.

## Initial observations

Counts below are summed over eight worlds during ticks 90,000–100,000.

| Case | Matter built | Signal energy emitted | Inflow blocked by terrain | Mixing amount attenuated |
|---|---:|---:|---:|---:|
| Coupled | 321 | 57 | 30,125 | 45,502 |
| Inert | 716 | 333 | 0 | 0 |
| Zero mutation | 0 | 0 | 0 | 0 |

The inert case still stores construction matter, emits signals, pays costs, and senses fields. Only terrain attenuation is disabled. The zero-mutation control creates no engineering instructions or fields.

Two coupled worlds maintained a terrain cell containing eight matter units at the final frame:

| Seed | Main builder genome prefix | Matter built by that genome in the window | Population at end | Produced copies in the window |
|---:|---|---:|---:|---:|
| 4 | `be1955f0aeca` | 156 | 1 | 0 |
| 5 | `db8682bab33d` | 157 | 1 | 0 |

Seed 5 also had a transient builder that placed one matter unit. Its maintained terrain cell at (6, 17) blocked 30,016 units of otherwise available inflow and attenuated 23,203 units of immediate resource mixing across the window. Seed 4's structure affected transport without blocking additional inflow in that window.

No signal energy remained in any final sampled frame, despite recorded emissions. These fields dissipated; emission counters alone do not establish a communication system.

## Matched feedback test

Each row compares two continuations from the same 100,000-tick snapshot. All physical state and RNGs were initially identical; only the terrain attenuation switch differed. Counts cover ticks 100,000–110,000.

| Seed | Copies, coupled | Copies, inert | Population, coupled | Population, inert |
|---:|---:|---:|---:|---:|
| 1 | 15,268 | 15,138 | 660 | 662 |
| 2 | 10,556 | 10,556 | 673 | 673 |
| 3 | 10,001 | 10,001 | 672 | 672 |
| 4 | 16,164 | 16,282 | 658 | 662 |
| 5 | 13,364 | 12,879 | 666 | 663 |
| 6 | 8,192 | 8,213 | 685 | 679 |
| 7 | 12,050 | 12,066 | 670 | 675 |
| 8 | 7,939 | 7,931 | 673 | 671 |

Seeds 2 and 3 had no construction in either continuation and matched on population and copying. Seed 5 produced 485 more copies under coupled terrain; seed 4 produced 118 fewer. Small transient construction events also caused some paired histories to diverge. The effect has no uniform direction in this sample.

These interventions show that the engineering feedback can affect subsequent dynamics. They do not attribute extra copies to the builder, establish a long-term advantage, or prove a new niche. The maintained builders did not reproduce in the measured initial window, so their activity must not be described as an established inherited adaptation.

## Validation

- Go tests, vet, and build passed; tests cover physical conservation, bounds, erosion, signal decay, tick-start sensing, snapshot/CLI replay, telemetry integrity, and assay source immutability.
- Matched-assay tests produced identical final hashes with one and 16 workers and rejected tampered source metadata.
- Eight Warp differential/rejection tests passed on the RTX 5070. Engineering itself is Go-only.
- Existing Stage 8/9 and Stage 10 trees still validate their frozen hashes.
- Desktop and 390-pixel mobile browser checks passed for world/layer selection, field coordinates, keyboard access, and actor tables, with no JavaScript errors or page overflow.

## Reproduce and continue

Use new output directories:

```powershell
./scripts/experiments.ps1 -OutputDirectory data/environment-repeat -Ticks 100000 -Every 1000 -Workers 16 -Seeds (1..8) -Cases environment,environment-inert,environment-no-mutation
go run ./cmd/summarize -input data/environment-repeat -window 10000 -format json
go run ./cmd/env-assay -input data/environment-repeat -out data/environment-repeat-assay -ticks 10000 -every 1000 -workers 16
go run ./cmd/council tree init -input data/environment-repeat -out data/environment-repeat-tree -window 10000
go run ./cmd/council tree archive -tree data/environment-repeat-tree
go run ./cmd/council tree export -tree data/environment-repeat-tree
```

The next investigation should track offspring of engineering genotypes and test whether specific lineages benefit from conditioned environments across repeated seeds. The Stage 11 criterion of a new adaptation creating new niches remains open. See [mechanics and intervention semantics](../../docs/environment.md).
