# Experiment analysis

Complete blocks: 20.
Time-to-adapt: W=20, reference=pre_all share, threshold=0.5 x pre_all share; reference share < 0.05 is NA. Censored values display as > n and enter paired tests as n_post + 1.
target_after_recovery_* is defined only for exactly one drift_on and one non-cleanup drift_off (transient); NA for onoff and permanent by design.
On/off aggregate shares pool decisions inside each phase after first drift_on; cell_end cleanup is excluded. Counts are decision and target-decision totals.
Holm family: every treatment x baseline pair for one metric in this directory; gamma/lambda sweep arms enlarge the family.

## Phases crossing build boundaries

| Round | Arm | Phases |
|---:|---|---|
| 1 | hybrid-linucb | 2_off |
| 1 | hybrid-linucb-d-arm-g098 | 2_off |
| 1 | hybrid-linucb-d-g095 | 2_off |
| 1 | hybrid-linucb-d-g098 | 2_off |
| 1 | hybrid-linucb-d-g099 | 2_off |
| 1 | leastloaded | 2_off |
| 1 | sed | 2_off |
| 2 | hybrid-linucb | 2_off |
| 2 | hybrid-linucb-d-arm-g098 | 2_off |
| 2 | hybrid-linucb-d-g095 | 2_off |
| 2 | hybrid-linucb-d-g098 | 2_off |
| 2 | hybrid-linucb-d-g099 | 2_off |
| 2 | leastloaded | 2_off |
| 2 | sed | 2_off |
| 3 | hybrid-linucb | 2_off |
| 3 | hybrid-linucb-d-arm-g098 | 2_off |
| 3 | hybrid-linucb-d-g095 | 2_off |
| 3 | hybrid-linucb-d-g098 | 2_off |
| 3 | hybrid-linucb-d-g099 | 2_off |
| 3 | leastloaded | 2_off |
| 3 | sed | 2_off |
| 4 | hybrid-linucb | 2_off |
| 4 | hybrid-linucb-d-arm-g098 | 2_off |
| 4 | hybrid-linucb-d-g095 | 2_off |
| 4 | hybrid-linucb-d-g098 | 2_off |
| 4 | hybrid-linucb-d-g099 | 2_off |
| 4 | leastloaded | 2_off |
| 4 | sed | 2_off |
| 5 | hybrid-linucb | 2_off |
| 5 | hybrid-linucb-d-arm-g098 | 2_off |
| 5 | hybrid-linucb-d-g095 | 2_off |
| 5 | hybrid-linucb-d-g098 | 2_off |
| 5 | hybrid-linucb-d-g099 | 2_off |
| 5 | leastloaded | 2_off |
| 5 | sed | 2_off |
| 6 | hybrid-linucb | 2_off |
| 6 | hybrid-linucb-d-arm-g098 | 2_off |
| 6 | hybrid-linucb-d-g095 | 2_off |
| 6 | hybrid-linucb-d-g098 | 2_off |
| 6 | hybrid-linucb-d-g099 | 2_off |
| 6 | leastloaded | 2_off |
| 6 | sed | 2_off |
| 7 | hybrid-linucb | 2_off |
| 7 | hybrid-linucb-d-arm-g098 | 2_off |
| 7 | hybrid-linucb-d-g095 | 2_off |
| 7 | hybrid-linucb-d-g098 | 2_off |
| 7 | hybrid-linucb-d-g099 | 2_off |
| 7 | leastloaded | 2_off |
| 7 | sed | 2_off |
| 8 | hybrid-linucb | 2_off |
| 8 | hybrid-linucb-d-arm-g098 | 2_off |
| 8 | hybrid-linucb-d-g095 | 2_off |
| 8 | hybrid-linucb-d-g098 | 2_off |
| 8 | hybrid-linucb-d-g099 | 2_off |
| 8 | leastloaded | 2_off |
| 8 | sed | 2_off |
| 9 | hybrid-linucb | 2_off |
| 9 | hybrid-linucb-d-arm-g098 | 2_off |
| 9 | hybrid-linucb-d-g095 | 2_off |
| 9 | hybrid-linucb-d-g098 | 2_off |
| 9 | hybrid-linucb-d-g099 | 2_off |
| 9 | leastloaded | 2_off |
| 9 | sed | 2_off |
| 10 | hybrid-linucb | 2_off |
| 10 | hybrid-linucb-d-arm-g098 | 2_off |
| 10 | hybrid-linucb-d-g095 | 2_off |
| 10 | hybrid-linucb-d-g098 | 2_off |
| 10 | hybrid-linucb-d-g099 | 2_off |
| 10 | leastloaded | 2_off |
| 10 | sed | 2_off |
| 11 | hybrid-linucb | 2_off |
| 11 | hybrid-linucb-d-arm-g098 | 2_off |
| 11 | hybrid-linucb-d-g095 | 2_off |
| 11 | hybrid-linucb-d-g098 | 2_off |
| 11 | hybrid-linucb-d-g099 | 2_off |
| 11 | leastloaded | 2_off |
| 11 | sed | 2_off |
| 12 | hybrid-linucb | 2_off |
| 12 | hybrid-linucb-d-arm-g098 | 2_off |
| 12 | hybrid-linucb-d-g095 | 2_off |
| 12 | hybrid-linucb-d-g098 | 2_off |
| 12 | hybrid-linucb-d-g099 | 2_off |
| 12 | leastloaded | 2_off |
| 12 | sed | 2_off |
| 13 | hybrid-linucb | 2_off |
| 13 | hybrid-linucb-d-arm-g098 | 2_off |
| 13 | hybrid-linucb-d-g095 | 2_off |
| 13 | hybrid-linucb-d-g098 | 2_off |
| 13 | hybrid-linucb-d-g099 | 2_off |
| 13 | leastloaded | 2_off |
| 13 | sed | 2_off |
| 14 | hybrid-linucb | 2_off |
| 14 | hybrid-linucb-d-arm-g098 | 2_off |
| 14 | hybrid-linucb-d-g095 | 2_off |
| 14 | hybrid-linucb-d-g098 | 2_off |
| 14 | hybrid-linucb-d-g099 | 2_off |
| 14 | leastloaded | 2_off |
| 14 | sed | 2_off |
| 15 | hybrid-linucb | 2_off |
| 15 | hybrid-linucb-d-arm-g098 | 2_off |
| 15 | hybrid-linucb-d-g095 | 2_off |
| 15 | hybrid-linucb-d-g098 | 2_off |
| 15 | hybrid-linucb-d-g099 | 2_off |
| 15 | leastloaded | 2_off |
| 15 | sed | 2_off |
| 16 | hybrid-linucb | 2_off |
| 16 | hybrid-linucb-d-arm-g098 | 2_off |
| 16 | hybrid-linucb-d-g095 | 2_off |
| 16 | hybrid-linucb-d-g098 | 2_off |
| 16 | hybrid-linucb-d-g099 | 2_off |
| 16 | leastloaded | 2_off |
| 16 | sed | 2_off |
| 17 | hybrid-linucb | 2_off |
| 17 | hybrid-linucb-d-arm-g098 | 2_off |
| 17 | hybrid-linucb-d-g095 | 2_off |
| 17 | hybrid-linucb-d-g098 | 2_off |
| 17 | hybrid-linucb-d-g099 | 2_off |
| 17 | leastloaded | 2_off |
| 17 | sed | 2_off |
| 18 | hybrid-linucb | 2_off |
| 18 | hybrid-linucb-d-arm-g098 | 2_off |
| 18 | hybrid-linucb-d-g095 | 2_off |
| 18 | hybrid-linucb-d-g098 | 2_off |
| 18 | hybrid-linucb-d-g099 | 2_off |
| 18 | leastloaded | 2_off |
| 18 | sed | 2_off |
| 19 | hybrid-linucb | 2_off |
| 19 | hybrid-linucb-d-arm-g098 | 2_off |
| 19 | hybrid-linucb-d-g095 | 2_off |
| 19 | hybrid-linucb-d-g098 | 2_off |
| 19 | hybrid-linucb-d-g099 | 2_off |
| 19 | leastloaded | 2_off |
| 19 | sed | 2_off |
| 20 | hybrid-linucb | 2_off |
| 20 | hybrid-linucb-d-arm-g098 | 2_off |
| 20 | hybrid-linucb-d-g095 | 2_off |
| 20 | hybrid-linucb-d-g098 | 2_off |
| 20 | hybrid-linucb-d-g099 | 2_off |
| 20 | leastloaded | 2_off |
| 20 | sed | 2_off |

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 174.2 | 20 |
| hybrid-linucb-d-arm-g098 | 175.2 | 20 |
| hybrid-linucb-d-g095 | 181 | 20 |
| hybrid-linucb-d-g098 | 178.9 | 20 |
| hybrid-linucb-d-g099 | 175.9 | 20 |
| leastloaded | 179 | 20 |
| sed | 175 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -3.457 [-6.531, -2.091] | 0.0003948 | 0.002764 | -0.8381 | 0 |
| hybrid-linucb | sed | 20 | -0.028 [-2.295, 1.793] | 0.8408 | 1 | 0.05714 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -4.577 [-5.886, -2.557] | 0.0002098 | 0.001678 | -0.8667 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.143 [-1.374, 1.632] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 1.968 [-1.926, 4.091] | 0.3683 | 1 | 0.2381 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 4.207 [3.451, 7.114] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.3515 [-5.113, 0.884] | 0.3488 | 1 | -0.2476 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 3.796 [1.187, 5.59] | 4.768e-05 | 0.0004292 | 0.9238 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -5.27 [-6.267, -0.246] | 0.001986 | 0.01191 | -0.7524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 1.136 [-0.092, 2.083] | 0.08255 | 0.4128 | 0.4476 | 0 |

