# Stage 2: diagnosis and a minimal ecology

Date: 2026-09-10. Go 1.26.1, Windows amd64. Kernel `0.1.0`, rules `ecology-2`, snapshot format 2. All quantities below come from executed runs, not an illustrative model.

## Why the original world stopped reproducing

Control runs on a 64×64 grid were extended to 100,000 ticks. Without matter transport, seeds 1, 7, and 42 ended with 11, 6, and 1 particles. The seed 1 control without mutations had no particles left by that point.

In the original seed 1 run, **4082 of 4096** matter units ended up in the region without energy input. Illuminated cells retained 3 units of free matter; another 11 units were held in particles. Only one new copy occurred in the last 50,000 ticks. With diagnostic counters added, the control reproduced the same particle states: 945,615 ALLOCATE attempts lacked available matter, and 34,724 lacked free space.

Mechanism: moving particles carry matter into the dark region, where it remains after they decay. There is no return transport. This is local depletion despite a conserved global supply; the explanation does not require genetic degradation.

The counterfactual change is conservative mixing of matter between neighboring cells every 4 ticks. It adds no material and does not depend on the genome.

| Configuration, 64×64, seed 1 | Particles at tick 100,000 | Genomes | Total copies | New copies in the final window |
| --- | ---: | ---: | ---: | ---: |
| Transport disabled | 11 | 8 | 8,077 | 1 in 50,000 ticks |
| Transport every 4 ticks | 2,481 | 130 | 351,562 | 32,259 in 10,000 ticks |
| Transport enabled, mutations disabled | 2,067 | 1 | 439,429 | 87,671 in 20,000 ticks |

With transport, seed 1 ends with 859 units of free matter in the illuminated region and 756 in the dark region; the remainder is held in particles. The control without mutations shows that restored reproduction does not require a new program to appear. This is strong evidence for the role of matter transport in this configuration, not a universal guarantee that arbitrary rules will remain stable.

## Minimal ecology rules

- Three chemical states X/Y/Z, storing 8/4/0 energy units.
- The energy field charges Z→X. Programs can execute X→Y + 4 energy and Y→Z + 4 energy.
- Conservative transport of matter and chemical states between neighbors.
- TARGET selects a neighbor, TAKE removes energy from it, BIND creates a bond, and UNBIND breaks it. Bonded particles cannot move.
- The world starts with one common seed program that executes both reactions. No predefined specialist populations are added.

An unsuccessful preliminary variant whose seed executed only X→Y went extinct: on a 32×32 grid, seed 1 produced 748 copies, but Y→Z did not appear before the available cycle was depleted. The current variant explicitly gives both reactions to a single seed. This is an initial experimental setup, not independently discovered metabolism.

## Long runs

```powershell
go run ./cmd/sim -width 32 -height 32 -seed 1 -ecology -matter-diffusion 4 -chemical-diffusion 4 -ticks 100000 -every 10000 -save data/eco.json -metrics data/eco.jsonl
```

Repeated with seeds 7 and 42, with 1% mutations on COPY and COPYMEM.

| Seed | Particles | Executable particles | Genomes | Total copies | Copies in the last 10,000 ticks |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 678 | 676 | 26 | 92 401 | 8 980 |
| 7 | 662 | 661 | 30 | 139 121 | 14 293 |
| 42 | 662 | 661 | 48 | 127 086 | 12 776 |

Unlike the original control, copying continues at the end. Both chemical conversions execute in every run. Their presence alone does not demonstrate interdependence between lineages: one program can perform both reactions.

## Two emergent strategies, seed 1

The first lineage (`05b26b88…`) appeared at tick 1339. A mutation replaced COPYMEM with BIND: the program bonds to a new particle, copies code into it, and transfers energy. The bond prevents subsequent movement. At tick 100,000, the lineage occupies 500 particles and has performed 53,549 copies and 53,550 bonding actions; the maximum generation among living particles is 72.

The second (`8c7d401d…`) appeared at tick 66,835. Its 9 instructions retain both reactions, an energy check, movement, ALLOCATE, and COPY. After COPY, execution wraps to the start of the program. COPYMEM and TRANSFER are absent: the reserve provided by ALLOCATE is enough for the offspring to begin acquiring energy independently. At tick 100,000 it has 112 particles, 7,717 copies, and no energy transfers or bonds; its maximum generation is 178.

| Tick | Bonded lineage: particles | Its cumulative copies | Short mobile lineage: particles | Its cumulative copies |
| ---: | ---: | ---: | ---: | ---: |
| 70 000 | 498 | 36 659 | 5 | 59 |
| 80 000 | 493 | 42 278 | 70 | 1 415 |
| 90 000 | 511 | 47 900 | 108 | 4 460 |
| 100 000 | 500 | 53 549 | 112 | 7 717 |

These are distinct heritable copying and spatial behaviors that emerged from one seed. Both lineages reproduce and coexist for more than 30,000 ticks after the second appears. This observation supports the Stage 2 criterion of multiple persistent strategies over the studied horizon; indefinite coexistence has not been established.

## Separate checks

Tests cover energy balance including chemical reserves, material conservation, reaction limits, transport along both axes at different intervals, local energy theft, movement blocked by bonds, and bond cleanup after death. Replay of a chemical world and continued reproduction without mutations under matter transport were also verified.

In an isolated test, a neighboring program obtains energy from Y only after another program produces Y from X and the material moves to the neighbor. This confirms that the interaction mechanism works; it **does not establish the spontaneous emergence of interdependent metabolic lineages**. That stronger Stage B criterion remains open.

Full replay check: seed 1, 32×32, chemistry mode, 101,000 continuous ticks versus 100,000 → snapshot → another 1000 ticks. Both runs ended with the same SHA-256:

```text
fef838f6e140cfd9d9be867dfc3e9cee26a2b87e51f03119eb4323c80ddcf302
```

To reproduce all configurations: `./scripts/experiments.ps1`. Raw snapshots and JSONL are written to `data/`; compact results from the current runs are in [summary.json](summary.json). The old Stage 1 benchmark does not apply to the new physics.
