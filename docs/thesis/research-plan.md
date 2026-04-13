# Research Plan — RL-based Scheduling for Distributed Build Systems

> Mục tiêu: Xây dựng nền tảng kiến thức đủ sâu để đưa ra quyết định thiết kế có cơ sở khoa học.
> Nguyên tắc: **Hiểu trước, quyết sau.** Không commit phương pháp cụ thể cho đến khi hoàn thành Phase 1-3.
> Thời lượng dự kiến: 6-8 tuần research, sau đó mới chuyển sang implementation plan.

---

## Phase 1: Nền tảng lý thuyết (Tuần 1-3)

Mục tiêu: Hiểu sâu 3 trụ cột lý thuyết — scheduling theory, RL foundations, distributed build systems.

### 1.1 Scheduling Theory

**Đọc:**
- Pinedo, M. — *Scheduling: Theory, Algorithms, and Systems* (textbook) — Chapter 1-3 (models, complexity), Chapter 5 (parallel machines)
- Graham, R.L. (1969) — "Bounds on multiprocessing timing anomalies" — classical list scheduling bound
- Lenstra, Shmoys, Tardos (1990) — "Approximation algorithms for scheduling unrelated parallel machines" — 2-approx, 3/2 lower bound

**Cần nắm được:**
- Ký hiệu chuẩn: `P||Cmax`, `Q||Cmax`, `R||Cmax` — phân biệt identical / uniform / unrelated machines
- Bài toán của ta thuộc mô hình nào? (R||Cmax — unrelated, vì compile time khác nhau tùy cặp task-worker)
- Tại sao NP-hard? Reduction từ bài toán nào?
- Graham's bound `(2 - 1/m)` áp dụng ra sao cho online list scheduling?
- HEFT algorithm (Heterogeneous Earliest Finish Time) — baseline chuẩn cho heterogeneous scheduling

**Checkpoint — bạn hiểu đủ khi trả lời được:**
- [ ] Bài toán scheduling trong Hybrid-Grid thuộc lớp nào trong phân loại Graham?
- [ ] Tại sao không thể có thuật toán optimal polynomial-time? (proof sketch)
- [ ] HEFT hoạt động ra sao? Ưu nhược điểm?
- [ ] Online vs offline scheduling khác gì? Competitive ratio nghĩa là gì?

---

### 1.2 Reinforcement Learning Foundations

**Đọc (theo thứ tự, không skip):**
- Sutton & Barto — *Reinforcement Learning: An Introduction* (2nd ed., free online)
  - Chapter 1-3: Multi-armed bandits, MDP, returns (3-4 giờ)
  - Chapter 4-5: Dynamic programming, Monte Carlo (3-4 giờ)
  - Chapter 6: **TD Learning** — quan trọng nhất (3 giờ, đọc kỹ)
  - Chapter 9: **Function approximation** — linear, neural (4 giờ, đọc kỹ)
  - Chapter 10: On-policy approximation (3 giờ)

**Thực hành bắt buộc (không chỉ đọc):**
- Implement tabular Q-learning cho GridWorld (Python, ~50 LOC) — thấy convergence bằng mắt
- Implement Q-learning với linear function approximation cho CartPole hoặc MountainCar
- Thử thay đổi: learning rate, discount factor, exploration rate — quan sát ảnh hưởng
- Note: mục đích là **hiểu cơ chế**, không phải dùng framework (viết from scratch, không dùng stable-baselines)

**Cần nắm được:**
- Bellman equation: `Q*(s,a) = E[r + γ max_a' Q*(s',a')]`
- TD update: `Q(s,a) ← Q(s,a) + α[r + γ max_a' Q(s',a') - Q(s,a)]`
- Tại sao function approximation cần? (state space lớn → tabular không scale)
- Linear function approximation: `Q(s,a) = θᵀφ(s,a)` — gradient update
- Convergence guarantees: khi nào TD với linear FA converge? Khi nào không?
- Deadly triad: function approximation + bootstrapping + off-policy → có thể diverge
- Exploration-exploitation tradeoff: ε-greedy, UCB, Boltzmann — trade-offs

