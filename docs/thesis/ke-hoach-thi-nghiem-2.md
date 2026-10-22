# Kế hoạch chạy Thí nghiệm 2 (drift ẩn)

> Bản v3 (bản cuối), ngày 2026-10-17, sau hai vòng phản biện của Codex (`gpt-5.6-sol`; v1 và
> v2 đều NOT READY; phản biện lưu ở `D:\bench-work\review\plan_tn2_review1.md`, `…review2.md`).
> Phần còn lại được kiểm bằng review mã và kết quả test ở mục 3 trước khi mở cổng.
> Đây là tài liệu vận hành, **không** phải tài liệu đăng ký trước. Cấu hình, giả thuyết,
> metric và tiêu chí loại dữ liệu đã đóng băng trong `docs/thesis/hypotheses.md` (commit
> `1b8339d`); runner và analyzer chạy đúng commit đó. Kế hoạch chỉ sửa **công cụ điều phối
> ngoài repo** (`D:\bench-work\*.sh`) để thực thi đăng ký chặt hơn; nó không đổi thiết kế.
> Chỗ nào mâu thuẫn với `hypotheses.md` thì `hypotheses.md` đúng.

## 0. Hiện trạng

| Hạng mục | Trạng thái |
|---|---|
| TN1 (E1-j6 … E1-2c) | Xong 2026-10-11: 400 cell, 0 cell hỏng, 0 lần chạy tiếp, `analysis rc=0` |
| Clone chạy `D:\bench-work\hybridgrid` | `1b8339d`, sạch, LF |
| Máy | Docker 72 CPU / 31 GiB, 0 container, không có app gây nhiễu, preflight không FAIL |
| Ổ D: | còn 74 GB (TN1 dùng 273 MB; TN2 ước tính ≤ 1 GB) |
| Cổng đợt 2 | Chưa có `out/official/PHASE2_APPROVED` |
| Pilot drift | Trước 17/10 chỉ có một cell E2-trans `leastloaded` với cấu hình cũ (`BUILDER_CPUS=1.0`, commit `07088e1`). Smoke S1–S4 ngày 17/10 (cấu hình chính thức, `out/pilot-e2/`) đều đạt (mục 3.2) |
| Codex | `gpt-5.5` bị máy chủ từ chối từ sau 11/10 (tài khoản ChatGPT). Reviewer/supervisor chuyển sang `gpt-5.6-sol` (khác model worker `gpt-6-sol`) |

## 1. Thiết kế đã đăng ký (tóm tắt, nguồn: hypotheses.md §2, §2b, §3)

- Sự kiện: `stress-ng --cpu 1` chạy **bên trong** `worker-5` (1.1 CPU, 3 slot) khi bộ đếm
  dispatch ≥ 120. Hạn mức CPU không đổi.
- Mỗi cell = phiên 2 build liên tiếp trên cùng coordinator (`SESSION_BUILDS=2`, ~590 quyết
  định), `-j8`, 1 client, `BUILDER_CPUS=2.0`. Injector chỉ khởi động ở build 1; khoảng giữa
  hai build đồng hồ drift dừng, stress-ng giữ trạng thái. Makespan phiên = makespan build 1 +
  makespan build 2.
- Mỗi lần gọi runner với `ROUND_START=1` chạy thêm một **cell làm nóng** (2 build, bị loại).

| Thư mục | DRIFT | Arm | Cell đo |
|---|---|---|---:|
| E2-perm | permanent | 7 arm (A7) | 140 |
| E2-trans | transient, 30 s | A7 | 140 |
| E2-onoff | onoff, 10 s / 10 s | A7 | 140 |
| E2-lambda | permanent | 10 arm (AL) | 200 |

A7 = `leastloaded sed hybrid-linucb hybrid-linucb-d-g095 hybrid-linucb-d-g098 hybrid-linucb-d-g099 hybrid-linucb-d-arm-g098`.
AL = `leastloaded sed hybrid-linucb hybrid-linucb-l0 hybrid-linucb-l025 hybrid-linucb-l100 hybrid-linucb-d-g098 hybrid-linucb-d-g098-l0 hybrid-linucb-d-g098-l025 hybrid-linucb-d-g098-l100`.

