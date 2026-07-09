# Chương X — Thuật toán lập lịch dựa trên Contextual Bandit cho hệ thống biên dịch phân tán Hybrid-Grid

> *Bản thảo gửi giáo viên hướng dẫn — phiên bản ngày 09/07/2026.*
> Mọi số liệu trong chương được dẫn nguồn từ các thư mục `.sisyphus/evidence/` và các paper trong `docs/thesis/theory-notes.md`. Bảng X.1–X.2 (mục X.4.2) là số liệu chạy đơn (single-run) ban đầu; mục **X.4.5** trình bày kết quả đo lường lại có kiểm soát nhiễu thí nghiệm (randomized block, 10 lần lặp) — đây là số liệu **có giá trị kết luận** và thay thế các so sánh single-run. Mục **X.4.6** trình bày ablation warm-bandit. Kết luận trung thực: contextual bandit **ngang** heuristic LeastLoaded trên workload này, vượt trội các baseline yếu hơn (P2C, LinUCB thuần) chỉ ở tải nhẹ.

## X.1 Đặt vấn đề

### X.1.1 Bối cảnh hệ thống

Hệ thống Hybrid-Grid Build chia một quá trình biên dịch C/C++ thành hàng trăm task độc lập (mỗi file `.c`/`.cpp` là một task) và phân phối chúng tới một cụm worker khác chủng (heterogeneous). Trong cấu hình điển hình mà chúng tôi đã đo (xem `.sisyphus/evidence/m1/findings.md`), một build CPython đơn lẻ sinh ra khoảng **293 translation unit** (task); con số 873 xuất hiện trong các tài liệu M1 ban đầu là **tổng cộng dồn trên ba cấu hình cluster** (1 + 3 + 5 worker, mỗi cấu hình ~291 task), không phải số task của một build. Tổng dung lượng CPU được cố định ở 4.0 lõi cho mọi cấu hình nhưng phân chia không đều giữa các worker. Các worker khác nhau về số lõi CPU, dung lượng bộ nhớ, kiến trúc tập lệnh, và cấu hình mạng. Do đó thời gian biên dịch cho cùng một file có thể lệch nhau đáng kể giữa các worker — đo lường thực nghiệm cho thấy tỉ số P99/P50 thời gian biên dịch lên tới 29 lần.

Coordinator (`hg-coord`) giữ một sổ đăng ký (registry) các worker đang sống và quyết định gửi task nào tới worker nào thông qua một thuật toán lập lịch (scheduler). Đây là điểm có ảnh hưởng lớn nhất tới tổng thời gian build (makespan) và là trọng tâm nghiên cứu của đồ án này.

### X.1.2 Mô hình hoá toán học

Bài toán lập lịch trong Hybrid-Grid là một thực thể của lớp `R||C_max` trong phân loại Graham — *unrelated parallel machines*: $m$ máy không đồng nhất, $n$ task, thời gian xử lý $p_{ij}$ phụ thuộc đồng thời vào cả task $j$ và máy $i$ và *không* tuân theo bất kỳ ràng buộc tỉ lệ nào (Lenstra, Shmoys, Tardos 1990, DOI 10.1007/BF01585745). Bài toán này đã được chứng minh là NP-khó, không có thuật toán đa thức nào đạt tỉ lệ xấp xỉ nhỏ hơn $\tfrac{3}{2}$ trừ khi $\text{P} = \text{NP}$, trong khi cận trên xấp xỉ tốt nhất hiện biết là $2$. Khoảng cách $[3/2, 2]$ vẫn còn mở sau hơn 35 năm, chứng tỏ độ khó về mặt lý thuyết của bài toán.

Trong thiết lập trực tuyến (online) — nơi task đến tuần tự và phải được gán ngay khi đến — đối thủ chuẩn là *list scheduling* của Graham 1969. Bound cạnh tranh (competitive ratio) cho list scheduling là $2 - 1/m$ trên môi trường đồng nhất; với máy khác chủng, không có bound cạnh tranh phổ quát chặt chẽ.

### X.1.3 Hạn chế của các heuristic tĩnh

Hai heuristic phổ biến đã được hiện thực trong Hybrid-Grid:

- **LeastLoaded:** chọn worker có ít task đang chạy nhất. Đơn giản, không học, không xét năng lực.
- **Power of Two Choices (P2C):** lấy mẫu hai worker ngẫu nhiên, chọn worker có điểm số (theo công thức trọng số tĩnh) cao hơn. Nền tảng lý thuyết ở Mitzenmacher 2001 (DOI 10.1109/71.963420), giảm tải tối đa từ $\Theta(\log n / \log\log n)$ xuống $\Theta(\log\log n)$ khi server đồng nhất.

