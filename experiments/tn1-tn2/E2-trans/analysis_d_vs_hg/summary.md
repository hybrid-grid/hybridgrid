# Experiment analysis

Complete blocks: 20.
Time-to-adapt: W=20, reference=pre_all share, threshold=0.5 x pre_all share; reference share < 0.05 is NA. Censored values display as > n and enter paired tests as n_post + 1.
target_after_recovery_* is defined only for exactly one drift_on and one non-cleanup drift_off (transient); NA for onoff and permanent by design.
On/off aggregate shares pool decisions inside each phase after first drift_on; cell_end cleanup is excluded. Counts are decision and target-decision totals.
Holm family: every treatment x baseline pair for one metric in this directory; gamma/lambda sweep arms enlarge the family.

## Phases crossing build boundaries

| Round | Arm | Phases |
|---:|---|---|
| 1 | hybrid-linucb | 2_off |
| 1 | hybrid-linucb-d-arm-g098 | 2_off |
| 1 | hybrid-linucb-d-g095 | 2_off |
| 1 | hybrid-linucb-d-g098 | 2_off |
| 1 | hybrid-linucb-d-g099 | 2_off |
| 1 | leastloaded | 2_off |
| 1 | sed | 2_off |
| 2 | hybrid-linucb | 2_off |
| 2 | hybrid-linucb-d-arm-g098 | 2_off |
| 2 | hybrid-linucb-d-g095 | 2_off |
| 2 | hybrid-linucb-d-g098 | 2_off |
| 2 | hybrid-linucb-d-g099 | 2_off |
| 2 | leastloaded | 2_off |
| 2 | sed | 2_off |
| 3 | hybrid-linucb | 2_off |
| 3 | hybrid-linucb-d-arm-g098 | 2_off |
| 3 | hybrid-linucb-d-g095 | 2_off |
| 3 | hybrid-linucb-d-g098 | 2_off |
| 3 | hybrid-linucb-d-g099 | 2_off |
| 3 | leastloaded | 2_off |
| 3 | sed | 2_off |
| 4 | hybrid-linucb | 2_off |
| 4 | hybrid-linucb-d-arm-g098 | 2_off |
| 4 | hybrid-linucb-d-g095 | 2_off |
| 4 | hybrid-linucb-d-g098 | 2_off |
| 4 | hybrid-linucb-d-g099 | 2_off |
| 4 | leastloaded | 2_off |
| 4 | sed | 2_off |
| 5 | hybrid-linucb | 2_off |
| 5 | hybrid-linucb-d-arm-g098 | 2_off |
| 5 | hybrid-linucb-d-g095 | 2_off |
| 5 | hybrid-linucb-d-g098 | 2_off |
| 5 | hybrid-linucb-d-g099 | 2_off |
| 5 | leastloaded | 2_off |
| 5 | sed | 2_off |
| 6 | hybrid-linucb | 2_off |
| 6 | hybrid-linucb-d-arm-g098 | 2_off |
| 6 | hybrid-linucb-d-g095 | 2_off |
| 6 | hybrid-linucb-d-g098 | 2_off |
| 6 | hybrid-linucb-d-g099 | 2_off |
| 6 | leastloaded | 2_off |
| 6 | sed | 2_off |
| 7 | hybrid-linucb | 2_off |
| 7 | hybrid-linucb-d-arm-g098 | 2_off |
| 7 | hybrid-linucb-d-g095 | 2_off |
| 7 | hybrid-linucb-d-g098 | 2_off |
| 7 | hybrid-linucb-d-g099 | 2_off |
| 7 | leastloaded | 2_off |
| 7 | sed | 2_off |
| 8 | hybrid-linucb | 2_off |
| 8 | hybrid-linucb-d-arm-g098 | 2_off |
| 8 | hybrid-linucb-d-g095 | 2_off |
| 8 | hybrid-linucb-d-g098 | 2_off |
| 8 | hybrid-linucb-d-g099 | 2_off |
| 8 | leastloaded | 2_off |
| 8 | sed | 2_off |
| 9 | hybrid-linucb | 2_off |
| 9 | hybrid-linucb-d-arm-g098 | 2_off |
| 9 | hybrid-linucb-d-g095 | 2_off |
| 9 | hybrid-linucb-d-g098 | 2_off |
| 9 | hybrid-linucb-d-g099 | 2_off |
| 9 | leastloaded | 2_off |
| 9 | sed | 2_off |
| 10 | hybrid-linucb | 2_off |
| 10 | hybrid-linucb-d-arm-g098 | 2_off |
| 10 | hybrid-linucb-d-g095 | 2_off |
| 10 | hybrid-linucb-d-g098 | 2_off |
| 10 | hybrid-linucb-d-g099 | 2_off |
| 10 | leastloaded | 2_off |
| 10 | sed | 2_off |
| 11 | hybrid-linucb | 2_off |
| 11 | hybrid-linucb-d-arm-g098 | 2_off |
| 11 | hybrid-linucb-d-g095 | 2_off |
| 11 | hybrid-linucb-d-g098 | 2_off |
| 11 | hybrid-linucb-d-g099 | 2_off |
| 11 | leastloaded | 2_off |
| 11 | sed | 2_off |
| 12 | hybrid-linucb | 2_off |
| 12 | hybrid-linucb-d-arm-g098 | 2_off |
| 12 | hybrid-linucb-d-g095 | 2_off |
| 12 | hybrid-linucb-d-g098 | 2_off |
| 12 | hybrid-linucb-d-g099 | 2_off |
| 12 | leastloaded | 2_off |
| 12 | sed | 2_off |
| 13 | hybrid-linucb | 2_off |
| 13 | hybrid-linucb-d-arm-g098 | 2_off |
| 13 | hybrid-linucb-d-g095 | 2_off |
| 13 | hybrid-linucb-d-g098 | 2_off |
| 13 | hybrid-linucb-d-g099 | 2_off |
| 13 | leastloaded | 2_off |
| 13 | sed | 2_off |
| 14 | hybrid-linucb | 2_off |
| 14 | hybrid-linucb-d-arm-g098 | 2_off |
| 14 | hybrid-linucb-d-g095 | 2_off |
| 14 | hybrid-linucb-d-g098 | 2_off |
| 14 | hybrid-linucb-d-g099 | 2_off |
| 14 | leastloaded | 2_off |
| 14 | sed | 2_off |
| 15 | hybrid-linucb | 2_off |
| 15 | hybrid-linucb-d-arm-g098 | 2_off |
| 15 | hybrid-linucb-d-g095 | 2_off |
| 15 | hybrid-linucb-d-g098 | 2_off |
| 15 | hybrid-linucb-d-g099 | 2_off |
| 15 | leastloaded | 2_off |
| 15 | sed | 2_off |
| 16 | hybrid-linucb | 2_off |
| 16 | hybrid-linucb-d-arm-g098 | 2_off |
| 16 | hybrid-linucb-d-g095 | 2_off |
| 16 | hybrid-linucb-d-g098 | 2_off |
| 16 | hybrid-linucb-d-g099 | 2_off |
| 16 | leastloaded | 2_off |
| 16 | sed | 2_off |
| 17 | hybrid-linucb | 2_off |
| 17 | hybrid-linucb-d-arm-g098 | 2_off |
| 17 | hybrid-linucb-d-g095 | 2_off |
| 17 | hybrid-linucb-d-g098 | 2_off |
| 17 | hybrid-linucb-d-g099 | 2_off |
| 17 | leastloaded | 2_off |
| 17 | sed | 2_off |
| 18 | hybrid-linucb | 2_off |
| 18 | hybrid-linucb-d-arm-g098 | 2_off |
| 18 | hybrid-linucb-d-g095 | 2_off |
| 18 | hybrid-linucb-d-g098 | 2_off |
| 18 | hybrid-linucb-d-g099 | 2_off |
| 18 | leastloaded | 2_off |
| 18 | sed | 2_off |
| 19 | hybrid-linucb | 2_off |
| 19 | hybrid-linucb-d-arm-g098 | 2_off |
| 19 | hybrid-linucb-d-g095 | 2_off |
| 19 | hybrid-linucb-d-g098 | 2_off |
| 19 | hybrid-linucb-d-g099 | 2_off |
| 19 | leastloaded | 2_off |
| 19 | sed | 2_off |
| 20 | hybrid-linucb | 2_off |
| 20 | hybrid-linucb-d-arm-g098 | 2_off |
| 20 | hybrid-linucb-d-g095 | 2_off |
| 20 | hybrid-linucb-d-g098 | 2_off |
| 20 | hybrid-linucb-d-g099 | 2_off |
| 20 | leastloaded | 2_off |
| 20 | sed | 2_off |

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 174.2 | 20 |
| hybrid-linucb-d-arm-g098 | 175.2 | 20 |
| hybrid-linucb-d-g095 | 181 | 20 |
| hybrid-linucb-d-g098 | 178.9 | 20 |
| hybrid-linucb-d-g099 | 175.9 | 20 |
| leastloaded | 179 | 20 |
| sed | 175 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.444 [-0.759, 1.61] | 0.8408 | 0.8408 | 0.05714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 5.876 [3.999, 7.502] | 0.0001049 | 0.0004196 | 0.8952 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 2.779 [0.9115, 5.016] | 0.002325 | 0.006975 | 0.7429 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 1.439 [-0.041, 2.641] | 0.33 | 0.66 | 0.2571 | 0 |

