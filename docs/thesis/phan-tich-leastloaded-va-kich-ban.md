# Phân tích nhược điểm của LeastLoaded và thiết kế kịch bản thực nghiệm

**Tài liệu đề xuất gửi thầy duyệt trước khi chạy đợt thực nghiệm mới**
Người viết: Nguyễn Trung Kiên — ngày 04/10/2026
Trạng thái: bản nháp, chưa chạy thí nghiệm nào; mọi con số "dự đoán" dưới đây là giả thuyết, chưa phải kết quả.

---

## 0. Tóm tắt

1. HG-LinUCB hoà LeastLoaded (LL) không phải vì bộ học kém, mà vì kịch bản hiện tại không có thứ gì để học: ở `-j5` mọi quyết định đều có worker rảnh (Mục 7.4), và mọi khác biệt giữa các worker đều nằm trong cấu hình tĩnh.
2. LL có bảy nhược điểm cấu trúc (Mục 3). Nhưng **chỉ ba trong số đó thực sự cần học mới sửa được**: drift ẩn, tương tác tác vụ × worker, và lỗi một phần (gray failure). Các nhược điểm còn lại một heuristic một dòng (SED, ngưỡng kích thước, circuit breaker, affinity) đã sửa được, nên thắng LL nhờ chúng không chứng minh được giá trị của việc học.
3. Audit code cho thấy mười điểm phải xử lý hoặc lưu ý trước khi chạy (Mục 2, P1–P10), trong đó quan trọng nhất: thành phần phạt tải $\lambda\cdot\text{LoadRatio}$ của HG-LinUCB **thừa hưởng đúng điểm mù của LL**, nên trong chính các kịch bản làm LL thua, HG-LinUCB có nguy cơ thua cả LinUCB thuần.
4. Đề xuất sáu kịch bản (Mục 5), ánh xạ vào Thí nghiệm 1–5 của thầy, cộng một kịch bản mới (gray failure). Với mỗi kịch bản ghi rõ dự đoán thứ tự và **điều kiện khiến dự đoán sai**.
5. Mục 8 liệt kê các quyết định cần thầy chốt trước khi viết `hypotheses.md`.

---

## 1. Vì sao kịch bản hiện tại không phân biệt được

Ở `-j5`, 100% quyết định dispatch được đưa ra khi có ít nhất một worker hoàn toàn rảnh (Bảng 5). Khi đó bài toán thu về câu hỏi "có đặt tác vụ lên máy rảnh hay không", và LL trả lời đúng theo định nghĩa (0,0% bỏ qua worker rảnh). Ở `-j10` mọi scheduler hội tụ. Ngoài ra workload CPython có ba đặc điểm khiến bộ học không có lợi thế:

- toàn bộ là C, nên đặc trưng "tác vụ là C++" là hằng số;
- phân phối chi phí tương đối đều, đuôi không đủ nặng để một quyết định sai ở cuối build kéo dài makespan;
- môi trường dừng (stationary): tốc độ mỗi worker không đổi trong suốt build, và tốc độ đó đã được mã hoá sẵn trong cấu hình.

Nói cách khác, mọi thông tin cần để lập lịch tốt đều quan sát được trực tiếp, nên không có gì để học từ phần thưởng.

---

## 2. Phát hiện từ audit code ảnh hưởng tới thiết kế thí nghiệm

Các điểm sau được kiểm tra trực tiếp trên mã nguồn nhánh `KienNT` (commit `66dd69b`).

**P1. Phạt tải $\lambda\cdot\text{LoadRatio}$ thừa hưởng điểm mù của LL.**
`LoadRatio = ActiveTasks / MaxParallel` (`linucb.go:550`). Thành phần này luôn đẩy lựa chọn về worker *đang ít tải*. Trong các kịch bản mà worker "xấu" lại là worker ít tải — vì tác vụ trên nó fail nhanh, hoặc vì nó có nhiều slot (worker 1,1 CPU có 3 slot nên LoadRatio thấp hơn các worker khác) — phạt tải đẩy HG-LinUCB về đúng lựa chọn sai của LL. HG-LinUCB chỉ thoát được khi chênh lệch giá trị đã học lớn hơn $\lambda\cdot\Delta\text{LoadRatio}$, tối đa 0,5 với $\lambda = 0{,}5$. Phần thưởng đã chuẩn hoá về $[-1, 0]$ nên khoảng chênh lệch học được thường chỉ vài phần mười. Đây là rủi ro chính đối với chuỗi thứ tự LL < LinUCB < HG-LinUCB.

**P2. Phần thưởng chỉ đo thời gian biên dịch trên worker.**
`reward = −log1p(CompilationTimeMs) / log1p(timeout)` (`grpc.go:1011`). Thời gian chờ hàng đợi, truyền dữ liệu và độ trễ mạng không có trong phần thưởng. Hệ quả: (a) kịch bản "worker sau đường mạng chậm" không thể giúp bộ học; (b) bộ học tối ưu thời gian compile từng tác vụ, không trực tiếp tối ưu makespan.

**P3. Đặc trưng [8] (độ trễ RPC) là hằng số.**
LinUCB tự tạo `LatencyTracker` riêng (`linucb.go:145-148`) nhưng không nơi nào gọi `Record` cho tracker đó (lời gọi duy nhất nằm trong `P2CScheduler.ReportSuccess`, mà chính hàm này cũng không được gọi ở đâu). Vậy `Get` luôn trả giá trị mặc định 100 ms, tức $x[8] = 1{,}0$ trên mọi worker và cộng tuyến với bias. Đây là cùng loại lỗi với R4 (đặc trưng hằng số âm thầm) mà bài báo đã mô tả.

