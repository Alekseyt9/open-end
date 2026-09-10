# Adaptive capability: first allocation round

**The new proxy found no sufficiently large adaptive information benefit in these eight worlds.** All selection and reserved-validation scores were zero at the predeclared 0.01 information-benefit threshold. This is a negative result for this assay, not a claim that the worlds have zero intelligence.

The experiment implements the revised priority: develop adaptive capability and functional diversity without requiring multicellularity, mutual exchange, or a particular organizational level. [Protocol, formula and limitations](../../docs/adaptivity.md).

## Runs and evidence

Source: the eight persistent-mode Stage 15 worlds at tick 100,000. Each received four challenges × three arms, for **96 probes of 1,000 ticks**, plus eight ordinary 1,000-tick behavior references. Two challenges determine selection and two are reserved for reporting validation. Four selected original worlds then continued normally for 20,000 ticks. The verified round used up to **16 workers** and completed in **9.23 seconds**, including evaluation and ordinary continuations.

| Seed | Selection robustness | Perception benefit | Memory benefit | Effective behavior profiles | Allocation |
|---:|---:|---:|---:|---:|---|
| 1 | 0.71602 | 0 | +0.001016 | 6.574 | Exploration |
| 2 | 0.71394 | 0 | −0.000183 | 4.903 | Retained source only |
| 3 | 0.71864 | 0 | −0.000280 | 2.568 | Behavioral spread |
| 4 | 0.72788 | 0 | +0.000214 | 8.078 | Behavioral spread |
| 5 | 0.72247 | 0 | +0.000126 | 4.043 | Retained source only |
| 6 | 0.72028 | 0 | +0.000182 | 6.562 | Retained source only |
| 7 | 0.71419 | 0 | −0.001970 | 4.301 | Behavioral spread |
| 8 | 0.71900 | 0 | +0.000675 | 5.056 | Retained source only |

Robustness is mean capped executable population retention, including descendants. Benefits are signed differences in that retention, not percentages of acquired intelligence. Effective behavior profiles are a coarse action-rate proxy; they are not species or certified distinct functions.

The largest selection information benefit was 0.000508 in seed 1, below 0.01. Its memory benefit reversed sign in reserved validation. No frozen-perception arm changed retention relative to its intact arm. These findings provide no basis for claiming an adaptive winner. Weak effects may reflect sparse exposure, short horizons, scratch-register disruption, or an environment that rarely rewards the measured capabilities.

With no positive adaptive scores, the scheduler used its declared fallback: seed 4 for highest functional diversity, then seeds 3 and 7 for descriptor spread, and seed 1 for deterministic exploration. Reserved validation results did not participate in that decision. Unselected original snapshots remain available.

## Verification

The independent Python verifier reconstructs all 96 intervention-start hashes from source snapshots, recalculates retention from per-tick executable traces, recomputes both score splits, and reproduces the allocation. It also verifies that all four ordinary continuation snapshots match the independent persistent arms of the Stage 15 symbol assay exactly. Normalized states are not needed for this last comparison: the full snapshot hashes agree.

Go tests verify source immutability, repeated probe equality, actual suppression of updated perceptions, threshold and signed-contrast handling, deterministic one-versus-sixteen worker results, preservation of exploration/diversity slots, lack of validation leakage, and ordinary continuation from original sources. Hand-written capability fixtures do not enter the evolutionary source cohort.

```powershell
go run ./cmd/adapt-assay -input data/symbols-stage15 -case symbols -out data/my-adaptivity -workers 16 -ticks 1000 -select 4 -continue-ticks 20000 -every 1000
python experiments/adaptivity/analyze.py --input data/adaptivity-round1-verified
```

Local verified artifacts: `data/adaptivity-round1-verified`. [Compact evidence](evidence.json) records protocol, source and endpoint hashes, all probe summaries and trace hashes, scores and selection reasons. The full per-tick traces remain in local `evaluation.json`. A four-world evidence dossier was prepared at `data/council-adaptivity-round1` without changing the UI or invoking an external AI API.

The next useful experiment should provide changing, structured contexts where retained information can improve outcomes, then measure transfer and acquisition efficiency. Simply lowering the threshold or rewarding code length would not establish greater adaptive capability.
