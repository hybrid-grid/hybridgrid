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
| hybrid-linucb | leastloaded | 20 | -0.843 [-2.716, 0.38] | 0.06193 | 0.06193 | -0.4762 | 0 |
| hybrid-linucb | sed | 20 | -6.981 [-7.254, -6.118] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.907 [0.4785, 2.479] | 0.00486 | 0.00972 | 0.6952 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -3.347 [-4.53, -2.384] | 1.907e-06 | 7.629e-06 | -1 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.4797 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.4797 [-0.5068, -0.4559] | 8.845e-05 | 0.0003533 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.08326 | 0.1665 | 1 | 17 |
| hybrid-linucb-d-g098 | sed | 20 | -0.4797 [-0.5068, -0.4542] | 8.832e-05 | 0.0003533 | -1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.4797 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.4797 [-0.5068, -0.4559] | 8.845e-05 | 0.0003538 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.25 | 0.5 | 1 | 17 |
| hybrid-linucb-d-g098 | sed | 20 | -0.4797 [-0.5068, -0.4539] | 8.845e-05 | 0.0003538 | -1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9023 | 20 |
| hybrid-linucb-d-g098 | 0.8159 | 20 |
| leastloaded | 0.8416 | 20 |
| sed | 0.5203 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.05985 [0.02865, 0.07394] | 0.0003948 | 0.0007896 | 0.8381 | 0 |
| hybrid-linucb | sed | 20 | 0.3841 [0.3319, 0.4181] | 1.907e-06 | 7.629e-06 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.03397 [-0.03901, 0.005797] | 0.03623 | 0.03623 | -0.5333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.3127 [0.2765, 0.3375] | 1.907e-06 | 7.629e-06 | 1 | 0 |

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
| hybrid-linucb | 0.3034 | 20 |
| hybrid-linucb-d-g098 | 0.278 | 20 |
| leastloaded | 0.2458 | 20 |
| sed | 0.3051 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03898 [0.02203, 0.06102] | 0.005107 | 0.02043 | 0.7143 | 0 |
| hybrid-linucb | sed | 20 | -0.02203 [-0.04576, 0.02542] | 0.2871 | 0.3119 | -0.2714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02203 [-0.008475, 0.03898] | 0.1559 | 0.3119 | 0.3619 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.04068 [-0.08983, -0.001695] | 0.009979 | 0.02994 | -0.6737 | 1 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.7712 | 20 |
| hybrid-linucb-d-g098 | 0.7932 | 20 |
| leastloaded | 0.7932 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.02881 [-0.06441, 0.00678] | 0.1453 | 0.2907 | -0.3714 | 0 |
| hybrid-linucb | sed | 20 | -0.2288 [-0.2322, -0.2085] | 8.72e-05 | 0.0003488 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.01186 [-0.0322, 0.03729] | 0.8813 | 0.8813 | 0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.2068 [-0.2169, -0.1678] | 8.82e-05 | 0.0003488 | -1 | 0 |

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