## build_1_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 89.88 | 20 |
| hybrid-linucb-d-arm-g098 | 90.36 | 20 |
| hybrid-linucb-d-g095 | 92.77 | 20 |
| hybrid-linucb-d-g098 | 90.61 | 20 |
| hybrid-linucb-d-g099 | 91.1 | 20 |
| leastloaded | 91.5 | 20 |
| sed | 89.68 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -1.398 [-2.029, -0.7095] | 0.03999 | 0.2399 | -0.5238 | 0 |
| hybrid-linucb | sed | 20 | -0.065 [-0.342, 1.894] | 0.33 | 0.7682 | 0.2571 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -1.216 [-2.44, -0.033] | 0.00486 | 0.03888 | -0.6952 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.4995 [-0.2605, 1.061] | 0.165 | 0.7682 | 0.3619 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.835 [-1.499, 3.364] | 0.2774 | 0.7682 | 0.2857 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 2.966 [1.457, 4.307] | 2.67e-05 | 0.000267 | 0.9429 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.576 [-1.532, 0.549] | 0.1536 | 0.7682 | -0.3714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 1.208 [-0.509, 1.902] | 0.02664 | 0.1865 | 0.5619 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.5245 [-1.569, 0.6215] | 0.2455 | 0.7682 | -0.3048 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 1.267 [0.71, 1.889] | 0.001986 | 0.01787 | 0.7524 | 0 |

