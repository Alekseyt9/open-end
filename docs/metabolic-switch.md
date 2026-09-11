# Experimental cost of metabolic switching

`metabolic_switch_cost` introduces an optional energetic cost for changing between the two baseline chemical reactions. It applies uniformly to all cells; it does not assign cell types, modify genomes or reward larger groups. The initial [120-trial screen](../experiments/metabolic-switch/REPORT.md) did not increase useful specialization, so zero remains the default.

## Physical semantics

`-metabolic-switch-cost` accepts an integer from 0 through 64 and requires ecology for positive values. Positive-cost worlds currently reject DSL rule state and rule installation/scheduling: the mechanism is defined for the two baseline reactions.

Each particle has a `metabolic_state`: 0 means unprepared, 1 means reaction 0 is prepared, and 2 means reaction 1 is prepared. For a valid baseline CONVERT with positive amount:

1. Pay the ordinary instruction cost and advance the instruction pointer as usual.
2. If the particle is unprepared or already prepared for this reaction, no extra energy is charged.
3. Otherwise pay the switching cost. If the remaining reserve is no greater than that cost, dissipate the remaining reserve, leave preparation unchanged and skip the reaction. The cell subsequently dies in normal maintenance.
4. After affordable preparation, record the selected state and attempt the ordinary reaction against the current substrate and energy capacity.

Preparation is paid even if no substrate is available. Repeating the same prepared reaction has no additional switching charge. Invalid reaction IDs and nonpositive amounts do not change preparation. Other instructions do not reset it. This is an anticipatory cost: a cell cannot finance a switch with the energy the blocked reaction would have released.

New allocations begin unprepared. COPY and COPYMEM do not inherit the parent's physiological preparation; genome and memory inheritance otherwise remain unchanged. On introducing the parameter into a previously untreated snapshot, all existing preparation states are zero, because earlier physics did not track them. The first attempted valid reaction is therefore free in all matched arms.

Switching dissipates energy into the existing accounting ledger. Additional counters record successful paid switches, switching starvation events and total switching heat, including partial payment by starved cells. These counters are not rewards or selection inputs.

## Persistence and compatibility

Positive-cost snapshots use **format 9**, storing configuration, per-particle preparation and cumulative switching accounting. The format supports the existing symbol and bond-motion fields. Downgrading such a snapshot is rejected. Loading with `sim` resumes the stored physical cost and preparation; command-line overrides of loaded parameters are prohibited.

At cost zero, preparation remains zero and new zero-valued fields are omitted. Older snapshot formats, hashes, mutation repertoires and physical behavior are retained. Council briefs and tree metadata preserve the new parameter. Reaction DSL proposals cannot be applied while the positive switching-cost model is active.

```powershell
go run ./cmd/sim -width 32 -height 32 -ecology -copy-model evolving -metabolic-switch-cost 2 -ticks 20000 -save data/switch-world.json
go run ./cmd/sim -load data/switch-world.json -ticks 20000 -save data/switch-continued.json
```

## Evolutionary screen

`cmd/switch-assay` clones each selected source into costs 0, 1, 2 and 4. Sources must have baseline ecology, no DSL state, no prior switching cost, no bond-motion treatment and no collective ablation. Genomes, positions, resource distributions and both RNG states start matched. The output follows the existing summary/results/JSONL format, with additional metabolic windows. Output directories must be new.

```powershell
go run ./cmd/switch-assay -input data/environment-stage11 -case environment -out data/my-switch-environment -workers 16 -ticks 20000 -every 1000 -group-age 100
go run ./cmd/switch-assay -input data/symbols-stage15 -case symbols -out data/my-switch-symbols -workers 16 -ticks 20000 -every 1000 -group-age 100
```

Reaction units are accumulated by particle over nonoverlapping 5,000-tick windows, or the shorter remaining interval. A profile requires at least 32 successful reaction units. It is labelled a reaction-0 or reaction-1 specialist if at least 90% of those units use that reaction; other qualifying profiles are generalists. Counts distinguish cells alive at the window end from cells that died. A persistence count requires the same living particle to meet the same specialist criterion in consecutive windows; a low-activity or missing window breaks the sequence.

An endpoint bond component is marked complementary only if it contains at least one living specialist of each kind. This is a descriptive candidate, not proof of useful interdependence, inherited specialization or group reproduction. Metabolic profiles below the activity threshold are not counted as specialists, and total population includes particles without qualifying profiles. A shorter final window is explicitly bounded by its From/To fields and should not be treated as equal exposure to full windows.

## Chemical witness comparison

```powershell
go run ./cmd/role-assay -suite switching -input data/multicell-mixed-witnesses -out data/my-switch-witnesses -workers 16 -ticks 5000 -every 100
```

Protocol 3 continues each preserved witness under the same four global physical costs. It observes selected original-cell COPY lineages with the [proportional chemical tracer](multicell-mechanisms.md). It does not perform the address-specific disabling interventions of the other role suites. An ordinary, unobserved replay with the same cost must reproduce each endpoint exactly.

The cost applies to the **whole world**, including cells outside the selected group. Thus chemical changes include the surrounding ecology's response. Initial source and configured-start hashes are both recorded. Although role-assay endpoint files retain the `released` naming convention, **the physical switching cost persists in their snapshots**. External lineage tags, tracer state and intervention callbacks do not persist; rerunning that observation protocol still starts from its original witness.

The screen and witness study answer different questions: evolution of window-level profiles across source populations, and short-term chemical contributions of related preselected groups. Neither is evidence of long-term inevitability or impossibility of multicellularity.
