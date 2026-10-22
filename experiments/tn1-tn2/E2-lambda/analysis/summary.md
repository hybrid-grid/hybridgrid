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
| 1 | hybrid-linucb-d-g098 | 1_on |
| 1 | hybrid-linucb-d-g098-l0 | 1_on |
| 1 | hybrid-linucb-d-g098-l025 | 1_on |
| 1 | hybrid-linucb-d-g098-l100 | 1_on |
| 1 | hybrid-linucb-l0 | 1_on |
| 1 | hybrid-linucb-l025 | 1_on |
| 1 | hybrid-linucb-l100 | 1_on |
| 1 | leastloaded | 1_on |
| 1 | sed | 1_on |
| 2 | hybrid-linucb | 1_on |
| 2 | hybrid-linucb-d-g098 | 1_on |
| 2 | hybrid-linucb-d-g098-l0 | 1_on |
| 2 | hybrid-linucb-d-g098-l025 | 1_on |
| 2 | hybrid-linucb-d-g098-l100 | 1_on |
| 2 | hybrid-linucb-l0 | 1_on |
| 2 | hybrid-linucb-l025 | 1_on |
| 2 | hybrid-linucb-l100 | 1_on |
| 2 | leastloaded | 1_on |
| 2 | sed | 1_on |
| 3 | hybrid-linucb | 1_on |
| 3 | hybrid-linucb-d-g098 | 1_on |
| 3 | hybrid-linucb-d-g098-l0 | 1_on |
| 3 | hybrid-linucb-d-g098-l025 | 1_on |
| 3 | hybrid-linucb-d-g098-l100 | 1_on |
| 3 | hybrid-linucb-l0 | 1_on |
| 3 | hybrid-linucb-l025 | 1_on |
| 3 | hybrid-linucb-l100 | 1_on |
| 3 | leastloaded | 1_on |
| 3 | sed | 1_on |
| 4 | hybrid-linucb | 1_on |
| 4 | hybrid-linucb-d-g098 | 1_on |
| 4 | hybrid-linucb-d-g098-l0 | 1_on |
| 4 | hybrid-linucb-d-g098-l025 | 1_on |
| 4 | hybrid-linucb-d-g098-l100 | 1_on |
| 4 | hybrid-linucb-l0 | 1_on |
| 4 | hybrid-linucb-l025 | 1_on |
| 4 | hybrid-linucb-l100 | 1_on |
| 4 | leastloaded | 1_on |
| 4 | sed | 1_on |
| 5 | hybrid-linucb | 1_on |
| 5 | hybrid-linucb-d-g098 | 1_on |
| 5 | hybrid-linucb-d-g098-l0 | 1_on |
| 5 | hybrid-linucb-d-g098-l025 | 1_on |
| 5 | hybrid-linucb-d-g098-l100 | 1_on |
| 5 | hybrid-linucb-l0 | 1_on |
| 5 | hybrid-linucb-l025 | 1_on |
| 5 | hybrid-linucb-l100 | 1_on |
| 5 | leastloaded | 1_on |
| 5 | sed | 1_on |
| 6 | hybrid-linucb | 1_on |
| 6 | hybrid-linucb-d-g098 | 1_on |
| 6 | hybrid-linucb-d-g098-l0 | 1_on |
| 6 | hybrid-linucb-d-g098-l025 | 1_on |
| 6 | hybrid-linucb-d-g098-l100 | 1_on |
| 6 | hybrid-linucb-l0 | 1_on |
| 6 | hybrid-linucb-l025 | 1_on |
| 6 | hybrid-linucb-l100 | 1_on |
| 6 | leastloaded | 1_on |
| 6 | sed | 1_on |
| 7 | hybrid-linucb | 1_on |
| 7 | hybrid-linucb-d-g098 | 1_on |
| 7 | hybrid-linucb-d-g098-l0 | 1_on |
| 7 | hybrid-linucb-d-g098-l025 | 1_on |
| 7 | hybrid-linucb-d-g098-l100 | 1_on |
| 7 | hybrid-linucb-l0 | 1_on |
| 7 | hybrid-linucb-l025 | 1_on |
| 7 | hybrid-linucb-l100 | 1_on |
| 7 | leastloaded | 1_on |
| 7 | sed | 1_on |
| 8 | hybrid-linucb | 1_on |
| 8 | hybrid-linucb-d-g098 | 1_on |
| 8 | hybrid-linucb-d-g098-l0 | 1_on |
| 8 | hybrid-linucb-d-g098-l025 | 1_on |
| 8 | hybrid-linucb-d-g098-l100 | 1_on |
| 8 | hybrid-linucb-l0 | 1_on |
| 8 | hybrid-linucb-l025 | 1_on |
| 8 | hybrid-linucb-l100 | 1_on |
| 8 | leastloaded | 1_on |
| 8 | sed | 1_on |
| 9 | hybrid-linucb | 1_on |
| 9 | hybrid-linucb-d-g098 | 1_on |
| 9 | hybrid-linucb-d-g098-l0 | 1_on |
| 9 | hybrid-linucb-d-g098-l025 | 1_on |
| 9 | hybrid-linucb-d-g098-l100 | 1_on |
| 9 | hybrid-linucb-l0 | 1_on |
| 9 | hybrid-linucb-l025 | 1_on |
| 9 | hybrid-linucb-l100 | 1_on |
| 9 | leastloaded | 1_on |
| 9 | sed | 1_on |
| 10 | hybrid-linucb | 1_on |
| 10 | hybrid-linucb-d-g098 | 1_on |
| 10 | hybrid-linucb-d-g098-l0 | 1_on |
| 10 | hybrid-linucb-d-g098-l025 | 1_on |
| 10 | hybrid-linucb-d-g098-l100 | 1_on |
| 10 | hybrid-linucb-l0 | 1_on |
| 10 | hybrid-linucb-l025 | 1_on |
| 10 | hybrid-linucb-l100 | 1_on |
| 10 | leastloaded | 1_on |
| 10 | sed | 1_on |
| 11 | hybrid-linucb | 1_on |
| 11 | hybrid-linucb-d-g098 | 1_on |
| 11 | hybrid-linucb-d-g098-l0 | 1_on |
| 11 | hybrid-linucb-d-g098-l025 | 1_on |
| 11 | hybrid-linucb-d-g098-l100 | 1_on |
| 11 | hybrid-linucb-l0 | 1_on |
| 11 | hybrid-linucb-l025 | 1_on |
| 11 | hybrid-linucb-l100 | 1_on |
| 11 | leastloaded | 1_on |
| 11 | sed | 1_on |
| 12 | hybrid-linucb | 1_on |
| 12 | hybrid-linucb-d-g098 | 1_on |
| 12 | hybrid-linucb-d-g098-l0 | 1_on |
| 12 | hybrid-linucb-d-g098-l025 | 1_on |
| 12 | hybrid-linucb-d-g098-l100 | 1_on |
| 12 | hybrid-linucb-l0 | 1_on |
| 12 | hybrid-linucb-l025 | 1_on |
| 12 | hybrid-linucb-l100 | 1_on |
| 12 | leastloaded | 1_on |
| 12 | sed | 1_on |
| 13 | hybrid-linucb | 1_on |
| 13 | hybrid-linucb-d-g098 | 1_on |
| 13 | hybrid-linucb-d-g098-l0 | 1_on |
| 13 | hybrid-linucb-d-g098-l025 | 1_on |
| 13 | hybrid-linucb-d-g098-l100 | 1_on |
| 13 | hybrid-linucb-l0 | 1_on |
| 13 | hybrid-linucb-l025 | 1_on |
| 13 | hybrid-linucb-l100 | 1_on |
| 13 | leastloaded | 1_on |
| 13 | sed | 1_on |
| 14 | hybrid-linucb | 1_on |
| 14 | hybrid-linucb-d-g098 | 1_on |
| 14 | hybrid-linucb-d-g098-l0 | 1_on |
| 14 | hybrid-linucb-d-g098-l025 | 1_on |
| 14 | hybrid-linucb-d-g098-l100 | 1_on |
| 14 | hybrid-linucb-l0 | 1_on |
| 14 | hybrid-linucb-l025 | 1_on |
| 14 | hybrid-linucb-l100 | 1_on |
| 14 | leastloaded | 1_on |
| 14 | sed | 1_on |
| 15 | hybrid-linucb | 1_on |
| 15 | hybrid-linucb-d-g098 | 1_on |
| 15 | hybrid-linucb-d-g098-l0 | 1_on |
| 15 | hybrid-linucb-d-g098-l025 | 1_on |
| 15 | hybrid-linucb-d-g098-l100 | 1_on |
| 15 | hybrid-linucb-l0 | 1_on |
| 15 | hybrid-linucb-l025 | 1_on |
| 15 | hybrid-linucb-l100 | 1_on |
| 15 | leastloaded | 1_on |
| 15 | sed | 1_on |
| 16 | hybrid-linucb | 1_on |
| 16 | hybrid-linucb-d-g098 | 1_on |
| 16 | hybrid-linucb-d-g098-l0 | 1_on |
| 16 | hybrid-linucb-d-g098-l025 | 1_on |
| 16 | hybrid-linucb-d-g098-l100 | 1_on |
| 16 | hybrid-linucb-l0 | 1_on |
| 16 | hybrid-linucb-l025 | 1_on |
| 16 | hybrid-linucb-l100 | 1_on |
| 16 | leastloaded | 1_on |
| 16 | sed | 1_on |
| 17 | hybrid-linucb | 1_on |
| 17 | hybrid-linucb-d-g098 | 1_on |
| 17 | hybrid-linucb-d-g098-l0 | 1_on |
| 17 | hybrid-linucb-d-g098-l025 | 1_on |
| 17 | hybrid-linucb-d-g098-l100 | 1_on |
| 17 | hybrid-linucb-l0 | 1_on |
| 17 | hybrid-linucb-l025 | 1_on |
| 17 | hybrid-linucb-l100 | 1_on |
| 17 | leastloaded | 1_on |
| 17 | sed | 1_on |
| 18 | hybrid-linucb | 1_on |
| 18 | hybrid-linucb-d-g098 | 1_on |
| 18 | hybrid-linucb-d-g098-l0 | 1_on |
| 18 | hybrid-linucb-d-g098-l025 | 1_on |
| 18 | hybrid-linucb-d-g098-l100 | 1_on |
| 18 | hybrid-linucb-l0 | 1_on |
| 18 | hybrid-linucb-l025 | 1_on |
| 18 | hybrid-linucb-l100 | 1_on |
| 18 | leastloaded | 1_on |
| 18 | sed | 1_on |
| 19 | hybrid-linucb | 1_on |
| 19 | hybrid-linucb-d-g098 | 1_on |
| 19 | hybrid-linucb-d-g098-l0 | 1_on |
| 19 | hybrid-linucb-d-g098-l025 | 1_on |
| 19 | hybrid-linucb-d-g098-l100 | 1_on |
| 19 | hybrid-linucb-l0 | 1_on |
| 19 | hybrid-linucb-l025 | 1_on |
| 19 | hybrid-linucb-l100 | 1_on |
| 19 | leastloaded | 1_on |
| 19 | sed | 1_on |
| 20 | hybrid-linucb | 1_on |
| 20 | hybrid-linucb-d-g098 | 1_on |
| 20 | hybrid-linucb-d-g098-l0 | 1_on |
| 20 | hybrid-linucb-d-g098-l025 | 1_on |
| 20 | hybrid-linucb-d-g098-l100 | 1_on |
| 20 | hybrid-linucb-l0 | 1_on |
| 20 | hybrid-linucb-l025 | 1_on |
| 20 | hybrid-linucb-l100 | 1_on |
| 20 | leastloaded | 1_on |
| 20 | sed | 1_on |

## makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 185.2 | 20 |
| hybrid-linucb-d-g098 | 188.8 | 20 |
| hybrid-linucb-d-g098-l0 | 193.4 | 20 |
| hybrid-linucb-d-g098-l025 | 189.3 | 20 |
| hybrid-linucb-d-g098-l100 | 188.4 | 20 |
| hybrid-linucb-l0 | 188.7 | 20 |
| hybrid-linucb-l025 | 186.1 | 20 |
| hybrid-linucb-l100 | 186.3 | 20 |
| leastloaded | 192.1 | 20 |
| sed | 186.6 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -6.93 [-8.707, -4.074] | 2.67e-05 | 0.0003738 | -0.9429 | 0 |
| hybrid-linucb | sed | 20 | -0.816 [-3.712, 1.635] | 0.3488 | 1 | -0.2476 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -1.995 [-6.591, -0.9355] | 0.001432 | 0.01719 | -0.7714 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.353 [0.735, 4.788] | 0.03277 | 0.1966 | 0.5429 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.98 [-0.8555, 3.637] | 0.3118 | 1 | 0.2667 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 7.584 [4.971, 9.73] | 1.335e-05 | 0.0002136 | 0.9619 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -3.032 [-5.468, -0.8805] | 0.00639 | 0.0639 | -0.6762 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 2.239 [-0.015, 5.929] | 0.01208 | 0.09663 | 0.6286 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -3.433 [-5.341, 0.27] | 0.005581 | 0.06139 | -0.6857 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 2.16 [-1.016, 6.253] | 0.05317 | 0.2658 | 0.4952 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | -4.088 [-5.255, -2.188] | 0.01069 | 0.0962 | -0.6381 | 0 |
| hybrid-linucb-l0 | sed | 20 | 2.837 [-0.0815, 5.886] | 0.01923 | 0.1346 | 0.5905 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | -6.036 [-6.957, -3.696] | 1.907e-05 | 0.0002861 | -0.9524 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.901 [-2.45, 1.614] | 0.7285 | 1 | -0.09524 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | -6.125 [-8.27, -2.025] | 0.0003948 | 0.005133 | -0.8381 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.2035 [-3.002, 3.394] | 1 | 1 | 0 | 0 |

## build_1_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 90.91 | 20 |
| hybrid-linucb-d-g098 | 92.11 | 20 |
| hybrid-linucb-d-g098-l0 | 94 | 20 |
| hybrid-linucb-d-g098-l025 | 92.38 | 20 |
| hybrid-linucb-d-g098-l100 | 92.24 | 20 |
| hybrid-linucb-l0 | 92.65 | 20 |
| hybrid-linucb-l025 | 92.1 | 20 |
| hybrid-linucb-l100 | 91.3 | 20 |
| leastloaded | 94.81 | 20 |
| sed | 91.42 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -2.21 [-5.675, -1.077] | 0.0007076 | 0.01132 | -0.8095 | 0 |
| hybrid-linucb | sed | 20 | -0.788 [-2.523, 0.54] | 0.2024 | 1 | -0.3333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -2.419 [-3.972, 0.0555] | 0.004221 | 0.05909 | -0.7048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.3835 [-1.68, 1.524] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | -1.105 [-1.697, 1.647] | 0.7012 | 1 | -0.1048 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 1.921 [0.338, 3.662] | 0.003654 | 0.05482 | 0.7143 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -1.374 [-3.313, 0.222] | 0.05317 | 0.5317 | -0.4952 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.1115 [-2.117, 2.576] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -1.404 [-4.327, 0.069] | 0.02664 | 0.2931 | -0.5619 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.704 [-2.661, 3.273] | 0.4524 | 1 | 0.2 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | -1.73 [-3.385, -0.1705] | 0.1054 | 0.9486 | -0.419 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.374 [-0.9115, 2.43] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | -3.081 [-4.275, 0.0465] | 0.01069 | 0.139 | -0.6381 | 0 |
| hybrid-linucb-l025 | sed | 20 | 0.598 [-1.698, 1.488] | 0.896 | 1 | 0.03333 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | -1.779 [-4.667, -0.1865] | 0.02395 | 0.2874 | -0.5714 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.809 [-1.851, 1.581] | 0.7841 | 1 | -0.07619 | 0 |