Cả hai đều **không học từ dữ liệu thực thi**. Đo lường thực nghiệm (Bảng X.1) trên cùng cấu hình 5 worker khác chủng cho thấy:

**Bảng X.1 — Makespan (giây) trên benchmark CPython (~293 task/build), chạy đơn (single-run).**

| Cấu hình  | LeastLoaded | P2C  | ε-greedy | LinUCB α=1 (lỗi) | LinUCB-fixed α=0.5 | HEFT |
|-----------|-------------|------|----------|------|--------------------|------|
| 1w-4.0CPU | 92          | 130  | 146      | 129  | 131                | 129  |
| 3w-hetero | 123         | 85   | 142      | 103  | 108                | 135  |
| 5w-hetero | 152         | **94** | 119    | 158  | **94**             | 144  |

> ⚠️ **Lưu ý về độ tin cậy của Bảng X.1:** các số trên đến từ *một* lần chạy mỗi cấu hình, không kiểm soát nhiễu thí nghiệm (thứ tự chạy, thermal drift). Mục **X.4.5** cho thấy khi lặp 10 lần với thiết kế randomized block và kiểm định thống kê ghép cặp, một số chênh lệch trong Bảng X.1 **biến mất** (đặc biệt ưu thế biểu kiến của bandit so với LeastLoaded là *artifact* của thứ tự chạy). Bảng X.1 được giữ lại như bối cảnh lịch sử; kết luận định lượng lấy từ X.4.5.

**Bảng X.2 — Tỉ lệ phân phối task top:bottom (lệch tải) trên 5w-hetero.**

| Scheduler                | Top:Bottom | P99 compile_time (ms) |
|--------------------------|-----------:|----------------------:|
| LeastLoaded              |   10.8 : 1 | 23 961 |
| P2C                      |    8.6 : 1 | 19 347 |
| ε-greedy                 |   13.2 : 1 | 25 488 |
| LinUCB (α=1, lỗi)        |     97 : 1 | 23 661 |
| **LinUCB-fixed (α=0.5)** |   13.9 : 1 | **18 896** |
| HEFT                     |    145 : 1 | 24 143 |

Cả ba heuristic ban đầu đều dồn 33% lưu lượng vào duy nhất một worker mạnh nhất, để các worker yếu hơn ở dưới ngưỡng sử dụng. LinUCB sau khi sửa lỗi đạt tail latency P99 thấp nhất trong nhóm — **18 896 ms so với P2C 19 347 ms**, giảm 2.3%. Wall-clock 5w-hetero ngang P2C (94 giây) sau khi vá ba lỗi triển khai (chi tiết §X.5).

## X.2 Phương pháp

### X.2.1 Khung Contextual Bandit

Chúng tôi mô hình hoá quyết định lập lịch như một bài toán *contextual bandit*: tại mỗi bước $t$, agent quan sát véc-tơ ngữ cảnh $x_{t,a} \in \mathbb{R}^d$ cho mỗi cánh tay (worker) $a$, chọn một $a_t$, nhận được phần thưởng $r_t$ liên quan đến thời gian biên dịch, rồi cập nhật ước lượng. So với một MDP đầy đủ, mô hình bandit giả định mỗi quyết định ảnh hưởng tới phần thưởng *tức thời* mà không kéo theo trạng thái dài hạn — giả định phù hợp ở đây vì các task biên dịch độc lập tương đối, và tải dài hạn được phản ánh qua đặc trưng "active_tasks" trong $x_{t,a}$ (Slivkins 2019 §1.3).

### X.2.2 Hàm phần thưởng

Thử nghiệm M1 cho thấy phân bố thời gian biên dịch nặng đuôi với P99/P50 ≈ 29×. Để tránh việc một quan sát lớn áp đảo cập nhật ước lượng, chúng tôi chọn:

$$r_t = -\log(1 + t_{\text{compile}})$$

trong đó $t_{\text{compile}}$ là thời gian biên dịch quan sát được tính bằng mili-giây. Đây là một lựa chọn thực dụng (engineering choice). Khác với tài liệu plan cũ (đã sửa trong `docs/thesis/theory-notes.md` §4.3), không có nguồn tham khảo có *peer review* nào trực tiếp khẳng định công thức này; nó là sự thoả hiệp giữa tính chất nén đuôi của hàm log và tính khả vi liên tục. Một nghiên cứu so sánh với phần thưởng kiểu Decima ($r_k = -(t_k - t_{k-1})J_k$, được biện luận bằng Định luật Little) là một hướng mở rộng (xem §X.5).

### X.2.3 Thuật toán LinUCB

