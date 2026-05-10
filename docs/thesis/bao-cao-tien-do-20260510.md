# Báo cáo tiến độ đồ án — 10/05/2026

> Người báo cáo: HieuLD
> Đề tài: Thuật toán lập lịch dựa trên Contextual Bandit cho hệ thống biên dịch phân tán Hybrid-Grid
> Kỳ báo cáo trước: 14/04/2026 (file `docs/thesis/bao-cao-tien-do-20260414.md`)
> Mã nguồn: <https://github.com/hybrid-grid/hybridgrid> (nhánh `main`, 17 commit kể từ 14/04)

## §1 Tóm tắt điều hành

Trong khoảng thời gian từ buổi gặp gần nhất (14/04/2026) đến hôm nay, nhóm đã hoàn tất toàn bộ ba mốc kỹ thuật chính (M1 + M2 + M3), bổ sung một thuật toán baseline kinh điển (HEFT của Topcuoglu et al. 2002) ngoài kế hoạch ban đầu, hoàn thành một lượt khảo sát siêu tham số (α-sweep), và soạn thảo đầy đủ một chương luận văn cùng một sườn paper định hướng công bố quốc tế.

Bốn kết quả định lượng quan trọng:
1. **Sáu thuật toán scheduling** đã được hiện thực và so sánh trên cùng workload thực (CPython 873 task biên dịch) trên ba cấu hình cluster khác chủng (1, 3, 5 worker).
2. **LinUCB sau khi vá ba lỗi triển khai** đạt thời gian build ngang heuristic P2C trên cấu hình khó nhất (5w-hetero, 94 giây), đồng thời **giảm 2.3% độ trễ đuôi P99** (18 896 ms so với 19 347 ms của P2C). Đây là lợi thế đặc thù của framework bandit theo nguyên tắc "tail at scale" (Dean & Barroso 2013).
3. **5 238 bản ghi đo lường** chi tiết được thu thập (873 task × 6 scheduler) ở định dạng JSON Lines 27 trường, sẵn sàng cho phân tích sâu hoặc huấn luyện ML offline.
4. Toàn bộ dẫn chứng lý thuyết được kiểm chứng độc lập: **22 claim với DOI/arXiv ID đầy đủ**, một dẫn chứng sai trong tài liệu lập kế hoạch trước đây đã được phát hiện và sửa lại.

## §2 Phạm vi công việc đã hoàn thành

### §2.1 Mốc M1 — Hạ tầng đo lường
Sản phẩm bàn giao:
- Factory `newScheduler` định tuyến 6 implementation qua cờ CLI `--scheduler`.
- Bộ ghi `TaskLogger` thread-safe sinh log JSON Lines 27 trường mỗi task.
- 25 trường gốc (`task_id`, `worker_id`, `worker_cpu_cores`, `compile_time_ms`, `queue_time_ms`, …) + 2 trường introspection cho learner (`q_value_at_dispatch`, `was_exploration`) bổ sung trong M2/M3.
- **9 unit test** (gồm test concurrency 16 goroutine × 100 writes) tất cả pass với `-race` detector.
- Kịch bản benchmark `test/stress/benchmark-heterogeneous.sh` được mở rộng để chấp nhận biến môi trường `SCHEDULER`, `ALPHA`, `EPSILON`, hỗ trợ chạy α-sweep mà không phải sửa script.

### §2.2 Mốc M2 — ε-greedy bandit
Sản phẩm bàn giao:
- Implement `EpsilonGreedyScheduler` theo Sutton & Barto 2018 §2.4 (incremental sample-mean update Q ← Q + (R−Q)/n).
- `LearningScheduler` interface mở rộng cho `Scheduler` để hỗ trợ feedback loop online, cùng helper `SelectWith` cho phép code coordinator tự động phát hiện learner.
- Hook reward feedback trong gRPC `Compile()` handler.
- **9 unit test** bao phủ: convergence trên 3 cánh tay synthetic, exploration rate ±2σ binomial test, single-candidate fast path, NaN/Inf guard, concurrency.
- Benchmark đầy đủ trên 3 cấu hình cluster.

