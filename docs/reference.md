# Simulator and experiment reference

Implementation of the [research plan](../open_ended_evolution_ai_world_plan.md): executable replicators, mini-ecology, a local-reaction DSL, telemetry, and stagnation detection. Stages 6–7 use `cmd/council` for AI observation in chat, validated rule proposals, and paired experiments. Requires Go 1.26+ and uses only its standard library.

Stage 4 provides event telemetry and automatic descriptions of selected tick windows. Go records the additional metrics with `-metrics`; Warp remains an experimental physics solver.

One seed replicator runs in a two-dimensional world, allocates matter, copies memory and code, and transfers energy. Copying errors produce heritable variants. Available space, matter, and energy determine competition; there is no fitness function.

## Running the simulator

From the repository root:

```powershell
go run ./cmd/sim -width 32 -height 32 -seed 1 -ecology -matter-diffusion 4 -chemical-diffusion 4 -ticks 10000 -every 1000
```

This enables chemical resources and their transport. Without flags, the original control world is **256×256**, with one seed replicator and a 1% mutation probability separately for each `COPY` and `COPYMEM`. Use `-matter-diffusion 4` for a baseline world with sustained matter turnover. Transport intervals are measured in ticks; `0` disables transport.

Saving state and metrics:

```powershell
go run ./cmd/sim -width 32 -height 32 -seed 1 -ecology -matter-diffusion 4 -chemical-diffusion 4 -ticks 10000 -save data/world.json -metrics data/run.jsonl
go run ./cmd/sim -load data/world.json -ticks 5000 -save data/continued.json
```

`-ticks` always means **additional** ticks. A snapshot contains configuration, both RNGs, chemical fields, bonds, programs, memory, ancestry, and resource counters. World parameters cannot be overridden with `-load`; they come from the snapshot. The final `state_sha256` line compares complete states. Snapshots use format **2** without DSL state and **3** with it; rules are **ecology-2**. Stage 1 snapshots (format 1) are rejected as incompatible; there is no implicit migration.

JSONL output requires a new file to avoid erasing an earlier experiment. Snapshots are written through a temporary file and rename; an existing destination snapshot may be replaced. Run files under `data/` are excluded from Git.

Control without mutations:

```powershell
go run ./cmd/sim -width 64 -height 64 -seed 1 -mutation-ppm 0 -matter-diffusion 4 -ticks 10000
```

All options: `go run ./cmd/sim -h`.

## Implemented features

- A toroidal world with four neighbors and one particle per cell.
- Integer energy, an external spatial gradient, and finite matter.
- A VM executing one instruction per particle per tick, with eight memory cells and bounded code length.
- Replication through `ALLOCATE → COPYMEM → COPY → TRANSFER`; movement breaks local access. There is no dedicated reproduction instruction.
- Instruction replacement, insertion, and deletion; block duplication and deletion; initial-memory mutation.
- Death when energy is exhausted, matter returned to the cell, and execution and maintenance costs.
- Deterministic SplitMix64 RNG, snapshot/replay, and format/kernel/rule versioning.
- Particle, program, inherited-variant, copy, death, energy-flow, and age metrics; variant ancestry in snapshots.
- Conservative matter transport; three chemical states X/Y/Z and two local conversions.
- Neighbor targeting, energy taking, particle bonds, and unbinding.
- Genome frequencies, generations, offspring, cumulative genome behavior, interval copy/death rates, and failed-operation causes.
- Reproducibility, resource-invariant, limit, mutation, and CLI tests; benchmarks.

## Architecture

| Directory | Responsibility |
| --- | --- |
| `internal/kernel` | Tick order, event collection/resolution, snapshot/replay |
| `internal/world` | Physical state, configuration, invariant checks |
| `internal/vm` | Instruction format, control flow, seed program |
| `internal/rules` | Basic physical effects, operation costs, inflow, decay |
| `internal/dsl` | Reaction parser, conservation checks, bounded bytecode, version history |
| `internal/experiment`, `cmd/assay` | Isolated genomes, mixtures, parallel control series |
| `internal/evolution` | Copying errors, RNG, inherited-state hashing |
| `internal/observer` | Metrics without affecting simulation state |
| `cmd/sim` | CLI and experiment files |
| `cmd/analyze` | Genome persistence analysis for completed JSONL batches |
| `cmd/summarize` | Validate and explain the last N ticks using event telemetry |