## build_2_makespan_s

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 93.8 | 20 |
| hybrid-linucb-d-g098 | 96.36 | 20 |
| hybrid-linucb-d-g098-l0 | 99.78 | 20 |
| hybrid-linucb-d-g098-l025 | 96.27 | 20 |
| hybrid-linucb-d-g098-l100 | 95.59 | 20 |
| hybrid-linucb-l0 | 96.25 | 20 |
| hybrid-linucb-l025 | 93.53 | 20 |
| hybrid-linucb-l100 | 93.7 | 20 |
| leastloaded | 98.2 | 20 |
| sed | 94.02 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | -3.757 [-4.946, -2.351] | 1.907e-05 | 0.0002861 | -0.9524 | 0 |
| hybrid-linucb | sed | 20 | 0.0635 [-1.058, 1.291] | 0.9563 | 1 | 0.01905 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.916 [-3.015, -0.233] | 0.03623 | 0.1812 | -0.5333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 2.811 [1.331, 3.207] | 0.004221 | 0.04221 | 0.7048 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 1.575 [0.6585, 2.591] | 0.005581 | 0.04465 | 0.6857 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 6.096 [2.338, 6.724] | 5.722e-06 | 9.155e-05 | 0.981 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -0.311 [-2.727, 0.6435] | 0.2455 | 0.9819 | -0.3048 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 1.926 [1.14, 3.165] | 0.0001049 | 0.001364 | 0.8952 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -1.6 [-2.697, -0.2125] | 0.004221 | 0.04221 | -0.7048 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 2.103 [0.478, 3.546] | 0.00639 | 0.04473 | 0.6762 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | -1.178 [-2.718, -0.2795] | 0.02148 | 0.1289 | -0.581 | 0 |
| hybrid-linucb-l0 | sed | 20 | 1.987 [0.853, 2.77] | 0.001209 | 0.0133 | 0.781 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | -4.141 [-5.135, -1.738] | 6.294e-05 | 0.0008812 | -0.9143 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.1395 [-1.107, 1.195] | 0.9563 | 1 | -0.01905 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | -3.533 [-5.194, -1.503] | 0.0001335 | 0.001602 | -0.8857 | 0 |
| hybrid-linucb-l100 | sed | 20 | 0.3265 [-0.425, 1.419] | 0.4749 | 1 | 0.1905 | 0 |

## skip_idle_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | 0.0822 | 20 |
| hybrid-linucb-d-g098-l025 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | 0 | 20 |
| hybrid-linucb-l0 | 0.02034 | 20 |
| hybrid-linucb-l025 | 0 | 20 |
| hybrid-linucb-l100 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.07373 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.07373 [-0.08051, -0.07034] | 8.782e-05 | 0.001405 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.1025 | 0.5124 | 1 | 17 |
| hybrid-linucb-d-g098 | sed | 20 | -0.07373 [-0.08051, -0.06949] | 8.807e-05 | 0.001405 | -1 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.0822 [0.06525, 0.09831] | 8.82e-05 | 0.001405 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.01271 [-0.01695, 0.03136] | 0.4221 | 1 | 0.2048 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0.002542] | 0.01576 | 0.1261 | 1 | 13 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.07288 [-0.08051, -0.0678] | 8.807e-05 | 0.001405 | -1 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [0, 0] | 0.08326 | 0.4996 | 1 | 17 |
| hybrid-linucb-d-g098-l100 | sed | 20 | -0.07373 [-0.07966, -0.06949] | 8.807e-05 | 0.001405 | -1 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.02034 [0.01441, 0.03644] | 8.832e-05 | 0.001405 | 1 | 0 |
| hybrid-linucb-l0 | sed | 20 | -0.05424 [-0.0661, -0.02712] | 0.02275 | 0.1592 | -0.581 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | sed | 20 | -0.07373 [-0.08051, -0.07034] | 8.782e-05 | 0.001405 | -1 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | sed | 20 | -0.07373 [-0.08051, -0.07034] | 8.782e-05 | 0.001405 | -1 | 0 |

## skip_idle_given_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | 0.1522 | 20 |
| hybrid-linucb-d-g098-l025 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | 0 | 20 |
| hybrid-linucb-l0 | 0.04253 | 20 |
| hybrid-linucb-l025 | 0 | 20 |
| hybrid-linucb-l100 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0.1666 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | -0.1666 [-0.1871, -0.1455] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 0.25 | 1 | 1 | 17 |
| hybrid-linucb-d-g098 | sed | 20 | -0.1666 [-0.1871, -0.1455] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.1522 [0.1295, 0.1753] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | -0.0001061 [-0.05008, 0.02676] | 0.4304 | 1 | -0.2095 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0.005047] | 0.01562 | 0.1094 | 1 | 13 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.1666 [-0.1854, -0.1405] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [0, 0] | 0.25 | 1 | 1 | 17 |
| hybrid-linucb-d-g098-l100 | sed | 20 | -0.1651 [-0.1859, -0.1436] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.04253 [0.0314, 0.07747] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-l0 | sed | 20 | -0.1224 [-0.1437, -0.06728] | 0.00486 | 0.03888 | -0.6952 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | sed | 20 | -0.1666 [-0.1871, -0.1455] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | sed | 20 | -0.1666 [-0.1871, -0.1455] | 1.907e-06 | 3.052e-05 | -1 | 0 |

## strongest_idle

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.9529 | 20 |
| hybrid-linucb-d-g098 | 0.9117 | 20 |
| hybrid-linucb-d-g098-l0 | 0.7784 | 20 |
| hybrid-linucb-d-g098-l025 | 0.906 | 20 |
| hybrid-linucb-d-g098-l100 | 0.9258 | 20 |
| hybrid-linucb-l0 | 0.8977 | 20 |
| hybrid-linucb-l025 | 0.9518 | 20 |
| hybrid-linucb-l100 | 0.9457 | 20 |
| leastloaded | 0.9235 | 20 |
| sed | 0.8334 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.02233 [0.007642, 0.04893] | 0.002712 | 0.01627 | 0.7333 | 0 |
| hybrid-linucb | sed | 20 | 0.1219 [0.09376, 0.1399] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.006713 [-0.02148, 0.00696] | 0.2024 | 0.4287 | -0.3333 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.07645 [0.06316, 0.09703] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | -0.1508 [-0.1762, -0.1218] | 1.907e-06 | 3.052e-05 | -1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | -0.0653 [-0.12, -0.01593] | 0.0003223 | 0.002579 | -0.8476 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -0.01587 [-0.02942, 0.003871] | 0.1429 | 0.4287 | -0.381 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.07654 [0.05505, 0.09982] | 4.768e-05 | 0.0004768 | 0.9238 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -0.005525 [-0.01795, 0.01345] | 0.8695 | 0.8695 | -0.04762 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.07835 [0.06467, 0.105] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | -0.01491 [-0.08516, -9.766e-05] | 0.03277 | 0.1638 | -0.5429 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.05878 [0.00704, 0.08595] | 0.07585 | 0.3034 | 0.4571 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.02202 [0.008022, 0.04314] | 4.768e-05 | 0.0004768 | 0.9238 | 0 |
| hybrid-linucb-l025 | sed | 20 | 0.107 [0.09544, 0.1447] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.02302 [0.002813, 0.04127] | 0.001432 | 0.01003 | 0.7714 | 0 |
| hybrid-linucb-l100 | sed | 20 | 0.1157 [0.08822, 0.1376] | 1.907e-06 | 3.052e-05 | 1 | 0 |

## free_slot_skip_all

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | 0 | 20 |
| hybrid-linucb-l0 | 0 | 20 |
| hybrid-linucb-l025 | 0 | 20 |
| hybrid-linucb-l100 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [0, 0] | 0.1573 | 1 | 1 | 18 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0 [0, 0] | 0.1573 | 1 | 1 | 18 |
| hybrid-linucb-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## free_slot_skip_given_free

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0 | 20 |
| hybrid-linucb-d-g098 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | 0 | 20 |
| hybrid-linucb-l0 | 0 | 20 |
| hybrid-linucb-l025 | 0 | 20 |
| hybrid-linucb-l100 | 0 | 20 |
| leastloaded | 0 | 20 |
| sed | 0 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [0, 0] | 0.1573 | 1 | 1 | 18 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0 [0, 0] | 0.1573 | 1 | 1 | 18 |
| hybrid-linucb-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 1 | 19 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## strongest_free_slot

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2144 | 20 |
| hybrid-linucb-d-g098 | 0.222 | 20 |
| hybrid-linucb-d-g098-l0 | 0.3 | 20 |
| hybrid-linucb-d-g098-l025 | 0.228 | 20 |
| hybrid-linucb-d-g098-l100 | 0.2297 | 20 |
| hybrid-linucb-l0 | 0.3042 | 20 |
| hybrid-linucb-l025 | 0.2297 | 20 |
| hybrid-linucb-l100 | 0.2051 | 20 |
| leastloaded | 0.2042 | 20 |
| sed | 0.2093 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.01864 [0, 0.04407] | 0.01303 | 0.1303 | 0.6333 | 0 |
| hybrid-linucb | sed | 20 | 0.01695 [-0.009322, 0.03814] | 0.08249 | 0.4949 | 0.4429 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01949 [0.007627, 0.02542] | 0.03034 | 0.2124 | 0.5524 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.001695 [-0.01525, 0.03729] | 0.3869 | 1 | 0.2263 | 1 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.1051 [0.07966, 0.1195] | 8.845e-05 | 0.001238 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.09492 [0.07288, 0.1059] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.03136 [0.01017, 0.05424] | 0.0002098 | 0.002518 | 0.8667 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.02542 [0.007627, 0.0339] | 0.006541 | 0.07195 | 0.7105 | 1 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.0178 [-0.005085, 0.04746] | 0.02041 | 0.1633 | 0.6 | 1 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.01695 [-0.002542, 0.02966] | 0.09551 | 0.4949 | 0.4421 | 1 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.1008 [0.08983, 0.1347] | 0.0001298 | 0.001687 | 0.9762 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.1102 [0.07627, 0.1305] | 1.907e-05 | 0.0002861 | 0.9524 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.02119 [0.002542, 0.05678] | 0.01332 | 0.1303 | 0.6474 | 1 |
| hybrid-linucb-l025 | sed | 20 | 0.02203 [-0.0178, 0.04322] | 0.1353 | 0.5412 | 0.381 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.002542 [-0.01356, 0.0322] | 0.5256 | 1 | 0.1619 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.01356 [-0.02797, 0.03644] | 0.7628 | 1 | -0.07895 | 1 |

