# Kế hoạch chạy benchmark đối chiếu icecream (icecc)

Tài liệu này để **thực thi trên máy khác**. Đọc hết mục 0 và 1 trước khi chạy bất cứ lệnh nào.

Mục tiêu: trả lời câu hỏi *"workload của chúng ta chạy bằng thuật toán của icecream thì tốt hơn hay tệ hơn?"*, và gỡ Limitation (6) của paper (chưa so với hệ thống production nào).

---

## 0. Bối cảnh — vì sao lại làm theo cách này

icecream **không** phải heuristic tĩnh. Đọc `scheduler/scheduler.cpp` của
github.com/icecc/icecream (master, 2608 dòng, tải ngày 2026-08-02) thì thuật toán mặc định
`fastest` học throughput từng node **online** từ các job đã hoàn thành:

```cpp
float f = (float)cs->cumCompiled().outputSize()
          / (float) cs->cumCompiled().compileTimeUser();      // dòng 263
f *= float(1000 - cs->load()) / 1000;                         // dòng 311
f *= (1.0f - (0.5f * cs->currentJobCount() / cs->maxJobs()));  // dòng 319
if (cs->lastCompiledJobs().size() < 7)
    f *= (-0.5 * cs->lastCompiledJobs().size() + 4.5);        // dòng 323
```

cộng với `pick_server_new()` — ưu tiên node chưa từng đo — và chọn ngẫu nhiên khi cả cụm chưa
có số liệu nào. Nói theo ngôn ngữ bandit: **bộ ước lượng giá trị chỉnh tay, không ngữ cảnh,
kèm khám phá ad-hoc**.

Vì vậy có **hai thí nghiệm khác nhau**, đừng lẫn lộn:

| | Đo cái gì | Giá trị | Trạng thái |
|---|---|---|---|
| **(A)** Port thuật toán icecc vào Hybrid-Grid | **Chỉ luật chọn worker** | Nội tại — sạch, không confound | ✅ Code xong, sẵn sàng chạy |
| **(B)** Chạy `icecc` thật, so đầu-cuối | Cả hệ thống | Ngoại tại — thực tế | ⏳ Chưa dựng, xem mục 5 |

**(A) là phần có khoa học.** Cùng transport, cùng cache, cùng admission rule, cùng workload —
khác biệt duy nhất là luật chọn worker, nên chênh lệch makespan quy được về thuật toán.
Chạy (A) trước.

---

## 1. Chuẩn bị

### 1.1. Mang code sang

Nhánh `KienNT`. Các thay đổi cần có mặt:

```
internal/coordinator/scheduler/icecc.go        (mới) port thuật toán icecc
internal/coordinator/scheduler/icecc_test.go   (mới) 9 test
internal/coordinator/scheduler/scheduler.go    thêm eligibleCandidates() dùng chung
internal/coordinator/scheduler/heft.go         ủy quyền sang helper trên
internal/coordinator/server/grpc.go            factory: case "icecc-fastest"
cmd/hg-coord/main.go                           hợp lệ hóa cờ --scheduler
scripts/analyze_rigorous.py                    BASELINES → suy ra từ dữ liệu
```

> ⚠️ `scripts/analyze_rigorous.py` trước đây hardcode `BASELINES = ["leastloaded","p2c","linucb"]`.
> Nếu quên mang file này sang, `icecc-fastest` vẫn hiện trong bảng tóm tắt nhưng **bị loại
> khỏi toàn bộ kiểm định cặp** — ra một bảng không ai đối chiếu được. Kiểm tra kỹ.

### 1.2. Yêu cầu máy

| Hạng mục | Cần |
|---|---|
| Docker | Đang chạy, engine Linux |
| CPU | **≥ 6 lõi rảnh** — cụm chiếm 5.25 (workers 4.0 + builder 1.0 + coordinator 0.25) |
| RAM | ≥ 8 GB cấp cho Docker |
| Đĩa | ≥ 20 GB (source CPython + layer Docker + cache) |
| Go | 1.25.x |
| Python | có `pandas`, `numpy`, `scipy`, `matplotlib` |

**Máy phải rảnh.** Nhiễu run-to-run là mối đe dọa nội tại số một của thí nghiệm này
(§7.4 của paper). Đóng IDE, trình duyệt, trình đồng bộ. Tắt sleep/hibernate. Cắm điện, đặt
power plan ở chế độ hiệu năng cao — throttling nhiệt chính là confound mà thiết kế khối
sinh ra để chống, đừng làm nó phải gánh thêm.

