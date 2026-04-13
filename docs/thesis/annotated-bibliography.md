# Annotated Bibliography — Scheduling in Distributed Build Systems

## Tier 1 — Foundation Papers

### 1. The Power of Two Choices in Randomized Load Balancing

**Mitzenmacher, M. (2001).** IEEE Transactions on Parallel and Distributed Systems, 12(10), 1094-1104. DOI: `10.1109/71.963420`

**Contribution:** Chứng minh chiến lược "chọn 2 server ngẫu nhiên, gán cho server ít tải hơn" giảm max load từ O(log n / log log n) xuống O(log log n) — cải thiện theo hàm mũ kép so với random placement, trong khi chi phí quyết định gần như không đổi (O(2) thay vì O(1)).

**Key results:**
- Single random choice: max load = Theta(log n / log log n) w.h.p.
- Two choices (d=2): max load = ln ln n / ln 2 + O(1) w.h.p.
- d choices: max load = ln ln n / ln d + O(1) — diminishing returns sau d=2
- Kết quả áp dụng cho cả mô hình tĩnh (one-shot) và động (continuous arrivals/departures)
- Phân tích dùng density-dependent jump Markov chains và hệ phương trình vi phân

**Limitations:**
- Giả định server đồng nhất (homogeneous) — không xét khác biệt năng lực giữa các node
- Giả định thông tin tải chính xác và cập nhật tức thời — không xét stale information
- Mô hình balls-into-bins không phản ánh task có thời gian xử lý khác nhau

**Relevance cho đề tài:** P2C là nền tảng lý thuyết của scheduler hiện tại trong Hybrid-Grid. Điểm mở rộng: thay tiêu chí "queue ngắn nhất" bằng "estimated completion time thấp nhất" từ learned cost model, biến thuật toán từ homogeneous-aware thành heterogeneous-aware mà giữ nguyên tính chất low-overhead của P2C.

---

### 2. Sparrow: Distributed, Low Latency Scheduling

**Ousterhout, K., Wendell, P., Zaharia, M., Stoica, I. (2013).** Proceedings of the 24th ACM SOSP, 69-84. DOI: `10.1145/2517349.2522716`

**Contribution:** Scheduler phân tán đạt scheduling latency dưới millisecond cho workload analytics (Spark) bằng hai kỹ thuật: batch sampling (probe d*m workers cho job m tasks) và late binding (chỉ gán task khi worker callback rảnh, giải quyết stale information).

**Key results:**
- 110 machines (440 slots) trên EC2, workload TPC-H trên Spark
- Median response time chỉ chậm 12% so với centralized omniscient scheduler
- Late binding giảm median wait time 55% so với early binding
- Per-job sampling độc lập giữa các scheduler, scale linearly

**Limitations:**
- Worker homogeneous — không có cơ chế ước lượng worker capacity
- Không hỗ trợ task priority, fairness policy phức tạp
- Không xét task có thời gian biến thiên lớn theo đặc tính worker

**Relevance:** Late binding đặc biệt hữu ích cho compilation: thời gian build một file biến thiên mạnh (10ms - 30s), gán tại thời điểm worker rảnh tốt hơn gán trước. Kế thừa cấu trúc probe + late binding nhưng thay tiêu chí chọn từ "queue length" thành "estimated build time" từ cost model.

---

### 3. Approximation Algorithms for Scheduling Unrelated Parallel Machines

**Lenstra, J.K., Shmoys, D.B., Tardos, E. (1990).** Mathematical Programming, 46(1-3), 259-271. DOI: `10.1007/BF01585745`

**Contribution:** Đưa ra thuật toán 2-approximation đầu tiên cho bài toán minimize makespan trên unrelated machines (R||C_max), đồng thời chứng minh lower bound 3/2 (NP-hard). Sau 35 nam gap [3/2, 2] vẫn chưa được thu hẹp đáng kể.

**Key results:**
- Mô hình R||C_max: m machines, n jobs, processing time p_ij (khác tùy cặp job-machine)
- Lower bound: không tồn tại polynomial-time algorithm với ratio < 3/2 trừ khi P = NP
- 2-approximation: binary search + LP relaxation + rounding dựa trên basic feasible solution
- Kỹ thuật rounding: mỗi machine nhận thêm tối đa 1 job so với LP solution