## had_idle_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.422 | 20 |
| hybrid-linucb-d-g098 | 0.4585 | 20 |
| hybrid-linucb-d-g098-l0 | 0.5449 | 20 |
| hybrid-linucb-d-g098-l025 | 0.4822 | 20 |
| hybrid-linucb-d-g098-l100 | 0.4602 | 20 |
| hybrid-linucb-l0 | 0.4932 | 20 |
| hybrid-linucb-l025 | 0.4237 | 20 |
| hybrid-linucb-l100 | 0.4195 | 20 |
| leastloaded | 0.4237 | 20 |
| sed | 0.461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.005085 [-0.02627, 0.01695] | 0.7227 | 1 | -0.09048 | 0 |
| hybrid-linucb | sed | 20 | -0.0339 [-0.07288, -0.01186] | 0.01603 | 0.1443 | -0.6143 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.01949 [-0.001695, 0.05763] | 0.04187 | 0.2931 | 0.519 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01949 [-0.04746, 0.02712] | 0.3701 | 1 | -0.2286 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.1169 [0.07881, 0.1627] | 8.845e-05 | 0.001327 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.07797 [0.05763, 0.1169] | 4.768e-05 | 0.0007629 | 0.9238 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.06186 [0.04322, 0.08051] | 0.0001678 | 0.002182 | 0.8762 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.01017 [-0.009322, 0.04153] | 0.08143 | 0.4886 | 0.4737 | 2 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.0339 [0.01441, 0.05424] | 0.0008909 | 0.01069 | 0.8476 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.004237 [-0.02203, 0.03136] | 0.9108 | 1 | 0.02857 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.06441 [0.02542, 0.0839] | 0.0001398 | 0.001957 | 0.9714 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.03898 [-0.01271, 0.06186] | 0.0263 | 0.2104 | 0.5667 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | -0.01186 [-0.0322, 0.007627] | 0.4329 | 1 | -0.2 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.04407 [-0.07881, -0.00678] | 0.008461 | 0.08461 | -0.6714 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.0008475 [-0.02627, 0.01695] | 0.5882 | 1 | -0.1381 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.03559 [-0.07881, -0.01102] | 0.00299 | 0.03289 | -0.7571 | 0 |

## had_free_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 1 | 20 |
| hybrid-linucb-d-g098 | 1 | 20 |
| hybrid-linucb-d-g098-l0 | 1 | 20 |
| hybrid-linucb-d-g098-l025 | 1 | 20 |
| hybrid-linucb-d-g098-l100 | 1 | 20 |
| hybrid-linucb-l0 | 1 | 20 |
| hybrid-linucb-l025 | 1 | 20 |
| hybrid-linucb-l100 | 1 | 20 |
| leastloaded | 1 | 20 |
| sed | 1 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l0 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l0 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l025 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |
| hybrid-linucb-l100 | sed | 20 | 0 [0, 0] | 1 | 1 | 0 | 20 |

## drift_pre_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2931 | 20 |
| hybrid-linucb-d-g098 | 0.2583 | 20 |
| hybrid-linucb-d-g098-l0 | 0.3391 | 20 |
| hybrid-linucb-d-g098-l025 | 0.2106 | 20 |
| hybrid-linucb-d-g098-l100 | 0.3393 | 20 |
| hybrid-linucb-l0 | 0.3443 | 20 |
| hybrid-linucb-l025 | 0.2586 | 20 |
| hybrid-linucb-l100 | 0.3498 | 20 |
| leastloaded | 0.2069 | 20 |
| sed | 0.2034 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.1034 [-0.08092, 0.2211] | 0.06412 | 0.7695 | 0.4842 | 1 |
| hybrid-linucb | sed | 20 | 0.1724 [0.05271, 0.2096] | 0.03332 | 0.4999 | 0.5429 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.02151 [-0.07623, 0.08448] | 0.7323 | 1 | 0.08947 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 0.09598 [0.01055, 0.1749] | 0.1212 | 1 | 0.3952 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.1404 [-0.007879, 0.2586] | 0.09291 | 1 | 0.4286 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.06724 [-0.01416, 0.2429] | 0.0382 | 0.5348 | 0.5421 | 1 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.08005 [-0.08826, 0.1523] | 0.2958 | 1 | 0.2667 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.05599 [-0.105, 0.2069] | 0.2413 | 1 | 0.3158 | 1 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.0883 [-0.1678, 0.2392] | 0.3884 | 1 | 0.2286 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.01379 [-0.06589, 0.29] | 0.2432 | 1 | 0.3053 | 1 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.1734 [-0.0003719, 0.3064] | 0.04405 | 0.5727 | 0.5143 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.2042 [0.01237, 0.3103] | 0.01866 | 0.2986 | 0.6 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.03448 [-0.1059, 0.1755] | 0.4723 | 1 | 0.193 | 2 |
| hybrid-linucb-l025 | sed | 20 | 0.101 [-0.06641, 0.1853] | 0.2789 | 1 | 0.2762 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.09622 [-0.1444, 0.2759] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-l100 | sed | 20 | 0.1107 [0, 0.1712] | 0.1415 | 1 | 0.4035 | 2 |

## drift_pre_all_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2752 | 20 |
| hybrid-linucb-d-g098 | 0.249 | 20 |
| hybrid-linucb-d-g098-l0 | 0.2763 | 20 |
| hybrid-linucb-d-g098-l025 | 0.2278 | 20 |
| hybrid-linucb-d-g098-l100 | 0.3008 | 20 |
| hybrid-linucb-l0 | 0.2879 | 20 |
| hybrid-linucb-l025 | 0.2471 | 20 |
| hybrid-linucb-l100 | 0.2664 | 20 |
| leastloaded | 0.2171 | 20 |
| sed | 0.2664 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.04447 [-0.02038, 0.08527] | 0.2179 | 1 | 0.3143 | 0 |
| hybrid-linucb | sed | 20 | 0.02286 [-0.03933, 0.07479] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.006318 [-0.02869, 0.04059] | 0.7562 | 1 | 0.08571 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.01667 [-0.01414, 0.05426] | 0.6726 | 1 | 0.1105 | 1 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.06756 [-0.06646, 0.1491] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.01005 [-0.05321, 0.09048] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.01873 [-0.04246, 0.08033] | 0.5958 | 1 | 0.1429 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.006662 [-0.05483, 0.02636] | 0.6507 | 1 | -0.1263 | 1 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.06952 [-0.04772, 0.1192] | 0.1054 | 1 | 0.419 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.03002 [-0.05079, 0.109] | 0.33 | 1 | 0.2571 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.06032 [-0.004592, 0.1447] | 0.08743 | 1 | 0.4526 | 1 |
| hybrid-linucb-l0 | sed | 20 | 0.07462 [-0.04016, 0.1067] | 0.1257 | 1 | 0.3905 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.03488 [-0.02372, 0.07752] | 0.3488 | 1 | 0.2476 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.008509 [-0.04389, 0.05921] | 0.8083 | 1 | -0.0619 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.03763 [-0.1139, 0.1349] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.00875 [-0.05148, 0.06589] | 0.9563 | 1 | -0.01905 | 0 |

