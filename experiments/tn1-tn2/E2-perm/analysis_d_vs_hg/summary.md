# Experiment analysis

Complete blocks: 20.
Time-to-adapt: W=20, reference=pre_all share, threshold=0.5 x pre_all share; reference share < 0.05 is NA. Censored values display as > n and enter paired tests as n_post + 1.
target_after_recovery_* is defined only for exactly one drift_on and one non-cleanup drift_off (transient); NA for onoff and permanent by design.
On/off aggregate shares pool decisions inside each phase after first drift_on; cell_end cleanup is excluded. Counts are decision and target-decision totals.
Holm family: every treatment x baseline pair for one metric in this directory; gamma/lambda sweep arms enlarge the family.

## Phases crossing build boundaries

| Round | Arm | Phases |
|---:|---|---|
| 1 | hybrid-linucb | 1_on |
| 1 | hybrid-linucb-d-arm-g098 | 1_on |
| 1 | hybrid-linucb-d-g095 | 1_on |
| 1 | hybrid-linucb-d-g098 | 1_on |
| 1 | hybrid-linucb-d-g099 | 1_on |
| 1 | leastloaded | 1_on |
| 1 | sed | 1_on |
| 2 | hybrid-linucb | 1_on |
| 2 | hybrid-linucb-d-arm-g098 | 1_on |
| 2 | hybrid-linucb-d-g095 | 1_on |
| 2 | hybrid-linucb-d-g098 | 1_on |
| 2 | hybrid-linucb-d-g099 | 1_on |
| 2 | leastloaded | 1_on |
| 2 | sed | 1_on |
| 3 | hybrid-linucb | 1_on |
| 3 | hybrid-linucb-d-arm-g098 | 1_on |
| 3 | hybrid-linucb-d-g095 | 1_on |
| 3 | hybrid-linucb-d-g098 | 1_on |
| 3 | hybrid-linucb-d-g099 | 1_on |
| 3 | leastloaded | 1_on |
| 3 | sed | 1_on |
| 4 | hybrid-linucb | 1_on |
| 4 | hybrid-linucb-d-arm-g098 | 1_on |
| 4 | hybrid-linucb-d-g095 | 1_on |
| 4 | hybrid-linucb-d-g098 | 1_on |
| 4 | hybrid-linucb-d-g099 | 1_on |
| 4 | leastloaded | 1_on |
| 4 | sed | 1_on |
| 5 | hybrid-linucb | 1_on |
| 5 | hybrid-linucb-d-arm-g098 | 1_on |
| 5 | hybrid-linucb-d-g095 | 1_on |
| 5 | hybrid-linucb-d-g098 | 1_on |
| 5 | hybrid-linucb-d-g099 | 1_on |
| 5 | leastloaded | 1_on |
| 5 | sed | 1_on |
| 6 | hybrid-linucb | 1_on |
| 6 | hybrid-linucb-d-arm-g098 | 1_on |
| 6 | hybrid-linucb-d-g095 | 1_on |
| 6 | hybrid-linucb-d-g098 | 1_on |
| 6 | hybrid-linucb-d-g099 | 1_on |
| 6 | leastloaded | 1_on |
| 6 | sed | 1_on |
| 7 | hybrid-linucb | 1_on |
| 7 | hybrid-linucb-d-arm-g098 | 1_on |
| 7 | hybrid-linucb-d-g095 | 1_on |
| 7 | hybrid-linucb-d-g098 | 1_on |
| 7 | hybrid-linucb-d-g099 | 1_on |
| 7 | leastloaded | 1_on |
| 7 | sed | 1_on |
| 8 | hybrid-linucb | 1_on |
| 8 | hybrid-linucb-d-arm-g098 | 1_on |
| 8 | hybrid-linucb-d-g095 | 1_on |
| 8 | hybrid-linucb-d-g098 | 1_on |
| 8 | hybrid-linucb-d-g099 | 1_on |
| 8 | leastloaded | 1_on |
| 8 | sed | 1_on |
| 9 | hybrid-linucb | 1_on |
| 9 | hybrid-linucb-d-arm-g098 | 1_on |
| 9 | hybrid-linucb-d-g095 | 1_on |
| 9 | hybrid-linucb-d-g098 | 1_on |
| 9 | hybrid-linucb-d-g099 | 1_on |
| 9 | leastloaded | 1_on |
| 9 | sed | 1_on |
| 10 | hybrid-linucb | 1_on |
| 10 | hybrid-linucb-d-arm-g098 | 1_on |
| 10 | hybrid-linucb-d-g095 | 1_on |
| 10 | hybrid-linucb-d-g098 | 1_on |
| 10 | hybrid-linucb-d-g099 | 1_on |
| 10 | leastloaded | 1_on |
| 10 | sed | 1_on |
| 11 | hybrid-linucb | 1_on |
| 11 | hybrid-linucb-d-arm-g098 | 1_on |
| 11 | hybrid-linucb-d-g095 | 1_on |
| 11 | hybrid-linucb-d-g098 | 1_on |
| 11 | hybrid-linucb-d-g099 | 1_on |
| 11 | leastloaded | 1_on |
| 11 | sed | 1_on |
| 12 | hybrid-linucb | 1_on |
| 12 | hybrid-linucb-d-arm-g098 | 1_on |
| 12 | hybrid-linucb-d-g095 | 1_on |
| 12 | hybrid-linucb-d-g098 | 1_on |
| 12 | hybrid-linucb-d-g099 | 1_on |
| 12 | leastloaded | 1_on |
| 12 | sed | 1_on |
| 13 | hybrid-linucb | 1_on |
| 13 | hybrid-linucb-d-arm-g098 | 1_on |
| 13 | hybrid-linucb-d-g095 | 1_on |
| 13 | hybrid-linucb-d-g098 | 1_on |
| 13 | hybrid-linucb-d-g099 | 1_on |
| 13 | leastloaded | 1_on |
| 13 | sed | 1_on |
| 14 | hybrid-linucb | 1_on |
| 14 | hybrid-linucb-d-arm-g098 | 1_on |
| 14 | hybrid-linucb-d-g095 | 1_on |
| 14 | hybrid-linucb-d-g098 | 1_on |
| 14 | hybrid-linucb-d-g099 | 1_on |
| 14 | leastloaded | 1_on |
| 14 | sed | 1_on |
| 15 | hybrid-linucb | 1_on |
| 15 | hybrid-linucb-d-arm-g098 | 1_on |
| 15 | hybrid-linucb-d-g095 | 1_on |
| 15 | hybrid-linucb-d-g098 | 1_on |
| 15 | hybrid-linucb-d-g099 | 1_on |
| 15 | leastloaded | 1_on |
| 15 | sed | 1_on |
| 16 | hybrid-linucb | 1_on |
| 16 | hybrid-linucb-d-arm-g098 | 1_on |
| 16 | hybrid-linucb-d-g095 | 1_on |
| 16 | hybrid-linucb-d-g098 | 1_on |
| 16 | hybrid-linucb-d-g099 | 1_on |
| 16 | leastloaded | 1_on |
| 16 | sed | 1_on |
| 17 | hybrid-linucb | 1_on |
| 17 | hybrid-linucb-d-arm-g098 | 1_on |
| 17 | hybrid-linucb-d-g095 | 1_on |
| 17 | hybrid-linucb-d-g098 | 1_on |
| 17 | hybrid-linucb-d-g099 | 1_on |
| 17 | leastloaded | 1_on |
| 17 | sed | 1_on |
| 18 | hybrid-linucb | 1_on |
| 18 | hybrid-linucb-d-arm-g098 | 1_on |
| 18 | hybrid-linucb-d-g095 | 1_on |
| 18 | hybrid-linucb-d-g098 | 1_on |
| 18 | hybrid-linucb-d-g099 | 1_on |
| 18 | leastloaded | 1_on |
| 18 | sed | 1_on |
| 19 | hybrid-linucb | 1_on |
| 19 | hybrid-linucb-d-arm-g098 | 1_on |
| 19 | hybrid-linucb-d-g095 | 1_on |
| 19 | hybrid-linucb-d-g098 | 1_on |
| 19 | hybrid-linucb-d-g099 | 1_on |
| 19 | leastloaded | 1_on |
| 19 | sed | 1_on |
| 20 | hybrid-linucb | 1_on |
| 20 | hybrid-linucb-d-arm-g098 | 1_on |
| 20 | hybrid-linucb-d-g095 | 1_on |
| 20 | hybrid-linucb-d-g098 | 1_on |
| 20 | hybrid-linucb-d-g099 | 1_on |
| 20 | leastloaded | 1_on |
| 20 | sed | 1_on |

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 184.1 | 20 |
| hybrid-linucb-d-arm-g098 | 184.1 | 20 |
| hybrid-linucb-d-g095 | 190.2 | 20 |
| hybrid-linucb-d-g098 | 188.5 | 20 |
| hybrid-linucb-d-g099 | 186.6 | 20 |
| leastloaded | 189.5 | 20 |
| sed | 186.3 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.3155 [-1.008, 1.166] | 0.8983 | 0.8983 | 0.0381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 6.342 [3.826, 7.774] | 4.768e-05 | 0.0001907 | 0.9238 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 3.969 [2.093, 6.066] | 0.0008507 | 0.002552 | 0.8 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 2.239 [-1.063, 5.114] | 0.06958 | 0.1392 | 0.4667 | 0 |