### Giả thuyết ↔ cột kết quả ↔ họ kiểm định

| Giả thuyết | Thư mục | Cột (tên chính xác) | Cặp | Họ Holm | Số so sánh |
|---|---|---|---|---|---:|
| H2.1 (a) | E2-perm | `time_to_adapt_rank` | `hybrid-linucb-d-g098` vs `hybrid-linucb` | `analysis_d_vs_hg` | 4 |
| H2.1 (b) | E2-perm | `drift_post_share` | như trên | `analysis_d_vs_hg` | 4 |
| H2.2 | E2-trans | `target_after_recovery_share` | như trên | `analysis_d_vs_hg` | 4 |
| H2.3 (a) | E2-perm | `makespan_s` | `hybrid-linucb-d-g098` vs `leastloaded` | `analysis` (mặc định) | 10 |
| H2.3 (b) | E2-onoff | `makespan_s` | như trên | `analysis` | 10 |
| H2.4 | E2-lambda | `drift_post_share` | median theo λ | không kiểm định | — |

**Cách báo cáo (không thêm quy tắc phân loại):** hypotheses.md chỉ quy định Wilcoxon hai phía,
Holm, α = 0,05 sau Holm và hướng của giả thuyết. Vì vậy mỗi dòng ở bảng trên được báo cáo
riêng: n ghép cặp, `n_zero`, median hai arm, median hiệu ghép cặp, CI 95 %, `r_rb`, `p_raw`,
`p_holm` và hướng quan sát được. Không gộp H2.1(a)+(b) thành "ủng hộ/một phần". Nếu cần một
câu tóm tắt cho thầy thì ghi rõ đó là **tóm tắt trình bày, không phải quy tắc đăng ký**.
Mọi so sánh khác analyzer sinh ra (γ = 0,95/0,99, `arm`, SED, metric phụ, cột `build_*`,
`drift_phase_*`) là **mô tả**.

## 2. Ngân sách thời gian

Số đo (cập nhật 17/10 sau smoke test, cấu hình chính thức): makespan phiên 2 build của các arm
HG-LinUCB-D có drift là 230–240 s (S1–S4; pilot cũ `leastloaded`, builder 1.0 CPU: 199 s). Phần
ngoài build (dựng/dọn cụm, khoảng giữa hai build, làm nóng chia đều) 25–45 s/cell (pilot: 23 s;
TN1: ~43 s). Dùng **≈ 272 s/cell** (260–285 s). Hypotheses.md ghi sai số ước tính ±15 %; ước tính
~25 h trong hypotheses.md dựa trên 222 s/cell nên thấp hơn thực tế; đây không phải thay đổi thiết
kế, chỉ là thời gian chạy.

| Bước | Cell đo | Ước tính điểm | Khoảng ±15 % |
|---|---:|---:|---:|
| Chuẩn bị + smoke test (mục 3) | 5 + 5 làm nóng (+ test dừng/chạy tiếp) | ~1,5 h (đã làm) | — |
| E2-perm | 140 | 10,6 h | 9,0–12,2 h |
| E2-trans | 140 | 10,6 h | 9,0–12,2 h |
| E2-onoff | 140 | 10,6 h | 9,0–12,2 h |
| **Ba biến thể** | 420 | **32 h** | **27–36,5 h** |
| E2-lambda | 200 | 15,1 h | 12,8–17,4 h |
| **Tổng** | 620 | **47 h** | **40–54 h** |

Ngân sách giữ máy = cận trên ±15 % + số lần chạy lại block tối đa theo đăng ký (2 mỗi kịch bản;
một block A7 ≈ 32 phút, một block AL ≈ 45 phút): ba biến thể 36,5 + 6 × 0,53 ≈ **40 h**; cả λ
54 + 6 × 0,53 + 2 × 0,76 ≈ **59 h**. Thời điểm xong dự kiến tính theo ước tính điểm từ lúc bắt đầu
campaign; ngân sách là mức xấu nhất.

