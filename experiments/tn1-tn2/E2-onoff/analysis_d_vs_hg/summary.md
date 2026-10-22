# Experiment analysis

Complete blocks: 20.
Time-to-adapt: W=20, reference=pre_all share, threshold=0.5 x pre_all share; reference share < 0.05 is NA. Censored values display as > n and enter paired tests as n_post + 1.
target_after_recovery_* is defined only for exactly one drift_on and one non-cleanup drift_off (transient); NA for onoff and permanent by design.
On/off aggregate shares pool decisions inside each phase after first drift_on; cell_end cleanup is excluded. Counts are decision and target-decision totals.
Holm family: every treatment x baseline pair for one metric in this directory; gamma/lambda sweep arms enlarge the family.

## Phases crossing build boundaries

| Round | Arm | Phases |
|---:|---|---|
| 1 | hybrid-linucb | 4_off |
| 1 | hybrid-linucb-d-arm-g098 | 4_off |
| 1 | hybrid-linucb-d-g095 | 4_off |
| 1 | hybrid-linucb-d-g098 | 4_off |
| 1 | hybrid-linucb-d-g099 | 4_off |
| 1 | leastloaded | 4_off |
| 1 | sed | 4_off |
| 2 | hybrid-linucb | 4_off |
| 2 | hybrid-linucb-d-arm-g098 | 4_off |
| 2 | hybrid-linucb-d-g095 | 4_off |
| 2 | hybrid-linucb-d-g098 | 4_off |
| 2 | hybrid-linucb-d-g099 | 4_off |
| 2 | leastloaded | 4_off |
| 2 | sed | 4_off |
| 3 | hybrid-linucb | 4_off |
| 3 | hybrid-linucb-d-arm-g098 | 4_off |
| 3 | hybrid-linucb-d-g095 | 4_off |
| 3 | hybrid-linucb-d-g098 | 4_off |
| 3 | hybrid-linucb-d-g099 | 4_off |
| 3 | leastloaded | 4_off |
| 3 | sed | 4_off |
| 4 | hybrid-linucb | 4_off |
| 4 | hybrid-linucb-d-arm-g098 | 4_off |
| 4 | hybrid-linucb-d-g095 | 4_off |
| 4 | hybrid-linucb-d-g098 | 4_off |
| 4 | hybrid-linucb-d-g099 | 4_off |
| 4 | leastloaded | 4_off |
| 4 | sed | 4_off |
| 5 | hybrid-linucb | 4_off |
| 5 | hybrid-linucb-d-arm-g098 | 4_off |
| 5 | hybrid-linucb-d-g095 | 4_off |
| 5 | hybrid-linucb-d-g098 | 4_off |
| 5 | hybrid-linucb-d-g099 | 4_off |
| 5 | leastloaded | 4_off |
| 5 | sed | 4_off |
| 6 | hybrid-linucb | 4_off |
| 6 | hybrid-linucb-d-arm-g098 | 4_off |
| 6 | hybrid-linucb-d-g095 | 4_off |
| 6 | hybrid-linucb-d-g098 | 4_off |
| 6 | hybrid-linucb-d-g099 | 4_off |
| 6 | leastloaded | 4_off |
| 6 | sed | 4_off |
| 7 | hybrid-linucb | 4_off |
| 7 | hybrid-linucb-d-arm-g098 | 4_off |
| 7 | hybrid-linucb-d-g095 | 4_off |
| 7 | hybrid-linucb-d-g098 | 4_off |
| 7 | hybrid-linucb-d-g099 | 4_off |
| 7 | leastloaded | 4_off |
| 7 | sed | 4_off |
| 8 | hybrid-linucb | 4_off |
| 8 | hybrid-linucb-d-arm-g098 | 4_off |
| 8 | hybrid-linucb-d-g095 | 5_on |
| 8 | hybrid-linucb-d-g098 | 4_off |
| 8 | hybrid-linucb-d-g099 | 4_off |
| 8 | leastloaded | 4_off |
| 8 | sed | 4_off |
| 9 | hybrid-linucb | 4_off |
| 9 | hybrid-linucb-d-arm-g098 | 4_off |
| 9 | hybrid-linucb-d-g095 | 4_off |
| 9 | hybrid-linucb-d-g098 | 4_off |
| 9 | hybrid-linucb-d-g099 | 4_off |
| 9 | leastloaded | 4_off |
| 9 | sed | 4_off |
| 10 | hybrid-linucb | 4_off |
| 10 | hybrid-linucb-d-arm-g098 | 4_off |
| 10 | hybrid-linucb-d-g095 | 5_on |
| 10 | hybrid-linucb-d-g098 | 4_off |
| 10 | hybrid-linucb-d-g099 | 4_off |
| 10 | leastloaded | 4_off |
| 10 | sed | 4_off |
| 11 | hybrid-linucb | 4_off |
| 11 | hybrid-linucb-d-arm-g098 | 4_off |
| 11 | hybrid-linucb-d-g095 | 4_off |
| 11 | hybrid-linucb-d-g098 | 4_off |
| 11 | hybrid-linucb-d-g099 | 4_off |
| 11 | leastloaded | 5_on |
| 11 | sed | 4_off |
| 12 | hybrid-linucb | 4_off |
| 12 | hybrid-linucb-d-arm-g098 | 4_off |
| 12 | hybrid-linucb-d-g095 | 4_off |
| 12 | hybrid-linucb-d-g098 | 4_off |
| 12 | hybrid-linucb-d-g099 | 4_off |
| 12 | leastloaded | 4_off |
| 12 | sed | 4_off |
| 13 | hybrid-linucb | 4_off |
| 13 | hybrid-linucb-d-arm-g098 | 4_off |
| 13 | hybrid-linucb-d-g095 | 4_off |
| 13 | hybrid-linucb-d-g098 | 4_off |
| 13 | hybrid-linucb-d-g099 | 4_off |
| 13 | leastloaded | 4_off |
| 13 | sed | 4_off |
| 14 | hybrid-linucb | 5_on |
| 14 | hybrid-linucb-d-arm-g098 | 4_off |
| 14 | hybrid-linucb-d-g095 | 4_off |
| 14 | hybrid-linucb-d-g098 | 4_off |
| 14 | hybrid-linucb-d-g099 | 4_off |
| 14 | leastloaded | 4_off |
| 14 | sed | 4_off |
| 15 | hybrid-linucb | 4_off |
| 15 | hybrid-linucb-d-arm-g098 | 4_off |
| 15 | hybrid-linucb-d-g095 | 4_off |
| 15 | hybrid-linucb-d-g098 | 4_off |
| 15 | hybrid-linucb-d-g099 | 4_off |
| 15 | leastloaded | 5_on |
| 15 | sed | 4_off |
| 16 | hybrid-linucb | 4_off |
| 16 | hybrid-linucb-d-arm-g098 | 4_off |
| 16 | hybrid-linucb-d-g095 | 4_off |
| 16 | hybrid-linucb-d-g098 | 4_off |
| 16 | hybrid-linucb-d-g099 | 4_off |
| 16 | leastloaded | 4_off |
| 16 | sed | 4_off |
| 17 | hybrid-linucb | 4_off |
| 17 | hybrid-linucb-d-arm-g098 | 4_off |
| 17 | hybrid-linucb-d-g095 | 4_off |
| 17 | hybrid-linucb-d-g098 | 4_off |
| 17 | hybrid-linucb-d-g099 | 4_off |
| 17 | leastloaded | 4_off |
| 17 | sed | 4_off |
| 18 | hybrid-linucb | 4_off |
| 18 | hybrid-linucb-d-arm-g098 | 4_off |
| 18 | hybrid-linucb-d-g095 | 4_off |
| 18 | hybrid-linucb-d-g098 | 4_off |
| 18 | hybrid-linucb-d-g099 | 4_off |
| 18 | leastloaded | 4_off |
| 18 | sed | 4_off |
| 19 | hybrid-linucb | 4_off |
| 19 | hybrid-linucb-d-arm-g098 | 4_off |
| 19 | hybrid-linucb-d-g095 | 4_off |
| 19 | hybrid-linucb-d-g098 | 4_off |
| 19 | hybrid-linucb-d-g099 | 4_off |
| 19 | leastloaded | 4_off |
| 19 | sed | 4_off |
| 20 | hybrid-linucb | 4_off |
| 20 | hybrid-linucb-d-arm-g098 | 4_off |
| 20 | hybrid-linucb-d-g095 | 4_off |
| 20 | hybrid-linucb-d-g098 | 4_off |
| 20 | hybrid-linucb-d-g099 | 4_off |
| 20 | leastloaded | 4_off |
| 20 | sed | 4_off |

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 181.2 | 20 |
| hybrid-linucb-d-arm-g098 | 181.1 | 20 |
| hybrid-linucb-d-g095 | 186.7 | 20 |
| hybrid-linucb-d-g098 | 183.8 | 20 |
| hybrid-linucb-d-g099 | 182.1 | 20 |
| leastloaded | 184.6 | 20 |
| sed | 180.2 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.1565 [-1.917, 1.395] | 0.6742 | 0.7366 | -0.1143 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 4.993 [2.422, 6.507] | 8.202e-05 | 0.0003281 | 0.9048 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 1.597 [-0.401, 3.976] | 0.02664 | 0.07993 | 0.5619 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.2585 [-0.4875, 2.614] | 0.3683 | 0.7366 | 0.2381 | 0 |

