# Giả thuyết đăng ký trước: Thí nghiệm 1 và 2 (HG-LinUCB)

> ĐĂNG KÝ TRƯỚC — được commit một lần, ngay trước lượt chạy chính thức đầu
> tiên, và không được sửa sau khi đã có kết quả (quy định chung số 3 của
> thầy). Mọi thay đổi sau đó chỉ được ghi thành kịch bản bổ sung ở file
> khác, kèm lý do. Các pilot ngày 2026-10-10 (một cell hoặc một block mỗi
> cấu hình) chỉ dùng để kiểm tra công cụ và chọn tham số; số liệu pilot
> không được gộp vào kết quả.
>
> Tài liệu này đăng ký cả Thí nghiệm 1 và 2. Thí nghiệm 1 chạy trước; Thí
> nghiệm 2 và E2-lambda chỉ bắt đầu sau khi nhóm xem kết quả Thí nghiệm 1 và
> xác nhận (mục 2b). Cấu hình Thí nghiệm 2 ở đây không được đổi sau khi đã
> thấy kết quả Thí nghiệm 1.

## 0. Cấu hình chung (cố định)

- Mã nguồn: nhánh `feat/exp-readiness`, commit chứa file này; commit được
  ghi trong `meta.json` của mỗi lượt chạy. Máy chạy: xem mục 4. Trước mỗi
  lần bắt đầu hoặc chạy tiếp một thư mục kịch bản chạy `preflight.sh` với
  `STRICT=1` và `EXPECT_COMMIT` = commit chứa file này; báo cáo lưu ở
  `preflight_r<k>.txt` cạnh kết quả (k = block bắt đầu). Toàn bộ chiến dịch
  chạy bằng `campaign.sh` theo thứ tự cố định ở mục 2b.
- Cụm 5 worker 0.5/0.6/0.8/1.0/1.1 CPU, max-parallel 1/2/2/2/3 (10 slot),
  coordinator 0.25 CPU, mỗi builder 2.0 CPU (`BUILDER_CPUS=2.0`, mọi kịch
  bản). Lý do (pilot 2026-10-10 trên máy chính thức, 1 cell mỗi cấu hình,
  1 client `-j8`): với 1.0 CPU builder chạm hạn mức ở 25 % số mẫu `docker
  stats` (p90 100 %); với 2.0 CPU builder dùng tới khoảng 1.2 CPU (p90
  123 %), tức hạn mức 1.0 CPU đã bó phía client. Bài báo dùng 1.0 CPU trên
  một máy khác, nên makespan tuyệt đối không so sánh trực tiếp với bài.
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
  `leastloaded`, `sed`. Ở E2-perm, E2-trans, E2-onoff, so sánh HG-LinUCB-D
  với HG-LinUCB là một họ riêng, chạy bằng lệnh thứ hai với nhãn liệt kê
  đầy đủ (analyzer không nhận ký tự đại diện): `--treatments
  hybrid-linucb-d-g095 hybrid-linucb-d-g098 hybrid-linucb-d-g099
  hybrid-linucb-d-arm-g098 --baselines hybrid-linucb --out
  <OUT_DIR>/analysis_d_vs_hg`.
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

Metric chính: makespan. Metric phụ dùng định nghĩa "rảnh" của Mục 7.4 /
Bảng 5 trong bài: worker khoẻ không có task nào đang chạy tại thời điểm
quyết định, đọc từ ảnh chụp cụm `candidates` trong task log:
- `strongest_idle`: trong các quyết định có ít nhất một worker rảnh, tỷ lệ
  chọn worker rảnh có `cpu_millis` lớn nhất;
- skip-idle (`skip_idle_all`, `skip_idle_given_idle`): gửi task vào worker
  đang bận trong khi có worker rảnh;
- `had_idle_share`: tỷ lệ quyết định có ít nhất một worker rảnh (mẫu số,
  như cột "had idle" của Bảng 5), luôn báo cáo kèm hai tỷ lệ trên.
