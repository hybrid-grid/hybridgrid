# Onboarding Guide — Hybrid-Grid Build

> Tài liệu dành cho thành viên mới tham gia nhóm đồ án.
> Đọc từ đầu đến cuối, theo thứ tự. Mỗi phần xây dựng trên phần trước.
> Last updated: 2026-04-13

---

## Mục lục

1. [Dự án này là gì?](#1-dự-án-này-là-gì)
2. [Bài toán gốc](#2-bài-toán-gốc)
3. [Kiến trúc hệ thống](#3-kiến-trúc-hệ-thống)
4. [Stack công nghệ](#4-stack-công-nghệ)
5. [Cấu trúc mã nguồn](#5-cấu-trúc-mã-nguồn)
6. [Trạng thái hiện tại — đã làm được gì](#6-trạng-thái-hiện-tại)
7. [Bài toán nghiên cứu — Scheduling](#7-bài-toán-nghiên-cứu)
8. [Hướng giải quyết — Reinforcement Learning](#8-hướng-giải-quyết)
9. [Roadmap tương lai](#9-roadmap-tương-lai)
10. [Setup môi trường phát triển](#10-setup-môi-trường)
11. [Quy trình làm việc](#11-quy-trình-làm-việc)
12. [Tài liệu tham khảo](#12-tài-liệu-tham-khảo)

---

## 1. Dự án này là gì?

**Hybrid-Grid Build** là một hệ thống **build phân tán** — tức là thay vì dùng 1 máy tính để biên dịch (compile) code, chúng ta chia công việc ra cho nhiều máy cùng làm song song, giúp build nhanh hơn.

Ví dụ thực tế: Một dự án C++ lớn (như CPython, Redis) có 500 file `.c`. Compile trên 1 máy mất 10 phút. Với Hybrid-Grid, 5 máy cùng compile → mỗi máy xử lý ~100 file → xong trong ~2 phút.

### Tại sao tên "Hybrid-Grid"?

- **Grid**: lưới máy tính cùng làm việc (giống grid computing)
- **Hybrid**: các máy trong lưới **không giống nhau** — có máy mạnh (desktop 16 cores), máy yếu (Raspberry Pi 4 cores), máy kiến trúc khác nhau (x86 vs ARM). Scheduler phải thông minh để phân chia công việc hợp lý.

### So sánh với giải pháp hiện có

| Hệ thống | Hỗ trợ | Scheduling | Học từ dữ liệu |
|---|---|---|---|
| **distcc** | C/C++ only | Round-robin (ngẫu nhiên) | Không |
| **icecream** | C/C++ only | Capability filter | Không |
| **sccache** | C/C++, Rust (cache only) | Không schedule | Không |
| **Bazel RBE** | Bazel projects only | Queue-based | Không |
| **Hybrid-Grid** | C/C++, Flutter, Unity, (sắp tới: Rust, Go) | **RL-based** (đang phát triển) | **Có** |

Điểm khác biệt chính: Hybrid-Grid sẽ dùng **Reinforcement Learning** để học cách phân chia task tối ưu. Đây là **đóng góp nghiên cứu chính** của đồ án.

---

## 2. Bài toán gốc

**Input**: Lập trình viên gõ `hgbuild make -j8` (giống `make` bình thường, nhưng thêm prefix `hgbuild`).

**Output**: Giống hệt `make` — file `.o` được tạo ra, project được build.

**Khác biệt**: Thay vì compile trên máy local, các file `.c/.cpp` được gửi đến **workers** (máy khác trong mạng) để compile song song.

### Luồng hoạt động (flow) đơn giản

```
Lập trình viên                    Coordinator                     Workers
     │                                │                              │
     │  hgbuild make -j8              │                              │
     │──────────────────────────────► │                              │
     │                                │  main.c → worker-1           │
     │                                │────────────────────────────► │ compile
     │                                │  utils.c → worker-2          │
     │                                │────────────────────────────► │ compile
     │                                │  config.c → worker-3         │
     │                                │────────────────────────────► │ compile
     │                                │                              │
     │                                │ ◄──── main.o (kết quả)      │
     │                                │ ◄──── utils.o               │
     │                                │ ◄──── config.o              │
     │ ◄──── tất cả .o files         │                              │
     │                                │                              │
     │  link thành binary             │                              │
```

---

## 3. Kiến trúc hệ thống

Hệ thống gồm **3 thành phần** (3 chương trình riêng biệt):

### 3.1. `hgbuild` (CLI Tool)

- Chạy trên **máy lập trình viên**
- Thay thế `gcc`/`g++`/`make`/`ninja`/`flutter` — lập trình viên dùng như bình thường
- Khi nhận lệnh compile, nó:
  1. Tiền xử lý file (preprocessor: `gcc -E`) trên máy local
  2. Kiểm tra cache — nếu đã compile rồi → trả kết quả ngay, không gửi đi
  3. Gửi source code tới **coordinator** qua gRPC
  4. Nhận kết quả (file `.o`) và ghi ra disk

### 3.2. `hg-coord` (Coordinator)

- **Bộ não trung tâm** — chạy trên 1 máy
- Nhiệm vụ:
  1. Quản lý danh sách workers (ai đang online, ai đang bận)
  2. **Quyết định giao task cho worker nào** ← đây là phần scheduling, phần ta sẽ cải tiến bằng RL
  3. Chuyển tiếp request tới worker được chọn
  4. Quản lý cache kết quả
  5. Cung cấp dashboard web (http://localhost:8080)
  6. Expose metrics cho Prometheus (http://localhost:8080/metrics)

### 3.3. `hg-worker` (Worker)

- **Máy thực thi** — chạy trên mỗi máy trong lưới
- Khi khởi động:
  1. Tự động tìm coordinator qua mDNS (zero-config) hoặc chỉ định thủ công
  2. Đăng ký bản thân: "tôi có 8 cores, 16GB RAM, hỗ trợ gcc/clang, kiến trúc x86_64"
  3. Gửi heartbeat định kỳ (mỗi 30s) để coordinator biết mình còn sống
- Khi nhận task:
  1. Compile source code bằng compiler trên máy mình
  2. Trả kết quả (.o file) về coordinator

### Giao tiếp giữa các thành phần

```
hgbuild ◄──── gRPC ────► hg-coord ◄──── gRPC ────► hg-worker
  (CLI)                 (Coordinator)               (Worker)
                             │
                        HTTP :8080
                             │
                        Dashboard / Metrics / Health
```

Tất cả giao tiếp dùng **gRPC** (Google RPC) — nhanh hơn REST API, dùng protobuf để serialize dữ liệu.

---

## 4. Stack công nghệ

### Ngôn ngữ: Go (Golang)

Toàn bộ project viết bằng Go. Nếu chưa biết Go, đọc theo thứ tự:

1. **Tour of Go** (2-3 giờ): https://go.dev/tour — bắt buộc
2. **Effective Go** (2 giờ): https://go.dev/doc/effective_go — conventions
3. **Go by Example** (tham khảo): https://gobyexample.com — tra cứu cú pháp

Go concepts cần biết cho project này:
- Goroutines + channels (concurrency)
- Interfaces (scheduler, executor đều dùng interface)
- Structs + methods (Go không có class)
- `sync.RWMutex` (concurrent access)
- Testing: `go test -race ./...`

### Công nghệ khác

| Công nghệ | Vai trò | Cần biết |
|---|---|---|
| **gRPC + Protobuf** | Giao tiếp giữa 3 components | Đọc `proto/hybridgrid/v1/build.proto` |
| **Docker** | Chạy workers trong container, cross-compile | `docker compose up` |
| **Prometheus** | Thu thập metrics (số task, thời gian, cache hit) | Không cần biết sâu |
| **OpenTelemetry** | Distributed tracing | Không cần biết sâu |
| **mDNS** | Auto-discovery workers trên LAN | Dùng thư viện, không cần hiểu protocol |

---

## 5. Cấu trúc mã nguồn

```
hybridgrid/
├── cmd/                          # Entry points (hàm main)
│   ├── hg-coord/main.go          #   Coordinator binary
│   ├── hg-worker/main.go         #   Worker binary
│   └── hgbuild/main.go           #   CLI binary
│
├── proto/hybridgrid/v1/          # Protobuf definitions
│   └── build.proto               #   Tất cả messages + services
│
├── gen/go/hybridgrid/v1/         # Code generated từ proto (KHÔNG sửa)
│   └── build.pb.go
│
├── internal/                     # Core logic (phần quan trọng nhất)
│   ├── coordinator/
│   │   ├── scheduler/            # ★ SCHEDULER — phần ta sẽ thêm RL
│   │   │   ├── scheduler.go      #   Interface + 3 implementations (Simple, LeastLoaded, P2C)
│   │   │   └── scheduler_test.go
│   │   ├── server/
│   │   │   ├── grpc.go           #   Coordinator gRPC server (Compile, Build, Register)
│   │   │   └── stats.go          #   Statistics tracking
│   │   ├── registry/
│   │   │   └── registry.go       #   Worker registry (who's online, capabilities)
│   │   ├── metrics/
│   │   │   ├── latency.go        #   Per-worker latency tracking (EWMA)
│   │   │   └── ewma.go           #   Exponential Weighted Moving Average
│   │   └── resilience/
│   │       └── circuit.go        #   Circuit breaker (fault tolerance)
│   │
│   ├── worker/
│   │   ├── executor/
│   │   │   ├── executor.go       #   Executor interface
│   │   │   ├── native.go         #   C/C++ executor (chạy gcc/clang)
│   │   │   ├── flutter.go        #   Flutter executor
│   │   │   ├── unity.go          #   Unity executor
│   │   │   └── docker.go         #   Docker cross-compile executor
│   │   └── server/
│   │       └── grpc.go           #   Worker gRPC server
│   │
│   ├── cache/
│   │   └── key.go                #   Cache key generation (xxhash)
│   │
│   ├── capability/
│   │   └── detect.go             #   Detect compilers, Flutter SDK, Unity, etc.
│   │
│   └── cli/                      #   CLI command packages
│       ├── flutter/command.go
│       └── unity/command.go
│
├── test/
│   ├── stress/                   #   Benchmark scripts (Docker-based)
│   │   ├── benchmark.sh          #   Scaling test
│   │   ├── benchmark-fair.sh     #   Equal resources test
│   │   └── benchmark-heterogeneous.sh  # Heterogeneous test
│   └── e2e/                      #   End-to-end tests
│
├── docs/
│   ├── thesis/                   #   Tài liệu luận văn
│   │   ├── bai-toan-lap-lich.md  #   Báo cáo bài toán scheduling (đã gửi GVHD)
│   │   └── annotated-bibliography.md  # 9 papers annotated
│   ├── BENCHMARK_REPORT.md       #   Benchmark report v0.2 (Jan 2026)
│   └── BENCHMARK_REPORT_v0.5.md  #   Benchmark report v0.5 (Apr 2026, mới nhất)
│
├── .sisyphus/                    #   Session management (plans, evidence, notes)
│   └── plans/
│       └── cost-scheduler.md     #   ★ Plan chi tiết cho scheduler mới
│
└── go.mod                        #   Go dependencies
```

### File quan trọng nhất cần đọc (theo thứ tự)

1. `proto/hybridgrid/v1/build.proto` — hiểu data model
2. `internal/coordinator/scheduler/scheduler.go` — scheduler hiện tại
3. `internal/coordinator/server/grpc.go` — luồng xử lý request
4. `internal/coordinator/registry/registry.go` — worker management
5. `internal/worker/executor/native.go` — cách worker compile code
6. `.sisyphus/plans/cost-scheduler.md` — plan scheduler mới (đang chuyển sang RL)

---

## 6. Trạng thái hiện tại

### Phiên bản đã release

| Version | Ngày | Nội dung chính |
|---|---|---|
| v0.1-v0.2 | Jan-Mar 2026 | Foundation: gRPC, mDNS discovery, C/C++ compilation, cache, circuit breaker |
| v0.2.3-v0.2.4 | Mar 2026 | TLS/mTLS, config validation, E2E verification |
| v0.3.0 | Mar 2026 | Observability: 12 Prometheus metrics, OpenTelemetry tracing, `/health` endpoints |
| v0.4.0 | Mar 2026 | Flutter Android distributed builds |
| v0.5.0 | Apr 2026 | Unity distributed builds (committed, chưa release) |

### Tính năng production-ready

- C/C++ distributed compilation (100+ file stress test OK)
- Flutter Android build (APK/AAB)
- Unity batch mode build (Android/iOS/Windows/Linux/macOS/WebGL)
- Content-addressable cache (xxhash)
- mDNS auto-discovery
- P2C scheduler (weighted heuristic)
- Circuit breaker + local fallback
- TLS/mTLS
- Prometheus + OpenTelemetry
- Web dashboard
- Docker cross-compile

### Benchmark hiện tại (April 2026, P2C scheduler)

| Scenario | Workers | Speedup |
|---|---|---|
| Scaling (thêm máy) | 5 | **5.01x** |
| Equal resources (cùng tổng CPU) | 3 (best) | **1.14x** |
| Heterogeneous (máy mạnh yếu lẫn lộn) | 5 | **1.23x** |

Con số 1.23x trên heterogeneous cluster là **baseline** mà scheduler mới (RL) cần beat.

---

## 7. Bài toán nghiên cứu

### Phát biểu bài toán

> Khi một task build đến coordinator, **nên giao cho worker nào?**

Đây là bài toán **online scheduling trên unrelated heterogeneous machines** — mỗi task có thời gian thực thi khác nhau trên mỗi worker (vì hardware khác nhau).

### Scheduler hiện tại và 5 hạn chế

Scheduler hiện tại dùng **P2C (Power of Two Choices)**: chọn 2 workers ngẫu nhiên, tính điểm mỗi worker bằng công thức cố định, giao cho worker điểm cao hơn.

```
score(w) = 50×(arch match) + 10×(CPU cores) + 5×(RAM GB)
         - 15×(active tasks) - 0.5×(latency ms) + 20×(LAN bonus)
```

**5 hạn chế:**
1. **Weights cố định** — 50, 10, 5, -15, -0.5, 20 là hằng số trong code, không có cơ sở tối ưu
2. **Không phân biệt loại task** — compile `helloworld.c` (0.2s) và `boost/spirit.hpp` (45s) bị coi như nhau
3. **ActiveTasks không phản ánh load thật** — "3 task đang chạy" có thể là 3 task nhỏ sắp xong hoặc 3 task to còn lâu
4. **Không học từ dữ liệu** — compile xong rồi nhưng scheduler không nhớ worker nào nhanh hơn
5. **Phạt cross-compile là 0/1** — thực tế Docker cross-compile chậm 1.5-5x tùy task, không phải binary

### Tại sao RL?

Reinforcement Learning phù hợp vì:
- **Sequential decisions**: quyết định giao task A cho worker 1 ảnh hưởng worker 1 bận hơn → ảnh hưởng quyết định tiếp theo
- **Reward rõ ràng**: task xong → biết completion time → reward = -completion_time
- **Adapt được**: worker mới join, workload thay đổi → model tự điều chỉnh
- **Exploration tự nhiên**: RL tự balance giữa "thử worker mới" (exploration) và "dùng worker tốt nhất đã biết" (exploitation)

Đọc thêm: `docs/thesis/bai-toan-lap-lich.md` (báo cáo đã gửi GVHD)

---

## 8. Hướng giải quyết

### Tổng quan: Q-Learning với Linear Function Approximation

Thay vì công thức tính điểm cố định (P2C), ta dùng một **model học được** để dự đoán thời gian hoàn thành:

```
P2C (hiện tại):    score = hardcoded_formula(worker_features)
RL (sắp tới):      Q(s,a) = θᵀ × φ(state, action)     ← θ được cập nhật liên tục
                    chọn worker có Q thấp nhất (= predicted cost thấp nhất)
```

### Q-Learning là gì? (giải thích đơn giản)

Hãy tưởng tượng bạn đi ăn trưa mỗi ngày và phải chọn 1 trong 5 quán. Mỗi quán có chất lượng khác nhau, thay đổi theo thời gian.

- **Q-value**: Ước lượng "quán này ngon cỡ nào" — `Q(quán_A) = 7.5/10`
- **Action**: Chọn quán nào hôm nay
- **Reward**: Sau khi ăn, bạn cho điểm (1-10)
- **Update**: Cập nhật ước lượng dựa trên trải nghiệm thực tế:
  ```
  Q_mới = Q_cũ + α × (reward_thực_tế - Q_cũ)
  ```
  α = learning rate (0 < α < 1), tốc độ cập nhật

Áp dụng vào scheduling:
- **Q-value**: Ước lượng "task này trên worker đó mất bao lâu"
- **Action**: Chọn worker nào
- **Reward**: Sau khi compile xong, biết thời gian thực → cập nhật model
- **Exploration**: Thỉnh thoảng thử worker khác (ε-greedy) để khám phá

### Linear Function Approximation

Thay vì lưu 1 Q-value cho **mỗi cặp (task, worker)** (quá nhiều), ta dùng **1 công thức tuyến tính** áp dụng cho tất cả:

```
Q(state, worker) = θ₁ × source_size
                 + θ₂ × worker_cpu_cores
                 + θ₃ × worker_active_tasks
                 + θ₄ × is_native_arch
                 + ...
                 + θ₁₈ × bias_term
```

`θ₁..θ₁₈` là **weights** — bắt đầu từ 0, được cập nhật bằng TD (Temporal Difference) learning mỗi khi có task hoàn thành.

### Features (18 đầu vào)

**Task features** (mô tả task cần compile):
- Source size (KB, normalized)
- Build type (C++ / Flutter / Unity)
- Size bucket (Small / Medium / Large / XLarge)
- Target architecture (x86_64 / arm64)

**Worker features** (mô tả worker đang xét):
- CPU cores (normalized)
- Memory GB (normalized)
- Load ratio (active_tasks / max_parallel, 0.0 - 1.0)
- Native arch match (0 or 1)
- Docker available (0 or 1)
- LAN connection (0 or 1)
- Network latency (ms, normalized)
- Historical avg compile time (ms, normalized)

**Interaction features** (kết hợp task + worker):
- source_size × cpu_cores (task lớn + worker mạnh = tốt)
- is_large_task × is_weak_worker (task lớn + worker yếu = xấu)
- Cluster total active tasks (tải tổng thể)

**Bias term**: 1.0 (constant)

### TD Update (sau mỗi task)

```
φ = buildFeatures(task, worker, clusterState)   // vector 18 features
Q_predicted = θᵀ × φ                             // dự đoán
actual_cost = completion_time_ms                  // thực tế
td_error = actual_cost - Q_predicted              // sai lệch
θ = θ + α × td_error × φ                         // cập nhật weights
```

`α` = 0.01 (learning rate)
`ε` = 0.1 → 0.05 (exploration rate, giảm dần)

### Luồng tổng hợp

```
1. Task mới đến coordinator
2. Lấy danh sách workers capable
3. Với mỗi worker:
   a. Xây dựng feature vector φ(task, worker)
   b. Tính Q = θᵀ × φ (predicted cost)
4. Chọn worker min Q (hoặc random nếu ε-greedy trigger)
5. Gửi task đến worker
6. Worker compile, trả kết quả + compilation_time_ms
7. Coordinator cập nhật θ (TD update)
8. Lặp lại từ bước 1
```

### Plan chi tiết

Đọc: `.sisyphus/plans/cost-scheduler.md` — plan đầy đủ với interface design, file list, phases, justifications. (Đang update sang RL version.)

---

## 9. Roadmap tương lai

### Phase 1: Foundation (Week 1-2)
- Task classification (phân loại task theo size bucket)
- Q-Model (linear function approximation, TD update)
- Interface mở rộng (CostAwareScheduler)

### Phase 2: Core RL Scheduler (Week 2-3)
- RLScheduler implementation (feature builder, selection, exploration)
- Unit tests (convergence, exploration, backward compat)

### Phase 3: Integration (Week 3-4)
- Wire vào coordinator (CLI flag `--scheduler=rl`)
- Feedback loop (compilation time → TD update)
- Full test suite pass

### Phase 4: Evaluation (Week 5-8)
- Benchmark matrix: 4 schedulers × 3 cluster configs
- Learning curve analysis
- Ablation studies (feature importance, hyperparameter sensitivity)
- Straggler injection test

### Phase 5: Thesis (Week 9-16)
- Viết luận văn (~50-70 trang)
- Figures, tables, diagrams
- Mock defense

### Có thể làm song song (khi RL scheduler đã ổn)
- Thêm Rust/Go build support (mở rộng platform)
- WAN registry (workers ngoài LAN)

---

## 10. Setup môi trường

### Prerequisites

```bash
# Go 1.24+
go version   # expect: go1.24.x

# Docker + Docker Compose
docker --version          # expect: 28.x
docker compose version    # expect: v2.x

# Git
git --version

# protoc (nếu cần sửa proto)
protoc --version
```

### Clone và build

```bash
# Clone
git clone https://github.com/hybrid-grid/hybridgrid.git
cd hybridgrid

# Build tất cả binaries
go build -o bin/ ./cmd/...
# Kết quả: bin/hg-coord, bin/hg-worker, bin/hgbuild

# Chạy tests
go test -race ./...

# Chạy lint (nếu có)
golangci-lint run
```

### Chạy thử local (3 terminal)

```bash
# Terminal 1: Coordinator
./bin/hg-coord serve --grpc-port=9000 --http-port=8080

# Terminal 2: Worker
./bin/hg-worker serve --coordinator=localhost:9000

# Terminal 3: Build test project
cd test/e2e/testdata
../../../bin/hgbuild make -j4
```

### Chạy bằng Docker Compose (đơn giản hơn)

```bash
docker compose up -d          # start coordinator + 2 workers
open http://localhost:8080     # dashboard
docker compose down            # stop
```

### Chạy benchmark

```bash
cd test/stress
./benchmark.sh                 # scaling test (~20 min)
./benchmark-fair.sh            # equal resources (~20 min)
./benchmark-heterogeneous.sh   # heterogeneous (~20 min)
cat /tmp/benchmark_results.txt # kết quả
```

---

## 11. Quy trình làm việc

### Git workflow
- Branch chính: `main`
- Commit message format: `type(scope): description`
  - `feat(scheduler): add Q-learning model`
  - `fix(grpc): handle nil worker response`
  - `test(scheduler): add convergence test`
  - `docs: update onboarding guide`

### Code conventions
- Go standard formatting (`gofmt`)
- Tests cùng package: `xxx_test.go`
- Interfaces nhỏ, concrete types lớn
- Error handling: return error, không panic
- Concurrency: dùng `sync.RWMutex` cho shared state

### Trước khi commit

```bash
go build ./...               # build OK?
go vet ./...                 # static analysis OK?
go test -race ./...          # tests pass? no race conditions?
```

### Tools hỗ trợ

- **Claude Code** (CLI): dùng để hỏi về codebase, generate code, review
- **Codex**: review plans, iterate
- Dashboard: http://localhost:8080 khi coordinator chạy

---

## 12. Tài liệu tham khảo

### Đọc bắt buộc (theo thứ tự)

1. **Go Tour**: https://go.dev/tour — học Go cơ bản (nếu chưa biết)
2. **gRPC Go Tutorial**: https://grpc.io/docs/languages/go/basics/ — hiểu gRPC
3. `docs/thesis/bai-toan-lap-lich.md` — bài toán scheduling (đã gửi GVHD)
4. `docs/thesis/annotated-bibliography.md` — 9 papers quan trọng
5. `.sisyphus/plans/cost-scheduler.md` — plan chi tiết scheduler mới

### Papers (đọc khi có thời gian)

| Paper | Tại sao đọc |
|---|---|
| Mitzenmacher 2001 — "Power of Two Choices" | Nền tảng P2C (scheduler hiện tại) |
| Ousterhout 2013 — "Sparrow" (SOSP) | P2C trong data center, gần bài toán của ta |
| Lenstra-Shmoys-Tardos 1990 | Lý thuyết scheduling trên unrelated machines |
| Sutton & Barto — "RL: An Introduction" Ch.6,9 | TD learning + function approximation |
| Watkins 1989 — "Q-learning" | Thuật toán Q-learning gốc |

### Tài liệu nội bộ

| File | Nội dung |
|---|---|
| `README.md` | Quick start, feature list, usage guide |
| `CLAUDE.md` | Rules cho Claude Code (workflow, conventions) |
| `docs/BENCHMARK_REPORT_v0.5.md` | Benchmark mới nhất (Apr 2026) |
| `.sisyphus/plans/` | Tất cả plans (v0.3 → v1.0) |
| `docs/system-architecture.md` | Kiến trúc chi tiết |

---

## Checklist cho member mới

- [ ] Đọc xong Go Tour (https://go.dev/tour)
- [ ] Clone repo, build thành công (`go build ./cmd/...`)
- [ ] Chạy tests thành công (`go test -race ./...`)
- [ ] Chạy coordinator + worker local, build test project
- [ ] Đọc `proto/hybridgrid/v1/build.proto`
- [ ] Đọc `internal/coordinator/scheduler/scheduler.go` — hiểu 3 scheduler hiện tại
- [ ] Đọc `internal/coordinator/server/grpc.go` — hiểu Compile() flow
- [ ] Đọc `docs/thesis/bai-toan-lap-lich.md` — hiểu bài toán
- [ ] Đọc `.sisyphus/plans/cost-scheduler.md` — hiểu plan scheduler mới
- [ ] Chạy benchmark (`test/stress/benchmark.sh`) — thấy hệ thống hoạt động
- [ ] Đọc Sutton & Barto Chapter 6 (TD Learning) — hiểu RL foundation
