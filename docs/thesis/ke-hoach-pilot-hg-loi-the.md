# Kế hoạch pilot: kịch bản HG-LinUCB có lợi thế ("máy mạnh bị chiếm ngầm")

*Viết 2026-10-23. Người đọc: phiên Claude Code trên máy khác, được giao chạy pilot. Đọc trước
`docs/thesis/TIEN-DO-THUC-NGHIEM.md` (quy ước chung). Đây là **pilot thăm dò**, không phải thí nghiệm chính
thức: không đóng băng giả thuyết, không commit dữ liệu, không ghi kết quả vào báo cáo luận văn.*

## 1. Mục đích

TN1–TN2 cho thấy HG-LinUCB không thắng LeastLoaded (LL) khi cụm bão hoà, và hoà với SED ngay cả khi có drift.
Trên máy tham chiếu, chúng tôi đã tìm được một kịch bản mà HG-LinUCB thắng cả LL lẫn SED (pilot 3/3 lần). Pilot
này nhằm:

1. **Lặp lại** kịch bản đó trên phần cứng khác: lợi thế có giữ được không?
2. **Dò ranh giới**: lợi thế còn không khi đổi workload, khi máy chỉ chậm từng đợt, khi cụm bão hoà?

Nếu kịch bản đứng vững, nó sẽ được đăng ký trước và chạy chính thức **trên máy tham chiếu** (không phải máy này).

## 2. Kịch bản và tính thực tế

**Tình huống ngoài đời:** một cụm build dùng chung. Máy mạnh nhất, sau khi đã đăng ký với coordinator, bị một
tác vụ khác chiếm phần lớn CPU và giữ như vậy đến hết phiên. Tác vụ đó có thể là job CI khác, build của người
khác, indexer, quét antivirus, hoặc CPU steal trên VM dùng chung. Thông tin năng lực máy (CPU quảng cáo) được gửi
lúc đăng ký và không được cập nhật. Đây là hiện tượng "noisy neighbour" phổ biến.

**Ba điều kiện cùng lúc tạo ra lợi thế cho bộ học:**

| Điều kiện | Cách tạo trong benchmark | Vì sao cần |
|---|---|---|
| Cụm còn dư năng lực | `JOBS=3` (ninja -j3) trên 5 worker, 10 slot, tổng 4,0 core | Khi bão hoà, mọi máy luôn bận, xếp lịch nào cũng gần bằng nhau (tổng việc ÷ 4,0 core). Ở -j8 LL chỉ chậm hơn cận dưới vật lý 5–16%; ở -j3 là 84% |
| Hiệu năng thật lệch khỏi quảng cáo | `stress-ng --cpu 3` trong worker-5 (quảng cáo 1,1 CPU, mạnh nhất) | LL, SED, P90 chỉ thấy CPU quảng cáo và số task đang chạy, không có trí nhớ về hiệu năng |
| Lệch xuất hiện **sau** khi đăng ký và **kéo dài** | `DRIFT=permanent`, `DRIFT_AT=101` (bắt đầu ở task thứ 101 của build 1, kéo dài hết phiên) | Kéo dài thì việc học mới có lời. Xuất hiện sau đăng ký thì cấu hình tĩnh hay đo một lần lúc khởi động đều không cứu được |

**Vì sao các scheduler phản ứng khác nhau:**
- **LL** chọn máy ít task nhất. Khi cụm rảnh, nhiều máy cùng 0 task nên LL chọn gần như ngẫu nhiên, kể cả máy bị chiếm.
- **SED** chọn theo `(task đang chạy + 1) / CPU quảng cáo`. Khi worker-5 rảnh, `1/1,1` nhỏ nhất nên SED **ưu tiên** đúng máy bị chiếm.
- **P90** đưa file nặng sang máy quảng cáo mạnh nhất, tức đúng máy bị chiếm.
- **HG-LinUCB** có mô hình riêng cho từng worker, cập nhật bằng thời gian hoàn thành thật, nên ghi nhớ "worker-5 chậm" và né nó kể cả khi nó đang rảnh.

