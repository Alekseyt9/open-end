# Open-Ended Evolution + External AI Agent
## Concept, architecture, and phased implementation plan

### Current research priority

Prioritize increasing organizational complexity toward multicellularity: persistent collectives, differentiation of useful functions, coordinated behavior, and reproduction that transmits organization. Mere growth in particle count, genome length, or connected-component size is insufficient evidence.

Use external matched experiments to investigate barriers to collective reproduction and the benefits of coordination. Allocate research effort according to evidence of collective heredity and useful differentiation while retaining alternative directions. Memory and perception assays remain supporting diagnostics rather than the sole selection target. Do not install an intelligence score inside organisms or equate longer code with greater intelligence. See [collective reproduction experiments](docs/multicell-motion.md) and [adaptive capability diagnostics](docs/adaptivity.md). Go is the active solver; Warp maintenance is paused.

---

## 0. Core idea

Create a digital world in which evolution is not limited to:

- mutations of a fixed genome;
- predefined species;
- a single fitness function;
- predefined entities such as `Agent`, `Species`, `Predator`, `Weapon`, or `Economy`;
- a fixed set of world laws.

The project's central idea:

> Local evolution takes place inside the simulation, while an external AI agent receives world telemetry and can propose changes to code, rules, primitives, inheritance mechanisms, and even the world's ontology.

The result is a coevolving system:

```text
world
↓
new structures / problems / stagnation emerge
↓
an external AI analyzes the state
↓
it proposes rule macromutations
↓
parallel world branches are created
↓
the simulation itself tests the changes
↓
successful branches are retained
↓
new mechanisms create new opportunities and selection pressures
↓
the cycle repeats
```

The central research question:

> Can a combination of local Darwinian evolution, a coevolving environment, and an external AI agent capable of changing the rule space sustain long-term growth in adaptive novelty?

---

# 1. What counts as Open-Ended Evolution?

OEE should not be confused with ordinary procedural generation.

Procedural generation:

```text
AI invents a new biome
AI invents a new monster
AI invents a new weapon
```

Open-ended evolution:

```text
a new mechanism appears
↓
it becomes part of the world
↓
other entities start using it
↓
selection pressures change
↓
new types of interaction emerge
↓
new levels of organization appear
↓
this opens another space of possibilities
```

Strong OEE is more than:

```text
x ∈ S
```

where evolution searches for new points within a fixed space `S`.

It is closer to:

```text
S0 → S1 → S2 → S3 → ...
```

where the space of possible states itself changes.

Example:

```text
S0 = physics
S1 = physics + replication
S2 = S1 + organisms
S3 = S2 + communication
S4 = S3 + symbols
S5 = S4 + culture
S6 = S5 + technology
S7 = S6 + new computational worlds
...
```

---

# 2. Overall architecture

The system is best divided into two layers.

```text
┌────────────────────────────────────────────┐
│              IMMUTABLE KERNEL              │
│                                            │
│ time                                      │
│ memory                                    │
│ sandbox                                    │
│ resource limits                           │
│ snapshots                                  │
│ rollback                                   │
│ module loading                            │
│ versions                                  │
│ branch management                          │
│ deterministic replay                      │
└────────────────────┬───────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────┐
│              MUTABLE UNIVERSE              │
│                                            │
│ physics rules                              │
│ resource rules                             │
│ chemistry-like rules                       │
│ executable matter                          │
│ inheritance                                │
│ reproduction                               │
│ signaling                                  │
│ memory                                     │
│ social mechanisms                          │
│ emergent abstractions                      │
│ AI-generated primitives                    │
└────────────────────────────────────────────┘
```

The kernel must remain immutable.

Almost everything specific to the world can potentially change.

---

# 3. Why the kernel must be immutable

Allowing an external AI to change the entire running process would:

- make experiments difficult to reproduce;
- make causality difficult to establish;
- prevent guaranteed rollback;
- risk damaging the runtime itself;
- blur the distinction between world evolution and a developer rewriting the program.

The kernel should define only:

```text
what counts as computation
how resources are allocated
how state is stored
how a patch is executed
how a new branch is created
how execution is bounded
how a past experiment is reproduced
```

The kernel should not have prior knowledge of:

```text
Species
Predator
Food
Language
Culture
Economy
Government
Technology
```

---

# 4. Minimal digital physics

The initial world should be very simple.

Primitives:

```text
Matter
Energy
Position
Connection
Signal
Memory
ExecutableRule
```

Minimal operations:

```text
move
bind
unbind
transfer_energy
store_energy
sense
emit_signal
copy
execute
spawn_structure
destroy
transform
```

Instead of predefined organisms, there should be structures that can assemble from these primitives.

---

# 5. Executable Matter

One of the central ideas:

> Information inside the world must have physical effects.

A structure can therefore store a program:

```text
LOAD
SENSE
COMPARE
JUMP
COPY
BIND
SEND
EXEC
```

Such a program can:

- read its local environment;
- change state;
- move matter;
- control energy flows;
- copy itself;
- transmit signals;
- create new structures;
- execute other programs.

A genome is then an executable program, rather than merely a set of parameters.

---

# 6. DSL / virtual machine

A safe DSL is preferable to directly modifying Go/C++ code.

Example:

```text
rule Photosynthesis {
    when:
        light > 0
        CO2 > 0

    consume:
        light 1
        CO2 1

    produce:
        stored_energy 0.7
}
```

Another example:

```text
rule Phototaxis {
    sensor:
        light_gradient

    action:
        move toward light_gradient
}
```

Advantages:

- sandbox;
- hot reload;
- rollback;
- versioning;
- fast execution of many branches;
- the ability to evolve code;
- straightforward rule generation by an external AI.

---

# 7. Evolution of new primitives

A particularly important mechanism:

> The system must be able to create new reusable building blocks.

For example:

```text
FOO :=
    LOAD
    COPY
    BIND
```

Once `FOO` appears, it becomes a new primitive.

Later:

```text
BAR :=
    FOO
    SIGNAL
    EXEC
```

This produces:

```text
primitive
↓
combination
↓
new primitive
↓
combination
↓
higher-level primitive
↓
...
```

This is one possible mechanism for genuine growth in complexity.

---

# 8. Three levels of evolution

## 8.1. Microevolution

This happens continuously inside the world:

```text
mutation
recombination
copy errors
resource competition
inheritance
local selection
```

AI is not needed here.

---

## 8.2. Macroevolution of mechanisms

When the world:

- stagnates;
- loses diversity;
- converges on one dominant strategy;
- experiences a mass extinction;
- develops an unusual persistent structure;

the external agent receives an aggregated description.

It proposes changes such as the following, rather than merely interesting content:

```text
sexual reproduction
horizontal transfer
symbiosis
new sensor
new binding mechanism
chemical memory
new signaling channel
new resource cycle
multicellular attachment
new mutation operator
```

---

## 8.3. Evolution of ontology

The most ambitious level.

The system may detect a new stable category emerging in the world.

For example, the starting point was:

```text
Matter
Energy
Connection
Signal
```

But over time the following appeared:

```text
persistent signal sequences
+
pattern storage
+
reuse
+
contextual interpretation
```

AI can propose a new abstract primitive:

```text
Symbol
```

Symbols can then be:

```text
copy
combine
store
transmit
interpret
transform
```

This may eventually lead to:

```text
signals
↓
symbols
↓
language
↓
culture
↓
institutions
↓
technology
```

---

# 9. Evolution of evolvability

The system should be able to change both genomes and the mechanisms of evolution themselves.

Potentially variable parameters:

```text
mutation rate
mutation operators
genome structure
recombination rules
inheritance encoding
copy fidelity
development process
horizontal gene transfer
error correction
reproduction strategy
```

In other words:

```text
evolution
↓
evolution of evolution
```

---

# 10. Major Evolutionary Transitions

Ideally, the system should be able to cross new levels of individuality on its own:

```text
replicators
↓
cooperative replicators
↓
colonies
↓
stable multicellular entities
↓
specialization
↓
communication
↓
shared memory
↓
collective decision making
↓
culture-like inheritance
```

The key criterion:

> A new unit of selection must emerge inside the world, rather than being predefined as a class.

---

# 11. Environmental coevolution

Do not use a fixed fitness function:

```text
fitness = f(agent)
```

Instead:

```text
fitness_t = f(
    agent,
    environment_t,
    population_t,
    history_t
)
```

The environment changes too:

```text
environment_{t+1}
=
g(
    environment_t,
    population_t
)
```

Example:

```text
photosynthesis appears
↓
the atmosphere changes
↓
new niches emerge
↓
a new metabolism evolves
↓
it changes the environment again
```

The key principle:

> Solving one evolutionary problem should create new problems.

---

# 12. The external AI agent

AI should not act as a god designing finished life forms.

A better role is:

```text
observe
↓
detect novelty / stagnation
↓
form hypothesis
↓
propose macro-mutation
↓
generate rule patch
↓
spawn branches
↓
compare outcomes
↓
retain promising branches
```

AI expands the space of possibilities.

Evolution determines whether a new mechanism will be used.

---

# 13. What AI receives

Hierarchical telemetry, rather than the entire world.

## Basic metrics

```text
population count
birth/death rates
energy distribution
resource consumption
average lifespan
genetic diversity
behavioral diversity
```

## Structural metrics

```text
persistent structures
graph complexity
modularity
hierarchy depth
community structure
```

## Information metrics

```text
entropy
mutual information
predictive information
signal reuse
memory usage
information flow
```

## Evolutionary metrics

```text
novelty rate
innovation persistence
extinction rate
new niches
new inheritance mechanisms
candidate major transitions
```

## Stagnation metrics

```text
diversity plateau
complexity plateau
repeated strategies
dominant monoculture
lack of new interactions
```

---

# 14. Request format for the external agent

Example:

```json
{
  "world_id": "W-1042",
  "generation": 850000,
  "summary": {
    "population": 1240000,
    "diversity": 0.18,
    "novelty_rate": 0.003,
    "stagnation_score": 0.87
  },
  "dominant_patterns": [
    "resource-hoarding colonies",
    "short-range chemical signaling"
  ],
  "emergent_structures": [
    "persistent 6-12 entity clusters"
  ],
  "constraints": {
    "max_new_primitives": 2,
    "max_cpu_overhead_pct": 5
  }
}
```

---

# 15. Agent response format

AI should return a structured patch proposal.

For example:

```json
{
  "hypothesis": "Persistent clusters may exploit specialized signaling",
  "changes": [
    {
      "type": "add_rule",
      "name": "directed_signal",
      "cost": 0.05
    }
  ],
  "expected_effects": [
    "higher coordination diversity",
    "possible specialization"
  ],
  "risks": [
    "single signaling strategy may dominate"
  ]
}
```

---

# 16. Patch Lifecycle

Each change:

```text
1. detect trigger
2. produce world summary
3. ask AI for hypotheses
4. generate N patches
5. static validation
6. sandbox compile
7. short smoke simulation
8. multi-seed test
9. compare metrics
10. archive all results
11. promote / retain / discard
```

---

# 17. AI as a generator of architectural mutations

Ordinary mutation:

```text
0.31 → 0.34
```

An LLM can:

```text
independent cells
↓
energy exchange via signals
↓
cooperation
↓
stable colony
↓
specialization
```

Thus, sometimes:

```text
Δparameter
```

is replaced by:

```text
Δarchitecture
```

This could be one of the main advantages of LLMs in an OEE system.

---

# 18. A tree of worlds

Do not select a single best branch.

```text
                    World 0
                 /     |     \
               W1      W2      W3
              /  \             / \
            W4    W5          W6  W7
                    \
                     W8
```

Retain different worlds according to different criteria:

```text
maximum novelty
maximum diversity
new cooperation
new hierarchy
new representation
new ecology
maximum persistence
```

---

# 19. Novelty Archive

Store discoveries as well as current winners:

```text
new structures
new functions
new signals
new interaction types
new inheritance mechanisms
new resource loops
new levels of organization
new representations
```

---

# 20. Measuring novelty

Ask more than whether the world became more efficient.

Possible criteria:

```text
behavioral novelty
structural novelty
functional novelty
causal novelty
ecological novelty
representation novelty
information-processing novelty
```

A particularly strong signal:

> A new structure enables a type of interaction that did not previously exist.

---

# 21. Causal Emergence

Causal emergence can be used as an analytical tool.

The idea:

> Find a coarse-graining at which the macrodescription has greater causal predictive power than the microdescription.

A potential sequence:

```text
particles
↓
persistent machine
↓
organism
↓
colony
↓
society
```

This could provide a way to discover new levels of organization automatically.

---

# 22. Viability / Free Energy Perspective

An organism can be defined as a system that keeps itself within a region of viable states, rather than as a class.

```text
state(t) ∈ viability region
```

Internal constraints:

```text
energy reserves
integrity
temperature-like state
resource balance
signal coherence
```

The system can change the world to remain viable.

---

# 23. Physics of Life

The world must be out of equilibrium.

Continuous flows are needed:

```text
energy source
↓
energy capture
↓
work
↓
waste / dissipation
```

If the energy gradient disappears, complexity should degrade.

---

# 24. Cultural evolution

Once complex communication emerges, a separate inheritance channel can be introduced:

```text
ideas
skills
symbols
recipes
strategies
construction programs
```

There are then:

```text
genetic inheritance
+
cultural inheritance
```

And over time:

```text
genetic evolution
↕
cultural evolution
```

---

# 25. External artifacts

Creatures should potentially be able to create persistent external objects.

Do not hardcode concepts such as:

```text
nest
tool
storage
trap
machine
```

Let these simply be persistent physical structures.

If they acquire functional uses, their semantics can be discovered later.

---

# 26. Technological evolution

A particularly interesting transition:

```text
tool
↓
tool-making tool
↓
machine
↓
factory-like system
↓
programmable machine
```

If such levels emerge without a predefined technology tree, this is a strong sign of open-endedness.

---

# 27. Internal virtual machines

A radical idea:

> Creatures inside the world can build their own VM.

For example:

```text
base physics
↓
organism builds VM-A
↓
VM-A defines symbolic language
↓
new digital entities exist in VM-A
↓
they build VM-B
↓
...
```

Analogy:

```text
biology
↓
brains
↓
language
↓
writing
↓
mathematics
↓
computers
↓
AI
```

Each layer creates a new space for evolution.

---

# 28. Self-modifying code

Allow changes to:

```text
behavior programs
reproduction programs
inheritance programs
mutation logic
communication
memory representation
world rules
resource cycles
local physics modules
```

But not to the immutable kernel.

---

# 29. AI patch safety

AI-generated code must run:

- without host OS access;
- without arbitrary filesystem access;
- without network access;
- with a CPU limit;
- with a RAM limit;
- with an instruction budget;
- with a timeout;
- with deterministic replay wherever possible.

Prefer:

```text
custom DSL
or
WASM-like sandbox
```

rather than native-code hot patching.

---

# 30. Reproducibility

Each change records:

```text
world_id
parent_world_id
seed
kernel_version
rule_version
AI_patch
AI_hypothesis
metrics_before
metrics_after
branch_result
```

Any branch can be reproduced.

---

# 31. Service architecture

```text
World Runner
    ↓
Telemetry Collector
    ↓
State Analyzer
    ↓
Novelty Detector
    ↓
Stagnation Detector
    ↓
AI Evolution Agent
    ↓
Patch Generator
    ↓
Patch Validator
    ↓
Branch Manager
    ↓
Experiment Evaluator
    ↓
Novelty Archive
```

---

# 32. Possible repository layout

```text
/cmd
  /world-runner
  /orchestrator

/internal
  /kernel
  /world
  /vm
  /dsl
  /physics
  /telemetry
  /evolution
  /novelty
  /branching
  /archive
  /agent
  /evaluation

/pkg
  /protocol

/experiments
  /baseline
  /macro_mutation
  /ontology_expansion

/web
  /viewer

/data
  /snapshots
  /archives
```

---

# 33. Technology stack

For a quick MVP:

```text
Simulation core: Go
AI orchestrator: Go / Python
Rule language: custom DSL
VM: custom bytecode or WASM
Storage: SQLite initially
Telemetry: Parquet later
Visualization: Web UI
Branching: processes / containers
```

If the simulation becomes CPU-bound:

```text
core → Rust/C++
```

GPU acceleration makes sense at a later stage.

---

# 34. Implementation stages

---

## Stage 0. Research framework

### Goal

Build a minimal reproducible world.

### Implement

- tick-based simulation;
- seeded RNG;
- snapshot;
- replay;
- a basic map;
- basic energy;
- resource accounting;
- benchmark.

### Completion criterion

The same snapshot and seed produce the same continuation.

---

## Stage 1. Executable replicators

### Add

```text
executable genome
copy
mutation
death
resource consumption
```

### Do not use

```text
Species
Predator
FitnessScore
```

### Criterion

Different heritable lineages emerge.

---

## Stage 2. Mini-ecology

Add:

```text
multiple resources
local competition
the ability to take resources
bonds between structures
```

### Criterion

At least several persistent strategies emerge.

---

## Stage 3. DSL / VM

Move mutable world rules into a DSL.

### Implement

- parser;
- bytecode;
- instruction budget;
- sandbox;
- rule versioning;
- hot reload;
- rollback.

### Criterion

A new rule can be loaded without restarting the kernel.

---

## Stage 4. Telemetry

Collect:

```text
population
energy
lineages
diversity
lifetimes
resource flows
structure sizes
interaction graph
```

### Criterion

Automatically explain what happened in the world over the last N thousand ticks.

---

## Stage 5. Novelty + Stagnation Detector

Start with simple heuristics:

```text
diversity plateau
population monoculture
repeated behavioral hashes
structural plateau
```

Add embedding-based or graph-based novelty later.

### Criterion

The system can automatically report:

```text
"the world is developing"
or
"the world is stuck"
```

---

## Stage 6. AI Observer

AI does not change anything yet.

It receives a telemetry summary and explains:

```text
what dominates
which niches exist
which structures have appeared
why stagnation may have occurred
```

### Criterion

AI conclusions can be compared with actual logs and visualizations.

---

## Stage 7. AI Macro-Mutations

Allow the agent to propose new rules.

For example:

```text
new sensor
new resource reaction
new binding rule
new communication channel
```

### Main rule

AI does not create a finished adaptation.

Bad:

```text
give every creature wings
```

Good:

```text
add a physical mechanism for lift
```

and see whether evolution makes use of it.

---

## Stage 8. Branching Worlds

For each AI patch:

```text
N variants × M seeds
```

Run in parallel.

### Evaluate

```text
novelty
diversity
persistence
structural complexity
niche count
```

### Criterion

The system retains several promising branches instead of a single winner.

---

## Stage 9. Novelty Archive + Pareto Selection

Do not reduce everything to a single fitness function.

Maintain a Pareto front over:

```text
novelty
diversity
complexity
persistence
new hierarchy
new information processing
```

### Criterion

Rare, unusual branches do not disappear merely because they are currently less efficient.

---

## Stage 10. Evolution of evolvability

Allow genomes to change:

```text
mutation rate
mutation operators
recombination
copy fidelity
inheritance format
```

### Criterion

Different lineages use different strategies for their own variability.

---

## Stage 11. Environmental coevolution

Creatures gain the ability to substantially change the world:

```text
resource distribution
local chemistry
energy gradients
terrain-like state
signal fields
```

### Criterion

A new adaptation changes the environment in ways that create new niches.

---

## Stage 12. Proto-multicellularity

Add only minimal mechanisms:

```text
persistent binding
resource sharing
signals
local specialization
```

Do not introduce a `MulticellularOrganism` class.

### Criterion

Some groups persist better than individual replicators and begin reproducing as wholes.

---

## Stage 13. Automatic Entity Discovery

Look for:

```text
persistent boundaries
internal coordination
resource sharing
common reproduction
information closure
```

### Result

A hierarchy:

```text
micro entity
↓
compound entity
↓
collective
↓
higher-level individual
```

---

## Stage 14. Causal Emergence Analysis

Try to find a macrolevel that better predicts the system's future.

### Goal

Automatically distinguish:

```text
a random group
```

from:

```text
a new causally significant entity
```

---

## Stage 15. Symbolic layer

Allow:

```text
persistent signals
signal composition
memory
arbitrary mapping
context-sensitive interpretation
```

### Criterion

Signal structures emerge whose meaning is not determined solely by the physical stimulus.

---

## Stage 16. Cultural Evolution

Add horizontal copying of:

```text
behavior programs
skills
symbol sequences
construction recipes
```

### Criterion

Information spreads independently of genetic lineage.

---

## Stage 17. Persistent Artifacts

Allow long-lived external structures.

### Criterion

Objects emerge that:

- outlive their creators;
- are used by others;
- become part of an adaptive strategy.

---

## Stage 18. Technology-like Evolution