## build_1_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 89.88 | 20 |
| hybrid-linucb-d-arm-g098 | 90.36 | 20 |
| hybrid-linucb-d-g095 | 92.77 | 20 |
| hybrid-linucb-d-g098 | 90.61 | 20 |
| hybrid-linucb-d-g099 | 91.1 | 20 |
| leastloaded | 91.5 | 20 |
| sed | 89.68 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.0485 [-1.729, 1.05] | 0.7012 | 0.9354 | -0.1048 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 1.843 [1.129, 3.991] | 0.002712 | 0.01085 | 0.7333 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.5395 [-0.5135, 1.258] | 0.3683 | 0.9354 | 0.2381 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.875 [-0.365, 1.976] | 0.3118 | 0.9354 | 0.2667 | 0 |

## build_2_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 84.08 | 20 |
| hybrid-linucb-d-arm-g098 | 84.17 | 20 |
| hybrid-linucb-d-g095 | 87.49 | 20 |
| hybrid-linucb-d-g098 | 87.12 | 20 |
| hybrid-linucb-d-g099 | 84.85 | 20 |
| leastloaded | 86.74 | 20 |
| sed | 85.08 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.4135 [-0.0305, 0.912] | 0.2774 | 0.5547 | 0.2857 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 3.116 [1.806, 4.994] | 0.0002613 | 0.001045 | 0.8571 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 2.861 [1.264, 3.707] | 0.0005856 | 0.001757 | 0.819 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.455 [-0.33, 1.223] | 0.3118 | 0.5547 | 0.2667 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.00339 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.08729 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.00339 [0.002542, 0.005085] | 0.0001788 | 0.0007154 | 1 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.00711 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1885 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.00711 [0.005112, 0.01035] | 0.0001961 | 0.0007844 | 1 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.949 | 20 |
| hybrid-linucb-d-arm-g098 | 0.9532 | 20 |
| hybrid-linucb-d-g095 | 0.8663 | 20 |
| hybrid-linucb-d-g098 | 0.8981 | 20 |
| hybrid-linucb-d-g099 | 0.9237 | 20 |
| leastloaded | 0.9067 | 20 |
| sed | 0.8115 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.002787 [-0.01949, 0.01384] | 0.8983 | 0.8983 | 0.0381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.06631 [-0.0948, -0.05641] | 3.815e-06 | 1.526e-05 | -0.9905 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.04106 [-0.05783, -0.03421] | 3.815e-06 | 1.526e-05 | -0.9905 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.02738 [-0.04928, 0.003453] | 0.01362 | 0.02723 | -0.619 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## free_slot_skip_given_free

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2542 | 20 |
| hybrid-linucb-d-arm-g098 | 0.272 | 20 |
| hybrid-linucb-d-g095 | 0.2686 | 20 |
| hybrid-linucb-d-g098 | 0.2602 | 20 |
| hybrid-linucb-d-g099 | 0.2475 | 20 |
| leastloaded | 0.2449 | 20 |
| sed | 0.2449 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.02119 [-0.01441, 0.03983] | 0.2871 | 0.8614 | 0.2714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02712 [-0.00339, 0.04407] | 0.08933 | 0.3573 | 0.4333 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.005085 [-0.0322, 0.03644] | 0.9256 | 1 | -0.02381 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.001695 [-0.01441, 0.02373] | 0.8666 | 1 | 0.04286 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.4059 | 20 |
| hybrid-linucb-d-arm-g098 | 0.4085 | 20 |
| hybrid-linucb-d-g095 | 0.4924 | 20 |
| hybrid-linucb-d-g098 | 0.4593 | 20 |
| hybrid-linucb-d-g099 | 0.4331 | 20 |
| leastloaded | 0.4314 | 20 |
| sed | 0.4576 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -2.776e-17 [-0.01949, 0.0178] | 0.8813 | 0.8813 | -0.0381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.07373 [0.04322, 0.1076] | 0.0002354 | 0.0009418 | 0.9381 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.03475 [0.0178, 0.06017] | 0.0004792 | 0.001438 | 0.8905 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01949 [-0.01864, 0.04661] | 0.2043 | 0.4086 | 0.3238 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-arm-g098 | 1 | 20 |
| hybrid-linucb-d-g095 | 1 | 20 |
| hybrid-linucb-d-g098 | 1 | 20 |
| hybrid-linucb-d-g099 | 1 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_pre_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2731 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2374 | 20 |
| hybrid-linucb-d-g095 | 0.1643 | 20 |
| hybrid-linucb-d-g098 | 0.1583 | 20 |
| hybrid-linucb-d-g099 | 0.3333 | 20 |
| leastloaded | 0.2222 | 20 |
| sed | 0.1927 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.06009 [-0.2563, 0.06449] | 0.2396 | 0.4791 | -0.3 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.1154 [-0.3, 0.03448] | 0.04405 | 0.1762 | -0.5143 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.189 [-0.2196, 0.02032] | 0.06372 | 0.1912 | -0.4762 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.03261 [-0.192, 0.2125] | 0.9273 | 0.9273 | 0.02857 | 0 |

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2558 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2248 | 20 |
| hybrid-linucb-d-g095 | 0.2101 | 20 |
| hybrid-linucb-d-g098 | 0.2287 | 20 |
| hybrid-linucb-d-g099 | 0.2578 | 20 |
| leastloaded | 0.2201 | 20 |
| sed | 0.2442 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.04283 [-0.09419, 0.02236] | 0.1769 | 0.3537 | -0.3524 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.04091 [-0.1136, 0.0155] | 0.06731 | 0.2019 | -0.4667 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.05936 [-0.109, -0.01448] | 0.03654 | 0.1462 | -0.5333 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01294 [-0.114, 0.08417] | 0.7285 | 0.7285 | -0.09524 | 0 |

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2495 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2557 | 20 |
| hybrid-linucb-d-g095 | 0.2449 | 20 |
| hybrid-linucb-d-g098 | 0.25 | 20 |
| hybrid-linucb-d-g099 | 0.2397 | 20 |
| leastloaded | 0.2459 | 20 |
| sed | 0.2497 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.0283 [-0.01409, 0.04354] | 0.165 | 0.6598 | 0.3619 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.009462 [-0.02844, 0.03606] | 0.7651 | 1 | 0.07619 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.006506 [-0.03386, 0.02974] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.002477 [-0.02065, 0.01133] | 0.6407 | 1 | -0.119 | 0 |

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 29 | 20 |
| hybrid-linucb-d-arm-g098 | 29.5 | 20 |
| hybrid-linucb-d-g095 | 29 | 20 |
| hybrid-linucb-d-g098 | 29 | 20 |
| hybrid-linucb-d-g099 | 29 | 20 |
| leastloaded | 29 | 20 |
| sed | 29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 1 [0, 1] | 0.0534 | 0.2136 | 0.5221 | 4 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-0.5, 0] | 0.6803 | 1 | 0.1364 | 9 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-1, 0] | 0.6886 | 1 | -0.1282 | 8 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-0.5, 0.5] | 0.8734 | 1 | 0.05128 | 8 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 129 | 20 |
| hybrid-linucb-d-arm-g098 | 129.5 | 20 |
| hybrid-linucb-d-g095 | 129 | 20 |
| hybrid-linucb-d-g098 | 129 | 20 |
| hybrid-linucb-d-g099 | 129 | 20 |
| leastloaded | 129 | 20 |
| sed | 129 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 1 [0, 1] | 0.0534 | 0.2136 | 0.5221 | 4 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-0.5, 0] | 0.6803 | 1 | 0.1364 | 9 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-1, 0] | 0.6886 | 1 | -0.1282 | 8 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-0.5, 0.5] | 0.8734 | 1 | 0.05128 | 8 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 461 | 20 |
| hybrid-linucb-d-arm-g098 | 460.5 | 20 |
| hybrid-linucb-d-g095 | 461 | 20 |
| hybrid-linucb-d-g098 | 461 | 20 |
| hybrid-linucb-d-g099 | 461 | 20 |
| leastloaded | 461 | 20 |
| sed | 461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -1 [-1, 0] | 0.0534 | 0.2136 | -0.5221 | 4 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [0, 0.5] | 0.6803 | 1 | -0.1364 | 9 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 1] | 0.6886 | 1 | 0.1282 | 8 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-0.5, 0.5] | 0.8734 | 1 | -0.05128 | 8 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2 | 20 |
| hybrid-linucb-d-arm-g098 | 0.19 | 20 |
| hybrid-linucb-d-g095 | 0.18 | 20 |
| hybrid-linucb-d-g098 | 0.17 | 20 |
| hybrid-linucb-d-g099 | 0.21 | 20 |
| leastloaded | 0.17 | 20 |
| sed | 0.13 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-0.12, 0.06] | 0.4772 | 0.9545 | -0.1961 | 3 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.02 [-0.08, 0] | 0.2758 | 0.8555 | -0.2924 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.04 [-0.09, 0.02] | 0.2139 | 0.8555 | -0.3333 | 2 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01 [-0.09, 0.09] | 0.9679 | 0.9679 | 0.01053 | 1 |

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2464 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2664 | 20 |
| hybrid-linucb-d-g095 | 0.257 | 20 |
| hybrid-linucb-d-g098 | 0.2655 | 20 |
| hybrid-linucb-d-g099 | 0.2518 | 20 |
| leastloaded | 0.2527 | 20 |
| sed | 0.2579 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.02744 [-0.002178, 0.05831] | 0.114 | 0.4559 | 0.4095 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.0009095 [-0.02217, 0.04587] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.002742 [-0.03438, 0.04431] | 0.8408 | 1 | 0.05714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.001217 [-0.02802, 0.02134] | 0.7228 | 1 | -0.09048 | 0 |

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.002477 | 20 |
| hybrid-linucb-d-arm-g098 | -0.15 | 20 |
| hybrid-linucb-d-g095 | -0.1474 | 20 |
| hybrid-linucb-d-g098 | -0.1036 | 20 |
| hybrid-linucb-d-g099 | -0.05619 | 20 |
| leastloaded | -0.237 | 20 |
| sed | -0.2195 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.1328 [-0.2073, 0.03821] | 0.08969 | 0.3588 | -0.4381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.07429 [-0.1809, 0.07427] | 0.1231 | 0.3693 | -0.4 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.005468 [-0.1633, 0.04438] | 0.5217 | 1 | -0.1714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01542 [-0.1174, 0.1367] | 0.8695 | 1 | 0.04762 | 0 |

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.1921 | 20 |
| hybrid-linucb-d-arm-g098 | -0.1686 | 20 |
| hybrid-linucb-d-g095 | -0.1576 | 20 |
| hybrid-linucb-d-g098 | -0.1697 | 20 |
| hybrid-linucb-d-g099 | -0.1843 | 20 |
| leastloaded | -0.2283 | 20 |
| sed | -0.1993 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.04721 [0.00893, 0.06199] | 0.01718 | 0.05154 | 0.6 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.03702 [0.02409, 0.05584] | 0.0003223 | 0.001289 | 0.8476 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.02391 [-0.02102, 0.0433] | 0.1429 | 0.2858 | 0.381 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01401 [-0.007971, 0.03] | 0.1429 | 0.2858 | 0.381 | 0 |

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 27 | 20 |
| hybrid-linucb-d-arm-g098 | 26.5 | 20 |
| hybrid-linucb-d-g095 | 26 | 20 |
| hybrid-linucb-d-g098 | 27 | 20 |
| hybrid-linucb-d-g099 | 42.5 | 20 |
| leastloaded | 40.5 | 20 |
| sed | 20 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-20, 31] | 1 | 1 | 0.006536 | 3 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-11, 11.5] | 0.9246 | 1 | 0.02614 | 3 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-21.5, 4] | 0.6628 | 1 | -0.117 | 2 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 2.5 [-0.5, 15] | 0.7224 | 1 | 0.09804 | 3 |

