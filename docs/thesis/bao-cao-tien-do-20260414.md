# Báo cáo tiến độ — Đồ án tốt nghiệp Hybrid-Grid Build

**Ngày báo cáo:** 14/04/2026
**Giai đoạn:** Sau định hướng nghiên cứu của GVHD (tập trung vào thuật toán lập lịch)

---

## 1. Tóm tắt

Kể từ buổi làm việc lần trước, nhóm đã hoàn thành ba việc chính: (1) chốt định hướng nghiên cứu tập trung vào thuật toán lập lịch và chuyển sang cách tiếp cận Reinforcement Learning, (2) làm sạch trạng thái hệ thống (đóng phiên bản v0.5.0, làm mới bộ số liệu benchmark baseline), và (3) chuẩn bị nền tảng cho phase nghiên cứu sâu — bao gồm tổng quan tài liệu sơ bộ và kế hoạch nghiên cứu 4 giai đoạn. Đồ án hiện đang chuyển từ giai đoạn phát triển sản phẩm (product engineering) sang giai đoạn nghiên cứu khoa học (research) với trọng tâm là xây dựng đóng góp mới có giá trị học thuật.

---

## 2. Các nội dung đã hoàn thành

### 2.1. Chốt định hướng nghiên cứu

Sau khi gửi thầy/cô báo cáo phân tích bài toán lập lịch (file `docs/thesis/bai-toan-lap-lich.md`), nhóm nhận được định hướng:

- Tạm dừng mở rộng các nền tảng build mới (Rust, Go, Node.js, Cocos), tập trung nguồn lực vào tối ưu thuật toán lập lịch.
- Phát triển thuật toán lập lịch thành một mô hình AI chuyên dụng — cụ thể là Reinforcement Learning model.
- Hiện đã có đủ các nền tảng build để minh họa hệ thống (C/C++, Flutter Android, Unity — ba loại đại diện cho compilation, mobile build, và game build).

Định hướng này thay đổi bản chất đồ án từ "hệ thống build phân tán đa nền tảng" sang "hệ thống build phân tán với scheduler học sâu trên cluster không đồng nhất".

### 2.2. Hoàn thiện trạng thái hệ thống hiện tại

Trước khi bước vào giai đoạn nghiên cứu, nhóm đã làm sạch trạng thái kỹ thuật:

- **Đóng phiên bản v0.5.0 (Unity)**: toàn bộ mã nguồn hỗ trợ build Unity phân tán đã được commit và đẩy lên repository (commit `392a3ea`). Các thành phần bao gồm executor chạy Unity batch mode, phát hiện năng lực (capability detection), routing gRPC, giao diện CLI, và dashboard stats.
- **Refresh benchmark baseline**: chạy lại toàn bộ ba kịch bản benchmark chuẩn (scaling, equal resources, heterogeneous) với scheduler Power of Two Choices (P2C) đang sử dụng. Kết quả lưu tại `docs/BENCHMARK_REPORT_v0.5.md`. Bảng dưới đây tóm tắt:

| Kịch bản | Số worker | Tăng tốc (speedup) |
|---|---|---|
| Scaling (thêm tài nguyên) | 5 | 5.01× |
| Equal resources (cố định tổng tài nguyên) | 3 (tốt nhất) | 1.14× |
| Heterogeneous (tài nguyên chia không đều) | 5 | **1.23×** |

Con số 1.23× trên cluster heterogeneous là mức chuẩn (baseline) mà scheduler mới cần vượt qua để chứng minh tính hiệu quả của phương pháp RL.

### 2.3. Khảo sát sơ bộ tài liệu liên quan

Nhóm đã tổng hợp bước đầu 9 công trình khoa học làm nền tảng cho nghiên cứu, lưu tại `docs/thesis/annotated-bibliography.md`. Các bài báo trải rộng trên ba trục:

- **Lý thuyết nền tảng lập lịch**: Mitzenmacher (2001) — Power of Two Choices; Lenstra, Shmoys, Tardos (1990) — lập lịch trên unrelated machines.
- **Lập lịch phân tán thực nghiệm**: Ousterhout et al. (2013) — Sparrow; Grandl et al. (2014) — Tetris; Ghodsi et al. (2011) — Dominant Resource Fairness.
- **Cân nhắc vấn đề scheduling + learning**: Delimitrou & Kozyrakis (2014) — Paragon/Quasar; Cortez et al. (2017) — Resource Central; Ananthanarayanan et al. (2010) — Mantri; Dean & Barroso (2013) — Tail at Scale.