Energy enters at the start of a tick. Particles are ordered by ID with a cyclic shift based on the tick number. The VM reads state before events are applied; SENSE captures its value in the event. Conflicts are resolved in this order, rechecking available resources. A new particle first executes on the next tick. At the end of the tick, maintenance is charged and zero-energy particles decay.

Basic rules are implemented in Go; the DSL can replace or add local `CONVERT` reactions. Changes to built-in semantics should increment `rules.Version` or `kernel.Version`: incompatible snapshots will not load. DSL modules have their own versions and hashes. Field transport uses a separate serialized RNG without consuming the mutation RNG stream.

## Instructions and costs

Every instruction attempt costs energy, including blocked actions. If energy is insufficient, the remainder dissipates and the instruction does not execute. Each particle also pays 1 unit of maintenance per tick.

| Instruction | Effect | Energy |
| --- | --- | ---: |
| `NOP` | No physical effect | 1 |
| `SENSE A B` | Store own energy (`A=0`), cell energy (`1`), or cell matter (`2`) in memory `B` | 1 |
| `MOVE A` | Move N/E/S/W (`0..3`); `A<0` searches for an empty neighbor from a random direction | 1 |
| `ABSORB A` | Take up to `A` energy units from the cell | 1 |
| `ALLOCATE A` | Spend 1 matter in a neighboring cell, create an empty particle, and remember its ID | 4 |
| `COPYMEM` | Copy eight memory cells into an empty neighboring target, with possible mutation | 8 |
| `COPY` | Copy code into an empty neighboring target, with possible mutation | Source code length |
| `TRANSFER A` | Give the target up to `A` energy units while retaining at least 1 | 1 |
| `WRITE A B` | Write value `B` to memory `A` | 1 |
| `READ A B` | Copy memory `A` to memory `B` | 1 |
| `COMPARE A B` | Set the flag to `memory[A] >= B` | 1 |
| `JUMP A B` | Jump to `A`: always (`B=0`), if flagged (`1`), or if not flagged (`-1`) | 1 |
| `CONVERT A B` | Built-in reaction `A=0`: X→Y; `A=1`: Y→Z; up to `B` chemical units | 1 |
| `TARGET A` | Select a neighboring particle by direction; `A<0` searches from a random direction | 1 |
| `TAKE A` | Take up to `A` energy from the neighboring target | 1 |
| `BIND` | Bond to the selected neighboring target | 1 |
| `UNBIND A` | Remove the target bond, or all own bonds if `A<0` | 1 |

`ALLOCATE` additionally **transfers**, rather than creates, 12 energy units from the source to the empty particle. This reserve provides time to copy memory and code. `COPY` allocates no matter and supplies no initial energy. The target is accessible only while adjacent; `COPYMEM` occurs before the target receives code. Code and memory addresses wrap modulo their size. Default limits are 64 instructions, 256 energy per particle, and 128 energy per cell.

In ecology mode, `SENSE A B` also reads X/Y/Z for `A=3/4/5`. A bond holds both particles in place: all own bonds must be removed before movement. Death removes bonds automatically. Bonds are unique and restricted to neighbors, so degree is at most four. In control mode, the new instructions are excluded from the mutation pool; chemical reactions require `-ecology`.

The following identities can be checked after every completed tick:

```text
field energy + particle energy + 8×X + 4×Y = initial energy + inflow − dissipation
field matter + particle count = initial matter
X + Y + Z = initial chemical amount
```

`genomes` counts distinct programs, ignoring memory. `lineages` counts active **code + initial inherited memory** variants. Thus `genomes=1` without mutations, but `lineages` may exceed one because the program passes on memory containing SENSE results. Use `genomes` as evidence of genetic variation. The `origins` history stores each variant's first origin and copy count; it is not yet a complete genealogy.

JSONL `active_genomes` records genome frequencies and behavior. `births` counts particles that received this code; `copies` counts copies produced by that genome, including mutated offspring. `last_copy_tick` helps distinguish a reproducing lineage from a long-lived remnant. Behavior counters are cumulative, including dead particles of that genome; subtract two frames to analyze a window. `interval` holds global reporting-window deltas. Origin and genome archives grow with discovered variants and are not yet compacted.

## Mini-ecology

X/Y/Z are chemical states with energy potentials 8/4/0. Field energy charges Z→X at a cost of 8 units; programs perform X→Y or Y→Z and receive 4 units each. No chemical transition creates energy. Y appears only after the first reaction executes. Transport lets neighboring cells use the products left behind.