## build_2_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 84.08 | 20 |
| hybrid-linucb-d-arm-g098 | 84.17 | 20 |
| hybrid-linucb-d-g095 | 87.49 | 20 |
| hybrid-linucb-d-g098 | 87.12 | 20 |
| hybrid-linucb-d-g099 | 84.85 | 20 |
| leastloaded | 86.74 | 20 |
| sed | 85.08 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -2.734 [-5.434, -1.57] | 0.0003948 | 0.002369 | -0.8381 | 0 |
| hybrid-linucb | sed | 20 | -0.9205 [-1.543, 0.36] | 0.1231 | 0.6155 | -0.4 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -2.446 [-4.392, -1.718] | 0.0003223 | 0.002256 | -0.8476 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.8855 [-1.255, 0.148] | 0.3884 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.06 [-1.018, 1.527] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 2.715 [1.762, 4.043] | 0.0001049 | 0.0009441 | 0.8952 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.3485 [-2.685, 1.721] | 0.4524 | 1 | -0.2 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.181 [1.346, 2.834] | 4.768e-05 | 0.0004768 | 0.9238 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -2.397 [-4.943, -0.956] | 0.0001335 | 0.001068 | -0.8857 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.195 [-0.941, 0.591] | 0.5706 | 1 | -0.1524 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.00339 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.08729 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 18 |
| hybrid-linucb | sed | 20 | -0.08729 [-0.09153, -0.07203] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.08729 [-0.09153, -0.07203] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.00339 [0.002542, 0.005085] | 0.0001779 | 0.0008894 | 1 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | -0.08051 [-0.08814, -0.06949] | 8.832e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g098 | sed | 20 | -0.08729 [-0.09153, -0.07203] | 8.807e-05 | 0.0008807 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g099 | sed | 20 | -0.08729 [-0.09153, -0.07203] | 8.807e-05 | 0.0008807 | -1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-arm-g098 | 0 | 20 |
| hybrid-linucb-d-g095 | 0.00711 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g099 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1885 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -0.3333 | 18 |
| hybrid-linucb | sed | 20 | -0.1885 [-0.1996, -0.1685] | 8.845e-05 | 0.000796 | -1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.1885 [-0.1996, -0.1685] | 8.845e-05 | 0.000796 | -1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.00711 [0.005112, 0.01035] | 0.0001961 | 0.0009805 | 1 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | -0.1789 [-0.1917, -0.1597] | 1.907e-06 | 1.907e-05 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1885 [-0.1996, -0.1685] | 8.845e-05 | 0.000796 | -1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | -1 | 19 |
| hybrid-linucb-d-g099 | sed | 20 | -0.1885 [-0.1996, -0.1685] | 8.845e-05 | 0.000796 | -1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.949 | 20 |
| hybrid-linucb-d-arm-g098 | 0.9532 | 20 |
| hybrid-linucb-d-g095 | 0.8663 | 20 |
| hybrid-linucb-d-g098 | 0.8981 | 20 |
| hybrid-linucb-d-g099 | 0.9237 | 20 |
| leastloaded | 0.9067 | 20 |
| sed | 0.8115 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03286 [0.02427, 0.04701] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb | sed | 20 | 0.1237 [0.1083, 0.1515] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04774 [0.0367, 0.0518] | 3.624e-05 | 0.0001812 | 0.9333 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.138 [0.1095, 0.1533] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.03556 [-0.05498, -0.01539] | 0.0001335 | 0.0005341 | -0.8857 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.05408 [0.02025, 0.07669] | 0.0002613 | 0.0007839 | 0.8571 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.007139 [-0.01615, 0.00443] | 0.1893 | 0.1893 | -0.3429 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.09251 [0.0618, 0.1016] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.01212 [0.001796, 0.02863] | 0.04405 | 0.08811 | 0.5143 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.1042 [0.09033, 0.1255] | 1.907e-06 | 1.907e-05 | 1 | 0 |

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
| hybrid-linucb | 0.2542 | 20 |
| hybrid-linucb-d-arm-g098 | 0.272 | 20 |
| hybrid-linucb-d-g095 | 0.2686 | 20 |
| hybrid-linucb-d-g098 | 0.2602 | 20 |
| hybrid-linucb-d-g099 | 0.2475 | 20 |
| leastloaded | 0.2449 | 20 |
| sed | 0.2449 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.007627 [-0.02627, 0.02458] | 0.444 | 1 | 0.1952 | 0 |
| hybrid-linucb | sed | 20 | 0.00678 [-0.01356, 0.02966] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.02458 [-0.001695, 0.03898] | 0.05214 | 0.365 | 0.4952 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.02034 [0.01525, 0.03729] | 0.02759 | 0.2483 | 0.5619 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02288 [0.01017, 0.04915] | 0.000671 | 0.00671 | 0.8895 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 0.02458 [-0.002542, 0.0661] | 0.0382 | 0.3056 | 0.5421 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.005932 [-0.01102, 0.02966] | 0.3133 | 1 | 0.2571 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01271 [-0.01356, 0.03305] | 0.3045 | 1 | 0.2684 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.007627 [-0.00339, 0.03559] | 0.09096 | 0.5458 | 0.4421 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01949 [-0.02288, 0.02881] | 0.629 | 1 | 0.1263 | 1 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.4059 | 20 |
| hybrid-linucb-d-arm-g098 | 0.4085 | 20 |
| hybrid-linucb-d-g095 | 0.4924 | 20 |
| hybrid-linucb-d-g098 | 0.4593 | 20 |
| hybrid-linucb-d-g099 | 0.4331 | 20 |
| leastloaded | 0.4314 | 20 |
| sed | 0.4576 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.03136 [-0.07119, 0] | 0.06654 | 0.3327 | -0.4971 | 2 |
| hybrid-linucb | sed | 20 | -0.05169 [-0.07034, -0.01102] | 0.006066 | 0.04246 | -0.7 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.02881 [-0.04492, -0.01441] | 0.001606 | 0.01445 | -0.8048 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.03729 [-0.06102, -0.01864] | 0.000889 | 0.00889 | -0.8476 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.04153 [0.02542, 0.0839] | 0.001941 | 0.01553 | 0.7905 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03814 [0.007627, 0.05] | 0.009451 | 0.05671 | 0.6619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02034 [-0.007627, 0.04237] | 0.1776 | 0.5327 | 0.3526 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 0.00678 [-0.02966, 0.0339] | 0.6012 | 0.6012 | 0.1333 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.01186 [-0.04237, 0.01695] | 0.2789 | 0.5578 | -0.2762 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.01864 [-0.05508, 0.007627] | 0.08576 | 0.343 | -0.4381 | 0 |

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
| hybrid-linucb | 0.2731 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2374 | 20 |
| hybrid-linucb-d-g095 | 0.1643 | 20 |
| hybrid-linucb-d-g098 | 0.1583 | 20 |
| hybrid-linucb-d-g099 | 0.3333 | 20 |
| leastloaded | 0.2222 | 20 |
| sed | 0.1927 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.00728 [-0.1125, 0.2429] | 0.2935 | 1 | 0.2842 | 1 |
| hybrid-linucb | sed | 20 | 0.03448 [-0.02072, 0.2362] | 0.1327 | 1 | 0.3905 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.04466 [-0.08575, 0.2054] | 0.7841 | 1 | 0.07619 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.06674 [-0.08974, 0.2069] | 0.4552 | 1 | 0.1905 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.04673 [-0.1974, 0.05057] | 0.2455 | 1 | -0.3048 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.01724 [-0.09429, 0.0702] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.1028 [-0.1814, 0.1023] | 0.498 | 1 | -0.181 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01786 [-0.1261, 0.1239] | 0.8405 | 1 | -0.05263 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.05357 [-0.02416, 0.2713] | 0.1956 | 1 | 0.3474 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | 0.1578 [0.05, 0.2325] | 0.08254 | 0.8254 | 0.4429 | 0 |

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2558 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2248 | 20 |
| hybrid-linucb-d-g095 | 0.2101 | 20 |
| hybrid-linucb-d-g098 | 0.2287 | 20 |
| hybrid-linucb-d-g099 | 0.2578 | 20 |
| leastloaded | 0.2201 | 20 |
| sed | 0.2442 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.01249 [-0.04904, 0.1267] | 0.1956 | 1 | 0.3474 | 1 |
| hybrid-linucb | sed | 20 | 0.01852 [-0.06698, 0.09856] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.03347 [-0.004383, 0.09204] | 0.4749 | 1 | 0.1905 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.005237 [-0.07759, 0.04264] | 0.5379 | 1 | -0.1571 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -0.007245 [-0.04518, 0.01719] | 0.5958 | 1 | -0.1429 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | -0.05213 [-0.08804, -0.02946] | 0.02395 | 0.2395 | -0.5714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01837 [-0.07093, 0.04447] | 0.8695 | 1 | 0.04762 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.02069 [-0.0691, 0.007752] | 0.1447 | 1 | -0.3895 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.03497 [9.092e-05, 0.08323] | 0.09731 | 0.8758 | 0.4286 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.001683 [-0.03837, 0.02684] | 0.8906 | 1 | -0.04211 | 1 |

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2495 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2557 | 20 |
| hybrid-linucb-d-g095 | 0.2449 | 20 |
| hybrid-linucb-d-g098 | 0.25 | 20 |
| hybrid-linucb-d-g099 | 0.2397 | 20 |
| leastloaded | 0.2459 | 20 |
| sed | 0.2497 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.0121 [-0.02949, 0.03016] | 0.953 | 1 | 0.02105 | 1 |
| hybrid-linucb | sed | 20 | -0.007592 [-0.01896, 0.01318] | 0.6541 | 1 | -0.1143 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01707 [-0.014, 0.02482] | 0.2774 | 1 | 0.2857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01397 [0.001085, 0.02851] | 0.06193 | 0.6193 | 0.4762 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.00684 [-0.01846, 0.02057] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.01142 [-0.04171, 0.03633] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.008408 [-0.03239, 0.03284] | 0.9273 | 1 | -0.02857 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.002169 [-0.02166, 0.02813] | 0.9854 | 1 | -0.009524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.008088 [-0.02198, 0.01457] | 0.5706 | 1 | -0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.005991 [-0.02469, 0.009408] | 0.5706 | 1 | -0.1524 | 0 |

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 29 | 20 |
| hybrid-linucb-d-arm-g098 | 29.5 | 20 |
| hybrid-linucb-d-g095 | 29 | 20 |
| hybrid-linucb-d-g098 | 29 | 20 |
| hybrid-linucb-d-g099 | 29 | 20 |
| leastloaded | 29 | 20 |
| sed | 29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [-1, 1] | 0.7092 | 1 | 0.1029 | 4 |
| hybrid-linucb | sed | 20 | 0 [-1, 0.5] | 0.6579 | 1 | -0.1319 | 7 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 1.5 [0, 2] | 0.08626 | 0.8045 | 0.4641 | 3 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.5 [0, 1] | 0.08045 | 0.8045 | 0.5275 | 7 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [-0.5, 1] | 0.4802 | 1 | 0.2 | 5 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [-1, 1] | 0.973 | 1 | 0.009524 | 6 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 0] | 0.8908 | 1 | -0.04545 | 9 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0] | 0.2362 | 1 | -0.3788 | 9 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 1 [0, 1] | 0.6191 | 1 | 0.1397 | 4 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 1] | 1 | 1 | 0 | 8 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 129 | 20 |
| hybrid-linucb-d-arm-g098 | 129.5 | 20 |
| hybrid-linucb-d-g095 | 129 | 20 |
| hybrid-linucb-d-g098 | 129 | 20 |
| hybrid-linucb-d-g099 | 129 | 20 |
| leastloaded | 129 | 20 |
| sed | 129 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [-1, 1] | 0.7092 | 1 | 0.1029 | 4 |
| hybrid-linucb | sed | 20 | 0 [-1, 0.5] | 0.6579 | 1 | -0.1319 | 7 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 1.5 [0, 2] | 0.08626 | 0.8045 | 0.4641 | 3 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.5 [0, 1] | 0.08045 | 0.8045 | 0.5275 | 7 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [-0.5, 1] | 0.4802 | 1 | 0.2 | 5 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [-1, 1] | 0.973 | 1 | 0.009524 | 6 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 0] | 0.8908 | 1 | -0.04545 | 9 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [-1, 0] | 0.2362 | 1 | -0.3788 | 9 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 1 [0, 1] | 0.6191 | 1 | 0.1397 | 4 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [0, 1] | 1 | 1 | 0 | 8 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 461 | 20 |
| hybrid-linucb-d-arm-g098 | 460.5 | 20 |
| hybrid-linucb-d-g095 | 461 | 20 |
| hybrid-linucb-d-g098 | 461 | 20 |
| hybrid-linucb-d-g099 | 461 | 20 |
| leastloaded | 461 | 20 |
| sed | 461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [-1, 1] | 0.7092 | 1 | -0.1029 | 4 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 1] | 0.6579 | 1 | 0.1319 | 7 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -1.5 [-2, 0] | 0.08626 | 0.8045 | -0.4641 | 3 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -0.5 [-1, 0] | 0.08045 | 0.8045 | -0.5275 | 7 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0 [-1, 0.5] | 0.4802 | 1 | -0.2 | 5 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [-1, 1] | 0.973 | 1 | -0.009524 | 6 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0.5] | 0.8908 | 1 | 0.04545 | 9 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 1] | 0.2362 | 1 | 0.3788 | 9 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -1 [-1, 0] | 0.6191 | 1 | -0.1397 | 4 |
| hybrid-linucb-d-g099 | sed | 20 | 0 [-1, 0] | 1 | 1 | 0 | 8 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2 | 20 |
| hybrid-linucb-d-arm-g098 | 0.19 | 20 |
| hybrid-linucb-d-g095 | 0.18 | 20 |
| hybrid-linucb-d-g098 | 0.17 | 20 |
| hybrid-linucb-d-g099 | 0.21 | 20 |
| leastloaded | 0.17 | 20 |
| sed | 0.13 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03 [-0.07, 0.14] | 0.3865 | 1 | 0.2263 | 1 |
| hybrid-linucb | sed | 20 | 0.07 [-0.03, 0.16] | 0.08699 | 0.8699 | 0.4474 | 1 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.02 [-0.02, 0.06] | 0.3949 | 1 | 0.2281 | 2 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.04 [-0.01, 0.1] | 0.1911 | 1 | 0.3509 | 2 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02 [-0.05, 0.05] | 0.6472 | 1 | 0.1228 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | 0.04 [-0.01, 0.07] | 0.2588 | 1 | 0.2947 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.02 [-0.06, 0.05] | 0.776 | 1 | -0.07843 | 3 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02 [-0.03, 0.11] | 0.5517 | 1 | 0.1691 | 4 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -1.388e-17 [-0.04, 0.08] | 0.4219 | 1 | 0.2048 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02 [-0.02, 0.11] | 0.1767 | 1 | 0.3626 | 2 |

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2464 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2664 | 20 |
| hybrid-linucb-d-g095 | 0.257 | 20 |
| hybrid-linucb-d-g098 | 0.2655 | 20 |
| hybrid-linucb-d-g099 | 0.2518 | 20 |
| leastloaded | 0.2527 | 20 |
| sed | 0.2579 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.002504 [-0.03727, 0.02281] | 0.5459 | 1 | -0.1619 | 0 |
| hybrid-linucb | sed | 20 | -0.01398 [-0.03, 0.00883] | 0.3124 | 1 | -0.2737 | 1 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01214 [-0.005855, 0.02293] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.007985 [-0.003409, 0.03441] | 0.1712 | 1 | 0.3579 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.01248 [-0.02559, 0.01915] | 0.8695 | 1 | 0.04762 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.0005863 [-0.04402, 0.03645] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.006994 [-0.02177, 0.03191] | 0.8983 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.001525 [-0.02187, 0.03167] | 0.8983 | 1 | -0.0381 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.01642 [-0.02684, 0.004063] | 0.3118 | 1 | -0.2667 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.01012 [-0.028, 0.009808] | 0.2943 | 1 | -0.2762 | 0 |

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.002477 | 20 |
| hybrid-linucb-d-arm-g098 | -0.15 | 20 |
| hybrid-linucb-d-g095 | -0.1474 | 20 |
| hybrid-linucb-d-g098 | -0.1036 | 20 |
| hybrid-linucb-d-g099 | -0.05619 | 20 |
| leastloaded | -0.237 | 20 |
| sed | -0.2195 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.2143 [-0.036, 0.3806] | 0.009436 | 0.07549 | 0.6476 | 0 |
| hybrid-linucb | sed | 20 | 0.1721 [-0.01999, 0.371] | 0.01531 | 0.07656 | 0.6095 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.1507 [0.03038, 0.2232] | 0.09731 | 0.2919 | 0.4286 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.08943 [-0.01432, 0.231] | 0.114 | 0.2919 | 0.4095 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.1092 [0.02219, 0.1957] | 0.00486 | 0.04374 | 0.6952 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03404 [-0.03049, 0.2608] | 0.2024 | 0.2919 | 0.3333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.1712 [-0.02609, 0.276] | 0.01069 | 0.07549 | 0.6381 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.1544 [0.0857, 0.2252] | 0.02395 | 0.0958 | 0.5714 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.1453 [0.08721, 0.3213] | 0.0003948 | 0.003948 | 0.8381 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.1878 [0.006538, 0.2657] | 0.009436 | 0.07549 | 0.6476 | 0 |

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.1921 | 20 |
| hybrid-linucb-d-arm-g098 | -0.1686 | 20 |
| hybrid-linucb-d-g095 | -0.1576 | 20 |
| hybrid-linucb-d-g098 | -0.1697 | 20 |
| hybrid-linucb-d-g099 | -0.1843 | 20 |
| leastloaded | -0.2283 | 20 |
| sed | -0.1993 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.04708 [0.02076, 0.0694] | 0.0008507 | 0.003538 | 0.8 | 0 |
| hybrid-linucb | sed | 20 | 0.006999 [-0.007628, 0.02988] | 0.2305 | 0.2305 | 0.3143 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.0854 [0.0414, 0.1021] | 1.335e-05 | 0.0001068 | 0.9619 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.04548 [0.02994, 0.0672] | 0.0007076 | 0.003538 | 0.8095 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.08904 [0.06307, 0.103] | 1.907e-06 | 1.907e-05 | 1 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.05211 [0.02894, 0.0781] | 6.294e-05 | 0.0003777 | 0.9143 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.05696 [0.03551, 0.07498] | 9.537e-06 | 8.583e-05 | 0.9714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02226 [0.006127, 0.06799] | 0.004221 | 0.01266 | 0.7048 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.06479 [0.02531, 0.07407] | 3.624e-05 | 0.0002537 | 0.9333 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.02365 [0.007712, 0.0418] | 0.008308 | 0.01662 | 0.6571 | 0 |

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 27 | 20 |
| hybrid-linucb-d-arm-g098 | 26.5 | 20 |
| hybrid-linucb-d-g095 | 26 | 20 |
| hybrid-linucb-d-g098 | 27 | 20 |
| hybrid-linucb-d-g099 | 42.5 | 20 |
| leastloaded | 40.5 | 20 |
| sed | 20 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -5.5 [-27, 27.5] | 0.8563 | 1 | -0.04737 | 1 |
| hybrid-linucb | sed | 20 | 3 [-2, 19.5] | 0.1356 | 1 | 0.4118 | 3 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -4.5 [-27, 4] | 0.5135 | 1 | -0.1754 | 2 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.5 [-8, 32.5] | 0.2461 | 1 | 0.3203 | 3 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -12 [-24.5, 0] | 0.6112 | 1 | -0.1503 | 3 |
| hybrid-linucb-d-g095 | sed | 20 | 2 [0, 19.5] | 0.04369 | 0.4369 | 0.5917 | 5 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -3.5 [-16, 9] | 0.6274 | 1 | -0.1238 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 4.5 [-2, 20] | 0.2095 | 1 | 0.3464 | 3 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.5 [-20.5, 17] | 0.7368 | 1 | -0.08571 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 5 [-3, 29] | 0.1219 | 1 | 0.4152 | 2 |