Qua khảo sát sơ bộ, nhóm nhận diện được khoảng trống (research gap) quan trọng: **hầu như chưa có công trình nào công bố về việc áp dụng Reinforcement Learning cho bài toán lập lịch trong hệ thống build phân tán, đặc biệt trên cluster heterogeneous và không cần simulator để huấn luyện**. Các công trình RL scheduling hiện có (Decima, DeepRM) đều yêu cầu pre-training trên simulator — điều không khả thi trong hệ thống build vì thời gian biên dịch phụ thuộc vào phần cứng vật lý.

### 2.4. Bổ sung nhân sự nhóm

Đồ án tiếp nhận thêm một thành viên. Để đảm bảo thành viên mới nắm được toàn bộ bối cảnh dự án, nhóm đã xây dựng tài liệu hướng dẫn toàn diện tại `docs/onboarding.md` (606 dòng) bao gồm: tổng quan hệ thống, kiến trúc 3 thành phần (hgbuild, hg-coord, hg-worker), stack công nghệ, cấu trúc mã nguồn, trạng thái hiện tại, bài toán nghiên cứu, hướng tiếp cận RL, roadmap, hướng dẫn thiết lập môi trường, và checklist 11 mục để thành viên mới tự kiểm tra tiến độ học tập.

---

## 3. Định hướng kế tiếp

### 3.1. Nguyên tắc: nghiên cứu trước, thực hiện sau

Nhóm nhận thức rằng để đồ án thực sự có giá trị khoa học (không chỉ dừng ở mức "áp dụng RL vào scheduling"), cần hiểu sâu trước khi cam kết một phương pháp cụ thể. Do đó, thay vì lập kế hoạch cài đặt ngay, nhóm xây dựng **kế hoạch nghiên cứu 4 giai đoạn** kéo dài 6-8 tuần (file `docs/thesis/research-plan.md`):

**Giai đoạn 1 (Tuần 1-3) — Nền tảng lý thuyết:** nắm vững ba trụ cột gồm lý thuyết lập lịch (scheduling theory), Reinforcement Learning (Sutton & Barto — TD learning, function approximation), và Multi-Armed Bandits. Kèm thực hành: cài đặt Q-learning cho bài toán đơn giản (GridWorld, CartPole) để hiểu cơ chế.

**Giai đoạn 2 (Tuần 3-5) — Khảo sát SOTA:** đọc sâu 14 công trình cốt lõi, xây dựng bảng so sánh có hệ thống theo các trục: loại RL, online/offline, heterogeneous, cache-aware, quy mô thử nghiệm. Viết phân tích 500 từ về khoảng trống nghiên cứu mà Hybrid-Grid hướng tới.

**Giai đoạn 3 (Tuần 5-7) — Thí nghiệm khám phá:** thu thập dữ liệu build thật (CPython, Redis) để hiểu đặc điểm workload thực tế — phân phối thời gian biên dịch, mối tương quan giữa đặc trưng file và thời gian compile, mức độ chênh lệch giữa các worker. Cài đặt nguyên mẫu RL đơn giản bằng Python để kiểm nghiệm khái niệm (proof of concept) trước khi triển khai trong Go.

**Giai đoạn 4 (Tuần 7-8) — Tổng hợp và quyết định:** dựa trên kết quả các giai đoạn trước, viết tài liệu tổng hợp ~5 trang trả lời rõ ràng: bài toán formal (MDP hay bandit), phương pháp phù hợp nhất, feature nào dùng, baseline nào so sánh, kế hoạch đánh giá cụ thể, các câu hỏi nghiên cứu chính. Họp nhóm + GVHD để chốt hướng trước khi chuyển sang giai đoạn cài đặt.

### 3.2. Tại sao đi theo trình tự này

Nghiên cứu sơ bộ cho thấy đa phần các công trình RL scheduling yếu bị reviewer bác bỏ vì ba lỗi lặp lại: (1) so sánh với baseline quá yếu (chỉ Random, FIFO), (2) đánh giá trên một workload duy nhất, (3) không có phân tích ablation. Để tránh các lỗi này, nhóm cần:

- Hiểu đủ sâu để biết baseline nào là chuẩn mực (ví dụ thuật toán HEFT — Heterogeneous Earliest Finish Time — là baseline bắt buộc cho lập lịch heterogeneous).
- Kiểm tra trên dữ liệu thực tế trước khi chọn phương pháp: nếu các đặc trưng không dự đoán được thời gian biên dịch (R² thấp), RL sẽ không hoạt động dù cài đặt hoàn hảo.
- Xác định trước các câu hỏi nghiên cứu chính (research questions) để bảo đảm đóng góp có cấu trúc.

### 3.3. Các câu hỏi nghiên cứu dự kiến (cần xin ý kiến GVHD)