The world starts with **one** program capable of both reactions and self-copying. Specialized species, roles, and cooperation rules are not predefined. In seed 1, a bonded, relatively immobile lineage and a short mobile replicator emerged and coexisted; their frequencies and offspring are recorded in the [Stage 2 report](../experiments/ecology/RESULTS.md).

The initial X→Y-only variant exhausted the cycle before another reaction emerged. That was an unsuccessful pilot, not evidence of ecology. Both reactions are explicit in the current seed; the experiment studies subsequent evolution, not the origin of metabolism from scratch.

## Verification

```powershell
go test ./...
go vet ./...
go test ./internal/kernel -run '^$' -bench BenchmarkStep -benchmem
```

[Historical Stage 1 results](../experiments/baseline/RESULTS.md) use `baseline-1` rules. [Long runs and Stage 2](../experiments/ecology/RESULTS.md) record diagnostics, no-mutation controls, and three ecology seeds.

[Extended validation with 16 workers](../experiments/ecology/parallel-300k/REPORT.md): 48 runs of 300,000 ticks, 16 seeds per configuration. Ordinary ecology retains 3–10 persistently reproducing genomes; distinct behavioral groups remain in 15 of 16 worlds. Controls show that this does not yet establish obligatory metabolite exchange.

Reproduce the batches and obtain JSONL, snapshots, and summaries (PowerShell 7+, several minutes):

```powershell
./scripts/experiments.ps1 -Workers 16
```

By default, the script runs four configurations for seeds 1/7/42: original control, matter transport, transport without mutations, and mini-ecology. Up to 16 independent processes run with `GOMAXPROCS=1` each. `-Workers` sets the process count without changing event order inside a world. On a Ryzen 7 5700X, this uses 16 logical threads across eight physical cores.

Results go to a new `data/experiments-...` directory: separate JSONL, snapshots, and stdout per world, sorted `summary.json`, and `manifest.json` containing parameters, status, source revision, binary hash, and batch duration. On failure or interruption, processes started by the script are stopped; partial data remains and is not marked complete. Duplicate seeds/configurations and existing output directories are rejected.

An ecology batch with 16 seeds and three control configurations:

```powershell
./scripts/experiments.ps1 -Workers 16 -Seeds (1..16) -Cases ecology,ecology-no-mutation,ecology-no-chemical-diffusion -Ticks 300000 -Every 10000 -OutputDirectory data/stage2-run
go run ./cmd/analyze -input data/stage2-run > data/stage2-run/analysis.json
```

`cmd/analyze` accepts only completed batches. By default, a genome is persistently reproducing if its count stays at least 5 at every boundary of the last three windows and it produces at least 5 copies in each window. Missing observations exclude a candidate: old cumulative copies cannot count as new. Thresholds are controlled by `-windows`, `-min-count`, and `-min-copies`. These are observer criteria, not selection rules inside the world.

Behavior groups describe actions only within the selected window: frequent binding means at least 0.25 binds per copy; frequent movement means at least 0.01 moves per instruction. With at least 100 chemical conversions, a Y→Z share of at most 10% means X→Y-dominated behavior; at least 90% means Y→Z-dominated behavior; otherwise reactions are mixed. Raw counters remain in the report. Both dominant reaction types occurring together identifies a complementarity candidate, not proof of interdependence.

Short runner check: `./scripts/experiments.ps1 -Ticks 100 -Every 50 -Seeds 1`. Equal-result verification with 1 and 16 processes: `./scripts/test-experiments.ps1`.

Matter transport removed the observed decline in copying over 100,000 ticks. Several reproducing strategies are observed; persistent interdependence between specialized metabolic lineages and growth in adaptive novelty remain unproven.

[Two-lineage isolation experiment](../experiments/ecology/isolation/REPORT.md): 96 runs of 100,000 ticks, 16 processes, 241.2 seconds. Both lineages reproduce alone at two initial densities and coexist in mixtures without chemical transport. Obligatory interdependence of this pair was not detected in the standardized environment.

## Stage 3: initial DSL implementation

Implemented: a strict JSON parser, resource-consumption/production bytecode, execution budgets, SHA-256 versions, module changes between ticks, and rule rollback. Examples: [baseline reactions](../examples/rules/baseline.json) and [direct X→Z with a new reverse reaction](../examples/rules/direct-x.json). [Smoke run and replay hashes](../experiments/dsl/SMOKE.md).

