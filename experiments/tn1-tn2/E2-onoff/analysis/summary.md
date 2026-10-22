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
| hybrid-linucb | leastloaded | 20 | -1.99 [-5.186, -0.988] | 0.0007076 | 0.005661 | -0.8095 | 0 |
| hybrid-linucb | sed | 20 | 1.26 [-0.922, 2.835] | 0.1769 | 0.3537 | 0.3524 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -3.846 [-6.146, -1.788] | 0.0001335 | 0.001202 | -0.8857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 1.417 [-0.8625, 2.185] | 0.2455 | 0.3537 | 0.3048 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 3.304 [-0.4205, 3.689] | 0.06958 | 0.233 | 0.4667 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 6.113 [3.759, 8.359] | 9.537e-06 | 9.537e-05 | 0.9714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -1.725 [-2.624, 0.158] | 0.05826 | 0.233 | -0.4857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.595 [0.553, 4.907] | 0.0008507 | 0.005955 | 0.8 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -2.917 [-4.667, -0.2775] | 0.01208 | 0.07248 | -0.6286 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 1.285 [-0.2035, 4.437] | 0.02664 | 0.1332 | 0.5619 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.15 [-1.589, 0.396] | 0.4524 | 0.9048 | -0.2 | 0 |
| hybrid-linucb | sed | 20 | 1.139 [0.2465, 1.749] | 0.04844 | 0.3084 | 0.5048 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -1.88 [-2.866, 0.1215] | 0.008308 | 0.06647 | -0.6571 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.585 [-0.476, 1.281] | 0.1893 | 0.7574 | 0.3429 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.358 [-0.1635, 1.486] | 0.2305 | 0.7574 | 0.3143 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 2.33 [0.7415, 4.364] | 0.0001335 | 0.001335 | 0.8857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.3385 [-1.7, 1.226] | 0.9854 | 0.9854 | -0.009524 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 1.595 [0.269, 4.04] | 0.004221 | 0.03799 | 0.7048 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.7035 [-2.342, 0.0235] | 0.09731 | 0.4865 | -0.4286 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 1.051 [-0.121, 1.623] | 0.04405 | 0.3084 | 0.5143 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -2.535 [-4.426, -1.518] | 0.001432 | 0.01146 | -0.7714 | 0 |
| hybrid-linucb | sed | 20 | -0.1935 [-1.114, 0.669] | 0.8695 | 0.8695 | -0.04762 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -1.615 [-3.465, -0.6195] | 0.0003948 | 0.003553 | -0.8381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.437 [-0.511, 1.129] | 0.4091 | 0.8182 | 0.219 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 1.01 [-0.6175, 2.846] | 0.1893 | 0.568 | 0.3429 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 3.475 [1.42, 5.31] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -1.449 [-2.877, -0.894] | 0.01718 | 0.1203 | -0.6 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.6235 [-0.0175, 1.284] | 0.03277 | 0.1966 | 0.5429 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -1.295 [-3.205, -0.3] | 0.05826 | 0.2913 | -0.4857 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.4505 [-0.2455, 2.828] | 0.09731 | 0.3892 | 0.4286 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.07627 [-0.0822, -0.0678] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.07627 [-0.0822, -0.0678] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.005085 [0.002542, 0.005932] | 0.0001817 | 0.0009087 | 1 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | -0.07373 [-0.07797, -0.06441] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.1573 | 0.6292 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.07627 [-0.0822, -0.0678] | 8.82e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | -0.07627 [-0.0822, -0.0678] | 8.807e-05 | 0.0008807 | -1 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.1648 [-0.1718, -0.1461] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.1648 [-0.1718, -0.1461] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01014 [0.005371, 0.0122] | 7.629e-06 | 3.815e-05 | 1 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | -0.1546 [-0.1701, -0.1391] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.5 | 1 | 1 | 18 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1648 [-0.1718, -0.1446] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | -0.1648 [-0.1718, -0.1461] | 1.907e-06 | 1.907e-05 | -1 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.03281 [0.02303, 0.04917] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb | sed | 20 | 0.1104 [0.09445, 0.1194] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04439 [0.02679, 0.04949] | 5.722e-06 | 2.861e-05 | 0.981 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.1088 [0.0885, 0.1323] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.02535 [-0.04628, -0.01967] | 4.768e-05 | 0.0001907 | -0.9238 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03979 [0.01812, 0.06471] | 0.0001678 | 0.0005035 | 0.8762 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.00299 [-0.01001, 0.02063] | 0.498 | 0.498 | 0.181 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07981 [0.05199, 0.08994] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.02029 [0.008446, 0.03519] | 0.0008507 | 0.001701 | 0.8 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.09379 [0.0753, 0.1117] | 1.907e-06 | 1.907e-05 | 1 | 0 |

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
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

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
| hybrid-linucb | leastloaded | 20 | 0.01525 [-0.001695, 0.04068] | 0.07615 | 0.4036 | 0.4524 | 0 |
| hybrid-linucb | sed | 20 | 0.02542 [0.007627, 0.04322] | 0.04377 | 0.3501 | 0.5143 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.0322 [-0.0178, 0.05169] | 0.1305 | 0.4036 | 0.3857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.0161 [0.00339, 0.03559] | 0.06726 | 0.4036 | 0.4667 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02627 [0.0161, 0.03898] | 0.002712 | 0.02712 | 0.7333 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.04068 [0.008475, 0.05593] | 0.006792 | 0.06113 | 0.6905 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.0161 [-0.005085, 0.05593] | 0.05448 | 0.3814 | 0.4905 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.03475 [-0.005932, 0.05339] | 0.08255 | 0.4036 | 0.4476 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.00678 [-0.009322, 0.04322] | 0.1956 | 0.4036 | 0.3474 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02288 [-0.02373, 0.04153] | 0.3133 | 0.4036 | 0.2571 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.02288 [-0.03983, 0.002542] | 0.1031 | 0.5154 | -0.4263 | 1 |
| hybrid-linucb | sed | 20 | -0.04407 [-0.06102, -0.02966] | 0.0001748 | 0.001573 | -0.9571 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.01695 [-0.03814, 0.01102] | 0.2471 | 0.8431 | -0.2952 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.04322 [-0.06525, -0.0178] | 0.0008507 | 0.006805 | -0.8 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.03983 [0.0339, 0.06186] | 1.144e-05 | 0.0001144 | 0.9789 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 0.01441 [-0.009322, 0.03051] | 0.2108 | 0.8431 | 0.319 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.00339 [-0.02627, 0.0339] | 0.4897 | 0.9794 | 0.1762 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.02627 [-0.05085, 0.008475] | 0.06726 | 0.4036 | -0.4667 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01102 [-0.01525, 0.02797] | 0.4897 | 0.9794 | 0.1762 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.02373 [-0.04153, -0.01356] | 0.00895 | 0.06265 | -0.6667 | 0 |

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
| hybrid-linucb | 0.2069 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2629 | 20 |
| hybrid-linucb-d-g095 | 0.2831 | 20 |
| hybrid-linucb-d-g098 | 0.3793 | 20 |
| hybrid-linucb-d-g099 | 0.2586 | 20 |
| leastloaded | 0.2707 | 20 |
| sed | 0.2457 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.00977 [-0.09762, 0.117] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb | sed | 20 | -0.04199 [-0.1055, 0.09515] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04803 [-0.1477, 0.1484] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.005172 [-0.1068, 0.2414] | 0.5196 | 1 | 0.1684 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.006158 [-0.1554, 0.1948] | 0.8248 | 1 | 0.05789 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 0.05057 [-0.05484, 0.233] | 0.444 | 1 | 0.1952 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.1466 [-2.776e-17, 0.2414] | 0.165 | 1 | 0.3619 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.1552 [0.01314, 0.2276] | 0.06958 | 0.6958 | 0.4667 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.04167 [-0.09163, 0.05714] | 0.7562 | 1 | -0.08571 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.005296 [-0.1349, 0.1207] | 0.9273 | 1 | 0.02857 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.007752 [-0.05689, 0.05539] | 0.9851 | 1 | 0.004762 | 0 |
| hybrid-linucb | sed | 20 | -0.04438 [-0.09421, 0.01938] | 0.1327 | 1 | -0.3905 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04341 [-0.003725, 0.08173] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01544 [-0.04844, 0.05178] | 0.9851 | 1 | -0.004762 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01398 [-0.05704, 0.08369] | 0.4781 | 1 | 0.181 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.01995 [-0.07047, 0.05294] | 0.7368 | 1 | -0.08571 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.03636 [-0.03612, 0.1053] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01346 [-0.06455, 0.04136] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.0043 [-0.02885, 0.02891] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.03977 [-0.1167, 0.01245] | 0.1327 | 1 | -0.3905 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.01358 [-0.004338, 0.04501] | 0.1084 | 0.7176 | 0.4095 | 0 |
| hybrid-linucb | sed | 20 | 0.03365 [-0.005675, 0.04755] | 0.04003 | 0.4003 | 0.5238 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01302 [-0.01548, 0.02823] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01122 [-0.01046, 0.02606] | 0.2753 | 1 | 0.2947 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.001894 [-0.0288, 0.02143] | 0.546 | 1 | -0.1579 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -0.00189 [-0.02174, 0.0179] | 0.9702 | 1 | 0.009524 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02005 [-0.02252, 0.03061] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01739 [-0.004619, 0.04283] | 0.09731 | 0.7176 | 0.4286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01655 [0.001085, 0.03682] | 0.08969 | 0.7176 | 0.4381 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.03176 [-0.0108, 0.04746] | 0.06958 | 0.6262 | 0.4667 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.4314 | 1 | 0.2889 | 11 |
| hybrid-linucb | sed | 20 | 0 [-1, 0] | 0.512 | 1 | -0.1905 | 6 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-0.5, 1] | 0.6912 | 1 | 0.1209 | 7 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [-1, 0] | 0.2218 | 1 | -0.3846 | 8 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [0, 1] | 0.08319 | 0.7488 | 0.5165 | 7 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 0.763 | 1 | 0.1111 | 12 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 1] | 0.6579 | 1 | 0.1319 | 7 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0] | 0.1841 | 1 | -0.3956 | 7 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [-1, 0] | 0.3687 | 1 | -0.2747 | 7 |
| hybrid-linucb-d-g099 | sed | 20 | -1 [-1, 0] | 0.01098 | 0.1098 | -0.7582 | 7 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.4314 | 1 | 0.2889 | 11 |
| hybrid-linucb | sed | 20 | 0 [-1, 0] | 0.512 | 1 | -0.1905 | 6 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-0.5, 1] | 0.6912 | 1 | 0.1209 | 7 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [-1, 0] | 0.2218 | 1 | -0.3846 | 8 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [0, 1] | 0.08319 | 0.7488 | 0.5165 | 7 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 0.763 | 1 | 0.1111 | 12 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 1] | 0.6579 | 1 | 0.1319 | 7 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0] | 0.1841 | 1 | -0.3956 | 7 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [-1, 0] | 0.3687 | 1 | -0.2747 | 7 |
| hybrid-linucb-d-g099 | sed | 20 | -1 [-1, 0] | 0.01098 | 0.1098 | -0.7582 | 7 |

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
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 0.4314 | 1 | -0.2889 | 11 |
| hybrid-linucb | sed | 20 | 0 [0, 1] | 0.512 | 1 | 0.1905 | 6 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [-1, 0.5] | 0.6912 | 1 | -0.1209 | 7 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0 [0, 1] | 0.2218 | 1 | 0.3846 | 8 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [-1, 0] | 0.08319 | 0.7488 | -0.5165 | 7 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [0, 0] | 0.763 | 1 | -0.1111 | 12 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-1, 0.5] | 0.6579 | 1 | -0.1319 | 7 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 1] | 0.1841 | 1 | 0.3956 | 7 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 1] | 0.3687 | 1 | 0.2747 | 7 |
| hybrid-linucb-d-g099 | sed | 20 | 1 [0, 1] | 0.01098 | 0.1098 | 0.7582 | 7 |

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
| hybrid-linucb | leastloaded | 20 | -0.01 [-0.09, 0.06] | 0.5325 | 1 | -0.1632 | 1 |
| hybrid-linucb | sed | 20 | 0.06 [-0.01, 0.09] | 0.3314 | 1 | 0.268 | 3 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.05 [-0.04, 0.08] | 0.4664 | 1 | 0.1857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.1 [0.04, 0.14] | 0.06426 | 0.5921 | 0.4714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.02 [-0.08, 0.12] | 0.3754 | 1 | 0.2316 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 0.08 [0.02, 0.15] | 0.05921 | 0.5921 | 0.481 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.03 [-0.03, 0.09] | 0.7936 | 1 | 0.06842 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07 [0.02, 0.12] | 0.1083 | 0.8664 | 0.4095 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.04 [-0.08, 0.03] | 0.2706 | 1 | -0.281 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02 [-0.03, 0.08] | 0.3542 | 1 | 0.2421 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.01797 [-0.0009537, 0.04936] | 0.05317 | 0.5317 | 0.4952 | 0 |
| hybrid-linucb | sed | 20 | 0.0259 [-0.0186, 0.05061] | 0.07585 | 0.6068 | 0.4571 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01247 [-0.004866, 0.02288] | 0.4326 | 1 | 0.2053 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.0002848 [-0.02098, 0.02644] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.000905 [-0.025, 0.01799] | 0.5256 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.006989 [-0.03045, 0.006399] | 0.4749 | 1 | -0.1905 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01794 [-0.018, 0.04077] | 0.1893 | 1 | 0.3429 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.006743 [-0.01553, 0.03285] | 0.3683 | 1 | 0.2381 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.02489 [0.005264, 0.04219] | 0.06372 | 0.5735 | 0.4762 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02802 [-0.01276, 0.04927] | 0.09731 | 0.6812 | 0.4286 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.1524 [0.08732, 0.2151] | 0.0004826 | 0.004826 | 0.8286 | 0 |
| hybrid-linucb | sed | 20 | 0.0859 [0.0323, 0.2531] | 0.01531 | 0.1072 | 0.6095 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.09256 [-0.03398, 0.2142] | 0.06372 | 0.319 | 0.4762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.09207 [-0.01606, 0.2999] | 0.09731 | 0.3302 | 0.4286 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.03712 [-0.01679, 0.1639] | 0.1327 | 0.3302 | 0.3905 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02618 [-0.04985, 0.203] | 0.2774 | 0.3302 | 0.2857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.1823 [0.09315, 0.2531] | 0.0004826 | 0.004826 | 0.8286 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.2117 [0.03377, 0.2858] | 0.004221 | 0.03377 | 0.7048 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.08335 [-0.0462, 0.1297] | 0.05317 | 0.319 | 0.4952 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.1069 [-0.06995, 0.176] | 0.08255 | 0.3302 | 0.4476 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.05305 [0.0252, 0.08093] | 4.768e-05 | 0.0003338 | 0.9238 | 0 |
| hybrid-linucb | sed | 20 | 0.03919 [0.01031, 0.06072] | 0.001209 | 0.002419 | 0.781 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.05366 [0.04132, 0.08128] | 0.0004826 | 0.00193 | 0.8286 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.03118 [0.01561, 0.04404] | 0.0005856 | 0.00193 | 0.819 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.0686 [0.0581, 0.08315] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.05659 [0.03906, 0.08431] | 5.722e-06 | 4.578e-05 | 0.981 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.06075 [0.03257, 0.07836] | 3.815e-06 | 3.433e-05 | 0.9905 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.03558 [0.01831, 0.06339] | 0.0001678 | 0.0008392 | 0.8762 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.05728 [0.03765, 0.07997] | 4.768e-05 | 0.0003338 | 0.9238 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.04546 [0.01512, 0.06211] | 0.001986 | 0.002419 | 0.7524 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 21.5 [-46.5, 54] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb | sed | 20 | 37.5 [0, 111] | 0.004094 | 0.04094 | 0.8162 | 4 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -10.5 [-52, 27.5] | 0.5678 | 1 | -0.1579 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 3.5 [0, 73] | 0.013 | 0.117 | 0.7059 | 4 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -11.5 [-72.5, 19.5] | 0.2396 | 1 | -0.3 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 8.5 [0, 51] | 0.01304 | 0.117 | 0.6667 | 2 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 1.5 [-65, 16.5] | 0.4204 | 1 | -0.2164 | 2 |
| hybrid-linucb-d-g098 | sed | 20 | 12.5 [0, 35.5] | 0.03354 | 0.2012 | 0.6029 | 4 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -14 [-62, 3] | 0.2145 | 1 | -0.3333 | 2 |
| hybrid-linucb-d-g099 | sed | 20 | 9 [0, 32] | 0.01675 | 0.1172 | 0.6601 | 3 |

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
| hybrid-linucb | leastloaded | 20 | 0.002042 [-0.02607, 0.03969] | 0.6215 | 1 | 0.1333 | 0 |
| hybrid-linucb | sed | 20 | 0.02839 [0.0004397, 0.04534] | 0.02958 | 0.2958 | 0.5524 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.001661 [-0.01756, 0.02674] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01414 [0.002205, 0.02611] | 0.1054 | 0.7378 | 0.419 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01642 [-0.02116, 0.04103] | 0.5217 | 1 | 0.1714 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02492 [-0.00885, 0.04408] | 0.05826 | 0.5243 | 0.4857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02102 [-0.037, 0.05095] | 0.4524 | 1 | 0.2 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01799 [-0.005907, 0.0663] | 0.06372 | 0.5243 | 0.4762 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.006215 [-0.02246, 0.03729] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.009853 [-0.006114, 0.06176] | 0.1327 | 0.7964 | 0.3905 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 7.5 [2, 12.5] | 0.02161 | 0.2161 | 0.5857 | 0 |
| hybrid-linucb | sed | 20 | -3 [-7, 6] | 0.8811 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 10 [2.5, 15.5] | 0.02385 | 0.2161 | 0.5762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 2.5 [-3, 5.5] | 0.6316 | 1 | 0.1287 | 2 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 4.5 [-1, 13] | 0.1363 | 0.8178 | 0.3895 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -1.5 [-11, 2.5] | 0.2954 | 1 | -0.2667 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 5 [0.5, 11] | 0.03793 | 0.3034 | 0.5421 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -1.5 [-7, 5.5] | 0.4089 | 1 | -0.2158 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 5 [1.5, 9] | 0.06403 | 0.4482 | 0.4714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -3.5 [-8, 3] | 0.3409 | 1 | -0.2429 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 2.5 [-5.5, 9.5] | 0.3979 | 1 | 0.2211 | 1 |
| hybrid-linucb | sed | 20 | 6 [0, 10] | 0.03102 | 0.3102 | 0.5789 | 2 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 1.5 [-4.5, 8] | 0.478 | 1 | 0.181 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 3.5 [0.5, 7] | 0.1526 | 0.9155 | 0.3737 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 2.5 [-6, 12] | 0.322 | 1 | 0.2524 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 6.5 [-1, 10] | 0.074 | 0.666 | 0.4795 | 2 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 6 [-7.5, 12] | 0.3039 | 1 | 0.2619 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 5.5 [-1, 14] | 0.07647 | 0.666 | 0.4632 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 3 [-7.5, 9.5] | 0.3698 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 3 [-2.5, 11.5] | 0.1071 | 0.7498 | 0.4211 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.0262 [-0.008834, 0.05229] | 0.05826 | 0.5243 | 0.4857 | 0 |
| hybrid-linucb | sed | 20 | 0.02417 [-0.01598, 0.06753] | 0.1054 | 0.7378 | 0.419 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01855 [0.006398, 0.03061] | 0.2162 | 0.9467 | 0.3238 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.02527 [-0.02944, 0.03952] | 0.5958 | 0.9467 | 0.1429 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.02073 [-0.03719, -0.0003785] | 0.1054 | 0.7378 | -0.419 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.02425 [-0.05003, 0.00598] | 0.1893 | 0.9467 | -0.3429 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01613 [-0.01831, 0.03064] | 0.2774 | 0.9467 | 0.2857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01285 [-0.01176, 0.03581] | 0.3884 | 0.9467 | 0.2286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.03004 [0.003513, 0.04872] | 0.02395 | 0.2395 | 0.5714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02142 [0.001171, 0.0555] | 0.07585 | 0.6068 | 0.4571 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -7.5 [-13, -1.5] | 0.01856 | 0.1856 | -0.6 | 0 |
| hybrid-linucb | sed | 20 | 2.5 [-5, 7.5] | 0.8403 | 1 | 0.05263 | 1 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -10 [-14.5, -2.5] | 0.02056 | 0.1856 | -0.5905 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -1 [-5, 4] | 0.6576 | 1 | -0.1158 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -4.5 [-14, -1] | 0.08568 | 0.532 | -0.4381 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 1.5 [-3, 11.5] | 0.3313 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -3.5 [-11, -2] | 0.03978 | 0.3183 | -0.5238 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2 [-5, 8] | 0.3902 | 1 | 0.219 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -5.5 [-9.5, 0] | 0.076 | 0.532 | -0.4524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 3.5 [-2.5, 9.5] | 0.2465 | 1 | 0.2952 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 3 [-4, 7] | 0.2511 | 1 | 0.3 | 1 |
| hybrid-linucb | sed | 20 | 6.5 [-7.5, 14.5] | 0.1401 | 0.9807 | 0.3762 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 2.5 [-2, 6] | 0.7471 | 1 | 0.08421 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 5 [-6.5, 7] | 0.7225 | 1 | 0.09048 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -3 [-12.5, 1] | 0.05897 | 0.5793 | -0.481 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -3.5 [-9, 0.5] | 0.2394 | 1 | -0.3 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.5 [-6.5, 7] | 0.6947 | 1 | 0.1 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.5 [-2.5, 9] | 0.277 | 1 | 0.2842 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 5 [-1, 11] | 0.05793 | 0.5793 | 0.5088 | 2 |
| hybrid-linucb-d-g099 | sed | 20 | 7.5 [-2.5, 14] | 0.07909 | 0.6327 | 0.4476 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.003894 [-0.0694, 0.05787] | 0.7562 | 1 | -0.08571 | 0 |
| hybrid-linucb | sed | 20 | 0.05133 [-0.001933, 0.1042] | 0.1769 | 1 | 0.3524 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.05317 [-0.05613, 0.08021] | 0.4304 | 1 | 0.2095 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.1012 [0.0678, 0.1367] | 0.05317 | 0.4785 | 0.4952 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.001801 [-0.07185, 0.1366] | 0.3683 | 1 | 0.2381 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.09127 [0.02302, 0.1756] | 0.03623 | 0.3623 | 0.5333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.04418 [-0.048, 0.1276] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.09503 [0.02626, 0.166] | 0.07585 | 0.6068 | 0.4571 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.03912 [-0.1089, 0.04441] | 0.3488 | 1 | -0.2476 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.03966 [-0.02512, 0.07181] | 0.2455 | 1 | 0.3048 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.001753 [-0.035, 0.02146] | 0.9843 | 1 | -0.01053 | 1 |
| hybrid-linucb | sed | 20 | -0.01762 [-0.06908, 0.0598] | 0.6477 | 1 | -0.1238 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.003208 [-0.06639, 0.0561] | 0.5459 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.01315 [-0.05218, 0.02393] | 0.3488 | 1 | -0.2476 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.01588 [-0.06484, 0.03419] | 0.4091 | 1 | -0.219 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.01985 [-0.0949, 0.03443] | 0.2162 | 1 | -0.3238 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.03796 [-0.06221, 0.01237] | 0.02582 | 0.2395 | -0.5789 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -0.04067 [-0.09056, -0.004657] | 0.02395 | 0.2395 | -0.5714 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.007304 [-0.02306, 0.06671] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02285 [-0.03993, 0.0805] | 0.8596 | 1 | 0.05263 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.007679 [-0.03618, 0.06159] | 0.4524 | 1 | 0.2 | 0 |
| hybrid-linucb | sed | 20 | -0.02487 [-0.04647, 0.03434] | 0.5459 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.008174 [-0.00656, 0.07454] | 0.4749 | 1 | 0.1905 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.007987 [-0.0659, 0.02704] | 0.4653 | 1 | -0.2 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01171 [-0.05108, 0.03711] | 0.7841 | 1 | -0.07619 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.03566 [-0.05704, -0.01454] | 0.03999 | 0.3999 | -0.5238 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.0406 [-0.01935, 0.06348] | 0.33 | 1 | 0.2571 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.004169 [-0.04963, 0.0414] | 0.5958 | 1 | -0.1429 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.005032 [-0.03895, 0.06927] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.006143 [-0.07633, 0.03502] | 0.3683 | 1 | -0.2381 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.05447 [0.004459, 0.08428] | 0.03999 | 0.3599 | 0.5238 | 0 |
| hybrid-linucb | sed | 20 | 0.0008502 [-0.03143, 0.08533] | 0.3702 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.02327 [0.01311, 0.08472] | 0.02395 | 0.2395 | 0.5714 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01904 [-0.01217, 0.03523] | 0.2611 | 1 | 0.2952 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02152 [-0.01974, 0.05978] | 0.1769 | 1 | 0.3524 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.01001 [-0.04559, 0.04165] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01991 [-0.03104, 0.0674] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.003178 [-0.06855, 0.05043] | 0.9843 | 1 | 0.01053 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.0209 [-0.02963, 0.06792] | 0.2611 | 1 | 0.2952 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01402 [-0.02701, 0.03776] | 0.7562 | 1 | 0.08571 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.003952 [-0.08602, 0.1187] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb | sed | 20 | 0.06175 [0.02101, 0.09375] | 0.0793 | 0.7137 | 0.4476 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04393 [-0.0475, 0.1217] | 0.2455 | 1 | 0.3048 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.06351 [0.004631, 0.129] | 0.02395 | 0.2395 | 0.5714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.04328 [-0.04853, 0.06454] | 0.6012 | 1 | 0.1333 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.04531 [-0.01379, 0.1074] | 0.1134 | 0.7936 | 0.4211 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.02819 [-0.07983, 0.04503] | 0.5628 | 1 | -0.1476 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.03212 [-0.04321, 0.09612] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.02878 [-0.0704, 0.09177] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.05252 [0.02583, 0.1275] | 0.08255 | 0.7137 | 0.4476 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.0893 [-0.12, 0.1787] | 0.498 | 1 | 0.181 | 0 |
| hybrid-linucb | sed | 20 | 0.06794 [-0.05691, 0.209] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.0105 [-0.09101, 0.1073] | 0.953 | 1 | -0.02105 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.02034 [-0.1057, 0.1063] | 0.8983 | 1 | 0.0381 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.04295 [-0.09465, -0.00662] | 0.2455 | 1 | -0.3048 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.07395 [-0.1623, 0.04067] | 0.2455 | 1 | -0.3048 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.04014 [-0.07901, 0.1737] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07985 [-0.02283, 0.1516] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.03852 [0.005254, 0.1721] | 0.1536 | 1 | 0.3714 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.08836 [-0.05262, 0.1567] | 0.08254 | 0.8254 | 0.4429 | 0 |

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
| hybrid-linucb | leastloaded | 20 | 0.04604 [-0.01775, 0.1057] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb | sed | 20 | 0.03352 [-0.09714, 0.1839] | 0.3506 | 1 | 0.2381 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.02777 [-0.1402, 0.0318] | 0.1536 | 1 | -0.3714 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.04971 [-0.09061, 0.06741] | 0.4749 | 1 | -0.1905 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.005894 [-0.1007, 0.03877] | 0.4524 | 1 | -0.2 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.01035 [-0.05915, 0.1139] | 0.7841 | 1 | 0.07619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.04293 [-0.0058, 0.0952] | 0.2024 | 1 | 0.3333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02434 [-0.02384, 0.1462] | 0.1769 | 1 | 0.3524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.006056 [-0.09228, 0.07542] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.04442 [-0.03419, 0.1234] | 0.33 | 1 | 0.2571 | 0 |

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
| hybrid-linucb | leastloaded | 20 | -0.01782 [-0.05286, 0.04393] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb | sed | 20 | 0.02163 [-0.03689, 0.06848] | 0.3118 | 1 | 0.2667 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.002825 [-0.1047, 0.07584] | 0.6226 | 1 | -0.1368 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.005914 [-0.04537, 0.05203] | 0.9854 | 1 | 0.009524 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.04524 [-0.1379, 0.03153] | 0.1231 | 1 | -0.4 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.03354 [-0.1392, 0.02799] | 0.165 | 1 | -0.3619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02021 [-0.03404, 0.09079] | 0.4091 | 1 | 0.219 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.002759 [-0.03378, 0.0718] | 0.6215 | 1 | 0.1333 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.02769 [-0.04531, 0.04457] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01455 [-0.05495, 0.07027] | 0.6226 | 1 | 0.1368 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.01129 [-0.02325, 0.03498] | 0.4653 | 1 | 0.2 | 1 |
| hybrid-linucb | sed | 20 | -0.00922 [-0.02825, 0.04593] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.008653 [-0.03584, 0.05793] | 0.5459 | 1 | 0.1619 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.001754 [-0.02877, 0.05759] | 0.6274 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.03069 [0.001625, 0.06586] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02972 [0.002909, 0.08446] | 0.08969 | 0.8969 | 0.4381 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.01667 [-0.06381, 0.05978] | 0.8983 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01275 [-0.04926, 0.04835] | 0.7285 | 1 | -0.09524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.0145 [-0.02404, 0.05127] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01138 [-0.02693, 0.04135] | 0.49 | 1 | 0.1895 | 1 |

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
| hybrid-linucb | leastloaded | 20 | 0.03041 [-0.003793, 0.07829] | 0.2101 | 1 | 0.3368 | 1 |
| hybrid-linucb | sed | 20 | 0.03654 [-0.02102, 0.08048] | 0.1231 | 0.9847 | 0.4 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.05093 [-0.026, 0.1244] | 0.1536 | 1 | 0.3714 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.02594 [0.01063, 0.1351] | 0.04844 | 0.4844 | 0.5048 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.006967 [-0.01712, 0.06619] | 0.4524 | 1 | 0.2 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03081 [0.008489, 0.07999] | 0.08255 | 0.743 | 0.4476 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02452 [-0.04359, 0.07967] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02755 [-0.02836, 0.07982] | 0.418 | 1 | 0.2211 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01062 [-0.03875, 0.08169] | 0.4091 | 1 | 0.219 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.04695 [-0.0318, 0.08174] | 0.2162 | 1 | 0.3238 | 0 |

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
| hybrid-linucb-d-g095 | leastloaded | 3 | -0.2 [-1, 0] | 0.5 | 1 | -1 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 1 | 0.8 [0.8, 0.8] | 1 | 1 | 1 | 0 |

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