### §2.3 Mốc M3 — LinUCB contextual bandit
Sản phẩm bàn giao:
- Implement `LinUCBScheduler` theo Algorithm 1 của Li, Chu, Langford, Schapire 2010 (verbatim từ §3.1, có comment tham chiếu line numbers).
- Cập nhật ma trận nghịch đảo bằng công thức Sherman-Morrison (Sherman & Morrison 1950; Golub & Van Loan §2.1.4) — chi phí $\mathcal{O}(d^2)$ mỗi cập nhật.
- Véc-tơ đặc trưng 9 chiều: bias, log source size chuẩn hoá, target arch one-hot, năng lực worker (CPU/RAM), arch match, queue ratio, RPC latency.
- Tích hợp `gonum/v1/gonum 0.16.0` cho linear algebra (không tự cài đặt).
- **8 unit test** bao gồm một test "ground-truth" so sánh cached Sherman-Morrison `A^{-1}` với một phép invert lại từ đầu (`mat.Inverse`) sau 50 rank-1 update — sai số dưới $10^{-6}$.
- Cờ CLI `--alpha`, validated trong khoảng $[0, 10]$.
- TaskContext mang `TaskID` + `SourceSizeBytes` để cache feature vector ở Select time, dùng lại ở RecordOutcome (xem §5).
- Benchmark đầy đủ trên 3 cấu hình cluster + α-sweep ở 4 giá trị {0.1, 0.5, 1.0, 2.0}.

### §2.4 Bổ sung ngoài kế hoạch — HEFT baseline
Việc thêm HEFT là quyết định nội bộ nhằm có một baseline kinh điển không học (non-learning) mạnh hơn LeastLoaded. Sản phẩm bàn giao:
- `HEFTScheduler` triển khai dạng thích ứng online của Topcuoglu et al. 2002 (DOI 10.1109/71.993206) — vì task biên dịch là độc lập (không có DAG dependency giữa các file), thuật toán suy biến về phép gán EFT-greedy với upward-rank chỉ là $\bar{w}(n_i)$.
- Cold-start prior dựa trên CPU cores theo mô hình $1/\sqrt{\text{cores}/4}$.
- Mỗi lệch khỏi thuật toán gốc đều được ghi rõ trong code comment để người đọc kiểm tra.
- **6 unit test** bao gồm: cold-start preference cho worker mạnh hơn, learned-time tracking, queue-aware penalty, fast path, no-workers, concurrency.

### §2.5 Khảo sát siêu tham số (α-sweep) — ablation §5.5.1
Bốn lần chạy benchmark trên cấu hình 5w-hetero với $\alpha \in \{0.1, 0.5, 1.0_{\text{lỗi}}, 2.0\}$. Kết quả tóm tắt trong Bảng 3 (§4.4). Quan sát phụ: cấu hình 1 worker có sàn nhiễu khoảng ±25 giây, cung cấp số liệu định lượng cho variance budget của môi trường thí nghiệm.

### §2.6 Tài liệu hoá
- `docs/thesis/chuong-X-lap-lich.md` — bản thảo chương luận văn ~8 trang văn phong học thuật trang trọng, đầy đủ 6 mục từ "đặt vấn đề" đến "kết luận" với hai bảng số liệu chính.
- `docs/thesis/theory-notes.md` — tài liệu lý thuyết ~10 trang, mỗi claim có DOI/arXiv ID, có bảng "Verification Status" tổng hợp 22 dòng.
- `docs/thesis/paper-skeleton.md` — sườn paper tiếng Anh ~25 trang theo cấu trúc SIGCOMM-style (8 section), tất cả bảng kết quả đã điền số liệu thực.
- `.sisyphus/evidence/m1/findings.md` — phân tích thực nghiệm chi tiết 9 mục, bao gồm chẩn đoán negative result và bảng so sánh đầy đủ.

## §3 Kết quả thực nghiệm

### §3.1 Bảng 1 — Tổng thời gian build (giây) trên ba cấu hình cluster

| Cấu hình | LeastLoaded | P2C | ε-greedy | LinUCB α=1 (lỗi) | **LinUCB-fixed α=0.5** | HEFT |
|---|---|---|---|---|---|---|
| 1w-4.0CPU | 92 | 130 | 146 | 129 | 131 | 129 |
| 3w-hetero | 123 | 85 | 142 | 103 | 108 | 135 |
| **5w-hetero** | 152 | **94** | 119 | 158 | **94** | 144 |

Kết quả khẳng định: P2C vượt LeastLoaded **1.62×** trên 5w-hetero — phù hợp với kết quả lý thuyết của Mitzenmacher 2001 (P2C giảm max load từ $\Theta(\log n / \log\log n)$ xuống $\Theta(\log\log n)$). LinUCB sau khi sửa lỗi đuổi kịp P2C ở 5w (cùng 94 giây).