### 1.3. Kiểm tra trước khi chạy

```bash
go build ./...
go test ./internal/coordinator/... -count=1
python -c "import pandas, numpy, scipy, matplotlib; print('deps ok')"
docker info --format 'cpus={{.NCPU}} mem={{.MemTotal}}'
```

Cả 4 lệnh phải xanh. Riêng test: **chỉ chạy `./internal/coordinator/...`**.
`./...` có một số test executor hỏng sẵn không liên quan đến thay đổi này, và trên Windows
thì `-race` không dùng được.

Xác nhận scheduler mới đã nối dây:

```bash
go run ./cmd/hg-coord serve --scheduler=bogus 2>&1 | head -2
# phải in: ... must be one of: leastloaded, simple, p2c, epsilon-greedy,
#          linucb, hybrid-linucb, heft, icecc-fastest
```

---

## 2. Bước 1 — Smoke test (~30 phút)

**Đừng bỏ qua bước này.** Nó bắt lỗi tích hợp trước khi tiêu 3 giờ.

```bash
cd <repo>
SCHEDULERS="leastloaded p2c linucb hybrid-linucb icecc-fastest" \
REPS=2 \
OUT_DIR=/tmp/bench_icecc_smoke \
bash scripts/benchmark_rigorous.sh
```

Nghiệm thu — **2 vòng không đủ lực thống kê, đừng đọc p-value**. Chỉ kiểm tra bốn thứ:

1. **Không cell nào lỗi.** Script `exit 1` ngay khi build fail, nên chạy hết là đạt.
2. **`icecc-fastest` có mặt** trong bảng của analyzer, cả phần tóm tắt lẫn phần Wilcoxon.
3. **Dispatch có trải ra 5 worker** — nếu dồn hết vào 1 worker là port sai:

   ```bash
   python - <<'EOF'
   import pandas as pd, glob
   for f in sorted(glob.glob("/tmp/bench_icecc_smoke/tasks_icecc-fastest_round*.jsonl")):
       d = pd.read_json(f, lines=True)
       print(f.split("_")[-1], "| tasks:", len(d))
       print(d["worker_id"].value_counts().to_string())
       print("exploration:", f"{d['was_exploration'].mean():.1%}")
   EOF
   ```

4. **Tỷ lệ exploration nhỏ.** icecc chỉ khám phá ở hai chỗ (cụm chưa đo, node chưa đo), nên
   trên ~293 task với 5 worker kỳ vọng **dưới 5%**. Nếu thấy hàng chục phần trăm → nghi
   `RecordOutcome` không được gọi, khiến `s.recorded` kẹt ở 0 và nhánh random chạy mãi.

Nếu (3) hoặc (4) sai, **dừng lại**, đừng chạy tiếp — báo lại kèm output.

---

## 3. Bước 2 — Chạy đầy đủ (~2–3 giờ)

```bash
SCHEDULERS="leastloaded p2c linucb hybrid-linucb icecc-fastest" \
REPS=10 \
COOLDOWN=8 \
OUT_DIR=/tmp/bench_icecc_full \
bash scripts/benchmark_rigorous.sh 2>&1 | tee /tmp/bench_icecc_full.log
```

Giữ nguyên 4 scheduler cũ để **bảng mới đối chiếu trực tiếp được với Bảng 3 của paper**.
Bỏ bớt là mất khả năng đó.

**Nếu bị ngắt giữa chừng**, script hỗ trợ tiếp tục — seed theo chỉ số vòng nên vòng chạy lại
khớp hệt như chạy liền mạch:

```bash
ROUND_START=6 REPS=10 OUT_DIR=/tmp/bench_icecc_full \
SCHEDULERS="leastloaded p2c linucb hybrid-linucb icecc-fastest" \
bash scripts/benchmark_rigorous.sh
```

Ước tính thời gian: 5 scheduler × 10 vòng = 50 cell, cộng 1 warm-up. Mỗi cell gồm compose
down/up, chờ 5/5 worker đăng ký, cooldown 8s, `make clean`, xóa cache, rồi build ~40s →
khoảng **2.5–4 phút/cell**.

---

## 4. Thu thập và diễn giải

### 4.1. Cần mang về