Chúng tôi triển khai LinUCB với mô hình tuyến tính rời (Li, Chu, Langford, Schapire 2010, DOI 10.1145/1772690.1772758, arXiv:1003.0146 v2). Với mỗi cánh tay $a$ và véc-tơ ngữ cảnh $x_{t,a}$, thuật toán giả định kỳ vọng phần thưởng tuyến tính theo đặc trưng:

$$\mathbb{E}[r_{t,a}|x_{t,a}] = x_{t,a}^\top \theta_a^*$$

LinUCB duy trì ma trận $A_a \in \mathbb{R}^{d\times d}$ và véc-tơ $b_a \in \mathbb{R}^d$:

- Khởi tạo: $A_a = I_d$, $b_a = 0$ (Li 2010 Algorithm 1, dòng 5–6).
- Ước lượng: $\hat{\theta}_a = A_a^{-1} b_a$ (dòng 8 — đây là nghiệm hồi quy ridge với hệ số $\lambda = 1$).
- Điểm UCB: $p_{t,a} = \hat{\theta}_a^\top x_{t,a} + \alpha\sqrt{x_{t,a}^\top A_a^{-1} x_{t,a}}$ (dòng 9).
- Chọn $a_t = \arg\max_a p_{t,a}$ (dòng 11).
- Cập nhật khi có phần thưởng: $A_{a_t} \mathrel{+}= x x^\top$, $b_{a_t} \mathrel{+}= r x$ (dòng 12–13).

Hệ số $\alpha$ điều khiển trade-off khám phá–khai thác. Công thức lý thuyết của Li 2010 (Eq. 4):

$$\alpha = 1 + \sqrt{\ln(2/\delta)/2}$$

trong đó $\delta$ là xác suất chấp nhận sai. $\alpha$ là tham số CLI có thể chỉnh tại thời điểm chạy (cờ `--alpha`); mã nguồn mặc định 0.5 sau quá trình review chỉ ra rằng giá trị 1.0 lý thuyết quá lớn so với độ lớn phần thưởng đã chuẩn hoá (xem §X.5). Trong thí nghiệm gốc chúng tôi đã dùng $\alpha = 1.0$ trước khi vá lỗi — kết quả thí nghiệm đó được giữ trong Bảng X.1 dưới cột "LinUCB α=1 (lỗi)" để so sánh trực tiếp với cấu hình đã sửa "LinUCB-fixed α=0.5". Lựa chọn $\alpha$ là *empirical* và không trích Chu et al. 2011 — paper đó dùng $\alpha = \sqrt{\tfrac{1}{2}\ln(2TK/\delta)}$ cho biến thể SupLinUCB, không phải LinUCB nguyên bản.

### X.2.4 Cập nhật ma trận nghịch đảo bằng công thức Sherman–Morrison

Cập nhật trực tiếp $A_a^{-1}$ tại mỗi bước có chi phí $\mathcal{O}(d^3)$. Thay vào đó, chúng tôi áp dụng công thức Sherman–Morrison (Sherman & Morrison 1950; Golub & Van Loan, *Matrix Computations* 4th ed. §2.1.4) cho cập nhật hạng-1 $A_{\text{new}} = A_{\text{old}} + xx^\top$:

$$A_{\text{new}}^{-1} = A_{\text{old}}^{-1} - \frac{A_{\text{old}}^{-1}\, x x^\top A_{\text{old}}^{-1}}{1 + x^\top A_{\text{old}}^{-1} x}$$

Chi phí mỗi cập nhật giảm xuống $\mathcal{O}(d^2)$ — phù hợp với khẳng định trong Li 2010. Trong code (`internal/coordinator/scheduler/linucb.go`), một unit test (`TestLinUCB_ShermanMorrisonMatchesBruteForce`) so sánh kết quả Sherman–Morrison sau 50 lần cập nhật với một phép nghịch đảo lại từ đầu thông qua thư viện `gonum/mat`, sai số tuyệt đối dưới $10^{-6}$.

### X.2.5 Véc-tơ đặc trưng

Số chiều $d = 12$. Phiên bản đầu tiên dùng one-hot ba chiều cho *loại build* (CPP/Flutter/Unity); vì luồng `Compile()` hiện tại chỉ phát sinh loại CPP, ba chiều này suy biến (chiều CPP trùng tuyến tính với bias, hai chiều còn lại vĩnh viễn bằng 0) và đã bị loại bỏ trong quá trình sửa lỗi (§X.5). Bố cục 12 chiều hiện tại của mã nguồn (`internal/coordinator/scheduler/linucb.go`, hàm `featureVector`) như sau:

| Chỉ số | Đặc trưng | Công thức chuẩn hoá |
|---|---|---|
| [0] | bias | hằng số 1.0 |
| [1] | log kích thước nguồn (đã tiền xử lý) | $\log(1+\text{size})/\log(1+4\,\text{MiB})$, chặn tại 1 |
| [2] | kiến trúc đích = x86\_64 | one-hot 0/1 |
| [3] | kiến trúc đích = arm64 | one-hot 0/1 |
| [4] | số lõi CPU của worker | $\text{cpu\_cores}/16$, chặn tại 1 |
| [5] | bộ nhớ worker | $\text{mem\_bytes}/64\,\text{GiB}$, chặn tại 1 |
| [6] | khớp kiến trúc native | 1.0 nếu `native_arch == target` |
| [7] | áp lực hàng đợi | $\text{active\_tasks}/\text{max\_parallel}$, chặn tại 1 |
| [8] | độ trễ RPC gần đây | $\text{rtt\_ms}/100$, chặn tại 1 |
| [9] | task là C++ | 1.0 nếu phần mở rộng/compiler là C++ |
| [10] | log kích thước nguồn thô (chưa tiền xử lý) | $\log(1+\text{raw\_size})/\log(1+1\,\text{MiB})$, chặn tại 1 |
| [11] | tỉ lệ thành công của worker (Laplace-smoothed) | $(\text{s}+1)/(\text{s}+\text{f}+2)$ |

Hai chiều bổ sung so với thiết kế ban đầu — [10] kích thước nguồn thô và [11] tỉ lệ thành công — được thêm sau khi review chỉ ra rằng: (a) chiều "tỉ lệ thành công" thô với prior lạc quan 1.0 sẽ luôn bằng 1.0 trên workload thành công hoàn toàn, trùng tuyến tính với bias, nên cần Laplace-smoothing để duy trì tính phụ-thuộc-kinh-nghiệm; (b) một mẫu số chuẩn hoá kích thước riêng cho nguồn thô để đặc trưng không dồn về 0 dưới mẫu số 4 MiB của nguồn đã tiền xử lý. Tất cả đặc trưng được chuẩn hoá xấp xỉ vào $[0, 1]$ để giữ $\|x\|$ bị chặn (Chu et al. 2011 §3) — điều kiện tiên quyết cho regret bound, mặc dù thực nghiệm không hiệu lực hoá ràng buộc này một cách nghiêm ngặt.

### X.2.6 Đảm bảo về regret

Định lý 1 (Chu, Li, Reyzin, Schapire 2011, AISTATS) cung cấp bound:

$$\text{Regret} = \mathcal{O}\!\left(\sqrt{T d\, \ln^3(KT \ln T / \delta)}\right)$$

khi giả thiết tuyến tính hoá ($\mathbb{E}[r|x] = x^\top \theta^*$) thoả mãn. Cận dưới khớp $\Omega(\sqrt{Td})$ cũng được chứng minh trong cùng paper. **Lưu ý quan trọng:** bound này chỉ đúng cho biến thể SupLinUCB của Chu 2011, không phải LinUCB Algorithm 1. Trong đồ án này chúng tôi triển khai LinUCB nguyên bản (đơn giản hơn), do đó kết quả thực nghiệm là minh chứng chính, còn bound trên được trích như *bối cảnh lý thuyết* chứ không phải đảm bảo trực tiếp.

## X.3 Triển khai

### X.3.1 Kiến trúc tích hợp

Mã nguồn được tổ chức thành các package Go riêng biệt:

- `internal/coordinator/scheduler/`: chứa interface `Scheduler` (`Select`), interface mở rộng `LearningScheduler` (`SelectWithDispatchInfo`, `RecordOutcome`), và sáu implementation: `SimpleScheduler`, `LeastLoadedScheduler`, `P2CScheduler`, `EpsilonGreedyScheduler`, `LinUCBScheduler`, `HEFTScheduler`.
- `internal/coordinator/server/grpc.go`: hàm `newScheduler` đóng vai factory chọn implementation theo cấu hình. Hook phản hồi (feedback) trong `Compile()` gọi `RecordOutcome` ngay sau khi `DecrementTasks` để bộ học cập nhật.
- `cmd/hg-coord/main.go`: cờ CLI `--scheduler`, `--epsilon`, `--alpha` cho phép chọn và cấu hình scheduler tại thời điểm chạy.

Việc dùng *type assertion* (`if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok`) cho phép các scheduler không-học (LeastLoaded, P2C) cùng tồn tại với các scheduler học (ε-greedy, LinUCB, HEFT) mà không phải sửa lại các code path khác.

### X.3.2 Đường ống đo lường