## 3. Việc trước khi bắt đầu

### 3.1 Sửa công cụ điều phối (ngoài repo, Codex worker `gpt-6-sol` làm, Claude giám sát, Codex reviewer duyệt)

Runner/analyzer không đổi (vẫn `1b8339d`). Mọi sửa đổi chỉ ở `D:\bench-work`, ghi vào
`D:\bench-work\CHANGELOG-phase2.md` kèm SHA-256 của file trước/sau, và được nêu trong báo cáo
như một thay đổi công cụ vận hành.

| # | File | Sửa | Lý do (phản biện) |
|---|---|---|---|
| T1 | `campaign.sh` | Sổ chạy lại **ghi nối tiếp** `$out/reruns.csv` (thời điểm, block, runner exit code, số lần trước/sau); số lần = số dòng của sổ, không reset khi gọi lại campaign; trần cố định 2 (bỏ biến `MAX_RESUMES` có thể ghi đè); lần hỏng thứ 3 → dừng chiến dịch | §3 "tối đa 2 lần mỗi kịch bản" (R1-3, R2-3) |
| T2 | `campaign.sh` | Trước mỗi lần chạy lại block k, chuyển **mọi** file của block đó (`tasks_*_round<k>*`, `events_*_round<k>*`, `stats_*_round<k>*`, `build_*_round<k>_*` kèm `.status`) và sao `results.csv`, `order_log.csv`, `failures.csv`, `meta.json`, `runner.log` sang `$out/archive/rerun_<n>_<YYYYmmdd-HHMMSS>/`; khi k = 1 thêm cả file của round 0 (làm nóng) | Giữ bằng chứng lỗi; runner ghi đè (R1-3, R2-4) |
| T3 | `campaign.sh` | Đọc quyết định λ từ `PHASE2_APPROVED` như **dữ liệu** (không `source`): phải có đúng một dòng khớp `^E2_LAMBDA=(RUN|SKIP)$`; thiếu, trùng hoặc sai → từ chối trước khi chạy cell nào. `SKIP` → tự đặt `STOP_AFTER=E2-onoff`; `STOP_AFTER` truyền vào mâu thuẫn với quyết định → từ chối. Ghi quyết định đã đọc vào `campaign.log` | Quyết định một lần (R1-5, R2-11) |
| T4 | `campaign.sh` | Mọi lần phân tích thành công (tất cả họ của kịch bản, rc = 0) thì **ghi sau cùng** `$out/analysis.done` (lệnh, SHA-256 của analyzer, SHA-256 các file đầu ra). Kịch bản đủ 20 block mà có `analysis.done` hợp lệ (hash khớp) thì không phân tích lại; thiếu/không khớp thì phân tích lại. Với 5 thư mục E1: tạo `analysis.done` một lần bằng tay sau khi kiểm `analysis*.log` của TN1 đều rc 0 và hash đầu ra (ghi trong CHANGELOG) | Không ghi đè E1; không nhận kết quả dở (R1-15, R2-5) |
| T5 | `campaign.sh` + `stop_campaign.sh` (mới) | Campaign ghi PID của chính nó vào `$CAMPAIGN/campaign.pid` và PID runner vào `$CAMPAIGN/runner.pid`; trước mỗi lần (chạy lại) runner, nếu có `$CAMPAIGN/STOP_REQUESTED` thì thoát sạch, không tính là chạy lại. `stop_campaign.sh "<lý do>"`: tạo `STOP_REQUESTED`, ghi `MANUAL STOP <lý do>` vào `campaign.log`, gửi TERM cho runner, chờ tối đa 120 s campaign thoát; quá hạn thì kill campaign/runner; sau đó luôn chạy `docker compose -p hgexp -f <file compose của runner> down` và kiểm theo nhãn `com.docker.compose.project=hgexp` rằng không còn container. Runner thoát do dừng tay **không** tính vào T1 (campaign thấy `STOP_REQUESTED`). Chạy tiếp: xoá `STOP_REQUESTED`, gọi lại lệnh chuẩn; block dở được chạy lại nguyên block, file của nó được lưu theo T2 | Runner mồ côi (R1-2, R2-2) |
| T6 | `supervisor.sh` | Chỉ xét các dòng `campaign.log` ghi **sau** khi supervisor khởi động, gồm cả `MANUAL STOP`; khoá chống chạy hai supervisor (`supervisor.lock`); model `gpt-5.6-sol`; khi `SEVERITY: ALERT` hoặc `WARN` → **chỉ thông báo** (toast Windows + dòng trong `supervisor.log`), **không tự dừng** campaign | Supervisor tự thoát ngay (R1-9); dừng theo đầu ra mô hình là quy tắc dừng chưa đăng ký (R2-C1) |
| T7 | `supervisor_prompt.md` | Thêm mục E2, tách rõ **tiêu chí đã đăng ký** (§3: build exit, `HGTIME`, 5 worker, task thất bại, số task ≥ build làm nóng, `drift_on` đầu tiên có `dispatch_count` trong [101, 160], ≥ 100 task sau đó, quy tắc chạy lại) và **chẩn đoán chỉ để báo cáo** (590 task/cell, E2-trans đúng 1 cặp on/off, stress-ng rò, pha bắc qua ranh giới build là bình thường); chẩn đoán không bao giờ là lý do loại/chạy lại; vẫn cấm diễn giải hiệu quả scheduler | R1-13, R2-9 |

