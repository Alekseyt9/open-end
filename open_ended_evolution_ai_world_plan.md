# Open-Ended Evolution + External AI Agent
## Концепция, архитектура и поэтапный план реализации

---

## 0. Краткая идея

Создать цифровой мир, в котором эволюция не ограничивается:

- мутацией фиксированного генома;
- заранее заданными видами;
- одной fitness-функцией;
- заранее определёнными сущностями вроде `Agent`, `Species`, `Predator`, `Weapon`, `Economy`;
- фиксированным набором законов мира.

Ключевая идея проекта:

> локальная эволюция происходит внутри симуляции, а внешний AI-агент получает телеметрию мира и может предлагать изменения кода, правил, примитивов, механизмов наследования и даже самой онтологии мира.

Получается коэволюционирующая система:

```text
мир
↓
возникают новые структуры / проблемы / стагнация
↓
внешний AI анализирует состояние
↓
предлагает макромутации правил
↓
создаются параллельные ветви миров
↓
изменения проверяются самой симуляцией
↓
успешные ветви сохраняются
↓
новые механизмы создают новые возможности и новые давления отбора
↓
цикл повторяется
```

Главный исследовательский вопрос:

> Может ли комбинация локальной дарвиновской эволюции, коэволюционирующей среды и внешнего AI-агента, способного менять пространство правил, поддерживать долговременный рост адаптивной новизны?

---

# 1. Что именно считать Open-Ended Evolution

Важно не путать OEE с обычной процедурной генерацией.

Процедурная генерация:

```text
AI придумал новый биом
AI придумал нового монстра
AI придумал новое оружие
```

Open-ended evolution:

```text
появился новый механизм
↓
он стал частью мира
↓
другие сущности начали его использовать
↓
изменились селективные давления
↓
возникли новые типы взаимодействий
↓
появились новые уровни организации
↓
это открыло ещё одно пространство возможностей
```

Сильная OEE — это не только:

```text
x ∈ S
```

где эволюция ищет новые точки внутри фиксированного пространства `S`.

А скорее:

```text
S0 → S1 → S2 → S3 → ...
```

где меняется само пространство возможных состояний.

Пример:

```text
S0 = физика
S1 = физика + репликация
S2 = S1 + организмы
S3 = S2 + коммуникация
S4 = S3 + символы
S5 = S4 + культура
S6 = S5 + технология
S7 = S6 + новые вычислительные миры
...
```

---

# 2. Основная архитектура

Систему лучше разделить на два слоя.

```text
┌────────────────────────────────────────────┐
│              IMMUTABLE KERNEL              │
│                                            │
│ время                                      │
│ память                                     │
│ sandbox                                    │
│ лимиты ресурсов                            │
│ snapshots                                  │
│ rollback                                   │
│ загрузка модулей                           │
│ версии                                     │
│ branch management                          │
│ детерминированный replay                   │
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

Kernel менять нельзя.

Почти всё, что относится к самому миру, потенциально можно менять.

---

# 3. Почему kernel должен быть неизменяемым

Если разрешить внешнему AI менять вообще весь исполняемый процесс:

- становится трудно воспроизводить эксперименты;
- сложно понять причинность;
- невозможно гарантировать rollback;
- появляется риск повредить сам runtime;
- теряется различие между "эволюцией мира" и "переписыванием программы разработчиком".

Kernel должен задавать только:

```text
что считается вычислением
как распределяются ресурсы
как хранится состояние
как запускается патч
как создаётся новая ветвь
как ограничивается выполнение
как воспроизводится прошлый эксперимент
```

Kernel не должен заранее знать:

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

# 4. Минимальная цифровая физика

Начальный мир должен быть очень простым.

Примитивы:

```text
Matter
Energy
Position
Connection
Signal
Memory
ExecutableRule
```

Минимальные операции:

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

Вместо заранее заданного организма должны существовать структуры, которые могут собираться из этих примитивов.

---

# 5. Executable Matter

Одна из ключевых идей:

> информация внутри мира должна быть физически действенной.

То есть структура может хранить программу:

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

Такая программа может:

- читать локальную среду;
- менять состояние;
- перемещать материю;
- управлять потоками энергии;
- копировать себя;
- передавать сигналы;
- создавать новые структуры;
- запускать другие программы.

Тогда "геном" — это не просто набор параметров, а исполняемая программа.

---

# 6. DSL / виртуальная машина

Вместо прямого изменения Go/C++-кода лучше создать безопасный DSL.

Пример:

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

Другой пример:

```text
rule Phototaxis {
    sensor:
        light_gradient

    action:
        move toward light_gradient
}
```

Преимущества:

- sandbox;
- hot reload;
- rollback;
- versioning;
- быстрый запуск множества ветвей;
- возможность эволюции кода;
- удобная генерация правил внешним AI.

---

# 7. Эволюция собственных примитивов

Очень важный механизм:

> система должна уметь создавать новые reusable building blocks.

Например:

```text
FOO :=
    LOAD
    COPY
    BIND
```

После появления `FOO` он становится новым примитивом.

Позже:

```text
BAR :=
    FOO
    SIGNAL
    EXEC
```

Получается:

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

Это один из возможных механизмов настоящего роста сложности.

---

# 8. Три уровня эволюции

## 8.1. Микроэволюция

Происходит постоянно внутри мира:

```text
mutation
recombination
copy errors
resource competition
inheritance
local selection
```

AI здесь не нужен.

---

## 8.2. Макроэволюция механизмов

Когда мир:

- стагнирует;
- теряет разнообразие;
- приходит к одной доминирующей стратегии;
- переживает массовое вымирание;
- обнаруживает необычную устойчивую структуру;

внешний агент получает агрегированное описание мира.

Он предлагает не "интересный контент", а изменения типа:

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

## 8.3. Эволюция онтологии

Самый сильный уровень.

Система может обнаружить, что в мире появилась новая стабильная категория.

Например было:

```text
Matter
Energy
Connection
Signal
```

Но со временем возникло:

```text
устойчивые сигнальные последовательности
+
хранение паттернов
+
повторное использование
+
контекстная интерпретация
```

AI может предложить новый абстрактный primitive:

```text
Symbol
```

После чего символы можно:

```text
copy
combine
store
transmit
interpret
transform
```

И потенциально появляется:

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

# 9. Эволюция evolvability

Система должна уметь менять не только геномы, но и сами механизмы эволюции.

Потенциально изменяемые параметры:

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

То есть:

```text
evolution
↓
evolution of evolution
```

---

# 10. Major Evolutionary Transitions

Хотелось бы, чтобы система могла сама переходить через новые уровни индивидуальности:

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

Ключевой критерий:

> новая единица отбора должна возникать внутри мира, а не быть заранее задана классом.

---

# 11. Коэволюция среды

Не использовать фиксированную fitness-функцию:

```text
fitness = f(agent)
```

Вместо этого:

```text
fitness_t = f(
    agent,
    environment_t,
    population_t,
    history_t
)
```

Среда тоже меняется:

```text
environment_{t+1}
=
g(
    environment_t,
    population_t
)
```

Пример:

```text
появился фотосинтез
↓
изменилась атмосфера
↓
появились новые ниши
↓
возник новый метаболизм
↓
он снова изменил среду
```

Ключевой принцип:

> решение одной эволюционной задачи должно создавать новые задачи.

---

# 12. Внешний AI-агент

AI не должен быть "богом", который проектирует конечные формы жизни.

Лучше роль:

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

AI расширяет пространство возможностей.

Эволюция решает, будет ли новый механизм использоваться.

---

# 13. Что получает AI

Не весь мир, а иерархическую telemetry.

## Базовые метрики

```text
population count
birth/death rates
energy distribution
resource consumption
average lifespan
genetic diversity
behavioral diversity
```

## Структурные метрики

```text
persistent structures
graph complexity
modularity
hierarchy depth
community structure
```

## Информационные метрики

```text
entropy
mutual information
predictive information
signal reuse
memory usage
information flow
```

## Эволюционные метрики

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

# 14. Формат запроса к внешнему агенту

Пример:

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

# 15. Формат ответа агента

AI должен возвращать структурированный patch proposal.

Например:

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

Каждое изменение:

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

# 17. AI как генератор архитектурных мутаций

Обычная мутация:

```text
0.31 → 0.34
```

LLM может сделать:

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

То есть иногда:

```text
Δparameter
```

заменяется на:

```text
Δarchitecture
```

Это может быть одним из главных преимуществ LLM в OEE-системе.

---

# 18. Дерево миров

Не выбирать одну "лучшую" ветвь.

```text
                    World 0
                 /     |     \
               W1      W2      W3
              /  \             / \
            W4    W5          W6  W7
                    \
                     W8
