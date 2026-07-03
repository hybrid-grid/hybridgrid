# Hybrid-LinUCB: Tài liệu Triển khai (Implementation Report)

> **Ngày:** 03/07/2026
> **Phạm vi:** Triển khai đầy đủ đề xuất trong [hybrid_linucb_proposal.md](hybrid_linucb_proposal.md), kèm hai bug fix phát sinh và bộ hạ tầng benchmark thống kê mới.
> **Trạng thái:** Code hoàn thành, unit test pass, smoke benchmark 1-rep pass end-to-end. Chờ chạy full benchmark 10-rep để có số liệu đánh giá.

---

## 1. Tổng quan

Tài liệu này ghi lại các thay đổi kỹ thuật khi nâng cấp scheduler LinUCB thuần túy thành **Hybrid-LinUCB** — kiến trúc lai giữa contextual bandit và heuristic chống nghẽn. Ba cơ chế của proposal đều đã được cài đặt:

| Cơ chế | Giải quyết vấn đề | Trạng thái |
|---|---|---|
| Warm-start (N=100) | Cold start — AI "học mò" 100 task đầu | ✅ Đã xác nhận chạy đúng trong production (smoke test) |
| Hybrid Q-value (λ·LoadRatio) | Delayed feedback — dồn việc vào máy nhanh | ✅ Có unit test cô lập |
| Feature vector d=9→12 | Bộ nhận thức nông, thiếu thông tin về task | ✅ Kèm quyết định thiết kế Laplace smoothing |

Ngoài ra, quá trình smoke test phơi ra **hai bug có sẵn** của hệ thống (không liên quan scheduler) chặn đứng benchmark, và cả hai đã được sửa: phân tán nhầm file assembly, và race điều phối gây từ chối task. Đây là ví dụ điển hình của giá trị "chạy thật để kiểm chứng" — cả hai bug đều không thể phát hiện bằng unit test.

**Nguyên tắc thiết kế xuyên suốt:** không tạo struct scheduler mới. `LinUCBScheduler` được mở rộng bằng hai tham số cấu hình mặc định bằng 0 — khi cả hai bằng 0, hành vi **giống hệt LinUCB cũ từng bit**. Điều này đảm bảo so sánh benchmark `linucb` vs `hybrid-linucb` là công bằng tuyệt đối (cùng code path, chỉ khác config).

---

## 2. Tính năng mới

### 2.1. Scheduler `hybrid-linucb`

Đăng ký như một scheduler type mới, chọn qua CLI:

```bash
hg-coord serve --scheduler=hybrid-linucb --alpha=0.5 --warm-start=100 --load-penalty=0.5
```

| Flag mới | Mặc định | Ràng buộc | Ý nghĩa |
|---|---|---|---|
| `--warm-start` | 100 | ≥ 0 | Số dispatch đầu tiên giao cho heuristic least-loaded; 0 = tắt |
| `--load-penalty` | 0.5 | [0, 5] | Hệ số λ phạt tải; 0 = tắt. λ > 5 khiến penalty áp đảo mọi ước lượng đã học, thoái hóa về leastloaded — nên bị chặn |

Cả hai flag chỉ có tác dụng với `hybrid-linucb`; các scheduler khác bỏ qua. Truyền `--warm-start=0` hoặc `--load-penalty=0` cho phép chạy **ablation study** (tách riêng đóng góp của từng cơ chế) — factory cố tình không tự đổi 0 thành default.

### 2.2. Công thức điểm lai (Hybrid Q-value)

$$Q_{hybrid} = \underbrace{\hat{\theta}^\top x}_{\text{giá trị đã học}} + \underbrace{\alpha \sqrt{x^\top A^{-1} x}}_{\text{bonus khám phá}} - \underbrace{\lambda \cdot \text{LoadRatio}}_{\text{phạt quá tải}}$$

