# Stage 8: branching worlds and retained continuations

Date: 2026-09-10. The Stage 8 harness preserves several world cohorts, their ancestry, frozen evidence, and repeatable continuations. An offline HTML explorer supports comparison and explicit selection. No automatic winner or Pareto archive is implemented.

## Recorded experiment

The source was the completed Stage 5 batch: 16 worlds at tick 100,000, consisting of ecology and ecology-no-mutation, seeds 1–8. The root dossier used a 50,000-tick observation window.

Generation 1 reused the validated Stage 7 `solar-y-recycle` response on the identical request. Each source world produced a paired control and a proposal continuation, 20,000 ticks each. **All 32 final physical hashes matched the archived Stage 7 trial.**

Generation 2 continued both resulting cohorts for another 20,000 ticks under their respective existing rules. No additional AI response or rule change was introduced.

| Run | Parents | Published children | World continuations | Additional ticks per world | Recorded duration |
|---|---|---|---:|---:|---:|
| r000001 | b000001 | b000002, b000003 | 32 | 20,000 | 19.15 s |
| r000002 | b000002, b000003 | b000004, b000005 | 32 | 20,000 | 19.18 s |

Both runs used 16 workers, `GOMAXPROCS=16`, reporting every 1000 ticks, and a 10,000-tick final analysis window. Parent cohorts within a generation execute sequentially; workers parallelize their independent worlds. Recorded durations cover trial execution and child dossier publication; initial import and preflight validation are excluded. These are actual experiment timings, not isolated performance benchmarks.

The result is **5 preserved cohorts, 80 world states, 2 generations, and 64 new continuations totaling 1.28 million world-ticks**. Both terminal cohorts remain selected. The root and both intermediate cohorts remain available.

## Terminal cohorts

Final observation window: ticks 130,000–140,000. Each row aggregates eight seeds within one case; populations and copies are sums, diversity is an arithmetic mean across worlds.

| Direction | Case | Final population | Mean effective diversity | Window copies |
|---|---|---:|---:|---:|
| Baseline continuation (b000004) | ecology | 5360 | 8.875 | 83,107 |
| Proposal continuation (b000005) | ecology | 5255 | 6.710 | 57,883 |
| Baseline continuation (b000004) | ecology-no-mutation | 4617 | 1.000 | 124,835 |
| Proposal continuation (b000005) | ecology-no-mutation | 4484 | 1.000 | 115,033 |

Every terminal world produced copies in the final window. Among mutation worlds, the minimum was 6332 copies for the baseline direction and 2463 for the proposal direction. The proposal's lower mean diversity and copying persist in this comparison; retaining it exercises alternative-direction preservation and does not accept it as an improvement.

Two baseline mutation worlds have a local `developing` detector status in this window. This is a heuristic observation, not evidence of open-ended evolution or a universal ranking against modified rules.

The child variant name `control` means unchanged rules relative to the immediate parent. Thus b000005 retains the earlier `solar-y-recycle` rules; it does not revert to baseline.

## Validation

- Multi-generation tests retain both directions and verify correct parentage.
- One-worker and 16-worker tree executions produce identical world results.
- The control-control path matches direct kernel continuation.
- Moving the original batch after import does not break tree growth.
- Invalid or duplicate selections, missing responses, modified snapshots, changed comparison metrics, and false ancestry are rejected.
- Exclusive writer locking blocks concurrent mutations.
- A failure after publishing one cohort preserves that cohort, records failure, leaves the previous selection intact, and allows a new run from the parents.
- HTML export can be refreshed and cannot overwrite a JSON artifact through a mistaken destination extension.
- Browser checks exercised five cards, 32 selected world rows, case filtering, saved/tip selections, evidence navigation, proposal commands, and option validation. Desktop and narrow layouts were checked; no page JavaScript errors occurred.
- `go test ./...`, `go vet ./...`, and `go build ./...` passed.

## Reproduction

```powershell
go run ./cmd/council tree init -input data/novelty-stage5 -out data/branching-repeat -window 50000

# Only valid for the exact original source batch/request.
Copy-Item -LiteralPath experiments/council/response.json -Destination data/branching-repeat/nodes/b000001/round/response.json

go run ./cmd/council tree grow -tree data/branching-repeat -proposals -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council tree grow -tree data/branching-repeat -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council tree export -tree data/branching-repeat
```

If the original batch is unavailable, reproduce it using the [Stage 5 commands](../novelty/REPORT.md). Changed inputs require a newly authored response tied to the resulting request.

[tree-compact.json](tree-compact.json) records every node, world metric, physical hash, request identity, selection, and run duration from this validation. Full snapshots, evidence, JSONL, and the HTML report are stored locally in `data/branching-stage8/`, outside Git. The compact archive alone is not a resumable tree.

[Interface and storage reference](../../docs/branching.md).
