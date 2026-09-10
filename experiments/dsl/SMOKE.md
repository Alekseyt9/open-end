# DSL: смена правил и воспроизводимость

2026-09-10. Первый срез этапа 3: локальные реакции, ограниченный байткод, версии и откат без перезапуска ядра.

Проверен мир 32×32, seed 1, перенос материи и химии каждые 4 тика, стандартные мутации. Начальный модуль — `examples/rules/baseline.json`; в начале тика 500 загружается `examples/rules/direct-x.json`, в начале тика 1000 возвращается предыдущий модуль. Всего 2000 тиков.

```powershell
go run ./cmd/sim -width 32 -height 32 -ecology -matter-diffusion 4 -chemical-diffusion 4 -rules examples/rules/baseline.json -rule-change 500=examples/rules/direct-x.json -rollback-at 1000 -ticks 2000 -every 500 -save data/dsl-full.json
go run ./cmd/sim -width 32 -height 32 -ecology -matter-diffusion 4 -chemical-diffusion 4 -rules examples/rules/baseline.json -rule-change 500=examples/rules/direct-x.json -rollback-at 1000 -ticks 750 -save data/dsl-split.json
go run ./cmd/sim -load data/dsl-split.json -ticks 1250 -save data/dsl-resumed.json
```

Полный и возобновлённый запуск дали одинаковые байты snapshot и SHA-256:

```text
fae07fd94bd6a8b6418637227ada4fff6e16a06daffbbc72dd03bb898f4ac6a2
```

Итог: 564 частицы (561 с кодом), 9 геномов, 3641 копирование, 3082 смерти. В журнале три события, бюджетный счётчик DSL — 2 528 119 единиц работы. Все ресурсные инварианты соблюдены.

Версии модулей:

| Версия | SHA-256 канонического источника |
| --- | --- |
| baseline-v1 | aa34583ba4e890cfefd45f849c9df47f1ae1adad1d18ed4839da24c9fa05f020 |
| direct-x-v1 | 97836395dd7c42429e279c0ba7241377903a5c3a00f472bf9059c5337db71ead |

Дополнительный CLI-тест продолжает snapshot после удаления исходных файлов обоих модулей. Ядро тестируется на совпадение физической траектории базового DSL со встроенными реакциями, новый ID 2 и его цену, сохранение ресурсов на каждом тике, атомарный отказ установки и восстановление правил без сброса мира. Загрузка snapshot с изменённым байткодом отвергается.

Снимок этапа 2 на тике 100 000 загружен без DSL; прежний SHA-256 сохранён: `2d39a7db78c520febcd31858e91a12ee200e1a374e4df4aa2f0a342d799a7cde`.

Это проверка механизма исполнения и воспроизводимости, не долгосрочный эксперимент по адаптации к смене физики. Остальные операции и поля мира пока не описываются DSL; CLI подаёт изменения по заранее замороженному расписанию.
