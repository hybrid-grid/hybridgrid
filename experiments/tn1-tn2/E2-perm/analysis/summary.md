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
| hybrid-linucb | leastloaded | 20 | -6.27 [-6.969, -4.33] | 8.202e-05 | 0.0007381 | -0.9048 | 0 |
| hybrid-linucb | sed | 20 | -0.7045 [-2.641, 0.2425] | 0.1327 | 0.3982 | -0.3905 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -5.214 [-8.072, -3.379] | 1.907e-05 | 0.0001907 | -0.9524 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -1.641 [-3.865, -0.0855] | 0.06372 | 0.3186 | -0.4762 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.7425 [-1.967, 3.225] | 0.7285 | 1 | 0.09524 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 3.876 [0.494, 6.982] | 0.002712 | 0.01899 | 0.7333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -2.254 [-3.729, -0.4015] | 0.03623 | 0.2174 | -0.5333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.75 [-0.593, 5.676] | 0.08255 | 0.3302 | 0.4476 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -1.871 [-7.137, -0.6555] | 0.002325 | 0.0186 | -0.7429 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.4205 [-1.858, 2.283] | 0.9854 | 1 | -0.009524 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -1.766 [-2.522, -1.192] | 0.01718 | 0.1375 | -0.6 | 0 |
| hybrid-linucb | sed | 20 | 0.2325 [-1.554, 2.211] | 0.9273 | 1 | 0.02857 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -1.882 [-4.935, -0.4975] | 0.007296 | 0.06566 | -0.6667 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.1295 [-3.256, 1.083] | 0.5217 | 1 | -0.1714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 1.006 [-0.1705, 1.96] | 0.4091 | 1 | 0.219 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 1.95 [0.7345, 4.59] | 0.005581 | 0.05581 | 0.6857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.961 [-2.215, 0.2465] | 0.04844 | 0.2906 | -0.5048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 1.605 [-0.903, 3.618] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -1.683 [-3.764, 0.198] | 0.02395 | 0.1677 | -0.5714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.4535 [-1.94, 1.995] | 0.8408 | 1 | 0.05714 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -4.748 [-5.077, -2.007] | 3.815e-06 | 3.815e-05 | -0.9905 | 0 |
| hybrid-linucb | sed | 20 | -1.127 [-3.338, 0.3365] | 0.01362 | 0.1089 | -0.619 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -3.535 [-5.021, -2.176] | 0.0001335 | 0.001202 | -0.8857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -1.262 [-3.105, 0.1415] | 0.05317 | 0.319 | -0.4952 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.3905 [-1.732, 1.098] | 0.5958 | 1 | -0.1429 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 1.739 [-0.367, 4.398] | 0.08255 | 0.4128 | 0.4476 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.818 [-1.79, 0.291] | 0.2024 | 0.7574 | -0.3333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.028 [-0.1625, 3.97] | 0.1893 | 0.7574 | 0.3429 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -2.152 [-4.765, 0.5905] | 0.01718 | 0.1203 | -0.6 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.169 [-2.831, 2.019] | 0.8408 | 1 | -0.05714 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb | sed | 20 | -0.07203 [-0.07881, -0.0661] | 8.795e-05 | 0.0008128 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.07203 [-0.07966, -0.0661] | 8.782e-05 | 0.0008128 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.005085 [0.002542, 0.00678] | 8.128e-05 | 0.0008128 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.06525 [-0.07627, -0.06186] | 8.733e-05 | 0.0008128 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.1573 | 0.6292 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.07203 [-0.07966, -0.0661] | 8.757e-05 | 0.0008128 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | -0.07203 [-0.07966, -0.0661] | 8.782e-05 | 0.0008128 | -1 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb | sed | 20 | -0.1614 [-0.1718, -0.1446] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.1614 [-0.1718, -0.1446] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.009694 [0.00516, 0.01316] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.152 [-0.1602, -0.1408] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.5 | 1 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1607 [-0.1718, -0.1446] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | -0.1614 [-0.1718, -0.1446] | 1.907e-06 | 1.907e-05 | -1 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.03056 [0.003493, 0.04805] | 0.007296 | 0.02189 | 0.6667 | 0 |
| hybrid-linucb | sed | 20 | 0.1113 [0.1005, 0.1262] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.03061 [0.01801, 0.04632] | 0.0001335 | 0.0006676 | 0.8857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.112 [0.099, 0.1214] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.03937 [-0.05567, -0.02958] | 1.907e-05 | 0.0001144 | -0.9524 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03682 [0.0229, 0.04575] | 0.0002613 | 0.001045 | 0.8571 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.005554 [-0.01811, 0.009622] | 0.7012 | 0.7012 | -0.1048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07054 [0.06245, 0.08688] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01637 [0.005479, 0.02553] | 0.01208 | 0.02416 | 0.6286 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.09357 [0.08231, 0.1056] | 1.907e-06 | 1.907e-05 | 1 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |

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
| hybrid-linucb | leastloaded | 20 | 0.0161 [-0.01271, 0.05508] | 0.08353 | 0.6682 | 0.4526 | 1 |
| hybrid-linucb | sed | 20 | 0.0161 [-0.009322, 0.03983] | 0.3341 | 0.9578 | 0.2526 | 1 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.0161 [-0.01017, 0.04492] | 0.1004 | 0.7026 | 0.419 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.0178 [-0.004237, 0.03136] | 0.1402 | 0.7026 | 0.3762 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.05339 [0.02119, 0.06695] | 8.845e-05 | 0.0008845 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.04492 [0.03136, 0.06102] | 0.0003392 | 0.003053 | 0.9368 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.005085 [-0.01271, 0.02881] | 0.2394 | 0.9578 | 0.3 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.0008475 [-0.008475, 0.02373] | 0.3802 | 0.9578 | 0.2238 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.0161 [0.00339, 0.02373] | 0.1044 | 0.7026 | 0.4143 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.008475 [-0.0008475, 0.02203] | 0.2955 | 0.9578 | 0.2667 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.004237 [-0.0322, 0.0161] | 0.3759 | 1 | -0.2316 | 1 |
| hybrid-linucb | sed | 20 | -0.02458 [-0.05593, -0.001695] | 0.005096 | 0.03057 | -0.7143 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.005085 [-0.01017, 0.01864] | 0.7651 | 1 | 0.07619 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.001695 [-0.06356, 0.009322] | 0.1841 | 0.9207 | -0.3474 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.05593 [0.03983, 0.08475] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03644 [0.005932, 0.06525] | 0.001711 | 0.0154 | 0.8 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.03305 [0.01017, 0.06441] | 0.002899 | 0.02029 | 0.7789 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 0.007627 [-0.009322, 0.02458] | 0.4325 | 1 | 0.2053 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.02373 [0.01949, 0.05169] | 0.002493 | 0.01994 | 0.7714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.007627 [-0.01271, 0.02458] | 0.5196 | 1 | 0.1684 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| hybrid-linucb | leastloaded | 20 | 0.09425 [-0.1046, 0.2329] | 0.3124 | 1 | 0.2737 | 1 |
| hybrid-linucb | sed | 20 | 0.101 [-0.04828, 0.1833] | 0.1536 | 0.9219 | 0.3714 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.1327 [0.008046, 0.3331] | 0.06193 | 0.4335 | 0.4762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.1755 [0.014, 0.2793] | 0.0349 | 0.3141 | 0.5381 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.1327 [-0.01182, 0.2664] | 0.04844 | 0.3875 | 0.5048 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.131 [0.05296, 0.2162] | 0.02275 | 0.2275 | 0.581 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02738 [-0.1291, 0.1188] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.04491 [-0.1205, 0.1534] | 0.5256 | 1 | 0.1619 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.05755 [-0.08436, 0.1107] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01268 [-0.08436, 0.1674] | 0.3884 | 1 | 0.2286 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.0299 [-0.0721, 0.08906] | 0.8227 | 1 | 0.05714 | 0 |
| hybrid-linucb | sed | 20 | 0.01061 [-0.05908, 0.05508] | 0.9563 | 1 | 0.01905 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04497 [-0.0144, 0.1082] | 0.1231 | 1 | 0.4 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.06194 [-0.008199, 0.1124] | 0.1402 | 1 | 0.3762 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.03802 [-0.05334, 0.09235] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02156 [-0.02139, 0.08374] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.00334 [-0.06047, 0.0518] | 0.8983 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01329 [-0.08822, 0.05932] | 1 | 1 | 0 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.02157 [-0.03419, 0.08801] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01976 [-0.01764, 0.07151] | 0.4091 | 1 | 0.219 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.01753 [-0.01352, 0.03695] | 0.06733 | 0.6733 | 0.4667 | 0 |
| hybrid-linucb | sed | 20 | 0.01386 [-0.02534, 0.03041] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01044 [-0.02949, 0.01859] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -5.331e-05 [-0.02363, 0.006324] | 0.2305 | 1 | -0.3143 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01169 [-0.01433, 0.02546] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.003281 [-0.01801, 0.01515] | 1 | 1 | 0 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.005606 [-0.01677, 0.0206] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01105 [-0.02432, 0.01865] | 0.6215 | 1 | -0.1333 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.002189 [-0.01456, 0.0253] | 0.6215 | 1 | 0.1333 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.006288 [-0.0167, 0.007407] | 0.498 | 1 | -0.181 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.5 [0, 1] | 0.1756 | 1 | 0.381 | 6 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 0.5] | 0.7963 | 1 | 0.07692 | 8 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-1, 0.5] | 0.6736 | 1 | 0.1238 | 6 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [-1, 0] | 0.5074 | 1 | -0.2121 | 9 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.5 [-1, 1] | 0.8268 | 1 | -0.05882 | 3 |
| hybrid-linucb-d-g095 | sed | 20 | -0.5 [-1, 0.5] | 0.242 | 1 | -0.3162 | 4 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-1, 1] | 0.4144 | 1 | 0.2206 | 4 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0.5] | 0.9215 | 1 | 0.02857 | 6 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [-1, 1] | 0.5596 | 1 | 0.1569 | 3 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [-1, 0.5] | 0.8856 | 1 | -0.04396 | 7 |

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
| hybrid-linucb | leastloaded | 20 | 0.5 [0, 1] | 0.1756 | 1 | 0.381 | 6 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 0.5] | 0.7963 | 1 | 0.07692 | 8 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-1, 0.5] | 0.6736 | 1 | 0.1238 | 6 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [-1, 0] | 0.5074 | 1 | -0.2121 | 9 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.5 [-1, 1] | 0.8268 | 1 | -0.05882 | 3 |
| hybrid-linucb-d-g095 | sed | 20 | -0.5 [-1, 0.5] | 0.242 | 1 | -0.3162 | 4 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-1, 1] | 0.4144 | 1 | 0.2206 | 4 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0.5] | 0.9215 | 1 | 0.02857 | 6 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [-1, 1] | 0.5596 | 1 | 0.1569 | 3 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [-1, 0.5] | 0.8856 | 1 | -0.04396 | 7 |

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
| hybrid-linucb | leastloaded | 20 | -0.5 [-1, 0] | 0.1756 | 1 | -0.381 | 6 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 0.5] | 0.7963 | 1 | -0.07692 | 8 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-0.5, 1] | 0.6736 | 1 | -0.1238 | 6 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [0, 1] | 0.5074 | 1 | 0.2121 | 9 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.5 [-1, 1] | 0.8268 | 1 | 0.05882 | 3 |
| hybrid-linucb-d-g095 | sed | 20 | 0.5 [-0.5, 1] | 0.242 | 1 | 0.3162 | 4 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-1, 1] | 0.4144 | 1 | -0.2206 | 4 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-0.5, 1] | 0.9215 | 1 | -0.02857 | 6 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [-1, 1] | 0.5596 | 1 | -0.1569 | 3 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [-0.5, 1] | 0.8856 | 1 | 0.04396 | 7 |

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
| hybrid-linucb | leastloaded | 20 | 0.05 [0.01, 0.1] | 0.0328 | 0.2963 | 0.5579 | 1 |
| hybrid-linucb | sed | 20 | 0.03 [-0.02, 0.12] | 0.103 | 0.5913 | 0.4263 | 1 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.06 [-0.02, 0.1] | 0.06081 | 0.4257 | 0.5029 | 2 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.04 [0.01, 0.1] | 0.04819 | 0.3855 | 0.5158 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.08 [0.02, 0.11] | 0.02963 | 0.2963 | 0.5684 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02 [-0.03, 0.13] | 0.09855 | 0.5913 | 0.4316 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.03 [-0.03, 0.12] | 0.1505 | 0.5913 | 0.386 | 2 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07 [-0.01, 0.1] | 0.09872 | 0.5913 | 0.4316 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01 [-0.06, 0.07] | 0.6012 | 1 | 0.1404 | 2 |
| hybrid-linucb-d-g099 | sed | 20 | 0.03 [-0.04, 0.07] | 0.6146 | 1 | 0.1316 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.01734 [-0.01265, 0.02901] | 0.1327 | 1 | 0.3905 | 0 |
| hybrid-linucb | sed | 20 | 0.001931 [-0.03138, 0.02729] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.003845 [-0.03379, 0.0152] | 0.6477 | 1 | -0.1238 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.007538 [-0.02513, 0.0022] | 0.03277 | 0.3277 | -0.5429 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.004323 [-0.02848, 0.02469] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.008671 [-0.02456, 0.01094] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.007306 [-0.02635, 0.03429] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.00654 [-0.03605, 0.02211] | 0.4781 | 1 | -0.181 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.00044 [-0.0251, 0.0249] | 0.8408 | 1 | 0.05714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.00164 [-0.02389, 0.01024] | 0.6226 | 1 | -0.1368 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.1899 [0.05717, 0.2537] | 0.0003948 | 0.003159 | 0.8381 | 0 |
| hybrid-linucb | sed | 20 | 0.167 [0.07102, 0.2666] | 0.001432 | 0.01003 | 0.7714 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.2013 [0.1272, 0.2786] | 0.0002098 | 0.002098 | 0.8667 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.168 [0.07774, 0.3007] | 0.0003223 | 0.002901 | 0.8476 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.1414 [0.04847, 0.2461] | 0.03277 | 0.1332 | 0.5429 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.08498 [0.03685, 0.1474] | 0.02664 | 0.1332 | 0.5619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.07722 [-0.003652, 0.1886] | 0.04844 | 0.1332 | 0.5048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.06551 [-0.04163, 0.1983] | 0.1231 | 0.1332 | 0.4 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.1366 [0.01456, 0.244] | 0.01069 | 0.06413 | 0.6381 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.07275 [-0.02722, 0.2336] | 0.03999 | 0.1332 | 0.5238 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.05007 [0.02275, 0.06036] | 3.624e-05 | 0.0002537 | 0.9333 | 0 |
| hybrid-linucb | sed | 20 | 0.02054 [-6.362e-05, 0.04516] | 0.08255 | 0.08811 | 0.4476 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04192 [0.01547, 0.05368] | 0.0007076 | 0.003538 | 0.8095 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.00847 [-0.001734, 0.02489] | 0.04405 | 0.08811 | 0.5143 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.09582 [0.0733, 0.1194] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.07674 [0.05937, 0.08887] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.04651 [0.02827, 0.06906] | 5.722e-06 | 4.578e-05 | 0.981 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02859 [0.009166, 0.04282] | 0.001017 | 0.004066 | 0.7905 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.03921 [0.02539, 0.04438] | 6.294e-05 | 0.0003777 | 0.9143 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01682 [0.001728, 0.02965] | 0.00486 | 0.01458 | 0.6952 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 9 [-2, 34] | 0.09774 | 0.585 | 0.4444 | 2 |
| hybrid-linucb | sed | 20 | 25 [7, 35] | 0.002642 | 0.02642 | 0.7667 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 10.5 [-4, 26] | 0.327 | 0.8901 | 0.2632 | 2 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 10.5 [1.5, 30.5] | 0.01978 | 0.1385 | 0.6257 | 2 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 11 [-1, 37.5] | 0.09751 | 0.585 | 0.4575 | 3 |
| hybrid-linucb-d-g095 | sed | 20 | 21 [1.5, 38.5] | 0.009983 | 0.07987 | 0.6571 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 9.5 [-3, 29] | 0.2461 | 0.8901 | 0.3203 | 3 |
| hybrid-linucb-d-g098 | sed | 20 | 17.5 [2.5, 31.5] | 0.00352 | 0.03168 | 0.7836 | 2 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.5 [-12, 24] | 0.619 | 0.8901 | 0.1373 | 3 |
| hybrid-linucb-d-g099 | sed | 20 | 5.5 [-3.5, 18.5] | 0.2225 | 0.8901 | 0.3275 | 2 |

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
| hybrid-linucb | leastloaded | 20 | 0.01753 [-0.01352, 0.03695] | 0.06733 | 0.6733 | 0.4667 | 0 |
| hybrid-linucb | sed | 20 | 0.01386 [-0.02534, 0.03041] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01044 [-0.02949, 0.01859] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -5.331e-05 [-0.02363, 0.006324] | 0.2305 | 1 | -0.3143 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01169 [-0.01433, 0.02546] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.003281 [-0.01801, 0.01515] | 1 | 1 | 0 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.005606 [-0.01677, 0.0206] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01105 [-0.02432, 0.01865] | 0.6215 | 1 | -0.1333 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.002189 [-0.01456, 0.0253] | 0.6215 | 1 | 0.1333 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.006288 [-0.0167, 0.007407] | 0.498 | 1 | -0.181 | 0 |

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