Điều kiện **dừng tự động duy nhất** vẫn là những cái đã có và độc lập với kết quả scheduler:
preflight STRICT fail, thư mục lẫn cấu hình (`MISMATCH`), runner hỏng quá 2 lần (T1), phân
tích báo block bị loại.

### 3.2 Kiểm thử công cụ (không dùng số liệu, thư mục `D:\bench-work\out\pilot-e2\`)

1. **Test có sẵn (đã chạy 17/10, đạt):** `bash scripts/tests/test_bench_exp_labels.sh` (kiểm
   `-arm`, `-g`, `-l0/-l025/-l100`) và `python3 -m unittest scripts/tests/test_analyze_exp.py`
   (26/26) trong clone `1b8339d`.
2. **Test điều phối bằng runner giả** (`D:\bench-work\tests\test_campaign.sh`, chạy trên bản
   sao `campaign.sh` trong thư mục tạm; lệnh gọi runner được thay bằng script giả trả về chuỗi
   exit code định trước và tạo file round như runner thật; bản chính thức không thêm biến cấu
   hình nào):
   (a) đi qua 5 thư mục E1 (bản sao) có `analysis.done` → không phân tích lại, hash không đổi;
   xoá một file đầu ra → phân tích lại (T4);
   (b) thiếu `PHASE2_APPROVED` → dừng ở cổng; `E2_LAMBDA` thiếu/trùng/sai → từ chối; `SKIP` không
   có `STOP_AFTER` → tự dừng sau E2-onoff; `RUN` + `STOP_AFTER=E2-onoff` → từ chối (T3);
   (c) runner hỏng ở block 1 hai lần, gọi lại campaign giữa chừng → `reruns.csv` vẫn đếm đúng,
   lần hỏng thứ 3 → dừng; mỗi lần có `archive/rerun_*` đủ file của block, `failures.csv` cũ còn
   trong archive (T1, T2);
   (d) có `STOP_REQUESTED` → campaign thoát sạch, không ghi vào `reruns.csv` (T5);
   (e) supervisor (Codex thay bằng script giả) với `campaign.log` có dòng kết thúc cũ → không
   thoát; thêm `MANUAL STOP` mới → đúng một audit cuối rồi thoát; supervisor thứ hai bị khoá
   từ chối (T6).