**Các giới hạn đã biết (không phải "thắng 100% mọi nơi"):**
- Cụm bão hoà (-j8): lợi thế gần như biến mất (TN1, TN2).
- Máy chậm **chớp nhoáng** (≤ 10 s) hoặc nhảy giữa nhiều máy: trí nhớ của HG trở thành bất lợi. TN2 E2-onoff 10 s/10 s ở -j8: HG so với SED +1,3 s, không có ý nghĩa.
- Máy chậm **ngay từ đầu và không đổi**: HG vẫn thắng, nhưng một luật "đo tốc độ mỗi máy một lần lúc khởi động rồi dùng SED" cũng xử lý được.
- Chưa so với đối thủ "chọn máy theo thời gian hoàn thành trung bình gần đây" (chưa có trong code). Đây là câu hỏi mở.

## 3. Bằng chứng hiện có (máy tham chiếu, 2 × Xeon E5-2686 v4, Docker 72 CPU / 31 GiB)

**Mô phỏng** (dựng từ log LeastLoaded của pilot TN3; xem `D:\bench-work\headroom` trên máy tham chiếu, ngoài
repo). "Bộ học hoàn hảo" = SED dùng tốc độ thật thay vì CPU quảng cáo:

| Workload, -j, số luồng chiếm | SED so với LL | Bộ học hoàn hảo so với LL | Hơn SED |
|---|---|---|---|
| absl, -j3, ×3 | +14,9% | +27,3% | +12,4 điểm (100% lần thử) |
| syn, -j3, ×3 | +12,8% | +27,3% | +14,4 điểm (100%) |
| absl/syn, -j4 | | | ~0–2 điểm |
| -j8, bất kỳ | | | ~0 |

**Pilot thật** (2026-10-23 00:23–01:38; absl, -j3, worker-5 bị chiếm 3 luồng từ task 101; 3 lần lặp; 0 lỗi).
Build 2, giây:

| Scheduler | Lần 1 | Lần 2 | Lần 3 | TB | So với LL | So với SED |
|---|---|---|---|---|---|---|
| hybrid-linucb-d-arm-g098 | 86,0 | 89,6 | 86,7 | 87,4 | −20,6% | −9,8% (3/3) |
| hybrid-linucb | 85,3 | 95,9 | 89,2 | 90,1 | −18,1% | −7,1% (3/3) |
| sed | 95,2 | 102,1 | 93,6 | 97,0 | −11,9% | — |
| linucb (thuần) | 96,1 | 111,0 | 106,2 | 104,4 | −5,2% | +7,7% |
| leastloaded | 106,1 | 110,7 | 113,5 | 110,1 | — | |
| p90 | 155,7 | 151,0 | 163,3 | 156,7 | +42,3% | |

Tỉ lệ task build 2 gửi vào worker-5: HG 3–4%, HG-D 2–9%, LinUCB thuần 2–3%, LL 12–16%, SED 14–15%, P90 16–18%.
Nghĩa là HG thắng đúng nhờ cơ chế dự kiến (né máy bị chiếm). LinUCB thuần cũng né được nhưng xếp phần còn lại
kém, nên warm-start và phạt tải của HG là thành phần quan trọng.

## 4. Chuẩn bị máy

1. **Code:** `git fetch origin && git checkout feat/exp3` (tối thiểu commit `00ef18b`). Không sửa code. Nếu cần
   sửa thì dừng lại, báo người dùng.
2. **Yêu cầu:**
   - Docker Engine hoặc Docker Desktop có `docker compose`, **≥ 8 CPU và ≥ 12 GB RAM cấp cho Docker**. Các
     container dùng khoảng 7 CPU theo quota: 5 worker 0,5–1,1, builder 2,0, coordinator 1,0. Nếu host ít core hơn
     thì cgroup quota không còn tương ứng với tốc độ thật, nên ghi rõ trong báo cáo.
   - Mạng Internet: runner clone abseil-cpp `20250512.1` từ GitHub bên trong container.
   - `bash`, `curl`, `git`, và `python3` có `pandas==2.3.3 scipy==1.17.1 numpy==2.4.6`. Xem
     `experiments/tooling/requirements.txt`; nên dùng venv.
3. **Windows + Git Bash:** đọc mục 4 của `TIEN-DO-THUC-NGHIEM.md`.
   - `python3` của Windows là bản giả, nên tạo venv và đưa một shim `python3` lên đầu `PATH`. Xem
     `experiments/tooling/bin/python3` và `env.sh`.
   - Runner đã tự xử lý việc Git Bash đổi đường dẫn `/...`.
   - Linux hoặc macOS không cần các bước này.