**P4. Circuit breaker không được nạp dữ liệu.**
`CircuitManager.Execute` không được gọi ở đâu trong đường dispatch; chỉ có `IsOpen`/`GetState` được đọc. Do đó circuit breaker không bao giờ mở, và bộ lọc "CB đóng" mô tả ở Mục 3.3 của bài báo luôn đúng một cách tầm thường. Riêng `LeastLoadedScheduler` thậm chí không nhận `CircuitChecker`. Với kịch bản gray failure, điều này có nghĩa là hiện không scheduler nào có cơ chế loại worker lỗi ngoài việc học; một phản biện sẽ đòi baseline "LL + circuit breaker thật".

**P5. Lỗi biên dịch thật sẽ làm hỏng build, không phải làm chậm build.**
Nếu worker trả về exit code khác 0, client trả thẳng lỗi đó cho `make` (`internal/cli/build/build.go`, nhánh "non-zero exit code"), nên build thất bại. Chỉ lỗi ở tầng hạ tầng (lỗi gRPC, timeout) mới khiến client chuyển sang compile cục bộ (local fallback). Vì vậy mọi kịch bản tiêm lỗi phải tiêm ở tầng RPC, không phải làm compiler fail.

**P6 (ảnh hưởng tới bài báo hiện tại). Đặc trưng tĩnh của worker là dư thừa trong mô hình disjoint.**
HG-LinUCB dùng LinUCB *disjoint*: mỗi worker có $\theta_a$ riêng. Với một worker cố định, các đặc trưng tĩnh của nó ([4] CPU, [5] RAM, [6] khớp kiến trúc) là hằng số qua mọi quan sát của arm đó, nên chúng chỉ là bản sao (có tỉ lệ) của bias [0]. Về mặt lớp hàm, mô hình không "nhìn thấy năng lực" qua [4]; nó học tốc độ từng worker qua hệ số chặn của chính arm đó, từ phần thưởng. Bản vá cgroup (R4) chỉ thay đổi độ lớn của hằng số đó, tức thay đổi cách ridge regularization co các ước lượng và độ rộng bonus (cả lúc đầu lẫn trong suốt quá trình học), chứ không cho mô hình học được quan hệ "năng lực → hiệu năng" *xuyên* các worker. (Phạm vi của nhận định: cụm hiện tại, một kiến trúc đích, năng lực worker cố định; chiều [6] sẽ biến thiên nếu một worker nhận nhiều kiến trúc đích qua cross-compile.) Kết luận ở Mục 7.6 của bài ("bộ học nhận được tín hiệu năng lực và hành động theo nó") vì vậy cần được kiểm chứng lại trước khi giữ nguyên trong bản nộp. Hệ quả cho thiết kế kịch bản: những gì LinUCB thực sự học được theo từng worker là **(i) tốc độ nền, (ii) độ dốc theo kích thước tác vụ, (iii) chênh lệch C/C++, (iv) phản ứng theo tải và tỉ lệ thành công**. Kịch bản phải khai thác đúng các kênh này.

**P7. Mã nguồn đã thay đổi so với lúc thu kết quả của bài báo.**
Coordinator hiện đã có hàng đợi dispatch có giới hạn (`DispatchQueueTimeout`, mặc định 30 s, `grpc.go:310-332`), trong khi Mục 3.3 và 8.3 của bài báo vẫn mô tả cơ chế retry ~700 ms rồi báo lỗi và "chưa có backpressure thật". Vì vậy: (a) mức `-j` tối đa chạy được có thể đã khác; (b) **các baseline phải được chạy lại trên cùng revision với kịch bản mới**, không trộn với số liệu cũ; (c) mỗi lần chạy ghi commit hash, và đếm riêng theo từng scheduler số tác vụ hết hạn chờ trong hàng đợi và số lần client phải compile cục bộ (fallback), vì cả hai đều làm thay đổi makespan.

**P8. Phần thưởng log nén chính cái đuôi mà kịch bản C++ muốn cải thiện.**
Phép biến đổi $-\log(1 + t)$ làm chênh lệch giữa tác vụ 5 s và 50 s nhỏ đi rất nhiều so với chênh lệch giữa 0,1 s và 1 s. Bộ học vì thế tối ưu thời gian compile *trung bình theo thang log* của từng tác vụ, trong khi makespan của một build đuôi nặng thường do một vài tác vụ dài nhất quyết định. Ngoài ra mọi tác vụ lỗi nhận cùng một mức phạt $-1$ bất kể tốn bao lâu. Đây không phải lý do để đổi phần thưởng ngay (đổi là đổi phương pháp), nhưng phải được nêu như một giới hạn khi diễn giải S2 và S3.

**P9. Task log chưa đủ để phân tích ở mức quyết định.**
Mỗi bản ghi chỉ chứa worker được chọn và tải của worker đó; không có trạng thái của các ứng viên khác, cũng không có điểm số của bộ học cho từng ứng viên. Muốn tính "chọn đúng worker mạnh nhất", time-to-adapt hay regret so với một oracle thì cần **ghi thêm ảnh chụp các ứng viên tại thời điểm quyết định** (ActiveTasks, LoadRatio, điểm $\hat\theta^\top x$, bonus, phạt). Các lần chọn bị worker từ chối (`ResourceExhausted`) rồi chọn lại cũng không đưa phần thưởng nào về bộ học, nên phải được đánh dấu riêng.

**P10. Thời gian compile trung bình trong registry bị kéo xuống bởi tác vụ lỗi.**
`DecrementTasks` (`registry.go:351-361`) tăng `FailedTasks` khi tác vụ lỗi nhưng vẫn chạy phép cập nhật trung bình với `compileTime = 0` và mẫu số `SuccessfulTasks` không đổi. Mỗi tác vụ lỗi làm `AvgCompileTime` của worker *giảm*. Mọi thành phần đọc giá trị này (HEFT, dashboard) sẽ thấy worker lỗi nhanh hơn thực tế. Phải sửa trước S4.