3. **Bốn cell thật cấu hình chính thức** (mỗi cái `REPS=1`, có cell làm nóng):

   | Cell | DRIFT | Arm | Kiểm thêm |
   |---|---|---|---|
   | S1 | permanent | `hybrid-linucb-d-arm-g098` | có `drift_off … cell_end` |
   | S2 | transient | `hybrid-linucb-d-g099` | đúng 1 cặp `drift_on`/`drift_off` không tính `cell_end`; `target_after_recovery_share` khác NaN |
   | S3 | onoff | `hybrid-linucb-d-g095` | on/off xen kẽ; không có sự kiện nào có timestamp nằm giữa lúc kết thúc build 1 và `build_start b=2`; cuối cell có thể kết thúc ở pha tắt (không bắt buộc `cell_end`) |
   | S4 | permanent | `hybrid-linucb-d-g098-l025` | cờ `--load-penalty=0.25` trong compose |

   Mọi cell: 590 `task_completed`, 0 thất bại; `drift_on` đầu tiên có `dispatch_count` trong
   [101, 160] (ngưỡng chấp nhận đã đăng ký; bình thường ≥ 120); ≥ 100 task sau `drift_on`;
   sau cell không còn container/stress-ng; `analyze_exp.py <dir> --strict` rc = 0.
4. **Test dừng/chạy tiếp thật 2 build** (S5): bản sao campaign trong thư mục thử với `REPS=2`,
   một arm (`hybrid-linucb`), DRIFT permanent, Docker thật. Giữa block 2 gọi `stop_campaign.sh`
   → trong 120 s không còn container nhãn `hgexp`, không còn runner; gọi lại → chạy lại đúng
   block 2, block 1 giữ nguyên, file dở của block 2 nằm trong `archive/`, `reruns.csv` rỗng.
5. Lệnh cụ thể dùng đường dẫn Git Bash có ngoặc kép, ví dụ:
   `ARMS="hybrid-linucb-d-g099" REPS=1 JOBS=8 CLIENTS=1 SESSION_BUILDS=2 DRIFT=transient BUILDER_CPUS=2.0 OUT_DIR="/d/bench-work/out/pilot-e2/S2" bash scripts/bench_exp.sh`
6. Lỗi nằm trong runner/analyzer (repo) → **dừng, báo người dùng**: sửa repo nghĩa là commit
   mới khác `EXPECT_COMMIT`, phải quyết định riêng và ghi thành sai lệch.

### 3.3 Chuẩn bị máy

- Tạm dừng Windows Update 7 ngày; campaign tự bật High performance (tắt ngủ) và trả lại.
- Đóng app nặng (đã làm 17/10); không dùng máy trong lúc chạy.
- Lưu manifest SHA-256 của toàn bộ `out/official/E1-*` (dữ liệu thô + `analysis*`) vào
  `out/official/manifest_E1.sha256`; so lại sau chiến dịch.

### 3.4 Quyết định một lần: có chạy E2-lambda không

| Phương án | Giữ máy | Hệ quả |
|---|---|---|
| **RUN** | ~47 h (tối đa 59 h) | H2.4 được đánh giá (mô tả); kết quả ba biến thể vẫn gửi thầy khi xong, không chờ λ |
| SKIP | ~32 h (tối đa 40 h) | H2.4 ghi "không được đánh giá"; muốn chạy λ sau này phải là kịch bản bổ sung |

Đề xuất: **RUN** nếu máy rảnh được ~2 ngày, vì thầy đã nêu quét λ và bỏ thì không bật lại được
trong cùng đăng ký.

## 4. Lệnh chạy (chỉ sau khi người dùng duyệt kế hoạch và chọn RUN/SKIP)