**Checkpoint:**
- [ ] Viết được Bellman equation cho bài toán scheduling (custom, không copy)
- [ ] Giải thích bằng lời: TD error nghĩa là gì về mặt trực giác?
- [ ] Implement Q-learning cho toy problem, vẽ được learning curve
- [ ] Giải thích: tại sao deep RL khó hơn linear RL? Khi nào dùng cái nào?
- [ ] Nêu được 3 rủi ro khi dùng RL cho real system

---

### 1.3 Multi-Armed Bandits & Contextual Bandits

**Tại sao cần hiểu riêng**: Bài toán scheduling có thể là bandit (horizon=1) hoặc full RL (horizon>1). Cần hiểu cả hai để quyết định đúng.

**Đọc:**
- Slivkins, A. (2019) — *Introduction to Multi-Armed Bandits* — Chapter 1-4 (free online)
- Li, L. et al. (2010) — "A Contextual-Bandit Approach to Personalized News Article Recommendation" (LinUCB paper)
- Lattimore & Szepesvári — *Bandit Algorithms* (2020) — Chapter 1-7 (optional, nếu muốn sâu)

**Cần nắm được:**
- Regret definition: `R(T) = Σ [r*(t) - r(t)]` — tổng loss so với optimal
- UCB (Upper Confidence Bound): `a* = argmax [Q(a) + c√(ln t / N(a))]`
- LinUCB: contextual bandit với linear payoff model
- Khi nào bandit đủ tốt vs khi nào cần full RL?
- Regret bounds: sublinear regret nghĩa là gì? O(√T) vs O(T)

**Checkpoint:**
- [ ] Phân biệt: multi-armed bandit vs contextual bandit vs full MDP
- [ ] Bài toán scheduling là bandit hay MDP? Lập luận cả 2 phía
- [ ] LinUCB hoạt động ra sao? Matrix inverse để làm gì?
- [ ] Khi nào regret O(√T) đạt được? Khi nào không?

---

## Phase 2: Khảo sát SOTA (Tuần 3-5)

Mục tiêu: Đọc và phân tích các paper RL scheduling quan trọng nhất. Xây dựng bảng so sánh.

### 2.1 Core Papers (đọc kỹ, ghi annotated notes)

| # | Paper | Venue | Tại sao đọc |
|---|---|---|---|
| 1 | Mitzenmacher 2001 — Power of Two Choices | IEEE TPDS | Nền tảng P2C, scheduler hiện tại |
| 2 | Ousterhout 2013 — Sparrow | SOSP | P2C cho data center, late binding |
| 3 | Mao 2016 — DeepRM | HotNets | RL đầu tiên cho resource management |
| 4 | Mao 2019 — Decima | SIGCOMM | RL + GNN cho job scheduling, benchmark chính |
| 5 | Delimitrou 2014 — Paragon/Quasar | ASPLOS | ML scheduling heterogeneous, collaborative filtering |
| 6 | Grandl 2014 — Tetris | SIGCOMM | Multi-resource bin packing heuristic |
| 7 | Topcuoglu 2002 — HEFT | IEEE TPDS | Baseline chuẩn cho heterogeneous scheduling |

### 2.2 Extended Papers (đọc abstract + method + results)

| # | Paper | Venue | Tại sao đọc |
|---|---|---|---|
| 8 | Qiao 2021 — Pollux | OSDI | Adaptive scheduling, so sánh approach |
| 9 | Cortez 2017 — Resource Central | SOSP | ML predict workload, Microsoft production |
| 10 | Ananthanarayanan 2013 — Late/Straggler clones | NSDI | Straggler mitigation |
| 11 | Dean 2013 — Tail at Scale | CACM | Tail latency why it matters |
| 12 | Li 2010 — LinUCB | WWW | Contextual bandit method |
| 13 | Watkins 1989/1992 — Q-learning | PhD thesis / ML | Original Q-learning |
| 14 | Lenstra-Shmoys-Tardos 1990 | Math. Prog. | Scheduling theory foundation |

