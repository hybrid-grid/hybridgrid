# Theory Notes — Bandit Scheduling for Distributed Compilation

> **Compiled by:** verified subagent research pass (2026-04-29).
> **Verification policy:** every claim below is anchored to a peer-reviewed paper or canonical textbook, with section/equation references and DOI/arXiv ID. Items marked *Needs further verification* are stated as such; do **not** cite them as established fact in the thesis without independent confirmation.

---

## 1. LinUCB Algorithm Canonical Form

**Source.** Li, L., Chu, W., Langford, J., Schapire, R.E. (2010). "A Contextual-Bandit Approach to Personalized News Article Recommendation." *WWW '10*. **DOI: 10.1145/1772690.1772758**. arXiv: 1003.0146 (v2, 2012-03-01).

### 1.1 Linear payoff assumption (§3.1, Eq. 2)

For each arm $a$ with context $x_{t,a} \in \mathbb{R}^d$ at round $t$:

$$
\mathbb{E}[r_{t,a} \mid x_{t,a}] = x_{t,a}^\top \theta_a^*
$$

The model is *disjoint* — one parameter vector per arm.

### 1.2 Algorithm 1 (LinUCB with disjoint linear models) — verbatim

```
Input: α ∈ ℝ⁺
for t = 1, 2, …, T:
  observe x_{t,a} ∈ ℝ^d for all a ∈ A_t
  for a ∈ A_t:
    if a is new:
      A_a ← I_d            (line 5)
      b_a ← 0_{d×1}        (line 6)
    θ̂_a ← A_a^{-1} b_a    (line 8)
    p_{t,a} ← θ̂_a^⊤ x_{t,a} + α √(x_{t,a}^⊤ A_a^{-1} x_{t,a})  (line 9)
  a_t ← argmax_a p_{t,a}  (line 11)
  observe r_t
  A_{a_t} ← A_{a_t} + x_{t,a_t} x_{t,a_t}^⊤   (line 12, rank-1 update)
  b_{a_t} ← b_{a_t} + r_t x_{t,a_t}            (line 13)
```

### 1.3 Exploration parameter $\alpha$ — Eq. (4)

The paper states (verbatim):

> "for any $\delta > 0$ and $x_{t,a} \in \mathbb{R}^d$, where $\alpha = 1 + \sqrt{\ln(2/\delta)/2}$ is a constant."

Followed by the practical caveat:

> "the value of $\alpha$ given in Eq. (4) may be conservatively large in some applications, and so optimizing this parameter may result in higher total payoffs in practice."

**Empirical convention.** Practitioners use $\alpha \in [0.1, 2.0]$. This range is *not* prescribed by Li 2010 or Chu 2011; it is an engineering heuristic. The thesis must report the value chosen and note that it was tuned empirically.

### 1.4 Initialisation and complexity

$A_a$ is initialised to the **unit identity** $I_d$ (Algorithm 1, line 5), implicitly performing $\lambda = 1$ ridge regularisation. The paper states the per-step matrix update cost is $\mathcal{O}(d^2)$ (using the Sherman–Morrison formula, §3 below).

### 1.5 Regret of plain LinUCB

Li 2010 reports an empirical regret of $\tilde{\mathcal{O}}(\sqrt{KdT})$. A formal regret proof for plain LinUCB is *not* given in the original paper; the proof exists for the modified SupLinUCB algorithm (§3 below).

---

## 2. Sherman–Morrison Formula (Incremental Inverse)

**Sources.** Sherman, J. & Morrison, W.J. (1950). *Annals of Mathematical Statistics* 21(1):124–127. Golub, G.H. & Van Loan, C.F. (2013). *Matrix Computations*, 4th ed. (Johns Hopkins University Press), §2.1.4. **ISBN 978-1-4214-0794-4**.

### 2.1 General formula

For an invertible $A \in \mathbb{R}^{n \times n}$ and column vectors $u, v \in \mathbb{R}^n$ such that $1 + v^\top A^{-1} u \neq 0$:

$$
(A + u v^\top)^{-1} = A^{-1} - \frac{A^{-1} u v^\top A^{-1}}{1 + v^\top A^{-1} u}
$$

### 2.2 Specialisation to LinUCB

LinUCB's update is $A_a \leftarrow A_a + x_{t,a} x_{t,a}^\top$ — the rank-1 case where $u = v = x_{t,a}$:

$$
A_{\text{new}}^{-1} = A_{\text{old}}^{-1} - \frac{A_{\text{old}}^{-1} x x^\top A_{\text{old}}^{-1}}{1 + x^\top A_{\text{old}}^{-1} x}
$$

### 2.3 Complexity claim

For $n \times n$ matrices the operation count is $3n^2$ scalar multiplications (Hager 1989; reproduced in Wikipedia). This matches the Li 2010 paper's claim of $\mathcal{O}(d^2)$ per update.

---

## 3. Regret Bound for Linear Contextual Bandits

**Source.** Chu, W., Li, L., Reyzin, L., Schapire, R.E. (2011). "Contextual Bandits with Linear Payoff Functions." *AISTATS 2011*, JMLR W&CP 15:208–214. <https://proceedings.mlr.press/v15/chu11a.html>.

### 3.1 Setting (§3)

> "We operate under the linear realisability assumption; that is, there exists an unknown weight vector $\theta^* \in \mathbb{R}^d$ with $\lVert\theta^*\rVert \le 1$ so that $\mathbb{E}[r_{t,a} \mid x_{t,a}] = x_{t,a}^\top \theta^*$ for all $t$ and $a$."

Noise: $r_{t,a}$ are independent random variables with that conditional mean. Contexts: $\lVert x_{t,a}\rVert \le 1$, "chosen arbitrarily by an oblivious adversary".

### 3.2 Theorem 1 — upper bound for SupLinUCB

> "If SupLinUCB is run with $\alpha = \sqrt{\tfrac{1}{2}\ln(2TK/\delta)}$, then with probability at least $1-\delta$, the regret of the algorithm is $\mathcal{O}\!\left(\sqrt{T d \ln^3(KT \ln T / \delta)}\right)$."

### 3.3 Theorem 2 — matching lower bound

For any algorithm:
$$
\mathbb{E}\!\left[\sum_t \max_a x_{t,a}^\top \theta^* - \sum_t r_{t,a_t}\right] \ge \gamma \sqrt{Td}
$$

so regret is $\Omega(\sqrt{Td})$. The two bounds match up to logarithmic factors.

### 3.4 Caveats relevant to this thesis

1. **The bound is for SupLinUCB**, not for plain LinUCB Algorithm 1. The paper explicitly states: "While experiments show LinUCB is probably sufficient in practice (Li et al., 2010), there is technical difficulty in analyzing it." The thesis must distinguish between *implementing* LinUCB (which we do) and *citing* the regret bound (which technically applies to SupLinUCB).
2. **Linear realisability is required.** When the true expected reward is non-linear in the features (very plausible for compile time), the bound does not apply; misspecification of magnitude $\varepsilon$ can inflate regret by $\mathcal{O}(\varepsilon \sqrt{T})$ (Lattimore & Szepesvári 2020, Ch. 24.4).
3. **No drift guarantee.** If $\theta^*$ changes over time (worker thermal throttling, GC pauses), the bound fails.

---

## 4. Reward Design

### 4.1 Sutton & Barto on reward design

**Source.** Sutton, R.S. & Barto, A.G. (2018). *Reinforcement Learning: An Introduction*, 2nd ed., MIT Press. <https://mitpress.mit.edu/9780262039246/>.

The relevant text is **§3.2 "Goals and Rewards"** (p. 57). It states the reward hypothesis:

> "all of what we mean by goals and purposes can be well thought of as the maximisation of the expected value of the cumulative sum of a received scalar signal (called reward)."

Sutton & Barto **does not give a formal theory of reward scaling or transformation**. Reward design is treated as a domain engineering choice.

### 4.2 Decima — actual reward function

**Source.** Mao, H. et al. (2019). "Learning Scheduling Algorithms for Data Processing Clusters." *SIGCOMM '19*. **DOI: 10.1145/3341302.3342080**. arXiv: 1810.01963.

Decima's reward, defined in **§5.2**, is verbatim:

> "if the objective is to minimise the average JCT, Decima penalises the agent $r_k = -(t_k - t_{k-1}) J_k$ after the $k$-th action, where $J_k$ is the number of jobs in the system during the interval $[t_{k-1}, t_k)$."

The justification is **Little's Law** ($L = \lambda W$): minimising the time-integrated job count minimises the average job-completion time.

