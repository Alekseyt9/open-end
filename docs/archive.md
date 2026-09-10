# Novelty archive and Pareto selection (Stage 9)

Stage 9 adds an external research selection policy to the [branching-world harness](branching.md). It compares several measured objectives without combining them into a single fitness score. A behavior-cell archive preserves unusual viable directions, including candidates outside the current Pareto front.

World execution is unchanged. Selection affects only which frozen cohorts will receive future computation.

## Run the archive

```powershell
# Analyze and archive a decision without changing the selected branches.
go run ./cmd/council tree archive -tree data/my-tree

# Recompute the decision and save its recommended continuation set.
go run ./cmd/council tree archive -tree data/my-tree -apply

# Continue that set using 16 workers, then archive the new evidence.
go run ./cmd/council tree grow -tree data/my-tree -ticks 20000 -every 1000 -window 10000 -workers 16
go run ./cmd/council tree archive -tree data/my-tree -apply
go run ./cmd/council tree export -tree data/my-tree
```

The archive never starts a simulation. `grow` still follows the saved selection or an explicit `-branches` override. Manual selection remains available.

In the HTML explorer, the archive panel shows:

- measured objectives and retention reasons for every archived cohort;
- a two-axis view of a selected comparison group;
- which tips are on the full Pareto front;
- protected behavior-cell representatives and recommended continuations;
- the historical observations used as novelty references;
- a button that loads recommendations into the command builder;
- controls for preparing archive refresh/apply commands.

A two-axis plot is a projection: the actual frontier uses every available objective. The report marks a decision as outdated when new tree nodes appear and disables its recommendation button until a fresh archive is exported.

## Evidence and comparability

Each archive refresh validates the tree, frozen requests, evidence hashes, and physical snapshots. Cohorts are matched by their sorted case/seed lists, complete physical configuration, actual window duration, sample count, detector configuration, and the four block durations.

**Novelty references** may come from any final tick in this matched context, including historical ancestors. **Pareto dominance** compares only current tips with the same context and the same vector of final ticks. An older world is not declared inferior merely because another had more time to evolve.

Rule modules may differ. The behavior descriptor uses only seven common physical actions whose meaning does not depend on a DSL reaction ID:

```text
allocate, copy, transfer, take, bind, unbind, absorbed_energy
```

Reaction-ID rates, DSL usage counts, rule names, and genome hashes do not enter the novelty descriptor. A new rule name or a new genome hash alone therefore cannot produce behavioral novelty.

## Measured objectives

| Objective, maximized | Definition | Interpretation |
|---|---|---|
| Novelty distance | Mean distance to up to K nearest distinct behavior descriptors in the compatible archive | Descriptive difference, not adaptive benefit |
| Effective diversity | Mean final effective genome count across cohort worlds | Includes neutral genetic variation |
| Bonded scale | Mean particle-weighted log component size | Structural scale proxy, not causal complexity |
| Persistence | Fraction of world/time blocks with code copying and a living executable population at the end | Window-local reproduction proxy |

For each world and common action, take the mean of its rates in the last two detector blocks, then transform it with `log2(1 + rate)`. Concatenate worlds in case/seed order. Distance is mean absolute difference across the resulting descriptor coordinates.

Rates are events, or absorbed energy, per 1000 particle-ticks. The denominator is the existing detector's trapezoidal estimate from sampled population frames. The archive does not claim exact particle exposure.

The default novelty score averages the nearest **3 distinct descriptors**, or fewer when fewer exist. If any compatible cohort has an identical descriptor, novelty is **zero**. Duplicate reference descriptors count once, so rerunning the same trajectory does not inflate the reference density. With no compatible peer, novelty is `null`, not zero; selection then uses the other available objectives.

For a world with population `P` and `count(s)` components of size `s`, bonded scale is:

```text
sum(count(s) × s × log2(s)) / max(1, P)
```

Singletons contribute zero. This does not measure a new level of biological organization.

A world contributes one quarter of a persistence unit for each of its four blocks with a positive copy rate, provided executable particles remain at the end. The cohort value is the mean over worlds. Code copying is not identical to allocating new particles.

**Causal structure, hierarchy depth, new information processing, and adaptive value are explicitly unavailable.** They are neither assigned zero nor inferred from component size. The implemented Stage 9 policy uses available proxies; the stronger causal filters and dimensions in the research plan require later measurement work.

## Admission filters

Defaults:

```text
minimum stable window:        10,000 ticks
minimum surviving worlds:     0.75
minimum productive blocks:    0.75
behavior cell bins/octave:    1
novelty neighbors:            3
```

Every world must have four usable behavior blocks, enough history, and stable rules throughout the analyzed window. Survival means executable particles remain, rather than merely empty allocated particles. Cohort survival and persistence must pass their configured fractions.

Filtered cohorts are retained as evidence with explicit reasons; they do not enter novelty references, the active frontier, or protected viable cells. A decision with no recommendations is archived, but `-apply` fails and leaves the previous selection intact.

Override the recorded policy explicitly:

```powershell
go run ./cmd/council tree archive -tree data/my-tree -min-ticks 10000 -min-survival 0.75 -min-persistence 0.75 -bins 2 -neighbors 3
```

These thresholds are research filters, not goals assigned to organisms.

## Pareto and cell retention

A tip dominates another comparable tip only when it is no worse on all available objectives and strictly better on at least one. Equal vectors and tradeoffs remain. There is no weighted total score.

For coarse behavior cells, average each of the seven descriptor coordinates across cohort worlds and take:

```text
cell coordinate = floor(mean_log_rate × bins_per_octave)
```

Each viable cell keeps its oldest recorded representative. The recommended set contains:

1. All non-dominated eligible tips, evaluated within matched horizons.
2. A representative of every otherwise uncovered viable cell in an active tip's context. Prefer a current tip in that cell; if none remains, recommend the historical representative.

The second rule can retain a globally dominated but unusual direction. It can also reactivate an older branch whose behavior no longer occurs at the tips. Cells represent coarse measurements, not species or discovered ecological niches.

No fixed branch-count cap is imposed. Applying a recommendation may select more than two cohorts; inspect the recorded set before allocating additional compute. The existing worker limit still bounds concurrency.

## Immutable decisions

```text
tree/
  archive/
    a000001/decision.json
    a000002/decision.json
  selection.json
  nodes/...
```

A decision contains its format version, SHA-256 identity, tree-node fingerprint, policy settings, measured points, comparison contexts, novelty references, dominance explanations, protected cells, and recommendation set.

Refreshing identical evidence with identical settings reuses the decision. New evidence or changed settings creates a new revision; previous revisions are not rewritten. Novelty is relative to the reference set and may change as the archive grows. Each revision preserves the result for its own reference set.

Snapshots and dossiers remain in the tree's existing immutable nodes. The archive indexes them rather than duplicating their large files. No archive operation deletes a branch. Writer locking and temporary-file publication follow the tree harness.

[Recorded Stage 9 validation](../experiments/archive/REPORT.md).
