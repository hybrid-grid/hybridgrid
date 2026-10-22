# Kế hoạch Thí nghiệm 3: workload C++ đuôi nặng

> Bản v2, ngày 2026-10-21, sau vòng phản biện 1 của Codex (`gpt-5.6-sol`, kết luận v1: NOT READY;
> phản biện lưu ở `D:\bench-work\review\plan_tn3_review1.md`). Đây là tài liệu kế hoạch, **chưa phải**
> đăng ký trước. Giả thuyết chính thức sẽ nằm trong `docs/thesis/hypotheses-tn3.md`, được đóng băng bằng
> commit **sau** pilot và **trước** lượt chạy chính thức đầu tiên.
> Nguồn yêu cầu: hướng dẫn của thầy (Thí nghiệm 3) và `phan-tich-leastloaded-va-kich-ban.md` (mục S2).

## 1. Khai thác điểm yếu nào của LeastLoaded

**Điểm yếu W2: LeastLoaded đếm số task, không nhìn độ nặng của task.**

Ví dụ siêu thị: LeastLoaded đưa khách tới quầy ít người xếp hàng nhất mà không nhìn xe hàng. Một khách đẩy
xe 100 món có thể bị xếp vào quầy chậm nhất (máy 0,5 CPU), chỉ vì quầy đó vừa trống. Các quầy khác làm xong
từ lâu, cả siêu thị vẫn chờ quầy này. Task về chót đó (**straggler**) quyết định makespan.

**Vì sao khác TN2:**
- Ở TN2, máy bị chậm làm hàng dài ra, nên LeastLoaded tự né được.
- Ở TN3, một task nặng vẫn chỉ tính là "1 task". Lúc quyết định, LeastLoaded không có cách nào biết task
  này nặng gấp 20 lần task khác.
- HG-LinUCB có thống kê **riêng cho từng worker**, gồm cả hệ số theo kích thước file (chiều [1]). Nhờ vậy,
  về nguyên tắc, nó học được "file lớn chạy trên máy 0,5 CPU rất chậm".

**Giới hạn đã biết** (phải kiểm trong pilot):
- Chiều [1] bão hoà ở 4 MiB file đã tiền xử lý, nên mọi file lớn hơn trông như nhau.
- Mỗi worker phải tự gặp đủ file nhỏ và file lớn thì mới học được.

## 2. Phát hiện chặn: hiện hệ thống chưa phân tán được build CMake + ninja

Code đã được kiểm chứng (Codex phát hiện, Claude xác nhận):
- CMake + ninja gọi compiler với `-MD -MT <file .o> -MF <file .d>`. Bộ phân tích tham số
  (`internal/compiler/parser.go:123-130`) coi file `.o` sau `-MT` là file đầu vào thứ hai, nên
  `IsDistributable()` trả về false. **Mọi file bị biên dịch tại máy build**, scheduler không có gì để
  chọn.
- `hgbuild ninja` chỉ đặt `CC`/`CXX` lúc ninja khởi động, trong khi CMake đã ghi cứng compiler vào
  `build.ninja` lúc configure. Thư mục wrapper tạm còn bị xoá khi lệnh kết thúc
  (`cmd/hgbuild/main.go:1215`).

**Phải sửa trước mọi việc khác.** Đây là sửa lỗi của hệ thống phân tán (parser và wrapper), không động tới
scheduler, có unit test:
- Parser hiểu đúng các tuỳ chọn có tham số đi kèm (`-MT`, `-MF`, `-MQ`, `-o`, `-I`, `-D`, `-include`,
  `-isystem`, `-x`…).
- File phụ thuộc `.d` vẫn được tạo cục bộ để ninja build lại đúng.
- Có wrapper `hgcc` / `hgcxx` cố định, để CMake configure thẳng với chúng.

**Cổng smoke bắt buộc:**
- Số file cần biên dịch = số task coordinator nhận.
- 0 lần biên dịch cục bộ hay fallback.
- Link thành công, chương trình test chạy được.

## 3. Thiết kế đề xuất

### 3.1 Scheduler (7 arm)

| Arm | Vai trò |
|---|---|
| LeastLoaded | baseline bị nhắm |
| SED | biết công suất máy, không biết độ nặng task |
| **P90** (mới) | heuristic của thầy: file > ngưỡng P90 thì vào máy có `cpu_millis` lớn nhất còn slot (hoà thì theo thứ tự registry); còn lại theo LeastLoaded |
| **SED + P90** (mới) | file lớn vào máy mạnh nhất còn slot, còn lại theo SED |
| **LinUCB thuần** (bổ sung theo phân tích) | không warm-start, không phạt tải (N = 0, λ = 0): tách được phần đóng góp của "học" khỏi các thành phần lai |
| HG-LinUCB | treatment chính (N = 100, λ = 0,5 như TN1–2) |
| HG-LinUCB-D | bắt buộc theo quy định chung; **cần chốt** bản nào (mục 8) |