## build_1_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 91.45 | 20 |
| hybrid-linucb-d-arm-g098 | 89.96 | 20 |
| hybrid-linucb-d-g095 | 94.17 | 20 |
| hybrid-linucb-d-g098 | 92.8 | 20 |
| hybrid-linucb-d-g099 | 91.34 | 20 |
| leastloaded | 92.82 | 20 |
| sed | 89.9 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.1215 [-2.057, 0.7195] | 0.5217 | 1 | -0.1714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 2.468 [1.387, 3.785] | 0.001017 | 0.004066 | 0.7905 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 1.033 [0.2975, 2.099] | 0.1044 | 0.3131 | 0.4143 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.0125 [-1.632, 2.021] | 0.6477 | 1 | 0.1238 | 0 |

## build_2_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 93.1 | 20 |
| hybrid-linucb-d-arm-g098 | 93.43 | 20 |
| hybrid-linucb-d-g095 | 97.72 | 20 |
| hybrid-linucb-d-g098 | 97.3 | 20 |
| hybrid-linucb-d-g099 | 94.46 | 20 |
| leastloaded | 97.51 | 20 |
| sed | 94.29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.021 [-0.6575, 1.126] | 0.7285 | 0.7285 | 0.09524 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 3.828 [1.73, 5.599] | 8.202e-05 | 0.000246 | 0.9048 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 3.062 [1.459, 4.224] | 9.537e-06 | 3.815e-05 | 0.9714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 1.371 [-0.1035, 3.9] | 0.01923 | 0.03847 | 0.5905 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.005085 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.07203 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.005085 [0.002542, 0.00678] | 0.0001228 | 0.0004911 | 1 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 0.5637 | 1 | 0.3333 | 17 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.009694 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1614 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.009694 [0.00516, 0.01316] | 3.815e-06 | 1.526e-05 | 0.9905 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 0 | 17 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9431 | 20 |
| hybrid-linucb-d-arm-g098 | 0.9487 | 20 |
| hybrid-linucb-d-g095 | 0.8776 | 20 |
| hybrid-linucb-d-g098 | 0.9175 | 20 |
| hybrid-linucb-d-g099 | 0.9269 | 20 |
| leastloaded | 0.9232 | 20 |
| sed | 0.8386 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.006203 [-0.01436, 0.01953] | 0.6215 | 0.6215 | 0.1333 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.07447 [-0.08999, -0.05134] | 1.907e-06 | 7.629e-06 | -1 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.03004 [-0.05065, -0.01495] | 0.001432 | 0.004297 | -0.7714 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.019 [-0.03712, 0.01483] | 0.2162 | 0.4323 | -0.3238 | 0 |

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
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |

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
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2212 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2246 | 20 |
| hybrid-linucb-d-g095 | 0.2508 | 20 |
| hybrid-linucb-d-g098 | 0.2136 | 20 |
| hybrid-linucb-d-g099 | 0.2186 | 20 |
| leastloaded | 0.2085 | 20 |
| sed | 0.2051 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.005085 [-0.02458, 0.01271] | 0.9519 | 1 | -0.01579 | 1 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.04746 [0.01017, 0.05508] | 0.002484 | 0.009937 | 0.7714 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.008475 [-0.02542, 0.009322] | 0.6541 | 1 | -0.1143 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.00678 [-0.03475, 0.02034] | 0.3602 | 1 | -0.2333 | 0 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.4195 | 20 |
| hybrid-linucb-d-arm-g098 | 0.4364 | 20 |
| hybrid-linucb-d-g095 | 0.4992 | 20 |
| hybrid-linucb-d-g098 | 0.4525 | 20 |
| hybrid-linucb-d-g099 | 0.4593 | 20 |
| leastloaded | 0.4288 | 20 |
| sed | 0.4508 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.01525 [-0.01864, 0.03305] | 0.3702 | 0.3702 | 0.2286 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.08305 [0.06271, 0.1025] | 4.768e-05 | 0.0001907 | 0.9238 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.04322 [0.02288, 0.06864] | 0.0004538 | 0.001162 | 0.9415 | 2 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.04153 [0.02712, 0.05339] | 0.0003875 | 0.001162 | 0.9048 | 0 |

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
| hybrid-linucb | 0.2929 | 20 |
| hybrid-linucb-d-arm-g098 | 0.344 | 20 |
| hybrid-linucb-d-g095 | 0.3929 | 20 |
| hybrid-linucb-d-g098 | 0.2583 | 20 |
| hybrid-linucb-d-g099 | 0.2659 | 20 |
| leastloaded | 0.2167 | 20 |
| sed | 0.2069 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.06034 [-0.01259, 0.1355] | 0.1429 | 0.5716 | 0.381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-0.05172, 0.1517] | 0.4171 | 0.8342 | 0.2281 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.03399 [-0.1641, 0.04259] | 0.2432 | 0.7295 | -0.3053 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.06466 [-0.1908, 0.1237] | 0.6742 | 0.8342 | -0.1143 | 0 |

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2539 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2734 | 20 |
| hybrid-linucb-d-g095 | 0.2869 | 20 |
| hybrid-linucb-d-g098 | 0.2346 | 20 |
| hybrid-linucb-d-g099 | 0.2731 | 20 |
| leastloaded | 0.2385 | 20 |
| sed | 0.2451 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.05783 [-0.03399, 0.09061] | 0.05317 | 0.2127 | 0.4952 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.01649 [-0.01997, 0.08612] | 0.2774 | 0.8321 | 0.2857 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.003089 [-0.06326, 0.04252] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.01274 [-0.0814, 0.08244] | 0.751 | 1 | 0.08095 | 0 |

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2046 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1939 | 20 |
| hybrid-linucb-d-g095 | 0.1961 | 20 |
| hybrid-linucb-d-g098 | 0.19 | 20 |
| hybrid-linucb-d-g099 | 0.1872 | 20 |
| leastloaded | 0.1876 | 20 |
| sed | 0.2028 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.01627 [-0.02675, 0.008879] | 0.08969 | 0.3588 | -0.4381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.0006964 [-0.04162, 0.02321] | 0.4653 | 0.8182 | -0.2 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.009397 [-0.03684, 0.01607] | 0.4091 | 0.8182 | -0.219 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.01475 [-0.03683, 0.01352] | 0.2707 | 0.8122 | -0.281 | 0 |

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 29 | 20 |
| hybrid-linucb-d-arm-g098 | 29 | 20 |
| hybrid-linucb-d-g095 | 28.5 | 20 |
| hybrid-linucb-d-g098 | 29 | 20 |
| hybrid-linucb-d-g099 | 29 | 20 |
| leastloaded | 29 | 20 |
| sed | 29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.5376 | 1 | -0.181 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-1.5, 1] | 0.2351 | 0.9404 | -0.3309 | 4 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-0.5, 0] | 0.7961 | 1 | -0.09091 | 10 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.5 [-1, 0.5] | 0.612 | 1 | -0.1397 | 4 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 129 | 20 |
| hybrid-linucb-d-arm-g098 | 129 | 20 |
| hybrid-linucb-d-g095 | 128.5 | 20 |
| hybrid-linucb-d-g098 | 129 | 20 |
| hybrid-linucb-d-g099 | 129 | 20 |
| leastloaded | 129 | 20 |
| sed | 129 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-1, 0.5] | 0.5376 | 1 | -0.181 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-1.5, 1] | 0.2351 | 0.9404 | -0.3309 | 4 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [-0.5, 0] | 0.7961 | 1 | -0.09091 | 10 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.5 [-1, 0.5] | 0.612 | 1 | -0.1397 | 4 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 461 | 20 |
| hybrid-linucb-d-arm-g098 | 461 | 20 |
| hybrid-linucb-d-g095 | 461.5 | 20 |
| hybrid-linucb-d-g098 | 461 | 20 |
| hybrid-linucb-d-g099 | 461 | 20 |
| leastloaded | 461 | 20 |
| sed | 461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0 [-0.5, 1] | 0.5376 | 1 | 0.181 | 6 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0 [-1, 1.5] | 0.2351 | 0.9404 | 0.3309 | 4 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0 [0, 0.5] | 0.7961 | 1 | 0.09091 | 10 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.5 [-0.5, 1] | 0.612 | 1 | 0.1397 | 4 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.23 | 20 |
| hybrid-linucb-d-arm-g098 | 0.23 | 20 |
| hybrid-linucb-d-g095 | 0.22 | 20 |
| hybrid-linucb-d-g098 | 0.22 | 20 |
| hybrid-linucb-d-g099 | 0.22 | 20 |
| leastloaded | 0.2 | 20 |
| sed | 0.17 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.03 [-0.07, 0.08] | 0.9305 | 1 | -0.02339 | 2 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.02 [-0.04, 0.06] | 0.9132 | 1 | 0.02924 | 2 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.02 [-0.08, 0.09] | 0.8247 | 1 | -0.05789 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.05 [-0.08, 0.02] | 0.07623 | 0.3049 | -0.4632 | 1 |

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2002 | 20 |
| hybrid-linucb-d-arm-g098 | 0.192 | 20 |
| hybrid-linucb-d-g095 | 0.1908 | 20 |
| hybrid-linucb-d-g098 | 0.1881 | 20 |
| hybrid-linucb-d-g099 | 0.1888 | 20 |
| leastloaded | 0.1949 | 20 |
| sed | 0.2029 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.009194 [-0.0265, 0.001211] | 0.03623 | 0.1449 | -0.5333 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.009483 [-0.04639, 0.02211] | 0.4749 | 1 | -0.1905 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.01612 [-0.03514, 0.01197] | 0.4524 | 1 | -0.2 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.009977 [-0.03577, 0.01031] | 0.498 | 1 | -0.181 | 0 |

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.02962 | 20 |
| hybrid-linucb-d-arm-g098 | -0.02711 | 20 |
| hybrid-linucb-d-g095 | -0.1047 | 20 |
| hybrid-linucb-d-g098 | -0.1133 | 20 |
| hybrid-linucb-d-g099 | -0.07983 | 20 |
| leastloaded | -0.237 | 20 |
| sed | -0.1945 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | 0.02682 [0.01042, 0.1192] | 0.2162 | 0.4323 | 0.3238 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.07733 [-0.1227, -0.00527] | 0.00486 | 0.01944 | -0.6952 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.08158 [-0.1495, -0.006652] | 0.02664 | 0.07993 | -0.5619 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | 0.007544 [-0.1319, 0.0953] | 0.4524 | 0.4524 | -0.2 | 0 |

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.2301 | 20 |
| hybrid-linucb-d-arm-g098 | -0.2438 | 20 |
| hybrid-linucb-d-g095 | -0.1786 | 20 |
| hybrid-linucb-d-g098 | -0.2236 | 20 |
| hybrid-linucb-d-g099 | -0.2359 | 20 |
| leastloaded | -0.2777 | 20 |
| sed | -0.2468 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.008108 [-0.0209, 0.01472] | 0.3488 | 1 | -0.2476 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 0.05876 [0.02401, 0.08007] | 8.202e-05 | 0.0003281 | 0.9048 | 0 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 0.01086 [-0.02207, 0.03778] | 0.4091 | 1 | 0.219 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.01305 [-0.02733, 0.01604] | 0.7285 | 1 | -0.09524 | 0 |

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 50 | 20 |
| hybrid-linucb-d-arm-g098 | 41 | 20 |
| hybrid-linucb-d-g095 | 50 | 20 |
| hybrid-linucb-d-g098 | 45.5 | 20 |
| hybrid-linucb-d-g099 | 32 | 20 |
| leastloaded | 26 | 20 |
| sed | 23 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.5 [-27.5, 11] | 0.5014 | 1 | -0.1714 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | 12 [-12, 22] | 0.5195 | 1 | 0.1684 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | 2.5 [-17.51, 11.5] | 0.7935 | 1 | -0.06842 | 1 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -12 [-33.5, -1] | 0.0701 | 0.2804 | -0.4737 | 1 |

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
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-arm-g098 | NA | 0 |
| hybrid-linucb-d-g095 | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g099 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_n

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