## target_after_recovery_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.002046 [-0.04402, 0.01606] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb | sed | 20 | -0.02292 [-0.04106, 0.002203] | 0.114 | 1 | -0.4095 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01612 [-0.003662, 0.0459] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01455 [-0.004629, 0.0379] | 0.1536 | 1 | 0.3714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.0004124 [-0.03106, 0.02789] | 0.9217 | 1 | -0.03158 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -0.007579 [-0.05716, 0.03948] | 0.8408 | 1 | -0.05714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.01435 [-0.0452, 0.02359] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0003242 [-0.05887, 0.02976] | 0.3884 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.02055 [-0.03427, 0.001943] | 0.165 | 1 | -0.3619 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.00771 [-0.03289, 0.006494] | 0.3884 | 1 | -0.2286 | 0 |

## target_after_recovery_tasks

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 87 | 20 |
| hybrid-linucb-d-arm-g098 | 94.5 | 20 |
| hybrid-linucb-d-g095 | 87 | 20 |
| hybrid-linucb-d-g098 | 86.5 | 20 |
| hybrid-linucb-d-g099 | 84.5 | 20 |
| leastloaded | 90 | 20 |
| sed | 91.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.5 [-14.5, 3.5] | 0.37 | 1 | -0.2286 | 0 |
| hybrid-linucb | sed | 20 | -6.5 [-14, -0.5] | 0.1504 | 1 | -0.3667 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 3.5 [-3.5, 10.5] | 0.3339 | 1 | 0.2526 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 2 [-4, 11] | 0.3756 | 1 | 0.2316 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -1.5 [-12.5, 6.5] | 0.6579 | 1 | -0.1158 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 1 [-17.5, 11] | 0.8082 | 1 | -0.0619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -1.5 [-16, 7] | 0.4008 | 1 | -0.2143 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -1 [-15, 9.5] | 0.3224 | 1 | -0.2524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -8 [-12.5, -0.5] | 0.147 | 1 | -0.3789 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | -4 [-10, 4] | 0.3441 | 1 | -0.2474 | 1 |