```

Сохранять разные миры по разным критериям:

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

Хранить не только текущих победителей, но и открытия:

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

# 20. Как измерять новизну

Не только "стал ли мир эффективнее".

Возможные критерии:

```text
behavioral novelty
structural novelty
functional novelty
causal novelty
ecological novelty
representation novelty
information-processing novelty
```

Особенно сильный признак:

> новая структура открывает тип взаимодействий, которого раньше не существовало.

---

# 21. Causal Emergence

Можно использовать causal emergence как аналитический инструмент.

Идея:

> искать coarse-graining, на котором макроописание имеет больше причинной предсказательной силы, чем микроописание.

Потенциальная цепочка:

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

Это может стать механизмом автоматического поиска новых уровней организации.

---

# 22. Viability / Free Energy Perspective

Организм можно определять не классом, а как систему, которая удерживает себя в области жизнеспособных состояний.

```text
state(t) ∈ viability region
```

Внутренние ограничения:

```text
energy reserves
integrity
temperature-like state
resource balance
signal coherence
```

Система может изменять мир, чтобы оставаться жизнеспособной.

---

# 23. Physics of Life

Мир должен быть неравновесным.

Нужны постоянные потоки:

```text
energy source
↓
energy capture
↓
work
↓
waste / dissipation
```

Если энергетический градиент исчезает, сложность должна деградировать.

---

# 24. Культурная эволюция

После появления сложной коммуникации можно допустить отдельный канал наследования:

```text
ideas
skills
symbols
recipes
strategies
construction programs
```

Тогда существуют:

```text
genetic inheritance
+
cultural inheritance
```

И со временем:

```text
genetic evolution
↕
cultural evolution
```

---

# 25. Внешние артефакты

Существа должны потенциально уметь создавать долговечные внешние объекты.

Не хардкодить понятия:

```text
nest
tool
storage
trap
machine
```

Пусть это будут просто устойчивые физические структуры.

Если они начнут использоваться функционально — семантику можно обнаружить позже.

---

# 26. Технологическая эволюция

Особенно интересный переход:

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

Если такие уровни возникают без заранее прописанной "technology tree", это сильный признак open-endedness.

---

# 27. Внутренние виртуальные машины

Радикальная идея:

> существа внутри мира могут построить собственную VM.

Например:

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

Аналогия:

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

Каждый слой создаёт новое пространство эволюции.

---

# 28. Самомодифицирующийся код

Разрешить менять:

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

Но не immutable kernel.

---

# 29. Безопасность AI-патчей

AI-generated code должен выполняться:

- без доступа к host OS;
- без произвольного filesystem;
- без сети;
- с лимитом CPU;
- с лимитом RAM;
- с instruction budget;
- с timeout;
- с deterministic replay там, где возможно.

Предпочтительно:

```text
custom DSL
или
WASM-like sandbox
```

а не native-code hot patch.

---

# 30. Reproducibility

Каждое изменение хранит:

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

Любую ветвь можно воспроизвести.

---

# 31. Архитектура сервисов

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

# 32. Возможная структура репозитория

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

# 33. Технологический стек

Для быстрого MVP:

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

Если симуляция станет CPU-bound:

```text
core → Rust/C++
```

GPU имеет смысл подключать позже.

---

# 34. Этапы реализации

---

## Этап 0. Исследовательский каркас

### Цель

Получить минимальный воспроизводимый мир.

### Сделать

- tick-based simulation;
- seeded RNG;
- snapshot;
- replay;
- базовую карту;
- базовую энергию;
- resource accounting;
- benchmark.

### Критерий завершения

Один и тот же snapshot + seed даёт одинаковое продолжение.

---

## Этап 1. Исполняемые репликаторы

### Добавить

```text
executable genome
copy
mutation
death
resource consumption
```

### Не использовать

```text
Species
Predator
FitnessScore
```

### Критерий

Появляются разные наследуемые линии.

---

## Этап 2. Мини-экология

Добавить:

```text
несколько ресурсов
локальную конкуренцию
возможность отбирать ресурсы
связи между структурами
```

### Критерий

Возникает минимум несколько устойчивых стратегий.

---

## Этап 3. DSL / VM

Вынести изменяемые правила мира в DSL.

### Реализовать

- parser;
- bytecode;
- instruction budget;
- sandbox;
- rule versioning;
- hot reload;
- rollback.

### Критерий

Новое правило можно загрузить без перезапуска kernel.

---

## Этап 4. Telemetry

Собирать:

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

### Критерий

Можно автоматически объяснить, что происходило в мире последние N тысяч ticks.

---

## Этап 5. Novelty + Stagnation Detector

Сначала простые эвристики:

```text
diversity plateau
population monoculture
repeated behavioral hashes
structural plateau
```

Позже добавить embedding / graph-based novelty.

### Критерий

Система умеет автоматически сказать:

```text
"мир развивается"
или
"мир застрял"
```

---

## Этап 6. AI Observer

AI пока ничего не меняет.

Он получает telemetry summary и отвечает:

```text
что доминирует
какие ниши существуют
какие структуры появились
почему могла возникнуть стагнация
```

### Критерий

Вывод AI можно сопоставить с реальными логами и визуализацией.

---

## Этап 7. AI Macro-Mutations

Разрешить агенту предлагать новые правила.

Например:

```text
new sensor
new resource reaction
new binding rule
new communication channel
```

### Главное правило

AI не создаёт готовую адаптацию.

Плохо:

```text
дать всем существам крылья
```

Хорошо:

```text
добавить физический механизм подъёмной силы
```

и посмотреть, воспользуется ли им эволюция.

---

## Этап 8. Branching Worlds

Для каждого AI patch:

```text
N variants × M seeds
```

Запускать параллельно.

### Оценивать

```text
novelty
diversity
persistence
structural complexity
niche count
```

### Критерий

Система умеет хранить несколько перспективных ветвей вместо одного winner.

---

## Этап 9. Novelty Archive + Pareto Selection

Не сводить всё к одной fitness-функции.

Хранить Pareto-front по:

```text
novelty
diversity
complexity
persistence
new hierarchy
new information processing
```

### Критерий

Редкие странные ветви не исчезают только потому, что они пока "менее эффективны".

---

## Этап 10. Эволюция evolvability

Разрешить геномам менять:

```text
mutation rate
mutation operators
recombination
copy fidelity
inheritance format
```

### Критерий

Разные линии используют разные стратегии собственной изменчивости.

---

## Этап 11. Коэволюция среды

Существа получают возможность заметно менять мир:

```text
resource distribution
local chemistry
energy gradients
terrain-like state
signal fields
```

### Критерий

Новая адаптация меняет среду так, что появляются новые ниши.

---

## Этап 12. Протомногоклеточность

Добавить только минимальные механизмы:

```text
persistent binding
resource sharing
signals
local specialization
```

Не вводить класс `MulticellularOrganism`.

### Критерий

Некоторые группы устойчивее отдельного репликатора и начинают воспроизводиться как целое.

---

## Этап 13. Automatic Entity Discovery

Искать:

```text
persistent boundaries
internal coordination
resource sharing
common reproduction
information closure
```

### Результат

Иерархия:

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

## Этап 14. Causal Emergence Analysis

Пытаться найти макроуровень, который лучше предсказывает будущее системы.

### Цель

Автоматически отличать:

```text
случайную группу
```

от:

```text
новой причинно значимой сущности
```

---

## Этап 15. Символический слой

Разрешить:

```text
persistent signals
signal composition
memory
arbitrary mapping
context-sensitive interpretation
```

### Критерий

Появляются сигнальные структуры, значение которых определяется не только физическим стимулом.

---

## Этап 16. Cultural Evolution

Добавить горизонтальное копирование:

```text
behavior programs
skills
symbol sequences
construction recipes
```

### Критерий

Информация распространяется независимо от генетической линии.

---

## Этап 17. Persistent Artifacts

Разрешить долгоживущие внешние структуры.

### Критерий

Возникают объекты, которые:

- переживают создателя;
- используются другими;
- становятся частью адаптивной стратегии.

---

## Этап 18. Technology-like Evolution

Разрешить конструкциям создавать конструкции.

### Цепочка

```text
tool
↓
tool-making tool
↓
machine
↓
construction network
```

### Критерий

Внешние артефакты начинают создавать новое пространство эволюционных возможностей.

---

## Этап 19. Self-Hosted VM

Самая амбициозная стадия.

Существа должны иметь возможность реализовать:

```text
memory
instruction encoding
decoder
execution loop
```

### Критерий

В мире появляется новый вычислительный слой поверх исходной VM.

---

## Этап 20. Recursive Evolution

Если появилась внутренняя VM:

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

Исследовать возможность:

```text
evolution inside evolution
```

---

## Этап 21. Strong OEE Experiment

Длительные эксперименты.

Следить:

```text
does novelty saturate?
does complexity saturate?
do new abstractions continue appearing?
do new levels of organization emerge?
does evolvability continue changing?
```

Главный критерий:

> рост разнообразия и адаптивной новизны не останавливается на очевидном потолке, непосредственно заданном разработчиком.

---

# 35. Основной сравнительный эксперимент

Сделать три режима.

## A. Control

```text
обычная внутренняя эволюция
без AI
```

## B. Parameter AI

AI меняет только числа:

```text
mutation rate
resource rate
energy cost
```

## C. Structural AI

AI может создавать:

```text
new rules
new interaction primitives
new inheritance mechanisms
new representations
```

Сравнивать:

```text
novelty over time
time to stagnation
structural complexity
persistent innovations
niche count
major transitions
```

Если C стабильно превосходит A/B — есть сильное основание развивать подход.

---

# 36. Критерии успеха

## Минимальный

- мир живёт долго;
- линии конкурируют;
- разнообразие не исчезает мгновенно;
- возникают неожиданные стратегии.

## Средний

- новые ecological niches;
- новые способы наследования;
- кооперация;
- паразитизм;
- устойчивые коллективы.

## Сильный

- новые уровни индивидуальности;
- symbol-like representations;
- cultural inheritance;
- persistent artifacts;
- новые вычислительные абстракции.

## Очень сильный

- self-created VM;
- новые языки;
- повторяющиеся major transitions;
- длительная novelty без saturation.

---

# 37. Что не делать в MVP

Не начинать с:

- Unreal Engine;
- 3D;
- realistic fluid physics;
- красивой графики;
- LLM у каждого NPC;
- политики;
- экономики;
- огромного мира;
- миллионов агентов;
- CUDA.

Главная неизвестная:

> возникает ли вообще интересная open-ended динамика?

Сначала доказать это.

---

# 38. Самый маленький MVP

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

Цель:

> проверить, способен ли внешний AI систематически выводить эволюцию из локальных тупиков, не задавая конечное решение напрямую.

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

# 40. Визуализация

Сначала scientific UI.

Показывать:

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

Особенно интересен UI истории онтологии:

```text
S0
↓
S1
↓
S2
↓
...
```

с пояснением:

```text
"в этой точке впервые возник persistent signaling"
"здесь появился новый inheritance channel"
"здесь collective стал единицей отбора"
```

---

# 41. Игровой вариант

После исследовательского прототипа из системы можно сделать игру.

Роль игрока:

```text
observer
scientist
curator
explorer
intervention agent
```

Самый интересный вариант:

> игрок физически существует внутри мира, но мир продолжает собственную OEE вокруг него.

Тогда прохождения могут радикально отличаться:

```text
World A → symbiotic super-organisms
World B → mobile colonies
World C → machine ecology
World D → symbolic civilization
```

---

# 42. Игровая мета-механика

Игрок не обязан "побеждать".

Он может:

- обнаруживать новые формы жизни;
- исследовать странные ветви эволюции;
- вмешиваться;
- сравнивать параллельные миры;
- восстанавливать причины major transition;
- путешествовать между branch worlds;
- сохранять редкие линии;
- запускать controlled perturbations.

По сути:

> "No Man's Sky", но генерируются не только формы мира — генерируются новые законы, экологии и уровни организации.

---

# 43. Самая сильная версия идеи

Стартовый разработчик создаёт только:

```text
space
matter
energy
connections
signals
execution
resource limits
```

А потом наблюдает:

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

Последний `???` — главный результат проекта.

Если заранее известно всё, что может появиться, это слабая версия OEE.

Если система начинает порождать категории, которые разработчик не проектировал непосредственно, эксперимент становится действительно интересным.

---

# 44. Итоговая схема

```text
MICRO EVOLUTION
миллионы дешёвых локальных изменений
        ↓