## build_1_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 90.66 | 20 |
| hybrid-linucb-d-arm-g098 | 89.87 | 20 |
| hybrid-linucb-d-g095 | 91.2 | 20 |
| hybrid-linucb-d-g098 | 91.31 | 20 |
| hybrid-linucb-d-g099 | 90.24 | 20 |
| leastloaded | 91.16 | 20 |
| sed | 89.41 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.325 [-2.33, 0.5925] | 0.1893 | 0.568 | -0.3429 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.6155 [-0.038, 3.446] | 0.09731 | 0.3892 | 0.4286 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.5785 [-0.519, 1.647] | 0.2455 | 0.568 | 0.3048 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.301 [-1.927, 0.968] | 0.4524 | 0.568 | -0.2 | 0 |

## build_2_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 90.09 | 20 |
| hybrid-linucb-d-arm-g098 | 90.83 | 20 |
| hybrid-linucb-d-g095 | 94.85 | 20 |
| hybrid-linucb-d-g098 | 91.21 | 20 |
| hybrid-linucb-d-g099 | 91.26 | 20 |
| leastloaded | 92.82 | 20 |
| sed | 90.45 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.7705 [-0.3575, 1.582] | 0.3488 | 0.3488 | 0.2476 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 2.717 [1.712, 5.061] | 1.907e-06 | 7.629e-06 | 1 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.6405 [0.1065, 1.198] | 0.01069 | 0.03207 | 0.6381 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.738 [0.175, 1.489] | 0.01069 | 0.03207 | 0.6381 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.005085 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.07627 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.005085 [0.002542, 0.005932] | 0.0001817 | 0.000727 | 1 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 0.1573 | 0.4719 | 1 | 18 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.01014 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1648 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01014 [0.005371, 0.0122] | 7.629e-06 | 3.052e-05 | 1 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 0.5 | 1 | 1 | 18 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9432 | 20 |
| hybrid-linucb-d-arm-g098 | 0.9461 | 20 |
| hybrid-linucb-d-g095 | 0.8791 | 20 |
| hybrid-linucb-d-g098 | 0.9066 | 20 |
| hybrid-linucb-d-g099 | 0.9313 | 20 |
| leastloaded | 0.907 | 20 |
| sed | 0.8352 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.007336 [-0.00887, 0.014] | 0.3884 | 0.3884 | 0.2286 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.06117 [-0.07474, -0.05228] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.02789 [-0.03713, -0.02008] | 5.722e-06 | 1.717e-05 | -0.981 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.007012 [-0.02747, -0.002841] | 0.01362 | 0.02723 | -0.619 | 0 |

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
| hybrid-linucb | 0.2619 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2644 | 20 |
| hybrid-linucb-d-g095 | 0.272 | 20 |
| hybrid-linucb-d-g098 | 0.2653 | 20 |
| hybrid-linucb-d-g099 | 0.2458 | 20 |
| leastloaded | 0.2424 | 20 |
| sed | 0.2347 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.007627 [-0.03814, 0.01864] | 0.3836 | 1 | -0.2339 | 2 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01525 [-0.001695, 0.02797] | 0.1166 | 0.4665 | 0.4 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.00678 [-0.02119, 0.02373] | 0.9273 | 1 | 0.02857 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.005932 [-0.02458, 0.009322] | 0.3683 | 1 | -0.2381 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.4127 | 20 |
| hybrid-linucb-d-arm-g098 | 0.4085 | 20 |
| hybrid-linucb-d-g095 | 0.4771 | 20 |
| hybrid-linucb-d-g098 | 0.4407 | 20 |
| hybrid-linucb-d-g099 | 0.4407 | 20 |
| leastloaded | 0.4305 | 20 |
| sed | 0.4695 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.005085 [-0.01186, 0.01949] | 0.8227 | 0.8227 | 0.05714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.05085 [0.03559, 0.08729] | 0.0001394 | 0.0005575 | 0.9714 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.02542 [0.0008475, 0.04407] | 0.0122 | 0.03661 | 0.6725 | 2 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01695 [-0.002542, 0.05085] | 0.04988 | 0.09976 | 0.5 | 0 |

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
| hybrid-linucb | 0.2069 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2629 | 20 |
| hybrid-linucb-d-g095 | 0.2831 | 20 |
| hybrid-linucb-d-g098 | 0.3793 | 20 |
| hybrid-linucb-d-g099 | 0.2586 | 20 |
| leastloaded | 0.2707 | 20 |
| sed | 0.2457 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.03624 [-0.09789, 0.1868] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.0124 [-0.1266, 0.1684] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.1046 [0.04004, 0.237] | 0.05457 | 0.2183 | 0.5053 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.007143 [-0.1082, 0.1379] | 0.8721 | 1 | 0.04211 | 1 |

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2433 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2829 | 20 |
| hybrid-linucb-d-g095 | 0.2636 | 20 |
| hybrid-linucb-d-g098 | 0.2918 | 20 |
| hybrid-linucb-d-g099 | 0.2326 | 20 |
| leastloaded | 0.2529 | 20 |
| sed | 0.3011 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.03775 [-0.02612, 0.08264] | 0.2305 | 0.6915 | 0.3143 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02027 [-0.024, 0.05053] | 0.3927 | 0.7855 | 0.2398 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.06158 [-0.02409, 0.1197] | 0.165 | 0.6598 | 0.3619 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01455 [-0.02524, 0.06587] | 0.7285 | 0.7855 | 0.09524 | 0 |

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2538 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2397 | 20 |
| hybrid-linucb-d-g095 | 0.2345 | 20 |
| hybrid-linucb-d-g098 | 0.2538 | 20 |
| hybrid-linucb-d-g099 | 0.2519 | 20 |
| leastloaded | 0.2343 | 20 |
| sed | 0.2269 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.03157 [-0.04959, 0.002411] | 0.08255 | 0.2783 | -0.4476 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.02103 [-0.05129, 0.006225] | 0.06958 | 0.2783 | -0.4667 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01544 [-0.03913, 0.03065] | 0.5958 | 1 | -0.1429 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.006236 [-0.03074, 0.02898] | 0.7012 | 1 | -0.1048 | 0 |

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 29 | 20 |
| hybrid-linucb-d-arm-g098 | 29 | 20 |
| hybrid-linucb-d-g095 | 29 | 20 |
| hybrid-linucb-d-g098 | 29 | 20 |
| hybrid-linucb-d-g099 | 29 | 20 |
| leastloaded | 29 | 20 |
| sed | 29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.6281 | 1 | -0.1429 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-0.5, 1] | 0.4145 | 1 | 0.225 | 5 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.5922 | 1 | -0.15 | 5 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-1.5, 0.5] | 0.15 | 0.5998 | -0.4286 | 6 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 129 | 20 |
| hybrid-linucb-d-arm-g098 | 129 | 20 |
| hybrid-linucb-d-g095 | 129 | 20 |
| hybrid-linucb-d-g098 | 129 | 20 |
| hybrid-linucb-d-g099 | 129 | 20 |
| leastloaded | 129 | 20 |
| sed | 129 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.6281 | 1 | -0.1429 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-0.5, 1] | 0.4145 | 1 | 0.225 | 5 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.5922 | 1 | -0.15 | 5 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-1.5, 0.5] | 0.15 | 0.5998 | -0.4286 | 6 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 461 | 20 |
| hybrid-linucb-d-arm-g098 | 461 | 20 |
| hybrid-linucb-d-g095 | 461 | 20 |
| hybrid-linucb-d-g098 | 461 | 20 |
| hybrid-linucb-d-g099 | 461 | 20 |
| leastloaded | 461 | 20 |
| sed | 461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-0.5, 1] | 0.6281 | 1 | 0.1429 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.4145 | 1 | -0.225 | 5 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-0.5, 1] | 0.5922 | 1 | 0.15 | 5 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-0.5, 1.5] | 0.15 | 0.5998 | 0.4286 | 6 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.22 | 20 |
| hybrid-linucb-d-arm-g098 | 0.21 | 20 |
| hybrid-linucb-d-g095 | 0.22 | 20 |
| hybrid-linucb-d-g098 | 0.22 | 20 |
| hybrid-linucb-d-g099 | 0.17 | 20 |
| leastloaded | 0.21 | 20 |
| sed | 0.16 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-0.04, 0.07] | 0.5691 | 1 | 0.1618 | 4 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02 [0, 0.07] | 0.214 | 0.856 | 0.3333 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.01 [-0.05, 0.06] | 0.7627 | 1 | 0.07895 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [-0.06, 0.04] | 0.7046 | 1 | -0.1046 | 3 |

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2643 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2375 | 20 |
| hybrid-linucb-d-g095 | 0.2382 | 20 |
| hybrid-linucb-d-g098 | 0.2555 | 20 |
| hybrid-linucb-d-g099 | 0.264 | 20 |
| leastloaded | 0.236 | 20 |
| sed | 0.2409 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.02616 [-0.04753, 0.005762] | 0.07585 | 0.2276 | -0.4571 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.01795 [-0.06118, 0.003878] | 0.02664 | 0.1066 | -0.5619 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01523 [-0.04629, 0.0292] | 0.4524 | 0.9048 | -0.2 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.004182 [-0.02526, 0.02366] | 0.7841 | 0.9048 | -0.07619 | 0 |

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.05395 | 20 |
| hybrid-linucb-d-arm-g098 | -0.0964 | 20 |
| hybrid-linucb-d-g095 | -0.1291 | 20 |
| hybrid-linucb-d-g098 | -0.01658 | 20 |
| hybrid-linucb-d-g099 | -0.08132 | 20 |
| leastloaded | -0.1792 | 20 |
| sed | -0.1745 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.02506 [-0.07416, 0.03975] | 0.5958 | 0.8321 | -0.1429 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.06406 [-0.2089, 0.07326] | 0.2024 | 0.8098 | -0.3333 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.03216 [-0.02412, 0.07214] | 0.33 | 0.8321 | 0.2571 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.05602 [-0.1717, 0.09561] | 0.2774 | 0.8321 | -0.2857 | 0 |

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.1801 | 20 |
| hybrid-linucb-d-arm-g098 | -0.1903 | 20 |
| hybrid-linucb-d-g095 | -0.1643 | 20 |
| hybrid-linucb-d-g098 | -0.1763 | 20 |
| hybrid-linucb-d-g099 | -0.1802 | 20 |
| leastloaded | -0.2408 | 20 |
| sed | -0.2243 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.0181 [-0.02964, 0.02318] | 0.6742 | 1 | -0.1143 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02858 [0.01343, 0.03654] | 0.002712 | 0.01085 | 0.7333 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.001969 [-0.01671, 0.02426] | 0.7285 | 1 | 0.09524 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.002277 [-0.01886, 0.02569] | 0.8983 | 1 | 0.0381 | 0 |

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 59 | 20 |
| hybrid-linucb-d-arm-g098 | 29.5 | 20 |
| hybrid-linucb-d-g095 | 33 | 20 |
| hybrid-linucb-d-g098 | 38 | 20 |
| hybrid-linucb-d-g099 | 37 | 20 |
| leastloaded | 43.5 | 20 |
| sed | 20 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.5 [-42.5, 35.5] | 0.7337 | 0.7337 | -0.09942 | 2 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-46, 11.5] | 0.2762 | 0.6288 | -0.2924 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-81.5, 14] | 0.2096 | 0.6288 | -0.3464 | 3 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -14.5 [-85.5, 14] | 0.1031 | 0.4124 | -0.4263 | 1 |