Chi tiết cho P90 và SED + P90:
- Mỗi workload có một ngưỡng P90 riêng, tính từ các build pilot bằng một phương pháp phân vị cố định.
- Ghi lại giá trị byte của ngưỡng cùng hash của manifest nguồn.
- Phải sửa cả đường gọi scheduler: hiện chỉ `LearningScheduler` nhận `TaskContext` (kích thước file), nên
  P90 cần một giao diện có ngữ cảnh cho scheduler không học.

### 3.2 Giao thức phiên (giải quyết vấn đề warm-start)

HG-LinUCB dùng 100 quyết định đầu theo LeastLoaded (warm-start). Với abseil (152 file biên dịch) như vậy chỉ còn
khoảng 52 quyết định do HG-LinUCB tự chọn. Đề xuất:
- Dùng **phiên 2 build** như TN2: build 1 là build "học", build 2 là build đo chính, với cùng coordinator
  và mô hình đã học.
- **Metric chính là makespan của build 2.** Makespan của build 1 vẫn được báo cáo.
- Scheduler không học thì hai build giống nhau, nên vẫn so sánh công bằng.

### 3.3 Kịch bản (mỗi kịch bản 20 block, cùng cụm 5 worker 0,5/0,6/0,8/1,0/1,1 CPU)

| Mã | Workload | Ghi chú |
|---|---|---|
| **E3-syn** | C++ tổng hợp khoảng 300 file, độ nặng theo Pareto (đuôi nặng), C/C++ khoảng 30/70, thứ tự cố định theo seed | biết trước file nào nặng |
| **E3-absl** | abseil-cpp LTS `20250512.1` (commit `76bb243`), 152 file biên dịch (156 file `.cc` không phải test trong repo) | project thật, thầy yêu cầu đầu tiên |
| **E3-protobuf** | protobuf (bản ghim cố định, có upb viết bằng C) | thầy yêu cầu thứ hai |
| E3-opencv | OpenCV core + imgproc | thầy yêu cầu thứ ba; xem mục 8 |
| **E3-syn-drift** | E3-syn + drift vĩnh viễn (stress-ng trên máy 1,1 CPU) | tổ hợp "TN3 + drift" thầy yêu cầu; cần đủ file lớn sau thời điểm drift **ghi nhận thực tế** |

### 3.4 Metric (gộp về **một giá trị mỗi build**; không coi từng file là một mẫu độc lập)

| Loại | Metric |
|---|---|
| Chính | makespan build 2 |
| Phụ (thầy yêu cầu) | top 5 % file lớn nhất: tỷ lệ đặt vào 2 máy mạnh nhất, và CPU trung bình của máy nhận |
| Phụ (thầy yêu cầu) | thời gian đuôi = `end_ts` của build − `dispatch_ts` lớn nhất của các task thành công (runner phải ghi thêm `start_ts`, `end_ts` vào `results.csv`) |
| Phụ | máy chạy task về chót và kích thước file của task đó; tỷ lệ quyết định top 5 % có ≥ 2 máy rảnh để chọn |
| Phụ (drift) | tỷ lệ file lớn gửi vào máy bị chậm sau sự kiện |
| Kiểm soát | thời gian tiền xử lý trên mỗi file, số lần fallback/OOM, mức bão hoà CPU của máy build |

### 3.5 Giả thuyết dự kiến

Trong `hypotheses-tn3.md`, mỗi giả thuyết sẽ ghi chính xác bộ ba kịch bản – cặp so sánh – metric, hướng
dự đoán, và họ Holm. Bản phác:
- **H3.1** Ở E3-syn và E3-absl, makespan build 2 của HG-LinUCB < LeastLoaded. Ở E3-syn, P90 < LeastLoaded.
- **H3.2** HG-LinUCB đặt top 5 % file lớn vào máy mạnh nhiều hơn LeastLoaded.
- **H3.3** Thời gian đuôi của HG-LinUCB < LeastLoaded.
- **H3.4** Ở E3-syn-drift, makespan build 2 của **một** treatment duy nhất chọn trước (HG-LinUCB hoặc bản
  D đã chốt) < P90.
