# Experiment analysis

Complete blocks: 20.
Time-to-adapt: W=20, reference=pre_all share, threshold=0.5 x pre_all share; reference share < 0.05 is NA. Censored values display as > n and enter paired tests as n_post + 1.
target_after_recovery_* is defined only for exactly one drift_on and one non-cleanup drift_off (transient); NA for onoff and permanent by design.
On/off aggregate shares pool decisions inside each phase after first drift_on; cell_end cleanup is excluded. Counts are decision and target-decision totals.
Holm family: every treatment x baseline pair for one metric in this directory; gamma/lambda sweep arms enlarge the family.

## Availability notes

- Drift metrics unavailable for cells without dispatch sequence and drift events.

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 165.3 | 20 |
| hybrid-linucb-d-g098 | 165.6 | 20 |
| leastloaded | 165.2 | 20 |
| sed | 164.8 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.1345 [-1.71, 1.151] | 0.8124 | 0.8124 | -0.06667 | 0 |

## client_1_elapsed_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 163 | 20 |
| hybrid-linucb-d-g098 | 163.9 | 20 |
| leastloaded | 163.2 | 20 |
| sed | 164.3 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 1.485 [0.077, 2.591] | 0.08255 | 0.08255 | 0.4476 | 0 |

## client_2_elapsed_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 163.1 | 20 |
| hybrid-linucb-d-g098 | 164.6 | 20 |
| leastloaded | 165.1 | 20 |
| sed | 163.4 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -1.326 [-2.991, 0.0925] | 0.2455 | 0.2455 | -0.3048 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0.005085 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.01017 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.01017 [0.008475, 0.01186] | 8.426e-05 | 8.426e-05 | 1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0.03116 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.06154 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.06154 [0.04884, 0.06884] | 1.907e-06 | 1.907e-06 | 1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.969 | 20 |
| hybrid-linucb-d-g098 | 0.9322 | 20 |
| leastloaded | 0.9604 | 20 |
| sed | 0.9385 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -0.0212 [-0.03254, -0.005545] | 4.768e-05 | 4.768e-05 | -0.9238 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## free_slot_skip_given_free

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9576 | 20 |
| hybrid-linucb-d-g098 | 0.9601 | 20 |
| leastloaded | 0.9483 | 20 |
| sed | 0.9644 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.01188 [0.00678, 0.02113] | 0.001018 | 0.001018 | 0.8381 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.1551 | 20 |
| hybrid-linucb-d-g098 | 0.1458 | 20 |
| leastloaded | 0.1508 | 20 |
| sed | 0.1644 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.005085 [-0.01949, 0.02119] | 0.9108 | 0.9108 | 0.02857 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-g098 | 0.9983 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 0.5176 | 0.5176 | 0.2857 | 14 |

## drift_pre_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## target_after_recovery_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## target_after_recovery_tasks

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 1 | leastloaded | worker-1:50051 | 500 |
| 1 | sed | worker-3:50051 | 800 |
| 2 | hybrid-linucb | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 2 | leastloaded | worker-2:50051 | 600 |
| 2 | sed | worker-3:50051 | 800 |
| 3 | hybrid-linucb | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 3 | leastloaded | worker-3:50051 | 800 |
| 3 | sed | worker-5:50051 | 1100 |
| 4 | hybrid-linucb | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 4 | leastloaded | worker-3:50051 | 800 |
| 4 | sed | worker-5:50051 | 1100 |
| 5 | hybrid-linucb | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 5 | leastloaded | worker-2:50051 | 600 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 6 | leastloaded | worker-3:50051 | 800 |
| 6 | sed | worker-4:50051 | 1000 |
| 7 | hybrid-linucb | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-3:50051 | 800 |
| 8 | hybrid-linucb | worker-2:50051 | 600 |
| 8 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 8 | leastloaded | worker-3:50051 | 800 |
| 8 | sed | worker-3:50051 | 800 |
| 9 | hybrid-linucb | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 9 | leastloaded | worker-5:50051 | 1100 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 10 | leastloaded | worker-1:50051 | 500 |
| 10 | sed | worker-2:50051 | 600 |
| 11 | hybrid-linucb | worker-2:50051 | 600 |
| 11 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 11 | leastloaded | worker-4:50051 | 1000 |
| 11 | sed | worker-2:50051 | 600 |
| 12 | hybrid-linucb | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-5:50051 | 1100 |
| 13 | hybrid-linucb | worker-2:50051 | 600 |
| 13 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 13 | leastloaded | worker-2:50051 | 600 |
| 13 | sed | worker-4:50051 | 1000 |
| 14 | hybrid-linucb | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 14 | leastloaded | worker-1:50051 | 500 |
| 14 | sed | worker-5:50051 | 1100 |
| 15 | hybrid-linucb | worker-3:50051 | 800 |
| 15 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 15 | leastloaded | worker-3:50051 | 800 |
| 15 | sed | worker-4:50051 | 1000 |
| 16 | hybrid-linucb | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 16 | leastloaded | worker-2:50051 | 600 |
| 16 | sed | worker-5:50051 | 1100 |
| 17 | hybrid-linucb | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | leastloaded | worker-2:50051 | 600 |
| 17 | sed | worker-3:50051 | 800 |
| 18 | hybrid-linucb | worker-1:50051 | 500 |
| 18 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 18 | leastloaded | worker-1:50051 | 500 |
| 18 | sed | worker-5:50051 | 1100 |
| 19 | hybrid-linucb | worker-2:50051 | 600 |
| 19 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 19 | leastloaded | worker-2:50051 | 600 |
| 19 | sed | worker-1:50051 | 500 |
| 20 | hybrid-linucb | worker-2:50051 | 600 |
| 20 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 20 | leastloaded | worker-1:50051 | 500 |
| 20 | sed | worker-4:50051 | 1000 |