WORLD DYNAMICS
экология + изменение среды
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

Главный принцип:

> внешний AI не проектирует конечные формы жизни. Он расширяет пространство возможных механизмов, а естественный отбор внутри симуляции решает, что из этого действительно закрепится.

---

# 45. Первый практический спринт

Если начинать реализацию прямо сейчас, первый спринт можно ограничить следующим.

## 1. Kernel

```text
World
Cell
Energy
RuleVM
Tick
Snapshot
```

## 2. Репликатор

Минимальный genome:

```text
SENSE_RESOURCE
MOVE
ABSORB
COPY
```

## 3. Мутации

```text
replace instruction
insert instruction
delete instruction
duplicate block
```

## 4. Метрики

```text
population
genome hash
lineage
energy
birth/death
behavior hash
```

## 5. Stagnation

Если:

```text
novelty_rate < threshold
```

на протяжении большого окна:

```text
trigger external agent
```

## 6. Первый AI patch

AI разрешено добавить ровно один новый DSL primitive.

Например:

```text
SEND_SIGNAL
```

Создать:

```text
control branch
+
patch branch
```

и сравнить их.

Это уже будет минимальный законченный эксперимент всей идеи.

---

# 46. Дальнейший главный milestone

После первого working loop:

```text
world
→ stagnation
→ AI patch
→ branch
→ selection
```

следующий действительно значимый milestone:

> получить первое новое устойчивое взаимодействие, которое не было конечной целью AI-патча.

Например:

AI добавил сигнализацию ради кооперации,

а эволюция использовала её для:

```text
паразитической имитации
или
обмана
или
территориальной маркировки
```

Вот такой результат будет гораздо более интересным, чем если мир просто использует механизм ровно так, как предполагал агент.

---

# 47. Конечная исследовательская цель

В идеале система должна перейти от:

```text
"AI придумывает новые правила"
```

к:

```text
"AI замечает, что эволюция сама приблизилась к новому уровню,
и лишь делает этот уровень доступным для дальнейшего исследования"
```

То есть внешний агент постепенно превращается из дизайнера в:

```text
наблюдателя
+
гипотезогенератор
+
оператор расширения пространства возможностей
```

И именно эта архитектура наиболее интересна как возможный мост между:

- classical evolutionary computation;
- artificial life;
- open-ended evolution;
- AI coding agents;
- causal emergence;
- digital physics;
- cultural evolution;
- self-modifying software.


---

# 48. Уточнение: в мире нет "проблем", есть только последствия

В базовой модели не должно существовать объекта или признака:

```text
Problem
```

Сама эволюция ничего не "считает проблемой".

Внутри мира происходят только:

```text
изменения среды
↓
изменения вероятности сохранения структур
↓
изменения вероятности репликации
↓
изменение состава популяции
```

Например:

```text
ресурсов стало меньше
↓
часть репликаторов исчезла
↓
другие продолжили существовать
```

Для мира это не "проблема нехватки ресурсов".

Это просто динамика.

Понятие проблемы появляется только у внешнего наблюдателя.

Поэтому правильнее говорить не:

```text
эволюция решает проблему
```

а:

```text
изменились условия
↓
изменились селективные давления
↓
некоторые структуры оказались более устойчивыми
```

---

# 49. Минимальная направленность без fitness-функции

Естественный отбор не требует явной функции:

```text
fitness(x)
```

Достаточно трёх свойств:

```text
variation
+
inheritance
+
differential persistence / reproduction
```

Если:

```text
A существует 10 ticks и исчезает

B копирует себя

C копирует себя ещё эффективнее
```

то через длительное время потомки B/C присутствуют, а A — нет.

Никто не задавал:

```text
goal = reproduce
```

Направленность возникает статистически:

> структуры, которые не оставляют причинных продолжений, перестают присутствовать в будущем состоянии мира.

---

# 50. Фундаментально все состояния нейтральны

С точки зрения базовой физики:

```text
организм продолжает существовать
```

и:

```text
организм распался
```

— просто два возможных состояния.

Вселенная не присваивает им:

```text
good
bad
success
failure
```

Это означает:

> "выживание" не должно быть встроенной моральной или целевой функцией мира.

Оно возникает как наблюдаемая статистическая асимметрия.

Мы видим долгоживущие структуры именно потому, что короткоживущие структуры уже исчезли.

---

# 51. Affordances вместо "проблем"

Более полезное понятие для OEE:

```text
affordance
```

Affordance — возможность взаимодействия, которая существует относительно конкретной структуры.

Например один и тот же объект может быть:

```text
для организма A → источник энергии
для организма B → токсин
для организма C → сигнал
для организма D → нейтральный фон
```

То есть нет:

```text
environment.problem = X
```

Есть:

```text
Affordance(agent, environment)
```

Новая динамика мира создаёт новые affordances.

Пример:

```text
организм A производит вещество X
↓
X накапливается
↓
организм B случайно способен использовать X
↓
появляется новая экологическая ниша
↓
B меняет среду
↓
возникают новые affordances
```

Никто не объявлял:

```text
Problem: use X
```

---

# 52. Значение возникает относительно структуры

До появления фототаксиса:

```text
градиент света
```

— просто физическая величина.

После появления соответствующего механизма:

```text
градиент света
```

становится информацией.

До появления копирующей machinery:

```text
последовательность
```

— просто структура.

После появления наследования:

```text
последовательность
```

становится генетической информацией.

То есть семантика возникает не как свойство среды самой по себе, а как отношение:

```text
structure ↔ environment
```

---

# 53. Внешний AI не должен быть Problem Solver

Первоначальную архитектуру лучше изменить.

Плохо:

```text
world has problem
↓
AI finds solution
```

Потому что это скрытая телеология.

Лучше:

```text
world develops new dynamics
↓
AI observes
↓
AI expands possibility space
↓
evolution exploits or ignores new mechanisms
```

Роль AI:

```text
AI Possibility Expander
```

а не:

```text
AI Problem Solver
```

---

# 54. Нейтральные расширения мира

AI можно периодически просить создавать:

```text
new local interaction primitive
new energy transformation
new signaling mechanism
new binding property
new state transition
new storage mechanism
```

При этом запрещать:

```text
"помоги виду X"
"реши проблему Y"
"сделай организм сильнее"
```

Хороший prompt:

```text
Предложи минимальное локальное расширение физики мира,
которое:
- имеет стоимость;
- не даёт прямого преимущества ни одной линии;
- допускает несколько потенциальных применений;
- не задаёт конечную функцию;
- может быть использовано или проигнорировано эволюцией.
```

---

# 55. Мир должен сам задавать направление расширения

Ещё более сильный вариант:

AI не генерирует случайные новые механизмы.

Он ищет структуры, которые уже начали возникать снизу.

Например:

```text
тысячи локальных связей
↓
возникает повторяющийся устойчивый паттерн
↓
AI обнаруживает latent structure
↓
предлагает сделать её новым primitive
```

Пример:

```text
сложная повторяющаяся граница
↓
candidate "membrane"
↓
Membrane становится оптимизированным primitive
↓
эволюция начинает строить структуры из Membrane
```

Позже:

```text
Membrane + metabolism + replication
↓
candidate higher-level unit
↓
новый primitive
```

---

# 56. Emergent Ontology Compilation

Эту идею можно формализовать как:

```text
emergent ontology compilation
```

Схема:

```text
микродинамика
↓
устойчивая повторяющаяся макроструктура
↓
обнаружение causal / functional coherence
↓
AI предлагает абстракцию
↓
абстракция компилируется в новый primitive
↓
новый primitive становится строительным блоком
↓
возникают ещё более сложные структуры
```

То есть:

```text
S0 → S1 → S2 → S3
```

где каждый следующий язык мира частично строится из устойчивых закономерностей предыдущего.

---

# 57. Но внешний метакритерий всё равно нужен

Если разрешить множество ветвей миров, часть из них будет:

```text
деградировать
упрощаться
переходить в хаос
замыкаться в одном паттерне
терять разнообразие
создавать случайный шум
```

Поэтому исследовательскому слою нужна оценка.

Но это не должна быть одна fitness-функция вроде:

```text
score = complexity
```

Иначе вся система начнёт оптимизировать наш измеритель.

Нужен:

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

# 58. Почему одной "сложности" недостаточно

Можно создать систему с огромной сложностью, которая не обладает интересной эволюцией.

Например:

```text
случайный шум
```

может иметь:

```text
высокую энтропию
низкую сжимаемость
много различных состояний
```

но почти нулевую устойчивую структуру.

С другой стороны:

```text
идеальный кристалл
```

очень структурирован, но почти не создаёт новизны.

Поэтому желаемая область находится между:

```text
rigid order
↔
adaptive structured novelty
↔
random chaos
```

---

# 59. Три класса деградации

## 59.1. Коллапс

```text
population → 0
energy loops disappear
persistent structures disappear
```

## 59.2. Замораживание

```text
одна стратегия доминирует
diversity → low
novelty → 0
world repeats itself
```

## 59.3. Хаотизация

```text
entropy high
structures short-lived
inheritance low
causal persistence low
```

Все три режима могут считаться нежелательными для OEE-эксперимента.

---

# 60. Вместо одной оценки — вектор качества мира

Для каждой ветви считать:

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

Не сводить этот вектор сразу к одному числу.

---

# 61. Viability

Показывает, существует ли вообще достаточно устойчивый процесс.

Возможные показатели:

```text
population persistence
energy throughput
reproduction continuity
mean lineage duration
fraction of persistent structures
```

Это baseline-фильтр.

Если:

```text
viability < minimum
```

ветвь можно не развивать дальше.

---

# 62. Diversity

Несколько типов разнообразия:

```text
genetic diversity
behavioral diversity
structural diversity
ecological diversity
functional diversity
```

Важно не считать разнообразием просто случайные различия.

---

# 63. Persistence

Новизна должна жить достаточно долго.

Например новый паттерн, существовавший:

```text
3 ticks
```

скорее шум.

А паттерн, который:

```text
появился
↓
распространился
↓
повлиял на другие структуры
↓
сохранился 100000 ticks
```

гораздо интереснее.

Можно определить:

```text
persistence_score =
lifetime × descendants × ecological_impact
```

---

# 64. Raw Novelty

Измеряет:

> насколько новое состояние отличается от архива прошлого.

Можно использовать:

```text
behavior embeddings
graph embeddings
genome embeddings
structure descriptors
interaction graphs
```

Например:

```text
novelty(x)
=
distance(
    descriptor(x),
    nearest historical descriptors
)
```

Но raw novelty легко награждает бессмысленный шум.

---

# 65. Adaptive Novelty

Поэтому нужна более сильная метрика:

```text
adaptive novelty
```

Новая структура считается интересной, если она:

1. раньше не наблюдалась;
2. сохраняется;
3. влияет на собственную устойчивость или воспроизводство;
4. меняет взаимодействия с другими структурами;
5. потенциально создаёт новые affordances.

Условно:

```text
AdaptiveNovelty
=
Novelty
× Persistence
× FunctionalImpact
```

Не как окончательная формула, а как идея.

---

# 66. Functional Complexity

Нужно различать:

```text
сложность структуры
```

и:

```text
сложность причинно полезной организации
```

Например считать:

```text
number of interacting subsystems
dependency graph depth
division of labor
number of coordinated functions
conditional behavior complexity
```

---

# 67. Causal Structure

Очень важный анти-шумовой фильтр.

Сложная система должна иметь:

```text
устойчивые причинные зависимости
```

а не просто случайную вариативность.

Возможные метрики:

```text
predictive information
effective information
transfer entropy
causal graph stability
macro-level predictive gain
```

Если структура сложна, но её части почти не влияют друг на друга устойчивым образом, это подозрение на шум.

---

# 68. Hierarchy Depth

Open-ended evolution особенно интересна, если появляются новые уровни:

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

Можно измерять:

```text
number of stable compositional levels
```

или:

```text
depth of reusable abstraction hierarchy
```

---

# 69. Niche Creation

Очень сильная метрика.

Новая адаптация интереснее, если она не только занимает нишу, но и создаёт новые.

Пример:

```text
новая структура
↓
создала новый ресурс
↓
другая линия начала использовать ресурс
↓
возникло новое взаимодействие
```

Можно считать:

```text
new affordances created
new resource loops
new interaction edges
new persistent dependent lineages
```

---

# 70. Evolvability

Измеряет:

> насколько система способна производить полезную наследуемую вариативность дальше.

Возможные proxy:

```text
fraction of viable mutations
diversity of descendant phenotypes
rate of persistent innovations
ability to escape local attractors
number of distinct accessible strategies
```

Очень важно:

```text
текущая эффективность
```

не должна уничтожать:

```text
будущую изменчивость
```

---

# 71. Ontology Growth

Самая интересная долгосрочная метрика.

Считать появление новых reusable abstractions:

```text
new primitives
new entity types discovered
new interaction categories
new inheritance channels
new representation systems
new levels of computation
```

Особенно важно не просто количество.

Новая абстракция должна:

```text
использоваться
комбинироваться
создавать новые производные структуры
```

---

# 72. Метрика "Generativity"

Можно ввести отдельную идею:

```text
Generativity
```

Вопрос:

> насколько новшество увеличило пространство последующих возможных инноваций?

Условный критерий:

```text
Generativity(x)
=
number of new persistent innovations
that become reachable after x
```

Это сложно измерить напрямую.

Практический proxy:

```text
branch world with innovation X
vs
counterfactual branch without X
```

Сравнить последующую novelty через большой горизонт.

Если после X:

```text
novelty rate ↑
niche count ↑
new primitives ↑
hierarchy depth ↑
```

то X было генеративным.

---

# 73. Контрфактуальные ветви

Для оценки настоящей ценности нового механизма полезно делать:

```text
World A = до изменения
World B = с изменением
```

Оба запускаются с:

```text
одинакового snapshot
одинаковых initial seeds
```

после чего расходятся.

Можно оценить:

```text
Δnovelty
Δdiversity
Δhierarchy
Δniche creation
Δevolvability
```

Это намного лучше, чем просто смотреть на абсолютные значения.

---

# 74. Novelty без деградации

Полезно разделить:

```text
novelty
```

и:

```text
productive novelty
```

Пример:

```text
random mutation explosion
→ high raw novelty
→ low persistence
→ low causal structure
→ low productive novelty
```

Другой пример:

```text
new signaling protocol
→ moderate raw novelty
→ high persistence
→ new cooperation
→ new niches
→ high productive novelty
```

---

# 75. Минимальные фильтры ветви

Перед попаданием в долгосрочный archive ветвь должна пройти минимальные ограничения:

```text
viability > Vmin
persistence > Pmin
causal_structure > Cmin
```

Это не "цель эволюции".

Это фильтр исследовательского эксперимента.

---

# 76. Pareto Frontier вместо одного score

Не использовать:

```text
Score =
0.4 novelty
+ 0.3 complexity
+ 0.3 diversity
```

Иначе система быстро научится эксплуатировать конкретные веса.

Вместо этого использовать Pareto selection.

Например:

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

Все три можно сохранить.

---

# 77. Quality-Diversity Archive