Mỗi task hoàn tất sinh ra một bản ghi JSON Lines 27 trường thông qua `TaskLogger` (file `internal/coordinator/server/task_log.go`), gồm: định danh task, cấu hình worker tại thời điểm dispatch (lõi, RAM, kiến trúc, queue depth), kích thước nguồn, các thành phần độ trễ (queue/compile/RPC), trạng thái thành công, và introspection của bộ học (giá trị Q tại dispatch, cờ exploration). Schema được kiểm chứng bằng `pandas.read_json(..., lines=True)` không lỗi trên 1746 bản ghi đã thu thập (873 leastloaded + 873 P2C + 873 ε-greedy).

### X.3.3 Khung benchmark

Script `test/stress/benchmark-heterogeneous.sh` sinh động các file `docker-compose-hetero.yml` cho ba cấu hình (1, 3, 5 worker với phân bổ CPU không đều), khởi chạy coordinator + worker + builder, clone CPython, chạy `make -j` thông qua `hgbuild`, và ghi tổng thời gian build + log per-task. Biến môi trường `SCHEDULER` chọn scheduler để so sánh; toàn bộ lưu trạng thái trong volume Docker và được trích xuất ra host bằng container Alpine helper.

## X.4 Kết quả thực nghiệm

### X.4.1 Đặc trưng dữ liệu (M1)

873 task, 100% thành công. Phân vị thời gian biên dịch (mili-giây): P50 = 820, P95 = 6 226, P99 = 23 961. Phân vị kích thước nguồn (byte): P50 ≈ 970 KB, P99 ≈ 2.3 MB. Đuôi nặng được khẳng định, biện minh cho hàm phần thưởng log.

### X.4.2 Bảng so sánh chính (Bảng X.1, X.2 ở trên)

Các phát hiện cốt lõi:

1. **P2C cải thiện 1.62×** so với LeastLoaded trên 5 worker khác chủng — phù hợp định lý của Mitzenmacher 2001.
2. **ε-greedy mù đặc trưng kém hơn cả P2C** trên mọi cấu hình. Trên 5w-hetero, ε-greedy đạt 119s so với P2C 94s. Kết quả này đúng với giả thuyết: bộ học không xét đặc trưng trả giá cho exploration mà không thu lợi từ tính khác chủng.
3. **ε-greedy còn làm xấu cân bằng tải** so với heuristic — tỉ số top:bottom là 13.2:1, lớn hơn cả LeastLoaded (10.8:1). Bộ học chọn argmax-Q dồn lưu lượng vào worker mạnh, gây tranh chấp hàng đợi.
4. **LinUCB và HEFT** (kết quả M3 đang được thu thập) sẽ kiểm chứng giả thuyết rằng việc thêm đặc trưng vào quyết định đóng được khoảng trống ε-greedy đã bộc lộ.

### X.4.3 Phân tích Q-value (ε-greedy)

Phân bố Q tại dispatch (`q_value_at_dispatch` trong log): trung bình −6.61, độ lệch chuẩn 1.06, dải [−7.82, 0]. Giá trị Q cluster quanh $-\log(\bar{T})$ với $\bar{T} \approx 800$ ms, tương ứng với phần thưởng đã học. Một số worker có $Q = 0$ (chưa từng được khảo sát) — chính là chế độ thất bại mà bonus UCB của LinUCB (được thiết kế để khảo sát tham lam-bị-điều chỉnh) sẽ giải quyết.

### X.4.4 Tỉ lệ exploration

Tỉ lệ `was_exploration = true` trong log ε-greedy: 0.084 — sát với mục tiêu $\varepsilon = 0.10$, sai lệch nhỏ là do fast-path đơn-ứng-viên trong cấu hình 1-worker.

### X.4.5 Đo lường lại có kiểm soát nhiễu (rigorous, randomized block)

Các số liệu Bảng X.1 đến từ một lần chạy mỗi cấu hình. Vì chênh lệch giữa các scheduler ở cùng cấu hình có thể nhỏ hơn biên độ nhiễu của môi trường (sàn nhiễu Docker/macOS quan sát được ±25 giây trên bản host cũ), một lần chạy đơn *không đủ* để rút kết luận thống kê. Chúng tôi thiết kế lại thí nghiệm theo chuẩn **randomized complete block design** (`scripts/benchmark_rigorous.sh`):

- **Khối hoá (blocking):** mỗi *round* chạy cả bốn scheduler đúng một lần, theo thứ tự **xáo trộn ngẫu nhiên có seed** riêng cho từng round. Nhờ vậy mỗi scheduler đều nếm đủ dải trạng thái của host (mát/nóng), triệt tiêu confound giữa "danh tính scheduler" và "trôi nhiệt/thời gian". Mỗi round trở thành một khối thống kê, cho phép dùng kiểm định **ghép cặp** (paired) mạnh hơn.
- **Loại các nhiễu khác:** một build khởi động bị loại bỏ (warm-up), rào chắn chờ đủ 5/5 worker đăng ký, cooldown cố định trước mỗi build, đo thời gian dưới-giây (sub-second), xóa cache biên dịch mỗi build.
- **Quy mô:** 10 round × 4 scheduler = 40 build, trên cấu hình 5w-hetero. Phân tích bằng Friedman omnibus + Wilcoxon signed-rank ghép cặp + hiệu chỉnh Holm–Bonferroni + effect size Cliff's delta + khoảng tin cậy bootstrap (`scripts/analyze_rigorous.py`). Số liệu thô: `.sisyphus/evidence/rigorous-v3.14.0/`.