### 4.3 Status of $r = -\log(1 + \text{latency})$

**No peer-reviewed source in our research uses this specific transform for scheduling rewards.** Earlier plan files (`linucb-scheduler-m2.md`, the M2 paper-skeleton draft) attribute it to "Decima §4.2 precedent". That attribution is **incorrect** — Decima's reward is the time-integrated job count, not a logarithm of latency.

Action items:
- The $-\log(1+\cdot)$ form must be presented in the thesis as an *engineering choice* motivated by heavy-tail compression, **not** cited to Decima or any other source.
- Consider implementing the Decima-style integrand $-(t_k - t_{k-1}) J_k$ as an ablation; this *is* citable.
- An ablation comparing raw $-\text{latency}$, $-\log(1+\text{latency})$, and Decima-style time-integrated penalty would strengthen the empirical justification.

---

## 5. Online Learning Under Non-Stationarity

**Source.** Lattimore, T. & Szepesvári, Cs. (2020). *Bandit Algorithms*, Cambridge University Press. Free PDF: <https://tor-lattimore.com/downloads/book/book.pdf>.

### 5.1 Relevant chapters

- **Ch. 19 "Stochastic Linear Bandits"** (p. 238) — confidence bounds for LinUCB / OFUL.
- **Ch. 20 "Confidence Bounds for Least Squares Estimators"** (p. 254) — self-normalised inequality.
- **Ch. 31 "Non-stationary Bandits"** (p. 379) — drift, change-points, tracking regret.

### 5.2 Drift assumption

Ch. 31 opens with:

> "With no assumptions, there is not much to be done. Because of this, it is usual to assume the distributions change infrequently or drift slowly."

The chapter develops *tracking regret* against the best policy in hindsight that switches at most $m$ times. Most techniques are stated for finite-armed bandits; the linear case is research-level.

### 5.3 LinUCB has no published drift guarantee

Standard LinUCB / OFUL analysis (Abbasi-Yadkori et al. 2011, NeurIPS) assumes a *fixed* unknown $\theta^*$. When $\theta^*$ drifts, the analysis breaks down.

Practical mitigations:

| Strategy | Guarantee | Reference |
|---|---|---|
| Sliding window | None tight in linear case | folklore |
| Exponential discounting | None tight in linear case | folklore |
| Change-point + restart | Bound exists for finite-armed, *not* linear | Lattimore & Szepesvári Ch. 31 |

For the thesis: thermal throttling and worker drift are **known limitations**, to be acknowledged in the limitations section.

### 5.4 The deadly triad

**Source.** Sutton & Barto 2018, **§11.3 "The Deadly Triad"**.

> "The instability and risk of divergence arise when we combine three factors: function approximation, bootstrapping, and off-policy training."

Verified by van Hasselt et al. (2018) "Deep Reinforcement Learning and the Deadly Triad", arXiv:1812.02648, which cites Sutton & Barto 2018 §11.3 directly.

LinUCB **does not invoke the triad** because it is a bandit (no value function, no bootstrapping). If the project later moves to full Q-learning with function approximation over multi-step state, the triad becomes a critical concern. State this as a scope boundary.

---

## 6. Heterogeneous Machine Scheduling

### 6.1 HEFT — Topcuoglu, Hariri, Wu 2002

**Source.** Topcuoglu, H., Hariri, S., Wu, M.-Y. (2002). "Performance-effective and low-complexity task scheduling for heterogeneous computing." *IEEE Transactions on Parallel and Distributed Systems* 13(3):260–274. **DOI: 10.1109/71.993206**.

**Phase 1 — upward rank** (computed from exit task to entry):

$$
\text{rank}_u(n_i) = \overline{w}(n_i) + \max_{n_j \in \text{succ}(n_i)} \left( \overline{c}(e_{i,j}) + \text{rank}_u(n_j) \right)
$$

with $\overline{w}(n_i) = (1/m)\sum_{j=1}^{m} w_{i,j}$ the average computation cost across $m$ processors, and $\overline{c}(e_{i,j})$ the average inter-task communication cost. For the exit task, $\text{rank}_u(n_{\text{exit}}) = \overline{w}(n_{\text{exit}})$.

**Phase 2 — earliest finish time on processor $p_j$:**

