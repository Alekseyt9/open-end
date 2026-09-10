# Trial rule comparison

Control and proposal continue the same source snapshot with identical initial RNG state. Once states diverge, identical initial RNGs do not imply identical events. No automatic winner is selected and source worlds remain unchanged.

| World | Variant | Population | Effective diversity | Largest structure | Window copies | Status |
|---|---|---:|---:|---:|---:|---|
| w001 (ecology, seed 1) | control | 677 | 2.615 | 338 | 9030 | mixed |
| w001 (ecology, seed 1) | solar-y-recycle | 690 | 2.697 | 102 | 6536 | mixed |
| w002 (ecology, seed 2) | control | 668 | 9.906 | 19 | 6412 | mixed |
| w002 (ecology, seed 2) | solar-y-recycle | 553 | 4.388 | 20 | 2614 | mixed |
| w003 (ecology, seed 3) | control | 673 | 8.979 | 109 | 7811 | mixed |
| w003 (ecology, seed 3) | solar-y-recycle | 684 | 11.105 | 99 | 5521 | mixed |
| w004 (ecology, seed 4) | control | 671 | 9.620 | 17 | 9769 | mixed |
| w004 (ecology, seed 4) | solar-y-recycle | 683 | 9.359 | 16 | 5336 | mixed |
| w005 (ecology, seed 5) | control | 671 | 11.862 | 1 | 11135 | mixed |
| w005 (ecology, seed 5) | solar-y-recycle | 671 | 11.297 | 1 | 7744 | mixed |
| w006 (ecology, seed 6) | control | 675 | 8.543 | 1 | 11406 | mixed |
| w006 (ecology, seed 6) | solar-y-recycle | 671 | 4.975 | 1 | 7560 | mixed |
| w007 (ecology, seed 7) | control | 662 | 7.772 | 26 | 14560 | mixed |
| w007 (ecology, seed 7) | solar-y-recycle | 666 | 5.627 | 26 | 9257 | stagnating |
| w008 (ecology, seed 8) | control | 667 | 3.497 | 4 | 14639 | mixed |
| w008 (ecology, seed 8) | solar-y-recycle | 655 | 3.702 | 1 | 9923 | mixed |
| w009 (ecology-no-mutation, seed 1) | control | 581 | 1.000 | 1 | 15748 | stagnating |
| w009 (ecology-no-mutation, seed 1) | solar-y-recycle | 569 | 1.000 | 1 | 14365 | mixed |
| w010 (ecology-no-mutation, seed 2) | control | 586 | 1.000 | 1 | 15737 | mixed |
| w010 (ecology-no-mutation, seed 2) | solar-y-recycle | 575 | 1.000 | 1 | 14445 | stagnating |
| w011 (ecology-no-mutation, seed 3) | control | 566 | 1.000 | 1 | 15670 | stagnating |
| w011 (ecology-no-mutation, seed 3) | solar-y-recycle | 572 | 1.000 | 1 | 14516 | mixed |
| w012 (ecology-no-mutation, seed 4) | control | 586 | 1.000 | 1 | 15826 | stagnating |
| w012 (ecology-no-mutation, seed 4) | solar-y-recycle | 560 | 1.000 | 1 | 14414 | mixed |
| w013 (ecology-no-mutation, seed 5) | control | 561 | 1.000 | 1 | 15487 | stagnating |
| w013 (ecology-no-mutation, seed 5) | solar-y-recycle | 551 | 1.000 | 1 | 14252 | mixed |
| w014 (ecology-no-mutation, seed 6) | control | 562 | 1.000 | 1 | 15594 | mixed |
| w014 (ecology-no-mutation, seed 6) | solar-y-recycle | 570 | 1.000 | 1 | 14481 | stagnating |
| w015 (ecology-no-mutation, seed 7) | control | 580 | 1.000 | 1 | 15566 | mixed |
| w015 (ecology-no-mutation, seed 7) | solar-y-recycle | 568 | 1.000 | 1 | 14480 | stagnating |
| w016 (ecology-no-mutation, seed 8) | control | 571 | 1.000 | 1 | 15446 | mixed |
| w016 (ecology-no-mutation, seed 8) | solar-y-recycle | 566 | 1.000 | 1 | 14481 | stagnating |

Values refer to the last requested window; actual boundaries are stored in the original trial's `results.json`. Differences require interpretation, repetitions, and longer runs. Behavioral hashes from different physical rule sets are not directly ranked.
