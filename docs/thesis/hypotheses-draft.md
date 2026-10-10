# Giả thuyết đăng ký trước: Thí nghiệm 1 và 2 (HG-LinUCB)

> BẢN NHÁP — chưa có hiệu lực. File này chỉ được commit một lần, ngay trước
> lượt chạy chính thức đầu tiên, và không được sửa sau khi đã có kết quả
> (quy định chung số 3 của thầy). Các pilot ngày 2026-10-10 (1 block mỗi
> kịch bản) chỉ dùng để kiểm tra công cụ và chọn tham số; số liệu pilot
> không được gộp vào kết quả.

## 0. Cấu hình chung (cố định)

- Mã nguồn: nhánh `feat/exp-readiness`, commit ghi trong `meta.json` của mỗi
  lượt chạy; cụm 5 worker 0.5/0.6/0.8/1.0/1.1 CPU, max-parallel 1/2/2/2/3
  (10 slot), coordinator 0.25 CPU, builder 1.0 CPU (`BUILDER_CPUS=1.0`).
- Workload: CPython v3.14.0, cấu hình nhẹ
  (`--disable-test-modules --disable-perf-trampoline`), 295 TU/build.
- Thiết kế: randomised complete block, **20 block**, thứ tự arm xáo trộn có
  seed mỗi block (`SEED_BASE=1000`), build làm nóng bị loại, cache xoá và
  `make clean` ngoài cửa sổ đo; coordinator khởi động lại mỗi cell (learner
  bắt đầu lạnh).
- Tham số scheduler mặc định: α = 0.5, warm-start = 100, λ = 0.5;
  HG-LinUCB-D: đồng hồ discount toàn cục (`global`) là biến thể chính,
  biến thể `arm` (γ = 0.98) là phụ.
- Thời gian build đo bên trong container builder sau một barrier chung
  (`HGTIME`); makespan của cell = max(end) − min(start) trên các client.
- Kiểm định: mỗi metric, mỗi thư mục kịch bản là một họ so sánh; Wilcoxon
  signed-rank hai phía trên các block hoàn chỉnh, hiệu chỉnh Holm trong họ,
  effect size matched-pairs rank-biserial r (âm = treatment nhỏ hơn), median
  hiệu ghép cặp với bootstrap CI 95 %. Ngưỡng α = 0.05 sau Holm.
- Treatment: các arm gốc `hybrid-linucb`/`hybrid-linucb-d`; baseline:
  `leastloaded`, `sed`. So sánh HG-LinUCB-D với HG-LinUCB là một họ riêng
  (`--treatments hybrid-linucb-d-* --baselines hybrid-linucb`).
- Lệnh phân tích: `python3 scripts/analyze_exp.py <OUT_DIR> --strict`
  (time-to-adapt: `--tta-ref pre_all`, W = 20, factor = 0.5).

## 1. Thí nghiệm 1: tải trung gian

Kịch bản (mỗi kịch bản một thư mục, arm: `leastloaded sed hybrid-linucb
hybrid-linucb-d-g098`):
- E1-j6, E1-j7, E1-j8, E1-j9: 1 client, `JOBS` = 6/7/8/9.
- E1-2c: 2 client đồng thời, mỗi client `-j8` (`CLIENTS=2`).

Tương ứng ngoài đời thật: E1-jN là một lập trình viên build với số job song
song khác nhau trên một cụm dùng chung có máy mạnh, máy yếu; E1-2c là hai
lập trình viên (hoặc hai job CI) cùng build một lúc trên cùng cụm.

Metric chính: makespan. Metric phụ: skip-idle (trên tất cả quyết định và
trên các quyết định có worker rảnh), tỷ lệ chọn đúng worker rảnh mạnh nhất
(`strongest_idle`), các biến thể "còn slot" là phụ thêm.

Giả thuyết:
- H1.1 Ở E1-j8, E1-j9 và E1-2c, makespan của HG-LinUCB nhỏ hơn LeastLoaded.
- H1.2 Ở mọi kịch bản E1, `strongest_idle` của HG-LinUCB lớn hơn
  LeastLoaded, và skip-idle của HG-LinUCB không lớn hơn LeastLoaded.
- H1.3 SED có skip-idle lớn hơn LeastLoaded (SED dồn việc vào worker mạnh
  dù có worker yếu đang rảnh).

## 2. Thí nghiệm 2: drift ẩn

Sự kiện: `stress-ng --cpu 1` chạy bên trong container worker 1.1 CPU
(`worker-5`) khi bộ đếm dispatch của coordinator đạt 120 (`DRIFT_AT=120`);
không dùng `docker update --cpus`. Mỗi cell là một phiên 2 build liên tiếp
trên cùng coordinator (`SESSION_BUILDS=2`, khoảng 590 quyết định), `-j8`,
1 client. Thời điểm và số dispatch của mỗi sự kiện được ghi vào task log
(`injected_event`).