## drift_post_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2148 | 20 |
| hybrid-linucb-d-g098 | 0.1924 | 20 |
| hybrid-linucb-d-g098-l0 | 0.1835 | 20 |
| hybrid-linucb-d-g098-l025 | 0.1855 | 20 |
| hybrid-linucb-d-g098-l100 | 0.2078 | 20 |
| hybrid-linucb-l0 | 0.1885 | 20 |
| hybrid-linucb-l025 | 0.205 | 20 |
| hybrid-linucb-l100 | 0.1911 | 20 |
| leastloaded | 0.1866 | 20 |
| sed | 0.2017 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.0154 [-0.01469, 0.04036] | 0.07283 | 1 | 0.4737 | 1 |
| hybrid-linucb | sed | 20 | 0.01166 [0.001818, 0.02118] | 0.1119 | 1 | 0.4158 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.0006588 [-0.01642, 0.01302] | 0.8596 | 1 | 0.05263 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -0.005418 [-0.01957, 0.02039] | 0.6676 | 1 | -0.1095 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.009594 [-0.02212, 0.02001] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | -0.006906 [-0.01816, 0.001993] | 0.2873 | 1 | -0.2714 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -0.01281 [-0.02083, 0.01516] | 0.5706 | 1 | -0.1524 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.006684 [-0.02084, 0.01145] | 0.2162 | 1 | -0.3238 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.01061 [-0.007359, 0.03559] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.005233 [-0.01302, 0.02979] | 0.5509 | 1 | 0.1696 | 2 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.01207 [-0.01848, 0.02712] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.001965 [-0.01883, 0.0232] | 0.8906 | 1 | -0.04211 | 1 |
| hybrid-linucb-l025 | leastloaded | 20 | -0.001944 [-0.01433, 0.03164] | 0.5459 | 1 | 0.1619 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.001316 [-0.02603, 0.02061] | 1 | 1 | 0 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.00175 [-0.01136, 0.03452] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.006087 [-0.03427, 0.02671] | 0.7938 | 1 | -0.06667 | 0 |

## drift_pre_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 29 | 20 |
| hybrid-linucb-d-g098 | 29 | 20 |
| hybrid-linucb-d-g098-l0 | 29 | 20 |
| hybrid-linucb-d-g098-l025 | 29 | 20 |
| hybrid-linucb-d-g098-l100 | 28.5 | 20 |
| hybrid-linucb-l0 | 29 | 20 |
| hybrid-linucb-l025 | 29 | 20 |
| hybrid-linucb-l100 | 29 | 20 |
| leastloaded | 29 | 20 |
| sed | 29 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [-0.5, 0] | 0.7146 | 1 | -0.1273 | 10 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 0] | 0.4926 | 1 | -0.2444 | 11 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.9043 | 1 | -0.03846 | 8 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 0.7815 | 1 | -0.09091 | 10 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [-1, 0.5] | 0.8157 | 1 | -0.06667 | 5 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [-1, 0] | 0.9145 | 1 | 0.03297 | 7 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [-0.5, 0] | 0.5222 | 1 | -0.2121 | 9 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [-1, 0] | 0.4386 | 1 | -0.2308 | 8 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -0.5 [-1, 0] | 0.03235 | 0.5176 | -0.6374 | 7 |
| hybrid-linucb-d-g098-l100 | sed | 20 | -1 [-1, 0] | 0.04911 | 0.7367 | -0.5333 | 5 |
| hybrid-linucb-l0 | leastloaded | 20 | -1 [-1, 1] | 0.7324 | 1 | -0.08772 | 2 |
| hybrid-linucb-l0 | sed | 20 | 0 [-1, 0.5] | 0.7971 | 1 | -0.07619 | 6 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.4669 | 1 | -0.2308 | 8 |
| hybrid-linucb-l025 | sed | 20 | 0 [-1, 0] | 0.4989 | 1 | -0.1978 | 7 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [-1, 1] | 0.7702 | 1 | 0.08088 | 4 |
| hybrid-linucb-l100 | sed | 20 | 0 [-1, 0.5] | 0.7436 | 1 | 0.0989 | 7 |

## drift_pre_all_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 129 | 20 |
| hybrid-linucb-d-g098 | 129 | 20 |
| hybrid-linucb-d-g098-l0 | 129 | 20 |
| hybrid-linucb-d-g098-l025 | 129 | 20 |
| hybrid-linucb-d-g098-l100 | 128.5 | 20 |
| hybrid-linucb-l0 | 129 | 20 |
| hybrid-linucb-l025 | 129 | 20 |
| hybrid-linucb-l100 | 129 | 20 |
| leastloaded | 129 | 20 |
| sed | 129 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [-0.5, 0] | 0.7146 | 1 | -0.1273 | 10 |
| hybrid-linucb | sed | 20 | 0 [-0.5, 0] | 0.4926 | 1 | -0.2444 | 11 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.9043 | 1 | -0.03846 | 8 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 0.7815 | 1 | -0.09091 | 10 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [-1, 0.5] | 0.8157 | 1 | -0.06667 | 5 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [-1, 0] | 0.9145 | 1 | 0.03297 | 7 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [-0.5, 0] | 0.5222 | 1 | -0.2121 | 9 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [-1, 0] | 0.4386 | 1 | -0.2308 | 8 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | -0.5 [-1, 0] | 0.03235 | 0.5176 | -0.6374 | 7 |
| hybrid-linucb-d-g098-l100 | sed | 20 | -1 [-1, 0] | 0.04911 | 0.7367 | -0.5333 | 5 |
| hybrid-linucb-l0 | leastloaded | 20 | -1 [-1, 1] | 0.7324 | 1 | -0.08772 | 2 |
| hybrid-linucb-l0 | sed | 20 | 0 [-1, 0.5] | 0.7971 | 1 | -0.07619 | 6 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.4669 | 1 | -0.2308 | 8 |
| hybrid-linucb-l025 | sed | 20 | 0 [-1, 0] | 0.4989 | 1 | -0.1978 | 7 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [-1, 1] | 0.7702 | 1 | 0.08088 | 4 |
| hybrid-linucb-l100 | sed | 20 | 0 [-1, 0.5] | 0.7436 | 1 | 0.0989 | 7 |

## drift_post_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 461 | 20 |
| hybrid-linucb-d-g098 | 461 | 20 |
| hybrid-linucb-d-g098-l0 | 461 | 20 |
| hybrid-linucb-d-g098-l025 | 461 | 20 |
| hybrid-linucb-d-g098-l100 | 461.5 | 20 |
| hybrid-linucb-l0 | 461 | 20 |
| hybrid-linucb-l025 | 461 | 20 |
| hybrid-linucb-l100 | 461 | 20 |
| leastloaded | 461 | 20 |
| sed | 461 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0 [0, 0.5] | 0.7146 | 1 | 0.1273 | 10 |
| hybrid-linucb | sed | 20 | 0 [0, 0.5] | 0.4926 | 1 | 0.2444 | 11 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.9043 | 1 | 0.03846 | 8 |
| hybrid-linucb-d-g098 | sed | 20 | 0 [0, 0] | 0.7815 | 1 | 0.09091 | 10 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0 [-0.5, 1] | 0.8157 | 1 | 0.06667 | 5 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0 [0, 1] | 0.9145 | 1 | -0.03297 | 7 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0 [0, 0.5] | 0.5222 | 1 | 0.2121 | 9 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0 [0, 1] | 0.4386 | 1 | 0.2308 | 8 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.5 [0, 1] | 0.03235 | 0.5176 | 0.6374 | 7 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 1 [0, 1] | 0.04911 | 0.7367 | 0.5333 | 5 |
| hybrid-linucb-l0 | leastloaded | 20 | 1 [-1, 1] | 0.7324 | 1 | 0.08772 | 2 |
| hybrid-linucb-l0 | sed | 20 | 0 [-0.5, 1] | 0.7971 | 1 | 0.07619 | 6 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [-0.5, 0.5] | 0.4669 | 1 | 0.2308 | 8 |
| hybrid-linucb-l025 | sed | 20 | 0 [0, 1] | 0.4989 | 1 | 0.1978 | 7 |
| hybrid-linucb-l100 | leastloaded | 20 | 0 [-1, 1] | 0.7702 | 1 | -0.08088 | 4 |
| hybrid-linucb-l100 | sed | 20 | 0 [-0.5, 1] | 0.7436 | 1 | -0.0989 | 7 |