Allow constructions to create constructions.

### Sequence

```text
tool
↓
tool-making tool
↓
machine
↓
construction network
```

### Criterion

External artifacts begin creating a new space of evolutionary possibilities.

---

## Stage 19. Self-Hosted VM

The most ambitious stage.

Creatures must be able to implement:

```text
memory
instruction encoding
decoder
execution loop
```

### Criterion

A new computational layer appears in the world on top of the original VM.

---

## Stage 20. Recursive Evolution

If an internal VM has appeared:

```text
world
↓
VM-A
↓
new programmable entities
↓
VM-B
↓
...
```

Investigate the possibility of:

```text
evolution inside evolution
```

---

## Stage 21. Strong OEE Experiment

Long-running experiments.

Monitor:

```text
does novelty saturate?
does complexity saturate?
do new abstractions continue appearing?
do new levels of organization emerge?
does evolvability continue changing?
```

The main criterion:

> Growth in diversity and adaptive novelty does not stop at an obvious ceiling directly imposed by the developer.

---

# 35. Main comparative experiment

Implement three modes.

## A. Control

```text
ordinary internal evolution
without AI
```

## B. Parameter AI

AI changes only numbers:

```text
mutation rate
resource rate
energy cost
```

## C. Structural AI

AI can create:

```text
new rules
new interaction primitives
new inheritance mechanisms
new representations
```

Compare:

```text
novelty over time
time to stagnation
structural complexity
persistent innovations
niche count
major transitions
```

If C consistently outperforms A/B, there is a strong reason to continue developing the approach.

---

# 36. Success criteria

## Minimal

- the world persists for a long time;
- lineages compete;
- diversity does not disappear immediately;
- unexpected strategies emerge.

## Intermediate

- new ecological niches;
- new inheritance mechanisms;
- cooperation;
- parasitism;
- persistent collectives.

## Strong

- new levels of individuality;
- symbol-like representations;
- cultural inheritance;
- persistent artifacts;
- new computational abstractions.

## Very strong

- self-created VM;
- new languages;
- repeated major transitions;
- sustained novelty without saturation.

---

# 37. What not to build in the MVP

Do not start with:

- Unreal Engine;
- 3D;
- realistic fluid physics;
- attractive graphics;
- an LLM for every NPC;
- politics;
- economics;
- a huge world;
- millions of agents;
- CUDA.

The main unknown:

> Does interesting open-ended dynamics emerge at all?

Establish that first.

---

# 38. The smallest MVP

```text
2D grid
+
energy
+
resources
+
executable genomes
+
mutation
+
copy
+
death
+
custom DSL
+
snapshots
+
telemetry
+
AI rule patches
+
parallel branches
+
novelty archive
```

Goal:

> Test whether an external AI can systematically move evolution out of local dead ends without directly specifying the final solution.

---

# 39. MVP Milestones

### M1 — Kernel
- [ ] deterministic simulation loop
- [ ] seeded RNG
- [ ] snapshot/replay
- [ ] resource accounting

### M2 — Digital life
- [ ] executable genome
- [ ] copy
- [ ] mutation
- [ ] death
- [ ] competition

### M3 — Evolution analytics
- [ ] lineage tracking
- [ ] diversity
- [ ] novelty
- [ ] stagnation

### M4 — DSL
- [ ] parser
- [ ] VM
- [ ] sandbox
- [ ] hot reload
- [ ] rollback

### M5 — External agent
- [ ] telemetry summary
- [ ] patch request
- [ ] patch validation
- [ ] branch run

### M6 — World tree
- [ ] branch graph
- [ ] Pareto archive
- [ ] multi-seed experiments

### M7 — Macro evolution
- [ ] new rules
- [ ] new primitives
- [ ] new mutation mechanisms

### M8 — Major transition prototype
- [ ] persistent binding
- [ ] resource sharing
- [ ] collective reproduction
- [ ] emergent entity detection

---

# 40. Visualization

Start with a scientific UI.

Display:

```text
world map
lineage tree
world branch tree
energy flows
interaction graph
novelty timeline
diversity timeline
complexity timeline
major transition events
AI patch history
```

An ontology-history UI is particularly interesting:

```text
S0
↓
S1
↓
S2
↓
...
```

with explanations such as:

```text
"persistent signaling first appeared here"
"a new inheritance channel appeared here"
"the collective became a unit of selection here"
```

---

# 41. A game version

After the research prototype, the system could become a game.

The player's role:

```text
observer
scientist
curator
explorer
intervention agent
```

The most interesting option:

> The player physically exists inside the world while the world continues its own OEE around them.

Playthroughs could then differ radically:

```text
World A → symbiotic super-organisms
World B → mobile colonies
World C → machine ecology
World D → symbolic civilization
```

---

# 42. Game metamechanics

The player does not have to win.

They can:

- discover new life forms;
- explore unusual evolutionary branches;
- intervene;
- compare parallel worlds;
- reconstruct the causes of a major transition;
- travel between world branches;
- preserve rare lineages;
- run controlled perturbations.

In effect:

> No Man's Sky, but generating new laws, ecologies, and levels of organization as well as world forms.

---

# 43. The most ambitious version of the idea

The initial developer creates only:

```text
space
matter
energy
connections
signals
execution
resource limits
```

Then observes:

```text
replication
↓
parasites
↓
cooperation
↓
multicellularity
↓
communication
↓
symbols
↓
culture
↓
technology
↓
new computation layers
↓
???
```

The final `???` is the project's central outcome.

If everything that can appear is known in advance, this is a weak form of OEE.

If the system starts producing categories that the developer did not directly design, the experiment becomes truly interesting.

---

# 44. Overall diagram

```text
MICRO EVOLUTION
millions of inexpensive local changes
        ↓
WORLD DYNAMICS
ecology + environmental change
        ↓
NOVELTY / STAGNATION DETECTOR
        ↓
EXTERNAL AI MACRO-EVOLUTION AGENT
        ↓
semantic / architectural mutations
        ↓
PARALLEL WORLD BRANCHES
        ↓
archive + Pareto selection
        ↓
new primitives / new ontology
        ↓
new evolutionary possibilities
        └──────────────↺
```

The main principle:

> The external AI does not design finished life forms. It expands the space of possible mechanisms, while natural selection inside the simulation determines what actually persists.

---

# 45. First practical sprint

If implementation starts now, the first sprint can be limited to the following.

## 1. Kernel

```text
World
Cell
Energy
RuleVM
Tick
Snapshot
```

## 2. Replicator

Minimal genome:

```text
SENSE_RESOURCE
MOVE
ABSORB
COPY
```

## 3. Mutations

```text
replace instruction
insert instruction
delete instruction
duplicate block
```

## 4. Metrics

```text
population
genome hash
lineage
energy
birth/death
behavior hash
```

## 5. Stagnation

If:

```text
novelty_rate < threshold
```

over a long window:

```text
trigger external agent
```

## 6. First AI patch

AI may add exactly one new DSL primitive.

For example:

```text
SEND_SIGNAL
```

Create:

```text
control branch
+
patch branch
```

and compare them.

This would already be a minimal complete experiment of the whole idea.

---

# 46. The next major milestone

After the first working loop:

```text
world
→ stagnation
→ AI patch
→ branch
→ selection
```

the next truly significant milestone is:

> Obtain the first new persistent interaction that was not the intended final outcome of the AI patch.

For example:

AI adds signaling to encourage cooperation,

but evolution uses it for:

```text
parasitic imitation
or
deception
or
territorial marking
```

That result would be much more interesting than the world using the mechanism exactly as the agent intended.

---

# 47. Ultimate research goal

Ideally, the system should move from:

```text
"AI invents new rules"
```

to:

```text
"AI notices that evolution has approached a new level on its own,
and merely makes that level accessible to further exploration"
```

The external agent gradually changes from a designer into:

```text
an observer
+
a hypothesis generator
+
an operator that expands the possibility space
```

This architecture is particularly interesting as a possible bridge between:

- classical evolutionary computation;
- artificial life;
- open-ended evolution;
- AI coding agents;
- causal emergence;
- digital physics;
- cultural evolution;
- self-modifying software.


---

# 48. Clarification: the world has no "problems," only consequences

The base model should not contain an object or attribute such as:

```text
Problem
```

Evolution itself does not consider anything a problem.

Inside the world, there are only:

```text
environmental changes
↓
changes in the probability of structures persisting
↓
changes in the probability of replication
↓
changes in population composition
```

For example:

```text
resources become scarcer
↓
some replicators disappear
↓
others persist
```

For the world, this is not a "resource shortage problem."

It is simply dynamics.

The notion of a problem arises only for an external observer.

It is therefore more accurate to replace:

```text
evolution solves a problem
```

with:

```text
conditions changed
↓
selection pressures changed
↓
some structures proved more persistent
```

---

# 49. Minimal directionality without a fitness function

Natural selection does not require an explicit function:

```text
fitness(x)
```

Three properties are sufficient:

```text
variation
+
inheritance
+
differential persistence / reproduction
```

If:

```text
A exists for 10 ticks and disappears

B copies itself

C copies itself even more efficiently
```

then, after a long time, descendants of B/C remain while A does not.

Nobody specified:

```text
goal = reproduce
```

Directionality emerges statistically:

> Structures that leave no causal continuations cease to be present in future world states.

---

# 50. Fundamentally, all states are neutral

From the perspective of basic physics:

```text
an organism continues to exist
```

and:

```text
an organism disintegrates
```

are simply two possible states.

The universe assigns neither:

```text
good
bad
success
failure
```

This means:

> Survival should not be a built-in moral value or objective function of the world.

It emerges as an observable statistical asymmetry.

We observe long-lived structures precisely because short-lived ones have already disappeared.

---

# 51. Affordances rather than "problems"

A more useful concept for OEE:

```text
affordance
```

An affordance is an opportunity for interaction that exists relative to a particular structure.

For example, the same object can be:

```text
for organism A → an energy source
for organism B → a toxin
for organism C → a signal
for organism D → neutral background
```

There is no:

```text
environment.problem = X
```

Instead, there is:

```text
Affordance(agent, environment)
```

New world dynamics creates new affordances.

Example:

```text
organism A produces substance X
↓
X accumulates
↓
organism B happens to be able to use X
↓
a new ecological niche emerges
↓
B changes the environment
↓
new affordances arise
```

Nobody declared:

```text
Problem: use X
```

---

# 52. Meaning emerges relative to a structure

Before phototaxis appears:

```text
a light gradient
```