Подход похож на quality-diversity / MAP-Elites.

Хранить лучшие миры в разных областях пространства характеристик.

Например оси:

```text
diversity
hierarchy depth
resource-cycle complexity
communication complexity
ontology depth
```

Это позволяет сохранять странные, но потенциально перспективные линии.

---

# 78. Защита от metric gaming

Если AI знает точную функцию:

```text
score(world)
```

он может начать оптимизировать её формально.

Например:

```text
нужно diversity
↓
AI создаёт миллион бессмысленных случайных типов
```

Поэтому:

1. использовать несколько метрик;
2. часть метрик скрывать от генератора патчей;
3. регулярно менять evaluator;
4. использовать human/AI qualitative review только как дополнительный слой;
5. использовать counterfactual tests;
6. проверять persistence;
7. проверять causal impact;
8. проверять generativity.

---

# 79. Surprise vs Novelty vs Progress

Разделять три понятия.

## Surprise

```text
неожиданно относительно модели
```

## Novelty

```text
раньше такого не было
```

## Progress

```text
новшество увеличивает дальнейшие возможности системы
```

OEE интересует прежде всего:

```text
persistent generative novelty
```

а не просто surprise.

---

# 80. Предлагаемая мета-оценка

Не одна формула, а pipeline:

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

Это намного устойчивее, чем:

```text
complexity_score > threshold
```

---

# 81. Как AI должен выбирать ветви

AI может помогать анализировать:

```text
"что здесь новое?"
"что здесь функционально?"
"какие новые affordances появились?"
"возник ли новый уровень организации?"
```

Но финальное решение желательно опирать на:

```text
measured metrics
+
counterfactual branches
+
archive comparison
```

а не только на текстовое мнение LLM.

---

# 82. Эталонная схема внешнего селектора

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

# 83. Важное философское разделение

Нужно сохранять два разных уровня.

## Внутри мира

Нет:

```text
good
bad
problem
progress
goal
```

Есть только:

```text
physics
interactions
persistence
inheritance
consequences
```

## На уровне исследователя

Есть цель эксперимента:

```text
найти системы,
которые продолжают создавать
устойчивую адаптивную новизну
```

То есть оценочная функция существует не как закон мира, а как:

```text
experimental selection criterion
```

---

# 84. Самая сильная версия внешнего критерия

Возможно, лучший вопрос не:

```text
"стала ли система сложнее?"
```

а:

> "стало ли после этого изменения больше способов стать сложнее в будущем?"

То есть оценивать не только состояние:

```text
Complexity(t)
```

а производную возможностей:

```text
FuturePossibilityGrowth
```

Условно:

```text
OEE quality
≈
persistent increase in accessible adaptive possibilities
```

Это ближе всего к сильной концепции open-ended evolution.

---

# 85. Обновлённая архитектура цикла

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

Ключевая идея:

> внутри мира нет "проблем". Снаружи есть только исследовательский критерий: сохранять ветви, которые демонстрируют устойчивую, причинно значимую и генеративную новизну.



---

# 86. Следующий уровень: оценивать не состояние, а пространство будущих возможностей

Главная слабость обычных метрик сложности:

```text
complexity(world_t)
```

измеряет только текущее состояние.

Для OEE важнее другое:

> создаёт ли текущее состояние новые возможные траектории дальнейшей эволюции?

Поэтому полезно мыслить через:

```text
Future Possibility Space
```

Условно:

```text
FPS(world_t)
=
множество качественно различных устойчивых состояний,
достижимых из текущего мира за горизонт T
```

Тогда интерес представляет не просто:

```text
FPS size
```

а:

```text
growth(FPS)
```

То есть:

```text
появляются ли новые классы достижимых будущих состояний
```

---

# 87. Generativity как центральная метрика

Можно формально ввести:

```text
Generativity(X)
```

для новшества `X`.

Интуиция:

> насколько появление X увеличило число последующих независимых инноваций?

Пример:

```text
X = новый тип межклеточной связи
```

После него становятся возможны:

```text
кооперация
специализация
ресурсный обмен
колонии
коллективная память
паразитирование на коллективе
защита коллектива
```

То есть X обладает высокой generativity.

---

# 88. Практическая оценка Generativity

Точное пространство будущего неизвестно.

Поэтому использовать выборку контрфактуальных ветвей.

Для новшества X:

```text
Snapshot S
```

создать:

```text
Branch A: X enabled
Branch B: X disabled
```

Для каждой ветки:

```text
K random seeds
×
T future ticks
```

После этого сравнить:

```text
new niches
new persistent structures
new interaction categories
new hierarchy levels
new inheritance mechanisms
new abstractions
```

Условно:

```text
G(X) =
ExpectedNoveltyFuture(X)
-
ExpectedNoveltyFuture(no X)
```

---

# 89. Оценивать не только количество, но и независимость новшеств

Проблема:

одна новая способность может создать тысячу почти одинаковых вариантов.

Например:

```text
1000 оттенков одного сигнала
```

Это не то же самое, что:

```text
сигнализация
+
коллективная память
+
новый канал наследования
```

Поэтому novelty archive должен кластеризовать новшества по:

```text
mechanism
function
causal role
interaction topology
representation
```

И считать:

```text
independent innovation classes
```

а не сырое число вариантов.

---

# 90. Open-Endedness Score не должен быть одной цифрой

Даже если нужен dashboard, лучше отображать несколько временных рядов:

```text
NoveltyRate(t)
GenerativityRate(t)
OntologyDepth(t)
HierarchyDepth(t)
Evolvability(t)
Diversity(t)
Persistence(t)
```

А затем отдельно:

```text
SaturationIndicators(t)
```

Например:

```text
novelty ↓
ontology growth = 0
niche creation ↓
same interaction patterns repeat
```

---

# 91. Детектор насыщения

Нужно различать:

```text
временное плато
```

и:

```text
структурное насыщение мира
```

Признаки структурного насыщения:

```text
1. новые геномы продолжают появляться;
2. поведение слегка меняется;
3. но новых функций нет;
4. новых ниш нет;
5. новых типов взаимодействий нет;
6. hierarchy depth не растёт;
7. ontology archive не получает новых классов.
```

Такой мир формально "эволюционирует", но не является open-ended.

---

# 92. Измерение семантической новизны

Особенно трудная задача:

> понять, что новый объект выполняет новую роль.

Необходимо анализировать не форму объекта, а его причинное использование.

Например:

```text
Structure A
```

может быть геометрически новой, но функционально делать то же самое.

А:

```text
Structure B
```

может выглядеть почти идентично старой, но начать использоваться как:

```text
memory
signal relay
energy store
construction template
```

Поэтому descriptor должен включать:

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

Для каждой устойчивой структуры можно строить:

```text
FunctionalSignature
```

Пример:

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

Novelty тогда измеряется не только по morphology, но и по:

```text
distance(FunctionalSignature)
```

---

# 94. Affordance Graph

Вместо списка сущностей можно строить граф возможностей:

```text
Entity / Structure
      ↓
can interact with
      ↓
Resource / Signal / Artifact / Other entity
      ↓
possible transformation
```

Пример:

```text
A --consume--> X
A --signal--> B
B --transform--> X
C --attach--> A
```

Когда появляется новый тип ребра:

```text
store
teach
imitate
delegate
encode
```

это сильнее, чем просто появление новой формы объекта.

---

# 95. Affordance Expansion Score

Можно оценивать:

```text
AES(t)
=
number of new persistent affordance classes
introduced over window Δt
```

Но учитывать только affordances, которые:

```text
реально используются
сохраняются
влияют на дальнейшую динамику
```

---

# 96. Эволюция отношений важнее эволюции объектов

Возможный ключевой принцип:

> OEE может расти в первую очередь за счёт появления новых типов отношений, а не новых типов объектов.

Например:

```text
есть два организма
```

но возникают отношения:

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

Каждый новый relation type резко увеличивает комбинаторное пространство дальнейшего развития.

---

# 97. Relation-First Ontology

Поэтому новая онтология может строиться не как:

```text
new EntityType
```

а как:

```text
new RelationType
```

Пример:

```text
transfer_energy
```

позже становится:

```text
share_energy
```

потом:

```text
conditional_share
```

потом:

```text
reciprocal_exchange
```

потом:

```text
credit-like relation
```

То есть социально-экономические структуры потенциально могут возникать из эволюции отношений.

---

# 98. Meta-Affordances

Особенно интересны affordances, которые создают другие affordances.

Например:

```text
language
```

не просто помогает передать сообщение.

Он создаёт возможность:

```text
instruction
promise
contract
teaching
coordination
planning
```

То есть:

```text
affordance → new affordance generator
```

Такие механизмы должны иметь особенно высокий generativity score.

---

# 99. Abstraction as Compression

Новый primitive можно считать полезным, если он позволяет компактнее описать множество повторяющихся процессов.

Например:

до появления `Membrane`:

```text
тысячи отдельных локальных bond rules
```

после:

```text
Membrane(...)
```

Если новая абстракция:

```text
сильно уменьшает описание
+
сохраняет предсказательную силу
```

это хороший кандидат на ontology compilation.

---

# 100. Minimum Description Length для новых примитивов

