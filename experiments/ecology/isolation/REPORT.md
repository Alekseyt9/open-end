# Extracted lineages: isolation and mixture

2026-09-10. **96 runs of 100,000 ticks**, 16 workers, 241.2 seconds for the full series. Rules `ecology-2`, 32×32 grid, mutations disabled. Two previously observed lineages were extracted from the seed 1 world at tick 100,000: bonded A (`05b26b88…`) and short mobile B (`8c7d401d…`). Exact programs and the source state SHA-256 are in [pair.json](pair.json).

## Conditions

For each seed from 1 to 16, the same 32 positions are selected. In the mixture, A receives even positions and B receives odd positions. Controls A16/B16 use the corresponding lineage's positions; A32/B32 use all 32 positions. Each founder receives 128 energy, zeroed memory, IP=0, and no bonds. Chemical fields and physical parameters are identical except for transport explicitly disabled in the sixth condition. Total initial energy differs between 16 and 32 founders, but energy per particle is identical; both density controls are included.

This tests **genomes in a standardized environment**, rather than continuing the entire historical ecosystem. Initial bonds, neighboring genomes, and acquired memory state from the source world are not carried over.

| Condition | Founders A/B | Final A: range (mean) | Final B: range (mean) |
| --- | ---: | ---: | ---: |
| A only | 32/0 | 669–691 (681.94) | 0 |
| B only | 0/32 | 0 | 604–632 (616.81) |
| A, same initial count as in the mixture | 16/0 | 668–687 (679.00) | 0 |
| B, same initial count as in the mixture | 0/16 | 0 | 608–632 (619.13) |
| Mixture | 16/16 | 290–427 (354.69) | 240–375 (315.38) |
| Mixture without chemical transport | 16/16 | 263–397 (337.81) | 264–392 (321.44) |

Every lineage present produces offspring in the final 90,000–100,000-tick window. In particular, A16 produces at least 8175 new copies, and B16 at least 17,671. In the mixture, the corresponding minima are 4795 and 3911. These are not merely long-lived remnants.

## Conclusion

**No obligate dependence of A on B or B on A was found in this environment.** Both programs sustain reproduction without a partner at both initial population sizes. In the mixture, each reaches a lower population than in its corresponding isolated control; this is consistent with competition for shared resources and space. The mixture also persists without chemical transport between cells.

The data do not rule out all beneficial effects or dependencies in other environments. They show that the previously observed coexistence of this pair cannot be interpreted as demonstrated metabolite-mediated interdependence. Mutations that could allow adaptation to partner removal during the experiment were disabled.

## Reproduction

```powershell
go run ./cmd/assay -pair experiments/ecology/isolation/pair.json -workers 16 -seeds 16 -ticks 100000 -every 10000 -output data/isolation-repeat
```

The output directory must be new and its parent must exist. The command creates separate JSONL files, a final summary, and a manifest. [summary.json](summary.json) contains all final counts, copy increments, and state hashes. [manifest.json](manifest.json) contains parameters and the selected programs. Raw JSONL for this experiment is in `data/isolation-assay/`, outside Git.