## target_after_recovery_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.03464 [0.002259, 0.07685] | 0.05317 | 0.2127 | 0.4952 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01741 [-0.02881, 0.04291] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.002662 [-0.04581, 0.05638] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.001825 [-0.02666, 0.04139] | 0.8983 | 1 | 0.0381 | 0 |

## target_after_recovery_tasks

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 87 | 20 |
| hybrid-linucb-d-arm-g098 | 94.5 | 20 |
| hybrid-linucb-d-g095 | 87 | 20 |
| hybrid-linucb-d-g098 | 86.5 | 20 |
| hybrid-linucb-d-g099 | 84.5 | 20 |
| leastloaded | 90 | 20 |
| sed | 91.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 9.5 [-1, 21] | 0.1167 | 0.4669 | 0.4 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 5 [-10.5, 12.5] | 0.5882 | 1 | 0.1381 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -1.5 [-17, 16] | 0.9839 | 1 | -0.005263 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 2.5 [-8.5, 13.5] | 0.8812 | 1 | 0.0381 | 0 |

## drift_on_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2062 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1781 | 20 |
| hybrid-linucb-d-g095 | 0.1993 | 20 |
| hybrid-linucb-d-g098 | 0.2024 | 20 |
| hybrid-linucb-d-g099 | 0.185 | 20 |
| leastloaded | 0.1776 | 20 |
| sed | 0.1641 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.01734 [-0.04463, 0.01109] | 0.2455 | 0.9819 | -0.3048 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.03796 [-0.05057, 0.03815] | 0.5459 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01876 [-0.03606, 0.04045] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.02504 [-0.03698, 0.01922] | 0.2943 | 0.9819 | -0.2762 | 0 |