is simply a physical quantity.

Once the corresponding mechanism appears:

```text
a light gradient
```

becomes information.

Before copying machinery appears:

```text
a sequence
```

is simply a structure.

Once inheritance appears:

```text
a sequence
```

becomes genetic information.

Semantics therefore emerges as a relation, rather than as an intrinsic property of the environment:

```text
structure ↔ environment
```

---

# 53. The external AI should not be a Problem Solver

The original architecture should be revised.

Bad:

```text
world has problem
↓
AI finds solution
```

Because this introduces hidden teleology.

Better:

```text
world develops new dynamics
↓
AI observes
↓
AI expands possibility space
↓
evolution exploits or ignores new mechanisms
```

The role of AI:

```text
AI Possibility Expander
```

rather than:

```text
AI Problem Solver
```

---

# 54. Neutral extensions of the world

AI can periodically be asked to create:

```text
new local interaction primitive
new energy transformation
new signaling mechanism
new binding property
new state transition
new storage mechanism
```

While prohibiting requests such as:

```text
"help species X"
"solve problem Y"
"make the organism stronger"
```

A good prompt:

```text
Propose a minimal local extension to the world's physics
that:
- has a cost;
- gives no lineage a direct advantage;
- allows several potential uses;
- does not prescribe a final function;
- can be used or ignored by evolution.
```

---

# 55. The world should guide its own expansion

An even stronger approach:

AI does not generate random new mechanisms.

It looks for structures that have already begun emerging from the bottom up.

For example:

```text
thousands of local bonds
↓
a recurring persistent pattern emerges
↓
AI detects a latent structure
↓
it proposes making that structure a new primitive
```

Example:

```text
a complex recurring boundary
↓
candidate "membrane"
↓
Membrane becomes an optimized primitive
↓
evolution starts building structures from Membrane
```

Later:

```text
Membrane + metabolism + replication
↓
candidate higher-level unit
↓
a new primitive
```

---

# 56. Emergent Ontology Compilation

This idea can be formalized as:

```text
emergent ontology compilation
```

Diagram:

```text
microdynamics
↓
a persistent recurring macrostructure
↓
detection of causal / functional coherence
↓
AI proposes an abstraction
↓
the abstraction is compiled into a new primitive
↓
the new primitive becomes a building block
↓
even more complex structures emerge
```

In other words:

```text
S0 → S1 → S2 → S3
```

where each successive world language is built partly from persistent regularities in the previous one.

---

# 57. An external metacriterion is still needed

If many world branches are allowed, some will:

```text
degrade
simplify
become chaotic
lock into a single pattern
lose diversity
produce random noise
```

The research layer therefore needs an evaluation method.

But it should not be a single fitness function such as:

```text
score = complexity
```

Otherwise, the entire system will start optimizing our measurement.

What is needed:

```text
multi-objective evaluation
+
archive
+
minimum viability filters
+
novelty detection
```

---

# 58. Why "complexity" alone is insufficient

A system can be enormously complex without exhibiting interesting evolution.

For example:

```text
random noise
```

can have:

```text
high entropy
low compressibility
many different states
```

but almost no persistent structure.

Conversely:

```text
a perfect crystal
```

is highly structured but produces almost no novelty.

The desired region therefore lies between:

```text
rigid order
↔
adaptive structured novelty
↔
random chaos
```

---

# 59. Three classes of degradation

## 59.1. Collapse

```text
population → 0
energy loops disappear
persistent structures disappear
```

## 59.2. Freezing

```text
one strategy dominates
diversity → low
novelty → 0
world repeats itself
```

## 59.3. Chaotization

```text
entropy high
structures short-lived
inheritance low
causal persistence low
```

All three regimes may be undesirable for an OEE experiment.

---

# 60. A world-quality vector instead of a single score

For each branch, compute:

```text
Q(world) = [
    viability,
    diversity,
    persistence,
    novelty,
    adaptive_novelty,
    structural_complexity,
    functional_complexity,
    causal_structure,
    hierarchy_depth,
    niche_creation,
    evolvability,
    ontology_growth
]
```

Do not immediately reduce this vector to a single number.

---

# 61. Viability

Indicates whether a sufficiently persistent process exists at all.

Possible measures:

```text
population persistence
energy throughput
reproduction continuity
mean lineage duration
fraction of persistent structures
```

This is a baseline filter.

If:

```text
viability < minimum
```

the branch need not be developed further.

---

# 62. Diversity

Several types of diversity:

```text
genetic diversity
behavioral diversity
structural diversity
ecological diversity
functional diversity
```

Random differences alone should not count as diversity.

---

# 63. Persistence

Novelty must persist long enough.

For example, a new pattern that existed for:

```text
3 ticks
```

is more likely to be noise.

Whereas a pattern that:

```text
appeared
↓
spread
↓
affected other structures
↓
persisted for 100000 ticks
```

is much more interesting.

One possible definition:

```text
persistence_score =
lifetime × descendants × ecological_impact
```

---

# 64. Raw Novelty

Measures:

> How much the new state differs from the archive of past states.

Possible approaches:

```text
behavior embeddings
graph embeddings
genome embeddings
structure descriptors
interaction graphs
```

For example:

```text
novelty(x)
=
distance(
    descriptor(x),
    nearest historical descriptors
)
```

However, raw novelty can easily reward meaningless noise.

---

# 65. Adaptive Novelty

A stronger metric is therefore needed:

```text
adaptive novelty
```

A new structure is interesting if it:

1. has not been observed before;
2. persists;
3. affects its own persistence or reproduction;
4. changes interactions with other structures;
5. potentially creates new affordances.

Conceptually:

```text
AdaptiveNovelty
=
Novelty
× Persistence
× FunctionalImpact
```

This is an idea, not a final formula.

---

# 66. Functional Complexity

Distinguish:

```text
structural complexity
```

from:

```text
the complexity of causally useful organization
```

For example, measure:

```text
number of interacting subsystems
dependency graph depth
division of labor
number of coordinated functions
conditional behavior complexity
```

---

# 67. Causal Structure

An important filter against noise.

A complex system should have:

```text
persistent causal dependencies
```

rather than merely random variation.

Possible metrics:

```text
predictive information
effective information
transfer entropy
causal graph stability
macro-level predictive gain
```

A complex structure whose parts have little persistent influence on one another may simply be noise.

---

# 68. Hierarchy Depth

Open-ended evolution is particularly interesting when new levels appear:

```text
primitive
↓
module
↓
entity
↓
collective
↓
society-like unit
```

Possible measures:

```text
number of stable compositional levels
```

or:

```text
depth of reusable abstraction hierarchy
```

---

# 69. Niche Creation

A particularly strong metric.

A new adaptation is more interesting if it creates niches as well as occupying one.

Example:

```text
a new structure
↓
creates a new resource
↓
another lineage starts using the resource
↓
a new interaction emerges
```

One possible measure:

```text
new affordances created
new resource loops
new interaction edges
new persistent dependent lineages
```

---

# 70. Evolvability

Measures:

> The system's capacity to continue producing useful heritable variation.

Possible proxies:

```text
fraction of viable mutations
diversity of descendant phenotypes
rate of persistent innovations
ability to escape local attractors
number of distinct accessible strategies
```

Crucially:

```text
current efficiency
```

must not destroy:

```text
future variability
```

---

# 71. Ontology Growth

The most interesting long-term metric.

Track the emergence of new reusable abstractions:

```text
new primitives
new entity types discovered
new interaction categories
new inheritance channels
new representation systems
new levels of computation
```

Their number alone is not enough.

A new abstraction should be:

```text
used
combined
used to create new derived structures
```

---

# 72. The "Generativity" metric

A separate concept can be introduced:

```text
Generativity
```

The question:

> How much did an innovation expand the space of subsequent possible innovations?

A conceptual criterion:

```text
Generativity(x)
=
number of new persistent innovations
that become reachable after x
```

This is difficult to measure directly.

A practical proxy:

```text
branch world with innovation X
vs
counterfactual branch without X
```

Compare subsequent novelty over a long horizon.

If, after X:

```text
novelty rate ↑
niche count ↑
new primitives ↑
hierarchy depth ↑
```

then X was generative.

---

# 73. Counterfactual branches

To evaluate the actual value of a new mechanism, it is useful to create:

```text
World A = before the change
World B = with the change
```

Both start from:

```text
the same snapshot
the same initial seeds
```

and then diverge.

One can estimate:

```text
Δnovelty
Δdiversity
Δhierarchy
Δniche creation
Δevolvability
```

This is much better than simply looking at absolute values.

---

# 74. Novelty without degradation

It is useful to distinguish:

```text
novelty
```

from:

```text
productive novelty
```

Example:

```text
random mutation explosion
→ high raw novelty
→ low persistence
→ low causal structure
→ low productive novelty
```

Another example:

```text
new signaling protocol
→ moderate raw novelty
→ high persistence
→ new cooperation
→ new niches
→ high productive novelty
```

---

# 75. Minimum branch filters

Before entering the long-term archive, a branch must satisfy minimum constraints:

```text
viability > Vmin
persistence > Pmin
causal_structure > Cmin
```

This is not the "goal of evolution."

It is a research-experiment filter.

---

# 76. A Pareto Frontier instead of a single score

Do not use:

```text
Score =
0.4 novelty
+ 0.3 complexity
+ 0.3 diversity
```

Otherwise, the system will quickly learn to exploit the particular weights.

Use Pareto selection instead.

For example:

World A:

```text
very high diversity
medium complexity
```

World B:

```text
medium diversity
high hierarchy
```

World C:

```text
low diversity
very high ontology growth
```

All three can be retained.

---

# 77. Quality-Diversity Archive

The approach resembles quality-diversity / MAP-Elites.

Retain the best worlds in different regions of feature space.

For example, the axes might be:

```text
diversity
hierarchy depth
resource-cycle complexity
communication complexity
ontology depth
```

This preserves unusual but potentially promising lineages.

---

# 78. Protection against metric gaming

If AI knows the exact function:

```text
score(world)
```

it may start optimizing it formally.

For example:

```text
diversity is required
↓
AI creates a million meaningless random types
```

Therefore:

1. use multiple metrics;
2. hide some metrics from the patch generator;
3. change the evaluator regularly;
4. use human/AI qualitative review only as an additional layer;
5. use counterfactual tests;
6. test persistence;
7. test causal impact;
8. test generativity.

---

# 79. Surprise vs Novelty vs Progress

Distinguish three concepts.

## Surprise