## target_after_recovery_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-arm-g098 | NA | 0 |
| hybrid-linucb-d-g095 | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g099 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## target_after_recovery_tasks

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-arm-g098 | NA | 0 |
| hybrid-linucb-d-g095 | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g099 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2033 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1946 | 20 |
| hybrid-linucb-d-g095 | 0.2099 | 20 |
| hybrid-linucb-d-g098 | 0.2128 | 20 |
| hybrid-linucb-d-g099 | 0.2072 | 20 |
| leastloaded | 0.203 | 20 |
| sed | 0.1818 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.03251 [-0.04281, 0.01824] | 0.1893 | 0.7574 | -0.3429 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.007477 [-0.03732, 0.02706] | 0.7012 | 1 | -0.1048 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.0005024 [-0.04833, 0.02794] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.002244 [-0.03073, 0.02047] | 0.8983 | 1 | -0.0381 | 0 |

## drift_on_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 242.5 | 20 |
| hybrid-linucb-d-arm-g098 | 246 | 20 |
| hybrid-linucb-d-g095 | 237.5 | 20 |
| hybrid-linucb-d-g098 | 239.5 | 20 |
| hybrid-linucb-d-g099 | 242.5 | 20 |
| leastloaded | 235.5 | 20 |
| sed | 246.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.5 [-9, 9] | 0.6537 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.5 [-10, 5.5] | 0.3756 | 1 | -0.2316 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.5 [-10.5, 4.5] | 0.4204 | 1 | -0.2105 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -2 [-10, 7] | 0.4326 | 1 | -0.2 | 0 |