```
/tmp/bench_icecc_full/results.csv              ← quan trọng nhất
/tmp/bench_icecc_full/order_log.csv            ← chứng minh thứ tự đã ngẫu nhiên hóa
/tmp/bench_icecc_full/tasks_*_round*.jsonl     ← 50 file, phân tích offline
/tmp/bench_icecc_full/makespan_rigorous.png
/tmp/bench_icecc_full.log
```

Chạy lại analyzer bất cứ lúc nào:

```bash
python scripts/analyze_rigorous.py /tmp/bench_icecc_full
```

### 4.2. Đọc kết quả thế nào

Analyzer kiểm định một phía `hybrid-linucb < baseline`. Với `icecc-fastest`:

| Kết quả | Nghĩa là | Viết vào paper thế nào |
|---|---|---|
| `SIGNIFICANT`, med diff âm | Hybrid-LinUCB **nhanh hơn** icecc | Học có ngữ cảnh thắng chỉnh tay thủ công — kết quả mạnh nhất có thể |
| `n.s.` | **Hòa** với icecc | Trung thực, và nhất quán với kết quả hòa LeastLoaded đã có |
| `n.s.` nhưng med diff **dương** rõ | icecc **nhanh hơn** ta | Cũng là kết quả tốt — cần chạy thêm Wilcoxon chiều ngược lại để khẳng định |

**Dự đoán trước khi chạy** (ghi lại đây để khỏi bị thiên kiến hậu nghiệm): icecc-fastest sẽ
**xấp xỉ LeastLoaded và Hybrid-LinUCB**, tức nằm trong nhóm ~40s, và **nhanh hơn P2C với
LinUCB thuần**. Lý do: LeastLoaded tuy mù về tốc độ node nhưng tự sửa sai qua backpressure
(worker chậm giữ `ActiveTasks` cao nên tự ngừng nhận việc), còn `server_speed` của icecc đo
đúng đại lượng đó một cách tường minh — hai đường đến cùng một chỗ trên workload đồng nhất.

Nếu kết quả **lệch hẳn dự đoán**, nghi ngờ port trước khi nghi ngờ thuật toán: kiểm tra lại
phân bố dispatch và tỷ lệ exploration ở mục 2.

### 4.3. Phân tích bổ sung nên làm

```bash
python - <<'EOF'
import pandas as pd, glob, re
rows = []
for f in glob.glob("/tmp/bench_icecc_full/tasks_*_round*.jsonl"):
    sched = re.search(r"tasks_(.+)_round", f).group(1)
    d = pd.read_json(f, lines=True)
    vc = d["worker_id"].value_counts()
    rows.append({"scheduler": sched,
                 "skew_top_bottom": vc.max() / max(vc.min(), 1),
                 "p99_ms": d["compile_time_ms"].quantile(0.99),
                 "p50_ms": d["compile_time_ms"].quantile(0.50)})
df = pd.DataFrame(rows).groupby("scheduler").median()
print(df.to_string())
EOF
```

Độ lệch tải (`skew_top_bottom`) là chỗ icecc **có thể** khác hẳn: hệ số bão hòa
`(1 - 0.5·jobs/maxJobs)` của nó phạt yếu hơn nhiều so với LeastLoaded, nên nhiều khả năng
lệch tải cao hơn. Paper đã báo cáo chỉ số này cho các scheduler khác (13.2:1 với ε-greedy,
8.6:1 với P2C) nên có sẵn mốc so sánh.

---

## 5. Giai đoạn (B) — chạy `icecc` thật, đầu-cuối

Làm **sau** khi (A) xong. Đây là thí nghiệm ngoại tại: nó đo cả hệ thống, không cô lập được
thuật toán, nhưng là thứ thầy hướng dẫn muốn thấy — "so với cái người ta đang dùng thật".

### 5.1. Ba confound bắt buộc phải xử lý

Không xử lý thì con số thu được vô nghĩa.

1. **icecc ưu tiên máy submitter.** Nó chạy job ngay trên máy gửi nếu máy đó rảnh và đạt
   ≥70% tốc độ node nhanh nhất. Container `builder` đang có `cpus: '1.0'` sẽ tự nuốt một
   phần lớn job → không còn là bài toán phân tán 5 worker nữa. **Bắt buộc** đặt daemon trên
   builder về 0 slot (`iceccd -m 0`), hoặc không chạy `iceccd` trên builder.
2. **Nén payload.** icecc nén job khi truyền; gRPC của ta mặc định không. Chênh lệch thu
   được sẽ lẫn phần này. Hoặc bật nén cho ta, hoặc ghi rõ trong phần threats to validity.