### 2.3 Cho mỗi paper, ghi lại:

```
Paper: [title]
Problem: [bài toán gì?]
Method: [dùng gì? RL type? features?]
Baselines: [so sánh với gì?]
Evaluation: [metrics, workloads, cluster size]
Key result: [số liệu chính]
Limitation: [tác giả nói gì? bạn thấy gì?]
Relevance: [liên quan gì đến Hybrid-Grid?]
Gap: [cái gì chưa giải quyết mà ta có thể?]
```

### 2.4 Xây dựng comparison table

Sau khi đọc xong, tạo bảng so sánh tổng hợp:

```
| Paper | RL Type | Online/Offline | Heterogeneous | Cache-aware | Simulator needed | Scale |
```

**Checkpoint:**
- [ ] Annotated notes cho ít nhất 10/14 papers
- [ ] Comparison table hoàn chỉnh
- [ ] Viết được 1 đoạn 500 từ: "Research gap mà Hybrid-Grid đang nhắm vào là gì và tại sao nó quan trọng"
- [ ] Trả lời được: Decima mạnh ở đâu và yếu ở đâu so với bài toán của ta?

---

## Phase 3: Thí nghiệm khám phá (Tuần 5-7)

Mục tiêu: Hands-on experiments để hiểu bài toán thực tế trước khi quyết định method.

### 3.1 Data Collection — Hiểu workload thực tế

Chạy build thật, thu thập dữ liệu:

```bash
# Build CPython, log chi tiết mỗi task
hgbuild -v make -j8 2>&1 | tee build_log.txt
```

**Thu thập cho mỗi compilation task:**
- Source file name, size (bytes)
- Preprocessed size (bytes)
- Compiler, flags
- Worker assigned
- Compilation time (ms)
- Cache hit/miss
- Worker CPU, RAM, arch

**Phân tích:**
- Distribution of compilation times — uniform? bimodal? long-tail?
- Correlation: source size vs compile time? Có phải linear?
- Top 10 slowest files — chúng có đặc điểm gì chung?
- Per-worker variance — cùng file trên 2 workers khác nhau bao nhiêu?

**Mục đích**: Hiểu data TRƯỚC khi chọn model. Nếu compile time gần như uniform → RL ít giúp. Nếu long-tail → RL có đất diễn.

### 3.2 Heuristic Baseline Profiling

Chạy 3 scheduler hiện có, thu thập chi tiết:

```bash
for sched in simple leastloaded p2c; do
    # modify grpc.go line 177 (hoặc sau khi implement --scheduler flag)
    # run benchmark, collect per-task data
done
```

**Phân tích:**
- Worker utilization heatmap theo thời gian
- File nào bị assign "sai" (file to trên worker yếu)?
- Makespan breakdown: scheduling time + queue time + compile time + network time
- Cache hit rate có khác nhau giữa schedulers không?

**Mục đích**: Biết chính xác heuristic fail ở đâu → biết RL cần cải thiện chỗ nào.

### 3.3 Feature Correlation Study

Trước khi build RL model, kiểm tra features có predict được compile time không:

```python
# Dùng Python vì tiện cho data analysis
import pandas as pd
from sklearn.linear_model import LinearRegression
from sklearn.ensemble import RandomForestRegressor

# Load collected data
df = pd.read_csv('build_log.csv')

# Feature correlation
print(df[['source_size', 'preprocessed_size', 'include_count',
           'worker_cpu', 'worker_ram', 'is_native_arch',
           'compile_time_ms']].corr())

# Can a simple model predict compile time?
X = df[features]
y = df['compile_time_ms']
model = LinearRegression().fit(X, y)
print(f"R² = {model.score(X, y)}")  # Nếu R² < 0.3 → features chưa đủ
```