## drift_on_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 133.5 | 20 |
| hybrid-linucb-d-arm-g098 | 143.5 | 20 |
| hybrid-linucb-d-g095 | 139 | 20 |
| hybrid-linucb-d-g098 | 138.5 | 20 |
| hybrid-linucb-d-g099 | 137.5 | 20 |
| leastloaded | 134 | 20 |
| sed | 136.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 5.5 [0, 13] | 0.01129 | 0.04515 | 0.6993 | 3 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 5 [-2, 9] | 0.3042 | 0.9127 | 0.2619 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 2.5 [-6.5, 8] | 0.7793 | 0.9127 | 0.07143 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 1 [-1, 3] | 0.3425 | 0.9127 | 0.2474 | 1 |

## drift_on_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 28 | 20 |
| hybrid-linucb-d-arm-g098 | 26 | 20 |
| hybrid-linucb-d-g095 | 27 | 20 |
| hybrid-linucb-d-g098 | 27 | 20 |
| hybrid-linucb-d-g099 | 26.5 | 20 |
| leastloaded | 23.5 | 20 |
| sed | 22 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -1 [-5, 4.5] | 0.6144 | 1 | -0.1316 | 1 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -4 [-7, 5.5] | 0.562 | 1 | -0.1476 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-4, 4.5] | 1 | 1 | 0 | 2 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -3.5 [-6.5, 5] | 0.4778 | 1 | -0.181 | 0 |

