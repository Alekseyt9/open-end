# Warp solver

Experimental GPU physics backend for `ecology-2`. Implements all 17 VM instructions, code and memory copying with mutation, conservative transport, chemical charging, bonds, creation/decay, both SplitMix64 RNGs, and installation/rollback of compiled DSL modules.

Copying here means the legacy model. Stage 10's `copy_model: fixed|evolving` is currently Go-only. Both the Go fixture bridge and the Python `Batch` API reject it explicitly; they never silently substitute legacy mutation semantics. See [evolvability](../docs/evolvability.md).

Stage 11's terrain, signals, and 19-instruction engineering repertoire are also Go-only and explicitly rejected. The existing backend retains the original 17-instruction ecology. See [environment engineering](../docs/environment.md).

Stage 12's collective ablations and per-tick ancestry observer run in Go. Both input paths explicitly reject nonempty `collective_ablation`; an unsupported treatment is never silently ignored. See [collective observation](../docs/collectives.md).

The recorded runs used the existing Python environment at `F:/src/game_arc/.venv-warp/Scripts/python.exe`, Warp 1.15.0, and an RTX 5070. Dependencies for other environments are listed in `requirements.txt`.

## Run from the repository root

```powershell
& F:/src/game_arc/.venv-warp/Scripts/python.exe warp-sim/run.py --snapshots data/dsl-split.json --ticks 1250 --output data/warp-result.json
```

Multiple snapshots can follow `--snapshots`; they must share physical parameters except for seed. The existing Go loader validates snapshots before execution. Instead of `--snapshots`, `--fixture` accepts an `initial` bundle from `cmd/warp-reference` or benchmark output. Input validation requires Go; the ARC source project and installed environment packages are not modified.

The `warp-physical-report-1` result contains physical state, global counters, both RNGs, and DSL state. **It is not a Go snapshot:** GPU execution does not build historical `genomes`/`origins` logs, their SHA-256 values, or cumulative per-genome metrics. Do not pass the report to `cmd/sim -load` or equate its `physical_sha256` with a full Go snapshot hash. Missing history is not fabricated. To continue within Python, call `Batch.run` repeatedly on the same instance; arrays remain on the device between calls.

Initial backend limits: common grid/settings within a batch; VM arguments in `[-2^30, 2^30)`; initial tick/ID values below `2^62`; estimated main-array storage capped at 6 GiB. Standard and mutated project programs fit these bounds. Interactive external module installation remains a Go API; Warp executes the frozen schedule from the snapshot.

## Speed comparison

```powershell
& F:/src/game_arc/.venv-warp/Scripts/python.exe warp-sim/benchmark.py --worlds 1,16,96,256 --size 32 --ticks 1000 --warmup 300 --repeats 3 --workers 16 --scenario ecology --output data/warp-benchmark-repeat
```

The output directory must be new. Before timing, each world runs for 300 Go ticks; CPU and GPU then start from identical state. CPU measurements use the actual `kernel.Step`, with 1 and 16 workers, rather than a Python equivalent. Initial preparation, snapshot loading, and final serialization are excluded from CPU time. Go continues maintaining its normal historical logs.

Warp records warm solver time with synchronization after every tick chunk, time including array preparation/transfers/physical-state extraction, and initial module loading/compilation time separately. Cold JIT is excluded from warm speedup. `warp_end_to_end_seconds` includes Python preparation and invariant checks but excludes final JSON report serialization. This compares the working Go kernel with an experimental backend doing less observer work; a Warp compute advantage is not a speedup of the complete Go pipeline.

Every repetition compares all cells, particles (ID, parent, code, memory, IP, target, age/generation), bonds, both RNGs, global accounting, and the DSL log. Exact equality is required, without tolerances. The summary contains all timings, medians, and source/initial/final-state hashes. Partial runs or failed comparisons receive `failed` status.

Scenarios: `ecology` — standard 1% mutations; `mutation` — 10% stress; `baseline` — no chemistry; `assay` — a mixture of two isolated lineages without mutations; `dsl` — transition at tick 100 and rollback at 200; `odd` — unequal grid dimensions and transport intervals. Use `--warmup 0` when testing DSL transitions so both occur during measured steps.

## Execution and checks

Conflict resolution and conditional RNG calls remain sequential inside each world. Free particle slots are reused; a separate list preserves ID order and cyclic shifting. A new particle executes starting next tick but pays maintenance in its creation tick.

The default `--lanes 32` assigns one CUDA warp per world with one active lane. This reduces interference from divergent world branches. `--lanes 1` packs neighboring worlds into one group and was much slower in the pilot. `--chunk 8` bounds each launch on a GPU also driving the display. Synchronization is included in timings. This uses [NVIDIA Warp](https://nvidia.github.io/warp/); conflicting actions inside a world are not parallelized.

```powershell
& F:/src/game_arc/.venv-warp/Scripts/python.exe -m unittest discover -s warp-sim -p test_solver.py -v
go test ./...
go vet ./...
```

Differential tests cover all opcodes, energy starvation, mutations, odd grids with different transport intervals, mixtures with bonds and slot reuse, DSL, and different tick chunking. Set `WARP_TEST_DEVICE=cpu` to test Warp CPU execution; that is a separate Warp backend, not the benchmark's Go reference.
