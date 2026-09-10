# Warp versus Go: physics solver

2026-09-10. **With 256 simultaneous worlds, warm Warp compute is 1.67× faster than Go with 16 workers. Go is faster for 1–96 worlds.** Including preparation and extraction, GPU execution loses in every tested short run. Go remains the primary solver for current small batches.

Hardware: Ryzen 7 5700X (8 physical cores / 16 logical threads), RTX 5070 12 GiB, Windows, NVIDIA driver 616.64. Go 1.26.1, Python 3.13.14, Warp 1.15.0 from the existing `F:/src/game_arc/.venv-warp` environment.

## Timing

Each world: 32×32, standard ecology with 1% mutations, matter/chemical transport every 4 ticks. Shared physical-state warmup: 300 Go ticks, followed by **1000 measured ticks per world**, with three repetitions from identical initial state. Medians below are seconds for the entire batch.

| Simultaneous worlds | Go, 1 worker | Go, 16 workers | Warp: compute | Warp: preparation + compute + extraction |
| ---: | ---: | ---: | ---: | ---: |
| 1 | 0.161 | 0.169 | 1.663 | 1.714 |
| 16 | 2.561 | 0.333 | 2.305 | 3.315 |
| 96 | 15.006 | 1.836 | 2.286 | 8.727 |
| 256 | 41.279 | 4.659 | 2.786 | 19.094 |

For one world, the Go pool executes one task regardless of the 16-worker limit. At 256 worlds, GPU throughput is about 91,885 world-ticks/s versus 54,950 for Go with 16 workers. This is batch throughput, not acceleration of an individual world.

Cold Warp compilation took about 55 seconds in a separate pilot and is excluded from warm timings. The main series used the compilation cache. Warp time includes explicit device synchronization after every 8-tick chunk; asynchronous launch is not treated as completed computation.

The last column additionally includes Python array preparation, GPU transfers, physical-state extraction, and resource-invariant checks; disk JSON output is excluded. CPU columns measure ordinary `kernel.Step` only, excluding preparation/snapshot parsing, final validation, and serialization. **Total observer work differs:** Go maintains genome/origin histories; Warp does not yet build them. Compute-kernel gains therefore cannot be presented as whole-application acceleration.

## Result parity

All three repetitions at every batch size matched exactly on:

- every cell and particle: resources, ID/parent, code, working and inherited memory, IP, flag, target, creation tick, and generation;
- bonds, tick number, next ID, and both SplitMix64 states;
- global resource, copy, death, reaction, and failure counters;
- DSL module state, schedule, transition log, and execution counters in separate DSL checks.

The table covers 1,107,000 GPU world-ticks compared against corresponding Go results. These hashes identify a canonical physical-state projection, not the complete Go snapshot with historical logs.

Six differential tests additionally cover all 17 opcodes, energy starvation, elevated mutation rates, odd grids and unequal transport intervals, mixtures with bonds and slot reuse, DSL load/rollback across chunk boundaries, and a new reaction ID consuming particle energy. `go test ./...`, `go vet ./...`, the build, and all Warp tests pass.

The Warp CLI also continued a real DSL snapshot from tick 750 to 2000 through its scheduled rollback. The result is compared with the physical projection of a full Go run; histories are not synthesized and the result is not presented as a Go snapshot.

## Implementation and practical conclusion

The port preserves sequential conflict resolution and conditional RNG calls inside each world. Worlds execute independently. It uses integer arrays, reusable slots, and a separate ID-ordered particle list. All VM actions and local DSL conversions execute on GPU.

The first version assigned neighboring worlds to neighboring lanes of one CUDA warp. A pilot with 16 worlds and 10% mutations took 16.35 s. Giving each world a separate warp with one active lane reduced this to 2.03 s with the same final physical hash. These are individual pilot measurements, not the main table's medians: [packed version](packed-pilot.json), [isolated worlds](isolated-pilot.json).

Sequential dependencies and memory accesses prevent a GPU advantage for a single world. Large batches hide some latency. Further Warp measurements make sense for long runs with hundreds of worlds while retaining device arrays between chunks. This benchmark did not test 100,000-tick runs and does not claim linear speedup over that horizon.

Next practical optimizations include avoiding unnecessary historical-data copying during array preparation, reducing Python state-conversion costs, and collecting observations without reconstructing every particle on CPU each reporting tick. Any change to within-world interaction order requires separate reproducibility checks.

## Reproduction and source data

```powershell
& F:/src/game_arc/.venv-warp/Scripts/python.exe warp-sim/benchmark.py --worlds 1,16,96,256 --size 32 --ticks 1000 --warmup 300 --repeats 3 --workers 16 --scenario ecology --output data/warp-repeat
& F:/src/game_arc/.venv-warp/Scripts/python.exe -m unittest discover -s warp-sim -p test_solver.py -v
```

[summary.json](summary.json) contains all repetitions, parameters, versions, source/binary hashes, and initial/final-state hashes. Full initial and reference states are under `data/warp-benchmark-32`, outside Git. The 256-world batch was interrupted and resumed: two complete CPU results were reused after hash verification; missing CPU repetitions and all GPU repetitions were rerun. This is recorded in the summary.

API details, independent execution from Go snapshots, and limitations: [warp-sim/README.md](../../warp-sim/README.md).
