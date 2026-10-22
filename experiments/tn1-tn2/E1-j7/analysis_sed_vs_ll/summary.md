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
| hybrid-linucb | 87.51 | 20 |
| hybrid-linucb-d-g098 | 90.74 | 20 |
| leastloaded | 89.78 | 20 |
| sed | 86.44 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -3.181 [-5.203, -1.695] | 1.907e-05 | 1.907e-05 | -0.9524 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.09153 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.08983 [0.08136, 0.1068] | 8.733e-05 | 8.733e-05 | 1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1432 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.1403 [0.124, 0.1602] | 1.907e-06 | 1.907e-06 | 1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9012 | 20 |
| hybrid-linucb-d-g098 | 0.8634 | 20 |
| leastloaded | 0.8758 | 20 |
| sed | 0.8568 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -0.03167 [-0.05626, -0.002278] | 0.08255 | 0.08255 | -0.4476 | 0 |

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
| hybrid-linucb | 0.2932 | 20 |
| hybrid-linucb-d-g098 | 0.2678 | 20 |
| leastloaded | 0.2559 | 20 |
| sed | 0.261 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | -0.001695 [-0.04237, 0.02034] | 0.8082 | 0.8082 | -0.0619 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.578 | 20 |
| hybrid-linucb-d-g098 | 0.6085 | 20 |
| leastloaded | 0.6068 | 20 |
| sed | 0.6542 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| sed | leastloaded | 20 | 0.04576 [0.01864, 0.08136] | 0.03654 | 0.03654 | 0.5333 | 0 |

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
| 1 | hybrid-linucb | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-2:50051 | 600 |
| 1 | sed | worker-3:50051 | 800 |
| 2 | hybrid-linucb | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 2 | leastloaded | worker-5:50051 | 1100 |
| 2 | sed | worker-5:50051 | 1100 |
| 3 | hybrid-linucb | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | leastloaded | worker-2:50051 | 600 |
| 3 | sed | worker-5:50051 | 1100 |
| 4 | hybrid-linucb | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 4 | leastloaded | worker-3:50051 | 800 |
| 4 | sed | worker-5:50051 | 1100 |
| 5 | hybrid-linucb | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 5 | leastloaded | worker-1:50051 | 500 |
| 5 | sed | worker-4:50051 | 1000 |
| 6 | hybrid-linucb | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 6 | leastloaded | worker-4:50051 | 1000 |
| 6 | sed | worker-5:50051 | 1100 |
| 7 | hybrid-linucb | worker-2:50051 | 600 |
| 7 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-5:50051 | 1100 |
| 8 | hybrid-linucb | worker-5:50051 | 1100 |
| 8 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 8 | leastloaded | worker-5:50051 | 1100 |
| 8 | sed | worker-3:50051 | 800 |
| 9 | hybrid-linucb | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 9 | leastloaded | worker-4:50051 | 1000 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 10 | leastloaded | worker-1:50051 | 500 |
| 10 | sed | worker-4:50051 | 1000 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 11 | leastloaded | worker-5:50051 | 1100 |
| 11 | sed | worker-5:50051 | 1100 |
| 12 | hybrid-linucb | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 12 | leastloaded | worker-2:50051 | 600 |
| 12 | sed | worker-5:50051 | 1100 |
| 13 | hybrid-linucb | worker-5:50051 | 1100 |
| 13 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 13 | leastloaded | worker-5:50051 | 1100 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 14 | leastloaded | worker-3:50051 | 800 |
| 14 | sed | worker-3:50051 | 800 |
| 15 | hybrid-linucb | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 15 | leastloaded | worker-1:50051 | 500 |
| 15 | sed | worker-5:50051 | 1100 |
| 16 | hybrid-linucb | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 16 | leastloaded | worker-1:50051 | 500 |
| 16 | sed | worker-5:50051 | 1100 |
| 17 | hybrid-linucb | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | leastloaded | worker-2:50051 | 600 |
| 17 | sed | worker-5:50051 | 1100 |
| 18 | hybrid-linucb | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 18 | leastloaded | worker-1:50051 | 500 |
| 18 | sed | worker-5:50051 | 1100 |
| 19 | hybrid-linucb | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 19 | leastloaded | worker-4:50051 | 1000 |
| 19 | sed | worker-5:50051 | 1100 |
| 20 | hybrid-linucb | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 20 | leastloaded | worker-3:50051 | 800 |
| 20 | sed | worker-4:50051 | 1000 |