- Khám phá: HG-LinUCB vs P90 và SED + P90; LinUCB thuần vs HG-LinUCB.

## 4. Việc phải làm

Làm trên một nhánh mới từ `feat/exp-readiness` (ví dụ `feat/exp3`). Codex `gpt-6-sol` viết code, Claude
kiểm tra, Codex `gpt-5.6-sol` review.

| # | Việc | Ước tính |
|---|---|---|
| C0 | **Sửa parser và wrapper để CMake + ninja phân tán được**, giữ file `.d`, có test; cổng smoke ở mục 2 | 1,5–2 ngày |
| C1 | Image: thêm `cmake`, `ninja-build` (Debian bookworm có cmake 3.25, ninja 1.11) | 0,5 ngày |
| C2 | Runner đa workload `WORKLOAD=cpython|syn|absl|protobuf|opencv`: ghim commit, configure ngoài cửa sổ đo, build bằng wrapper cố định, dọn bằng `ninja clean`, ghi thêm `start_ts`/`end_ts`, phiên 2 build | 1,5 ngày |
| C3 | Bộ sinh workload tổng hợp (có seed, Pareto, C/C++, `CMakeLists.txt`) | 1–1,5 ngày |
| C4 | Scheduler `p90`, `sed-p90` cùng giao diện có ngữ cảnh; nhãn arm trong runner và analyzer | 1 ngày |
| C5 | Analyzer: top 5 %, thời gian đuôi, task về chót, thời gian tiền xử lý; test | 1 ngày |
| C6 | Campaign, preflight, supervisor cho E3 | 0,5 ngày |
| P | Pilot (mục 5) | 1–1,5 ngày |
| R | `hypotheses-tn3.md`, Codex phản biện, đóng băng | 0,5–1 ngày |
| | **Tổng chuẩn bị** | **khoảng 9–11 ngày làm việc** |

## 5. Pilot và cổng "có đáng chạy không"

Chạy nhiều cell pilot với LeastLoaded cho mỗi workload; dữ liệu bỏ đi. Các tiêu chí sẽ được chốt thành
số trước khi đăng ký:
1. 100 % file cần biên dịch được gửi đi phân tán; 0 fallback, 0 lỗi, 0 OOM.
2. Máy build không phải nút cổ chai: đo thời gian tiền xử lý, mức bão hoà CPU của builder, khoảng trống
   dispatch. Thử builder 2 CPU và 4 CPU.
3. Kích thước file dự đoán được thời gian biên dịch: Spearman ρ ≥ 0,5. Tỷ lệ file vượt mức bão hoà
   4 MiB của chiều [1] phải được báo cáo; nếu đuôi bị bão hoà thì chốt lại cách chuẩn hoá trước khi đăng ký.
4. Task nặng thật sự thành straggler: với LeastLoaded, task về chót thuộc top 10 % kích thước ở phần lớn
   các cell pilot, và thời gian đuôi ≥ 10 % makespan.
5. Đủ quyết định top 5 % có ≥ 2 máy rảnh để chọn, và với E3-syn-drift có đủ file lớn sau thời điểm drift
   thực tế.
6. Với workload tổng hợp: chọn tham số để một build mất khoảng 3–4 phút và top 5 % file nặng gấp ≥ 5 lần
   median.

## 6. Thời gian

Ước tính cho phiên 2 build, 7 arm × 20 block = 140 cell mỗi kịch bản. Thời gian build của các project
thật chưa đo, nên sai số lớn cho tới khi có pilot.

| Kịch bản | Một build (ước) | Một cell | Thời gian |
|---|---:|---:|---:|
| E3-syn | ~3,5 phút | ~8 phút | ~19 h |
| E3-syn-drift | ~3,5 phút | ~8 phút | ~19 h |
| E3-absl | ~3–4 phút | ~8–9 phút | ~19–21 h |
| E3-protobuf | ~9–11 phút | ~20–23 phút | ~47–54 h |
| E3-opencv | ~8–10 phút | ~18–21 phút | ~42–49 h |

| Phạm vi | Máy chạy liên tục |
|---|---|
| Lõi: syn + syn-drift + absl | **~57–60 h (≈ 2,5 ngày)** |
| Lõi + protobuf | **~105–115 h (≈ 4,5 ngày)** |
| Đủ cả ba project của thầy | **~150–165 h (≈ 6,5–7 ngày)** |