4. **Trạng thái máy trước mỗi lượt chạy (bắt buộc):**
   - Buộc tắt ứng dụng không cần thiết: trình duyệt, IDE nặng, Unity, game, chat. Giữ IDE chứa phiên Claude,
     `codex`, và Remote Desktop nếu có.
   - Power plan High performance, tắt sleep.
   - Không chạy thứ gì khác trong lúc pilot chạy.
   - Kiểm tra `docker ps` trống và CPU host rảnh (< 5%).
5. **Thư mục dữ liệu:** bên ngoài repo, ví dụ `~/bench-out/pilot-hidden` hoặc `D:\bench-work\out\pilot-hidden`.
   Không đặt trong repo.

## 5. Script chạy

Tạo file `pilot_hidden.sh` **bên ngoài repo**. Sửa `REPO` và `OUT_ROOT` cho máy này.

```bash
#!/bin/bash
# EXPLORATORY pilot (never official): hidden slowdown of the advertised-strongest worker.
# usage: pilot_hidden.sh <name> <workload> <jobs> <drift> <drift_cpu> <reps> [arms...]
#   drift: permanent | onoff | none     (onoff uses DRIFT_ON/DRIFT_OFF seconds from the environment)
REPO=${REPO:-"$HOME/hybridgrid"}            # repo checkout at feat/exp3
OUT_ROOT=${OUT_ROOT:-"$HOME/bench-out/pilot-hidden"}
cd "$REPO" || exit 1
name=$1 w=$2 j=$3 drift=$4 dc=$5 reps=$6; shift 6
arms=${*:-"leastloaded sed p90 linucb hybrid-linucb hybrid-linucb-d-arm-g098"}
thr=2356237; [[ $w == syn ]] && thr=150189          # registered TN3 P90 thresholds
P=$OUT_ROOT/$name
rm -rf "$P"; mkdir -p "$OUT_ROOT"
echo "$(date '+%F %T') START $name workload=$w jobs=$j drift=$drift drift_cpu=$dc reps=$reps arms=$arms"
env WORKLOAD=$w SYN_MEDIAN_UNITS=120 SYN_MAX_UNITS=4500 JOBS=$j CLIENTS=1 BUILDER_CPUS=2.0 BUILDER_MEM=2G \
  ARMS="$arms" P90_THRESHOLD=$thr REPS=$reps SESSION_BUILDS=2 REQUEST_TIMEOUT=600s \
  DRIFT=$drift DRIFT_AT=101 DRIFT_TARGET=worker-5 DRIFT_CPU=$dc \
  DRIFT_ON=${DRIFT_ON:-30} DRIFT_OFF=${DRIFT_OFF:-30} \
  OUT_DIR="$P" bash scripts/bench_exp.sh > "$P.log" 2>&1
rc=$?
echo "$(date '+%F %T') END $name rc=$rc"; tail -n 3 "$P.log"; tail -n +2 "$P/failures.csv" 2>/dev/null
```

Các tham số khác giữ mặc định của runner: `WARM_START=100`, `ALPHA=0.5`, `COOLDOWN=8`, `SEED_BASE=1000`.
`DRIFT_AT` phải ≥ 101; runner kiểm tra điều này và kiểm tra có `drift_on` hợp lệ.

## 6. Thứ tự pilot (thời gian ước tính theo máy tham chiếu)

| # | Tên | Lệnh | Ước tính | Mục đích |
|---|---|---|---|---|
| **P1** | `absl_j3_h3` | `pilot_hidden.sh absl_j3_h3 absl 3 permanent 3 3` | ~75 phút | **Bắt buộc.** Lặp lại đúng pilot của máy tham chiếu (6 scheduler) |
| P2 | `syn_j3_h3` | `pilot_hidden.sh syn_j3_h3 syn 3 permanent 3 3 leastloaded sed linucb hybrid-linucb hybrid-linucb-d-arm-g098` | ~2,5–3 giờ | Workload thứ hai (đuôi nặng tổng hợp) |
| P3 | `absl_j3_onoff30` | `DRIFT_ON=30 DRIFT_OFF=30 pilot_hidden.sh absl_j3_onoff30 absl 3 onoff 3 3 leastloaded sed linucb hybrid-linucb hybrid-linucb-d-arm-g098` | ~65 phút | Ranh giới: máy chậm từng đợt 30 s |
| P4 | `absl_j8_h3` | `pilot_hidden.sh absl_j8_h3 absl 8 permanent 3 3 leastloaded sed hybrid-linucb hybrid-linucb-d-arm-g098` | ~45 phút | Đối chứng bão hoà (dự kiến hoà) |

