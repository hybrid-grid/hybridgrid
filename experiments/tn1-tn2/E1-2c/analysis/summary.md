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
| hybrid-linucb | leastloaded | 20 | 0.223 [-1.936, 0.996] | 0.8695 | 1 | -0.04762 | 0 |
| hybrid-linucb | sed | 20 | 1.017 [-1.484, 3.059] | 0.4304 | 1 | 0.2095 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.354 [-1.204, 2.316] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.4845 [-0.8675, 1.952] | 0.2943 | 1 | 0.2762 | 0 |

## client_1_elapsed_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 163 | 20 |
| hybrid-linucb-d-g098 | 163.9 | 20 |
| leastloaded | 163.2 | 20 |
| sed | 164.3 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.81 [-1.626, 2.472] | 0.7841 | 1 | 0.07619 | 0 |
| hybrid-linucb | sed | 20 | -0.8655 [-2.595, 0.7745] | 0.3884 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 1.207 [-0.3315, 2.263] | 0.1429 | 0.5716 | 0.381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.459 [-1.321, 0.408] | 0.498 | 1 | -0.181 | 0 |

## client_2_elapsed_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 163.1 | 20 |
| hybrid-linucb-d-g098 | 164.6 | 20 |
| leastloaded | 165.1 | 20 |
| sed | 163.4 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -1.684 [-3.808, 1.011] | 0.2455 | 0.9819 | -0.3048 | 0 |
| hybrid-linucb | sed | 20 | 0.8095 [-1.184, 2.557] | 0.5459 | 1 | 0.1619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.761 [-1.862, 1.915] | 0.7562 | 1 | -0.08571 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.2365 [-1.108, 2.345] | 0.4091 | 1 | 0.219 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0.005085 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.01017 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.05878 | 0.05878 | 1 | 16 |
| hybrid-linucb | sed | 20 | -0.008475 [-0.01102, -0.00678] | 8.439e-05 | 0.0003375 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.005085 [0.00339, 0.005085] | 0.0001174 | 0.0003521 | 1 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -0.005085 [-0.007627, -0.002542] | 0.0008233 | 0.001647 | -0.9216 | 3 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0.03116 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.06154 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.125 | 0.125 | 1 | 16 |
| hybrid-linucb | sed | 20 | -0.05572 [-0.06591, -0.04382] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.03116 [0.02326, 0.03573] | 0.0001314 | 0.0003943 | 1 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -0.02787 [-0.04298, -0.01588] | 0.0003223 | 0.0006447 | -0.8476 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.969 | 20 |
| hybrid-linucb-d-g098 | 0.9322 | 20 |
| leastloaded | 0.9604 | 20 |
| sed | 0.9385 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.007408 [-0.01172, 0.01181] | 0.7012 | 0.8609 | 0.1048 | 0 |
| hybrid-linucb | sed | 20 | 0.01783 [0.006303, 0.04324] | 0.001986 | 0.005957 | 0.7524 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.02473 [-0.04146, -0.01225] | 0.0001335 | 0.0005341 | -0.8857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.004142 [-0.01824, 0.004989] | 0.4304 | 0.8609 | -0.2095 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9576 | 20 |
| hybrid-linucb-d-g098 | 0.9601 | 20 |
| leastloaded | 0.9483 | 20 |
| sed | 0.9644 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.007607 [0.0008014, 0.0153] | 0.01808 | 0.05424 | 0.6105 | 1 |
| hybrid-linucb | sed | 20 | -0.006715 [-0.01191, -0.001754] | 0.01958 | 0.05424 | -0.6105 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.008482 [0.006068, 0.01096] | 9.537e-06 | 3.815e-05 | 0.9714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0008532 [-0.009331, 0.003282] | 0.3118 | 0.3118 | -0.2667 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.1551 | 20 |
| hybrid-linucb-d-g098 | 0.1458 | 20 |
| leastloaded | 0.1508 | 20 |
| sed | 0.1644 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.00339 [-0.01271, 0.01186] | 0.8812 | 1 | 0.0381 | 0 |
| hybrid-linucb | sed | 20 | -0.00339 [-0.02542, 0.02458] | 0.7171 | 1 | -0.09474 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.005932 [-0.02373, 0.01949] | 0.3957 | 1 | -0.2281 | 2 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01525 [-0.02966, 0.00678] | 0.3224 | 1 | -0.2524 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-g098 | 0.9983 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.8516 | 1 | 0.06667 | 11 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 0.6234 | 1 | -0.1778 | 11 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.001695 [-0.001695, 0] | 0.08415 | 0.2525 | -0.4857 | 6 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0008475 [-0.001695, 0] | 0.01868 | 0.07474 | -0.7179 | 8 |

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