## drift_on_phases_target_n

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

## drift_off_phases_share

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

## drift_off_phases_n

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

## drift_off_phases_target_n

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

## drift_phase_1_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2046 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1939 | 20 |
| hybrid-linucb-d-g095 | 0.1961 | 20 |
| hybrid-linucb-d-g098 | 0.19 | 20 |
| hybrid-linucb-d-g099 | 0.1872 | 20 |
| leastloaded | 0.1876 | 20 |
| sed | 0.2028 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb-d-arm-g098 | hybrid-linucb | 20 | -0.01627 [-0.02675, 0.008879] | 0.08969 | 0.3588 | -0.4381 | 0 |
| hybrid-linucb-d-g095 | hybrid-linucb | 20 | -0.0006964 [-0.04162, 0.02321] | 0.4653 | 0.8182 | -0.2 | 1 |
| hybrid-linucb-d-g098 | hybrid-linucb | 20 | -0.009397 [-0.03684, 0.01607] | 0.4091 | 0.8182 | -0.219 | 0 |
| hybrid-linucb-d-g099 | hybrid-linucb | 20 | -0.01475 [-0.03683, 0.01352] | 0.2707 | 0.8122 | -0.281 | 0 |

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 1 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 1 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 1 | leastloaded | worker-2:50051 | 600 |
| 1 | sed | worker-4:50051 | 1000 |
| 2 | hybrid-linucb | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 2 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 2 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 2 | leastloaded | worker-2:50051 | 600 |
| 2 | sed | worker-3:50051 | 800 |
| 3 | hybrid-linucb | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g099 | worker-2:50051 | 600 |
| 3 | leastloaded | worker-3:50051 | 800 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 4 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 4 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 4 | leastloaded | worker-1:50051 | 500 |
| 4 | sed | worker-5:50051 | 1100 |
| 5 | hybrid-linucb | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 5 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 5 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 5 | leastloaded | worker-1:50051 | 500 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 6 | leastloaded | worker-1:50051 | 500 |
| 6 | sed | worker-5:50051 | 1100 |
| 7 | hybrid-linucb | worker-4:50051 | 1000 |
| 7 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 7 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 7 | leastloaded | worker-5:50051 | 1100 |
| 7 | sed | worker-3:50051 | 800 |
| 8 | hybrid-linucb | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 8 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 8 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 8 | leastloaded | worker-5:50051 | 1100 |
| 8 | sed | worker-5:50051 | 1100 |
| 9 | hybrid-linucb | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 9 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 9 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 9 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 9 | leastloaded | worker-3:50051 | 800 |
| 9 | sed | worker-4:50051 | 1000 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 10 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 10 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 10 | leastloaded | worker-2:50051 | 600 |
| 10 | sed | worker-3:50051 | 800 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 11 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 11 | hybrid-linucb-d-g099 | worker-2:50051 | 600 |
| 11 | leastloaded | worker-4:50051 | 1000 |
| 11 | sed | worker-5:50051 | 1100 |
| 12 | hybrid-linucb | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 12 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-3:50051 | 800 |
| 13 | hybrid-linucb | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 13 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 13 | leastloaded | worker-2:50051 | 600 |
| 13 | sed | worker-3:50051 | 800 |
| 14 | hybrid-linucb | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 14 | leastloaded | worker-4:50051 | 1000 |
| 14 | sed | worker-5:50051 | 1100 |
| 15 | hybrid-linucb | worker-3:50051 | 800 |
| 15 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 15 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 15 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 15 | leastloaded | worker-3:50051 | 800 |
| 15 | sed | worker-4:50051 | 1000 |
| 16 | hybrid-linucb | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 16 | leastloaded | worker-2:50051 | 600 |
| 16 | sed | worker-3:50051 | 800 |
| 17 | hybrid-linucb | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 17 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | hybrid-linucb-d-g099 | worker-2:50051 | 600 |
| 17 | leastloaded | worker-3:50051 | 800 |
| 17 | sed | worker-2:50051 | 600 |
| 18 | hybrid-linucb | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 18 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 18 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 18 | leastloaded | worker-3:50051 | 800 |
| 18 | sed | worker-4:50051 | 1000 |
| 19 | hybrid-linucb | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 19 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 19 | leastloaded | worker-5:50051 | 1100 |
| 19 | sed | worker-5:50051 | 1100 |
| 20 | hybrid-linucb | worker-4:50051 | 1000 |
| 20 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 20 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 20 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 20 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 20 | leastloaded | worker-3:50051 | 800 |
| 20 | sed | worker-4:50051 | 1000 |