```powershell
go run ./cmd/sim -width 32 -height 32 -ecology -matter-diffusion 4 -chemical-diffusion 4 -rules examples/rules/baseline.json -rule-change 500=examples/rules/direct-x.json -rollback-at 1000 -ticks 2000 -every 250 -save data/dsl-world.json
```

At the start of tick 500, `CONVERT 0` changes from X→Y with 4 energy output to X→Z with 8. At tick 1000, the original module returns. Particles, memory, RNGs, and resources continue. `-rule-change tick=path` is repeatable; ticks must be unique. A transition scheduled for tick N occurs on the next `Step` when `World.Tick == N`: a snapshot after N steps still has it queued.

Source and bytecode for all scheduled modules are frozen before execution and stored in the snapshot. `-load` resumes them without reading source files; loading cannot override rules through flags. Worlds without DSL retain format 2 and their existing state hash; worlds with DSL history use format 3, including the queue, history, transition log, and counters. Loading recompiles sources and compares their bytecode and hashes.

The world owner calls `kernel.ReloadRules`, `kernel.ScheduleRules`, and `kernel.RollbackRules` between steps without restarting the kernel. Installation is atomic: compilation errors, history overflow, or invalid future schedules leave the world unchanged. These APIs are not intended to run concurrently with `Step` in another goroutine. Rollback changes only the rule module; restoring physical state requires a snapshot. The CLI currently uses predefined schedules rather than watching files.

The DSL is limited to local X/Y/Z, `field_energy` (cell energy), and `energy` (particle energy). Documents require `format: 1`, a text `version`, `instruction_budget` from 1 to 64, and 1–16 rules. Each rule has unique `id` (0–15) and `name`, nonempty `consume`/`produce`, attempt cost `energy_cost`, and batch limit `max_batch` (both 1–64). Coefficients are integers from 1 to 64. A module replaces the listed IDs; omitted IDs 0/1 retain their built-in reactions, while other missing IDs do nothing. The example adds ID 2 to charge Z→X using particle energy. Mutation currently chooses reaction arguments only from 0/1; an explicitly written program can use a new ID.

The compiler enforces conservation of X+Y+Z and energy at potentials 8/4/0/1/1. Attempt cost dissipates separately through the normal mechanism. Execution checks resources and capacity, then applies the entire feasible batch atomically. Per-call work is `2 × bytecode instruction count + 5`; bounded batch size does not create an execution loop. This computational budget is separate from physical operation cost. There are no jumps, recursion, external functions, filesystem access, or network access. Source is limited to 64 KiB; unknown fields, duplicate JSON keys, and nesting beyond 12 are rejected. History is limited to 32 modules, the queue to 32 changes, and the log plus queue to 256 events.

Exact reaction usage is stored in `world.rule_state.usage` by `hash/name`, interpreter work in `instructions`, and activations in `events`. The CLI prints the active version and hash. Legacy `converted[0/1]` metrics count batches for those IDs; after replacement, consult the corresponding module for their chemical meaning. New IDs use DSL counters in the snapshot.

Validation covers malformed rules, atomic rejection, new IDs and attempt costs, conservation, change/rollback without resetting the world, snapshot continuation after source files are removed, and baseline DSL parity with built-in reactions. Arbitrary new fields, DSL versions of other operations, interactive patch delivery to a running process, automatic branch selection, and AI API calls remain unimplemented.

## Stage 4: telemetry and window explanations

```powershell
go run ./cmd/sim -width 32 -height 32 -seed 1 -ecology -matter-diffusion 4 -chemical-diffusion 4 -ticks 20000 -every 1000 -metrics data/telemetry-run.jsonl -save data/telemetry-run.json
go run ./cmd/summarize -input data/telemetry-run.jsonl -window 10000
go run ./cmd/summarize -input data/telemetry-run.jsonl -window 10000 -format json
```

`cmd/sim -metrics` and `cmd/assay` automatically add telemetry version 1. The observer is separate from the world: it consumes no RNG, is not serialized into the physical snapshot, and does not change its hash. `observer.Observe` still reads basic metrics; complete recording uses `observer.NewTracker`, `kernel.StepObserved`, and `Tracker.Frame`.

Collected data:

- population, genomes, and ancestry lineages; Shannon entropy in natural logarithms, effective genome count `exp(H)`, inverse Simpson diversity, and dominant-genome share;
- P50/P90/maximum ages of particles alive at the frame, separate from completed lifetimes of particles that died during the interval;
- completed-lifetime count, mean, minimum/maximum, and histogram buckets 0, 1, 2–3, 4–7, and so on; deaths between sparse frames are retained;
- energy inflow/dissipation, absorption, TRANSFER, TAKE, ALLOCATE startup reserves, chemical charging, conversions by ID, and DSL usage by hash/name;
- bond-graph components, sizes, bonded-particle counts, mixed-genome structures, and existing bonds between genomes;
- interval graphs of successful ALLOCATE/COPY/TRANSFER/TAKE/BIND/UNBIND operations, deaths by genome, and new genomes, including those that appear and disappear between frames.

Graph nodes represent **code genomes**, not individual particles or memory-bearing `origin` variants. An empty hash denotes an unprogrammed particle. TRANSFER and TAKE follow actual energy flow; for TAKE, the victim is the source and the executor the recipient. ALLOCATE records its 12-unit reserve transfer. Zero or failed actions are not successful edges. BIND/UNBIND represent program-initiated actions; death removes bonds physically but is not reported as voluntary UNBIND. Structure sizes include singletons; bonded components contain more than one particle.

A dead particle's lifetime is the completed death-tick boundary minus `Created`: creation and decay within one tick yield 1. The mean covers only particles that died in the selected window, including empty particles; living particles are right-censored and reported separately by age. This is not an estimate of the entire population's mean lifetime. Age quantiles use the nearest rank. Chemical fields are anonymous: matter transfers are not attributed to specific genomes without a separate tracking mechanism.

The summary reports actual window boundaries, population/diversity changes, dominant genomes and reproduction, new genomes and those absent at the end, lifetimes, structures, flows, and failures. Cumulative genome-action deltas are reported only when the baseline is known: the genome was present initially or first appeared within the window. A returning old genome's unknown baseline is not treated as zero.

`-window` is measured in ticks. Analysis takes the last frame at or before the requested start and every subsequent interval, without interpolation. With insufficient history, it begins at the first available frame. Population and structure extrema refer to reporting frames, not every tick. Graphs, deaths, new genomes, and flows include events from all ticks.

Every frame contains the seed, observation start, and SHA-256 of the initial JSON `World` state without its snapshot envelope. This identifies the observation session and is **not** `kernel.Hash`. Resuming a physical snapshot creates a new session; it cannot silently be merged with the previous one. Analysis rejects gaps, mixed sessions, decreasing counters, inconsistent genome births/deaths, flows, and graphs. Older JSONL without event telemetry remains usable by `cmd/analyze`, but cannot recover exact lifetimes and interactions through `cmd/summarize`.

A 16-process batch and a summary for every world:

```powershell
./scripts/experiments.ps1 -Workers 16 -Seeds (1..16) -Cases ecology -Ticks 20000 -Every 1000 -OutputDirectory data/telemetry-batch
go run ./cmd/summarize -input data/telemetry-batch -window 10000 -format json > data/telemetry-batch/window-summary.json
```

Directory input requires a complete manifest and agreement between final metrics and the runner summary. A single JSONL file is described through its last recorded frame without claiming the entire run is complete. DSL transitions and versions are included: identical reaction IDs can have different physical meanings under different modules.

[Stage 4 validation on 16 worlds and an automatically generated explanation](../experiments/telemetry/REPORT.md).

## Stage 5: novelty and stagnation

```powershell
go run ./cmd/summarize -input data/telemetry-run.jsonl -window 10000 -detect
./scripts/experiments.ps1 -Workers 16 -Seeds (1..8) -Cases ecology,ecology-no-mutation -Ticks 100000 -Every 1000 -OutputDirectory data/novelty-batch
go run ./cmd/summarize -input data/novelty-batch -window 20000 -detect -format json
```

`-detect` adds a separate version 1 `dynamics` block to JSON and an explanation to text output. Defaults require at least 10,000 ticks and real boundaries in all four quarters of the window. The same telemetry integrity checks apply as for summaries. Analysis runs after simulation and changes no world, snapshot, RNG, or physical hash.

Features and decisions:

- **Diversity plateau:** effective genome count range `(max-min)/max(1,min)` is at most 10% across all window frames.
- **Monoculture:** the same dominant genome occupies at least 90% of executable particles at every frame. This is a separate diagnostic and does not replace other stagnation conditions.
- **Structural plateau:** largest-component range is at most 10%, and total variation distance between the initial size distribution and each later frame is at most 0.1. The distribution is particle-weighted, with size buckets 1, 2–3, 4–7, and so on.
- **Behavioral hash:** successful ALLOCATE/COPY/TRANSFER/TAKE/BIND/UNBIND, absorbed energy, X/Y conversions, Z charging, and total DSL units. Events are normalized per 1000 particle-ticks; the denominator is estimated from sampled populations by the trapezoidal rule. Rates are quantized as `round(4*log2(1+rate))`, then hashed with SHA-256. Genomes and seeds are excluded.
- **Stagnation:** simultaneous diversity, structure, and population plateaus, plus at least 75% repeated behavioral hashes. A repeat has appeared in any earlier block of this window; the first block is excluded from the denominator. With four blocks, 75% effectively requires all four to match.
- **Persistent behavioral change:** the last two blocks share a hash absent from the first two. Each new block must also differ from each reference block by at least a full logarithmic bin on at least one feature: merely crossing a rounding boundary is insufficient.
- **Persistent structural growth:** at every frame after the start of the second half, the largest component exceeds the first-half maximum by more than `max(2, 10% of the initial maximum)` particles.

Statuses: `stagnating` — joint plateau; `developing` — persistent behavioral change or structural growth; `mixed` — ambiguous dynamics; `insufficient_history` — short history or sparse frames; `extinct` — no executable particles; `rule_change` — rules changed within a sufficiently long window, so changes are not attributed to internal evolution. Extinction is checked before minimum history length.

JSON records configuration, actual boundaries, rates and hashes of all four blocks, ranges, and reasons. CLI options: `-detect-min-ticks`, `-detect-tolerance`, `-detect-resolution`. Full configuration is also available through `observer.DefaultDynamicsConfig` and `observer.DetectDynamics`.

These are local window heuristics, not proof of adaptive novelty. New genomes are counted separately and do not increase the development assessment. Sensitivity depends on the window, sampling cadence, and quantization. Global action rates may conceal a rare strategy; movement, signals, individual-lineage behavior profiles, and graph topology are not yet hashed. A long-term novelty archive belongs to Stage 9. `developing` calls for subsequent persistence and usefulness checks; it does not automatically authorize rule changes.

[Stage 5 experiment: mutation and control, 16 worlds × 100,000 ticks](../experiments/novelty/REPORT.md).

## Stages 6–7: AI through chat and a file protocol

Codex in chat acts as observer and proposal author. `cmd/council` prepares dossiers, validates responses, and compares proposed rules with controls on copied snapshots:

```powershell
go run ./cmd/council prepare -input data/novelty-stage5 -out data/my-round -window 50000
# Ask Codex to read brief.md and complete response.json.
go run ./cmd/council check -round data/my-round
go run ./cmd/council trial -round data/my-round -out data/my-trial -workers 16
# Prepare the selected branch for the next discussion:
go run ./cmd/council prepare -input data/my-trial -variant solar-y-recycle -out data/my-next-round -window 10000
```

The dossier contains identified facts, summaries, and source-snapshot copies. Claims cite facts; hypotheses are separated from observations. Macromutations are full DSL modules with a mechanism, prediction, and risk. Validation checks hashes, references, conservation, and reaction reachability in the current VM. Source worlds remain unchanged; runs produce `comparison.md`, JSONL, snapshots, and branch summaries. No API key is required.

[Interface and exchange format](council.md) · [Completed round on 16 source worlds](../experiments/council/REPORT.md). Observation and macromutations are implemented within the current reaction DSL. Automatic AI invocation is not enabled. The [Stage 8 tree harness](branching.md) preserves multiple selected cohorts, their ancestry, and repeated continuations, with an offline HTML explorer. The [Stage 9 archive](archive.md) adds immutable selection decisions, Pareto comparisons on measured proxies, and protected behavior cells. The next major stage is **10, Evolution of Evolvability**.

## Experimental Warp solver

[warp-sim](../warp-sim/README.md) implements a GPU physics backend: all 17 opcodes, mutations, resource transport, bonds, both RNGs, and scheduled DSL transitions. It accepts validated Go snapshots and computes independent worlds on GPU. Genome and origin histories remain a Go feature; Warp outputs a physical report incompatible with `cmd/sim -load`.

[RTX 5070 versus Ryzen 7 5700X comparison](../experiments/warp/REPORT.md): 32×32 worlds, 1000 ticks each, medians of three repetitions. For 16 worlds, Go with 16 workers is faster (0.33 s versus 2.31 s); for 256 worlds, Warp compute is 1.67× faster (2.79 s versus 4.66 s). Including preparation and extraction, the short GPU run takes about 19 s for 256 worlds. All physical fields, RNGs, and global counters match exactly; Go remains the default.