## drift_on_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2062 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1781 | 20 |
| hybrid-linucb-d-g095 | 0.1993 | 20 |
| hybrid-linucb-d-g098 | 0.2024 | 20 |
| hybrid-linucb-d-g099 | 0.185 | 20 |
| leastloaded | 0.1776 | 20 |
| sed | 0.1641 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.02548 [-0.02, 0.092] | 0.165 | 1 | 0.3619 | 0 |
| hybrid-linucb | sed | 20 | 0.04999 [-0.01211, 0.08344] | 0.09731 | 0.9731 | 0.4286 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.006461 [-0.02374, 0.02851] | 0.8695 | 1 | 0.04762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01791 [-0.01221, 0.04118] | 0.2305 | 1 | 0.3143 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02773 [-0.02376, 0.06111] | 0.2024 | 1 | 0.3333 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03365 [-0.02698, 0.05857] | 0.2774 | 1 | 0.2857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01698 [-0.01569, 0.05539] | 0.1769 | 1 | 0.3524 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.04115 [-0.024, 0.06784] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.007123 [-0.02175, 0.04675] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01165 [-0.03456, 0.0565] | 0.3488 | 1 | 0.2476 | 0 |

## drift_on_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 133.5 | 20 |
| hybrid-linucb-d-arm-g098 | 143.5 | 20 |
| hybrid-linucb-d-g095 | 139 | 20 |
| hybrid-linucb-d-g098 | 138.5 | 20 |
| hybrid-linucb-d-g099 | 137.5 | 20 |
| leastloaded | 134 | 20 |
| sed | 136.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 1 [-2, 5.5] | 0.5126 | 1 | 0.1667 | 0 |
| hybrid-linucb | sed | 20 | -0.5 [-5, 4] | 0.7603 | 1 | -0.08187 | 2 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 5 [2, 16] | 0.002026 | 0.02026 | 0.7857 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 6 [2, 10] | 0.00251 | 0.02259 | 0.7895 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 3.5 [0, 8.5] | 0.08912 | 0.7129 | 0.4561 | 2 |
| hybrid-linucb-d-g095 | sed | 20 | 0 [-2.5, 8] | 0.5415 | 1 | 0.1637 | 2 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 4.5 [0.5, 6.5] | 0.3041 | 1 | 0.2684 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 2.5 [-4.5, 5] | 0.9107 | 1 | 0.02857 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 4 [-3, 9] | 0.2191 | 1 | 0.3211 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | 1.5 [-4.5, 6.5] | 0.6146 | 1 | 0.1316 | 1 |