Можно использовать идею MDL:

```text
DescriptionLength(before)
vs
DescriptionLength(after abstraction)
```

Если:

```text
DL(after) << DL(before)
```

а предсказательная способность не ухудшается,

новая абстракция может быть настоящей структурой, а не выдумкой AI.

---

# 101. Три условия для компиляции новой онтологии

Новый primitive разрешается добавить только если:

```text
1. Compression
2. Persistence
3. Causal usefulness
```

То есть он:

- сокращает описание мира;
- устойчиво встречается;
- помогает предсказывать последствия.

Это снижает риск, что AI будет создавать искусственные категории ради novelty score.

---

# 102. Causal Emergence + Ontology Compilation

Процесс:

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

Получается:

```text
causal emergence
→
new ontology
→
new evolutionary building block
```

---

# 103. Новая роль AI: ученый-компилятор

Внешний агент лучше разделить на несколько ролей.

```text
Observer
Hypothesis Generator
Ontology Miner
Patch Generator
Critic
Experiment Designer
```

Особенно интересен:

```text
Ontology Miner
```

Он не придумывает сущности с нуля.

Он спрашивает:

> какие повторяющиеся макропаттерны уже существуют, но пока представлены только как тысячи микровзаимодействий?

---

# 104. Ensemble Evaluator

Нельзя давать одному AI одновременно:

```text
generate patch
+
evaluate patch
```

Иначе он начнёт оценивать собственные решения благосклонно.

Лучше:

```text
Agent A → proposes
Agent B → critiques
Metrics → measure
Agent C → interprets
Archive → decides retention
```

Финальный селектор опирается прежде всего на измеримые результаты.

---

# 105. Blind Evaluation

Для части экспериментов evaluator не должен знать:

```text
какой patch применён
кто его создал
какая гипотеза была
```

Он получает только:

```text
before / after telemetry
```

Это уменьшает confirmation bias.

---

# 106. Hidden Metrics

Не все метрики следует показывать Patch Generator.

Например generator видит:

```text
world description
constraints
```

но не знает точную формулу:

```text
GenerativityEvaluator
```

Это снижает Goodhart effect.

---

# 107. Goodhart Resistance

Основное правило:

> когда метрика становится целью, она перестаёт быть хорошей метрикой.

Поэтому OEE evaluator должен регулярно проверять:

```text
metric gaming
```

Примеры:

```text
diversity score ↑
через бессмысленный noise

complexity score ↑
через огромные бесполезные структуры

novelty score ↑
через постоянную случайную смену состояний
```

Антидоты:

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

Отдельный агент может пытаться объяснить:

> почему якобы новое явление на самом деле не является прогрессом OEE.

Пример:

```text
"Новый тип поведения является лишь параметрической вариацией старого механизма."
```

Или:

```text
"Рост diversity вызван случайностью и не наследуется."
```

В archive попадают только кандидаты, выдержавшие такую критику.

---

# 109. Novelty Lineage

Каждая инновация должна иметь собственную историю:

```text
InnovationID
parent innovations
first appearance
lineages using it
dependent innovations
descendant affordances
```

Пример:

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

Это позволит измерять не только число инноваций, но и глубину их причинной генеалогии.

---

# 110. Innovation DAG

Лучше хранить не дерево, а DAG:

```text
Innovation A ─┐
              ├→ Innovation C
Innovation B ─┘
```

Потому что многие открытия возникают из комбинации предыдущих.

Сильная OEE должна показывать:

```text
increasing DAG depth
+
increasing recombination
```

---

# 111. Reuse как признак настоящей сложности

Если новый механизм возникает один раз и исчезает — это слабый результат.

Если механизм становится reusable building block:

```text
X
↓
используется в 20 независимых линиях
↓
комбинируется с Y и Z
↓
становится основой новых систем
```

это гораздо сильнее.

Можно считать:

```text
ReuseScore(X)
```

---

# 112. Compositionality Score

Важный признак open-endedness:

> новые элементы должны комбинироваться.

Например:

```text
A
B
C
```

дают:

```text
AB
AC
BC
ABC
```

Если каждая инновация независима и некомпонуема, пространство возможностей растёт медленно.

Если innovations compositional:

```text
possibility space
```

может расти комбинаторно.

---

# 113. Эволюция модульности

Полезно наблюдать, возникает ли:

```text
module
```

то есть часть системы, которая:

- имеет локальную функцию;
- переиспользуется;
- относительно независима;
- может комбинироваться с другими.

Модульность может оказаться одним из важнейших механизмов OEE.

---

# 114. Complexity Budget

Чтобы AI не создавал безгранично дорогие правила:

каждый новый primitive получает цену:

```text
compute cost
memory cost
energy cost
description cost
```

Новая возможность должна конкурировать за ограниченный бюджет.

Иначе:

```text
AI просто добавляет всё подряд
```

и пространство мира искусственно растёт без отбора.

---

# 115. Закон сохранения вычислительных ресурсов

Очень желательно:

```text
new capability ≠ free capability
```

Если добавляется:

```text
long-range signaling
```

он должен иметь цену:

```text
energy
latency
noise
memory
bandwidth
```

Это создаёт trade-offs.

Trade-offs — один из двигателей разнообразия.

---

# 116. Trade-off Generator

Новый primitive должен по возможности создавать:

```text
advantage A
↔
cost B
```

Примеры:

```text
быстрое размножение ↔ низкая точность

дальний сигнал ↔ высокая энергия

толстая мембрана ↔ медленный обмен

большая память ↔ вычислительная стоимость
```

Без trade-offs эволюция легко схлопнется в один универсальный лучший вариант.

---

# 117. Нет универсально лучшего организма

Архитектура мира должна стремиться к:

```text
context-dependent fitness
```

а не:

```text
global optimum
```

Идеально, если:

```text
Strategy A beats B
B beats C
C beats A
```

или эффективность зависит от:

```text
environment
population composition
history
local resource structure
```

Это поддерживает длительную коэволюцию.

---

# 118. Red Queen Dynamics

Желательный режим:

```text
вид A адаптируется к B
↓
B адаптируется к A
↓
A снова меняется
↓
...
```

Но важно:

обычная Red Queen dynamics может бесконечно вращаться внутри одного пространства стратегий.

Для OEE требуется иногда:

```text
Red Queen cycle
↓
new interaction mechanism
↓
new strategy space
```

---

# 119. Major Transition Detector

Автоматически искать признаки нового уровня индивидуальности.

Кандидат-группа должна иметь:

```text
persistent boundary
internal resource sharing
internal signaling
division of labor
common reproduction
reduced internal conflict
shared fate
```

Если эти признаки растут вместе:

```text
candidate major transition
```

---

# 120. Conflict Suppression как признак major transition

В истории эволюции новые уровни организации часто требуют подавления внутреннего конфликта.

Пример:

```text
клетки организма
```

не должны бесконечно конкурировать друг с другом.

Поэтому полезная метрика:

```text
internal competition
vs
collective fitness coupling
```

Если группа становится более целостной, внутренний конфликт должен снижаться или регулироваться.

---

# 121. Новые каналы наследования

Система должна отслеживать появление:

```text
genetic inheritance
epigenetic-like inheritance
horizontal transfer
behavioral imitation
cultural transmission
artifact inheritance
environmental inheritance
```

Каждый новый inheritance channel потенциально резко увеличивает evolvability.

---

# 122. Environmental Inheritance

Очень интересный вариант:

организм может передавать потомкам не информацию внутри себя, а изменённую среду.

Например:

```text
строит структуру
↓
умирает
↓
потомки используют структуру
```

Это уже наследование через:

```text
niche construction
```

---

# 123. Ecological Memory

Среда сама может хранить историю.

Например:

```text
chemical traces
persistent structures
resource depletion
constructed channels
symbol markers
```

Таким образом мир становится внешней памятью эволюции.

---

# 124. Evolutionary Memory Stack

Можно представить несколько слоёв памяти:

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

Появление нового слоя памяти — потенциальный major transition.

---

# 125. Time-Scale Separation

Важно моделировать разные временные масштабы.

Например:

```text
physics: каждый tick
behavior: десятки ticks
lifetime: тысячи ticks
ecology: миллионы ticks
ontology change: десятки миллионов ticks
```

Если все уровни меняются с одинаковой скоростью, стабильные структуры могут не успевать возникать.

---

# 126. Slow Macro-Evolution

AI macro-patches должны быть гораздо реже микроэволюции.

Например:

```text
micro mutation: постоянно
macro rule proposal: только после большого окна наблюдения
ontology compilation: ещё реже
```

Иначе внешний AI станет главным автором мира.

---

# 127. Budget внешнего вмешательства

Ввести:

```text
InterventionBudget
```

Например за миллион ticks AI может:

```text
добавить не более 1 primitive
изменить не более 0.1% rule space
```

Это позволяет проверить:

> может ли маленькое число семантических расширений поддержать огромный объём внутренней эволюции?

---

# 128. Минимальное вмешательство как исследовательский принцип

Лучший AI patch:

> минимальное изменение, которое максимально расширяет будущие возможности.

Это можно оптимизировать как:

```text
Generativity
----------------
PatchComplexity
```

То есть высокий:

```text
Generativity per added rule
```

---

# 129. Эквивалент Occam для OEE

Если два patch дают одинаковую generativity:

```text
Patch A = 2 new primitives
Patch B = 50 new primitives
```

предпочтительнее A.

Это позволяет не превращать AI в бесконечный генератор контента.

---

# 130. Открытый мир против расширяемого мира

Различать:

## Open state space

```text
огромное число состояний
```

и:

## Expanding state space

```text
появляются новые типы состояний
```

Для сильной OEE интереснее второе.

---

# 131. Онтологический event log

Отдельно хранить события:

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

Это станет "историей Вселенной".

---

# 132. Automatic Scientific Narration

AI Observer может строить научный журнал:

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

Это особенно полезно, если мир работает неделями или месяцами.

---

# 133. Branch Archaeology

Интересная функция:

> взять современное сложное явление и восстановить цепочку его происхождения.

Например:

```text
современная symbolic system
↓
какие innovation nodes были необходимы?
↓
какие могли отсутствовать?
↓
какой transition был критическим?
```

Можно автоматически запускать counterfactual branches из прошлого.

---

# 134. Causal Importance of Innovation

Для каждого исторического события X:

```text
удалить X из прошлого snapshot
↓
повторить множество прогонов
```

Если без X дальнейший класс структур почти никогда не возникает:

```text
X = evolutionary bottleneck / key innovation
```

---

# 135. Convergent Evolution Test

Особенно интересный эксперимент:

```text
одинаковый physics
+
разные seeds
```

Возникают ли независимо:

```text
membranes?
parasites?
communication?
multicellularity?
symbol systems?
```

Если да, это говорит о глубоких attractors пространства возможностей.

---

# 136. Contingency vs Necessity

Запускать тысячи историй.

Для каждого major transition считать:

```text
P(transition | physics)
```

Если событие появляется почти всегда:

```text
likely structural necessity
```

Если чрезвычайно редко:

```text
historical contingency
```

Это уже делает систему интересной как модель фундаментальных вопросов эволюции.

---

# 137. Search for Universal Evolutionary Patterns

Можно искать:

```text
повторяется ли паразитизм?
возникает ли кооперация?
нужна ли модульность?
возникает ли иерархия?
появляется ли разделение труда?
```

Если независимые цифровые миры регулярно приходят к одним и тем же абстрактным решениям, это может быть более интересным результатом, чем конкретный красивый организм.

---

# 138. Метрика сложности мира как причинной сети

Вместо количества сущностей:

```text
Complexity ≈ structure of causal dependency graph
```

Интересны:

```text
depth
modularity
feedback loops
cross-scale dependencies
reusable motifs
```

---

# 139. Multi-Scale Causal Graph

Хранить:

```text
micro causal graph
meso causal graph
macro causal graph
```

И смотреть:

```text
на каких масштабах появляются устойчивые причинные законы
```

Это потенциально связывает систему с causal emergence.

---

# 140. Когда абстракция становится "реальной"

Рабочее определение:

> макрообъект считается реальным уровнем симуляции, если его использование улучшает предсказание и управление по сравнению с чистым микроописанием.

То есть:

```text
predictive gain
+
compression gain
+
causal usefulness
```

---

# 141. Финальная форма внешнего селектора

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

Ни один отдельный слой не определяет "прогресс".

---

# 142. Возможный главный критерий OEE

Рабочая формулировка:

> система демонстрирует сильную open-ended evolution, если на длинных временных масштабах она продолжает производить новые устойчивые причинно-функциональные структуры, которые расширяют множество доступных последующих адаптаций и создают новые уровни организации.

В короткой форме:

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

# 143. Самый важный практический тест

Первый действительно убедительный эксперимент:

1. Запустить обычный мир без внешнего AI.
2. Измерить время до структурного saturation.
3. Запустить тот же мир с AI Possibility Expander.
4. Ограничить AI маленьким intervention budget.
5. Не давать AI конечной цели.
6. Сравнить:
   - novelty;
   - generativity;
   - ontology depth;
   - hierarchy depth;
   - niche creation;
   - time to saturation.

Если при небольшом числе нейтральных семантических расширений:

```text
AI-world
```

стабильно дольше создаёт новые функциональные категории, чем:

```text
control-world
```

это будет уже очень интересный результат.

---

# 144. Более сильный тест

После успешного предыдущего эксперимента:

AI разрешено добавлять primitive только тогда, когда:

```text
primitive
```

является компиляцией уже возникшего паттерна.

То есть AI запрещено "изобретать снаружи".

Схема:

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

Если и такой режим поддерживает OEE:

> направление расширения действительно идёт из мира, а AI лишь ускоряет переход между уровнями.

Это намного сильнее первоначальной архитектуры.

---

# 145. Ultimate Experiment

Самая амбициозная версия:

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

И дальше вопрос:

> насколько далеко система сможет построить собственный язык описания мира поверх исходных примитивов?

В идеале:

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

Последние уровни не должны быть заранее названы разработчиком.

Именно появление чего-то, для чего нам после эксперимента придётся придумать новое понятие, было бы самым сильным признаком настоящей open-ended evolution.


---

# 146. Как может выглядеть основа мира в коде

Главный принцип:

> базовый мир не должен знать, что такое организм, вид, хищник, пища, язык, культура или экономика.

Он должен знать только:

```text
пространство
материя
энергия
состояние
связи
память
исполняемые правила
```

То есть избегаем:

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

и вместо этого строим универсальные примитивы.

---

# 147. Базовая сущность

Для первой версии:

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

Ключевой элемент:

```go
Properties map[PropertyID]Value
```

Не нужно заранее добавлять поля:

```text
Health
Sex
Vision
Species
Intelligence
```

Позже новые механизмы могут добавить:

```text
electrical_charge
membrane_permeability
signal_frequency
chemical_affinity
symbol_memory
```

без изменения базового типа.

---

# 148. Динамические свойства

```go
type PropertyID uint32

type Value struct {
    Number float64
    Bytes  []byte
}
```

В более развитой версии можно перейти к tagged union:

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

`World` — только текущее состояние.

Он не содержит семантики вроде:

```text
population
species
ecosystem
society
```

Эти понятия должны появляться во внешнем Observer.

---

# 150. Пространственные поля

```go
type Fields struct {
    Energy Grid[float64]
    Matter Grid[float64]

    Dynamic map[FieldID]*ScalarField
}
```

Изначально:

```text
Energy
Matter
```

Позже можно динамически добавить:

```text
Light
Temperature
Chemical_A
Chemical_B
SignalField
Charge
```

---

# 151. Базовый Rule Interface

Kernel не должен знать смысл правил.

```go
type Rule interface {
    ID() RuleID

    Apply(
        ctx *RuleContext,
        out *EventBuffer,
    )
}
```

Примеры встроенных правил:

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

Основной принцип:

> Rule не должен напрямую изменять World.

Он генерирует события.

---

# 153. Event-driven изменение мира

```go
type Event interface {
    Apply(*World)
}
```

Примеры:

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

Получается:

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

Это даёт:

```text
replay
debugging
causal history
branching
counterfactual experiments
```

---

# 154. Исполняемая материя

Внутри сущности хранится программа.

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

Минимальная программа может выглядеть так:

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

Но операция `COPY` не должна означать:

```text
reproduce organism
```

Она должна означать только:

> скопировать структуру / программу при наличии ресурсов.

---

# 155. Репликацию не хардкодить

Избегать:

```go
func (e *Entity) Reproduce() *Entity
```

Лучше набор низкоуровневых операций:

```text
allocate matter
copy memory
copy instructions
transfer energy
create bond
detach
```

Тогда эволюция сама может создать алгоритм репликации:

```text
создать новую структуру
↓
скопировать код
↓
передать энергию
↓
отсоединить
```

Так эволюционировать сможет не только геном, но и сам механизм размножения.

---

# 156. Entity не обязан быть организмом

Ещё лучше рассматривать `Entity` как атомарный узел:

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

Тогда:

```text
1 Node
```

может быть бессмысленным,

а:

```text
100 connected Nodes
```

могут образовать:

```text
репликатор
мембрану
машину
колонию
```

Сам engine этого не знает.

---

# 157. Relations как first-class citizens

Связи нужно сделать полноценной частью мира.

```go
type Relation struct {
    ID RelationID

    From EntityID
    To   EntityID

    Kind PrimitiveID

    State []float64
}
```

Потому что OEE может развиваться через новые отношения:

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

Ключевой слой расширяемого мира:

```go
type Registry struct {
    Properties map[PropertyID]PropertyDefinition
    Fields     map[FieldID]FieldDefinition
    Primitives map[PrimitiveID]PrimitiveDefinition
    Opcodes    map[OpcodeID]OpcodeDefinition
}
```

Начальный Registry:

```text
energy
matter
position

move
bind
transfer
copy
```

Поздний Registry потенциально:

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

Это и есть практическая реализация:

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

Kernel знает только:

```text
как исполнять инструкции
как применять события
как распределять ресурсы
как создавать snapshot
как валидировать patch
как создавать новую ветвь
```

