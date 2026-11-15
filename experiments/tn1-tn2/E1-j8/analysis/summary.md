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
| hybrid-linucb | 86.39 | 20 |
| hybrid-linucb-d-g098 | 87.47 | 20 |
| leastloaded | 87.4 | 20 |
| sed | 85.71 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.842 [-2.986, -0.2525] | 0.03277 | 0.0983 | -0.5429 | 0 |
| hybrid-linucb | sed | 20 | 0.572 [-0.117, 1.544] | 0.07585 | 0.1517 | 0.4571 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.198 [-2.072, 1.584] | 0.5958 | 0.5958 | -0.1429 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 1.426 [0.633, 3.174] | 0.004221 | 0.01688 | 0.7048 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.08136 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.08136 [-0.09153, -0.0661] | 8.807e-05 | 0.0003523 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.1573 | 0.3146 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.08136 [-0.09153, -0.0661] | 8.807e-05 | 0.0003523 | -1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1861 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.1861 [-0.2144, -0.1581] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.5 | 1 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1861 [-0.2144, -0.1581] | 1.907e-06 | 7.629e-06 | -1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9383 | 20 |
| hybrid-linucb-d-g098 | 0.9071 | 20 |
| leastloaded | 0.9137 | 20 |
| sed | 0.8139 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03078 [0.0122, 0.04691] | 0.00169 | 0.00338 | 0.7619 | 0 |
| hybrid-linucb | sed | 20 | 0.1195 [0.08068, 0.154] | 1.907e-06 | 7.629e-06 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.005374 [-0.02643, 0.01958] | 0.5958 | 0.5958 | -0.1429 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.08388 [0.03373, 0.1088] | 1.907e-06 | 7.629e-06 | 1 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## free_slot_skip_given_free

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3017 | 20 |
| hybrid-linucb-d-g098 | 0.2695 | 20 |
| leastloaded | 0.2932 | 20 |
| sed | 0.2729 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03051 [-0.02203, 0.08136] | 0.1841 | 0.7365 | 0.3474 | 1 |
| hybrid-linucb | sed | 20 | 0.03898 [-0.0322, 0.0661] | 0.1977 | 0.7365 | 0.3286 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01017 [-0.03559, 0.04746] | 0.5014 | 0.7733 | 0.1714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.008475 [-0.01525, 0.03898] | 0.3866 | 0.7733 | 0.2263 | 1 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.4 | 20 |
| hybrid-linucb-d-g098 | 0.4492 | 20 |
| leastloaded | 0.4339 | 20 |
| sed | 0.4695 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.02542 [-0.06441, 0.001695] | 0.05092 | 0.2037 | -0.5105 | 1 |
| hybrid-linucb | sed | 20 | -0.04915 [-0.08475, -0.01186] | 0.07613 | 0.2284 | -0.4524 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02203 [-0.04746, 0.03898] | 0.9108 | 0.9553 | 0.02857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.005085 [-0.02881, 0.02034] | 0.4777 | 0.9553 | -0.181 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-g098 | 1 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| 1 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 1 | leastloaded | worker-5:50051 | 1100 |
| 1 | sed | worker-4:50051 | 1000 |
| 2 | hybrid-linucb | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 2 | leastloaded | worker-1:50051 | 500 |
| 2 | sed | worker-5:50051 | 1100 |
| 3 | hybrid-linucb | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 3 | leastloaded | worker-5:50051 | 1100 |
| 3 | sed | worker-2:50051 | 600 |
| 4 | hybrid-linucb | worker-5:50051 | 1100 |
| 4 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 4 | leastloaded | worker-2:50051 | 600 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 5 | leastloaded | worker-1:50051 | 500 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-5:50051 | 1100 |
| 6 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 6 | leastloaded | worker-5:50051 | 1100 |
| 6 | sed | worker-5:50051 | 1100 |
| 7 | hybrid-linucb | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-4:50051 | 1000 |
| 8 | hybrid-linucb | worker-5:50051 | 1100 |
| 8 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 8 | leastloaded | worker-5:50051 | 1100 |
| 8 | sed | worker-4:50051 | 1000 |
| 9 | hybrid-linucb | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 9 | leastloaded | worker-4:50051 | 1000 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 10 | leastloaded | worker-3:50051 | 800 |
| 10 | sed | worker-4:50051 | 1000 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 11 | leastloaded | worker-2:50051 | 600 |
| 11 | sed | worker-5:50051 | 1100 |
| 12 | hybrid-linucb | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 12 | leastloaded | worker-4:50051 | 1000 |
| 12 | sed | worker-4:50051 | 1000 |
| 13 | hybrid-linucb | worker-5:50051 | 1100 |
| 13 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 13 | leastloaded | worker-4:50051 | 1000 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 14 | leastloaded | worker-4:50051 | 1000 |
| 14 | sed | worker-4:50051 | 1000 |
| 15 | hybrid-linucb | worker-5:50051 | 1100 |
| 15 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 15 | leastloaded | worker-5:50051 | 1100 |
| 15 | sed | worker-5:50051 | 1100 |
| 16 | hybrid-linucb | worker-1:50051 | 500 |
| 16 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 16 | leastloaded | worker-5:50051 | 1100 |
| 16 | sed | worker-3:50051 | 800 |
| 17 | hybrid-linucb | worker-3:50051 | 800 |
| 17 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 17 | leastloaded | worker-1:50051 | 500 |
| 17 | sed | worker-5:50051 | 1100 |
| 18 | hybrid-linucb | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 18 | leastloaded | worker-3:50051 | 800 |
| 18 | sed | worker-2:50051 | 600 |
| 19 | hybrid-linucb | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 19 | leastloaded | worker-1:50051 | 500 |
| 19 | sed | worker-5:50051 | 1100 |
| 20 | hybrid-linucb | worker-4:50051 | 1000 |
| 20 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 20 | leastloaded | worker-1:50051 | 500 |
| 20 | sed | worker-3:50051 | 800 |