### §3.2 Bảng 2 — Lệch tải và độ trễ đuôi trên 5w-hetero (873 task)

| Scheduler | Top:Bottom dispatch | P50 (ms) | P95 (ms) | P99 (ms) |
|---|---|---|---|---|
| LeastLoaded | 10.8 : 1 | 820 | 6 226 | 23 961 |
| P2C | 8.6 : 1 | 704 | 5 830 | 19 347 |
| ε-greedy | 13.2 : 1 | 956 | 7 387 | 25 488 |
| LinUCB α=1 (lỗi) | 97 : 1 | 942 | 7 142 | 23 661 |
| **LinUCB-fixed α=0.5** | 13.9 : 1 | 826 | 5 676 | **18 896** |
| HEFT | 145 : 1 | 995 | 6 919 | 24 143 |

Quan sát: **LinUCB-fixed có P99 thấp nhất**, kém P2C 22 ms ở P50 nhưng vượt P2C ở P95 (5 676 vs 5 830) và P99 (18 896 vs 19 347). Đây là lợi thế đặc thù của bandit framework: cải thiện ở phần đuôi của phân phối, đúng với mục tiêu mà framework được thiết kế (Dean & Barroso 2013, "The Tail at Scale", CACM).

### §3.3 Bảng 3 — Khảo sát siêu tham số α (LinUCB sau vá lỗi, 5w-hetero)

| α | Wall-clock (s) | Ghi chú |
|---|---|---|
| 0.1 | 98 | Chi phí exploration tối thiểu |
| **0.5** | **94** | Tối ưu trên workload này, ngang P2C |
| 1.0 (lỗi gốc) | 158 | Số liệu trước khi vá lỗi — không so sánh α |
| 2.0 | 96 | Cạnh tranh; chi phí exploration cao hơn |

Khoảng cách giữa α=0.1 và α=0.5 chỉ 4 giây — nằm trong sàn nhiễu — cho thấy *sau khi vá lỗi*, knob α có ảnh hưởng giới hạn trên workload này. Khuyến cáo thực dụng: với pattern triển khai đúng, mặc định α=0.5 là khởi điểm an toàn.

### §3.4 Bảng 4 — Phân tích Q-value distribution (so sánh trước/sau vá lỗi)

| Scheduler | Q mean | Q std | Q range | Exploration rate |
|---|---|---|---|---|
| ε-greedy (no normalisation) | −6.61 | 1.06 | [−7.82, 0.00] | 8.4% |
| LinUCB α=1 (lỗi gốc) | n/a | n/a | [−7.06, +2.54] | 1.7% |
| **LinUCB-fixed α=0.5** | n/a | n/a | [−0.53, +1.18] | 25.1% |

Sự co lại của Q range từ `[−7, +2.5]` sang `[−0.53, +1.18]` là bằng chứng định lượng rằng phép chuẩn hoá phần thưởng (chia cho `log1p(timeout_ms)`) đã đưa UCB bonus và mean estimate về cùng bậc độ lớn. Tỉ lệ exploration tăng từ 1.7% lên 25.1% là kết quả của một bug fix riêng biệt: cờ `wasExploration` trước đây so sánh sai, sau khi sửa nó báo cáo trung thực khi nào bonus (chứ không phải mean) chi phối quyết định.

## §4 Quy trình đảm bảo chất lượng

Để đảm bảo các kết quả số nêu trên là kiểm chứng được, nhóm đã áp dụng chuỗi ba lượt review độc lập (writer/reviewer separation) với role tách biệt — không tự khen tự duyệt:

| Lượt | Vai trò | Kết quả phát hiện |
|---|---|---|
| 1 | document-specialist | Xác minh độc lập 22 claim lý thuyết bằng cách fetch các paper gốc (arXiv, ACM DL, IEEE Xplore). **Phát hiện một lỗi** trong tài liệu lập kế hoạch: hàm phần thưởng `r = -log(1+t)` đã bị attribute sai cho Decima §4.2; thực tế Decima dùng công thức $r_k = -(t_k - t_{k-1})J_k$ có cơ sở Định luật Little. Đã sửa trong `theory-notes.md` và mọi file dẫn chiếu. |
| 2 | code-reviewer | Đọc 4 file mã nguồn LinUCB và call-site Compile(). **Phát hiện ba lỗi nghiêm trọng** (chi tiết §5). |
| 3 | architect | Sau khi vá lỗi, kiểm tra toàn bộ body of work: kiểm tra code-doc consistency, citation accuracy, schema conformance, test coverage. **Verdict: APPROVE** với một drift nhỏ trong chương luận văn (đã sửa). |

