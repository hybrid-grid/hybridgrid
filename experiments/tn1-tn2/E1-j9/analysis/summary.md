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
| hybrid-linucb | 84.99 | 20 |
| hybrid-linucb-d-g098 | 84.85 | 20 |
| leastloaded | 84.53 | 20 |
| sed | 84.95 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.62 [-0.3495, 2.171] | 0.2162 | 0.8647 | 0.3238 | 0 |
| hybrid-linucb | sed | 20 | -0.04 [-0.651, 1.74] | 0.2943 | 0.8828 | 0.2762 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.298 [-0.482, 2.164] | 0.3683 | 0.8828 | 0.2381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0245 [-0.675, 1.325] | 0.5459 | 0.8828 | 0.1619 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.05085 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.05085 [-0.06102, -0.04237] | 8.733e-05 | 0.0003493 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-d-g098 | sed | 20 | -0.05085 [-0.06102, -0.04237] | 8.733e-05 | 0.0003493 | -1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.128 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.128 [-0.1434, -0.1011] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-d-g098 | sed | 20 | -0.128 [-0.1434, -0.1011] | 1.907e-06 | 7.629e-06 | -1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9357 | 20 |
| hybrid-linucb-d-g098 | 0.9439 | 20 |
| leastloaded | 0.9021 | 20 |
| sed | 0.872 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.02981 [0.01561, 0.04642] | 0.0003223 | 0.0006447 | 0.8476 | 0 |
| hybrid-linucb | sed | 20 | 0.06238 [0.04288, 0.0832] | 3.624e-05 | 0.0001087 | 0.9333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02995 [0.01632, 0.04376] | 0.00486 | 0.00486 | 0.6952 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.05345 [0.02913, 0.08649] | 9.537e-06 | 3.815e-05 | 0.9714 | 0 |

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
| hybrid-linucb | 0.339 | 20 |
| hybrid-linucb-d-g098 | 0.3424 | 20 |
| leastloaded | 0.2915 | 20 |
| sed | 0.6237 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.0339 [-0.01356, 0.08136] | 0.04988 | 0.04988 | 0.5 | 0 |
| hybrid-linucb | sed | 20 | -0.3 [-0.3203, -0.2678] | 8.845e-05 | 0.0003528 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.05424 [0.02373, 0.07627] | 0.003153 | 0.006306 | 0.7238 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.2932 [-0.3085, -0.2475] | 8.82e-05 | 0.0003528 | -1 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2966 | 20 |
| hybrid-linucb-d-g098 | 0.2864 | 20 |
| leastloaded | 0.2475 | 20 |
| sed | 0.4051 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.05254 [0.0322, 0.0678] | 0.0008992 | 0.001798 | 0.8684 | 1 |
| hybrid-linucb | sed | 20 | -0.1051 [-0.1373, -0.07966] | 8.845e-05 | 0.0002653 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.05085 [0.01525, 0.0661] | 0.002484 | 0.002484 | 0.7714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1254 [-0.1525, -0.08644] | 3.815e-06 | 1.526e-05 | -0.9905 | 0 |

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
| 1 | hybrid-linucb | worker-4:50051 | 1000 |
| 1 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-4:50051 | 1000 |
| 1 | sed | worker-5:50051 | 1100 |
| 2 | hybrid-linucb | worker-2:50051 | 600 |
| 2 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 2 | leastloaded | worker-3:50051 | 800 |
| 2 | sed | worker-3:50051 | 800 |
| 3 | hybrid-linucb | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | leastloaded | worker-4:50051 | 1000 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-3:50051 | 800 |
| 4 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 4 | leastloaded | worker-5:50051 | 1100 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-2:50051 | 600 |
| 5 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 5 | leastloaded | worker-5:50051 | 1100 |
| 5 | sed | worker-4:50051 | 1000 |
| 6 | hybrid-linucb | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 6 | leastloaded | worker-3:50051 | 800 |
| 6 | sed | worker-4:50051 | 1000 |
| 7 | hybrid-linucb | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 7 | leastloaded | worker-3:50051 | 800 |
| 7 | sed | worker-5:50051 | 1100 |
| 8 | hybrid-linucb | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 8 | leastloaded | worker-4:50051 | 1000 |
| 8 | sed | worker-5:50051 | 1100 |
| 9 | hybrid-linucb | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 9 | leastloaded | worker-5:50051 | 1100 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 10 | leastloaded | worker-1:50051 | 500 |
| 10 | sed | worker-4:50051 | 1000 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 11 | leastloaded | worker-2:50051 | 600 |
| 11 | sed | worker-4:50051 | 1000 |
| 12 | hybrid-linucb | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 12 | leastloaded | worker-2:50051 | 600 |
| 12 | sed | worker-5:50051 | 1100 |
| 13 | hybrid-linucb | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 13 | leastloaded | worker-3:50051 | 800 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-1:50051 | 500 |
| 14 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 14 | leastloaded | worker-3:50051 | 800 |
| 14 | sed | worker-4:50051 | 1000 |
| 15 | hybrid-linucb | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 15 | leastloaded | worker-5:50051 | 1100 |
| 15 | sed | worker-5:50051 | 1100 |
| 16 | hybrid-linucb | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 16 | leastloaded | worker-1:50051 | 500 |
| 16 | sed | worker-5:50051 | 1100 |
| 17 | hybrid-linucb | worker-2:50051 | 600 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | leastloaded | worker-4:50051 | 1000 |
| 17 | sed | worker-5:50051 | 1100 |
| 18 | hybrid-linucb | worker-5:50051 | 1100 |
| 18 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 18 | leastloaded | worker-2:50051 | 600 |
| 18 | sed | worker-4:50051 | 1000 |
| 19 | hybrid-linucb | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 19 | leastloaded | worker-5:50051 | 1100 |
| 19 | sed | worker-3:50051 | 800 |
| 20 | hybrid-linucb | worker-3:50051 | 800 |
| 20 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 20 | leastloaded | worker-4:50051 | 1000 |
| 20 | sed | worker-2:50051 | 600 |