## drift_on_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 28 | 20 |
| hybrid-linucb-d-arm-g098 | 26 | 20 |
| hybrid-linucb-d-g095 | 27 | 20 |
| hybrid-linucb-d-g098 | 27 | 20 |
| hybrid-linucb-d-g099 | 26.5 | 20 |
| leastloaded | 23.5 | 20 |
| sed | 22 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 5.5 [-3, 12] | 0.112 | 0.8962 | 0.4048 | 0 |
| hybrid-linucb | sed | 20 | 6.5 [-3, 10.5] | 0.08584 | 0.7726 | 0.4381 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.5 [-2, 5] | 0.3154 | 0.9866 | 0.269 | 2 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 4 [-1.5, 7.5] | 0.07632 | 0.7632 | 0.4632 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 5 [-2, 9] | 0.1582 | 0.9866 | 0.3684 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 4 [-2, 8.5] | 0.2106 | 0.9866 | 0.319 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 2.5 [-1.5, 9] | 0.1585 | 0.9866 | 0.3684 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 6 [-2.5, 10] | 0.1409 | 0.9866 | 0.3842 | 1 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 2.5 [-2.5, 6.5] | 0.3538 | 0.9866 | 0.2421 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | 2 [-3, 7.5] | 0.3042 | 0.9866 | 0.2684 | 1 |