Nhóm đề xuất các câu hỏi nghiên cứu sau để định hướng toàn bộ phần đánh giá:

1. **RQ1 (Feature Importance):** Đặc trưng nào dự đoán tốt nhất thời gian biên dịch trên cluster không đồng nhất?
2. **RQ2 (Sample Efficiency):** Mô hình RL cần quan sát bao nhiêu task để vượt qua heuristic cố định?
3. **RQ3 (Heterogeneity Effect):** Mức cải thiện của RL so với heuristic phụ thuộc vào độ heterogeneous của cluster ra sao?
4. **RQ4 (Non-stationarity Adaptation):** Khi workload thay đổi (ví dụ chuyển project), mô hình thích ứng nhanh ra sao?
5. **RQ5 (Failure Mode):** Trong trường hợp nào RL thua heuristic, và tại sao?

RQ5 đặc biệt quan trọng: một báo cáo chỉ nêu các trường hợp RL thắng sẽ không đủ tin cậy. Báo cáo trung thực cả trường hợp thất bại kèm phân tích lý do mới là đóng góp khoa học đúng nghĩa.

### 3.4. Phân công nhân sự

Hai thành viên làm việc song song trong giai đoạn 1-2 (đọc tài liệu), cùng cộng tác trong giai đoạn 3 (thí nghiệm). Chi tiết phân công theo tuần lưu tại mục "Phân công 2 thành viên" của file `research-plan.md`.

---

## 4. Các vấn đề xin ý kiến thầy/cô

Nhóm xin được thầy/cô cho ý kiến về các nội dung sau trước khi bắt đầu giai đoạn nghiên cứu sâu:

1. **Các câu hỏi nghiên cứu RQ1-RQ5 ở mục 3.3 có phù hợp với kỳ vọng đóng góp mới của đồ án không?** Nếu cần điều chỉnh hoặc bổ sung, nhóm xin ý kiến cụ thể.

2. **Về mức độ lý thuyết trong luận văn:** liệu đồ án cần chứng minh tính hội tụ hoặc phân tích competitive ratio của thuật toán RL đề xuất, hay tập trung chủ yếu vào đánh giá thực nghiệm là đủ?

3. **Về baseline so sánh:** ngoài các scheduler nội bộ (Simple, LeastLoaded, P2C, HEFT), có cần so sánh với hệ thống công nghiệp thực tế như distcc, icecream, Bazel Remote Execution hay không? Nếu có, việc cài đặt các hệ thống này để benchmark đối sánh sẽ tốn thêm khoảng 2-3 tuần.

4. **Về benchmark corpus:** nhóm đề xuất sử dụng CPython, Redis, và một tập con của LLVM làm workload chính. Thầy/cô có gợi ý bổ sung hoặc thay thế nào không?

5. **Về khả năng công bố:** nhóm có thể hướng tới việc chỉnh sửa kết quả thành một bài báo gửi hội nghị chuyên ngành (IEEE CCGrid, ACM CloudCom, ICPP workshop) sau khi bảo vệ không? Nếu có, việc chuẩn bị dữ liệu và phân tích cần theo chuẩn cao hơn ngay từ đầu.

---

## 5. Timeline tổng thể dự kiến

```
Tuần 1-8    (Tháng 4-6/2026)   : Nghiên cứu sâu (4 giai đoạn trên)
Tuần 9-14   (Tháng 6-7/2026)   : Cài đặt thuật toán RL được chọn
Tuần 15-18  (Tháng 7-8/2026)   : Đánh giá thực nghiệm, phân tích ablation
Tuần 19-26  (Tháng 8-10/2026)  : Viết luận văn (khoảng 50-70 trang)
Tuần 27-30  (Tháng 10-11/2026) : Chỉnh sửa, hoàn thiện figures, bảo vệ thử
Dự kiến bảo vệ chính thức      : Tháng 12/2026 - Tháng 1/2027
```

Thời lượng còn lại đến dự kiến bảo vệ là khoảng 9 tháng, đủ thoải mái để làm nghiên cứu tử tế mà không phải ép tiến độ.

---

## 6. Tài liệu đính kèm (trong repository)

- `docs/thesis/bai-toan-lap-lich.md` — Báo cáo phân tích bài toán lập lịch (đã gửi lần trước)
- `docs/thesis/annotated-bibliography.md` — Tổng quan 9 công trình nền tảng
- `docs/thesis/research-plan.md` — Kế hoạch nghiên cứu chi tiết 4 giai đoạn
- `docs/onboarding.md` — Tài liệu hướng dẫn cho thành viên mới
- `docs/BENCHMARK_REPORT_v0.5.md` — Kết quả benchmark baseline P2C (tháng 4/2026)