Он не знает:

```text
organism
species
culture
language
technology
```

---

# 160. AI Patch должен быть декларативным

Не давать AI писать arbitrary native code в kernel.

Лучше:

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

# 161. AI Patch не применяется сразу к основной ветви

```go
func TestPatch(
    snapshot Snapshot,
    patch Patch,
    seeds []int64,
) []ExperimentResult
```

Схема:

```text
             Snapshot W
             /    |    \
            /     |     \
       original   P1     P2
                           \
                            P3
```

AI создаёт новые ветки.

Мир не переписывается необратимо.

---

# 162. Минимальный Tick Loop

```go
func (w *World) Step(k *Kernel) {
    ctx := TickContext{
        Tick:     w.Tick,
        World:    w.ReadOnlyView(),
        Registry: &k.Registry,
    }

    events := NewEventBuffer()

    // 1. Исполняемый код сущностей.
    k.VM.ExecuteAll(ctx, events)

    // 2. Универсальные правила мира.
    for _, rule := range k.Registry.ActiveRules() {
        rule.Evaluate(ctx, events)
    }

    // 3. Разрешение конфликтов.
    resolved := ResolveEvents(events)

    // 4. Изменение состояния мира.
    ApplyEvents(w, resolved)

    // 5. Потери / диссипация.
    ApplyDissipation(w)

    w.Tick++
}
```

Цель:

> всё сложное должно расти поверх максимально маленького loop.

---

# 163. Observer вынести за пределы World

```go
type Observer struct {
    Metrics     MetricsEngine
    Novelty     NoveltyDetector
    Patterns    PatternDetector
    CausalModel CausalAnalyzer
}
```

Observer может говорить:

```text
"похоже, появилась популяция"
"похоже, появилась мембрана"
"похоже, возникла новая ниша"
```

Но эти понятия не существуют внутри физики мира.

---

# 164. World Summary для внешнего AI

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

Даже `CandidateMacroEntities` — только гипотеза Observer.

---

# 165. Чистое разделение системы

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

# 166. Самая первая версия кода

Для MVP можно ещё сильнее упростить:

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

И всего 8 инструкций:

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

Мир:

```text
256 × 256
```

Каждая инструкция стоит энергию.

Энергия поступает извне как физический градиент.

Если:

```text
Energy == 0
```

структура перестаёт выполняться / распадается.

---

# 167. Что НЕ писать

Особенно избегать:

```go
type Agent interface {
    Think()
    Act()
    Reproduce()
}
```

Потому что это заранее вводит:

```text
агентность
мышление
действие
размножение
```

Вместо этого:

```text
matter executes local transformations
```

Агентность должна быть интерпретацией устойчивого паттерна.

---

# 168. Минимальная структура Go-проекта

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

# 169. Следующий этап после базового kernel

После того как:

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

работают, следующий этап — НЕ подключать AI.

Следующий этап:

# получить автономную внутреннюю эволюцию без внешней помощи.

Главный вопрос:

> сможет ли минимальная цифровая физика поддерживать репликацию, мутации, конкуренцию и появление устойчивых линий без понятия Organism?

---

# 170. Этап A — Baseline Digital Evolution

## Цель

Получить замкнутый эксперимент:

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

# 171. Что реализовать на этапе A

## Мир

```text
256 × 256 grid
```

## Ресурсы

Минимум:

```text
EnergyField
Matter
```

## VM

Инструкции:

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

## Ограничения

Каждая операция имеет:

```text
energy cost
instruction cost
memory cost
```

---

# 172. Репликация на этапе A

Не использовать:

```text
Reproduce()
```

Нужно, чтобы программа сама выполнила:

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

Для первой версии допустимо иметь низкоуровневую:

```text
ALLOCATE
```

но не высокоуровневую:

```text
REPRODUCE
```

---

# 173. Мутации

Минимальный набор:

```text
instruction replacement
instruction insertion
instruction deletion
block duplication
block deletion
memory initialization mutation
```

Позже:

```text
recombination
horizontal transfer
```

---

# 174. Начальный эксперимент

Есть два варианта.

## Вариант 1 — Seed Replicator

В мир помещается один очень простой рабочий репликатор.

Цель:

```text
не проверять происхождение жизни,
а проверить эволюционную динамику.
```

Это лучший MVP.

## Вариант 2 — Abiogenesis Search

Начать со случайного executable matter и ждать появления репликации.

Это намного сложнее.

Для первой версии не рекомендуется.

---

# 175. Почему лучше начать с Seed Replicator

Если система не эволюционирует, нужно понимать:

```text
проблема в происхождении репликатора?
или
проблема в самой эволюционной архитектуре?
```

Seed Replicator разделяет эти две задачи.

Сначала проверить:

```text
replication → mutation → ecology
```

А происхождение репликации исследовать отдельно позже.

---

# 176. Критерии успеха этапа A

Нужно получить:

1. длительно существующую популяцию;
2. несколько lineage;
3. наследуемые различия;
4. изменение частот lineage со временем;
5. появление новых устойчивых программ;
6. отсутствие необходимости вручную назначать fitness.

---

# 177. Первый очень интересный результат

Проверять:

> возникнет ли паразитизм?

Например мутант:

```text
не копирует весь replication machinery
```

а использует:

```text
ресурсы / copy machinery соседей.
```

Если такая стратегия возникает сама:

```text
replicator
↓
parasite
↓
host defense
↓
parasite adaptation
```

то это уже сильный сигнал, что базовая экология работает.

---

# 178. Второй интересный результат

Проверять появление:

```text
cooperation
```

Например одна линия:

```text
собирает энергию
```

а другая:

```text
эффективно реплицируется
```

и между ними возникает устойчивый обмен.

Важно:

```text
не писать CooperationRule
```

Достаточно возможности:

```text
transfer_energy
```

---

# 179. Что измерять на этапе A

Минимальная telemetry:

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

Одного genome hash недостаточно.

Два разных генома могут делать одно и то же.

Для каждого lineage считать приблизительный:

```text
BehaviorSignature
```

Например:

```text
energy absorbed
distance moved
energy transferred
copies created
relations created
signals emitted
```

Это даст первую behavioral novelty.

---

# 181. Stage A Saturation

Запустить много длинных прогонов и выяснить:

```text
через сколько времени novelty перестаёт расти?
```

Это создаёт baseline:

```text
T_saturation_control
```

Позже именно с ним будет сравниваться AI-assisted OEE.

---

# 182. Следующий этап после A

Только если baseline evolution действительно работает:

# Этап B — Emergent Ecology

Добавить:

```text
несколько ресурсов
spatial heterogeneity
relations
resource transformation
waste products
local fields
```

Но всё ещё без внешнего AI.

Цель:

> добиться того, чтобы организмы сами создавали друг другу новые affordances.

---

# 183. Пример Emergent Ecology

```text
Lineage A:
Resource X → Waste Y

Lineage B:
Waste Y → Energy

↓
A создаёт нишу для B
```

Дальше:

```text
B меняет концентрацию Y
↓
это влияет на A
↓
возникает коэволюция
```

Никакой "задачи" здесь нет.

---

# 184. Критерий завершения этапа B

Хотя бы один устойчивый случай:

```text
одна линия изменяет среду
↓
это создаёт новый affordance
↓
другая линия его использует
↓
возникает длительная взаимозависимость
```

---

# 185. Только затем — этап C: Observer

После появления внутренней экологии подключить Observer:

```text
pattern detection
novelty
stagnation
affordances
lineages
candidate macro-entities
```

Observer пока:

```text
read-only
```

Он ничего не меняет.

---

# 186. Этап D — Branching Experiments

После Observer добавить:

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

Это подготовка к внешнему AI.

---

# 187. Этап E — AI Possibility Expander

Только здесь подключать внешний AI.

Ему разрешено:

```text
propose new primitive
propose new relation
propose new resource transformation
propose new local field
```

Но не:

```text
create species
create predator defense
create intelligence
```

---

# 188. Первый AI experiment

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

Важно:

AI не говорит:

```text
"используйте сигнал для кооперации"
```

Он просто добавляет возможность.

Дальше эволюция сама решает, что с ней делать.

---

# 189. Очень сильный первый результат

Например AI добавил:

```text
SEND_SIGNAL
```

в расчёте на расширение коммуникации.

А эволюция использовала его как:

```text
ложный сигнал
↓
приманку
↓
паразитическую стратегию
```

Такой результат интереснее, чем ожидаемое использование.

Он показывает:

> мир действительно исследует новое пространство сам.

---

# 190. Рекомендуемый порядок следующих этапов

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

# 191. Ближайший практический milestone

Самая ближайшая цель после проектирования кода:

> запустить мир, в котором один простой seed-replicator самостоятельно размножается, мутирует и образует несколько конкурирующих lineage.

Не нужно пока:

```text
AI
OEE evaluator
causal emergence
symbols
culture
```

Нужно доказать базовую предпосылку:

```text
минимальная физика
+
исполняемая материя
+
ограниченные ресурсы
```

уже создают нормальный дарвиновский процесс.

Только после этого имеет смысл расширять систему к open-ended evolution.