**Bảng 1.** Các chiều của véc-tơ ngữ cảnh có mang tín hiệu hay không, trên workload hiện tại và trên các kịch bản đề xuất.

| Chiều | Ý nghĩa | CPython hiện tại | Kịch bản đề xuất |
|---|---|---|---|
| [0] | bias (tốc độ nền của worker) | có | có |
| [1], [10] | kích thước tác vụ | có, nhưng phân phối hẹp | có, đuôi nặng (S2, S3) |
| [2], [3], [6] | kiến trúc đích / khớp kiến trúc | hằng số | hằng số (cùng host x86-64) |
| [4], [5] | CPU, RAM của worker | dư thừa (hằng số trong mỗi arm) | dư thừa |
| [7] | LoadRatio | có | có |
| [8] | độ trễ RPC | **hằng số do P3** | hằng số trừ khi sửa P3 |
| [9] | tác vụ là C++ | hằng số (toàn C) | có, nếu workload trộn C/C++ |
| [11] | tỉ lệ thành công (Laplace) | gần như chỉ đếm số tác vụ đã xong | có (S4) |

---

## 3. Nhược điểm cấu trúc của LeastLoaded

LL (`scheduler.go:139-180`): trong các worker còn slot, chọn worker có `ActiveTasks` nhỏ nhất; hoà thì lấy worker đứng trước. Không trạng thái, không bộ nhớ, không circuit breaker.

**Bảng 2.** Nhược điểm của LL, kênh mà LinUCB khai thác, và liệu một heuristic đơn giản có sửa được không.

| # | Nhược điểm | LinUCB khai thác qua | Heuristic đơn giản sửa được? | Cần học? |
|---|---|---|---|---|
| W1 | **Mù năng lực**: so `ActiveTasks` tuyệt đối; worker 0,5 CPU rảnh được ưu tiên hơn worker 1,1 CPU đang chạy 1 tác vụ | bias theo từng arm | **Có — SED** | Không |
| W2 | **Mù kích thước tác vụ**: TU lớn có thể rơi vào máy yếu ở cuối build, thành straggler | độ dốc theo [1], [10] trong từng arm | Một phần — ngưỡng P90 | Một phần |
| W3 | **Mù drift ẩn**: worker chậm đi mà quota không đổi (throttling, noisy neighbour) | phần thưởng xấu đi → bias của arm dịch chuyển | Không (SED đọc quota, quota không đổi) | **Có** |
| W4 | **Mù tương tác tác vụ × worker**: worker nhanh với tác vụ nhỏ nhưng rất chậm với tác vụ lớn (thiếu RAM), hoặc chậm riêng với C++ | độ dốc riêng của từng arm theo [1], [10], [9] | **Không** — heuristic P90 còn chọn sai, vì gửi TU lớn vào máy "mạnh nhất" | **Có** |
| W5 | **Hố đen (gray failure)**: worker lỗi một phần trả về nhanh, nên luôn có `ActiveTasks` thấp và LL dồn thêm việc vào | phần thưởng −1 cho tác vụ lỗi, [11] | Một phần — circuit breaker, nhưng chỉ khi tỉ lệ lỗi vượt ngưỡng (60%) | **Có**, với lỗi dưới ngưỡng |
| W6 | Mù độ trễ mạng | [8] — nhưng hằng số (P3), và phần thưởng không chứa chi phí mạng (P2) | Có — đo RTT | Hiện không khai thác được |
| W7 | Mù locality của cache | chưa có đặc trưng | Có — heuristic affinity | Một phần (đánh đổi affinity ↔ cân bằng tải) |

**Hai nhận xét định hướng thiết kế.**

- *LL tự sửa sai được, nhưng chỉ khi có hàng đợi.* Worker chậm giữ tác vụ lâu hơn nên `ActiveTasks` của nó cao hơn và LL tự né. Cơ chế này chỉ có tác dụng khi tải đủ cao để hàng đợi hình thành. Ở tải thấp và trung bình, LL hoàn toàn mù với W3 và W4. Ngược lại, với W5 (lỗi trả về nhanh), cơ chế tự sửa này **chạy ngược**: worker lỗi càng nhanh càng trông "rảnh".
- *Chỉ W3, W4, W5 là nơi việc học có giá trị không thay thế được.* Kịch bản chủ lực phải nhắm vào ba nhược điểm này, và luôn có SED cùng heuristic tương ứng làm đối chứng, để mọi chiến thắng đều quy được cho việc học.

---

## 4. Điều kiện để có chuỗi thứ tự LL < LinUCB < HG-LinUCB

**Bậc 1 — LinUCB thắng LL.** LinUCB thuần trả một chi phí khám phá đáng kể: ở `-j5` nó bỏ qua worker rảnh 35% số lần và chậm hơn HG-LinUCB 8,95 s. Vậy nó chỉ thắng LL khi *tổng thiệt hại do các quyết định sai của LL* lớn hơn chi phí đó. Kịch bản phải khiến một quyết định sai của LL **đắt** (gấp nhiều lần thời gian một tác vụ: swap, fallback cục bộ, straggler cuối build), chứ không chỉ chậm hơn vài chục phần trăm.