```text
unexpected relative to a model
```

## Novelty

```text
something that did not exist before
```

## Progress

```text
an innovation increases the system's future possibilities
```

OEE is primarily concerned with:

```text
persistent generative novelty
```

rather than surprise alone.

---

# 80. Proposed meta-evaluation

A pipeline, rather than a single formula:

```text
candidate branch
↓
VIABILITY FILTER
↓
PERSISTENCE FILTER
↓
NOVELTY CHECK
↓
CAUSAL / FUNCTIONAL CHECK
↓
GENERATIVITY ESTIMATE
↓
PARETO ARCHIVE
```

This is much more robust than:

```text
complexity_score > threshold
```

---

# 81. How AI should select branches

AI can help analyze:

```text
"What is new here?"
"What is functional here?"
"What new affordances appeared?"
"Has a new level of organization emerged?"
```

However, the final decision should preferably be based on:

```text
measured metrics
+
counterfactual branches
+
archive comparison
```

rather than solely on the LLM's written opinion.

---

# 82. Reference design for the external selector

```text
                      ┌──────────────┐
                      │ Candidate W  │
                      └──────┬───────┘
                             ▼
                     Viability Filter
                             │
                 fail ───────┴────── pass
                                   ▼
                             Persistence
                                   │
                                   ▼
                               Novelty
                                   │
                                   ▼
                           Causal Structure
                                   │
                                   ▼
                           Generativity Test
                                   │
                                   ▼
                           Pareto / QD Archive
```

---

# 83. An important philosophical distinction

Keep two distinct levels.

## Inside the world

There is no:

```text
good
bad
problem
progress
goal
```

There is only:

```text
physics
interactions
persistence
inheritance
consequences
```

## At the researcher level

There is an experimental objective:

```text
find systems
that continue producing
persistent adaptive novelty
```

The evaluation function therefore exists as the following, rather than as a world law:

```text
experimental selection criterion
```

---

# 84. The strongest version of the external criterion

Perhaps the best question is not:

```text
"Did the system become more complex?"
```

but:

> "Did this change create more ways for the system to become more complex in the future?"

Evaluate not only the state:

```text
Complexity(t)
```

but the rate of change in possibilities:

```text
FuturePossibilityGrowth
```

Conceptually:

```text
OEE quality
≈
persistent increase in accessible adaptive possibilities
```

This comes closest to the strong concept of open-ended evolution.

---

# 85. Updated cycle architecture

```text
MICRO EVOLUTION
        ↓
WORLD DYNAMICS
        ↓
new affordances emerge
        ↓
persistent structures appear
        ↓
EMERGENT PATTERN DETECTOR
        ↓
AI POSSIBILITY EXPANDER
        ↓
optional ontology compilation
        ↓
candidate world branches
        ↓
VIABILITY + NOVELTY + CAUSALITY + GENERATIVITY
        ↓
PARETO / QUALITY-DIVERSITY ARCHIVE
        ↓
promising worlds continue
        └────────────────────────↺
```

The central idea:

> Inside the world there are no problems. Outside it, the research criterion is to retain branches exhibiting persistent, causally significant, generative novelty.



---

# 86. The next level: evaluate future possibilities, not just the current state

The main weakness of ordinary complexity metrics:

```text
complexity(world_t)
```

measures only the current state.

For OEE, a different question matters more:

> Does the current state create new possible trajectories for subsequent evolution?

It is therefore useful to think in terms of:

```text
Future Possibility Space
```

Conceptually:

```text
FPS(world_t)
=
the set of qualitatively distinct persistent states
reachable from the current world within horizon T
```

The quantity of interest is then not merely:

```text
FPS size
```

but:

```text
growth(FPS)
```

In other words:

```text
whether new classes of reachable future states appear
```

---

# 87. Generativity as the central metric

One can formally introduce:

```text
Generativity(X)
```

for an innovation `X`.

The intuition:

> How much did the emergence of X increase the number of subsequent independent innovations?

Example:

```text
X = a new type of intercellular bond
```

It makes the following possible:

```text
cooperation
specialization
resource exchange
colonies
collective memory
parasitism of the collective
collective defense
```

X therefore has high generativity.

---

# 88. Practical evaluation of Generativity

The exact space of future possibilities is unknown.

Use a sample of counterfactual branches instead.

For innovation X:

```text
Snapshot S
```

create:

```text
Branch A: X enabled
Branch B: X disabled
```

For each branch:

```text
K random seeds
×
T future ticks
```

Then compare:

```text
new niches
new persistent structures
new interaction categories
new hierarchy levels
new inheritance mechanisms
new abstractions
```

Conceptually:

```text
G(X) =
ExpectedNoveltyFuture(X)
-
ExpectedNoveltyFuture(no X)
```

---

# 89. Evaluate the independence of innovations as well as their number

The problem:

One new capability may create a thousand nearly identical variants.

For example:

```text
1000 shades of the same signal
```

This is not equivalent to:

```text
signaling
+
collective memory
+
a new inheritance channel
```

The novelty archive should therefore cluster innovations by:

```text
mechanism
function
causal role
interaction topology
representation
```

And count:

```text
independent innovation classes
```

rather than the raw number of variants.

---

# 90. An Open-Endedness Score should not be a single number

Even when a dashboard is needed, display several time series:

```text
NoveltyRate(t)
GenerativityRate(t)
OntologyDepth(t)
HierarchyDepth(t)
Evolvability(t)
Diversity(t)
Persistence(t)
```

And then, separately:

```text
SaturationIndicators(t)
```

For example:

```text
novelty ↓
ontology growth = 0
niche creation ↓
same interaction patterns repeat
```

---

# 91. Saturation detector

Distinguish:

```text
a temporary plateau
```

from:

```text
structural saturation of the world
```

Signs of structural saturation:

```text
1. new genomes keep appearing;
2. behavior changes slightly;
3. but there are no new functions;
4. no new niches;
5. no new types of interaction;
6. hierarchy depth does not increase;
7. the ontology archive gains no new classes.
```

Such a world is technically evolving, but is not open-ended.

---

# 92. Measuring semantic novelty

A particularly difficult task:

> Determine whether a new object performs a new role.

Analyze its causal use rather than its form.

For example:

```text
Structure A
```

may be geometrically novel but perform the same function.

Whereas:

```text
Structure B
```

may look almost identical to an old structure but start being used as:

```text
memory
signal relay
energy store
construction template
```

The descriptor should therefore include:

```text
inputs
outputs
causal dependencies
interaction partners
environmental effects
descendant effects
```

---

# 93. Functional Signature

For each persistent structure, one can build:

```text
FunctionalSignature
```

Example:

```text
inputs:
    photons
    molecule_A

outputs:
    stored_energy
    molecule_B

effects:
    increases local survival
    changes resource field

interactions:
    used by lineage_42
    consumed by lineage_71
```

Novelty is then measured not only by morphology, but also by:

```text
distance(FunctionalSignature)
```

---

# 94. Affordance Graph

Instead of a list of entities, build a graph of possibilities:

```text
Entity / Structure
      ↓
can interact with
      ↓
Resource / Signal / Artifact / Other entity
      ↓
possible transformation
```

Example:

```text
A --consume--> X
A --signal--> B
B --transform--> X
C --attach--> A
```

When a new edge type appears:

```text
store
teach
imitate
delegate
encode
```

this is stronger evidence than a new object shape alone.

---

# 95. Affordance Expansion Score

One can estimate:

```text
AES(t)
=
number of new persistent affordance classes
introduced over window Δt
```

But count only affordances that:

```text
are actually used
persist
affect subsequent dynamics
```

---

# 96. Evolution of relations matters more than evolution of objects

A possible central principle:

> OEE may grow primarily through new types of relations, rather than new types of objects.

For example:

```text
there are two organisms
```

but relations emerge:

```text
predation
symbiosis
parasitism
teaching
trade
trust
delegation
signaling
inheritance
```

Each new relation type sharply expands the combinatorial space of subsequent development.

---

# 97. Relation-First Ontology

A new ontology can therefore be built as follows, instead of:

```text
new EntityType
```

use:

```text
new RelationType
```

Example:

```text
transfer_energy
```

later becomes:

```text
share_energy
```

then:

```text
conditional_share
```

then:

```text
reciprocal_exchange
```

then:

```text
credit-like relation
```

Social and economic structures could therefore emerge from the evolution of relations.

---

# 98. Meta-Affordances

Affordances that create other affordances are particularly interesting.

For example:

```text
language
```

does more than help transmit a message.

It enables:

```text
instruction
promise
contract
teaching
coordination
planning
```

In other words:

```text
affordance → new affordance generator
```

Such mechanisms should have particularly high generativity scores.

---

# 99. Abstraction as Compression

A new primitive can be considered useful if it describes many recurring processes more compactly.

For example:

before `Membrane`:

```text
thousands of separate local bond rules
```

after:

```text
Membrane(...)
```

If the new abstraction:

```text
substantially reduces description length
+
preserves predictive power
```

it is a good candidate for ontology compilation.

---

# 100. Minimum Description Length for new primitives

The MDL idea can be used:

```text
DescriptionLength(before)
vs
DescriptionLength(after abstraction)
```

If:

```text
DL(after) << DL(before)
```

and predictive ability does not deteriorate,

the new abstraction may correspond to a real structure rather than an AI invention.

---

# 101. Three conditions for compiling a new ontology

A new primitive may be added only if:

```text
1. Compression
2. Persistence
3. Causal usefulness
```

In other words, it:

- compresses the description of the world;
- occurs persistently;
- helps predict consequences.

This reduces the risk of AI creating artificial categories merely to increase a novelty score.

---

# 102. Causal Emergence + Ontology Compilation

The process:

```text
microstates
↓
candidate coarse-graining
↓
measure predictive / effective information
↓
macro variable explains dynamics better
↓
compile macro variable as new primitive
```

This produces:

```text
causal emergence
→
new ontology
→
new evolutionary building block
```

---

# 103. A new AI role: scientist-compiler

The external agent should be separated into several roles.

```text
Observer
Hypothesis Generator
Ontology Miner
Patch Generator
Critic
Experiment Designer
```

Particularly interesting:

```text
Ontology Miner
```

It does not invent entities from scratch.

It asks:

> Which recurring macropatterns already exist but are still represented only as thousands of microinteractions?

---

# 104. Ensemble Evaluator

Do not give one AI all of the following roles simultaneously:

```text
generate patch
+
evaluate patch
```

Otherwise, it may evaluate its own decisions favorably.