Metric phụ thêm theo định nghĩa "còn slot" (active < max_parallel):
`free_slot_skip_all`, `free_slot_skip_given_free`, `strongest_free_slot`,
`had_free_share`. Ở `-j8`/`-j9` ít quyết định có worker rảnh hoàn toàn, nên
định nghĩa "còn slot" được báo cáo để thấy phần còn lại.

Giả thuyết:
- H1.1 Ở E1-j8, E1-j9 và E1-2c, makespan của HG-LinUCB nhỏ hơn LeastLoaded.
- H1.2 Ở mọi kịch bản E1, `strongest_idle` của HG-LinUCB lớn hơn
  LeastLoaded.
- Skip-idle so với LeastLoaded chỉ báo cáo mô tả, không kiểm định định
  hướng: LeastLoaded chọn worker có ít task đang chạy nhất (số task thô,
  không chia theo CPU; hoà thì theo thứ tự duyệt registry), nên khi có
  worker rảnh nó luôn chọn một worker rảnh và skip-idle của nó gần như bằng
  0 theo cấu tạo. Phần khác 0 nhỏ, nếu có, chỉ là sai lệch do ảnh chụp
  `candidates` và lựa chọn của scheduler được đọc ở hai thời điểm cách nhau
  rất ngắn. Chỉ số này chỉ cho biết HG-LinUCB có tệ hơn LeastLoaded hay
  không, không thể cho thấy nó tốt hơn. Cũng vì LeastLoaded không phân biệt
  worker rảnh mạnh hay yếu, `strongest_idle` (H1.2) là chỉ số có thể tách
  hai scheduler.
- H1.3 SED có skip-idle lớn hơn LeastLoaded (SED dồn việc vào worker mạnh
  dù có worker yếu đang rảnh). Cố định: metric là `skip_idle_all` (tính trên
  mọi quyết định); so sánh SED với LeastLoaded là một họ riêng, chạy bằng
  lệnh `analyze_exp.py <OUT_DIR> --strict --treatments sed --baselines
  leastloaded --out <OUT_DIR>/analysis_sed_vs_ll` ở mỗi kịch bản E1;
  Wilcoxon hai phía hiệu chỉnh Holm trong từng kịch bản E1 (họ gồm riêng cặp
  này, theo từng metric). `skip_idle_given_idle` chỉ báo cáo mô tả.

## 2. Thí nghiệm 2: drift ẩn

Sự kiện: `stress-ng --cpu 1` chạy bên trong container worker 1.1 CPU
(`worker-5`) khi bộ đếm dispatch của coordinator đạt 120 (`DRIFT_AT=120`);
không dùng `docker update --cpus`. Mỗi cell là một phiên 2 build liên tiếp
trên cùng coordinator (`SESSION_BUILDS=2`, khoảng 590 quyết định), `-j8`,
1 client. Thời điểm và số dispatch của mỗi sự kiện được ghi vào task log
(`injected_event`).

Lý do chọn phiên 2 build (quyết định của nhóm, cố định tại thời điểm đăng
ký; câu trả lời của thầy về Q8, nếu đến sau, chỉ được ghi thành nhận xét,
mục 5): trên máy chính thức một build 295 TU kéo dài khoảng 94 s; trong
pilot, biến thể 30 s bật ở dispatch ~124 và hồi phục ở dispatch ~250, nên
một build chỉ còn khoảng 45 quyết định sau hồi phục, còn phiên 2 build còn
khoảng 340; phiên dài cũng cho discount thêm lịch sử trước drift. Đây là
điểm khác với giao thức cũ của bài (khởi động lại coordinator mỗi build);
Thí nghiệm 1 giữ giao thức cũ. Khoảng cách thực giữa hai sự kiện bật/tắt
dài hơn `DRIFT_DURATION` vài giây (độ trễ `docker exec`), nên phân tích dùng
mốc ghi trong task log, không dùng 30 s danh nghĩa.

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
- H2.4 (khám phá, không kiểm định xu hướng) Trong E2-lambda, khi λ giảm
  thì tỷ lệ gửi vào worker bị chậm sau sự kiện của HG-LinUCB và
  HG-LinUCB-D giảm (học được nhiều hơn từ phần thưởng). Chỉ báo cáo median
  theo từng mức λ và các so sánh cặp mà analyzer tạo ra; không có kiểm định
  xu hướng và không rút kết luận xác nhận từ H2.4.