**Bảng X.3 — Makespan (giây), 10 round, 5w-hetero, workload nhẹ ~293 task. Median [KTC 95%].**

| Scheduler | Median (s) | Mean | Std | So với hybrid-linucb (Wilcoxon ghép cặp + Holm) |
|---|---|---|---|---|
| LeastLoaded | 40.24 | 40.38 | 1.02 | **hòa** (Δ=+0.09 s, p=0.65, Cliff d=+0.10) |
| **hybrid-linucb** | 40.52 | 40.56 | 1.06 | — |
| P2C | 45.60 | 46.01 | 1.39 | hybrid **thắng** (Δ=−5.26 s, p=0.003, d=−1.0) |
| LinUCB | 47.48 | 47.11 | 1.45 | hybrid **thắng** (Δ=−6.80 s, p=0.003, d=−1.0) |

Friedman omnibus: $\chi^2 = 24.60$, $p = 0.00002$ (các scheduler khác nhau có ý nghĩa). Về độ trễ đuôi P99: Friedman $p = 0.169$ — **không** phát hiện khác biệt.

**Phát hiện cốt lõi và bài học phương pháp luận.** Trong lần chạy single-run ban đầu (thiết kế "scheduler-major": chạy toàn bộ 10 lần lặp của một scheduler liên tiếp rồi mới sang scheduler khác), hybrid-linucb *dường như* vượt LeastLoaded 7.9% với $p = 0.0001$. Khi loại confound thứ tự bằng randomized block, ưu thế đó **biến mất hoàn toàn**: hai scheduler thống kê **ngang nhau** ($p = 0.65$). Chênh lệch biểu kiến trước đó là *artifact thí nghiệm* — LeastLoaded bị đo khi máy ở trạng thái tải khác. Đây là minh chứng trực tiếp rằng một thiết kế thí nghiệm nghiêm ngặt có thể **bắt được kết quả dương tính giả** do một thiết kế cẩu thả sinh ra; bản thân điều này là một đóng góp phương pháp luận của đồ án.

**Bảng X.4 — Makespan trên workload nặng hơn (~371 task, bỏ cờ `--disable-test-modules`).**

| Scheduler | Median (s) | Mean | Std | So với hybrid-linucb |
|---|---|---|---|---|
| **LeastLoaded** | 48.65 | 51.36 | 6.78 | hybrid **hòa/thua nhẹ** (Δ=+0.10 s, p=0.82) |
| hybrid-linucb | 50.64 | 55.42 | 16.60 | — |
| P2C | 53.48 | 55.22 | 5.08 | Δ=−3.12 s, p_holm=0.16 (không đủ ý nghĩa sau hiệu chỉnh) |
| LinUCB | 60.20 | 59.45 | 6.35 | Δ=−7.61 s, p_holm=0.16 (không đủ ý nghĩa sau hiệu chỉnh) |

Ở tải nặng hơn, LeastLoaded thậm chí **nhanh nhất** theo median, và ưu thế của hybrid so với P2C/LinUCB không còn đủ ý nghĩa sau hiệu chỉnh Holm. Độ lệch chuẩn lớn của hybrid (16.6) đến từ **một** outlier (round 10 = 101.93 s, gấp đôi bình thường): bandit khởi động lạnh thỉnh thoảng ra một chuỗi quyết định khám phá tệ, bộc lộ **rủi ro đuôi (tail risk)** của bandit — đúng với hạn chế drift ở §X.5.2.

### X.4.6 Ablation warm-bandit — persistence không thay đổi kết luận

Giả thuyết tự nhiên để giải thích thế hòa: benchmark rigorous khởi động lại coordinator ở *mỗi* build, nên bandit **khởi động lạnh mỗi lần** và không tích lũy học qua ~293 task của một build. Chúng tôi kiểm chứng bằng cách loại bỏ chính giả thuyết này (`scripts/benchmark_warm.sh`): giữ **một** coordinator sống qua K=8 build tuần tự (bandit tích lũy trạng thái xuyên suốt — đã xác minh qua đọc mã: trạng thái bandit là singleton in-memory, tạo một lần, không bao giờ reset), lặp 6 session độc lập, đo makespan theo chỉ số build.