Điểm tinh tế: `LoadRatio` (= ActiveTasks/MaxParallel) **vốn đã là feature x[7]** trong vector đặc trưng — về lý thuyết, LinUCB có thể tự học trọng số âm cho nó. Vậy penalty để làm gì? Trả lời: penalty là **prior thủ công** cho giai đoạn mô hình chưa học xong. Do phản hồi trễ (reward về sau vài giây), trong cửa sổ đó AI không có dữ liệu để biết máy đang quá tải; λ·LoadRatio dùng thông tin thời-gian-thực (đếm task theo mili-giây, như LeastLoaded) để chặn hành vi dồn việc **trước khi** mô hình đủ dữ liệu tự nhận ra. Đây là luận điểm "lai tạo" trung tâm của thesis.

### 2.3. Warm-start với học ngầm (passive learning)

Trong `warmStartTasks` dispatch đầu:

1. Chọn worker theo **min ActiveTasks** trên tập candidates đã qua lọc admission (unhealthy, circuit-open, hết capacity) — mirror ngữ nghĩa `LeastLoadedScheduler.Select`, nhưng không instantiate scheduler đó.
2. **Vẫn tính feature vector** của worker được chọn và lưu vào `pendingX[TaskID]` → khi reward về, `RecordOutcome` áp dụng update Sherman-Morrison như bình thường. AI "quan sát và học" đúng như proposal mô tả — chỉ mất quyền quyết định, không mất dữ liệu học.
3. Sau N dispatch, AI tiếp quản với ma trận $A$, vector $b$ đã được "làm ấm" bằng ~N quan sát thật.

### 2.4. Feature vector 12 chiều

Layout đầy đủ (3 chiều mới in đậm):

| # | Feature | Chuẩn hóa | Nguồn |
|---|---|---|---|
| 0 | bias | = 1.0 | — |
| 1 | Kích thước source (preprocessed + raw) | log1p / log1p(4 MiB) | `TaskContext.SourceSizeBytes` |
| 2 | target là x86_64 | one-hot | request |
| 3 | target là ARM64 | one-hot | request |
| 4 | CPU cores của worker | /16, cap 1.0 | capabilities |
| 5 | RAM của worker | /64 GiB, cap 1.0 | capabilities |
| 6 | native arch khớp target | 1.0/0.0 | capabilities |
| 7 | LoadRatio (ActiveTasks/MaxParallel) | cap 1.0 | registry |
| 8 | RPC latency gần đây | /100ms, cap 1.0 | LatencyTracker |
| **9** | **Task là C++ (vs C)** | 1.0/0.0 | extension file hoặc compiler driver |
| **10** | **Kích thước raw source** | log1p / log1p(1 MiB) | `TaskContext.RawSourceSizeBytes` |
| **11** | **Success rate của worker (Laplace)** | (s+1)/(s+f+2) | registry counters |

Chi tiết ba chiều mới:

- **[9] Ngôn ngữ:** suy luận qua `isCppSource(filename, compiler)` — extension `.cpp/.cc/.cxx/.c++/.hpp/.hh` (và `.C` viết hoa, quy ước Unix) → C++; fallback: compiler driver chứa `"++"` (g++, clang++) vì theo ngữ nghĩa GCC, g++ compile cả file `.c` như C++. Denominator riêng 1 MiB cho [10] vì raw source (chưa expand header) nhỏ hơn preprocessed 10-100×; dùng chung denominator 4 MiB sẽ nén mọi giá trị về gần 0.
- **[11] Success rate — quyết định thiết kế quan trọng:** proposal gốc đề xuất rate thô với prior lạc quan 1.0. Khi review phát hiện: trên workload benchmark (CPython, 100% success) rate thô **≡ 1.0 = hằng số = trùng hoàn toàn với bias x[0]** — chính là lớp lỗi collinearity CRITICAL-2 mà dự án từng phải gỡ (one-hot build-type). Giải pháp: **Laplace smoothing** $(s+1)/(s+f+2)$ — worker chưa có dữ liệu nhận 0.5, giá trị tăng dần theo kinh nghiệm, không bao giờ là cột hằng. Ghi chú trung thực cho thesis: trên workload build sạch, feature này tiệm cận 1 và đóng góp ít; giá trị thật của nó nằm ở cluster thực tế có failure.

### 2.5. Đường ống dữ liệu ngữ cảnh (TaskContext)