```bash
# 1. Ghi quyết định và mở cổng (một lần). File chỉ chứa đúng một dòng; lý do ghi vào campaign.log.
#    Lần đọc đầu tiên campaign khoá quyết định vào PHASE2_DECISION; đổi file sau đó bị từ chối.
printf 'E2_LAMBDA=RUN\n' > /d/bench-work/out/official/PHASE2_APPROVED
printf '%s PHASE2_APPROVED E2_LAMBDA=RUN reason=<...>\n' "$(date '+%F %T')" >> /d/bench-work/out/official/campaign.log

# 2. Campaign (lệnh chuẩn, dùng cả khi chạy tiếp; không cần STOP_AFTER vì T3 tự đặt)
cd /d/bench-work && EXPECT_COMMIT=1b8339d BUILDER_CPUS=2.0 nohup bash campaign.sh >> out/official/stdout.log 2>&1 &

# 3. Supervisor
nohup bash /d/bench-work/supervisor.sh 5400 >> /d/bench-work/out/official/supervisor_stdout.log 2>&1 &

# Dừng an toàn khi cần dùng máy
bash /d/bench-work/stop_campaign.sh "<lý do>"
```

## 5. Vận hành trong lúc chạy

- **Bịt mắt kết quả:** trong lúc chạy chỉ xem tính toàn vẹn (số cell, `failures.csv`, sự kiện
  drift, cấu hình). Không mở `analysis*/` của bất kỳ thư mục E2 nào cho đến khi **cả ba**
  E2-perm/trans/onoff xong (campaign vẫn tự chạy phân tích; chỉ không đọc). E2-lambda: không mở
  cho đến khi xong.
- **Giám sát:** supervisor 90 phút/lần (~18–20 audit cho ba biến thể, ~8 nữa cho λ).
  `ALERT`/`WARN` → chỉ thông báo (T6). Claude (khi phiên còn mở) hoặc người dùng đọc bằng
  chứng; chỉ dừng tay bằng `stop_campaign.sh` khi xác nhận một vấn đề **toàn vẹn/an toàn** độc
  lập với kết quả scheduler (ví dụ cấu hình lệch, container rò, máy quá tải), không bao giờ vì
  makespan hay xu hướng. Mỗi lần dừng tay ghi lý do. Claude không bảo đảm thức suốt 40 h, nên
  người dùng là người nhận thông báo; thời gian phản hồi mục tiêu ≤ 12 h.
- **Lỗi:** đúng hypotheses.md §3 (runner chạy lại cell 1 lần → hỏng tiếp thì bỏ block →
  campaign chạy lại block, tối đa 2 lần/kịch bản theo sổ `reruns.csv` → quá thì dừng).
- **Khi xong mỗi kịch bản:** ghi SHA-256 toàn bộ file thô của thư mục vào
  `out/official/manifest_<kịch bản>.sha256`.

## 6. Phân tích sau khi xong

1. **Kiểm toán trước khi suy luận** (với mỗi dòng ở bảng mục 1): `n = 20`, `n_zero`, không có
   NaN, đủ số so sánh trong họ (4 hoặc 10); E2-trans: đúng một cặp on/off ở mọi cell. Metric
   NaN không phải tiêu chí loại đã đăng ký, nên nếu xảy ra thì báo cáo n thực tế (Wilcoxon chạy
   trên số cặp còn lại) hoặc "không đánh giá được"; **không** đặt quy tắc loại/chạy lại mới.
2. Bảng xác nhận H2.1–H2.3 theo cách báo cáo ở mục 1.
3. Bảng mô tả mỗi thư mục (tên cột chính xác): `makespan_s`, `build_1_makespan_s`,
   `build_2_makespan_s`, `drift_pre_all_share`, `drift_pre_share`, `drift_post_share`,
   `drift_post_early_share`, `drift_post_late_share`, `time_to_adapt` + tỷ lệ kiểm duyệt
   (`time_to_adapt_censored`), `excess_target_share_post`, `target_after_recovery_share`
   (E2-trans), `drift_on_phases_share` / `drift_off_phases_share` (E2-onoff),
   `last_finisher_address`. E2-onoff thêm: thời lượng pha quan sát (từ timestamp sự kiện),
   `dispatch_count` của từng sự kiện on/off, số task mỗi pha, pha bắc qua ranh giới build.
   Mọi thư mục: báo cáo số lần chạy lại cell/block và lý do (từ `failures.csv`, `reruns.csv`,
   `archive/`).