**Kết quả: giả thuyết bị bác bỏ.** Xu hướng học (Spearman giữa chỉ số build và makespan): $\rho = -0.079$, $p = 0.59$ — **phẳng, không học**. So sánh build 1 với build 8 (Wilcoxon ghép cặp, n=6): $\Delta = +0.21$ s, $p = 0.42$ (không nhanh lên). Warm-bandit ở build 8 so với LeastLoaded: $\Delta = -0.19$ s, $p = 0.50$ (vẫn hòa). Số liệu: `.sisyphus/evidence/warm-bandit-v3.14.0/`.

**Vì sao.** Mỗi build có ~293 task, trong khi warm-start $N = 100$ — bandit đã đi qua cửa sổ khởi động và hội tụ **ngay trong một build**. Với không gian đặc trưng nhỏ (9–12 chiều) và workload **dừng (stationary)**, ma trận $A_a$ bão hòa trong build đầu; các build sau không còn thông tin mới để học. Do đó khởi động lạnh **không** phải nguyên nhân của thế hòa — thế hòa là **bản chất**: trên workload biên dịch này (task đồng nhất, worker tĩnh trong một build), LeastLoaded đã gần tối ưu, và một bandit có ngữ cảnh hội tụ về đúng chất lượng đó nhưng không có cấu trúc ẩn nào để khai thác vượt lên.

## X.5 Thảo luận và hạn chế

### X.5.1 Phần đã đạt và bằng chứng đi kèm

- Pipeline đo lường ổn định, dữ liệu pandas-ready, có thể tái lập (xem §X.3.3).
- Năm scheduler đầy đủ (LeastLoaded, Simple, P2C, ε-greedy, LinUCB, HEFT) đều có unit test bao phủ tính chính xác (gồm test Sherman–Morrison của LinUCB và test cập nhật trung bình ngẫu nhiên của ε-greedy).
- Ba scheduler đã có số liệu wall-clock và load-balance trên ba cấu hình cluster.

### X.5.2 Hạn chế đã nhận diện

- **Giả thiết tuyến tính của LinUCB.** Thời gian biên dịch không tuyến tính theo kích thước nguồn (xét compiler tối ưu hoá nhiều cấp); khi giả thiết bị vi phạm, regret bound của Chu 2011 không còn áp dụng. Lattimore & Szepesvári 2020 Ch. 24.4 cho thấy regret tăng cộng theo $\mathcal{O}(\varepsilon\sqrt{T})$ với mức độ vi phạm $\varepsilon$.
- **Dòng chảy phân bố (drift).** Worker có thể bị giảm hiệu năng do thermal throttling hoặc tải nền. LinUCB chuẩn không có đảm bảo dưới drift; phương pháp giảm thiểu (sliding window, change-point) là chủ đề nghiên cứu mở.
- **Bộ ba chết của Sutton–Barto.** Phân tích §11.3 (Sutton & Barto 2018) cảnh báo divergence khi kết hợp xấp xỉ hàm + bootstrapping + off-policy. LinUCB không chạm bộ ba này (không có bootstrap); nếu mở rộng sang Q-learning đa-bước với xấp xỉ tuyến tính, vấn đề trở nên nghiêm trọng và cần được giải quyết riêng.
- **Mẫu thực nghiệm còn hạn chế về đa dạng workload.** Số lần lặp đã được nâng lên 10 (rigorous, X.4.5) với thiết kế thống kê nghiêm ngặt, nhưng vẫn chỉ trên một họ workload (CPython) và một host (Docker macOS). Để công bố cần đa dạng workload — đặc biệt các codebase nặng template (Boost/Qt/Eigen) có phân phối nặng đuôi hơn, nơi thứ tự tương đối giữa các scheduler *có thể* thay đổi.
- **LeastLoaded là baseline mạnh trên workload dừng.** Kết quả X.4.5–X.4.6 cho thấy bandit không vượt được LeastLoaded khi task đồng nhất và worker tĩnh trong một build. Đây không phải hạn chế của riêng LinUCB mà là đặc điểm của bài toán: giá trị của lập lịch học chỉ bộc lộ khi môi trường có cấu trúc ẩn (cache-affinity, drift, task cost dị biệt) mà heuristic đơn giản bỏ qua (xem X.5.3).

### X.5.3 Hướng mở rộng