## drift_on_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 50.5 | 20 |
| hybrid-linucb-d-arm-g098 | 47 | 20 |
| hybrid-linucb-d-g095 | 51 | 20 |
| hybrid-linucb-d-g098 | 51.5 | 20 |
| hybrid-linucb-d-g099 | 47.5 | 20 |
| leastloaded | 48.5 | 20 |
| sed | 44.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -6.5 [-12, 5.5] | 0.1847 | 0.7389 | -0.3381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -2 [-11, 6] | 0.4896 | 1 | -0.1762 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-9.5, 7.5] | 0.8812 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.5 [-7, 7] | 0.6271 | 1 | -0.1238 | 0 |

## drift_off_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3016 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2876 | 20 |
| hybrid-linucb-d-g095 | 0.2699 | 20 |
| hybrid-linucb-d-g098 | 0.2883 | 20 |
| hybrid-linucb-d-g099 | 0.305 | 20 |
| leastloaded | 0.2778 | 20 |
| sed | 0.2776 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.03043 [-0.05166, 0.005373] | 0.1769 | 0.5306 | -0.3524 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.03548 [-0.0703, -0.001049] | 0.008308 | 0.03323 | -0.6571 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.009319 [-0.05275, 0.01401] | 0.2774 | 0.5547 | -0.2857 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.01602 [-0.03078, 0.02087] | 0.5958 | 0.5958 | -0.1429 | 0 |

