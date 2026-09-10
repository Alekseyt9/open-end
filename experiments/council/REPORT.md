# First AI observer and macromutation round

Date: 2026-09-10. Codex in the current chat performed the Stage 6 and 7 roles. `cmd/council` prepared data, validated the structured response, and ran the experiment; no external AI API was used.

## Inputs and observer findings

Sixteen Stage 5 source snapshots at tick 100,000: ecology and ecology-no-mutation, seeds 1–8. Observation used ticks 50,000–100,000. The dossier was checked against original telemetry, the manifest, and final snapshots.

The observer described differences in dominance, possible energy-acquisition strategies, bonded structures, and stagnation-detector limitations. Niches and stagnation causes were stated as hypotheses. For example, ecology seed 1 has a largest component of 338 particles; it is not declared a multicellular organism. Direct energy absorption is much greater in ecology seed 2 than seed 1; isolation tests were proposed to establish separate niches.

[Validated response with facts](review.md) · [Machine-readable response](response.json) · [Dossier](request.json).

## Proposed mechanism

`solar-y-recycle` retains reaction 0 and replaces reaction 1:

```text
ID 0: X → Y + 4 particle energy
ID 1: Y + 4 field energy → X
```

Each attempt costs 1; the maximum batch is 64. Chemical-unit count and energy potential are conserved. This is classified as a structural change: consumed and produced resources change, not just cost coefficients.

ID 1 is available to existing programs and mutations. A new ID unsupported by the mutation generator could remain unused, so that option was not chosen. Genomes, individuals, and resources were not edited. Loss of the former Y → Z energy pathway and reduced viability were recorded as risks before the experiment.

## Trial comparison

For each source snapshot: an unchanged-rule control and a proposal branch. **32 continuations × 20,000 ticks**, 16 worker goroutines, `GOMAXPROCS=16`, recording every 1000 ticks. The entire trial took **17.05 s**, including result output. Final analysis window: ticks 110,000–120,000.

| Group, 8 source worlds | Control | Proposal |
|---|---:|---:|
| With mutations: mean final effective diversity | 7.849 | 6.644 |
| Without mutations: mean final effective diversity | 1.000 | 1.000 |
| With mutations: total window copies relative to control | 100% | 64.29% |
| Without mutations: total window copies relative to control | 100% | 92.29% |

The new reaction is used in all 16 modified branches. With mutations, effective diversity increased relative to paired control in 3 of 8 worlds and decreased in 5; copying decreased in all eight. No world was extinct at the end. Without mutations, no new genomes appeared, as expected.

Decision: **retain the proposal as an experiment, not as an improvement to baseline rules**. Mean diversity and copying declined. Individual improvements do not establish long-term benefit; this is a short comparison without adaptive-value assessment. Detector transitions between `mixed` and `stagnating` are not an automatic winning criterion.

[All paired results](comparison.md) · [Compact data and physical hashes](results-compact.json) · [Rules](solar-y-recycle.rules.json) · [Trial parameters](manifest.json).

## Harness validation

`go test ./...`, `go vet ./...`, and `go build ./...` passed. Checks covered:

- incomplete batches, incorrect seeds, and modified evidence, facts, or snapshots;
- stale responses, missing sources, required topics, and hypothesis caveats;
- duplicate JSON keys, wrong base modules, energy creation, cosmetic renaming, and unreachable new IDs;
- observer-only responses with no proposals;
- identical physical hashes with 1 and 2 workers, and control parity with direct `kernel.Step` continuation;
- source-snapshot preservation, output-overwrite rejection, and complete manifests;
- preparing the next round from a selected trial variant.

Using real data, `data/council-round2` was also prepared from 16 `solar-y-recycle` branches. This validates protocol reuse; it does not accept the patch. A second AI response and experiment have not been performed.

## Reproduction

```powershell
go run ./cmd/council prepare -input data/novelty-stage5 -out data/council-replay-round -window 50000
# Copy this archive's response.json into the prepared round.
# Identical inputs yield the same request_sha256; otherwise prepare a new response.
go run ./cmd/council check -round data/council-replay-round
go run ./cmd/council trial -round data/council-replay-round -out data/council-replay-trial -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council prepare -input data/council-replay-trial -variant solar-y-recycle -out data/council-replay-next -window 10000
```

Full evidence files, JSONL, and snapshots are stored locally in `data/council-round1` and `data/council-trial1`. The archive contains the dossier, response, review, validated module, manifest, table, compact results, and source hashes. Source-reference checks do not automatically verify the meaning of all AI claims. This English document translates the report; hash-bound machine-readable records retain their original contents.