**Mục đích**: Nếu features không predict được compile time → RL sẽ không work (garbage in, garbage out). Cần tìm features tốt hơn trước khi implement RL.

### 3.4 Toy RL Experiment

Implement RL scheduler đơn giản nhất có thể (prototype, không cần clean code):

```python
# Python prototype, KHÔNG phải Go production code
# Mục đích: kiểm tra RL CÓ THỂ học được gì từ data này không

class SimpleQLearning:
    def __init__(self, n_features, alpha=0.01, epsilon=0.1):
        self.theta = np.zeros(n_features)
        ...

    def select(self, task_features, worker_features_list):
        # Q(s,a) = theta @ phi(task, worker) cho mỗi worker
        ...

    def update(self, features, reward):
        # TD update
        ...

# Chạy trên collected data (offline replay)
for task in build_log:
    action = agent.select(task.features, workers)
    reward = -task.actual_compile_time
    agent.update(features, reward)

# Plot learning curve
```

**Mục đích**: Kiểm tra concept trước khi commit 6 tuần implement trong Go. Nếu prototype không learn → cần thay đổi approach, chứ không phải "code Go tốt hơn".

**Checkpoint Phase 3:**
- [ ] Có dataset build log ≥ 1000 tasks từ ≥ 2 projects
- [ ] Biết distribution of compile times (histogram)
- [ ] Biết R² của linear model predict compile time
- [ ] Biết features nào correlate mạnh nhất với compile time
- [ ] Prototype RL có show learning curve giảm hay không
- [ ] Report 1-2 trang: "Findings từ exploratory experiments"

---

## Phase 4: Tổng hợp & Quyết định (Tuần 7-8)

Mục tiêu: Từ tất cả kiến thức Phase 1-3, chốt hướng đi cụ thể.

### 4.1 Synthesis Document

Viết 1 tài liệu ~5 trang trả lời:

1. **Bài toán chính xác là gì?** (MDP formulation hoặc bandit, có lý do)
2. **Phương pháp nào phù hợp nhất?** (Q-learning? Policy gradient? Bandit? Hybrid? — dựa trên data thực nghiệm Phase 3)
3. **Features nào dùng?** (dựa trên correlation study)
4. **Baselines nào?** (ít nhất 4, bao gồm HEFT)
5. **Evaluation plan cụ thể?** (workloads, clusters, metrics, statistical tests)
6. **Research questions chính?** (3-5 RQs, mỗi cái trả lời được bằng thí nghiệm)
7. **Limitations dự kiến?** (biết trước thì defend tốt hơn)
8. **Timeline implementation + evaluation + writing?**

### 4.2 Decision Meeting

Họp nhóm (2 thành viên) + GVHD:
- Present synthesis document
- Xin GVHD approve hướng đi
- Chốt scope cuối cùng trước khi implement

### 4.3 Sau khi chốt → chuyển sang Implementation Plan

Lúc này mới viết implementation plan (files, phases, tests, commits) — vì bây giờ bạn biết chính xác mình đang implement CÁI GÌ và TẠI SAO.

**Checkpoint Phase 4:**
- [ ] Synthesis document hoàn chỉnh
- [ ] GVHD approved hướng đi
- [ ] Cả 2 thành viên đều hiểu và đồng thuận
- [ ] Implementation plan viết xong (lúc này mới viết)
- [ ] Bắt đầu code

---

## Timeline tổng hợp