## drift_off_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 218 | 20 |
| hybrid-linucb-d-arm-g098 | 215 | 20 |
| hybrid-linucb-d-g095 | 223 | 20 |
| hybrid-linucb-d-g098 | 220.5 | 20 |
| hybrid-linucb-d-g099 | 220 | 20 |
| leastloaded | 225.5 | 20 |
| sed | 214.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-9, 10] | 0.8369 | 1 | -0.05238 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.5 [-6.5, 10] | 0.3978 | 1 | 0.2211 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.5 [-4, 10] | 0.4206 | 1 | 0.2105 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 1.5 [-6, 10.5] | 0.3312 | 1 | 0.2476 | 0 |

## drift_off_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 65 | 20 |
| hybrid-linucb-d-arm-g098 | 64 | 20 |
| hybrid-linucb-d-g095 | 61 | 20 |
| hybrid-linucb-d-g098 | 64 | 20 |
| hybrid-linucb-d-g099 | 67 | 20 |
| leastloaded | 63 | 20 |
| sed | 60 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -5.5 [-9.5, 3] | 0.1846 | 0.5538 | -0.3381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -8 [-17, 1] | 0.01652 | 0.06606 | -0.6263 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -2.5 [-11.5, 7] | 0.3602 | 0.7203 | -0.2333 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -3.5 [-7, 7.5] | 0.8372 | 0.8372 | -0.05238 | 0 |