4. **Phụ lục khám phá (không đăng ký trước, không dùng ngôn ngữ xác nhận), định nghĩa chốt
   ngay bây giờ:** (a) kiểm tra sự kiện có làm chậm worker-5, chỉ trên E2-perm: task hợp lệ là
   task `success=true`, `compile_time_ms > 0`, `preprocessed_size_bytes > 0` (đếm và báo cáo số
   task bị bỏ); tỷ số r = `compile_time_ms / (preprocessed_size_bytes / 1e5)` (ms trên 100 kB).
   Với mỗi cell và mỗi worker ∈ {worker-5, worker-4}: median r trước onset ([1, onset)) và sau
   onset ([onset, hết phiên]); chỉ số của cell = (median sau / median trước) của worker-5 chia
   cho cùng tỷ số của worker-4; cell thiếu task ở một trong bốn ô thì bỏ và đếm. Báo cáo median
   và IQR của chỉ số này qua 140 cell (gộp mọi arm), không kiểm định; (b) hình tỷ lệ trượt
   20 quyết định gửi vào worker-5 theo `dispatch_seq`, trung vị qua 20 block, mỗi arm một
   đường, vạch onset/offset trung vị.
5. Báo cáo gửi thầy sau ba biến thể, cùng định dạng báo cáo TN1.

## 7. Hạn chế báo cáo trước (không sửa sau khi có kết quả)

- `time_to_adapt` là chỉ số thứ bậc bị chặn (`n_post + 1` khi không đạt), không phải phân
  tích sống sót; mức chặn khác nhau giữa các cell; nhiều giá trị trùng khiến Wilcoxon dùng
  xấp xỉ. Báo cáo tỷ lệ kiểm duyệt từng arm và n ghép cặp.
- H2.3 nằm trong họ 10 so sánh (5 treatment × 2 baseline), 20 block: sức mạnh thấp. Không nới
  họ sau khi thấy kết quả.
- Bật/tắt chỉ kiểm soát xấp xỉ (poll 0,2 s, độ trễ `docker exec`, đồng hồ dừng giữa build);
  số quyết định mỗi pha phụ thuộc scheduler.
- H2.4 chỉ dựa trên median theo λ; các so sánh với LL/SED analyzer sinh ra (16 cặp mỗi metric)
  không phải kiểm định xu hướng.
- Thứ tự kịch bản cố định (nhiễu theo thời gian không được ngẫu nhiên hoá giữa kịch bản); thứ
  tự arm trong block dùng cùng cơ chế seed (`SEED_BASE=1000`) ở mọi kịch bản nên block k có cùng
  hoán vị giữa các kịch bản có cùng danh sách arm; một máy, một workload; phiên 2 build khác giao
  thức của bài và của TN1.
- `gpt-5.5` không còn dùng được; supervisor TN2 dùng `gpt-5.6-sol`, khác TN1.

## 8. Rủi ro còn lại

| Rủi ro | Dấu hiệu | Xử lý |
|---|---|---|
| `drift_on` quá muộn (> 160) | `failures.csv` | §3 |
| stress-ng rò sang cell sau | `stress-ng cleanup failed`, cell hỏng | Runner đánh dấu hỏng; supervisor kiểm |
| Không ai đọc thông báo ALERT kịp | toast/`supervisor.log` | Người dùng kiểm tra ít nhất sáng/tối; các điều kiện dừng tự động xác định vẫn hoạt động |
| Windows Update / mất điện | khoảng trống trong log | Tạm dừng update; chạy tiếp từ block dở |
| Sự kiện quá yếu | phụ lục 4(a) không thấy chậm | Báo cáo đúng; đề xuất kịch bản bổ sung nếu cần |

## 9. Không làm

- Không sửa `hypotheses.md`, runner, analyzer; không đổi arm, γ, λ, `DRIFT_AT`, số block.
- Không xem so sánh scheduler trước khi xong cả ba biến thể; không dừng sớm vì kết quả.
- Không push, không commit kết quả thô khi chưa được đồng ý.
- Không tạo `PHASE2_APPROVED` khi người dùng chưa duyệt kế hoạch và chọn RUN/SKIP.