`TaskContext` được mở rộng thêm 3 field: `RawSourceSizeBytes`, `SourceFilename`, `Compiler`. Phát hiện đáng chú ý: **dữ liệu này luôn có sẵn** trên `CompileRequest` tại đúng điểm điều phối (coordinator gRPC handler) — chỉ là chưa ai nối dây vào TaskContext. Không cần sửa proto, không cần sửa client.

---

## 3. Bug fix phát sinh (ngoài phạm vi proposal nhưng chặn benchmark)

### 3.1. File assembly bị phân tán → object hỏng, chỉ lộ khi link

**Triệu chứng:** build CPython fail tại bước link `Programs/_freeze_module` với `undefined reference to _Py_trampoline_func_start` — dù mọi task compile đều "thành công".

**Nguyên nhân gốc:** `ParsedArgs.IsDistributable()` (internal/compiler/parser.go) coi `.s/.S` là source phân tán được. Đường ống preprocess-rồi-gửi-remote làm hỏng object của file assembly (`asm_trampoline_x86_64.S`) — object trả về **không chứa symbol definitions**, và lỗi chỉ bùng phát sau đó rất xa, ở bước link cục bộ. Đây là dạng lỗi nguy hiểm nhất: sai ở A, nổ ở B.

**Fix:** assembly không được phân tán — compile cục bộ. Chi phí bằng 0 về hiệu năng: assemble không phải phần đắt của build (CPython chỉ có ~2 file .S trên ~300 file C).

### 3.2. Race điều phối → worker từ chối task → build fail ngẫu nhiên

**Triệu chứng:** `make -j5` lúc cold-start thỉnh thoảng fail ngay task đầu tiên với `rpc error: ResourceExhausted: too many concurrent tasks`. Có run sống 299 task rồi mới chết, có run chết ngay — hoàn toàn xác suất.

**Nguyên nhân gốc:** giữa lúc `Select` **đọc** `ActiveTasks` và lúc `IncrementTasks` **ghi** nó, các request song song có cửa sổ race. Burst 5 request đồng thời khi mọi worker đều rảnh → nhiều request cùng chọn một worker nhỏ (max-parallel=1) → worker từ chối thẳng (không xếp hàng) → coordinator dịch lời từ chối thành "compile failure" trả về client → make chết. Lưu ý: race này tồn tại với **mọi** scheduler từ trước (kể cả LeastLoaded); benchmark tháng 5 chỉ đơn giản là chưa gặp vận rủi. Với 40 run × ~300 task của benchmark thống kê, xác suất nổ giữa chừng là gần chắc chắn — bắt buộc phải fix.

**Fix:** vòng dispatch trong coordinator `Compile` handler retry tối đa **3 lần** khi (và chỉ khi) lỗi là `ResourceExhausted`: hoàn trả booking (`DecrementTasks` với success=false — lời từ chối là tín hiệu quá tải chính đáng cho thống kê worker), backoff 25ms×attempt để các booking đang race kịp ghi vào registry, rồi Select lại. Learner không nhận `RecordOutcome` cho attempt bị hủy — re-Select tự ghi đè `pendingX`. Ngữ nghĩa metrics (queue time, dashboard events, gauges) được giữ nguyên.

---

## 4. Hạ tầng benchmark thống kê (mới)

### 4.1. Vấn đề với phương pháp cũ

Số liệu tháng 5 chủ yếu là **single-run** (chính báo cáo tiến độ đã tự nhận đây là hạn chế). Noise floor ±25s trên nền chênh lệch vài chục giây khiến kết luận đơn lẻ không vững.

### 4.2. Các thành phần mới

| File | Vai trò |
|---|---|
| `scripts/benchmark_statistical.sh` | Wrapper chạy R×S build (mặc định 10 reps × 4 schedulers: leastloaded, p2c, linucb, hybrid-linucb) trên cụm 5w-hetero; xuất `results.csv` + per-run `tasks_*.jsonl`. Restart coordinator mỗi rep → arms reset → các mẫu độc lập thống kê |
| `scripts/analyze_benchmark.py` | pandas/scipy/matplotlib: bảng mean±std makespan, **Mann-Whitney U một phía** (hybrid < baseline?, α=0.05), boxplot, P99 per-run, và **warm-up curve** (rolling mean compile-time + Q-value 150 task đầu, linucb vs hybrid) |
| `test/stress/benchmark-heterogeneous.sh` | Nâng cấp thành *sourceable library* (`HG_BENCH_LIB=1`): tái sử dụng compose generators, `run_build`, `clone_cpython`… mà không chạy full 1w/3w/5w |

