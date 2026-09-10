# DSL: rule changes and reproducibility

2026-09-10. Initial Stage 3 implementation: local reactions, bounded bytecode, versions, and rollback without restarting the kernel.

Tested a 32×32 world, seed 1, matter and chemical transport every 4 ticks, and standard mutations. Initial module: `examples/rules/baseline.json`; `examples/rules/direct-x.json` loads at the start of tick 500, and the previous module returns at the start of tick 1000. Total: 2000 ticks.

```powershell
go run ./cmd/sim -width 32 -height 32 -ecology -matter-diffusion 4 -chemical-diffusion 4 -rules examples/rules/baseline.json -rule-change 500=examples/rules/direct-x.json -rollback-at 1000 -ticks 2000 -every 500 -save data/dsl-full.json
go run ./cmd/sim -width 32 -height 32 -ecology -matter-diffusion 4 -chemical-diffusion 4 -rules examples/rules/baseline.json -rule-change 500=examples/rules/direct-x.json -rollback-at 1000 -ticks 750 -save data/dsl-split.json
go run ./cmd/sim -load data/dsl-split.json -ticks 1250 -save data/dsl-resumed.json
```

Continuous and resumed runs produced identical snapshot bytes and SHA-256:

```text
fae07fd94bd6a8b6418637227ada4fff6e16a06daffbbc72dd03bb898f4ac6a2
```

Final state: 564 particles (561 with code), 9 genomes, 3641 copies, and 3082 deaths. The log contains three events; the DSL budget counter records 2,528,119 work units. All resource invariants hold.

Module versions:

| Version | Canonical-source SHA-256 |
| --- | --- |
| baseline-v1 | aa34583ba4e890cfefd45f849c9df47f1ae1adad1d18ed4839da24c9fa05f020 |
| direct-x-v1 | 97836395dd7c42429e279c0ba7241377903a5c3a00f472bf9059c5337db71ead |

An additional CLI test resumes a snapshot after both module source files are deleted. Kernel tests verify baseline DSL physical-trajectory parity with built-in reactions, new ID 2 and its cost, conservation at every tick, atomic installation rejection, and rule restoration without resetting the world. Loading a snapshot with altered bytecode is rejected.

A Stage 2 snapshot at tick 100,000 loaded without DSL and retained its SHA-256: `2d39a7db78c520febcd31858e91a12ee200e1a374e4df4aa2f0a342d799a7cde`.

This validates execution and reproducibility, not long-term adaptation to changed physics. Other operations and world fields are not yet described by the DSL; the CLI applies changes from a schedule frozen before execution.