## drift_post_early_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.23 | 20 |
| hybrid-linucb-d-g098 | 0.21 | 20 |
| hybrid-linucb-d-g098-l0 | 0.24 | 20 |
| hybrid-linucb-d-g098-l025 | 0.2 | 20 |
| hybrid-linucb-d-g098-l100 | 0.2 | 20 |
| hybrid-linucb-l0 | 0.25 | 20 |
| hybrid-linucb-l025 | 0.16 | 20 |
| hybrid-linucb-l100 | 0.19 | 20 |
| leastloaded | 0.16 | 20 |
| sed | 0.15 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.07 [-0.02, 0.12] | 0.06981 | 0.6693 | 0.4619 | 0 |
| hybrid-linucb | sed | 20 | 0.07 [0, 0.13] | 0.01299 | 0.1818 | 0.6667 | 2 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.06 [-0.02, 0.11] | 0.06693 | 0.6693 | 0.4789 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 0.1 [0.01, 0.11] | 0.02166 | 0.2599 | 0.6 | 1 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.08 [-0.02, 0.13] | 0.03649 | 0.4014 | 0.5614 | 2 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.09 [0.01, 0.14] | 0.01757 | 0.2284 | 0.6374 | 2 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.04 [-0.03, 0.09] | 0.2857 | 1 | 0.2789 | 1 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.05 [-0.06, 0.11] | 0.2392 | 1 | 0.3 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.05 [-0.07, 0.1] | 0.344 | 1 | 0.2474 | 1 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.03 [0, 0.11] | 0.08822 | 0.7057 | 0.4706 | 3 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.08 [0.03, 0.14] | 0.007974 | 0.1196 | 0.6762 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.04 [0.01, 0.17] | 0.003053 | 0.04885 | 0.817 | 3 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [-0.08, 0.04] | 0.887 | 1 | -0.03922 | 3 |
| hybrid-linucb-l025 | sed | 20 | 0.03 [-0.04, 0.09] | 0.4438 | 1 | 0.1952 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.06 [-0.07, 0.11] | 0.4551 | 1 | 0.1905 | 0 |
| hybrid-linucb-l100 | sed | 20 | 0.08 [-0.05, 0.15] | 0.1712 | 1 | 0.3579 | 1 |

## drift_post_late_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2202 | 20 |
| hybrid-linucb-d-g098 | 0.1876 | 20 |
| hybrid-linucb-d-g098-l0 | 0.18 | 20 |
| hybrid-linucb-d-g098-l025 | 0.1839 | 20 |
| hybrid-linucb-d-g098-l100 | 0.2063 | 20 |
| hybrid-linucb-l0 | 0.1788 | 20 |
| hybrid-linucb-l025 | 0.2022 | 20 |
| hybrid-linucb-l100 | 0.1934 | 20 |
| leastloaded | 0.1912 | 20 |
| sed | 0.1995 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.01264 [-0.01703, 0.04477] | 0.2162 | 1 | 0.3238 | 0 |
| hybrid-linucb | sed | 20 | 0.01307 [-0.01582, 0.02407] | 0.3603 | 1 | 0.2333 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | -0.001698 [-0.01558, 0.008516] | 0.7012 | 1 | -0.1048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | -0.01076 [-0.02436, 0.004866] | 0.165 | 1 | -0.3619 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | -0.001689 [-0.02633, 0.01071] | 0.6215 | 1 | -0.1333 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | -0.01281 [-0.0221, -0.002678] | 0.04187 | 0.6699 | -0.519 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -0.01359 [-0.02511, -0.0007158] | 0.3321 | 1 | -0.2632 | 1 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.02117 [-0.03136, 0.004657] | 0.08936 | 1 | -0.4333 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.01459 [-0.02023, 0.04651] | 0.4524 | 1 | 0.2 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.0009656 [-0.02303, 0.03583] | 0.9217 | 1 | 0.03158 | 1 |
| hybrid-linucb-l0 | leastloaded | 20 | -0.005153 [-0.01629, 0.01751] | 0.8983 | 1 | -0.0381 | 0 |
| hybrid-linucb-l0 | sed | 20 | -0.01254 [-0.02837, 0.01073] | 0.2455 | 1 | -0.3048 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.002881 [-0.01312, 0.03528] | 0.5217 | 1 | 0.1714 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.0005282 [-0.02555, 0.02798] | 0.9702 | 1 | -0.009524 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.009511 [-0.01603, 0.02344] | 0.5706 | 1 | 0.1524 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.01174 [-0.02459, 0.01292] | 0.4749 | 1 | -0.1905 | 0 |

## excess_target_share_pre

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.007333 | 20 |
| hybrid-linucb-d-g098 | -0.1568 | 20 |
| hybrid-linucb-d-g098-l0 | 0.01234 | 20 |
| hybrid-linucb-d-g098-l025 | -0.04036 | 20 |
| hybrid-linucb-d-g098-l100 | -0.1474 | 20 |
| hybrid-linucb-l0 | 0.1301 | 20 |
| hybrid-linucb-l025 | 0.04992 | 20 |
| hybrid-linucb-l100 | -0.1075 | 20 |
| leastloaded | -0.2551 | 20 |
| sed | -0.2151 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.2562 [0.1263, 0.3251] | 0.0001049 | 0.001154 | 0.8952 | 0 |
| hybrid-linucb | sed | 20 | 0.1966 [0.1012, 0.2286] | 4.768e-05 | 0.0005722 | 0.9238 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.1047 [0.04566, 0.1641] | 0.004221 | 0.02955 | 0.7048 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.1465 [-0.03853, 0.2571] | 0.06958 | 0.2783 | 0.4667 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.2527 [0.1103, 0.3531] | 3.624e-05 | 0.0004711 | 0.9333 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.2121 [0.152, 0.3452] | 0.0003223 | 0.003223 | 0.8476 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.1261 [0.04176, 0.2645] | 0.0007076 | 0.006369 | 0.8095 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.1695 [0.03758, 0.2328] | 0.02148 | 0.1289 | 0.581 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.09385 [-0.04483, 0.2252] | 0.07585 | 0.2783 | 0.4571 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.008722 [-0.05489, 0.1931] | 0.3488 | 0.3488 | 0.2476 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.3575 [0.2325, 0.5081] | 5.722e-06 | 9.155e-05 | 0.981 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.3363 [0.2332, 0.4139] | 1.907e-05 | 0.0002861 | 0.9524 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.2687 [0.122, 0.3599] | 2.67e-05 | 0.0003738 | 0.9429 | 0 |
| hybrid-linucb-l025 | sed | 20 | 0.2815 [0.1223, 0.3409] | 0.001432 | 0.01146 | 0.7714 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.1302 [-0.06231, 0.2841] | 0.04844 | 0.2422 | 0.5048 | 0 |
| hybrid-linucb-l100 | sed | 20 | 0.13 [-0.02017, 0.211] | 0.07585 | 0.2783 | 0.4571 | 0 |

## excess_target_share_post

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | -0.2229 | 20 |
| hybrid-linucb-d-g098 | -0.2285 | 20 |
| hybrid-linucb-d-g098-l0 | -0.1759 | 20 |
| hybrid-linucb-d-g098-l025 | -0.2121 | 20 |
| hybrid-linucb-d-g098-l100 | -0.2216 | 20 |
| hybrid-linucb-l0 | -0.2091 | 20 |
| hybrid-linucb-l025 | -0.2178 | 20 |
| hybrid-linucb-l100 | -0.243 | 20 |
| leastloaded | -0.2764 | 20 |
| sed | -0.2497 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.03737 [0.01461, 0.0646] | 0.0003223 | 0.002352 | 0.8476 | 0 |
| hybrid-linucb | sed | 20 | 0.02113 [0.008051, 0.03442] | 0.003153 | 0.01261 | 0.7238 | 0 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.04526 [0.02794, 0.07106] | 6.294e-05 | 0.0007553 | 0.9143 | 0 |
| hybrid-linucb-d-g098 | sed | 20 | 0.02077 [0.0117, 0.04737] | 0.0007076 | 0.004246 | 0.8095 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.09761 [0.07252, 0.1199] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 0.08122 [0.06514, 0.104] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 0.05687 [0.04879, 0.07941] | 2.67e-05 | 0.0003471 | 0.9429 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 0.02781 [0.02146, 0.05092] | 1.907e-06 | 3.052e-05 | 1 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.05311 [0.02555, 0.07139] | 0.0002613 | 0.002352 | 0.8571 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.02616 [0.00678, 0.03477] | 0.007296 | 0.02189 | 0.6667 | 0 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.09252 [0.0354, 0.1087] | 0.0001049 | 0.001049 | 0.8952 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.05247 [0.02621, 0.07675] | 0.0002613 | 0.002352 | 0.8571 | 0 |
| hybrid-linucb-l025 | leastloaded | 20 | 0.04689 [0.02615, 0.0695] | 8.202e-05 | 0.0009022 | 0.9048 | 0 |
| hybrid-linucb-l025 | sed | 20 | 0.01788 [-0.002191, 0.06755] | 0.03623 | 0.07247 | 0.5333 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.03203 [0.005311, 0.04441] | 0.002325 | 0.01163 | 0.7429 | 0 |
| hybrid-linucb-l100 | sed | 20 | -9.669e-05 [-0.01469, 0.03291] | 0.5958 | 0.5958 | 0.1429 | 0 |

