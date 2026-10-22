# Giả thuyết đăng ký trước: Thí nghiệm 3 (workload C++ đuôi nặng)

> **BẢN NHÁP, chưa đóng băng.** Chỉ đóng băng sau khi: (1) pilot xong và điền các giá trị đánh dấu ⟨pilot⟩;
> (2) Codex phản biện; (3) người dùng xác nhận các quyết định ở mục 6. Sau khi đóng băng bằng commit,
> file này không được sửa; mọi thay đổi về sau chỉ được ghi thành kịch bản bổ sung ở file khác, kèm lý do
> (quy định chung số 3 của thầy). Kế hoạch: `docs/thesis/ke-hoach-thi-nghiem-3.md`.
> Pilot (thư mục `D:\bench-work\out\pilot-tn3\`) chỉ dùng để kiểm tra công cụ và chọn tham số; số liệu
> pilot không được gộp vào kết quả.

## 0. Cấu hình chung (cố định)

- **Mã nguồn:** nhánh `feat/exp3`, commit chứa file này; commit được ghi trong `meta.json` của mỗi lượt
  chạy. Trước mỗi lần bắt đầu hoặc chạy tiếp một thư mục kịch bản: chạy `preflight.sh` với `STRICT=1`
  và `EXPECT_COMMIT` bằng commit đó; buộc tắt các ứng dụng không cần thiết.
- **Cụm:** 5 worker 0,5/0,6/0,8/1,0/1,1 CPU, max-parallel 1/2/2/2/3, coordinator 0,25 CPU (như TN1–2);
  builder 2,0 CPU, 2 GiB RAM (`BUILDER_CPUS=2.0`, `BUILDER_MEM=2G`); `JOBS=8`, 1 client.
- **Build:** CMake + Ninja qua `hgbuild ninja --no-fallback`. Không task nào được biên dịch tại chỗ;
  configure nằm ngoài cửa sổ đo và không gửi task tới coordinator đang đo.
- **Timeout mỗi task của coordinator:** 600 s (`--request-timeout=600s`; mặc định của hệ thống là
  120 s; đây là hạn chót của lời gọi tới worker, còn thời gian chờ trong hàng đợi dispatch vẫn giới hạn
  30 s như cũ). Lý do, từ pilot:
  - P90 và SED + P90 dồn các file nặng vào worker 1,1 CPU. Không có drift, task dài nhất đã tới 113 s;
    có drift, sau khi sửa, task dài nhất là 193 s.
  - Với giới hạn 120 s, các task này bị cắt. Trước khi sửa, SED + P90 có drift hỏng 2/2 lần; sau khi
    sửa, 0 task hỏng.
  - Task bị cắt làm hỏng cell của một số scheduler nhiều hơn các scheduler khác, làm lệch phép so sánh.
    Timeout không phải đối tượng của TN3.
- **Hạn chót phía client:** wrapper `hgcc`/`hgcxx` dùng `HG_TIMEOUT` = timeout coordinator + 60 s
  (660 s). Pilot có drift cho thấy với hạn chót mặc định 2 phút, client huỷ các task dài rồi gửi lại
  chính task đó trong khi worker vẫn đang chạy bản cũ; P90 chọn lại đúng máy đó, gặp "đầy" và build hỏng
  (SED + P90 hỏng 2/2 lần). TN1–2 không bị ảnh hưởng vì task CPython dài nhất khoảng 15 s.
- **Thang phần thưởng của bộ học** được giữ cố định ở 120 s (−log1p(ms)/log1p(120000)), độc lập với
  timeout, nên task thành công cho cùng tín hiệu như ở TN1–2. Task thất bại được chấm như thể chạy hết
  timeout: −log1p(max(120000, timeout_ms))/log1p(120000), bằng −1 khi timeout là 120 s (như TN1–2) và
  ≈ −1,14 khi timeout là 600 s, để không task thành công nào bị chấm tệ hơn task thất bại.
- **Phiên:** mỗi cell là một phiên 2 build liên tiếp trên cùng coordinator (`SESSION_BUILDS=2`). Build 1
  là build "học"; **build 2 là build đo chính**. Lý do: warm-start 100 quyết định chiếm phần lớn một build
  abseil (152 file), nên một build đơn gần như không đo được bộ học.
- **Thiết kế:** randomised complete block, **20 block** mỗi kịch bản, thứ tự arm xáo trộn có seed mỗi
  block (`SEED_BASE=3000`, khác TN1–2); build làm nóng bị loại; coordinator khởi động lại mỗi cell.
- **Tham số scheduler mặc định:** α = 0,5, warm-start = 100, λ = 0,5 (như TN1–2). LinUCB thuần: không
  warm-start, không phạt tải.
- **Kiểm định:** mỗi metric, mỗi thư mục kịch bản, mỗi họ là một họ so sánh; Wilcoxon signed-rank hai
  phía trên các block hoàn chỉnh, hiệu chỉnh Holm trong họ, effect size matched-pairs rank-biserial r,
  median hiệu ghép cặp với CI 95 % bootstrap. Ngưỡng α = 0,05 sau Holm. Giả thuyết có hướng được báo cáo
  kèm hướng quan sát được.

## 1. Workload (cố định)

| Mã | Workload | Tham số |
|---|---|---|
| syn | C++/C tổng hợp, `scripts/gen_synthetic_cpp.py` | `SYN_SEED=3001`, 300 file, C 30 %, Pareto α = 1,5, median 120 đơn vị, trần 4.500 đơn vị, chi phí ẩn 0 |
| absl | abseil-cpp LTS `20250512.1`, commit `76bb24329e8bf5f39704eb10d21b9a80befa7c81` | Release, C++17, `ABSL_BUILD_TESTING=OFF`; 152 file biên dịch |

Ngưỡng P90 (kích thước tiền xử lý, byte; phân vị 90 tuyến tính trên các file của một build LeastLoaded
pilot; task "lớn" khi kích thước **lớn hơn hẳn** ngưỡng):
- absl: **2.356.237 byte** (16/152 file vượt).
- syn: **150.189 byte** (build LeastLoaded pilot ngày 2026-10-22, `SYN_SEED=3001`; khoảng 30 file vượt).
- E3-syn-drift dùng ngưỡng của syn.

## 2. Kịch bản và scheduler

| Thư mục | Workload | Drift |
|---|---|---|
| **E3-syn** | syn | không |
| **E3-absl** | absl | không |
| **E3-syn-drift** | syn | vĩnh viễn: `stress-ng --cpu 1` trong worker 1,1 CPU từ `DRIFT_AT=120` (như E2-perm); build 2 diễn ra hoàn toàn trong drift |

Arm (7, như nhau ở mọi kịch bản): `leastloaded sed p90 sed-p90 linucb hybrid-linucb hybrid-linucb-d-arm-g098`.
- `p90`: file lớn vào máy có `cpu_millis` lớn nhất còn slot (hoà theo thứ tự registry), còn lại như
  LeastLoaded.
- `sed-p90`: như trên, phần còn lại như SED.
- `hybrid-linucb-d-arm-g098`: HG-LinUCB-D, γ = 0,98, đồng hồ discount theo từng máy. Chọn bản này dựa
  trên kết quả TN2, quyết định ở mục 6.

## 3. Metric

Mọi metric phụ tính **trên build 2** của mỗi cell (một giá trị mỗi cell; không coi từng file là mẫu độc
lập):
- **Chính:** `build_2_makespan_s`.
- **Phụ:**
  - `top5_strong_share`: trong top 5 % task lớn nhất theo kích thước tiền xử lý (⌈0,05 n⌉, hoà theo
    `dispatch_seq`), tỷ lệ gửi vào 2 máy có `cpu_millis` lớn nhất.
  - `top5_mean_cpu_millis`.
  - `tail_s`: `end_ts` của build trừ `dispatch_ts` lớn nhất của các task thành công.
  - `last_finisher_size_bytes` và `last_finisher_rank_pct`.
  - `top5_multi_choice_share` (mẫu số: tỷ lệ quyết định top 5 % có ≥ 2 máy còn slot).
- Makespan build 1 và tổng phiên được báo cáo mô tả.

## 4. Giả thuyết

Họ A (mặc định của analyzer, mỗi kịch bản, mỗi metric): treatment `hybrid-linucb`,
`hybrid-linucb-d-arm-g098` × baseline `leastloaded`, `sed`, `p90`, `sed-p90` = 8 so sánh.
Họ B (mỗi kịch bản, mỗi metric): `--treatments p90 sed-p90 --baselines leastloaded` = 2 so sánh.

| GT | Kịch bản | Metric | So sánh | Hướng | Họ |
|---|---|---|---|---|---|
| H3.1 | E3-syn | `build_2_makespan_s` | `hybrid-linucb` vs `leastloaded` | nhỏ hơn | A (8) |
| H3.2 | E3-syn | `build_2_makespan_s` | `p90` vs `leastloaded` | nhỏ hơn | B (2) |
| H3.3 | E3-absl | `build_2_makespan_s` | `hybrid-linucb` vs `leastloaded` | nhỏ hơn | A (8) |
| H3.4 | E3-syn | `top5_strong_share` | `hybrid-linucb` vs `leastloaded` | lớn hơn | A (8) |
| H3.5 | E3-syn | `tail_s` | `hybrid-linucb` vs `leastloaded` | nhỏ hơn | A (8) |
| H3.6 | E3-syn-drift | `build_2_makespan_s` | `hybrid-linucb-d-arm-g098` vs `p90` | nhỏ hơn | A (8) |

**Số liệu thiếu:** mỗi dòng giả thuyết cần đủ 20 cặp giá trị hữu hạn.
- Nếu một block có giá trị NaN ở metric của giả thuyết mà không vi phạm tiêu chí loại ở mục 5, thì dòng
  giả thuyết đó được báo cáo là "không đánh giá được".
- Không bỏ cặp, không nội suy, không chạy lại block vì lý do này.
- Kiểm toán `n = 20` được làm trước khi đọc kết quả.

Khám phá (không kiểm định xác nhận):
- HG-LinUCB so với P90 và SED + P90 ở E3-syn và E3-absl (không dự đoán hướng); các so sánh này đã có
  trong họ A.
- LinUCB thuần so với HG-LinUCB: chạy sau chiến dịch với tham số cố định
  `analyze_exp.py <OUT> --strict --treatments linucb --baselines hybrid-linucb --out <OUT>/analysis_linucb_vs_hg`.
- E3-absl so với E3-syn về mức độ khác biệt.
- Mọi metric phụ khác.

## 5. Tiêu chí loại dữ liệu và chạy lại (cố định)

Một cell bị coi là hỏng khi xảy ra một trong các trường hợp sau:
- build thoát khác 0, hoặc log có `FAILED:`;
- thiếu `HGTIME`;
- ít hơn 5 worker đăng ký;
- có task thất bại, hoặc số task thành công nhỏ hơn số của build làm nóng.

Với E3-syn-drift, cell còn bị coi là hỏng khi:
- không có `drift_on` với `dispatch_count` ≥ 101;
- `drift_on` đầu tiên có `dispatch_count` > `DRIFT_AT` + 40;
- có ít hơn 100 task hoàn thành sau sự kiện.

Quy trình chạy lại có hai tầng như TN1–2:
- Cell hỏng được chạy lại ngay 1 lần.
- Nếu hỏng lần nữa thì bỏ cả block; campaign chạy lại nguyên block, tối đa 2 lần mỗi kịch bản (sổ
  `reruns.csv`).
- Không ghép phần block cũ với block mới.
- Tạm dừng theo yêu cầu không tính là chạy lại; được ghi vào `interruptions.csv`.

Một kịch bản chỉ được phân tích khi đủ 20 block.

## 6. Quyết định đã chốt (người dùng xác nhận ngày 2026-10-22)

1. **Phạm vi:** đăng ký này gồm syn, absl, syn + drift. protobuf và OpenCV, nếu thầy yêu cầu, sẽ là đăng
   ký riêng sau, vì runner chưa hỗ trợ (protobuf còn phụ thuộc abseil).
2. **Bản HG-LinUCB-D:** "chỉ máy vừa nhận kết quả quên" (`hybrid-linucb-d-arm-g098`, γ = 0,98). Chọn dựa
   trên TN2: bản "mọi máy cùng quên" chậm hơn HG-LinUCB có ý nghĩa, còn bản này thì không.
3. **Phiên 2 build, đo build 2** (mục 0). Đã cân nhắc phương án 1 build khoảng 300 file và loại: build
   được đo khi đó có 1/3 đầu do warm-start LeastLoaded quyết định ở mọi scheduler, làm loãng khác biệt.
4. **Thời điểm chạy:** trước khi thầy xem báo cáo TN1–2, theo yêu cầu của người dùng; bắt đầu tối
   2026-10-22.

## 7. Mối đe dọa đến tính hợp lệ (ghi nhận trước)

- Một máy, một cụm Docker (như TN1–2).
- Workload tổng hợp được thiết kế để kích thước dự đoán tốt chi phí (ρ ≈ 0,95 khi biên dịch riêng lẻ;
  0,73 trên cụm). Kết quả ở syn vì vậy có lợi cho các scheduler dùng kích thước; abseil là phép kiểm chứng
  trên project thật.
- Ngưỡng P90 được tính từ một build pilot LeastLoaded của chính workload đó.
- Kích thước tiền xử lý của file C++ thật lớn (median khoảng 1,5 MB với abseil); chiều [1] của LinUCB
  bão hoà ở 4 MiB, nhưng không file nào vượt mức này.
- Phiên 2 build khác giao thức của bài báo và của TN1.
- `SEED_BASE` khác TN1–2 nhưng dùng chung cho cả ba kịch bản TN3.
- H3.6 dùng `hybrid-linucb-d-arm-g098`, được chọn sau khi đã xem kết quả TN2. Đây là quyết định có
  thông tin trước, được ghi rõ ở đây.
- Thời gian ước tính từ pilot ngày 2026-10-22 (1 block mỗi kịch bản): E3-absl khoảng 21 phút/block
  (≈ 7 h cho 20 block), E3-syn khoảng 60 phút/block (≈ 20 h), E3-syn-drift khoảng 52 phút/block
  (≈ 17,5 h); tổng ≈ 45 h, sai số khoảng ±15 %, chưa kể chạy lại block.