**Bậc 2 — HG-LinUCB thắng LinUCB.** Warm-start và phạt tải cắt bớt khám phá lãng phí. Nhưng theo P1, phạt tải đồng thời kéo HG-LinUCB về phía lựa chọn của LL. HG-LinUCB thắng LinUCB khi lợi ích "không bỏ qua worker rảnh" lớn hơn thiệt hại "bị kéo về worker xấu nhưng ít tải". Đây là câu hỏi thực nghiệm, và **sweep $\lambda \in \{0; 0{,}25; 0{,}5; 1{,}0\}$ chính là phép đo trực tiếp đánh đổi này**.

Cần nói rõ: HG-LinUCB **không có thêm thông tin nào** so với LinUCB thuần — LoadRatio vốn đã là chiều [7]. Ưu thế duy nhất của nó là phân bổ an toàn hơn trong giai đoạn đầu khi phần thưởng còn đến chậm. Warm-start cũng thừa hưởng điểm mù của LL theo một cách khác: trong 100 dispatch đầu, dữ liệu do LL sinh ra hầu như không chứa các lựa chọn "gửi vào worker nhanh đang bận" hay "gửi TU lớn vào worker cụ thể", nên mô hình khởi đầu thiếu đúng loại quan sát cần thiết để thoát khỏi lựa chọn của LL.

Vì vậy, một sweep $\lambda$ đơn lẻ không đủ để quy ưu thế (nếu có) cho từng cơ chế. Em đề xuất thêm **ablation 2 × 2** $(N, \lambda) \in \{(0, 0), (100, 0), (0, 0{,}5), (100, 0{,}5)\}$ ở các kịch bản chủ lực. Ô $(0, 0)$ chính là LinUCB thuần, ô $(100, 0{,}5)$ là HG-LinUCB.

**Hệ quả cần nói trước:** chuỗi thứ tự đầy đủ là khó nhất. Dự đoán của em là ở W4 và W5 có khả năng xảy ra LL < HG-LinUCB ≤ LinUCB ở $\lambda = 0{,}5$, và HG-LinUCB chỉ vượt LinUCB ở $\lambda$ nhỏ. Nếu kết quả như vậy, đó là một phát hiện có giá trị về giới hạn của cơ chế phạt tải, và sẽ được báo cáo như vậy.

---

## 5. Kịch bản đề xuất

Quy ước chung cho mọi kịch bản: 20 khối ngẫu nhiên; tối thiểu 4 scheduler (LL, SED, HG-LinUCB, HG-LinUCB-D) cộng LinUCB thuần để kiểm tra chuỗi thứ tự; mọi sự kiện xảy ra sau dispatch thứ 100; timestamp sự kiện ghi vào task log.

### S0 — Tải trung gian (Thí nghiệm 1)

- **Nhắm vào:** W1.
- **Ngoài đời:** cụm build dùng chung trong nhóm, số job đồng thời dao động theo giờ trong ngày, hiếm khi đúng 50% hay 100% năng lực.
- **Thiết lập:** CPython ở `-j6`…`-j9`; thêm hai build đồng thời từ hai client.
- **Dự đoán:** SED ≥ HG-LinUCB ≥ LL; LinUCB thuần thua LL. Kịch bản này **không nhằm** tạo chuỗi thứ tự, mà để vẽ ranh giới chế độ và kiểm tra liệu ưu thế (nếu có) của HG-LinUCB có lớn hơn SED hay không.
- **Metric chính:** makespan. **Phụ:** tỉ lệ bỏ qua worker rảnh; tỉ lệ chọn đúng worker rảnh có năng lực *theo tác vụ* cao nhất (lưu ý: worker 1,1 CPU với 3 slot đang đầy chỉ cho mỗi tác vụ khoảng 0,37 lõi, kém worker 1,0 CPU với 2 slot đang đầy là 0,5 lõi, nên "mạnh nhất" phải tính theo lõi hiệu dụng cho một tác vụ mới).

### S1 — Drift ẩn (Thí nghiệm 2)

