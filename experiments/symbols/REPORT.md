# Stage 15: symbolic primitives

**Primitive use appeared, but a beneficial symbolic convention was not demonstrated.** Eight evolving worlds generated inscriptions and memory lookups. One world read nonempty inscriptions, including foreign and two-token words. Changing or suppressing reception did not change the final physical state after normalization described below.

## Protocol

32 worlds: seeds 1–8 in `symbols`, `symbols-scrambled`, `symbols-unreadable`, and `symbols-no-mutation`. Each ran 100,000 ticks on a 32×32 grid with ecology, diffusion intervals 4, evolving copying, and coupled engineering. The normal 13-instruction ecology seed was unchanged. Reporting interval: 1,000 ticks. Sixteen worker processes completed the batch in **109.93 seconds**, including runner overhead. This is a recorded run duration, not an isolated performance benchmark.

From the eight persistent-mode endpoints, a further 24 matched continuations tested persistent, scrambled, and unreadable reception for 20,000 ticks each, again using 16 workers. The verified continuation batch took **14.78 seconds**. All source words, memories, programs and both RNGs were preserved at branching. No nonempty word was read in the persistent continuation arms.

## Development observations

Counters below cover the complete 100,000-tick persistent runs. “Context” counts LOOKUP executions with a nonzero context offset, not demonstrated context-dependent meaning.

| Seed | Writes | Reads | Nonempty | Pairs read | Foreign reads | Context lookups |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 3,489 | 754 | 0 | 0 | 0 | 15,051 |
| 2 | 11,141 | 13,364 | 0 | 0 | 0 | 19,246 |
| 3 | 745 | 3,270 | 0 | 0 | 0 | 4 |
| 4 | 208,395 | 51 | 0 | 0 | 0 | 11,003 |
| 5 | 18 | 3,839 | 0 | 0 | 0 | 26,000 |
| 6 | 2,936 | 397 | 0 | 0 | 0 | 242 |
| 7 | 4,113 | 109 | 0 | 0 | 0 | 471 |
| 8 | 768 | 249,032 | 108 | 31 | 104 | 43 |

Six historical genomes in seed 8 read nonempty inscriptions. None of these genomes executed LOOKUP. The main reader accounted for 97 nonempty reads; its LISTEN wrote memory register 1, while its conditional copying routine compared register 0. This is consistent with reception having no useful behavioral effect in that lineage. It is not a general proof that all symbol information was unused at every intermediate tick.

All eight no-mutation controls recorded zero symbol operations, confirming that the mechanisms were not programmed into the ancestor.

## Controls and limits

For all 16 development comparisons and all 16 paired continuation comparisons, removing only `config.symbols` and the `world.symbols` observation ledger made the endpoint worlds exactly equal. This comparison retains physical inscriptions, particle memories, programs, both RNGs, ancestry, environment fields and resource counters. Full snapshot hashes still differ because they include control configuration and counters.

For all 16 development comparisons, every sampled metric was also identical after removing symbol metrics and the session-initial hash. These checks do not establish equality of every intermediate memory value. The absence of nonempty reads in the later continuation makes that experiment a low-exposure control, not a strong challenge of an active convention.

Hand-written tests separately establish that ordered tokens survive snapshot replay, context can select arbitrary memory entries, and a reception intervention can change a subsequent conditional physical action. Those fixtures are capability checks, not evidence of evolutionary discovery.

## Reproduction

See [mechanics and commands](../../docs/symbols.md). Local source batch: `data/symbols-stage15`; matched continuation: `data/symbol-assay-stage15-verified`.

```powershell
python experiments/symbols/analyze.py --assay data/symbol-assay-stage15-verified
```

The independent standard-library verifier checks every snapshot hash, source/initial intervention hash, final telemetry ledger, counter sum and reported difference. It writes [evidence.json](evidence.json), including all case/seed counts, reader programs and comparisons. Large physical snapshots and JSONL files remain local under `data/`.

The next priority is to measure adaptive capability and functional diversity directly. Neither multicellularity nor symbolic communication is a required form of progress. A single-particle system may improve without either phenomenon.