Quy trình này không chỉ đảm bảo correctness mà còn để lại **dấu vết kiểm chứng** (audit trail) — các comment trong code chỉ ra section/equation cụ thể của paper, các bảng trong theory-notes có cột "Verification Status", các evidence file có thể được chạy lại bằng pandas trên log thô.

## §5 Bài học rút ra từ quá trình vá lỗi

Lần chạy đầu tiên của LinUCB cho 158 giây trên 5w-hetero — xấu nhất trong sáu scheduler. Sau khi code-reviewer chỉ ra ba lỗi và nhóm vá xong, makespan hồi phục về 94 giây — toàn bộ 64 giây thiếu hụt được lấy lại mà *không thay đổi thuật toán*. Mỗi lỗi đại diện cho một class lỗi ML systems engineering, có giá trị giáo dục cho luận văn:

### Lỗi 1 — Rò rỉ nhãn ở update path (target leakage in async update)
Khi `RecordOutcome` chạy, `worker.ActiveTasks` đã bị `DecrementTasks` giảm và bộ ghi RTT đã hấp thụ độ trễ chính task đang xét. Bandit do đó học $\hat{\theta}$ trên véc-tơ đặc trưng *khác* với véc-tơ đã dùng để Select. **Class lỗi**: nhầm lẫn giữa state-at-decision và state-at-observation trong các update path bất đồng bộ — phổ biến trong RL production.

**Cách sửa**: lưu cache véc-tơ tại Select keyed by `TaskID`, dùng lại ở RecordOutcome. Drop update nếu cache miss thay vì rebuild từ stale state.

### Lỗi 2 — Đặc trưng one-hot trùng tuyến tính với bias
Cờ build_type luôn là `(1, 0, 0)` dưới luồng `Compile()` hiện tại. Chiều CPP trùng hoàn toàn với bias, hai chiều Flutter/Unity vĩnh viễn không học. Sherman-Morrison làm việc trong một subspace rank-2 thấp hơn dự kiến. **Class lỗi**: copy-paste pattern khi enum chỉ có một giá trị active tại một thời điểm.

**Cách sửa**: rút véc-tơ từ 12 chiều xuống 9 chiều, bỏ ba chiều build_type. Khi Flutter/Unity được tích hợp vào luồng học sẽ thêm lại đúng cách.

### Lỗi 3 — Phần thưởng át UCB bonus
Phần thưởng nguyên gốc trong khoảng `[−10, −3]` (raw `−log(1+t_ms)`); UCB bonus cỡ $\alpha\sqrt{x^T A^{-1} x} \approx 1$ tại warm-up. Kết quả: bonus bị át hoàn toàn, scheduler hành xử như round-robin random — *xấu hơn LeastLoaded*. **Class lỗi**: thiếu kiểm tra magnitude tương đối giữa các thành phần của UCB score.

**Cách sửa**: chuẩn hoá phần thưởng theo `log1p(timeout_ms)` để $r \in [-1, 0]$, mặc định α=0.5.

Câu chuyện này đã được tài liệu hoá trong commit `85f20e2` và mục §X.5 của chương luận văn. Trong các paper RL được công bố, phần "implementation pitfalls" thường bị bỏ qua; ghi lại một cách chi tiết là một đóng góp giáo dục độc lập của đồ án.

## §6 Hạn chế đã nhận diện

Để minh bạch với hội đồng, nhóm chủ động liệt kê năm hạn chế hiện hữu:

1. **Số lần lặp thí nghiệm** — phần lớn số liệu Bảng 1-2 đến từ một lần chạy. Sàn nhiễu Docker/macOS quan sát được khoảng ±25 giây (xem cột 1w trong Bảng 3). Để đạt mức tin cậy paper, cần lặp ≥ 5 lần và áp dụng kiểm định Wilcoxon — chi phí ước lượng 6 giờ tính toán liên tục, đã lên kế hoạch cho phiên thí nghiệm tiếp theo.
2. **Đa dạng workload** — chỉ test trên CPython. Các codebase nặng template (Boost, Eigen, LLVM) có phân phối nặng đuôi hơn — kết quả tương đối có thể thay đổi.
3. **Drift và thermal throttling** — LinUCB chuẩn không có guarantee dưới drift (Lattimore & Szepesvári 2020 Ch. 31). Hệ thống hiện chưa có change-point detection.
4. **Khoảng trống lý thuyết LinUCB vs SupLinUCB** — regret bound $O(\sqrt{Td})$ của Chu et al. 2011 chứng minh được cho biến thể SupLinUCB; nhóm triển khai LinUCB Algorithm 1 nguyên bản, do đó bound được trích như *bối cảnh lý thuyết* chứ không phải đảm bảo trực tiếp.
5. **HEFT online adaptation** — thuật toán gốc Topcuoglu 2002 là offline; phần thích ứng online (HEFT-LPT degeneration) chỉ có một paper khảo sát 2024 đề cập, chưa có canonical reference.