- **Nhắm vào:** W3.
- **Ngoài đời:** máy bị thermal throttling, job nền (backup, indexing, antivirus), noisy neighbour trên VM đám mây.
- **Thiết lập:** tại dispatch thứ 120, `stress-ng --cpu 1` bên trong container worker 1,1 CPU; ba biến thể vĩnh viễn / 30 s rồi hồi phục / bật-tắt.
- **Mức chậm thực tế (ước tính theo CFS):** tiến trình stress chia quota 1,1 lõi với $k$ tiến trình compile, nên mỗi tác vụ còn khoảng $1{,}1/(k+1)$ lõi: 0,55 khi $k = 1$ (chậm 45%), 0,37 khi $k = 2$, 0,28 khi $k = 3$ (chậm 25%). Worker này rơi về mức worker yếu nhất chứ không trở thành máy tệ hẳn. Cần đo kiểm chứng trước; nếu hiệu ứng quá nhỏ sẽ báo lại thầy, không tự ý đổi sang `--cpu 2`.
- **Vấn đề thời lượng (phải giải quyết trước khi chạy):** build CPython ở `-j5` chỉ khoảng 40 s, và dispatch thứ 120/295 rơi vào khoảng giây thứ 15–17. Biến thể "chậm 30 s rồi hồi phục" vì thế kéo dài tới hết build, tức trùng với biến thể vĩnh viễn; còn biến thể vĩnh viễn chỉ để lại khoảng 175 quyết định cho việc thích nghi. Hai hướng xử lý (Q8): (a) dùng một **phiên build dài**, ví dụ build sạch CPython ba lần liên tiếp mà không khởi động lại coordinator (khoảng 885 dispatch), sự kiện vẫn đặt ở dispatch 120; (b) giữ một build nhưng rút ngắn pha chậm xuống khoảng 10 s. Em nghiêng về (a), vì nó giữ đúng 30 s như thầy giao.
- **Kiểm chứng drift có thật:** đọc `cpu.stat` (thời gian dùng CPU, số lần bị throttle) của container trong lúc chạy và ghi vào log, để chứng minh tốc độ hiệu dụng thực sự giảm — nhãn "drift" thôi chưa đủ.
- **Hai build đồng thời:** dispatch thứ 120 tính theo bộ đếm toàn cục của coordinator, vì hai build dùng chung một trạng thái học.
- **Baseline thích nghi không học:** phản biện chắc chắn sẽ hỏi "một EWMA thời gian compile theo worker có đủ không". HEFT đã cài có ý tưởng đó, nhưng **bản hiện tại không dùng được làm baseline drift**: nó lấy EWMA của `AvgCompileTime` trong registry, mà giá trị này là trung bình cộng dồn từ đầu phiên (`registry.go:357-361`), nên phản ứng với drift rất chậm. Cần sửa HEFT để cập nhật EWMA từ thời gian compile của từng tác vụ trước khi đưa vào S1 và S4.
- **Dự đoán:** ở tải thấp/trung bình, HG-LinUCB-D < LL về makespan ở biến thể vĩnh viễn; ở biến thể hồi phục, HG-LinUCB không discount sẽ chậm quay lại worker đã khoẻ. Ở `-j10`, LL tự sửa qua hàng đợi nên khác biệt có thể biến mất.
- **Dự đoán sai khi:** mức chậm quá nhỏ so với nhiễu; hoặc số quyết định còn lại sau sự kiện không đủ để bộ học thích nghi.
- **Thiết kế HG-LinUCB-D:** discount áp cho **mọi arm ở mỗi bước** (không chỉ arm được chọn), để arm bị né có $A$ co dần về $I$, bonus khám phá tăng lên và tự kích hoạt thử lại — điều kiện cần cho biến thể hồi phục. Hệ quả cần nói rõ: **đồng hồ discount là số dispatch toàn cục**, nên bộ nhớ hiệu dụng $\approx 1/(1-\gamma)$ được đo bằng dispatch của *cả cụm*, không phải số quan sát của một worker. Với $\gamma = 0{,}99$ và một worker nhận khoảng 25% dispatch, bộ nhớ hiệu dụng của arm đó chỉ còn khoảng 25 quan sát của chính nó; với hai client đồng thời, nhịp discount cũng thay đổi. Đồng hồ discount và cách quy đổi này sẽ được ghi trong `hypotheses.md` và báo cáo cùng kết quả. Bonus khám phá dùng $A^{-1}$ theo công thức thầy giao (biến thể đơn giản hoá của D-LinUCB, Russac và cộng sự 2019; bài gốc dùng thêm ma trận $\tilde V$ với $\gamma^2$); bài sẽ ghi rõ điều này.
- **Metric chính:** makespan. **Phụ:** tỉ lệ tác vụ gửi vào worker bị chậm trước/sau sự kiện; time-to-adapt; worker chứa tác vụ hoàn thành cuối cùng.
- **Định nghĩa time-to-adapt (đề xuất, sẽ cố định trong `hypotheses.md`):** tính từ thời điểm sự kiện tới khi tỉ lệ chọn worker bị chậm trong cửa sổ trượt 20 dispatch xuống dưới tỉ lệ theo năng lực *sau* drift, và giữ dưới ngưỡng đó trong 20 dispatch liên tiếp. Báo cáo theo ba đơn vị: số dispatch, số giây, và **số phần thưởng của worker bị chậm mà bộ học đã nhận** — đơn vị cuối là công bằng nhất, vì phần thưởng đến trễ nên đếm dispatch đơn thuần sẽ phạt oan bộ học. Không đạt trước khi phiên kết thúc thì ghi là "không thích nghi" (kiểm duyệt phải, censored).
- **Thiết kế D cho cả LinUCB thuần:** để quy ưu thế thích nghi cho đúng cơ chế, chạy thêm LinUCB-D (discount, không warm-start, không phạt tải) ở $\gamma$ tốt nhất.

### S2 — Workload C++ đuôi nặng (Thí nghiệm 3)

- **Nhắm vào:** W2, và làm chiều [9] có tín hiệu.
- **Ngoài đời:** build codebase C++ lớn (Chromium, LLVM, game engine), vài file template khổng lồ quyết định thời gian build.
- **Thiết lập:** workload tổng hợp có kiểm soát (Mục 6) rồi abseil-cpp → protobuf (có upb viết bằng C) → OpenCV core + imgproc, build qua CMake + ninja với hgbuild. Baseline thêm: "TU có kích thước > P90 thì gửi vào worker mạnh nhất còn slot", ngưỡng P90 tính từ một build mẫu và cố định trước. Thêm cả tổ hợp **SED + P90** (TU lớn → worker mạnh nhất còn slot, còn lại theo SED), vì đây là heuristic không học mạnh nhất mà một kỹ sư sẽ viết.
- **Dự đoán:** P90-heuristic và HG-LinUCB đều thắng LL về tail time; khác biệt giữa HG-LinUCB và P90-heuristic không chắc chắn.
- **Dự đoán sai khi:** các TU lớn được dispatch sớm (thứ tự của ninja), nên không thành straggler; chi phí compile không tương quan đủ với kích thước tiền xử lý (số byte là thước đo thô của độ phức tạp template); phần lớn quyết định chỉ có một ứng viên; hoặc phạt tải của HG-LinUCB đẩy TU lớn khỏi worker nhanh đang bận sang worker chậm đang rảnh. Thêm vào đó, theo P8, phần thưởng log làm bộ học ít nhạy với chính các TU dài nhất.
- **Metric chính:** makespan. **Phụ:** phân bố top 5% TU lớn nhất theo worker; thời gian từ dispatch cuối cùng đến khi build xong.
- **Tổ hợp S2 + S1** như thầy giao: P90-heuristic tiếp tục gửi TU lớn vào worker 1,1 CPU dù nó đã bị chậm.