```
Tuần 1-2:  Phase 1.1-1.2 — Scheduling theory + RL foundations (đọc + thực hành)
Tuần 3:    Phase 1.3 + 2.1 — Bandits + bắt đầu đọc core papers
Tuần 4-5:  Phase 2.2-2.4 — Đọc xong papers, comparison table
Tuần 5-6:  Phase 3.1-3.2 — Data collection, heuristic profiling
Tuần 6-7:  Phase 3.3-3.4 — Feature study, toy RL prototype
Tuần 7-8:  Phase 4 — Synthesis, quyết định, implementation plan

Tuần 9+:   Implementation (6 tuần)
Tuần 15+:  Evaluation (4 tuần)
Tuần 19+:  Thesis writing (6-8 tuần)
```

---

## Phân công 2 thành viên

### Song song trong Phase 1-2 (đọc)

| Tuần | Thành viên 1 (bạn) | Thành viên 2 (mới) |
|---|---|---|
| 1 | Sutton & Barto Ch.1-6 (RL) | Go Tour + project setup + codebase reading |
| 2 | Sutton & Barto Ch.9-10 (Function Approx) | Scheduling theory (Pinedo Ch.1-3, HEFT paper) |
| 3 | Papers 1-4 (Mitzenmacher, Sparrow, DeepRM, Decima) | Papers 5-7 (Paragon, Tetris, HEFT) + Slivkins bandits Ch.1-4 |
| 4 | Papers 8-10 + annotated notes | Papers 11-14 + annotated notes |
| 5 | Comparison table + gap analysis | Implement Q-learning toy (GridWorld, CartPole) |

### Cùng làm trong Phase 3 (experiments)

| Việc | Ai |
|---|---|
| Instrument hgbuild để log chi tiết (Go code) | Thành viên 1 |
| Chạy benchmark, collect data | Cả 2 |
| Feature correlation study (Python) | Thành viên 2 |
| Toy RL prototype (Python) | Thành viên 1 |
| Viết findings report | Cả 2 cùng viết |

### Phase 4 (quyết định)

Cả 2 cùng viết synthesis document, cùng present cho GVHD.

---

## Deliverables mỗi tuần

Mỗi cuối tuần, mỗi thành viên nộp:

1. **Reading notes** — bullet points, không cần dài. Ưu tiên: "tôi hiểu gì", "tôi không hiểu gì", "cái này liên quan đến Hybrid-Grid ra sao"
2. **Checkpoint answers** — trả lời các câu hỏi checkpoint tương ứng
3. **Questions log** — câu hỏi nảy sinh, để discuss nhóm hoặc hỏi GVHD

Lưu tất cả vào `docs/thesis/research-notes/week-N-[tên].md`

---

## Tài liệu tham khảo chính

### Sách (đọc selected chapters)
- Sutton & Barto — *Reinforcement Learning: An Introduction* (2nd ed.) — http://incompleteideas.net/book/the-book-2nd.html
- Slivkins — *Introduction to Multi-Armed Bandits* — https://arxiv.org/abs/1904.07272
- Pinedo — *Scheduling: Theory, Algorithms, and Systems* (library / institutional access)

### Papers (DOI / search trên Google Scholar)
- Mitzenmacher 2001, DOI: 10.1109/71.963420
- Ousterhout 2013 (Sparrow), DOI: 10.1145/2517349.2522716
- Mao 2016 (DeepRM), DOI: 10.1145/3005745.3005750
- Mao 2019 (Decima), DOI: 10.1145/3341302.3342080
- Delimitrou 2014 (Paragon), DOI: 10.1145/2541940.2541941
- Grandl 2014 (Tetris), DOI: 10.1145/2619239.2626334
- Topcuoglu 2002 (HEFT), DOI: 10.1109/71.993206
- Lenstra-Shmoys-Tardos 1990, DOI: 10.1007/BF01585745
- Li 2010 (LinUCB), DOI: 10.1145/1772690.1772758
- Dean 2013 (Tail at Scale), DOI: 10.1145/2408776.2408794

### Tools
- Python (data analysis): pandas, sklearn, matplotlib
- Go (implementation): standard library + gonum (nếu cần linear algebra)
- Jupyter Notebook: cho exploratory analysis