Chi tiết chống-sai-số trong wrapper (mỗi mục là một bug đã gặp thật trong smoke):

- **Truncate task-log giữa các rep** — volume `task-logs` sống qua `compose down`, logger mở O_APPEND; không truncate thì JSONL của run sau chứa cả run trước → P99/warm-up sai toàn bộ. (Trên Git Bash phải bọc `sh -c` vì MSYS tự đổi `/tmp/...` thành đường dẫn Windows.)
- **Fail loudly khi build lỗi** — build fail dở chừng kết thúc *nhanh hơn* build thành công; nếu vẫn ghi CSV sẽ làm đẹp số liệu một cách âm thầm. Wrapper grep `make: ***` trong build log và abort cả batch.
- **Guard `./configure`** — chỉ chạy khi Makefile chưa tồn tại (make clean không xóa Makefile), tiết kiệm ~1.5 phút × 40 run.
- **Regenerate compose file mỗi scheduler** — `SCHED_ARGS` được bake vào file lúc generate; dùng lại file cũ là chạy nhầm scheduler.

### 4.3. Phát hiện quan trọng: workload phải được pin version

Harness cũ clone CPython `main` **không pin** → workload trôi theo thời gian:

- Tháng 5/2026: main lúc đó ≈ 873 translation units (con số trong mọi bảng số liệu thesis).
- Tháng 7/2026: main = 3.16-dev, **build hỏng** dưới distributed compilation, và số TU đã khác.

Fix: pin `CPYTHON_VERSION=v3.14.0` (marker file trong volume tự phát hiện mismatch và reclone). Hệ quả cần ghi rõ trong thesis: **v3.14.0 cho ~295 task, không phải 873** — con số 873 không thể tái tạo vì gắn với một commit main vô danh. Điều thống kê cần là workload *cố định giữa các run*, và giờ đã có. Số liệu cũ và mới **không so sánh trực tiếp được** với nhau; benchmark mới tự sinh baseline của nó.

---

## 5. Kiểm thử

### 5.1. Unit test mới (internal/coordinator/scheduler/linucb_test.go)

| Test | Kiểm chứng |
|---|---|
| `TestLinUCB_WarmStartDelegatesToLeastLoadedAndLearns` | Warm-start chọn đúng worker ít tải nhất **và** arms vẫn được update (học ngầm) |
| `TestLinUCB_WarmStartHandsOverAfterN` | Ranh giới bàn giao chính xác: dispatch N theo least-loaded, dispatch N+1 theo UCB argmax |
| `TestLinUCB_LoadPenaltySteersAwayFromLoadedWorker` | Cô lập λ: hai worker giống hệt, chỉ khác tải → chọn máy rảnh |
| `TestIsCppSource` | Table test 9 case suy luận ngôn ngữ |
| `TestLinUCB_SuccessRatePrior` | Giá trị Laplace: 0/0→0.5, 3/1→0.667, 0/2→0.25 |
| `TestLinUCB_TotalDispatchesCountsFastPath` | Counter đếm dispatch volume, kể cả fast-path 1 candidate |

**Bẫy kỹ thuật cần biết khi viết test:** constructor `NewLinUCBScheduler` đổi `Alpha == 0` thành default 0.5. Test muốn tắt exploration phải truyền `Alpha: -1` (clamp về 0). Nếu quên: worker đang tải có ‖x‖ lớn hơn (LoadRatio là feature) → bonus lớn hơn → test penalty cho kết quả sai lệch mà vẫn có thể pass ngẫu nhiên.

### 5.2. Fix test cũ

- `TestLinUCB_LearnsBestArm` chứa **bug ngầm**: không set `TaskID` → `pendingX` không bao giờ được lưu → `RecordOutcome` bỏ mọi update → test "pass" suốt thời gian qua chỉ nhờ bonus ∝ ‖x‖ thiên vị worker nhiều core, chứ **không hề kiểm chứng việc học**. Đã fix bằng TaskID duy nhất mỗi trial.
- `TestLinUCB_FeatureVectorDimensions`: cập nhật cho layout d=12.
- `TestIsDistributable` (compiler): thêm 2 case assembly.
- Factory test: thêm case `hybrid-linucb` (và bổ sung `linucb`/`heft` vốn thiếu).