Prefer:

```text
Agent A → proposes
Agent B → critiques
Metrics → measure
Agent C → interprets
Archive → decides retention
```

The final selector relies primarily on measurable results.

---

# 105. Blind Evaluation

For some experiments, the evaluator should not know:

```text
which patch was applied
who created it
what the hypothesis was
```

It receives only:

```text
before / after telemetry
```

This reduces confirmation bias.

---

# 106. Hidden Metrics

Not all metrics should be shown to the Patch Generator.

For example, the generator sees:

```text
world description
constraints
```

but does not know the exact formula for:

```text
GenerativityEvaluator
```

This reduces the Goodhart effect.

---

# 107. Goodhart Resistance

The main rule:

> When a measure becomes a target, it ceases to be a good measure.

The OEE evaluator should therefore regularly check:

```text
metric gaming
```

Examples:

```text
diversity score ↑
through meaningless noise

complexity score ↑
through huge, useless structures

novelty score ↑
through constant random changes of state
```

Countermeasures:

```text
persistence
causal usefulness
counterfactual impact
compression
generativity
multi-objective archive
```

---

# 108. Red-Team Evaluator

A separate agent can try to explain:

> Why an allegedly new phenomenon is not actually progress toward OEE.

Example:

```text
"The new behavior is merely a parameter variation of an existing mechanism."
```

Or:

```text
"The increase in diversity is caused by randomness and is not inherited."
```

Only candidates that withstand this criticism enter the archive.

---

# 109. Novelty Lineage

Each innovation should have its own history:

```text
InnovationID
parent innovations
first appearance
lineages using it
dependent innovations
descendant affordances
```

Example:

```text
Signal
↓
DirectedSignal
↓
GroupCoordination
↓
Specialization
↓
CollectiveReproduction
```

This makes it possible to measure both the number of innovations and the depth of their causal genealogy.

---

# 110. Innovation DAG

A DAG is preferable to a tree:

```text
Innovation A ─┐
              ├→ Innovation C
Innovation B ─┘
```

Because many discoveries combine previous ones.

Strong OEE should exhibit:

```text
increasing DAG depth
+
increasing recombination
```

---

# 111. Reuse as evidence of meaningful complexity

A mechanism that appears once and disappears is a weak result.

If a mechanism becomes a reusable building block:

```text
X
↓
it is used by 20 independent lineages
↓
combined with Y and Z
↓
and becomes the foundation of new systems
```

the result is much stronger.

One can measure:

```text
ReuseScore(X)
```

---

# 112. Compositionality Score

An important sign of open-endedness:

> New elements must be composable.

For example:

```text
A
B
C
```

produce:

```text
AB
AC
BC
ABC
```

If every innovation is independent and noncomposable, the possibility space grows slowly.

If innovations are compositional:

```text
possibility space
```

can grow combinatorially.

---

# 113. Evolution of modularity

It is useful to observe whether the following emerges:

```text
module
```

a part of the system that:

- has a local function;
- is reused;
- is relatively independent;
- can be combined with other parts.

Modularity may prove to be one of the most important mechanisms of OEE.

---

# 114. Complexity Budget

To prevent AI from creating arbitrarily expensive rules:

each new primitive receives a cost:

```text
compute cost
memory cost
energy cost
description cost
```

A new capability must compete for a limited budget.

Otherwise:

```text
AI simply adds everything
```

and the world's possibility space grows artificially, without selection.

---

# 115. Conservation of computational resources

A highly desirable property:

```text
new capability ≠ free capability
```

If the following is added:

```text
long-range signaling
```

it must have a cost:

```text
energy
latency
noise
memory
bandwidth
```

This creates trade-offs.

Trade-offs are one of the drivers of diversity.

---

# 116. Trade-off Generator

A new primitive should, where possible, create:

```text
advantage A
↔
cost B
```

Examples:

```text
rapid reproduction ↔ low fidelity

long-range signaling ↔ high energy cost

thick membrane ↔ slow exchange

large memory ↔ computational cost
```

Without trade-offs, evolution can easily collapse into one universally best variant.

---

# 117. No universally best organism

The world's architecture should aim for:

```text
context-dependent fitness
```

rather than:

```text
global optimum
```

Ideally:

```text
Strategy A beats B
B beats C
C beats A
```

or efficiency depends on:

```text
environment
population composition
history
local resource structure
```

This supports prolonged coevolution.

---

# 118. Red Queen Dynamics

A desirable regime:

```text
species A adapts to B
↓
B adapts to A
↓
A changes again
↓
...
```

However:

Ordinary Red Queen dynamics can cycle indefinitely within a single strategy space.

OEE requires occasional:

```text
Red Queen cycle
↓
new interaction mechanism
↓
new strategy space
```

---

# 119. Major Transition Detector

Automatically look for signs of a new level of individuality.

A candidate group should have:

```text
persistent boundary
internal resource sharing
internal signaling
division of labor
common reproduction
reduced internal conflict
shared fate
```

If these properties grow together:

```text
candidate major transition
```

---

# 120. Conflict Suppression as evidence of a major transition

In evolutionary history, new levels of organization often require suppression of internal conflict.

Example:

```text
the cells of an organism
```

should not compete with one another indefinitely.

A useful metric is therefore:

```text
internal competition
vs
collective fitness coupling
```

As a group becomes more integrated, internal conflict should decrease or become regulated.

---

# 121. New inheritance channels

The system should track the emergence of:

```text
genetic inheritance
epigenetic-like inheritance
horizontal transfer
behavioral imitation
cultural transmission
artifact inheritance
environmental inheritance
```

Each new inheritance channel can potentially increase evolvability substantially.

---

# 122. Environmental Inheritance

A particularly interesting possibility:

An organism can pass an altered environment to its descendants, rather than information stored inside itself.

For example:

```text
it builds a structure
↓
dies
↓
its descendants use the structure
```

This is already inheritance through:

```text
niche construction
```

---

# 123. Ecological Memory

The environment itself can store history.

For example:

```text
chemical traces
persistent structures
resource depletion
constructed channels
symbol markers
```

The world thus becomes evolution's external memory.

---

# 124. Evolutionary Memory Stack

Several memory layers can be envisioned:

```text
genome
↓
cell state
↓
organism memory
↓
social memory
↓
environmental memory
↓
artifacts
↓
symbolic records
```

The emergence of a new memory layer is a potential major transition.

---

# 125. Time-Scale Separation

It is important to model different time scales.

For example:

```text
physics: every tick
behavior: tens of ticks
lifetime: thousands of ticks
ecology: millions of ticks
ontology change: tens of millions of ticks
```

If all levels change at the same speed, persistent structures may not have time to emerge.

---

# 126. Slow Macro-Evolution

AI macropatches should occur much less frequently than microevolution.

For example:

```text
micro mutation: continuously
macro rule proposal: only after a long observation window
ontology compilation: even less frequently
```

Otherwise, the external AI becomes the world's main author.

---

# 127. External intervention budget

Introduce:

```text
InterventionBudget
```

For example, per million ticks AI may:

```text
add no more than 1 primitive
change no more than 0.1% of the rule space
```

This makes it possible to test:

> Can a small number of semantic extensions support a vast amount of internal evolution?

---

# 128. Minimal intervention as a research principle

The best AI patch:

> The smallest change that expands future possibilities the most.

This can be optimized as:

```text
Generativity
----------------
PatchComplexity
```

In other words, a high:

```text
Generativity per added rule
```

---

# 129. An Occam principle for OEE

If two patches produce the same generativity:

```text
Patch A = 2 new primitives
Patch B = 50 new primitives
```

A is preferable.

This keeps AI from becoming an endless content generator.

---

# 130. An open world versus an extensible world

Distinguish:

## Open state space

```text
a vast number of states
```

from:

## Expanding state space

```text
new types of states emerge
```

The latter is more interesting for strong OEE.

---

# 131. Ontological event log

Record events separately:

```text
first replicator
first stable parasite
first cooperative exchange
first persistent group
first new inheritance channel
first symbol-like representation
first external memory
first self-built interpreter
```

This becomes the "history of the universe."

---

# 132. Automatic Scientific Narration

The AI Observer can build a scientific journal:

```text
Generation 2.4M:
Persistent energy-sharing clusters emerged.

Generation 2.7M:
A lineage began exploiting clusters parasitically.

Generation 3.1M:
Clusters evolved selective binding.

Generation 3.8M:
Evidence suggests collective reproduction.
```

This is especially useful when the world runs for weeks or months.

---

# 133. Branch Archaeology

An interesting capability:

> Take a present-day complex phenomenon and reconstruct its chain of origins.

For example:

```text
a present-day symbolic system
↓
which innovation nodes were necessary?
↓
which could have been absent?
↓
which transition was critical?
```

Counterfactual branches can be launched automatically from past states.

---

# 134. Causal Importance of Innovation

For each historical event X:

```text
remove X from a past snapshot
↓
repeat many runs
```

If the later class of structures almost never appears without X:

```text
X = evolutionary bottleneck / key innovation
```

---

# 135. Convergent Evolution Test

A particularly interesting experiment:

```text
the same physics
+
different seeds
```

Do the following emerge independently:

```text
membranes?
parasites?
communication?
multicellularity?
symbol systems?
```

If so, this suggests deep attractors in the possibility space.

---

# 136. Contingency vs Necessity

Run thousands of histories.

For each major transition, estimate:

```text
P(transition | physics)
```

If the event appears almost always:

```text
likely structural necessity
```

If it is extremely rare:

```text
historical contingency
```

This already makes the system interesting as a model for fundamental questions about evolution.

---

# 137. Search for Universal Evolutionary Patterns

One can ask:

```text
does parasitism recur?
does cooperation emerge?
is modularity necessary?
does hierarchy emerge?
does division of labor appear?
```

If independent digital worlds regularly arrive at the same abstract solutions, this may be more interesting than a particular attractive organism.

---

# 138. World complexity as a causal network

Instead of entity counts:

```text
Complexity ≈ structure of causal dependency graph
```

Quantities of interest:

```text
depth
modularity
feedback loops
cross-scale dependencies
reusable motifs
```

---

# 139. Multi-Scale Causal Graph

Store:

```text
micro causal graph
meso causal graph
macro causal graph
```

And examine:

```text
the scales at which persistent causal laws emerge
```

This potentially connects the system to causal emergence.

---

# 140. When an abstraction becomes "real"

A working definition:

> A macro-object counts as a real simulation level if using it improves prediction and control compared with a purely microscopic description.

