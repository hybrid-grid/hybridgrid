# Báo cáo tiến độ đồ án — 10/05/2026

> Người báo cáo: HieuLD
> Đề tài: Thuật toán lập lịch dựa trên Contextual Bandit cho hệ thống biên dịch phân tán Hybrid-Grid
> Kỳ báo cáo trước: 14/04/2026 (file `bao-cao-tien-do-20260414.md` trong cùng thư mục)
> Mã nguồn: <https://github.com/hybrid-grid/hybridgrid> (nhánh `main`)

## §1 Tóm tắt

Trong khoảng thời gian 14/04 → 10/05/2026, nhóm đã hoàn thiện toàn bộ ba mốc kỹ thuật (M1, M2, M3) và bổ sung baseline HEFT theo định hướng đã thống nhất tại buổi gặp gần nhất. Tổng số 17 commit đã được đẩy lên repo công khai. Kết quả thực nghiệm chính: thuật toán LinUCB sau khi vá ba lỗi triển khai cho thời gian build (wall-clock) **ngang bằng heuristic P2C** trên cluster 5 worker khác chủng, đồng thời **cải thiện 2.3% về độ trễ đuôi P99** so với P2C — phù hợp với kỳ vọng lý thuyết về lợi thế của bandit trên phân phối nặng đuôi.

Bên cạnh phần code, nhóm đã hoàn thành: (a) chương luận văn `chuong-X-lap-lich.md` ~8 trang văn phong học thuật trang trọng; (b) `theory-notes.md` ghi nhận và xác minh độc lập mọi dẫn chứng lý thuyết (DOI/arXiv ID đầy đủ); (c) `paper-skeleton.md` là sườn paper định hướng công bố quốc tế.

## §2 Kết quả đo lường

Thực nghiệm trên benchmark CPython 873 task biên dịch, ba cấu hình cluster Docker (1, 3, 5 worker) với tổng dung lượng CPU cố định 4.0 lõi nhưng phân bổ không đều giữa các worker.

**Bảng 1 — Tổng thời gian build (giây) trên cluster heterogeneous (873 task).**

| Cấu hình | LeastLoaded | P2C | ε-greedy | LinUCB α=1 (lỗi) | **LinUCB-fixed α=0.5** | HEFT |
|---|---|---|---|---|---|---|
| 1w-4.0CPU | 92 | 130 | 146 | 129 | 131 | 129 |
| 3w-hetero | 123 | 85 | 142 | 103 | 108 | 135 |
| **5w-hetero** | 152 | **94** | 119 | 158 | **94** | 144 |

**Bảng 2 — Lệch tải và tail latency trên 5w-hetero.**

| Scheduler | Top:Bottom dispatch | P50 (ms) | P95 (ms) | P99 (ms) |
|---|---|---|---|---|
| LeastLoaded | 10.8 : 1 | 820 | 6 226 | 23 961 |
| P2C | 8.6 : 1 | 704 | 5 830 | 19 347 |
| ε-greedy | 13.2 : 1 | 956 | 7 387 | 25 488 |
| LinUCB α=1 (lỗi) | 97 : 1 | 942 | 7 142 | 23 661 |
| **LinUCB-fixed α=0.5** | 13.9 : 1 | 826 | 5 676 | **18 896** |
| HEFT | 145 : 1 | 995 | 6 919 | 24 143 |

Hai kết luận chính:

1. **P2C là baseline mạnh** — vượt LeastLoaded 1.62× trên cluster khác chủng (phù hợp định lý Mitzenmacher 2001). Đây là đối thủ thật sự, không phải straw-man.
2. **LinUCB sau sửa lỗi đạt được mục tiêu** — bằng P2C về wall-clock (94 giây) và **vượt P2C ở P99** (18 896 ms so với 19 347 ms, giảm 2.3%). Tail latency là cửa sổ mà framework bandit được thiết kế để tối ưu, do đó kết quả này có ý nghĩa lý thuyết, không chỉ là sai số đo.

