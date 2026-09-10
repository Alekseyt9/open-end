# Initial Stage 1 validation

Date: 2026-09-10. Windows amd64, Go 1.26.1. Kernel `0.1.0`, rules `baseline-1`.

## Experiment

```powershell
go run ./cmd/sim -width 64 -height 64 -seed 1 -ticks 10000 -every 2000
```

Repeated with seeds 7 and 42. Mutation probability: 10,000 ppm (1%) on COPY and independently on COPYMEM. Initial state: one particle, identical seed code, finite matter, and an external energy gradient; no fixed fitness function.

State at tick 10,000:

| Seed | Particles | Executable particles | Genomes | Code + initial memory | Cumulative copies | Cumulative deaths |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 519 | 518 | 16 | 65 | 7 857 | 7 354 |
| 7 | 491 | 491 | 14 | 75 | 7 858 | 7 400 |
| 42 | 504 | 504 | 11 | 61 | 7 867 | 7 402 |

A particle can temporarily have no code between ALLOCATE and COPY. The particle count can therefore exceed the number of executable programs.

For seed 1, particle counts at ticks 2,000 / 4,000 / 6,000 / 8,000 / 10,000 were **1501 / 1228 / 900 / 675 / 519**. This run demonstrates copying, variation, and the continued existence of different programs over the observed horizon. It does not demonstrate a stable equilibrium, the usefulness of every variant, or open-ended evolution. The population decline still requires investigation.

## Control without mutations

```powershell
go run ./cmd/sim -width 64 -height 64 -seed 1 -mutation-ppm 0 -ticks 10000 -every 5000
```

At tick 10,000: **518 particles, 1 genome, 7,630 copies**. Inherited memory produces 20 state variants, but no new programs. Genetic diversity in the main experiment cannot be explained by memory changes alone.

## Replay

Compared:

1. A continuous seed 1 run on a 64×64 grid for 10,500 ticks.
2. A 10,000-tick run → JSON save → reload → another 500 ticks.

Full state SHA-256 in both cases:

```text
2c04669ea700113bab214259327665efde79a980f1ad722db69f9a84c17ce99a
```

Physical state, RNG, programs, memory, variant ancestry, and counters match.

## Automated checks

`go test ./...` and `go vet ./...` passed. Checks cover:

- Identical seeds and snapshot continuation produce identical states.
- Energy and matter balances hold at every tick of the test runs.
- Particles go extinct without available energy; the genome is preserved without mutations.
- Copied code does not share mutable storage with the parent's slice.
- Mutations preserve valid instructions and keep program length within the limit.
- ALLOCATE creates an empty particle; an occupied cell cannot be allocated again.
- A new particle does not execute on its creation tick; the particle limit is respected.
- Corrupt or incompatible snapshots and conflicting CLI arguments are rejected.
- The observer does not modify the world and emits metrics in a stable order.

## Performance

On an AMD Ryzen 7 5700X, a 24×24 grid benchmark with 300 particles at the start of the measured window took **about 79.9 µs/tick**, with 27,536 bytes and 18 allocations per tick. The benchmark repeats the same 128-tick window and excludes snapshot loading from the timing; later extinction cannot artificially improve the result. This is a reference for the current implementation, not a performance estimate for large worlds.

## Limits of the result

The Stage 1 criterion that different heritable lineages appear is supported at the level of distinct copied programs. Rigorous evaluation of adaptation, the persistence of individual new programs, novelty saturation, and ecological interactions requires further experiments. The full Stage A described in later sections of the specification is not yet complete.