In other words:

```text
predictive gain
+
compression gain
+
causal usefulness
```

---

# 141. Final form of the external selector

```text
WORLD BRANCH
    ↓
Minimum Viability
    ↓
Persistence
    ↓
Raw Novelty
    ↓
Functional / Causal Novelty
    ↓
Affordance Expansion
    ↓
Generativity
    ↓
Ontology Growth
    ↓
Quality-Diversity / Pareto Archive
```

No single layer defines "progress."

---

# 142. A possible main OEE criterion

A working formulation:

> A system exhibits strong open-ended evolution if, over long time scales, it continues producing new persistent causal-functional structures that expand the set of subsequent available adaptations and create new levels of organization.

In brief:

```text
OEE
=
persistent
+
causal
+
generative
+
compositional
+
non-saturating novelty
```

---

# 143. The most important practical test

The first truly convincing experiment:

1. Run an ordinary world without external AI.
2. Measure the time to structural saturation.
3. Run the same world with an AI Possibility Expander.
4. Limit AI to a small intervention budget.
5. Do not give AI a final objective.
6. Compare:
   - novelty;
   - generativity;
   - ontology depth;
   - hierarchy depth;
   - niche creation;
   - time to saturation.

If, with a small number of neutral semantic extensions:

```text
AI-world
```

consistently produces new functional categories for longer than:

```text
control-world
```

that would already be a very interesting result.

---

# 144. A stronger test

After the previous experiment succeeds:

AI may add a primitive only when:

```text
primitive
```

compiles a pattern that has already emerged.

In other words, AI may not invent mechanisms from outside.

Diagram:

```text
world produces latent pattern
↓
detector finds it
↓
AI abstracts it
↓
experiment validates abstraction
↓
kernel exposes primitive
```

If this mode also sustains OEE:

> The direction of expansion truly comes from the world, while AI merely accelerates transitions between levels.

This is much stronger than the original architecture.

---

# 145. Ultimate Experiment

The most ambitious version:

```text
immutable minimal kernel
+
very small initial physics
+
micro evolution
+
automatic macro-pattern discovery
+
ontology compilation
+
external AI abstraction engine
+
quality-diversity world archive
```

The question then becomes:

> How far can the system build its own language for describing the world on top of the initial primitives?

Ideally:

```text
primitive physics
↓
replicators
↓
boundaries
↓
organisms
↓
collectives
↓
communication
↓
symbols
↓
culture
↓
technology
↓
new computation
↓
unknown abstractions
```

The final levels should not be named in advance by the developer.

The strongest sign of genuine open-ended evolution would be something for which we must invent a new concept after the experiment.


---

# 146. What the world's foundation might look like in code

The main principle:

> The base world should not know what an organism, species, predator, food, language, culture, or economy is.

It should know only:

```text
space
matter
energy
state
bonds
memory
executable rules
```

In other words, avoid:

```go
type Organism struct {
    Health       float64
    Age          int
    Species      SpeciesID
    Attack       float64
    Defense      float64
    Intelligence float64
}
```

and build general-purpose primitives instead.

---

# 147. Basic entity

For the first version:

```go
type EntityID uint64
type Tick uint64

type Vec2 struct {
    X, Y int
}

type Entity struct {
    ID EntityID

    Pos Vec2

    Energy float64
    Matter float64

    Memory []float64

    Program Program

    Bonds []Bond

    Properties map[PropertyID]Value
}
```

The key element:

```go
Properties map[PropertyID]Value
```

Do not add predefined fields such as:

```text
Health
Sex
Vision
Species
Intelligence
```

Later, new mechanisms may add:

```text
electrical_charge
membrane_permeability
signal_frequency
chemical_affinity
symbol_memory
```

without changing the base type.

---

# 148. Dynamic properties

```go
type PropertyID uint32

type Value struct {
    Number float64
    Bytes  []byte
}
```

A more advanced version can use a tagged union:

```go
type ValueType uint8

const (
    ValueNumber ValueType = iota
    ValueBytes
    ValueVector
    ValueEntityRef
)

type Value struct {
    Type   ValueType
    Number float64
    Bytes  []byte
    Vec    []float64
    Ref    EntityID
}
```

---

# 149. World

```go
type World struct {
    Tick Tick

    Width  int
    Height int

    Entities map[EntityID]*Entity

    Fields Fields

    Rules RuleSet

    Events []Event

    RNG *rand.Rand
}
```

`World` is only the current state.

It contains no semantics such as:

```text
population
species
ecosystem
society
```

These concepts should emerge in the external Observer.

---

# 150. Spatial fields

```go
type Fields struct {
    Energy Grid[float64]
    Matter Grid[float64]

    Dynamic map[FieldID]*ScalarField
}
```

Initially:

```text
Energy
Matter
```

Later, one can dynamically add:

```text
Light
Temperature
Chemical_A
Chemical_B
SignalField
Charge
```

---

# 151. Basic Rule Interface

The kernel should not know the meaning of rules.

```go
type Rule interface {
    ID() RuleID

    Apply(
        ctx *RuleContext,
        out *EventBuffer,
    )
}
```

Examples of built-in rules:

```go
type MoveRule struct{}
type EnergyTransferRule struct{}
type BondRule struct{}
type DecayRule struct{}
type ProgramExecutionRule struct{}
```

---

# 152. RuleSet

```go
type RuleSet struct {
    Rules []Rule
}
```

The main principle:

> A Rule must not modify the World directly.

It generates events.

---

# 153. Event-driven world changes

```go
type Event interface {
    Apply(*World)
}
```

Examples:

```go
type MoveEvent struct {
    Entity EntityID
    To     Vec2
}

type TransferEnergyEvent struct {
    From   EntityID
    To     EntityID
    Amount float64
}

type SpawnEvent struct {
    Parent EntityID
    Child  Entity
}

type DestroyEvent struct {
    Entity EntityID
}

type SetPropertyEvent struct {
    Entity   EntityID
    Property PropertyID
    Value    Value
}
```

This produces:

```text
World(t)
↓
Rules / Programs
↓
Events
↓
Conflict Resolution
↓
World(t+1)
```

This provides:

```text
replay
debugging
causal history
branching
counterfactual experiments
```

---

# 154. Executable matter

An entity stores a program.

```go
type Opcode byte

const (
    OpSense Opcode = iota
    OpMove
    OpAbsorb
    OpEmit
    OpBind
    OpUnbind
    OpCopy
    OpJump
    OpCompare
)

type Instruction struct {
    Op Opcode
    A  int32
    B  int32
}

type Program struct {
    Code []Instruction
    IP   int
}
```

A minimal program might look like:

```text
SENSE ENERGY
COMPARE > 0.3
JUMP 5
MOVE RANDOM
JUMP 0

ABSORB ENERGY
COPY
JUMP 0
```

However, `COPY` should not mean:

```text
reproduce organism
```

It should mean only:

> Copy a structure or program when resources are available.

---

# 155. Do not hardcode replication

Avoid:

```go
func (e *Entity) Reproduce() *Entity
```

Prefer a set of low-level operations:

```text
allocate matter
copy memory
copy instructions
transfer energy
create bond
detach
```

Evolution can then develop a replication algorithm itself:

```text
create a new structure
↓
copy code
↓
transfer energy
↓
detach
```

The reproduction mechanism itself can then evolve, as well as the genome.

---

# 156. An Entity does not have to be an organism

Better still, treat `Entity` as an atomic node:

```go
type Node struct {
    ID EntityID

    Pos Vec2

    Energy float64
    Matter MatterType

    Memory []byte

    Program []Instruction

    Bonds []EntityID
}
```

Then:

```text
1 Node
```

may be meaningless,

while:

```text
100 connected Nodes
```

may form:

```text
a replicator
a membrane
a machine
a colony
```

The engine itself does not know this.

---

# 157. Relations as first-class citizens

Bonds should be a full-fledged part of the world.

```go
type Relation struct {
    ID RelationID

    From EntityID
    To   EntityID

    Kind PrimitiveID

    State []float64
}
```

Because OEE may develop through new relations:

```text
mechanical bond
energy transfer
signal
conditional exchange
memory reference
ownership-like relation
teaching-like relation
```

---

# 158. Primitive Registry

A central layer of an extensible world:

```go
type Registry struct {
    Properties map[PropertyID]PropertyDefinition
    Fields     map[FieldID]FieldDefinition
    Primitives map[PrimitiveID]PrimitiveDefinition
    Opcodes    map[OpcodeID]OpcodeDefinition
}
```

Initial Registry:

```text
energy
matter
position

move
bind
transfer
copy
```

A possible later Registry:

```text
energy
matter
position
charge
light
chemical_A
chemical_B

move
bind
transfer
copy
signal
catalyze
store
encode
interpret
...
```

This is the practical implementation of:

```text
S0 → S1 → S2 → ...
```

---

# 159. Immutable Kernel

```go
type Kernel struct {
    VM       VM
    Registry Registry

    Limits Limits

    Snapshotter Snapshotter

    Validator PatchValidator
}
```

The kernel knows only:

```text
how to execute instructions
how to apply events
how to allocate resources
how to create a snapshot
how to validate a patch
how to create a new branch
```

It does not know:

```text
organism
species
culture
language
technology
```

---

# 160. AI patches must be declarative

Do not allow AI to write arbitrary native code in the kernel.

Prefer:

```yaml
version: 1

add_property:
  name: polarity
  type: float
  range: [-1, 1]

add_rule:
  name: polarity_binding

  when:
    distance: "< 1"
    condition:
      - "a.polarity * b.polarity < 0"

  effect:
    create_relation:
      type: bond

  cost:
    energy: 0.02
```

Lifecycle:

```text
parse
↓
validate
↓
compile to VM/bytecode
↓
sandbox
↓
branch world
```

---

# 161. AI patches are not applied directly to the main branch

```go
func TestPatch(
    snapshot Snapshot,
    patch Patch,
    seeds []int64,
) []ExperimentResult
```

Diagram:

```text
             Snapshot W
             /    |    \
            /     |     \
       original   P1     P2
                           \
                            P3
```

AI creates new branches.

The world is not irreversibly rewritten.

---

# 162. Minimal tick loop