### S3 — Worker mạnh CPU nhưng thiếu RAM (Thí nghiệm 4) — kịch bản chủ lực

- **Nhắm vào:** W4. Đây là kịch bản mà LL, SED **và** P90-heuristic đều chọn sai theo cấu trúc, và chỉ mô hình có độ dốc theo kích thước riêng cho từng worker mới phân biệt được "worker này tốt với TU nhỏ, tệ với TU lớn".
- **Ngoài đời:** cụm không đồng nhất trong đó máy CPU nhanh lại ít RAM (laptop/mini-PC đời mới cạnh server cũ nhiều RAM); CI runner bị giới hạn bộ nhớ.
- **Thiết lập:** worker 1,1 CPU đổi RAM thành 768 MB, bật swap bằng `memswap_limit`; workload S2. Lưu ý worker này có 3 slot, nên áp lực bộ nhớ phụ thuộc cả vào việc *bao nhiêu TU lớn chạy cùng lúc* — một tương tác mà bộ học thấy được qua [7].
- **Kiểm tra trước:** Docker Desktop trên WSL2 có thực sự cho container swap không (phụ thuộc cấu hình swap của VM trong `.wslconfig`); đo RSS đỉnh của các TU lớn để chắc chắn chúng vượt ngưỡng. Trong lúc chạy ghi lại lượng swap sử dụng và số sự kiện OOM của container.
- **Điều kiện tiên quyết — mô hình có thấy được tương tác này không:** chiều [7] đếm *số* tác vụ đang chạy chứ không đếm bộ nhớ của chúng, và kích thước nguồn chỉ dự đoán được RSS đỉnh nếu workload khiến quan hệ đó ổn định. OOM chỉ cho tín hiệu lỗi thô, không phải độ chậm tăng dần theo kích thước. Vì vậy trước khi coi S3 là phép thử "độ dốc theo từng worker", phải đo trên một build pilot quan hệ giữa kích thước đã ghi log, số TU nặng chạy đồng thời, RSS và thời gian compile. Nếu kích thước không dự đoán được độ chậm do swap, S3 sẽ không phân biệt được các scheduler và cần thiết kế lại workload (Mục 6).
- **Baseline thêm:** heuristic theo bộ nhớ (TU có RSS dự kiến lớn thì không gửi vào worker ít RAM) — tương đương P90 nhưng biết thông tin RAM; nếu HG-LinUCB không thắng được heuristic này thì giá trị của việc học ở S3 chỉ còn là "không cần biết trước RSS".
- **Dự đoán:** LL, SED, P90 < LinUCB. Theo P1, HG-LinUCB với $\lambda = 0{,}5$ có thể bị phạt tải kéo về worker 3 slot (LoadRatio thấp), nên HG-LinUCB ≤ LinUCB là khả năng thật; HG-LinUCB với $\lambda$ nhỏ dự đoán thắng.
- **Metric:** như S2, cộng số tác vụ OOM/fail và phân bố TU lớn trên worker thiếu RAM.

### S4 — Worker lỗi một phần / gray failure (đề xuất mới, chờ thầy duyệt)

- **Nhắm vào:** W5.
- **Ngoài đời:** gray failure — thành phần hỏng một phần nhưng vẫn báo khoẻ (Huang và cộng sự, HotOS 2017): worker có đĩa tạm gần đầy, phiên bản toolchain lệch với một số flag, card mạng chập chờn. Heartbeat vẫn bình thường nên registry vẫn coi là khoẻ mạnh.
- **Thiết lập:** sau dispatch thứ 120, một worker trả lỗi gRPC (`Unavailable`) cho khoảng 30% tác vụ (dưới ngưỡng 60% của circuit breaker), client chuyển sang compile cục bộ trên container builder 1 CPU. Theo P5, lỗi phải tiêm ở tầng RPC (cờ fault-injection trong worker, chỉ bật trong benchmark), không làm compiler fail.
- **Đường đi của một tác vụ lỗi (đã kiểm tra trong code):** lỗi `Unavailable` được client coi là lỗi tạm thời và **gửi lại tối đa 3 lần** (`build.go:400-430`, backoff 100 ms rồi 200 ms), mỗi lần là một lượt dispatch mới qua coordinator; chỉ khi hết số lần thử mới compile cục bộ. Với LL, lượt gửi lại nhiều khả năng lại rơi vào đúng worker lỗi (nó vừa trả lỗi nên đang rảnh), nên một tác vụ có thể tốn ba lần lỗi cộng backoff rồi mới fallback. Với bộ học, lượt gửi lại có thể đi sang worker khác. Chính sách retry phải được cố định cho thí nghiệm, và log phải đếm riêng số lần thử, số lượt gọi coordinator thất bại và số lần fallback cục bộ.
- **Vì sao lỗi của LL đắt:** ngoài các lần thử lại ở trên, fallback chạy trên builder 1 CPU vốn cũng đang chạy `make`. LL lại ưu tiên chính worker lỗi vì nó trả kết quả nhanh và luôn trông rảnh.
- **Baseline thêm:** HEFT đã sửa (xem S1; ngoài ra hiện `DecrementTasks` vẫn cập nhật `AvgCompileTime` khi tác vụ lỗi với thời gian 0 mà không tăng mẫu số, nên **tác vụ lỗi làm worker lỗi trông nhanh hơn** — lỗi này phải sửa trước S4, và cách xử lý tác vụ lỗi của baseline phải nêu rõ trong bài); và LL + circuit breaker được nạp dữ liệu thật (cần sửa P4) — để chứng minh lợi thế đến từ việc học lỗi *dưới ngưỡng*, không phải từ việc thiếu circuit breaker. Khi đã nối circuit breaker, tỉ lệ lỗi *trung bình* 30% chưa đảm bảo breaker luôn đóng: nó mở khi tỉ lệ *quan sát được* đạt 60% chỉ sau 3 yêu cầu trong cửa sổ 10 s, nên một chuỗi lỗi ngẫu nhiên dồn cục vẫn có thể làm nó mở. Cần dùng chuỗi lỗi xác định theo seed (giống nhau giữa các scheduler trong cùng khối) và ghi trạng thái breaker theo thời gian.
- **Dự đoán:** LL < LinUCB. HG-LinUCB: rủi ro P1 cao nhất ở kịch bản này, vì worker lỗi nhanh luôn có LoadRatio thấp.
- **Dự đoán sai khi:** chi phí một lần fallback nhỏ (builder rảnh), hoặc tỉ lệ lỗi quá thấp so với nhiễu.

