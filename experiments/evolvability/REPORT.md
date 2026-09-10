# Stage 10: inherited copying strategies

The implementation allows different genomes to execute different copying policies. In this experiment, new policies appeared and were used, including one local recombination event. **All sampled worlds still had only one persistent code-copying policy in each final observation window.** This is evidence that the mechanism operates, not that selection maintains a diversity of evolvability strategies.

## Protocol

- Eight seeds (1–8) in each of three cases: evolving policies, fixed policies, and evolving policies with zero mutation.
- 32×32 ecology worlds, matter and chemical diffusion every four ticks, a single unchanged seed replicator, base mutation rate 1% in the two mutation-enabled cases.
- 100,000 ticks per world, metrics every 1,000 ticks, 16 workers with one Go thread per batch worker.
- An unchanged-rules continuation of all 24 frozen worlds for another 20,000 ticks, through the branch harness with 16 workers.
- Windows: 90,000–100,000 and 110,000–120,000, each with 11 frames. A persistent genome/policy requires at least five live individuals in every frame and a successful copy in every interval.

The initial batch took **78.21 seconds**, including build, reporting, and result collection. The tree continuation took **15.79 seconds**. These are wall-clock measurements of this workload, not isolated physics benchmarks or a CPU/GPU comparison.

The [compact evidence](evidence.json) contains the original batch manifest, binary/source identity, continuation settings, per-world snapshot hashes, policy aggregates, and cumulative history counters. Large snapshots and JSONL remain under the ignored local directories `data/evolvability-stage10` and `data/branching-stage10`.

## Results

Counts below sum the eight worlds in each case within the specified window. “Changed” counts actual changed offspring code, not attempted mutation events.

| End tick | Case | Code copies | Changed | Recombined | Persistent policies per world |
|---:|---|---:|---:|---:|---|
| 100,000 | Evolving | 77,857 | 831 | 0 | 1 in all eight |
| 100,000 | Fixed | 90,837 | 905 | 0 | 1 in all eight |
| 100,000 | Zero mutation | 125,037 | 0 | 0 | 1 in all eight |
| 120,000 | Evolving | 77,127 | 711 | 1 | 1 in all eight |
| 120,000 | Fixed | 92,795 | 963 | 0 | 1 in all eight |
| 120,000 | Zero mutation | 125,074 | 0 | 0 | 1 in all eight |

The number of executed code policies in the evolving worlds was:

| Seed | Window ending 100k | Window ending 120k | Distinct policies executed over the full 120k history |
|---:|---:|---:|---:|
| 1 | 1 | 2 | 5 |
| 2 | 1 | 1 | 9 |
| 3 | 1 | 1 | 8 |
| 4 | 2 | 1 | 6 |
| 5 | 2 | 4 | 9 |
| 6 | 2 | 2 | 6 |
| 7 | 1 | 1 | 3 |
| 8 | 3 | 1 | 9 |

Fixed-policy controls used exactly one code policy. Zero-mutation controls preserved a single genome and recorded no changed code or recombination. Different genomes can share the same decoded policy, so persistent genome counts must not be presented as persistent strategy counts.

In evolving seed 1 during ticks 110,000–120,000, a nondefault policy used a 0.25% effective rate, mixed operators, recombination, and proofreading. It produced 20 copies, one changed offspring code, and one recorded donor splice. It did not satisfy the persistence threshold. The same world recorded 8,704 copies using the default policy.

The evolving case did not produce more copies than its controls in these windows. This small seed sample does not establish a general benefit or cost of evolvability. Fixed and evolving worlds use the same six mutation operators, but their random trajectories diverge when inherited policies begin affecting execution. Neither raw genome novelty nor a donor splice demonstrates adaptation.

## Verification and next question

Go tests, vet, and build passed. Tests cover policy mechanics, controls, accounting, snapshot replay, and telemetry integrity. All seven Warp differential/rejection tests passed on the RTX 5070; encoded policies themselves are currently Go-only. The existing seven-node Stage 8/9 tree still validates with unchanged frozen hashes.

The new tree contains two nodes and 48 preserved world snapshots. The archive accepts the continued cohort. The offline inspector was checked at desktop and 390-pixel mobile widths: world selection, code/memory filters, and persistence filtering work without JavaScript errors or page overflow.

The next research question is whether an environment with changing local opportunities can maintain multiple copying strategies. Stage 11 concerns environmental coevolution. The present result should remain the baseline when adding those mechanisms; persistent strategy diversity and its adaptive value remain open criteria.

## Reproduce

Use fresh output directories:

```powershell
./scripts/experiments.ps1 -OutputDirectory data/evolvability-repeat -Ticks 100000 -Every 1000 -Workers 16 -Seeds (1..8) -Cases evolvability,evolvability-fixed,evolvability-no-mutation
go run ./cmd/summarize -input data/evolvability-repeat -window 10000 -format json
go run ./cmd/council tree init -input data/evolvability-repeat -out data/evolvability-repeat-tree -window 10000
go run ./cmd/council tree grow -tree data/evolvability-repeat-tree -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council tree archive -tree data/evolvability-repeat-tree
go run ./cmd/council tree export -tree data/evolvability-repeat-tree
```

See [policy encoding and interpretation](../../docs/evolvability.md) for exact mechanics and limitations.