Kịch bản (mỗi kịch bản một thư mục):
- E2-perm: chậm vĩnh viễn đến hết phiên.
- E2-trans: chậm 30 s rồi hồi phục (`DRIFT_DURATION=30`).
- E2-onoff: bật 10 s / tắt 10 s đến hết phiên.
  Arm: `leastloaded sed hybrid-linucb hybrid-linucb-d-g095
  hybrid-linucb-d-g098 hybrid-linucb-d-g099 hybrid-linucb-d-arm-g098`.
- E2-lambda (chỉ trên drift vĩnh viễn): `leastloaded sed hybrid-linucb
  hybrid-linucb-l0 hybrid-linucb-l025 hybrid-linucb-l100
  hybrid-linucb-d-g098 hybrid-linucb-d-g098-l0 hybrid-linucb-d-g098-l025
  hybrid-linucb-d-g098-l100`.

Tương ứng ngoài đời thật: một máy trong cụm bị tiến trình khác (job nền,
antivirus, người dùng khác) chiếm CPU mà hạn mức của container không đổi,
nên scheduler nhìn cấu hình thì không thấy gì; vĩnh viễn ~ một job nền chạy
dài, 30 s ~ một tác vụ bảo trì ngắn, bật/tắt ~ tải nền dao động theo chu kỳ.

Metric chính: makespan của phiên. Metric phụ: tỷ lệ quyết định gửi vào
worker bị chậm trước sự kiện (cửa sổ [101, onset) theo yêu cầu, và
[1, onset) làm tham chiếu ổn định) và sau sự kiện; time-to-adapt (số quyết
định sau onset đến khi tỷ lệ trượt 20 quyết định ≤ 0.5 × tỷ lệ tham chiếu
[1, onset); không đạt thì kiểm duyệt "> n"); excess target share; tỷ lệ theo
pha bật/tắt; tỷ lệ sau hồi phục (E2-trans); worker chạy task hoàn thành cuối.

Giả thuyết:
- H2.1 Ở E2-perm, HG-LinUCB-D (γ = 0.98, global) có time-to-adapt nhỏ hơn và
  tỷ lệ gửi vào worker bị chậm sau sự kiện thấp hơn HG-LinUCB.
- H2.2 Ở E2-trans, sau khi hồi phục HG-LinUCB-D quay lại dùng worker 1.1 CPU
  nhiều hơn HG-LinUCB (tỷ lệ sau hồi phục cao hơn).
- H2.3 Ở E2-perm và E2-onoff, makespan của HG-LinUCB-D (γ = 0.98) nhỏ hơn
  LeastLoaded.
- H2.4 (khám phá, không kiểm định định hướng) Trong E2-lambda, khi λ giảm
  thì tỷ lệ gửi vào worker bị chậm sau sự kiện của HG-LinUCB và
  HG-LinUCB-D giảm (học được nhiều hơn từ phần thưởng).

## 3. Tiêu chí loại dữ liệu (cố định trước)

Một cell bị chạy lại (tối đa 1 lần) và ghi vào `failures.csv` nếu: build
thoát khác 0, thiếu `HGTIME`, ít hơn 5 worker đăng ký, có task thất bại, số
task thành công nhỏ hơn số của build làm nóng; với drift: không có
`drift_on`, onset > `DRIFT_AT` + 40, hoặc ít hơn 100 task sau onset. Phân
tích chỉ dùng block hoàn chỉnh; mọi block bị loại được liệt kê trong
`summary.md`.

## 4. Mối đe dọa đến tính hợp lệ (ghi nhận trước)

Máy chạy là Windows + Docker Desktop, CPU lai P/E core (nhiễu thời gian);
builder 1.0 CPU có lúc chạm hạn mức khi 1 client ở `-j8` (đo bằng
`stats_*.csv`); stress-ng cạnh tranh trong chính hạn mức CFS của worker;
LeastLoaded tự giảm gửi vào worker chậm qua số task đang chạy nên không
"mù" hoàn toàn; đặc trưng [8] (độ trễ RPC) của LinUCB là hằng; reward chỉ
dùng thời gian biên dịch.

## 5. Còn chờ quyết định trước khi commit file này

1. Phiên dài cho thí nghiệm 2 (`SESSION_BUILDS=2`) có đúng ý thầy không (Q8).
2. Giữ builder 1.0 CPU (so sánh được với bài báo) hay nâng lên 2.0 để client
   không làm nghẽn ở kịch bản 1 client.
3. "Worker rảnh" = không có task nào đang chạy (như Bảng 5 của bài); biến thể
   "còn slot" chỉ là phụ.
4. Ngân sách máy: khoảng 6,6 h (TN1) + 11 h (TN2 ba biến thể) + 5,2 h (λ),
   tổng ~23 h.