$$
\text{EST}(v_i, p_j) = \max\!\left(\text{avail}[j], \max_{v_t \in \text{pred}(v_i)} \text{AFT}(v_t) + c_{t,i}\right)
$$
$$
\text{EFT}(v_i, p_j) = w_{ij} + \text{EST}(v_i, p_j)
$$

The selected processor is $\arg\min_{p_j}\text{EFT}(v_i, p_j)$, with **insertion-based scheduling** (a task may fill an idle gap between already-scheduled tasks).

Complexity is $\mathcal{O}(eq)$, where $e$ = edge count and $q$ = processor count.

### 6.2 HEFT for streaming arrivals — *needs further verification*

HEFT as published is **offline**: it requires the full DAG upfront to compute upward ranks. No paper found in this research pass establishes a canonical "online HEFT". The 2024 survey arXiv:2408.02938 acknowledges dynamic / RL-based extensions exist; a thesis adaptation must clearly document the deviation from Topcuoglu 2002.

For independent compilation tasks (no DAG, $\text{succ}(n_i) = \emptyset$), HEFT degenerates to **Longest Processing Time (LPT)** assignment: rank is just $\overline{w}(n_i)$, and we pick the processor minimising completion time. This *can* run online and is what we will implement as the HEFT baseline.

### 6.3 Lenstra–Shmoys–Tardos R||C_max

**Source.** Lenstra, J.K., Shmoys, D.B., Tardos, É. (1990). *Mathematical Programming* 46(1–3):259–271. **DOI: 10.1007/BF01585745**.

Confirmed:
- 2-approximation (LP-relaxation + rounding)
- $\frac{3}{2}$ lower bound (no polynomial-time algorithm achieves a worst-case ratio strictly less than $\tfrac{3}{2}$ unless P = NP)
- The $[3/2, 2]$ gap remains open after 35 years.

---

## 7. Verification Status (Summary)

| Claim | Status | Source |
|---|---|---|
| LinUCB Algorithm 1 pseudocode | **Verified** | Li 2010 §3.1 |
| $\alpha = 1 + \sqrt{\ln(2/\delta)/2}$ | **Verified** | Li 2010 Eq. (4) |
| $A_a \leftarrow I_d$ initialisation | **Verified** | Li 2010 line 5 |
| Sherman–Morrison formula | **Verified** | Wikipedia + Golub–Van Loan §2.1.4 |
| $\mathcal{O}(d^2)$ update cost | **Verified** | Li 2010 + Hager 1989 |
| SupLinUCB regret $\mathcal{O}(\sqrt{Td}\log^{3/2})$ | **Verified** | Chu 2011 Thm. 1 |
| $\Omega(\sqrt{Td})$ lower bound | **Verified** | Chu 2011 Thm. 2 |
| Decima reward $-(t_k - t_{k-1})J_k$ | **Verified** | Mao 2019 §5.2 |
| $r = -\log(1+\text{latency})$ | **Empirical only** — *no peer-reviewed support* | — |
| Sutton & Barto §11.3 deadly triad | **Verified** | S&B 2018 §11.3 + van Hasselt 2018 |
| HEFT rank function | **Verified** | Topcuoglu 2002 |
| Online HEFT canonical reference | **Needs further verification** | survey arXiv:2408.02938 |
| LinUCB drift guarantee | **Verified absent** | Lattimore & Szepesvári Ch. 31 |
| Lenstra–Shmoys–Tardos 2-approx + 3/2 lb | **Verified** | LST 1990 |

---

## 8. Action Items for Subsequent Work

1. **Code comments** in `internal/coordinator/scheduler/linucb.go` must cite Li 2010 §3.1 lines 5–13 and the Eq. (4) form of $\alpha$.
2. **Reward function code path** must label $-\log(1+\cdot)$ as an empirical heuristic, not a Decima derivation. Consider implementing the Little's-Law-based reward as an alternative.
3. **Paper limitations section** must call out: linear realisability assumption, no drift guarantee, plain-LinUCB regret proof gap relative to SupLinUCB.
4. **HEFT baseline** in `internal/coordinator/scheduler/heft.go` must document the online adaptation explicitly: independent tasks, no DAG, rank = $\overline{w}(n_i)$, EFT-greedy assignment.
5. **Pending verifications** (do not cite without re-checking): online HEFT prior art (single 2024 survey reference is thin); the practical $\alpha$ tuning range $[0.1, 2.0]$ should be cited to specific empirical papers if used.