3. **Cache.** Ta có CAS nội bộ (xxhash) và harness xóa cache mỗi lần chạy. icecc thường ghép
   với ccache. **Phải tắt ccache** (`unset CCACHE_PREFIX`) để cả hai phía đều không cache.

### 5.2. Khung việc

1. Viết `test/stress/Dockerfile.icecc` — cài `icecc` (Debian/Ubuntu: `apt-get install icecc`).
2. Viết `test/stress/docker-compose-icecc.yml` — 1 `icecc-scheduler`, 5 `iceccd` với **đúng
   cùng cgroup CPU** 0.5/0.6/0.8/1.0/1.1, 1 builder chạy `iceccd -m 0`.
3. Cùng nguồn CPython, cùng `CONFIGURE_FLAGS`, cùng `-j5`, cùng cooldown, cùng warm-up bỏ đi.
4. Xen kẽ `hgbuild` và `icecc` trong **cùng một vòng khối** để chia sẻ nhiễu host — không
   chạy hết bên này rồi mới sang bên kia, đó chính là lỗi run-order confounding mà paper đã
   phát hiện và sửa ở §6.2.
5. Đo bằng cùng đồng hồ `python time.time()` bao quanh **chỉ** lệnh build.

Số liệu icecc lấy được từ `icecc --scheduler-log` hoặc `icemon`, nhưng **không có** gì tương
đương 27 trường JSONL của ta — nên với (B) chỉ so được makespan, không so được phân bố
dispatch. Đây là lý do (A) mới là phần mang tính khoa học.

---

## 6. Sau khi có số — sửa paper

Ba chỗ hiện **đang sai** và phải sửa bất kể kết quả benchmark ra sao:

1. `docs/thesis/paper-hybridgrid-en.md` §2.4 dòng 50: *"None learn online."* → **Sai.**
   Thay bằng một đoạn mô tả đúng `server_speed()`, có trích công thức. Bản thân đoạn đó đã là
   đóng góp khảo sát: chưa thấy paper nào mô tả thuật toán này tử tế.
2. §1 dòng 19: *"heuristics ... never learn from observed outcomes"* → **Sai** với icecream.
3. §4.4(b): hệ số bão hòa hàng đợi của icecc trùng ý tưởng cấu trúc với `λ·LoadRatio`. Khác
   biệt còn lại là thật nhưng **hẹp** — icecc chỉ có phạt thủ công, ta có LoadRatio vừa là
   feature học được `x[7]` vừa là prior thủ công. Nêu đúng mức, đừng thổi phồng.

Và định vị lại đóng góp: **không phải** "đầu tiên học online trong biên dịch phân tán" (sai),
mà là **"formulation contextual bandit có nguyên lý, so với ước lượng throughput vô hướng
chỉnh tay"** — đúng và bảo vệ được.

Bản ghi chi tiết: `~/.claude/.../memory/icecream-learns-online.md`.

---

## 7. Sự cố thường gặp

| Triệu chứng | Nguyên nhân / xử lý |
|---|---|
| `only N/5 workers registered` | Máy thiếu CPU. Cần ≥6 lõi rảnh. Tăng timeout trong `wait_registration` không giải quyết gốc. |
| Build fail, script `exit 1` | Xem `/tmp/build_rig_<sched>-r<N>.log`. Thường do hết đĩa hoặc OOM. |
| `icecc-fastest` không có trong bảng Wilcoxon | Quên mang `scripts/analyze_rigorous.py` sang. Xem mục 1.1. |
| Exploration của icecc-fastest > 20% | `RecordOutcome` không được gọi → `s.recorded` kẹt 0. Kiểm tra type assertion `LearningScheduler` ở `grpc.go`. |
| Makespan vọt gấp đôi ở một vòng lẻ | Đã gặp với hybrid-linucb ở workload nặng (một vòng 101.93s). Giữ lại điểm đó, **đừng loại** — nó là tail risk có thật, paper đã báo cáo. |
| Lỗi path trên Windows/Git-Bash | Đường dẫn `/tmp` của Git-Bash ánh xạ sang `C:\Users\<user>\AppData\Local\Temp`. Dùng `pwd -W` để lấy đường dẫn Windows thật khi cần đưa cho công cụ ngoài. |
| `gofmt -l` báo gần như mọi file | CRLF sẵn có trong working tree Windows. **Đừng "sửa"** — sẽ tạo diff khổng lồ giả. CI Linux checkout LF nên không ảnh hưởng. |