```go
func (w *World) Step(k *Kernel) {
    ctx := TickContext{
        Tick:     w.Tick,
        World:    w.ReadOnlyView(),
        Registry: &k.Registry,
    }

    events := NewEventBuffer()

    // 1. Executable entity code.
    k.VM.ExecuteAll(ctx, events)

    // 2. General world rules.
    for _, rule := range k.Registry.ActiveRules() {
        rule.Evaluate(ctx, events)
    }

    // 3. Conflict resolution.
    resolved := ResolveEvents(events)

    // 4. World state changes.
    ApplyEvents(w, resolved)

    // 5. Losses / dissipation.
    ApplyDissipation(w)

    w.Tick++
}
```

Goal:

> All complexity should grow on top of the smallest possible loop.

---

# 163. Keep the Observer outside the World

```go
type Observer struct {
    Metrics     MetricsEngine
    Novelty     NoveltyDetector
    Patterns    PatternDetector
    CausalModel CausalAnalyzer
}
```

The Observer may report:

```text
"a population appears to have emerged"
"a membrane appears to have emerged"
"a new niche appears to have emerged"
```

But these concepts do not exist inside the world's physics.

---

# 164. World Summary for the external AI

```go
type WorldSummary struct {
    Tick uint64

    EnergyFlow float64

    Diversity DiversityMetrics
    Novelty   NoveltyMetrics

    PersistentPatterns []Pattern
    NewAffordances     []Affordance

    CandidateMacroEntities []MacroEntity

    Saturation SaturationMetrics
}
```

Even `CandidateMacroEntities` is only an Observer hypothesis.

---

# 165. A clear separation of the system

```text
┌─────────────────────────────┐
│          KERNEL             │
│ computation                 │
│ resources                   │
│ events                      │
│ sandbox                     │
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│           WORLD             │
│ matter                      │
│ energy                      │
│ programs                    │
│ relations                   │
│ fields                      │
└─────────────┬───────────────┘
              │ observe
              ▼
┌─────────────────────────────┐
│         OBSERVER            │
│ organisms?                  │
│ niches?                     │
│ hierarchy?                  │
│ novelty?                    │
│ affordances?                │
└─────────────┬───────────────┘
              ▼
┌─────────────────────────────┐
│      AI META-EVOLUTION      │
│ hypothesis                  │
│ ontology compilation        │
│ possibility expansion       │
└─────────────────────────────┘
```

---

# 166. The very first code version

The MVP can be simplified further:

```go
type Particle struct {
    ID EntityID

    X int
    Y int

    Energy uint16

    Code []byte

    Memory [8]int16
}
```

With just 8 instructions:

```text
NOP
MOVE
SENSE
ABSORB
TRANSFER
WRITE
COPY
JUMP
```

World:

```text
256 × 256
```

Each instruction costs energy.

Energy enters from outside as a physical gradient.

If:

```text
Energy == 0
```

the structure stops executing or disintegrates.

---

# 167. What NOT to write

In particular, avoid:

```go
type Agent interface {
    Think()
    Act()
    Reproduce()
}
```

Because this introduces the following in advance:

```text
agency
thinking
action
reproduction
```

Instead:

```text
matter executes local transformations
```

Agency should be an interpretation of a persistent pattern.

---

# 168. Minimal Go project layout

```text
/internal/kernel
    tick.go
    events.go
    limits.go
    snapshot.go

/internal/world
    world.go
    particle.go
    field.go
    relation.go

/internal/vm
    opcode.go
    program.go
    executor.go

/internal/rules
    registry.go
    builtin.go

/internal/evolution
    mutation.go
    lineage.go

/internal/observer
    metrics.go
    patterns.go
    novelty.go
    affordance.go

/internal/branch
    manager.go
    experiment.go

/internal/agent
    protocol.go
    patch.go
    validator.go

/cmd/sim
    main.go
```

---

# 169. The next stage after the basic kernel

Once:

```text
World
+
Tick Loop
+
VM
+
Energy
+
Snapshot/Replay
```

are working, the next step is NOT to connect AI.

The next step is:

# Establish autonomous internal evolution without external help.

The main question:

> Can minimal digital physics support replication, mutation, competition, and persistent lineages without an Organism concept?

---

# 170. Stage A — Baseline Digital Evolution

## Goal

Build a complete experimental loop:

```text
energy gradient
↓
executable matter
↓
copy errors
↓
inheritance
↓
differential persistence
↓
evolution
```

---

# 171. What to implement in Stage A

## World

```text
256 × 256 grid
```

## Resources

At minimum:

```text
EnergyField
Matter
```

## VM

Instructions:

```text
NOP
SENSE
MOVE
ABSORB
TRANSFER
WRITE
READ
COPY
JUMP
COMPARE
BIND
UNBIND
```

## Constraints

Each operation has:

```text
energy cost
instruction cost
memory cost
```

---

# 172. Replication in Stage A

Do not use:

```text
Reproduce()
```

The program itself must execute:

```text
allocate / spawn
↓
copy code
↓
copy state
↓
transfer initial energy
↓
detach
```

For the first version, a low-level operation is acceptable:

```text
ALLOCATE
```

but not a high-level one:

```text
REPRODUCE
```

---

# 173. Mutations

The minimal set:

```text
instruction replacement
instruction insertion
instruction deletion
block duplication
block deletion
memory initialization mutation
```

Later:

```text
recombination
horizontal transfer
```

---

# 174. Initial experiment

There are two options.

## Option 1 — Seed Replicator

Place one very simple working replicator in the world.

Goal:

```text
test evolutionary dynamics,
rather than the origin of life.
```

This is the best MVP.

## Option 2 — Abiogenesis Search

Start with random executable matter and wait for replication to emerge.

This is much harder.

It is not recommended for the first version.

---

# 175. Why start with a Seed Replicator?

If the system does not evolve, one must distinguish:

```text
is the problem the origin of the replicator?
or
is the problem the evolutionary architecture itself?
```

A Seed Replicator separates these two questions.

First test:

```text
replication → mutation → ecology
```

Investigate the origin of replication separately, later.

---

# 176. Stage A success criteria

Obtain:

1. a long-lived population;
2. several lineages;
3. heritable differences;
4. changes in lineage frequencies over time;
5. new persistent programs;
6. no need to assign fitness manually.

---

# 177. The first particularly interesting result

Test:

> Does parasitism emerge?

For example, a mutant:

```text
does not copy the full replication machinery
```

but uses:

```text
its neighbors' resources or copying machinery.
```

If this strategy emerges independently:

```text
replicator
↓
parasite
↓
host defense
↓
parasite adaptation
```

that is already a strong sign that basic ecology is working.

---

# 178. A second interesting result

Look for the emergence of:

```text
cooperation
```

For example, one lineage:

```text
collects energy
```

while another:

```text
replicates efficiently
```

and persistent exchange emerges between them.

Important:

```text
do not write a CooperationRule
```

It is enough to enable:

```text
transfer_energy
```

---

# 179. What to measure in Stage A

Minimal telemetry:

```text
entity count
energy throughput
lineage count
genome diversity
behavior diversity
average lifetime
replication rate
extinction rate
resource distribution
interaction graph
```

---

# 180. Behavior Hash

A genome hash alone is insufficient.

Two different genomes can do the same thing.

For each lineage, estimate a:

```text
BehaviorSignature
```

For example:

```text
energy absorbed
distance moved
energy transferred
copies created
relations created
signals emitted
```

This provides an initial measure of behavioral novelty.

---

# 181. Stage A Saturation

Run many long experiments and determine:

```text
how long does novelty take to stop growing?
```

This establishes a baseline:

```text
T_saturation_control
```

AI-assisted OEE will later be compared against it.

---

# 182. The next stage after A

Only if baseline evolution actually works:

# Stage B — Emergent Ecology

Add:

```text
multiple resources
spatial heterogeneity
relations
resource transformation
waste products
local fields
```

Still without external AI.

Goal:

> Get organisms to create new affordances for one another.

---

# 183. An Emergent Ecology example

```text
Lineage A:
Resource X → Waste Y

Lineage B:
Waste Y → Energy

↓
A creates a niche for B
```

Next:

```text
B changes the concentration of Y
↓
this affects A
↓
coevolution emerges
```

There is no "task" here.

---

# 184. Stage B completion criterion

At least one persistent case in which:

```text
one lineage changes the environment
↓
this creates a new affordance
↓
another lineage uses it
↓
long-term interdependence emerges
```

---

# 185. Only then — Stage C: Observer

Once internal ecology has emerged, connect an Observer:

```text
pattern detection
novelty
stagnation
affordances
lineages
candidate macro-entities
```

For now, the Observer:

```text
read-only
```

does not change anything.

---

# 186. Stage D — Branching Experiments

After the Observer, add:

```text
snapshot
↓
clone world
↓
change one rule
↓
run many seeds
↓
compare
```

This prepares the system for external AI.

---

# 187. Stage E — AI Possibility Expander

Connect the external AI only at this point.

It may:

```text
propose new primitive
propose new relation
propose new resource transformation
propose new local field
```

But may not:

```text
create species
create predator defense
create intelligence
```

---

# 188. First AI experiment

Control:

```text
World A
```

Patch:

```text
add SEND_SIGNAL primitive
```

Branch:

```text
World B
```

Important:

AI does not say:

```text
"use the signal for cooperation"
```

It simply adds a possibility.

Evolution then determines what to do with it.

---

# 189. A very strong first result

For example, AI adds:

```text
SEND_SIGNAL
```

with the intention of expanding communication.

But evolution uses it as:

```text
a false signal
↓
a lure
↓
a parasitic strategy
```

This result is more interesting than the expected use.

It shows:

> The world is actually exploring the new space on its own.

---

# 190. Recommended order of subsequent stages

```text
0. Kernel + deterministic replay
↓
1. VM + executable matter
↓
2. Seed replicator
↓
3. Mutation + lineage evolution
↓
4. Baseline saturation measurement
↓
5. Emergent ecology
↓
6. Affordance-producing environment
↓
7. Observer
↓
8. Novelty / saturation detector
↓
9. Branching worlds
↓
10. Counterfactual evaluator
↓
11. AI Possibility Expander
↓
12. AI-generated primitives
↓
13. Evolvability evolution
↓
14. Major transition detection
↓
15. Ontology compilation
↓
16. Cultural / symbolic evolution
↓
17. Self-hosted computation
```

---

# 191. The next practical milestone

The immediate goal after designing the code:

> Run a world in which one simple seed replicator reproduces, mutates, and forms several competing lineages on its own.

There is no need yet for:

```text
AI
OEE evaluator
causal emergence
symbols
culture
```

Establish the basic premise:

```text
minimal physics
+
executable matter
+
limited resources
```

already produce a normal Darwinian process.

Only then does it make sense to extend the system toward open-ended evolution.