## time_to_adapt_rank

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 52.5 | 20 |
| hybrid-linucb-d-g098 | 47 | 20 |
| hybrid-linucb-d-g098-l0 | 42.5 | 20 |
| hybrid-linucb-d-g098-l025 | 25 | 20 |
| hybrid-linucb-d-g098-l100 | 28.5 | 20 |
| hybrid-linucb-l0 | 56.5 | 20 |
| hybrid-linucb-l025 | 26 | 20 |
| hybrid-linucb-l100 | 40 | 20 |
| leastloaded | 23.5 | 20 |
| sed | 22 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 17 [-7.5, 45] | 0.04994 | 0.799 | 0.5 | 0 |
| hybrid-linucb | sed | 20 | 27 [0, 55.5] | 0.0979 | 1 | 0.4444 | 2 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 4.5 [-8.5, 30.5] | 0.3047 | 1 | 0.2684 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | 2 [-8.5, 37] | 0.2958 | 1 | 0.2807 | 2 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 9.5 [0, 30.5] | 0.2758 | 1 | 0.3007 | 3 |
| hybrid-linucb-d-g098-l0 | sed | 20 | 6 [-13.5, 21.5] | 0.4079 | 1 | 0.2222 | 2 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | 1 [-3.5, 28.5] | 0.3654 | 1 | 0.2574 | 4 |
| hybrid-linucb-d-g098-l025 | sed | 20 | 2 [-7, 36] | 0.4588 | 1 | 0.1988 | 2 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0 [-11, 17] | 0.7111 | 1 | 0.09942 | 2 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0 [-15, 12] | 0.7959 | 1 | -0.07353 | 4 |
| hybrid-linucb-l0 | leastloaded | 20 | 13 [0.5, 46] | 0.07327 | 1 | 0.4684 | 1 |
| hybrid-linucb-l0 | sed | 20 | 4.5 [0, 24.5] | 0.1329 | 1 | 0.4035 | 2 |
| hybrid-linucb-l025 | leastloaded | 20 | 0 [-22.5, 7.5] | 0.8276 | 1 | -0.05848 | 2 |
| hybrid-linucb-l025 | sed | 20 | 0.5 [-11.5, 5] | 0.8361 | 1 | 0.05882 | 4 |
| hybrid-linucb-l100 | leastloaded | 20 | 6 [-14.5, 16.5] | 0.5327 | 1 | 0.1632 | 1 |
| hybrid-linucb-l100 | sed | 20 | 1.5 [-9.5, 29] | 0.4586 | 1 | 0.2157 | 3 |

## target_after_recovery_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## target_after_recovery_tasks

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_on_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_off_phases_target_n

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | NA | 0 |
| hybrid-linucb-d-g098 | NA | 0 |
| hybrid-linucb-d-g098-l0 | NA | 0 |
| hybrid-linucb-d-g098-l025 | NA | 0 |
| hybrid-linucb-d-g098-l100 | NA | 0 |
| hybrid-linucb-l0 | NA | 0 |
| hybrid-linucb-l025 | NA | 0 |
| hybrid-linucb-l100 | NA | 0 |
| leastloaded | NA | 0 |
| sed | NA | 0 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|

## drift_phase_1_on_share

| Arm | Median | n |
|---|---:|---:|
| hybrid-linucb | 0.2148 | 20 |
| hybrid-linucb-d-g098 | 0.1924 | 20 |
| hybrid-linucb-d-g098-l0 | 0.1835 | 20 |
| hybrid-linucb-d-g098-l025 | 0.1855 | 20 |
| hybrid-linucb-d-g098-l100 | 0.2078 | 20 |
| hybrid-linucb-l0 | 0.1885 | 20 |
| hybrid-linucb-l025 | 0.205 | 20 |
| hybrid-linucb-l100 | 0.1911 | 20 |
| leastloaded | 0.1866 | 20 |
| sed | 0.2017 | 20 |

| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |
|---|---|---:|---:|---:|---:|---:|---:|
| hybrid-linucb | leastloaded | 20 | 0.0154 [-0.01469, 0.04036] | 0.07283 | 1 | 0.4737 | 1 |
| hybrid-linucb | sed | 20 | 0.01166 [0.001818, 0.02118] | 0.1119 | 1 | 0.4158 | 1 |
| hybrid-linucb-d-g098 | leastloaded | 20 | 0.0006588 [-0.01642, 0.01302] | 0.8596 | 1 | 0.05263 | 1 |
| hybrid-linucb-d-g098 | sed | 20 | -0.005418 [-0.01957, 0.02039] | 0.6676 | 1 | -0.1095 | 0 |
| hybrid-linucb-d-g098-l0 | leastloaded | 20 | 0.009594 [-0.02212, 0.02001] | 0.8124 | 1 | 0.06667 | 0 |
| hybrid-linucb-d-g098-l0 | sed | 20 | -0.006906 [-0.01816, 0.001993] | 0.2873 | 1 | -0.2714 | 0 |
| hybrid-linucb-d-g098-l025 | leastloaded | 20 | -0.01281 [-0.02083, 0.01516] | 0.5706 | 1 | -0.1524 | 0 |
| hybrid-linucb-d-g098-l025 | sed | 20 | -0.006684 [-0.02084, 0.01145] | 0.2162 | 1 | -0.3238 | 0 |
| hybrid-linucb-d-g098-l100 | leastloaded | 20 | 0.01061 [-0.007359, 0.03559] | 0.2943 | 1 | 0.2762 | 0 |
| hybrid-linucb-d-g098-l100 | sed | 20 | 0.005233 [-0.01302, 0.02979] | 0.5509 | 1 | 0.1696 | 2 |
| hybrid-linucb-l0 | leastloaded | 20 | 0.01207 [-0.01848, 0.02712] | 0.6742 | 1 | 0.1143 | 0 |
| hybrid-linucb-l0 | sed | 20 | 0.001965 [-0.01883, 0.0232] | 0.8906 | 1 | -0.04211 | 1 |
| hybrid-linucb-l025 | leastloaded | 20 | -0.001944 [-0.01433, 0.03164] | 0.5459 | 1 | 0.1619 | 0 |
| hybrid-linucb-l025 | sed | 20 | -0.001316 [-0.02603, 0.02386] | 1 | 1 | 0 | 0 |
| hybrid-linucb-l100 | leastloaded | 20 | 0.00175 [-0.01136, 0.03452] | 0.6477 | 1 | 0.1238 | 0 |
| hybrid-linucb-l100 | sed | 20 | -0.006087 [-0.03427, 0.02671] | 0.7938 | 1 | -0.06667 | 0 |

## Last finisher