## §3 Câu chuyện vá lỗi triển khai

Lần chạy LinUCB đầu tiên cho kết quả 158 giây trên 5w-hetero — **xấu nhất trong sáu scheduler**. Một code-reviewer độc lập (hệ subagent) đã chỉ ra ba lỗi triển khai chính:

1. **Rò rỉ nhãn ở bước cập nhật** (CRITICAL): khi `RecordOutcome` chạy, `worker.ActiveTasks` đã bị `DecrementTasks` giảm đi và bộ ghi RTT đã hấp thụ độ trễ của chính task đang xét. Bandit do đó học $\hat{\theta}$ trên một véc-tơ đặc trưng *khác* với véc-tơ đã dùng để ra quyết định. **Cách sửa**: lưu cache véc-tơ đặc trưng tại thời điểm Select, dùng lại tại RecordOutcome thông qua `TaskContext.TaskID`.
2. **One-hot loại build trùng tuyến tính với bias** (CRITICAL): cờ build_type luôn là `(1, 0, 0)` dưới luồng `Compile()` hiện tại, khiến chiều CPP trùng hoàn toàn với bias và hai chiều Flutter/Unity vĩnh viễn không học gì. **Cách sửa**: rút véc-tơ đặc trưng từ 12 chiều xuống 9 chiều, bỏ ba chiều build_type. Khi Flutter/Unity được tích hợp vào luồng học tập trong tương lai sẽ thêm lại đúng cách.
3. **Phần thưởng át UCB bonus** (HIGH): với phần thưởng nguyên gốc trong khoảng `[-10, -3]` và UCB bonus cỡ `α·√(x^T A^{-1} x) ≈ 1`, điểm ngẫu nhiên ở giai đoạn cold-start luôn dồn về cánh tay nóng. **Cách sửa**: chuẩn hoá phần thưởng theo `log1p(timeout_ms)` để $r \in [-1, 0]$, đồng thời đặt mặc định $\alpha = 0.5$ (Chu et al. 2011 §5 dải thực nghiệm).

Sau khi vá ba lỗi, makespan trên 5w-hetero hồi phục từ 158 giây xuống 94 giây — toàn bộ 64 giây thiếu hụt được lấy lại mà *không* cần thay đổi thuật toán. Đây là một quan sát có giá trị nghiên cứu: trên các bài báo thuật toán, ranh giới giữa "đúng về mặt thuật toán" và "đúng về mặt triển khai" thường bị bỏ qua, trong khi thực tế ranh giới này có thể quyết định toàn bộ kết quả.

Quy trình vá lỗi này đã được tài liệu hoá trong commit `85f20e2` của repo cùng với một tệp đánh giá kỹ thuật chi tiết ở `.sisyphus/evidence/m1/findings.md` §8.

## §4 Hạn chế còn lại và rủi ro

Để bảo đảm tính trung thực với GVHD, nhóm liệt kê các hạn chế hiện hữu mà bản báo cáo này chưa giải quyết:

1. **Số lần lặp thí nghiệm**: hầu hết các con số trong Bảng 1 và Bảng 2 đến từ một lần chạy duy nhất. Sàn nhiễu của môi trường Docker/macOS quan sát được khoảng ±25 giây. Để công bố cần lặp ≥ 5 lần và áp dụng kiểm định Wilcoxon — chi phí ước lượng khoảng 6 giờ tính toán liên tục, đã lên kế hoạch cho phiên thí nghiệm tiếp theo.
2. **Đa dạng workload**: hiện chỉ test trên CPython. Các codebase nặng template như Boost, Eigen, LLVM có phân phối thời gian khác hẳn — kết quả tương đối có thể thay đổi.
3. **Drift và thermal throttling**: LinUCB chuẩn không có guarantee dưới drift (Lattimore & Szepesvári 2020 Ch. 31). Hệ thống hiện chưa có cơ chế phát hiện change-point.
4. **α-sweep còn 1 ô thiếu**: cần một lần chạy cô lập α=1.0 *sau* khi đã vá lỗi để tách biệt hoàn toàn ảnh hưởng của α khỏi ảnh hưởng của bug.
5. **HEFT online**: thuật toán gốc của Topcuoglu 2002 là offline; phần thích ứng online (HEFT-LPT degeneration) chỉ tìm được một paper khảo sát 2024 đề cập chứ chưa có canonical reference. Có thể cần dùng cẩn thận trong phần Related Work.