### 5.3. Kết quả smoke benchmark (1 rep, hybrid-linucb, v3.14.0)

- 295/295 task **success** — retry fix hoạt động, không còn ResourceExhausted.
- 100 record đầu: **99 `was_exploration=true`** (cửa sổ warm-start); record 101-200: chỉ còn 8 — bàn giao đúng tại N=100 trong môi trường thật.
- Load trải đều 5 worker trong warm-start (pattern least-loaded rõ ràng).
- Pipeline phân tích chạy trọn: summary, boxplot, P99. Makespan 42s/run trên máy hiện tại.

---

## 6. Hướng dẫn chạy đánh giá đầy đủ

```bash
# Full: 10 reps × 4 schedulers ≈ 40 build ≈ 1-1.5 giờ trên máy hiện tại
bash scripts/benchmark_statistical.sh

# Ablation / sweep (ví dụ):
WARM_START=35 bash scripts/benchmark_statistical.sh          # giữ tỉ lệ ~11% như proposal giả định
SCHEDULERS="linucb hybrid-linucb" REPS=5 bash scripts/benchmark_statistical.sh
```

**Tiêu chí "cải thiện" đã thống nhất:**

1. Makespan trung bình hybrid ≤ P2C, Mann-Whitney p<0.05 vs LinUCB cũ.
2. P99 compile-time per-run ≤ LinUCB cũ, cùng kiểm định.
3. Warm-up curve cho thấy hybrid né được spike compile-time giai đoạn ~100 task đầu của LinUCB.
4. Nếu (1) chỉ hòa P2C nhưng (2)+(3) đạt → claim đóng góp chuyển trọng tâm sang tail latency + cold-start (output của analyze script hỗ trợ trực tiếp).

**Caveat phải ghi trong thesis:**

- Warm-start N=100 chiếm **~34%** workload 295 task (proposal giả định 11% của 873) → cân nhắc sweep `WARM_START` ∈ {35, 100} và báo cáo cả hai.
- `was_exploration=true` trong cửa sổ warm-start là chủ đích (giúp phân tích offline tách cửa sổ) — exploration-rate tổng của hybrid vì thế cộng thêm ~34%; phân tích phải tách hai giai đoạn.
- Máy dev Windows không chạy được `go test -race` (thiếu gcc/cgo) — race detector dựa vào CI Linux.

---

## 7. Danh mục file thay đổi

| File | Thay đổi |
|---|---|
| `internal/coordinator/scheduler/feedback.go` | TaskContext +3 fields |
| `internal/coordinator/scheduler/linucb.go` | d=12, `isCppSource`, `loadRatio`, warm-start, load penalty, config mới |
| `internal/coordinator/scheduler/linucb_test.go` | +6 test mới, fix 2 test cũ |
| `internal/coordinator/server/grpc.go` | Populate TaskContext; Config + factory `hybrid-linucb`; **vòng dispatch retry** |
| `internal/coordinator/server/scheduler_factory_test.go` | +3 case factory |
| `internal/compiler/parser.go` (+test) | Assembly không phân tán |
| `cmd/hg-coord/main.go` | Flags `--warm-start`, `--load-penalty` + validation |
| `test/stress/benchmark-heterogeneous.sh` | `hybrid-linucb`, `compute_sched_args()`, sourceable, pin CPython v3.14.0 |
| `scripts/benchmark_statistical.sh` | **MỚI** — benchmark thống kê |
| `scripts/analyze_benchmark.py` | **MỚI** — phân tích + kiểm định |

**Nợ tài liệu (chưa sửa trong đợt này, cần cập nhật sau):** các chỗ trong `chuong-X-lap-lich.md` / `paper-skeleton.md` còn tham chiếu d=9 và công thức Q chưa có penalty term; doc comment `DispatchInfo` trong feedback.go mô tả Q của LinUCB chưa nhắc dạng hybrid.