Tổng thời gian từ lúc bắt đầu: khoảng 9–11 ngày chuẩn bị, cộng 2,5–7 ngày chạy, tức **khoảng 2,5–3,5
tuần**. Theo thứ tự thầy giao, lượt chạy chính thức nên chờ thầy xem kết quả TN1–2. Phần sửa code, image,
bộ sinh workload và pilot thì làm được ngay.

## 7. Kết quả mong muốn và khi nào thì sai

**Mong muốn:**
- P90 và HG-LinUCB nhanh hơn LeastLoaded ở E3-syn, nhờ không để file lớn rơi vào máy 0,5 CPU.
- HG-LinUCB đặt file lớn vào máy mạnh nhiều hơn, và thời gian đuôi ngắn hơn.
- Ở E3-syn-drift, bộ học tránh được máy mạnh đã bị chậm còn P90 thì không, nên **bộ học thắng P90**. Đây là
  chỗ "học" có giá trị hơn "luật cứng".

**Có thể sai vì:**
- ninja dispatch file lớn từ sớm, nên không có straggler.
- Kích thước tiền xử lý là thước đo thô của độ phức tạp template, và chiều [1] bão hoà ở 4 MiB.
- Máy build (tiền xử lý C++ cục bộ) thành nút cổ chai, xoá mờ khác biệt giữa các scheduler.
- Phần lớn quyết định chỉ có một máy rảnh để chọn.
- Phạt tải λ đẩy file lớn khỏi máy nhanh đang bận.
- Phần thưởng dạng log ít nhạy với chính các file dài nhất.
- P90 ngang hoặc thắng HG-LinUCB, vì luật cứng không cần thời gian học. Kết quả này cũng phải được báo cáo
  đầy đủ.

## 8. Cần chốt (người dùng hoặc thầy)

1. **Phạm vi project:** thầy viết "thử theo thứ tự abseil-cpp, protobuf, OpenCV". Có hai cách hiểu:
   - (a) chạy đủ cả ba;
   - (b) thử theo thứ tự, dùng project đầu tiên đáp ứng được yêu cầu.

   Đề xuất: làm lõi (syn + absl + syn-drift) trước, rồi protobuf; OpenCV tuỳ thầy. Nên hỏi thầy.
2. **Bản HG-LinUCB-D:** đề xuất bản "chỉ máy vừa nhận kết quả quên" (γ = 0,98), ghi rõ quyết định dựa trên
   TN2.
3. **Giao thức phiên 2 build, đo build 2** (mục 3.2): hay giữ 1 build và chấp nhận warm-start chiếm phần
   lớn abseil?
4. **Thời điểm chạy chính thức:** trước hay sau khi thầy xem báo cáo TN1–2.

## 9. Trạng thái môi trường (kiểm tra 21/10/2026, 22:10)

| Hạng mục | Trạng thái |
|---|---|
| Docker | 72 CPU / 31 GiB, Server 29.4.2, 0 container |
| Ổ D: | còn 73 GB (TN3 ước tính 1–3 GB log) |
| Mạng | clone được abseil-cpp từ GitHub (`D:\bench-work\pilot-src\abseil-cpp`, LTS `20250512.1`) |
| Gói cho image | Debian bookworm có `cmake` 3.25.1 và `ninja-build` 1.11.1 |
| Image Go để chạy unit test | `golang:1.25-bookworm` đã có sẵn (Windows không cài Go, CMake, ninja; không cần) |
| Image hiện tại `hgexp-*` | chỉ dành cho CPython, chưa có cmake/ninja |
| Repo | nhánh `feat/exp-readiness` ở `1b8339d`; các file báo cáo/kế hoạch chưa commit |
| Codex | `gpt-6-sol` (worker) và `gpt-5.6-sol` (reviewer) dùng được |
| App nền | Edge (29 tiến trình) và Unity đang mở: cần đóng trước pilot và lượt chạy chính thức |
| Windows Update | đang tạm dừng tới 24/10: cần gia hạn trước khi chạy dài |
| Power plan | Balanced: campaign tự chuyển sang High performance khi chạy |

## 10. Kế hoạch chạy chính thức (chốt ngày 2026-10-22, bắt đầu tối cùng ngày)

Các quyết định đã chốt: phạm vi syn + absl + syn-drift; HG-LinUCB-D bản "chỉ máy vừa nhận kết quả quên"
(γ = 0,98); phiên 2 build, đo build 2; ngưỡng P90 syn 150.189 / absl 2.356.237 byte; timeout coordinator
600 s, client 660 s. Chi tiết trong `hypotheses-tn3.md`.

### 10.1 Trạng thái trước khi chạy