| Round | Arm | Worker address | CPU millicores |
|---:|---|---|---:|
| 1 | hybrid-linucb | worker-2:50051 | 600 |
| 1 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 1 | hybrid-linucb-d-g098-l0 | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 1 | hybrid-linucb-d-g098-l100 | worker-4:50051 | 1000 |
| 1 | hybrid-linucb-l0 | worker-2:50051 | 600 |
| 1 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 1 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 1 | leastloaded | worker-1:50051 | 500 |
| 1 | sed | worker-4:50051 | 1000 |
| 2 | hybrid-linucb | worker-2:50051 | 600 |
| 2 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 2 | hybrid-linucb-d-g098-l0 | worker-3:50051 | 800 |
| 2 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 2 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 2 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 2 | hybrid-linucb-l025 | worker-2:50051 | 600 |
| 2 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 2 | leastloaded | worker-3:50051 | 800 |
| 2 | sed | worker-2:50051 | 600 |
| 3 | hybrid-linucb | worker-4:50051 | 1000 |
| 3 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 3 | hybrid-linucb-d-g098-l0 | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g098-l025 | worker-2:50051 | 600 |
| 3 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 3 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 3 | hybrid-linucb-l025 | worker-5:50051 | 1100 |
| 3 | hybrid-linucb-l100 | worker-1:50051 | 500 |
| 3 | leastloaded | worker-1:50051 | 500 |
| 3 | sed | worker-4:50051 | 1000 |
| 4 | hybrid-linucb | worker-5:50051 | 1100 |
| 4 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 4 | hybrid-linucb-d-g098-l0 | worker-3:50051 | 800 |
| 4 | hybrid-linucb-d-g098-l025 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 4 | hybrid-linucb-l0 | worker-2:50051 | 600 |
| 4 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 4 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 4 | leastloaded | worker-3:50051 | 800 |
| 4 | sed | worker-4:50051 | 1000 |
| 5 | hybrid-linucb | worker-4:50051 | 1000 |
| 5 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-d-g098-l0 | worker-3:50051 | 800 |
| 5 | hybrid-linucb-d-g098-l025 | worker-2:50051 | 600 |
| 5 | hybrid-linucb-d-g098-l100 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-l0 | worker-3:50051 | 800 |
| 5 | hybrid-linucb-l025 | worker-5:50051 | 1100 |
| 5 | hybrid-linucb-l100 | worker-5:50051 | 1100 |
| 5 | leastloaded | worker-4:50051 | 1000 |
| 5 | sed | worker-4:50051 | 1000 |
| 6 | hybrid-linucb | worker-3:50051 | 800 |
| 6 | hybrid-linucb-d-g098 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g098-l0 | worker-1:50051 | 500 |
| 6 | hybrid-linucb-d-g098-l025 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 6 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 6 | hybrid-linucb-l025 | worker-3:50051 | 800 |
| 6 | hybrid-linucb-l100 | worker-5:50051 | 1100 |
| 6 | leastloaded | worker-1:50051 | 500 |
| 6 | sed | worker-3:50051 | 800 |
| 7 | hybrid-linucb | worker-5:50051 | 1100 |
| 7 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g098-l0 | worker-1:50051 | 500 |
| 7 | hybrid-linucb-d-g098-l025 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-l0 | worker-3:50051 | 800 |
| 7 | hybrid-linucb-l025 | worker-1:50051 | 500 |
| 7 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 7 | leastloaded | worker-1:50051 | 500 |
| 7 | sed | worker-5:50051 | 1100 |
| 8 | hybrid-linucb | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 8 | hybrid-linucb-d-g098-l0 | worker-1:50051 | 500 |
| 8 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 8 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 8 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 8 | hybrid-linucb-l025 | worker-3:50051 | 800 |
| 8 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 8 | leastloaded | worker-1:50051 | 500 |
| 8 | sed | worker-3:50051 | 800 |
| 9 | hybrid-linucb | worker-2:50051 | 600 |
| 9 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098-l0 | worker-3:50051 | 800 |
| 9 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 9 | hybrid-linucb-d-g098-l100 | worker-1:50051 | 500 |
| 9 | hybrid-linucb-l0 | worker-5:50051 | 1100 |
| 9 | hybrid-linucb-l025 | worker-3:50051 | 800 |
| 9 | hybrid-linucb-l100 | worker-5:50051 | 1100 |
| 9 | leastloaded | worker-1:50051 | 500 |
| 9 | sed | worker-5:50051 | 1100 |
| 10 | hybrid-linucb | worker-4:50051 | 1000 |
| 10 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098-l025 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-d-g098-l100 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-l0 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-l025 | worker-5:50051 | 1100 |
| 10 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 10 | leastloaded | worker-2:50051 | 600 |
| 10 | sed | worker-2:50051 | 600 |
| 11 | hybrid-linucb | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 11 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 11 | hybrid-linucb-d-g098-l025 | worker-3:50051 | 800 |
| 11 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 11 | hybrid-linucb-l0 | worker-3:50051 | 800 |
| 11 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 11 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 11 | leastloaded | worker-2:50051 | 600 |
| 11 | sed | worker-5:50051 | 1100 |
| 12 | hybrid-linucb | worker-3:50051 | 800 |
| 12 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 12 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 12 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 12 | hybrid-linucb-l0 | worker-2:50051 | 600 |
| 12 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 12 | hybrid-linucb-l100 | worker-5:50051 | 1100 |
| 12 | leastloaded | worker-2:50051 | 600 |
| 12 | sed | worker-3:50051 | 800 |
| 13 | hybrid-linucb | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-g098-l0 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-d-g098-l025 | worker-3:50051 | 800 |
| 13 | hybrid-linucb-d-g098-l100 | worker-1:50051 | 500 |
| 13 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 13 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 13 | leastloaded | worker-3:50051 | 800 |
| 13 | sed | worker-3:50051 | 800 |
| 14 | hybrid-linucb | worker-2:50051 | 600 |
| 14 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 14 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 14 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 14 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 14 | hybrid-linucb-l0 | worker-3:50051 | 800 |
| 14 | hybrid-linucb-l025 | worker-1:50051 | 500 |
| 14 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 14 | leastloaded | worker-1:50051 | 500 |
| 14 | sed | worker-4:50051 | 1000 |
| 15 | hybrid-linucb | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-d-g098 | worker-3:50051 | 800 |
| 15 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 15 | hybrid-linucb-d-g098-l025 | worker-2:50051 | 600 |
| 15 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 15 | hybrid-linucb-l0 | worker-5:50051 | 1100 |
| 15 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 15 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 15 | leastloaded | worker-1:50051 | 500 |
| 15 | sed | worker-5:50051 | 1100 |
| 16 | hybrid-linucb | worker-3:50051 | 800 |
| 16 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 16 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-d-g098-l025 | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 16 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 16 | hybrid-linucb-l025 | worker-5:50051 | 1100 |
| 16 | hybrid-linucb-l100 | worker-5:50051 | 1100 |
| 16 | leastloaded | worker-1:50051 | 500 |
| 16 | sed | worker-3:50051 | 800 |
| 17 | hybrid-linucb | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g098-l0 | worker-5:50051 | 1100 |
| 17 | hybrid-linucb-d-g098-l025 | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 17 | hybrid-linucb-l0 | worker-4:50051 | 1000 |
| 17 | hybrid-linucb-l025 | worker-3:50051 | 800 |
| 17 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 17 | leastloaded | worker-5:50051 | 1100 |
| 17 | sed | worker-3:50051 | 800 |
| 18 | hybrid-linucb | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g098 | worker-1:50051 | 500 |
| 18 | hybrid-linucb-d-g098-l0 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g098-l025 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-d-g098-l100 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-l0 | worker-2:50051 | 600 |
| 18 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 18 | hybrid-linucb-l100 | worker-2:50051 | 600 |
| 18 | leastloaded | worker-1:50051 | 500 |
| 18 | sed | worker-3:50051 | 800 |
| 19 | hybrid-linucb | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-d-g098 | worker-5:50051 | 1100 |
| 19 | hybrid-linucb-d-g098-l0 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-d-g098-l025 | worker-1:50051 | 500 |
| 19 | hybrid-linucb-d-g098-l100 | worker-3:50051 | 800 |
| 19 | hybrid-linucb-l0 | worker-1:50051 | 500 |
| 19 | hybrid-linucb-l025 | worker-4:50051 | 1000 |
| 19 | hybrid-linucb-l100 | worker-4:50051 | 1000 |
| 19 | leastloaded | worker-4:50051 | 1000 |
| 19 | sed | worker-2:50051 | 600 |
| 20 | hybrid-linucb | worker-3:50051 | 800 |
| 20 | hybrid-linucb-d-g098 | worker-2:50051 | 600 |
| 20 | hybrid-linucb-d-g098-l0 | worker-2:50051 | 600 |
| 20 | hybrid-linucb-d-g098-l025 | worker-4:50051 | 1000 |
| 20 | hybrid-linucb-d-g098-l100 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-l0 | worker-5:50051 | 1100 |
| 20 | hybrid-linucb-l025 | worker-3:50051 | 800 |
| 20 | hybrid-linucb-l100 | worker-3:50051 | 800 |
| 20 | leastloaded | worker-2:50051 | 600 |
| 20 | sed | worker-4:50051 | 1000 |