## 2b. Thứ tự chạy, ngân sách và điều kiện dừng

- Thứ tự cố định: E1-j6, E1-j7, E1-j8, E1-j9, E1-2c, E2-perm, E2-trans,
  E2-onoff, E2-lambda. Mỗi thư mục kịch bản chạy đủ 20 block rồi mới sang
  kịch bản sau.
- Ước tính trên máy chính thức (ngoại suy từ 3 cell pilot, sai số khoảng
  ±15 %): TN1 ~15 h (400 cell), TN2 ba biến thể ~25 h (420 cell),
  E2-lambda ~12 h (200 cell, 10 arm); tổng ~52 h.
- Một kịch bản chỉ được phân tích khi đủ 20 block hoàn chỉnh. Nếu runner
  dừng vì lỗi lặp lại, chạy tiếp từ block bị dừng (`ROUND_START`), không
  chạy lại các block đã hoàn thành. Kịch bản không hoàn thành được báo cáo
  là "không hoàn thành", không phân tích một phần.
- Chiến dịch được chia thành hai đợt có xác nhận ở giữa: đợt 1 gồm các kịch
  bản E1-*, đợt 2 gồm E2-perm, E2-trans, E2-onoff rồi E2-lambda. Script
  `campaign.sh` tự dừng sau E1-2c và chỉ chạy đợt 2 khi người vận hành tạo
  tệp `PHASE2_APPROVED` trong thư mục kết quả sau khi xem kết quả đợt 1.
  Việc xem kết quả đợt 1 không được dẫn tới đổi cấu hình, giả thuyết hay
  danh sách arm của đợt 2; muốn đổi thì chạy thành kịch bản bổ sung và nói
  rõ là bổ sung.
- Việc có chạy E2-lambda hay không được quyết định MỘT LẦN, khi tạo
  `PHASE2_APPROVED` (trước khi có bất kỳ kết quả TN2 nào): nếu không chạy thì
  đặt `STOP_AFTER=E2-onoff` và ghi quyết định cùng lý do (thời gian, thiết
  bị) vào `campaign.log`. Sau đó không đổi quyết định. Nếu E2-lambda không
  chạy thì H2.4 được báo cáo là "không được đánh giá". Kết quả gửi thầy sau
  khi xong TN1 và ba biến thể TN2, không chờ E2-lambda.

## 3. Tiêu chí loại dữ liệu (cố định trước)

Một cell bị coi là hỏng và ghi vào `failures.csv` nếu: build thoát khác 0,
thiếu `HGTIME`, ít hơn 5 worker đăng ký, có task thất bại, số task thành
công nhỏ hơn số của build làm nóng; với drift: không có `drift_on` với
`dispatch_count` ≥ 101, `dispatch_count` của sự kiện `drift_on` đầu tiên
> `DRIFT_AT` + 40 (analyzer đặt onset = `dispatch_count` + 1 để tính cửa sổ,
nhưng ngưỡng loại dùng chính `dispatch_count` như runner), hoặc ít hơn 100
task hoàn thành sau sự kiện đó.

Chạy lại có hai tầng, cố định trước:
1. Cell hỏng được runner chạy lại ngay 1 lần (cụm mới).
2. Nếu lần chạy lại cũng hỏng, runner dừng và block đó bị bỏ toàn bộ:
   `campaign.sh` chạy lại cả block từ đầu (mọi arm, cùng seed và thứ tự),
   tối đa 2 lần cho mỗi kịch bản; quá số đó thì dừng chiến dịch để kiểm
   tra bằng tay. Không bao giờ giữ một phần block cũ ghép với block mới.

Mọi lần cell hỏng được ghi trong `failures.csv` (kèm lý do) và `runner.log`
của kịch bản; `campaign.log` ghi mỗi lần runner dừng và chạy tiếp. Tất cả
được báo cáo kèm kết quả. Phân tích chỉ dùng block hoàn chỉnh; một kịch bản
chỉ được phân tích khi đủ 20 block hoàn chỉnh (mục 2b).