## drift_phase_1_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2133 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2212 | 20 |
| hybrid-linucb-d-g095 | 0.2158 | 20 |
| hybrid-linucb-d-g098 | 0.2243 | 20 |
| hybrid-linucb-d-g099 | 0.1809 | 20 |
| leastloaded | 0.202 | 20 |
| sed | 0.1296 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.008543 [-0.03507, 0.06392] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02752 [0.0006773, 0.08892] | 0.1536 | 0.6146 | 0.3714 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.01469 [-0.03109, 0.07902] | 0.5217 | 1 | 0.1714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.002889 [-0.05874, 0.05592] | 0.8695 | 1 | -0.04762 | 0 |

## drift_phase_2_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3069 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2674 | 20 |
| hybrid-linucb-d-g095 | 0.2651 | 20 |
| hybrid-linucb-d-g098 | 0.2373 | 20 |
| hybrid-linucb-d-g099 | 0.3025 | 20 |
| leastloaded | 0.291 | 20 |
| sed | 0.2892 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.02762 [-0.08594, 0.03893] | 0.5706 | 1 | -0.1524 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.02515 [-0.08682, 0.006148] | 0.3465 | 1 | -0.2632 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01463 [-0.1046, 0.02729] | 0.1536 | 0.6146 | -0.3714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01317 [-0.02184, 0.07584] | 0.7012 | 1 | 0.1048 | 0 |