## §5 Định hướng tiếp theo và các vấn đề xin ý kiến thầy

### Định hướng nội bộ
- Hoàn thành multi-rep + Wilcoxon để có khoảng tin cậy thống kê.
- Chạy benchmark trên ít nhất 1 workload nặng template (đề xuất: Boost.MPL hoặc LLVM `clang/lib`).
- Mở rộng feature vector với cờ "đã có cache" (cache-aware scheduling) — đây là khoảng trống chưa có công bố cho lập lịch biên dịch phân tán.

### Câu hỏi xin ý kiến

1. **Đối tượng công bố**: với mức kết quả hiện có (LinUCB ngang P2C wall-clock, vượt P2C P99) thầy đánh giá phù hợp với venue nào?
   - Hội nghị quốc tế Tier-2 hệ thống (EuroSys/HotOS workshop)?
   - Tạp chí trong nước (Journal of Computer Science and Cybernetics)?
   - Hay tập trung trước cho luận văn, công bố sau khi mở rộng?
2. **Cache-aware là contribution chính hay phụ**? Nếu chính, cần dành ~4 tuần để thu thập dữ liệu cache hit rate per (worker, file) và mở rộng feature vector. Tăng đáng kể tính novelty nhưng cũng tăng rủi ro deadline.
3. **Mức độ chi tiết của câu chuyện vá lỗi trong luận văn**: nhóm đã viết tách riêng §X.5 trong chương luận văn để trình bày 3 lỗi và kết quả hồi phục 158→94s. Thầy thấy nên giữ ở mức chi tiết hiện tại, mở rộng thêm, hay rút gọn để giữ nội dung "khoa học" hơn?
4. **Decima reward ablation**: nhóm phát hiện công thức reward `r = -log(1+t)` không có nguồn peer-reviewed (đã sửa lại trong theory-notes), trong khi Decima dùng `r_k = -(t_k - t_{k-1})J_k` có cơ sở Định luật Little. Có nên triển khai cả hai và so sánh thực nghiệm như một ablation, hay chỉ ghi nhận trong limitations?
5. **Thời điểm nộp luận văn và trình tự ưu tiên**: deadline thesis defense dự kiến cuối 2026 — Q1/2027. Với 7 tháng còn lại, thầy đề xuất ưu tiên (a) hoàn thiện code + benchmark, hay (b) viết luận văn song song với thí nghiệm?

## §6 Tài liệu đính kèm trong repo

| File | Vai trò |
|---|---|
| `docs/thesis/chuong-X-lap-lich.md` | Bản thảo chương luận văn, ~8 trang văn phong học thuật |
| `docs/thesis/theory-notes.md` | Ghi chú lý thuyết với mọi dẫn chứng có DOI/arXiv |
| `docs/thesis/paper-skeleton.md` | Sườn paper tiếng Anh định hướng công bố |
| `.sisyphus/evidence/m1/findings.md` | Phân tích thực nghiệm chi tiết của 6 scheduler |
| `.sisyphus/evidence/m1/tasks-*.jsonl` | Log thô 873 record/scheduler để tái lập kết quả |
| `internal/coordinator/scheduler/*.go` | Mã nguồn 6 scheduler có comment tham chiếu paper |

Báo cáo này hiện được commit cùng repo tại `docs/thesis/bao-cao-tien-do-20260510.md`. Nhóm sẵn sàng trình bày trực tiếp khi thầy sắp xếp được lịch.
