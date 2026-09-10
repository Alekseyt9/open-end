# AI observation and rule proposals through chat

In Stages 6–7, Codex in the current task performs the AI role. Files provide the exchange format, so another AI or a person can use the same protocol. No keys, network API, or automatic model invocation are required. The interface consists of a CLI and readable Markdown dossiers/reports.

## First round

From the repository root:

```powershell
go run ./cmd/council prepare -input data/novelty-stage5 -out data/my-round -window 50000
```

Input must be a completed `scripts/experiments.ps1` directory containing JSONL and final snapshots. Output directories must be new. `prepare` checks the manifest, final summary, seed, physical hash, and final snapshot metrics. The current ecology DSL is supported; worlds with pending future rule changes must complete those transitions first.

The directory contains:

| File | Purpose |
|---|---|
| `brief.md` | Instructions for AI and facts about each world |
| `request.json` | Protocol version, request hash, limits, rules, and identified facts |
| `response.template.json` | Observer-response template; the empty template fails validation |
| `wNNN.evidence.json` | Complete validated window summary and detector result |
| `wNNN.snapshot.json` | Source-snapshot copy for trial continuations |

You can simply ask in the task:

> Read `data/my-round/brief.md` and act as observer and macromutation author. Write `data/my-round/response.json`, validate it, and compare the proposal with control using 16 workers.

AI covers four topics: `dominance`, `niches`, `structures`, and `stagnation`. Each claim has an `observation` or `hypothesis` kind, text, a list of fact IDs, and a `caveat`. Hypotheses require a limitation or proposed test. A rule proposal includes its mechanism, rationale, prediction, risk, evidence references, and full DSL module. Up to four proposals are allowed; the array may be empty for Stage 6.

Completed example: [response.json](../experiments/council/response.json). It is bound to its own request; another round requires its own `request_sha256` and fact references. Historical machine-readable records retain their original language and hashes.

## Validation and experiment

```powershell
go run ./cmd/council check -round data/my-round
go run ./cmd/council trial -round data/my-round -out data/my-trial -ticks 20000 -every 1000 -window 10000 -workers 16
```

The default input is `response.json` in the round directory; `-response` selects another file. `check` prints Markdown containing the response and actual cited values. Save it with `> data/my-round/review.md`.

Validation covers:

- request version/hash, kernel version, and integrity of frozen snapshots and evidence files;
- correspondence of each fact to its source-summary field, and existing, unique references;
- required topics, author, hypothesis caveats, and proposal mechanism, prediction, and risks;
- strict JSON with no duplicate keys or unknown fields;
- base rule hash, compilation, DSL work budget, matter, and energy;
- substantive rule changes: simple renaming is rejected; parameter and structural changes are distinguished;
- structural-change reachability through IDs 0/1. Adding a new ID alone is currently insufficient because mutations do not generate it.

Reference validation does not prove textual claims, niches, or causality. The observer evaluates meaning by comparing claims with cited values and full evidence files. Hashes guard against accidental changes and mixed rounds; they are not a cryptographic author signature.

One response currently applies to worlds sharing the same base rule module. Different modules require separate rounds. Proposals contain no commands, executables, patch paths, genome edits, or RNG edits. The harness executes only compiled declarative DSL.

For every source snapshot, `trial` creates a control and one branch per proposal. All start from identical state and RNG; events differ after states diverge. Defaults are 16 worker goroutines and `GOMAXPROCS=16`. Source snapshots remain available for replay. New files contain:

- `manifest.json` — parameters, request/response identity, and `running`/`complete`/`failed` status;
- `request.json`, `response.json`, `*.rules.json` — accepted inputs;
- `*.jsonl` and `*.snapshot.json` — telemetry and final states for all branches;
- `results.json` — physical hashes, summaries, and detector output per branch;
- `comparison.md` — a table comparing control and proposals.

Final `complete` status is published after results are written. Interrupted runs remain `running`; branch errors produce `failed`. Such directories cannot be used as completed next rounds. There is no automatic winner criterion yet. Novelty-detector statuses describe the last window within a branch, not a comparative ranking of different physical rules.

## Next round

```powershell
go run ./cmd/council prepare -input data/my-trial -variant solar-y-recycle -out data/my-next-round -window 10000
```

`-variant control` selects controls; a proposal name selects its continuations. Selecting a branch for discussion does not declare it an improvement. For longer manual continuation of one world, use ordinary `cmd/sim -load` with the desired final snapshot and a new metrics path.

Rolling back an experiment means returning to the frozen source snapshot or control. Switching rules alone does not undo elapsed ticks. The [Stage 8 tree harness](branching.md) now preserves and continues several selected cohorts across generations. A long-term novelty archive and automatic Pareto selection remain Stage 9.

[First completed round](../experiments/council/REPORT.md).