Làm P1 trước. Chỉ chạy tiếp P2–P4 nếu P1 hợp lệ (0 lỗi) **và** người dùng đồng ý. Mỗi lượt cần buộc tắt app lại
và kiểm tra máy rảnh trước khi chạy.

## 7. Phân tích

```python
# usage: python3 summarize.py <pilot dir>
import csv, json, glob, sys, collections, statistics as st
P = sys.argv[1]
rows = [r for r in csv.DictReader(open(f'{P}/results.csv')) if r['build_idx'] == '2']
by = collections.defaultdict(dict)
for r in rows: by[r['arm']][int(r['round'])] = float(r['elapsed_s'])
R = sorted({k for d in by.values() for k in d})
ll, sed = by['leastloaded'], by.get('sed', {})
print('build 2 seconds; rounds', R)
for a, d in sorted(by.items(), key=lambda kv: st.mean(kv[1].values())):
    v = [d[k] for k in R]
    vl = [(d[k] - ll[k]) / ll[k] * 100 for k in R]
    vs = [(d[k] - sed[k]) / sed[k] * 100 for k in R] if sed else []
    print(f'{a:26} {v} mean={st.mean(v):6.1f} vsLL={[round(x,1) for x in vl]} vsSED={[round(x,1) for x in vs]}')
print('\nshare of build-2 tasks on worker-5 per round:')
for f in sorted(glob.glob(f'{P}/tasks_*_round*.jsonl')):
    arm, rnd = f.split('tasks_')[1].rsplit('_round', 1)
    c = collections.Counter()
    for l in open(f):
        d = json.loads(l)
        if d.get('event') == 'task_completed' and '-b2-' in d.get('build_id', ''):
            c[d['worker_address'].split(':')[0]] += 1
    n = sum(c.values())
    if n: print(f'{arm:26} round {rnd.split(".")[0]}: w5 {100*c["worker-5"]/n:4.1f}%  {dict(sorted(c.items()))}')
```

## 8. Tiêu chí đánh giá (chốt trước khi chạy)

P1 được coi là **lặp lại thành công** khi đạt đủ cả bốn:
1. Hợp lệ: `failures.csv` không có dòng nào, runner rc = 0.
2. `hybrid-linucb` **hoặc** `hybrid-linucb-d-arm-g098` nhanh hơn `leastloaded` **và** `sed` ở **cả 3/3 lần**
   (build 2, so từng lần lặp).
3. Mức hơn SED trung bình ≥ 5%.
4. Cơ chế: tỉ lệ task vào worker-5 của biến thể HG đó ≤ một nửa của SED.

Với P2–P4 chỉ mô tả kết quả, không áp tiêu chí. Dự kiến: P2 tương tự P1, P3 không chắc, P4 hoà.

## 9. Quy ước

- Không commit dữ liệu pilot, không sửa code, không sửa `hypotheses*.md`, không chạy thí nghiệm chính thức.
- Không chạy lại một lượt để "lấy kết quả đẹp hơn". Nếu một lượt hỏng (lỗi build, máy bị can thiệp), ghi lý do và
  chạy lại **một** lần, báo cáo cả hai.
- Có thể nhờ Codex CLI review phân tích (reviewer `gpt-5.6-sol`; không dùng "astra"), không bắt buộc.

## 10. Báo cáo gửi lại người dùng

Một file Markdown (ngoài repo), gồm:
- Máy: CPU (model, số core), RAM, OS, phiên bản Docker, CPU/RAM cấp cho Docker, commit đã chạy.
- Với mỗi pilot đã chạy: thời gian bắt đầu và kết thúc, rc, `failures.csv`, đầu ra của `summarize.py`.
- Kết luận P1 theo đủ 4 tiêu chí ở mục 8, ghi rõ đạt hay không đạt từng tiêu chí.
- Bất thường: lỗi, timeout, ứng dụng nền, nhiệt độ, v.v.
- Nén thư mục pilot (`tar -cJf <tên>.tar.xz <thư mục>`) để người dùng chuyển về máy tham chiếu nếu cần.