## drift_off_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.002046 [-0.04402, 0.01606] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb | sed | 20 | -0.02292 [-0.04106, 0.002203] | 0.114 | 1 | -0.4095 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01612 [-0.003662, 0.0459] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01455 [-0.004629, 0.0379] | 0.1536 | 1 | 0.3714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.0004124 [-0.03106, 0.02789] | 0.9217 | 1 | -0.03158 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -0.007579 [-0.05716, 0.03948] | 0.8408 | 1 | -0.05714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.01435 [-0.0452, 0.02359] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0003242 [-0.05887, 0.02976] | 0.3884 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.02055 [-0.03427, 0.001943] | 0.165 | 1 | -0.3619 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.00771 [-0.03289, 0.006494] | 0.3884 | 1 | -0.2286 | 0 |

## drift_off_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 328 | 20 |
| hybrid-linucb-d-arm-g098 | 317 | 20 |
| hybrid-linucb-d-g095 | 321.5 | 20 |
| hybrid-linucb-d-g098 | 322.5 | 20 |
| hybrid-linucb-d-g099 | 323.5 | 20 |
| leastloaded | 327.5 | 20 |
| sed | 325 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -1 [-7, 2.5] | 0.5325 | 1 | -0.1632 | 1 |
| hybrid-linucb | sed | 20 | 1.5 [-4, 5.5] | 0.601 | 1 | 0.1333 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -4.5 [-16.5, -2] | 0.001182 | 0.01182 | -0.8474 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | -6.5 [-11, -2.5] | 0.002458 | 0.02212 | -0.7714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -4 [-11, -0.5] | 0.07623 | 0.6098 | -0.4632 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -1 [-7.5, 2] | 0.5252 | 1 | -0.1619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -4 [-6.5, -0.5] | 0.2856 | 1 | -0.2789 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -3 [-4.5, 5] | 0.9701 | 1 | 0.009524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -6 [-8.5, 3] | 0.1907 | 1 | -0.3333 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -1.5 [-7, 4.5] | 0.5624 | 1 | -0.1476 | 0 |

## drift_off_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 87 | 20 |
| hybrid-linucb-d-arm-g098 | 94.5 | 20 |
| hybrid-linucb-d-g095 | 87 | 20 |
| hybrid-linucb-d-g098 | 86.5 | 20 |
| hybrid-linucb-d-g099 | 84.5 | 20 |
| leastloaded | 90 | 20 |
| sed | 91.5 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -0.5 [-14.5, 3.5] | 0.37 | 1 | -0.2286 | 0 |
| hybrid-linucb | sed | 20 | -6.5 [-14, -0.5] | 0.1504 | 1 | -0.3667 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 3.5 [-3, 10.5] | 0.3339 | 1 | 0.2526 | 1 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 2 [-4, 11] | 0.3756 | 1 | 0.2316 | 1 |
| hybrid-linucb-d-g095 | leastloaded | 20 | -1.5 [-12.5, 6.5] | 0.6579 | 1 | -0.1158 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | 1 [-17.5, 11] | 0.8082 | 1 | -0.0619 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -1.5 [-16, 7] | 0.4008 | 1 | -0.2143 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -1 [-15, 9.5] | 0.3224 | 1 | -0.2524 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -8 [-12.5, -0.5] | 0.147 | 1 | -0.3789 | 1 |
| hybrid-linucb-d-g099 | sed | 20 | -4 [-10, 4] | 0.3441 | 1 | -0.2474 | 1 |

## drift_phase_1_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2062 | 20 |
| hybrid-linucb-d-arm-g098 | 0.1781 | 20 |
| hybrid-linucb-d-g095 | 0.1993 | 20 |
| hybrid-linucb-d-g098 | 0.2024 | 20 |
| hybrid-linucb-d-g099 | 0.185 | 20 |
| leastloaded | 0.1776 | 20 |
| sed | 0.1641 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.02548 [-0.02, 0.092] | 0.165 | 1 | 0.3619 | 0 |
| hybrid-linucb | sed | 20 | 0.04999 [-0.01211, 0.08344] | 0.09731 | 0.9731 | 0.4286 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | -0.006461 [-0.02374, 0.02851] | 0.8695 | 1 | 0.04762 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01791 [-0.01221, 0.04118] | 0.2305 | 1 | 0.3143 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.02773 [-0.02376, 0.06111] | 0.2024 | 1 | 0.3333 | 0 |
| hybrid-linucb-d-g095 | sed | 20 | 0.03365 [-0.02698, 0.05857] | 0.2774 | 1 | 0.2857 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01698 [-0.01569, 0.05539] | 0.1769 | 1 | 0.3524 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.04115 [-0.024, 0.06784] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | 0.007123 [-0.02175, 0.04675] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | 0.01165 [-0.03456, 0.0565] | 0.3488 | 1 | 0.2476 | 0 |