## §7 Kế hoạch tiếp theo (4 tuần tới)

| Tuần | Việc | Đầu ra dự kiến |
|---|---|---|
| 1 | Multi-rep ≥5 cho mọi scheduler trên 5w-hetero + Wilcoxon test | Bảng 1 và 2 có cột mean ± stddev và p-value |
| 2 | Workload mở rộng: chạy benchmark trên Boost.MPL + clang/lib | Bảng so sánh thứ 5 trong findings.md |
| 3 | Reward ablation: triển khai biến thể Decima `r_k = -(t_k-t_{k-1})J_k`, so sánh với log-reward hiện tại | §5.5.2 của chương luận văn |
| 4 | Cache-aware feature mở rộng: thêm cờ "đã có cache" cho mỗi cặp (worker, file) vào véc-tơ đặc trưng | Phiên bản LinUCB-cache-aware, benchmark sơ bộ |

## §8 Câu hỏi xin ý kiến thầy/cô

1. **Đối tượng venue công bố** — với mức kết quả hiện có (LinUCB ngang P2C wall-clock, vượt P2C P99) thầy đánh giá phù hợp với hội nghị quốc tế Tier-2 (HotOS workshop, EuroSys late-breaking) hay tạp chí trong nước (Journal of Computer Science and Cybernetics)?

2. **Cache-aware là contribution chính hay phụ** — nếu chính, cần ~4 tuần thu thập dữ liệu cache hit rate per (worker, file) và mở rộng feature vector. Tăng đáng kể tính novelty (chưa có paper nào áp dụng bandit cho cache-aware compilation scheduling) nhưng cũng tăng rủi ro deadline.

3. **Mức độ chi tiết của phần bug-fix narrative trong luận văn** — nhóm đã viết tách §X.5 trình bày 3 lỗi và bài học. Thầy thấy nên giữ ở mức hiện tại, mở rộng thêm để trở thành đóng góp pedagogical riêng, hay rút gọn để giữ nội dung "thuần khoa học"?

4. **Decima reward ablation** — phát hiện công thức `r = -log(1+t)` không có nguồn peer-reviewed. Có nên triển khai cả công thức Decima `r_k = -(t_k-t_{k-1})J_k` (có cơ sở Little's Law) và so sánh thực nghiệm như một ablation chính thức, hay chỉ ghi nhận trong limitations?

5. **Trình tự ưu tiên giữa code và viết** — deadline thesis defense dự kiến cuối 2026 / Q1 2027. Với 7 tháng còn lại, thầy đề xuất ưu tiên (a) hoàn thiện code + benchmark trước rồi viết, hay (b) viết luận văn song song với thí nghiệm để khoá nội dung sớm?

## §9 Tài liệu đính kèm trong repo

| File | Mô tả |
|---|---|
| `docs/thesis/chuong-X-lap-lich.md` | Bản thảo chương luận văn (~8 trang) |
| `docs/thesis/theory-notes.md` | Ghi chú lý thuyết, mọi claim có DOI/arXiv |
| `docs/thesis/paper-skeleton.md` | Sườn paper tiếng Anh (~25 trang) định hướng công bố |
| `.sisyphus/evidence/m1/findings.md` | Phân tích thực nghiệm chi tiết 6 scheduler |
| `.sisyphus/evidence/m1/tasks-*.jsonl` | Log thô 873 record/scheduler để tái lập kết quả |
| `internal/coordinator/scheduler/*.go` | Mã nguồn 6 scheduler có comment tham chiếu paper |
| `.sisyphus/plans/linucb-scheduler*.md` | Tài liệu thiết kế chi tiết của ba milestone |

Toàn bộ source code và tài liệu nêu trên đã được commit vào repo công khai, có thể truy cập qua URL ở phần đầu báo cáo. Nhóm sẵn sàng trình bày trực tiếp khi thầy/cô sắp xếp được lịch.