| Hạng mục | Trạng thái |
|---|---|
| Sửa CMake + ninja phân tán (C0) | xong, Codex duyệt; abseil 152/152 task, CPython 295 task như TN1–2 |
| Runner đa workload, bộ sinh, P90, analyzer (C2–C5) | xong, test pass; analyzer tính lại E1/E2 giống hệt |
| Timeout và thang phần thưởng | coordinator `--request-timeout`, client `HG_TIMEOUT`, thang thưởng cố định 120 s, task thất bại chấm theo timeout; test pass |
| Điều phối `campaign_tn3.sh` | 19/19 test; khoá cứng ngưỡng P90, tách khỏi dữ liệu TN1–2 |
| Pilot | absl, syn, syn-drift: mỗi kịch bản 1 block × 7 scheduler; lần cuối 0 cell hỏng |
| Codex rà soát cuối | không lỗi chặn; mọi góp ý đã sửa |

### 10.2 Các bước tối nay (theo thứ tự; dừng nếu bước nào hỏng)

1. **Rà lại lần cuối:** chạy toàn bộ test (Go race, Python, bash), `git diff` toàn nhánh; Codex xác nhận
   nhanh bản đăng ký cuối.
2. **Đóng băng:** bỏ dòng "BẢN NHÁP" trong `hypotheses-tn3.md`; commit code + test + đăng ký + kế hoạch
   trên nhánh `feat/exp3` (không push). Ghi SHA của commit này làm `EXPECT_COMMIT`.
3. **Bản clone chạy:** `git clone` nhánh `feat/exp3` vào `D:\bench-work\hybridgrid-tn3` với
   `core.autocrlf=false`, kiểm tra HEAD = commit đóng băng, cây sạch, LF.
4. **Máy:**
   - gia hạn tạm dừng Windows Update tới khoảng 31/10 (hạn hiện tại 24/10 22:28, sẽ hết giữa lúc chạy);
   - buộc tắt app không cần thiết (giữ Antigravity, claude, codex, Remote Desktop);
   - kiểm tra Docker 72 CPU / 31 GiB, 0 container;
   - chạy `STRICT=1 EXPECT_BRANCH=feat/exp3 BENCH_REPO=<clone> EXPECT_COMMIT=<sha> preflight.sh` không có
     FAIL ngoài power plan.
5. **Chạy:**
   ```bash
   cd /d/bench-work && TN3_REPO=/d/bench-work/hybridgrid-tn3 EXPECT_COMMIT=<sha> \
     P90_SYN=150189 P90_ABSL=2356237 nohup bash campaign_tn3.sh >> out/official-tn3/stdout.log 2>&1 &
   CAMPAIGN=/d/bench-work/out/official-tn3 nohup bash supervisor_tn3.sh 5400 >> out/official-tn3/supervisor_stdout.log 2>&1 &
   ```
6. **Kiểm tra sau khoảng 20 phút:**
   - cell làm nóng và cell đầu xong;
   - đủ 2 × 301 task, 0 lỗi, không có "too many concurrent tasks";
   - audit đầu tiên OK.

### 10.3 Thứ tự và thời gian dự kiến (nếu bắt đầu khoảng 20:00 thứ Năm 22/10)

| Kịch bản | Thời gian | Xong khoảng |
|---|---:|---|
| E3-syn | ≈ 20 h | 16:00 thứ Sáu 23/10 |
| E3-absl | ≈ 7 h | 23:00 thứ Sáu 23/10 |
| E3-syn-drift | ≈ 17,5 h | 16:30 thứ Bảy 24/10 |

Sai số khoảng ±15 % (trễ nhất khoảng sáng Chủ nhật 25/10), chưa kể chạy lại block. Trong lúc chạy:
không dùng máy; nếu cần dừng thì `touch D:\bench-work\out\official-tn3\STOP_REQUESTED` (dừng sạch sau cell
hiện tại); chạy tiếp theo quy trình `resume_check` như TN2.

### 10.4 Sau khi chạy xong

1. Kiểm toán trước khi đọc kết quả: đủ 20 block mỗi kịch bản, `n = 20` và không NaN ở 6 dòng giả thuyết,
   số so sánh trong họ (A = 8, B = 2), manifest SHA-256 dữ liệu.
2. Họ A và B do campaign tự chạy; thêm so sánh khám phá LinUCB vs HG-LinUCB với tham số đã ghi trong đăng
   ký.
3. Codex tính lại độc lập các dòng xác nhận; báo cáo TN3 theo mẫu Table 3, Codex rà soát báo cáo.