## drift_phase_2_off_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2699 | 20 |
| hybrid-linucb-d-arm-g098 | 0.2957 | 20 |
| hybrid-linucb-d-g095 | 0.2676 | 20 |
| hybrid-linucb-d-g098 | 0.2699 | 20 |
| hybrid-linucb-d-g099 | 0.2581 | 20 |
| leastloaded | 0.2719 | 20 |
| sed | 0.2829 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.002046 [-0.04402, 0.01606] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb | sed | 20 | -0.02292 [-0.04106, 0.002203] | 0.114 | 1 | -0.4095 | 0 |
| hybrid-linucb-d-arm-g098 | leastloaded | 20 | 0.01612 [-0.003662, 0.0459] | 0.1429 | 1 | 0.381 | 0 |
| hybrid-linucb-d-arm-g098 | sed | 20 | 0.01455 [-0.004629, 0.0379] | 0.1536 | 1 | 0.3714 | 0 |
| hybrid-linucb-d-g095 | leastloaded | 20 | 0.0004124 [-0.03106, 0.02789] | 0.9217 | 1 | -0.03158 | 1 |
| hybrid-linucb-d-g095 | sed | 20 | -0.007579 [-0.05716, 0.03948] | 0.8408 | 1 | -0.05714 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.01435 [-0.0452, 0.02359] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.0003242 [-0.05887, 0.02976] | 0.3884 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g099 | leastloaded | 20 | -0.02055 [-0.03427, 0.001943] | 0.165 | 1 | -0.3619 | 0 |
| hybrid-linucb-d-g099 | sed | 20 | -0.00771 [-0.03289, 0.006494] | 0.3884 | 1 | -0.2286 | 0 |

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-arm-g098 | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 1 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 1 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 1 | leastloaded | worker-4:50051 | 1000 |
| 1 | sed | worker-4:50051 | 1000 |
| 2 | hybrid-linucb | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 2 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 2 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 2 | leastloaded | worker-2:50051 | 600 |
| 2 | sed | worker-5:50051 | 1100 |
| 3 | hybrid-linucb | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 3 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 3 | leastloaded | worker-3:50051 | 800 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 4 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 4 | leastloaded | worker-1:50051 | 500 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 5 | leastloaded | worker-4:50051 | 1000 |
| 5 | sed | worker-5:50051 | 1100 |
| 6 | hybrid-linucb | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 6 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 6 | leastloaded | worker-1:50051 | 500 |
| 6 | sed | worker-5:50051 | 1100 |
| 7 | hybrid-linucb | worker-1:50051 | 500 |
| 7 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 7 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-4:50051 | 1000 |
| 8 | hybrid-linucb | worker-1:50051 | 500 |
| 8 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 8 | leastloaded | worker-3:50051 | 800 |
| 8 | sed | worker-5:50051 | 1100 |
| 9 | hybrid-linucb | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-d-g095 | worker-2:50051 | 600 |
| 9 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 9 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 9 | leastloaded | worker-1:50051 | 500 |
| 9 | sed | worker-3:50051 | 800 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g095 | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 10 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 10 | leastloaded | worker-1:50051 | 500 |
| 10 | sed | worker-5:50051 | 1100 |
| 11 | hybrid-linucb | worker-5:50051 | 1100 |
| 11 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 11 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 11 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 11 | leastloaded | worker-3:50051 | 800 |
| 11 | sed | worker-3:50051 | 800 |
| 12 | hybrid-linucb | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 12 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 12 | leastloaded | worker-1:50051 | 500 |
| 12 | sed | worker-4:50051 | 1000 |
| 13 | hybrid-linucb | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 13 | leastloaded | worker-2:50051 | 600 |
| 13 | sed | worker-5:50051 | 1100 |
| 14 | hybrid-linucb | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 14 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 14 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 14 | leastloaded | worker-1:50051 | 500 |
| 14 | sed | worker-2:50051 | 600 |
| 15 | hybrid-linucb | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 15 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 15 | leastloaded | worker-3:50051 | 800 |
| 15 | sed | worker-2:50051 | 600 |
| 16 | hybrid-linucb | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-arm-g098 | worker-2:50051 | 600 |
| 16 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 16 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 16 | leastloaded | worker-1:50051 | 500 |
| 16 | sed | worker-4:50051 | 1000 |
| 17 | hybrid-linucb | worker-2:50051 | 600 |
| 17 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 17 | hybrid-linucb-d-g099 | worker-5:50051 | 1100 |
| 17 | leastloaded | worker-2:50051 | 600 |
| 17 | sed | worker-2:50051 | 600 |
| 18 | hybrid-linucb | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-arm-g098 | worker-3:50051 | 800 |
| 18 | hybrid-linucb-d-g095 | worker-1:50051 | 500 |
| 18 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g099 | worker-4:50051 | 1000 |
| 18 | leastloaded | worker-5:50051 | 1100 |
| 18 | sed | worker-5:50051 | 1100 |
| 19 | hybrid-linucb | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-arm-g098 | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-d-g095 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g099 | worker-3:50051 | 800 |
| 19 | leastloaded | worker-1:50051 | 500 |
| 19 | sed | worker-4:50051 | 1000 |
| 20 | hybrid-linucb | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-arm-g098 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-g095 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 20 | hybrid-linucb-d-g099 | worker-1:50051 | 500 |
| 20 | leastloaded | worker-4:50051 | 1000 |
| 20 | sed | worker-5:50051 | 1100 |