## drift_off_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.03464 [0.002259, 0.07685] | 0.05317 | 0.2127 | 0.4952 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01741 [-0.02881, 0.04291] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.002662 [-0.04581, 0.05638] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.001825 [-0.02666, 0.04139] | 0.8983 | 1 | 0.0381 | 0 |

## drift_off_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 328 | 20 |
| hybrid-linucb-d-arm-g098 | 317 | 20 |
| hybrid-linucb-d-g095 | 321.5 | 20 |
| hybrid-linucb-d-g098 | 322.5 | 20 |
| hybrid-linucb-d-g099 | 323.5 | 20 |
| leastloaded | 327.5 | 20 |
| sed | 325 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -8 [-14, -0.5] | 0.0088 | 0.0352 | -0.6842 | 1 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -5.5 [-9.5, 1.5] | 0.3313 | 0.6625 | -0.2476 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -4 [-8, 7.5] | 0.808 | 0.808 | -0.0619 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -1.5 [-4, 0.5] | 0.2061 | 0.6182 | -0.3392 | 2 |

## drift_off_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 87 | 20 |
| hybrid-linucb-d-arm-g098 | 94.5 | 20 |
| hybrid-linucb-d-g095 | 87 | 20 |
| hybrid-linucb-d-g098 | 86.5 | 20 |
| hybrid-linucb-d-g099 | 84.5 | 20 |
| leastloaded | 90 | 20 |
| sed | 91.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 9.5 [-1, 21] | 0.1167 | 0.4669 | 0.4 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 5 [-10.5, 12.5] | 0.5882 | 1 | 0.1381 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -1.5 [-17, 16] | 0.9839 | 1 | -0.005263 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 2.5 [-8.5, 13.5] | 0.8812 | 1 | 0.0381 | 0 |

