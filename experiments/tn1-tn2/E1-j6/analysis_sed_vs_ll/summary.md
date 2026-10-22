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
| hybrid-linucb | 89.63 | 20 |
| hybrid-linucb-d-g098 | 92.95 | 20 |
| leastloaded | 91.22 | 20 |
| sed | 96.34 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 5.379 [3.899, 7.237] | 1.907e-06 | 1.907e-06 | 1 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.4797 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.4797 [0.4559, 0.5068] | 8.845e-05 | 8.845e-05 | 1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.4797 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.4797 [0.4559, 0.5068] | 8.845e-05 | 8.845e-05 | 1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9023 | 20 |
| hybrid-linucb-d-g098 | 0.8159 | 20 |
| leastloaded | 0.8416 | 20 |
| sed | 0.5203 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -0.3353 [-0.3741, -0.3117] | 1.907e-06 | 1.907e-06 | -1 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## free_slot_skip_given_free

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3034 | 20 |
| hybrid-linucb-d-g098 | 0.278 | 20 |
| leastloaded | 0.2458 | 20 |
| sed | 0.3051 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.05593 [0.02034, 0.08983] | 0.0005471 | 0.0005471 | 0.881 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.7712 | 20 |
| hybrid-linucb-d-g098 | 0.7932 | 20 |
| leastloaded | 0.7932 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.2068 [0.161, 0.2136] | 8.782e-05 | 8.782e-05 | 1 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-g098 | 1 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| 1 | hybrid-linucb | worker-5:50051 | 1100 |
| 1 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-4:50051 | 1000 |
| 1 | sed | worker-5:50051 | 1100 |
| 2 | hybrid-linucb | worker-3:50051 | 800 |
| 2 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 2 | leastloaded | worker-4:50051 | 1000 |
| 2 | sed | worker-3:50051 | 800 |
| 3 | hybrid-linucb | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | leastloaded | worker-5:50051 | 1100 |
| 3 | sed | worker-5:50051 | 1100 |
| 4 | hybrid-linucb | worker-5:50051 | 1100 |
| 4 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 4 | leastloaded | worker-3:50051 | 800 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 5 | leastloaded | worker-1:50051 | 500 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 6 | leastloaded | worker-1:50051 | 500 |
| 6 | sed | worker-4:50051 | 1000 |
| 7 | hybrid-linucb | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-5:50051 | 1100 |
| 8 | hybrid-linucb | worker-5:50051 | 1100 |
| 8 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 8 | leastloaded | worker-5:50051 | 1100 |
| 8 | sed | worker-4:50051 | 1000 |
| 9 | hybrid-linucb | worker-4:50051 | 1000 |
| 9 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 9 | leastloaded | worker-3:50051 | 800 |
| 9 | sed | worker-4:50051 | 1000 |
| 10 | hybrid-linucb | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 10 | leastloaded | worker-4:50051 | 1000 |
| 10 | sed | worker-5:50051 | 1100 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 11 | leastloaded | worker-4:50051 | 1000 |
| 11 | sed | worker-5:50051 | 1100 |
| 12 | hybrid-linucb | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-5:50051 | 1100 |
| 13 | hybrid-linucb | worker-2:50051 | 600 |
| 13 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 13 | leastloaded | worker-2:50051 | 600 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 14 | leastloaded | worker-5:50051 | 1100 |
| 14 | sed | worker-5:50051 | 1100 |
| 15 | hybrid-linucb | worker-5:50051 | 1100 |
| 15 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 15 | leastloaded | worker-2:50051 | 600 |
| 15 | sed | worker-5:50051 | 1100 |
| 16 | hybrid-linucb | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 16 | leastloaded | worker-5:50051 | 1100 |
| 16 | sed | worker-4:50051 | 1000 |
| 17 | hybrid-linucb | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 17 | leastloaded | worker-5:50051 | 1100 |
| 17 | sed | worker-4:50051 | 1000 |
| 18 | hybrid-linucb | worker-4:50051 | 1000 |
| 18 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 18 | leastloaded | worker-5:50051 | 1100 |
| 18 | sed | worker-4:50051 | 1000 |
| 19 | hybrid-linucb | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 19 | leastloaded | worker-3:50051 | 800 |
| 19 | sed | worker-5:50051 | 1100 |
| 20 | hybrid-linucb | worker-4:50051 | 1000 |
| 20 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 20 | leastloaded | worker-1:50051 | 500 |
| 20 | sed | worker-4:50051 | 1000 |