## drift_phase_3_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.198 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1786 | 20 |
| hybrid-linucb-d-g095 | 0.1758 | 20 |
| hybrid-linucb-d-g098 | 0.198 | 20 |
| hybrid-linucb-d-g099 | 0.1846 | 20 |
| leastloaded | 0.1716 | 20 |
| sed | 0.2115 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.01826 [-0.06457, 0.04669] | 0.7841 | 1 | -0.07619 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.01493 [-0.04509, 0.02981] | 0.5217 | 1 | -0.1714 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.00975 [-0.04915, 0.03512] | 0.7381 | 1 | -0.09474 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.006818 [-0.0535, 0.02318] | 0.6705 | 1 | -0.1228 | 2 |

## drift_phase_4_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2857 | 20 |
| hybrid-linucb-d-arm-g098 | 0.3025 | 20 |
| hybrid-linucb-d-g095 | 0.266 | 20 |
| hybrid-linucb-d-g098 | 0.2633 | 20 |
| hybrid-linucb-d-g099 | 0.272 | 20 |
| leastloaded | 0.2429 | 20 |
| sed | 0.2683 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.004534 [-0.03599, 0.04043] | 1 | 1 | 0 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.01907 [-0.04608, 0.01216] | 0.2413 | 0.7074 | -0.3158 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.04548 [-0.08634, 0.008799] | 0.2305 | 0.7074 | -0.3143 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.02202 [-0.06312, 0.02009] | 0.1769 | 0.7074 | -0.3524 | 0 |

## drift_phase_5_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.1967 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2258 | 20 |
| hybrid-linucb-d-g095 | 0.1875 | 20 |
| hybrid-linucb-d-g098 | 0.1791 | 20 |
| hybrid-linucb-d-g099 | 0.2205 | 20 |
| leastloaded | 0.1895 | 20 |
| sed | 0.1667 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.01047 [-0.0225, 0.06568] | 0.4565 | 1 | 0.1947 | 1 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.002228 [-0.02978, 0.03613] | 0.9217 | 1 | 0.03158 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.007576 [-0.07353, 0.03134] | 0.6291 | 1 | -0.1263 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.0261 [-0.03779, 0.09549] | 0.4221 | 1 | 0.2048 | 0 |

## drift_phase_6_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3302 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2903 | 20 |
| hybrid-linucb-d-g095 | 0.2426 | 20 |
| hybrid-linucb-d-g098 | 0.3377 | 20 |
| hybrid-linucb-d-g099 | 0.3431 | 20 |
| leastloaded | 0.2817 | 20 |
| sed | 0.2639 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.0837 [-0.1441, 0.04167] | 0.1536 | 0.4609 | -0.3714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.0893 [-0.2077, 0.04594] | 0.08969 | 0.3588 | -0.4381 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01134 [-0.1048, 0.1123] | 1 | 1 | 0 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.0003064 [-0.09707, 0.08881] | 0.8124 | 1 | -0.06667 | 0 |

## drift_phase_7_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2386 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1473 | 20 |
| hybrid-linucb-d-g095 | 0.1695 | 20 |
| hybrid-linucb-d-g098 | 0.2589 | 20 |
| hybrid-linucb-d-g099 | 0.2018 | 20 |
| leastloaded | 0.1884 | 20 |
| sed | 0.1821 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.07086 [-0.1204, -0.01217] | 0.01373 | 0.05493 | -0.6286 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.06208 [-0.09541, -0.01144] | 0.2162 | 0.6485 | -0.3238 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.03312 [-0.09642, 0.07172] | 0.8408 | 1 | -0.05714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.04272 [-0.1197, 0.1076] | 0.7562 | 1 | -0.08571 | 0 |

## drift_phase_8_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.3277 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2832 | 20 |
| hybrid-linucb-d-g095 | 0.2673 | 20 |
| hybrid-linucb-d-g098 | 0.3018 | 20 |
| hybrid-linucb-d-g099 | 0.2798 | 20 |
| leastloaded | 0.2981 | 20 |
| sed | 0.2789 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.04631 [-0.07814, 0.05093] | 0.6742 | 1 | -0.1143 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.05843 [-0.1085, -0.01933] | 0.05317 | 0.2127 | -0.4952 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01445 [-0.083, 0.06766] | 0.8408 | 1 | -0.05714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.007915 [-0.08285, 0.06005] | 0.5217 | 1 | -0.1714 | 0 |

