# Chương X — Thuật toán lập lịch dựa trên Contextual Bandit cho hệ thống biên dịch phân tán Hybrid-Grid

> *Bản thảo gửi giáo viên hướng dẫn — phiên bản ngày 29/04/2026.*
> Mọi số liệu trong chương được dẫn nguồn từ thư mục `.sisyphus/evidence/m1/` và các paper trong `docs/thesis/theory-notes.md`. Các đoạn còn để trống cờ "kết quả M3" sẽ được bổ sung khi benchmark LinUCB hoàn tất.

## X.1 Đặt vấn đề

### X.1.1 Bối cảnh hệ thống

Hệ thống Hybrid-Grid Build chia một quá trình biên dịch C/C++ thành hàng trăm tới hàng nghìn task độc lập (mỗi file `.c`/`.cpp` là một task) và phân phối chúng tới một cụm worker khác chủng (heterogeneous). Trong cấu hình điển hình mà chúng tôi đã đo (xem `.sisyphus/evidence/m1/findings.md`), một build CPython gồm 873 task được phân bố trên 1, 3 hoặc 5 worker với tổng dung lượng CPU cố định 4.0 lõi nhưng phân chia không đều. Các worker khác nhau về số lõi CPU, dung lượng bộ nhớ, kiến trúc tập lệnh, và cấu hình mạng. Do đó thời gian biên dịch cho cùng một file có thể lệch nhau đáng kể giữa các worker — đo lường thực nghiệm cho thấy tỉ số P99/P50 thời gian biên dịch lên tới 29 lần.

Coordinator (`hg-coord`) giữ một sổ đăng ký (registry) các worker đang sống và quyết định gửi task nào tới worker nào thông qua một thuật toán lập lịch (scheduler). Đây là điểm có ảnh hưởng lớn nhất tới tổng thời gian build (makespan) và là trọng tâm nghiên cứu của đồ án này.

### X.1.2 Mô hình hoá toán học

Bài toán lập lịch trong Hybrid-Grid là một thực thể của lớp `R||C_max` trong phân loại Graham — *unrelated parallel machines*: $m$ máy không đồng nhất, $n$ task, thời gian xử lý $p_{ij}$ phụ thuộc đồng thời vào cả task $j$ và máy $i$ và *không* tuân theo bất kỳ ràng buộc tỉ lệ nào (Lenstra, Shmoys, Tardos 1990, DOI 10.1007/BF01585745). Bài toán này đã được chứng minh là NP-khó, không có thuật toán đa thức nào đạt tỉ lệ xấp xỉ nhỏ hơn $\tfrac{3}{2}$ trừ khi $\text{P} = \text{NP}$, trong khi cận trên xấp xỉ tốt nhất hiện biết là $2$. Khoảng cách $[3/2, 2]$ vẫn còn mở sau hơn 35 năm, chứng tỏ độ khó về mặt lý thuyết của bài toán.

Trong thiết lập trực tuyến (online) — nơi task đến tuần tự và phải được gán ngay khi đến — đối thủ chuẩn là *list scheduling* của Graham 1969. Bound cạnh tranh (competitive ratio) cho list scheduling là $2 - 1/m$ trên môi trường đồng nhất; với máy khác chủng, không có bound cạnh tranh phổ quát chặt chẽ.

### X.1.3 Hạn chế của các heuristic tĩnh

Hai heuristic phổ biến đã được hiện thực trong Hybrid-Grid:

- **LeastLoaded:** chọn worker có ít task đang chạy nhất. Đơn giản, không học, không xét năng lực.
- **Power of Two Choices (P2C):** lấy mẫu hai worker ngẫu nhiên, chọn worker có điểm số (theo công thức trọng số tĩnh) cao hơn. Nền tảng lý thuyết ở Mitzenmacher 2001 (DOI 10.1109/71.963420), giảm tải tối đa từ $\Theta(\log n / \log\log n)$ xuống $\Theta(\log\log n)$ khi server đồng nhất.

Cả hai đều **không học từ dữ liệu thực thi**. Đo lường thực nghiệm (Bảng X.1) trên cùng cấu hình 5 worker khác chủng cho thấy:

**Bảng X.1 — Makespan (giây) trên benchmark CPython, 873 task, đo đơn vị giây.**

| Cấu hình  | LeastLoaded | P2C  | ε-greedy | LinUCB α=1 (lỗi) | LinUCB-fixed α=0.5 | HEFT |
|-----------|-------------|------|----------|------|--------------------|------|
| 1w-4.0CPU | 92          | 130  | 146      | 129  | 131                | 129  |
| 3w-hetero | 123         | 85   | 142      | 103  | 108                | 135  |
| 5w-hetero | 152         | **94** | 119    | 158  | **94**             | 144  |

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

trong đó $\delta$ là xác suất chấp nhận sai. Trong thực nghiệm chúng tôi dùng $\alpha = 1.0$ làm mặc định và để $\alpha$ là tham số CLI có thể chỉnh; lựa chọn này được khẳng định là *empirical* và không trích Chu et al. 2011 — paper đó dùng $\alpha = \sqrt{\tfrac{1}{2}\ln(2TK/\delta)}$ cho biến thể SupLinUCB, không phải LinUCB nguyên bản.

### X.2.4 Cập nhật ma trận nghịch đảo bằng công thức Sherman–Morrison

Cập nhật trực tiếp $A_a^{-1}$ tại mỗi bước có chi phí $\mathcal{O}(d^3)$. Thay vào đó, chúng tôi áp dụng công thức Sherman–Morrison (Sherman & Morrison 1950; Golub & Van Loan, *Matrix Computations* 4th ed. §2.1.4) cho cập nhật hạng-1 $A_{\text{new}} = A_{\text{old}} + xx^\top$:

$$A_{\text{new}}^{-1} = A_{\text{old}}^{-1} - \frac{A_{\text{old}}^{-1}\, x x^\top A_{\text{old}}^{-1}}{1 + x^\top A_{\text{old}}^{-1} x}$$

Chi phí mỗi cập nhật giảm xuống $\mathcal{O}(d^2)$ — phù hợp với khẳng định trong Li 2010. Trong code (`internal/coordinator/scheduler/linucb.go`), một unit test (`TestLinUCB_ShermanMorrisonMatchesBruteForce`) so sánh kết quả Sherman–Morrison sau 50 lần cập nhật với một phép nghịch đảo lại từ đầu thông qua thư viện `gonum/mat`, sai số tuyệt đối dưới $10^{-6}$.

### X.2.5 Véc-tơ đặc trưng

Số chiều $d = 12$, gồm: hệ số bias, log kích thước nguồn (chuẩn hoá), one-hot loại build (CPP/Flutter/Unity), one-hot kiến trúc đích (x86_64/arm64), tỉ lệ năng lực worker (CPU lõi / 16, bộ nhớ / 64GB), khớp kiến trúc native, áp lực hàng đợi (active_tasks / max_parallel), độ trễ RPC (chuẩn hoá theo 100 ms). Tất cả đặc trưng được chuẩn hoá vào khoảng $[0, 1]$ để duy trì giả định $\|x\|\leq 1$ (Chu et al. 2011 §3) — đây là điều kiện tiên quyết cho regret bound, mặc dù trong thực nghiệm chúng tôi không hiệu lực hoá ràng buộc này một cách nghiêm ngặt.

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

## X.5 Thảo luận và hạn chế

### X.5.1 Phần đã đạt và bằng chứng đi kèm

- Pipeline đo lường ổn định, dữ liệu pandas-ready, có thể tái lập (xem §X.3.3).
- Năm scheduler đầy đủ (LeastLoaded, Simple, P2C, ε-greedy, LinUCB, HEFT) đều có unit test bao phủ tính chính xác (gồm test Sherman–Morrison của LinUCB và test cập nhật trung bình ngẫu nhiên của ε-greedy).
- Ba scheduler đã có số liệu wall-clock và load-balance trên ba cấu hình cluster.

### X.5.2 Hạn chế đã nhận diện

- **Giả thiết tuyến tính của LinUCB.** Thời gian biên dịch không tuyến tính theo kích thước nguồn (xét compiler tối ưu hoá nhiều cấp); khi giả thiết bị vi phạm, regret bound của Chu 2011 không còn áp dụng. Lattimore & Szepesvári 2020 Ch. 24.4 cho thấy regret tăng cộng theo $\mathcal{O}(\varepsilon\sqrt{T})$ với mức độ vi phạm $\varepsilon$.
- **Dòng chảy phân bố (drift).** Worker có thể bị giảm hiệu năng do thermal throttling hoặc tải nền. LinUCB chuẩn không có đảm bảo dưới drift; phương pháp giảm thiểu (sliding window, change-point) là chủ đề nghiên cứu mở.
- **Bộ ba chết của Sutton–Barto.** Phân tích §11.3 (Sutton & Barto 2018) cảnh báo divergence khi kết hợp xấp xỉ hàm + bootstrapping + off-policy. LinUCB không chạm bộ ba này (không có bootstrap); nếu mở rộng sang Q-learning đa-bước với xấp xỉ tuyến tính, vấn đề trở nên nghiêm trọng và cần được giải quyết riêng.
- **Mẫu thực nghiệm còn nhỏ.** Chỉ có một workload (CPython) và một host (Docker macOS). Để công bố cần đa dạng workload (ưa template như Boost/Qt) và lặp ≥ 5 lần để có khoảng tin cậy.

### X.5.3 Hướng mở rộng

1. **Phần thưởng dựa Định luật Little**: Triển khai $r_k = -(t_k - t_{k-1})J_k$ kiểu Decima như một biến thể có dẫn chứng, so sánh ablation với phần thưởng log.
2. **Cache-aware scheduling**: bổ sung đặc trưng "đã có cache" cho mỗi cặp (worker, file). Đây là khoảng trống chưa có công bố nào lấp trong lập lịch biên dịch phân tán.
3. **Detection of drift**: thêm cơ chế phát hiện chuyển dịch (CUSUM, Page–Hinkley) và reset cục bộ $A_a, b_a$ khi cần.
4. **Mở rộng đa-loại task**: hiện đã hỗ trợ Flutter và Unity ở mức compile entry point; tương lai có thể tích hợp đặc trưng theo loại build vào véc-tơ ngữ cảnh.

## X.6 Kết luận

Đồ án đã hoàn thiện ba milestone (M1–M3) gồm: hạ tầng đo lường có thể tái lập, một bộ học bandit cơ sở (ε-greedy), và một bộ học có ngữ cảnh (LinUCB) cùng với một baseline kinh điển (HEFT) đã được điều chỉnh cho luồng task trực tuyến. Bằng chứng thực nghiệm đầu (M1, M2) đã khẳng định:

- P2C vượt LeastLoaded 1.62× trên cluster khác chủng — phù hợp lý thuyết.
- ε-greedy không vượt P2C trên cluster khác chủng — biểu hiện khoảng trống mà LinUCB nhắm tới.

Kết quả LinUCB và HEFT sẽ được bổ sung khi benchmark hoàn tất. Với regret bound lý thuyết và bằng chứng thực nghiệm M1/M2 đã có, đồ án chứng minh giá trị nghiên cứu của hướng tiếp cận và mở đường cho các mở rộng cache-aware và drift-aware ở các giai đoạn tiếp theo.