### S5 — Cache affinity trên CI (Thí nghiệm 5)

- **Nhắm vào:** W7.
- **Ngoài đời:** CI build mỗi commit; giữa hai commit liên tiếp phần lớn TU không đổi.
- **Lưu ý triển khai:** cache hiện nằm phía client (`cache.NewStore` chỉ được gọi trong `hgbuild`); client cache hit thì tác vụ không tới coordinator. Kịch bản cần thêm cache phía worker, bảng `hash → worker` ở coordinator và đặc trưng `worker_has_artifact`. Đây là một **tính năng hệ thống mới**, không phải một cấu hình scheduler: nó thay đổi cả dòng tác vụ tới coordinator lẫn các phương pháp được so sánh, và cần kiểm soát riêng về tỉ lệ hit, dung lượng lưu trữ và tính đúng của artifact. Em đề xuất tách S5 thành phần mở rộng có phạm vi riêng, không nằm trong cam kết chính S0–S4, và làm sau cùng như thầy đã xếp.

### Hoãn lại: worker sau đường mạng chậm (W6)

Không đưa vào đợt này vì theo P2 và P3 bộ học hiện không thấy chi phí mạng. Chỉ có ý nghĩa nếu đổi phần thưởng sang thời gian đầu-cuối và sửa [8] (Mục 8, câu hỏi Q2, Q3).

**Bảng 3.** Tổng hợp dự đoán (giả thuyết, chưa phải kết quả).

| Kịch bản | Nhược điểm | Heuristic đối chứng | Dự đoán LL vs LinUCB | Dự đoán LinUCB vs HG-LinUCB ($\lambda = 0{,}5$) |
|---|---|---|---|---|
| S0 | W1 | SED | LL thắng | HG thắng |
| S1 | W3 | SED | LinUCB thắng ở tải thấp; hoà ở `-j10` | HG-D thắng ở biến thể hồi phục |
| S2 | W2 | P90 | không chắc | HG thắng |
| S3 | W4 | SED, P90 | LinUCB thắng | **rủi ro: HG ≤ LinUCB** |
| S4 | W5 | LL + CB | LinUCB thắng | **rủi ro cao: HG ≤ LinUCB** |
| S5 (mở rộng) | W7 | Affinity | chưa dự đoán | chưa dự đoán |

---

## 6. Xây dựng dữ liệu: workload tổng hợp có kiểm soát

Bên cạnh project thật, em đề xuất một bộ sinh workload C/C++ tổng hợp, để kiểm soát chính xác các đặc tính mà kịch bản cần:

- khoảng 300 TU; chi phí compile điều khiển bằng độ sâu instantiation template (hoặc số hàm sinh ra), lấy mẫu theo phân phối Pareto với tham số hình dạng chọn được (ví dụ 1,5 và 2,5 cho đuôi nặng / vừa);
- tỉ lệ C/C++ chọn được (ví dụ 30/70) để chiều [9] có tín hiệu;
- một nhóm TU "nặng bộ nhớ" với RSS đỉnh biết trước, cho S3;
- thứ tự TU trong build cố định theo seed, để các khối so sánh được với nhau.

Lợi ích: biết trước ground truth tác vụ nào nặng; thay đổi được độ nặng của đuôi một cách có hệ thống; tách được tác động của từng đặc tính. Project thật (abseil, protobuf, OpenCV) dùng để kiểm chứng tính ngoại suy, không phải để dò tìm kịch bản thắng.

---

## 7. Thống kê và tính toàn vẹn của thiết kế