## 4. Mối đe dọa đến tính hợp lệ (ghi nhận trước)

Máy chạy là Windows 10 + Docker Desktop (WSL2), hai CPU Xeon E5-2686 v4
2.3 GHz (36 nhân/72 luồng, không có P/E core); máy dùng chung với công
việc khác nên các ứng dụng nền (IDE, trình duyệt) được đóng và power plan
đặt High performance trong lúc chạy. Báo cáo kiểm tra môi trường
(`preflight_r<k>.txt`: commit, phiên bản thư viện, tài nguyên Docker, CPU, hệ
điều hành, power plan, ứng dụng nền) được lưu cùng thư mục kết quả của từng
kịch bản; `meta.json` ghi thêm nproc, CPU/RAM Docker và giới hạn của từng
worker. Pilot ngày 2026-10-10 chạy trên một máy khác (i5-14600KF, P/E
core), nên thời gian pilot không dùng để ước lượng thời gian trên máy chính
thức. Một phép đo thay thế (biên dịch một file C tổng hợp bằng `gcc -O2`,
5 lần lặp, độ lệch khoảng 2 %) cho thấy thời gian biên dịch trong container
trên máy chính thức dài hơn khoảng 2.6 lần (hạn mức 1.1 CPU) và 3.1 lần
(0.5 CPU) so với máy pilot; vì vậy các hằng số tính theo giây (30 s,
10 s/10 s) tương ứng một phần khác của build so với pilot;
builder được nâng lên 2.0 CPU vì 1.0 CPU chạm hạn mức ở 1 client `-j8`
(mục 0); mức sử dụng của builder và coordinator vẫn được ghi trong
`stats_*.csv` của từng cell; stress-ng cạnh tranh trong chính hạn mức CFS
của worker;
LeastLoaded tự giảm gửi vào worker chậm qua số task đang chạy nên không
"mù" hoàn toàn; đặc trưng [8] (độ trễ RPC) của LinUCB là hằng; reward chỉ
dùng thời gian biên dịch.

## 5. Quyết định đã chốt và câu hỏi gửi thầy

Đã chốt (2026-10-10):
- (A) Chạy toàn bộ trên máy ở mục 4.
- (B) Thí nghiệm 2 dùng phiên 2 build cho cả ba biến thể (mục 2).
- (D) "Rảnh" = không có task nào đang chạy (Mục 7.4 / Bảng 5 của bài);
  "còn slot" là phụ; mẫu số `had_idle_share` / `had_free_share` luôn báo cáo.
- (E) Thứ tự, ngân sách và điều kiện dừng như mục 2b.
- (C) `BUILDER_CPUS=2.0` cho mọi kịch bản, theo quy tắc đặt trước khi chạy
  pilot: giữ 1.0 nếu builder không chạm hạn mức, nâng lên 2.0 nếu chạm (số
  liệu ở mục 0). Ở kịch bản 2 client, mỗi builder 1.0 CPU chỉ dùng trung vị
  25–30 % (p90 khoảng 70 %), nên 2.0 CPU không thay đổi kịch bản đó về bản
  chất.

Câu hỏi gửi thầy (không chặn việc chạy; câu trả lời đến sau khi file này đã
commit chỉ được ghi thành nhận xét trong báo cáo, không sửa file này; nếu
thầy muốn cấu hình khác thì chạy thành kịch bản bổ sung):
- Q8: phiên 2 build cho Thí nghiệm 2 có đúng ý thầy không?
- Công thức A ← γA + xxᵀ + (1−γ)I: nhóm chạy cả cách "mọi arm cùng giảm theo
  mỗi quyết định" (`global`, chính, γ ∈ {0.95, 0.98, 0.99}) và cách đọc
  nguyên văn "chỉ arm vừa nhận kết quả" (`arm`, γ = 0.98). Thầy muốn cách nào
  là chính?
- Log thô của cả đợt khoảng 400–500 MB (khoảng 50 MB khi nén): đưa lên repo
  dạng nén, dùng Git LFS, hay đính kèm riêng?