**Limitations:**
- LP-based, complexity cao cho online setting
- Offline: giả định biết trước tất cả jobs và processing times p_ij
- Gap [3/2, 2] chưa closed

**Relevance:** Cung cấp nền tảng lý thuyết: mô hình R||C_max trực tiếp mô tả distributed build (mỗi compile unit có thời gian khác nhau trên mỗi worker). Kết quả 3/2-hardness cho thấy không thể có thuật toán tối ưu, justify việc dùng heuristic. Thay p_ij chính xác bằng estimated p_ij từ learned cost model, đánh giá competitive ratio thực nghiệm.

---

## Tier 2 — Extended Reading

### 4. Dominant Resource Fairness (DRF)

**Ghodsi, A. et al. (2011).** "Dominant Resource Fairness: Fair Allocation of Multiple Resource Types." NSDI '11.

Mở rộng max-min fairness cho multi-resource (CPU, memory, disk). Workers trong build cluster có resource profiles khác nhau; DRF cung cấp framework công bằng khi chia sẻ cluster giữa nhiều build jobs đồng thời.

### 5. Tetris: Multi-Resource Packing

**Grandl, R. et al. (2014).** "Multi-Resource Packing for Cluster Schedulers." SIGCOMM '14.

Multi-dimensional bin packing heuristic kết hợp scoring function match task resource requirements với available resources. Có thể adapt scoring function để tích hợp learned cost model.

### 6. Quasar: Resource-Efficient Cluster Management

**Delimitrou, C., Kozyrakis, C. (2014).** ASPLOS '14.

Dùng collaborative filtering (tương tự recommendation systems) và classification để dự đoán performance trên các machine types khác nhau. Không dùng RL — dùng matrix factorization và nearest-neighbor regression. Rất gần mục tiêu đề tài: học cost model từ historical data cho placement trên heterogeneous cluster.

### 7. Resource Central (Microsoft)

**Cortez, E. et al. (2017).** "Resource Central: Understanding and Predicting Workloads for Improved Resource Management in Large Cloud Platforms." SOSP '17.

Dùng random forests và gradient boosted trees để dự đoán VM lifetime, resource usage. ML đơn giản (không RL), trained trên historical traces. Phương pháp luận có thể áp dụng: EMA, linear regression, hoặc gradient boosted trees cho predict build time.

### 8. Mantri: Straggler Mitigation

**Ananthanarayanan, G. et al. (2010).** "Reining in the Outliers in MapReduce Clusters using Mantri." OSDI '10.

Cause-aware restart: phân loại nguyên nhân straggler và chọn hành động phù hợp. Nếu build task chạy chậm hơn predicted time quá ngưỡng, có thể dùng Mantri-style speculative re-execution.

### 9. The Tail at Scale

**Dean, J., Barroso, L.A. (2013).** Communications of the ACM, 56(2), 74-80.

Survey từ Google về tail latency: hedged requests, tied requests, micro-partitioning. Trong distributed build, 1 file compile chậm quyết định tổng makespan. Hedged requests có thể áp dụng cho critical-path compilation units.

---

## Mapping vào đề tài

| Khía cạnh | Paper nền tảng | Ứng dụng trong Hybrid-Grid |
|---|---|---|
| Scheduling baseline | Mitzenmacher (2001) | P2C làm baseline, mở rộng bằng cost-weighted choice |
| Distributed architecture | Ousterhout (2013) — Sparrow | Probe + late binding cho low-latency decisions |
| Theoretical hardness | Lenstra et al. (1990) | R\|\|C_max model, 2-approx bound, justify heuristic |
| Heterogeneous matching | Grandl (2014), Ghodsi (2011) | Multi-resource scoring function |
| Learned cost model | Delimitrou (2014), Cortez (2017) | Collaborative filtering, regression, EMA |
| Straggler handling | Ananthanarayanan (2010), Dean (2013) | Progress monitoring, speculative execution |

---

> **Luu y:** Khi sử dụng trong luận văn, cần truy cập bản gốc qua IEEE Xplore, ACM DL, SpringerLink để xác minh trích dẫn và số liệu cụ thể.