- **Pre-registration:** mỗi thí nghiệm có `hypotheses.md` commit lên git trước khi chạy (timestamp commit là bằng chứng). Ghi một metric chính, danh sách so sánh trong họ Holm, kiểm định một/hai phía, tiêu chí loại khối.
- **Chuỗi thứ tự** LL < LinUCB < HG-LinUCB được kiểm định như hai giả thuyết có thứ tự (fixed-sequence): chỉ kiểm định bậc 2 khi bậc 1 có ý nghĩa — cách này giữ được tỉ lệ lỗi tổng thể mà không phải chia α.
- **Effect size:** matched-pairs rank-biserial $r = (T^+ - T^-)/(T^+ + T^-)$ thay cho Cliff's d (Cliff's d dành cho mẫu độc lập, không khớp với Wilcoxon ghép cặp đang dùng).
- **Báo cáo đầy đủ:** giữ nguyên kết quả `-j5` hoà LL trong bài. Mục tiêu là một **bản đồ chế độ** — khi nào heuristic là đủ, khi nào việc học có lợi và nhờ cơ chế nào — chứ không phải chứng minh HG-LinUCB luôn thắng.
- **Đơn vị thống kê là build (hoặc phiên build), không phải tác vụ.** 295 tác vụ trong một build tương quan với nhau, không phải 295 lần lặp độc lập. Các metric cấp tác vụ chỉ dùng để giải thích cơ chế, không dùng để kiểm định.
- **Khác biệt tối thiểu có ý nghĩa thực tế** được ghi trước (ví dụ 2% makespan), báo cáo kèm khoảng tin cậy của trung vị hiệu số ghép cặp, để phân biệt "có ý nghĩa thống kê" với "đáng kể trong thực tế".
- **Giữ họ so sánh nhỏ:** mỗi kịch bản chỉ một tương phản chính; các tổ hợp $\gamma$, $\lambda$ và ablation 2 × 2 là phân tích phụ, báo cáo đầy đủ nhưng không dùng làm khẳng định chính.
- **Không điều chỉnh phương pháp sau khi thấy kết quả.** Nếu thầy đồng ý sửa phương pháp (Q4), việc sửa phải xong và được ghi vào `hypotheses.md` trước khi chạy.

**Ước tính chi phí chạy** (khoảng 1,5 phút/build gồm khởi động lại và xoá cache): S0 khoảng 400 build (~10 giờ); S1 khoảng 540 build (~13–14 giờ) nếu sweep $\lambda$ không chạy chéo với $\gamma$. Workload C++ thật sẽ lâu hơn đáng kể.

---

## 8. Các quyết định cần thầy chốt

| # | Câu hỏi | Đề xuất của em |
|---|---|---|
| Q1 | Thêm kịch bản S4 (gray failure) và baseline LL + circuit breaker thật? | Có — đây là kịch bản LL thua rõ nhất về mặt cơ chế |
| Q2 | Đổi phần thưởng sang thời gian đầu-cuối (gồm chờ và truyền)? | Giữ phần thưởng hiện tại cho đợt này; ghi là hạn chế. Đổi phần thưởng là đổi phương pháp |
| Q3 | Xử lý chiều [8] hằng số (P3): nối dữ liệu thật hay bỏ chiều (d = 11)? | Nối dữ liệu thật trước khi chạy đợt mới, ghi rõ trong bài là phiên bản sửa |
| Q4 | Có được điều chỉnh cơ chế phạt tải trước khi chạy (ví dụ $\lambda$ giảm dần theo độ tự tin của mô hình) không? | Không sửa; dùng sweep $\lambda$ để đo đánh đổi, báo cáo trung thực. Ý tưởng sửa để ở hướng phát triển |
| Q5 | Sweep $\lambda$ chạy với HG-LinUCB thường hay chạy chéo cả ba giá trị $\gamma$? Và chọn $\gamma$ cho khẳng định chính bằng cách nào? | Sweep $\lambda$ với HG-LinUCB thường. $\gamma$ cho khẳng định chính được **chọn trên các khối pilot riêng**, cố định trong `hypotheses.md`, rồi mới chạy 20 khối xác nhận; ba giá trị $\gamma$ vẫn chạy đủ nhưng chỉ là phân tích khám phá. Chọn $\gamma$ "tốt nhất" trên chính dữ liệu đánh giá là khớp theo nhiễu |
| Q6 | Dùng thêm workload tổng hợp có kiểm soát (Mục 6)? | Có |
| Q7 | P6: kiểm chứng lại kết luận "bộ học thấy năng lực" ở Mục 7.6? | Có, chạy cùng đợt S0: ablation đặt [4] = hằng số chung, **kèm** đường cong học theo từng worker và phân bổ dispatch ngay sau warm-start, để tách "chấm điểm ban đầu nhờ đặc trưng năng lực" khỏi "tốc độ học được qua phần thưởng". Riêng ablation không đủ để phân biệt hai cách giải thích |
| Q8 | Thời lượng S1: phiên build dài (CPython × 3, không khởi động lại coordinator) hay rút ngắn pha chậm? | Phiên build dài — với điều kiện xoá cache phía client giữa các build và kiểm tra số dispatch thực tế ≈ 885 (nếu không, cache client sẽ chặn phần lớn tác vụ của build thứ hai và ba trước khi tới coordinator) |
| Q9 | Thêm HEFT (EWMA, sau khi sửa để học từ từng tác vụ) làm baseline thích nghi không học cho S1, S4, và ablation 2 × 2 $(N, \lambda)$ cho các kịch bản chủ lực? | Có |

## 9. Lộ trình

1. Code: SED và SED + P90; HG-LinUCB-D (và LinUCB-D); sửa P3 (nếu Q3 = có); sửa P10 và cho HEFT học từ từng tác vụ; script chạy phiên build dài có xoá cache client giữa các build; ghi ảnh chụp ứng viên tại thời điểm quyết định (P9); script tiêm sự kiện, ghi event và `cpu.stat` vào log; metric mới trong `scripts/analyze_rigorous.py` (rank-biserial, bỏ qua worker rảnh theo lõi hiệu dụng, time-to-adapt ba đơn vị, đếm fallback và hết hạn hàng đợi).
2. Chạy lại baseline `-j5` trên revision hiện tại (P7) để có mốc so sánh cùng mã nguồn.
3. Đo kiểm chứng mức chậm của `stress-ng` (chưa tính vào kết quả).
4. Viết `hypotheses.md` cho S0 và S1, commit.
5. Chạy S0, S1 → gửi kết quả cho thầy.
6. Sau khi thầy duyệt: S2 → S3 → S4; S5 là phần mở rộng làm sau cùng nếu còn thời gian.