## drift_phase_9_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2264 | 20 |
| hybrid-linucb-d-arm-g098 | 0.215 | 20 |
| hybrid-linucb-d-g095 | 0.2411 | 20 |
| hybrid-linucb-d-g098 | 0.1883 | 20 |
| hybrid-linucb-d-g099 | 0.2308 | 20 |
| leastloaded | 0.2041 | 20 |
| sed | 0.2234 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.004757 [-0.03088, 0.02646] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.02486 [-0.02614, 0.06175] | 0.2611 | 1 | 0.2952 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01777 [-0.05818, 0.04156] | 0.498 | 1 | -0.181 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01554 [-0.02863, 0.05196] | 0.5706 | 1 | 0.1524 | 0 |

## drift_phase_10_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2911 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2992 | 20 |
| hybrid-linucb-d-g095 | 0.2665 | 20 |
| hybrid-linucb-d-g098 | 0.2857 | 20 |
| hybrid-linucb-d-g099 | 0.2836 | 20 |
| leastloaded | 0.2429 | 20 |
| sed | 0.2466 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.02614 [-0.04265, 0.05401] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.004121 [-0.061, 0.05387] | 0.8596 | 1 | -0.05263 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.01597 [-0.07166, 0.05911] | 0.7285 | 1 | -0.09524 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.01943 [-0.07079, 0.07001] | 0.9563 | 1 | -0.01905 | 0 |

## drift_phase_11_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 1 |
| hybrid-linucb-d-arm-g098 | NA | 0 |
| hybrid-linucb-d-g095 | 0 | 7 |
| hybrid-linucb-d-g098 | 0.5 | 2 |
| hybrid-linucb-d-g099 | 0 | 1 |
| leastloaded | 0.1 | 6 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-4:50051 | 1000 |
| 1 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 1 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-5:50051 | 1100 |
| 1 | sed | worker-5:50051 | 1100 |
| 2 | hybrid-linucb | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 2 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 2 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-d-g099 | worker-2:50051 | 600 |
| 2 | leastloaded | worker-4:50051 | 1000 |
| 2 | sed | worker-5:50051 | 1100 |
| 3 | hybrid-linucb | worker-5:50051 | 1100 |
| 3 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 3 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 3 | leastloaded | worker-5:50051 | 1100 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 4 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 4 | leastloaded | worker-1:50051 | 500 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 5 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 5 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 5 | leastloaded | worker-1:50051 | 500 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-5:50051 | 1100 |
| 6 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 6 | leastloaded | worker-2:50051 | 600 |
| 6 | sed | worker-4:50051 | 1000 |
| 7 | hybrid-linucb | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 7 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 7 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 7 | leastloaded | worker-2:50051 | 600 |
| 7 | sed | worker-4:50051 | 1000 |
| 8 | hybrid-linucb | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 8 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 8 | leastloaded | worker-4:50051 | 1000 |
| 8 | sed | worker-4:50051 | 1000 |
| 9 | hybrid-linucb | worker-4:50051 | 1000 |
| 9 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 9 | leastloaded | worker-5:50051 | 1100 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-3:50051 | 800 |
| 10 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 10 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g099 | worker-2:50051 | 600 |
| 10 | leastloaded | worker-3:50051 | 800 |
| 10 | sed | worker-5:50051 | 1100 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 11 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 11 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 11 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 11 | leastloaded | worker-4:50051 | 1000 |
| 11 | sed | worker-4:50051 | 1000 |
| 12 | hybrid-linucb | worker-5:50051 | 1100 |
| 12 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-3:50051 | 800 |
| 13 | hybrid-linucb | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 13 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 13 | leastloaded | worker-3:50051 | 800 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 14 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 14 | leastloaded | worker-1:50051 | 500 |
| 14 | sed | worker-4:50051 | 1000 |
| 15 | hybrid-linucb | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 15 | leastloaded | worker-3:50051 | 800 |
| 15 | sed | worker-3:50051 | 800 |
| 16 | hybrid-linucb | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 16 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 16 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 16 | leastloaded | worker-4:50051 | 1000 |
| 16 | sed | worker-5:50051 | 1100 |
| 17 | hybrid-linucb | worker-3:50051 | 800 |
| 17 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 17 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 17 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 17 | leastloaded | worker-3:50051 | 800 |
| 17 | sed | worker-5:50051 | 1100 |
| 18 | hybrid-linucb | worker-5:50051 | 1100 |
| 18 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 18 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 18 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 18 | leastloaded | worker-1:50051 | 500 |
| 18 | sed | worker-4:50051 | 1000 |
| 19 | hybrid-linucb | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 19 | leastloaded | worker-5:50051 | 1100 |
| 19 | sed | worker-5:50051 | 1100 |
| 20 | hybrid-linucb | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 20 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 20 | leastloaded | worker-4:50051 | 1000 |
| 20 | sed | worker-3:50051 | 800 |