## drift_phase_1_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2062 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1781 | 20 |
| hybrid-linucb-d-g095 | 0.1993 | 20 |
| hybrid-linucb-d-g098 | 0.2024 | 20 |
| hybrid-linucb-d-g099 | 0.185 | 20 |
| leastloaded | 0.1776 | 20 |
| sed | 0.1641 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.01734 [-0.04463, 0.01109] | 0.2455 | 0.9819 | -0.3048 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.03796 [-0.05057, 0.03815] | 0.5459 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01876 [-0.03606, 0.04045] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.02504 [-0.03698, 0.01922] | 0.2943 | 0.9819 | -0.2762 | 0 |

## drift_phase_2_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.03464 [0.002259, 0.07685] | 0.05317 | 0.2127 | 0.4952 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01741 [-0.02881, 0.04291] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.002662 [-0.04581, 0.05638] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.001825 [-0.02666, 0.04139] | 0.8983 | 1 | 0.0381 | 0 |

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 1 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-4:50051 | 1000 |
| 1 | sed | worker-4:50051 | 1000 |
| 2 | hybrid-linucb | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 2 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 2 | leastloaded | worker-2:50051 | 600 |
| 2 | sed | worker-5:50051 | 1100 |
| 3 | hybrid-linucb | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 3 | leastloaded | worker-3:50051 | 800 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 4 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 4 | leastloaded | worker-1:50051 | 500 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 5 | leastloaded | worker-4:50051 | 1000 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 6 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 6 | leastloaded | worker-1:50051 | 500 |
| 6 | sed | worker-5:50051 | 1100 |
| 7 | hybrid-linucb | worker-1:50051 | 500 |
| 7 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 7 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-4:50051 | 1000 |
| 8 | hybrid-linucb | worker-1:50051 | 500 |
| 8 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 8 | leastloaded | worker-3:50051 | 800 |
| 8 | sed | worker-5:50051 | 1100 |
| 9 | hybrid-linucb | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 9 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 9 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 9 | leastloaded | worker-1:50051 | 500 |
| 9 | sed | worker-3:50051 | 800 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 10 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 10 | leastloaded | worker-1:50051 | 500 |
| 10 | sed | worker-5:50051 | 1100 |
| 11 | hybrid-linucb | worker-5:50051 | 1100 |
| 11 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 11 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 11 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 11 | leastloaded | worker-3:50051 | 800 |
| 11 | sed | worker-3:50051 | 800 |
| 12 | hybrid-linucb | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 12 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-4:50051 | 1000 |
| 13 | hybrid-linucb | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 13 | leastloaded | worker-2:50051 | 600 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 14 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 14 | leastloaded | worker-1:50051 | 500 |
| 14 | sed | worker-2:50051 | 600 |
| 15 | hybrid-linucb | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 15 | leastloaded | worker-3:50051 | 800 |
| 15 | sed | worker-2:50051 | 600 |
| 16 | hybrid-linucb | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 16 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 16 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 16 | leastloaded | worker-1:50051 | 500 |
| 16 | sed | worker-4:50051 | 1000 |
| 17 | hybrid-linucb | worker-2:50051 | 600 |
| 17 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 17 | leastloaded | worker-2:50051 | 600 |
| 17 | sed | worker-2:50051 | 600 |
| 18 | hybrid-linucb | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 18 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 18 | leastloaded | worker-5:50051 | 1100 |
| 18 | sed | worker-5:50051 | 1100 |
| 19 | hybrid-linucb | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 19 | leastloaded | worker-1:50051 | 500 |
| 19 | sed | worker-4:50051 | 1000 |
| 20 | hybrid-linucb | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 20 | leastloaded | worker-4:50051 | 1000 |
| 20 | sed | worker-5:50051 | 1100 |