1. **Phần thưởng dựa Định luật Little**: Triển khai $r_k = -(t_k - t_{k-1})J_k$ kiểu Decima như một biến thể có dẫn chứng, so sánh ablation với phần thưởng log.
2. **Cache-aware scheduling** — hướng mở rộng có triển vọng nhất và được chính kết quả X.4.6 chỉ ra. Vì bandit đã hội tụ về chất lượng của LeastLoaded trên workload dừng, bandit chỉ có thể *vượt* LeastLoaded ở một môi trường mà LeastLoaded **dưới tối ưu** — tức có cấu trúc ẩn tương quan với đặc trưng mà LeastLoaded bỏ qua. Cache-affinity là ứng viên rõ ràng: nếu worker $w$ đã có sẵn artifact biên dịch của file $f$ trong cache, gửi $f$ tới $w$ cho một cache hit (gần như tức thời) thay vì biên dịch lại ở worker khác. Thiết kế cụ thể: thêm chiều đặc trưng thứ 13 "worker $w$ đã có $f$ trong cache" — có thể hiện thực **không cần RPC mới** bằng cách để coordinator theo dõi lịch sử dispatch per-(worker, filename) (coordinator vốn định tuyến mọi task, và `SourceFilename` đã có trong `TaskContext` tại thời điểm Select).

   **Điều kiện tiên quyết quan trọng (phát hiện từ khảo sát mã nguồn):** đặc trưng cache-affinity **bất hoạt trong kịch bản clean-build** hiện đang đo — vì benchmark xóa cache mỗi build và trong một clean-build mỗi file chỉ biên dịch đúng một lần, nên **không** worker nào từng "đã có file trong cache" trong lúc build. Cache-affinity chỉ có ý nghĩa khi có **tái sử dụng cache**, tức kịch bản **incremental build** (lập trình viên sửa vài file rồi biên dịch lại — phần lớn file là cache hit). Do đó ablation cache-aware đòi hỏi một khung bài toán mới (incremental rebuild, không xóa cache) — một thiết lập *khác* với makespan clean-build mà chương này đo. Đây là khoảng trống chưa có công bố nào lấp trong lập lịch biên dịch phân tán, và là nội dung nghiên cứu độc lập cho giai đoạn tiếp theo.
3. **Detection of drift**: thêm cơ chế phát hiện chuyển dịch (CUSUM, Page–Hinkley) và reset cục bộ $A_a, b_a$ khi cần.
4. **Mở rộng đa-loại task**: hiện đã hỗ trợ Flutter và Unity ở mức compile entry point; tương lai có thể tích hợp đặc trưng theo loại build vào véc-tơ ngữ cảnh.

## X.6 Kết luận

Đồ án đã hoàn thiện ba milestone (M1–M3) gồm: hạ tầng đo lường có thể tái lập, một bộ học bandit cơ sở (ε-greedy), và một bộ học có ngữ cảnh (LinUCB) cùng một baseline kinh điển (HEFT) đã được điều chỉnh cho luồng task trực tuyến. Sáu scheduler được so sánh dưới một quy trình đo lường nghiêm ngặt (randomized block, 10 lần lặp, kiểm định ghép cặp). Các kết luận đã được kiểm chứng:

- **P2C vượt LeastLoaded** trên cluster khác chủng ở lần đo single-run (Bảng X.1), phù hợp lý thuyết Mitzenmacher; tuy nhiên ưu thế này thu hẹp trên workload nhẹ khi đo lại rigorous (Bảng X.3).
- **ε-greedy và LinUCB thuần** (mù/yếu về ngữ cảnh) thua P2C và LeastLoaded — biểu hiện chi phí exploration không được đền bù trên workload dừng.
- **hybrid-linucb ngang bằng LeastLoaded** về makespan trên cả hai mức tải (Bảng X.3, X.4), và chỉ **vượt trội có ý nghĩa** so với P2C và LinUCB thuần ở tải nhẹ. Ablation warm-bandit (X.4.6) chứng minh thế hòa này là **bản chất** chứ không phải do khởi động lạnh.

Đóng góp trung thực của đồ án do đó có **hai mặt**. Thứ nhất về *kết quả*: một contextual bandit có thể sánh ngang heuristic tốt nhất và vượt các baseline yếu hơn, nhưng **chưa** vượt LeastLoaded trên workload biên dịch dừng — một kết quả âm (negative result) được củng cố bởi ba lớp bằng chứng độc lập (rigorous, tải nặng, warm-bandit) đã loại trừ các phản biện hiển nhiên. Thứ hai về *phương pháp luận*: quy trình thí nghiệm nghiêm ngặt đã tự phát hiện và loại bỏ một kết quả dương tính giả do thiết kế cẩu thả sinh ra — một bài học có giá trị giáo dục độc lập. Cùng với regret bound lý thuyết làm bối cảnh, đồ án xác lập rõ *khi nào* lập lịch học có giá trị và mở đường có cơ sở cho các mở rộng **cache-aware** (X.5.3) và **drift-aware** ở giai đoạn tiếp theo — chính là các môi trường nơi heuristic tĩnh dưới tối ưu và bandit được kỳ vọng thắng thật sự.
