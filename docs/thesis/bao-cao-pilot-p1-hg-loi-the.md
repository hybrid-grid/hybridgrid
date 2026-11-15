# Báo cáo pilot P1: "máy mạnh bị chiếm ngầm" (absl, -j3)

*Pilot thăm dò, không phải thí nghiệm chính thức. 2026-10-23. Chạy trên máy phát triển của Kien, không phải máy tham chiếu.*

## 1. Máy và cấu hình

| Mục | Giá trị |
|---|---|
| CPU | Intel Core i5-14600KF, 14 core / 20 luồng (lai P-core + E-core) |
| RAM | 47,8 GB |
| OS | Windows 11 Pro 10.0.26200 |
| Docker | Docker Desktop, Engine 29.3.1, Compose 5.1.0 |
| Cấp cho Docker | 20 CPU, 23,4 GiB RAM |
| Code | `feat/exp3` @ `6e0fd9244745fff7aa2e0b3fa0dfac2deac717db`, `git_dirty=false` (theo `meta.json`) |
| Bản chạy | clone LF riêng `C:\bench-work\hybridgrid` (`core.autocrlf=false`), vì checkout chính có CRLF |
| Dữ liệu | `C:\bench-work\out\pilot-hidden\` (ngoài repo) |
| Script | `pilot_hidden.sh`, `summarize.py` đúng như mục 5 và 7 của kế hoạch; chỉ đổi `REPO` và `OUT_ROOT` |
| Python | 3.11.9, pandas 2.3.3, scipy 1.17.1, numpy 2.4.6 (đúng bản yêu cầu) |

Trạng thái máy trước khi chạy:
- Power plan High performance (trước đó Balanced), sleep và hibernate trên AC đều là "never".
- Windows Update đã pause đến 2026-11-27 (do Kien thực hiện, phiên Claude không có quyền admin).
- Đã tắt Edge, Fork, Riot Client, PhoneExperienceHost. Giữ nguyên Antigravity IDE (kèm java và codex), Docker Desktop, Chrome Remote Desktop, TeamViewer, UltraViewer và Radmin VPN.
- `docker ps` trống. CPU host rảnh trung bình 2,8% (đỉnh 6,1% ở một mẫu trong 24 s).
- Preflight không STRICT: đạt, trừ 2 mục báo FAIL do đường dẫn `D:\bench-work` không tồn tại trên máy này (shim python3 và ổ D:). Kết quả lưu ở `preflight.txt`.

**Khác biệt với máy tham chiếu** (2 × Xeon E5-2686 v4, Docker 72 CPU / 31 GiB):
- CPU nhanh hơn nhiều và là CPU lai, nên quota cgroup của container (tổng khoảng 7 CPU) không còn tương ứng tốc độ thật.
- Docker có 20 CPU và 23,4 GiB.

## 2. P1 `absl_j3_h3`

| Mục | Giá trị |
|---|---|
| Lệnh | `pilot_hidden.sh absl_j3_h3 absl 3 permanent 3 3` |
| Bắt đầu / kết thúc | 2026-10-23 09:24:50 / 09:52:08 (27 phút 18 giây; kế hoạch ước tính ~75 phút, máy này nhanh hơn) |
| rc | 0 |
| `failures.csv` | chỉ có dòng tiêu đề, 0 lỗi |
| Chạy lại | không có lượt nào bị chạy lại |
| Drift | 19/19 ô (kể cả ô khởi động `leastloaded round0`) có đúng 1 sự kiện `drift_on` (dispatch ≈ 107, `stress_started trigger=101`) và 1 `drift_off` ở cuối ô |
| Thứ tự | ngẫu nhiên theo vòng (`order_log.csv`) |

### Đầu ra `summarize.py` (nguyên văn)

```
build 2 seconds; rounds [1, 2, 3]
hybrid-linucb-d-arm-g098   [28.892, 30.384, 27.277] mean=  28.9 vsLL=[-17.0, -8.1, -19.0] vsSED=[-5.1, -0.7, -9.3]
hybrid-linucb              [31.942, 30.223, 28.895] mean=  30.4 vsLL=[-8.2, -8.5, -14.2] vsSED=[4.9, -1.3, -3.9]
sed                        [30.447, 30.608, 30.065] mean=  30.4 vsLL=[-12.5, -7.4, -10.7] vsSED=[0.0, 0.0, 0.0]
leastloaded                [34.811, 33.045, 33.665] mean=  33.8 vsLL=[0.0, 0.0, 0.0] vsSED=[14.3, 8.0, 12.0]
linucb                     [35.04, 36.246, 41.464] mean=  37.6 vsLL=[0.7, 9.7, 23.2] vsSED=[15.1, 18.4, 37.9]
p90                        [46.133, 48.276, 47.584] mean=  47.3 vsLL=[32.5, 46.1, 41.3] vsSED=[51.5, 57.7, 58.3]

share of build-2 tasks on worker-5 per round:
hybrid-linucb-d-arm-g098   round 1: w5  6.6%  {'worker-1': 7, 'worker-2': 11, 'worker-3': 52, 'worker-4': 72, 'worker-5': 10}
hybrid-linucb-d-arm-g098   round 2: w5  5.9%  {'worker-1': 45, 'worker-2': 9, 'worker-3': 57, 'worker-4': 32, 'worker-5': 9}
hybrid-linucb-d-arm-g098   round 3: w5  2.0%  {'worker-1': 7, 'worker-2': 25, 'worker-3': 53, 'worker-4': 64, 'worker-5': 3}
hybrid-linucb              round 1: w5  6.6%  {'worker-1': 6, 'worker-2': 44, 'worker-3': 20, 'worker-4': 72, 'worker-5': 10}
hybrid-linucb              round 2: w5  3.3%  {'worker-1': 5, 'worker-2': 15, 'worker-3': 56, 'worker-4': 71, 'worker-5': 5}
hybrid-linucb              round 3: w5  3.3%  {'worker-1': 19, 'worker-2': 6, 'worker-3': 63, 'worker-4': 59, 'worker-5': 5}
leastloaded                round 1: w5 12.5%  {'worker-1': 44, 'worker-2': 11, 'worker-3': 47, 'worker-4': 31, 'worker-5': 19}
leastloaded                round 2: w5 10.5%  {'worker-1': 24, 'worker-2': 29, 'worker-3': 32, 'worker-4': 51, 'worker-5': 16}
leastloaded                round 3: w5 11.8%  {'worker-1': 34, 'worker-2': 22, 'worker-3': 59, 'worker-4': 19, 'worker-5': 18}
linucb                     round 1: w5  3.3%  {'worker-1': 22, 'worker-2': 13, 'worker-3': 16, 'worker-4': 96, 'worker-5': 5}
linucb                     round 2: w5  9.9%  {'worker-1': 37, 'worker-2': 9, 'worker-3': 11, 'worker-4': 80, 'worker-5': 15}
linucb                     round 3: w5 29.6%  {'worker-1': 14, 'worker-2': 15, 'worker-3': 13, 'worker-4': 65, 'worker-5': 45}
p90                        round 1: w5 15.8%  {'worker-1': 20, 'worker-2': 32, 'worker-3': 53, 'worker-4': 23, 'worker-5': 24}
p90                        round 2: w5 16.4%  {'worker-1': 33, 'worker-2': 56, 'worker-3': 13, 'worker-4': 25, 'worker-5': 25}
p90                        round 3: w5 15.8%  {'worker-1': 27, 'worker-2': 24, 'worker-3': 20, 'worker-4': 57, 'worker-5': 24}
sed                        round 1: w5 14.5%  {'worker-3': 59, 'worker-4': 71, 'worker-5': 22}
sed                        round 2: w5 15.1%  {'worker-3': 58, 'worker-4': 71, 'worker-5': 23}
sed                        round 3: w5 17.1%  {'worker-3': 54, 'worker-4': 72, 'worker-5': 26}
```

### Tóm tắt (build 2, giây)

| Scheduler | Lần 1 | Lần 2 | Lần 3 | TB | So với LL (TB) | So với SED (TB) | Tỉ lệ task vào worker-5 (TB 3 lần) |
|---|---|---|---|---|---|---|---|
| hybrid-linucb-d-arm-g098 | 28,89 | 30,38 | 27,28 | 28,85 | −14,7% | −5,01% | 4,8% |
| hybrid-linucb | 31,94 | 30,22 | 28,90 | 30,35 | −10,3% | −0,07% | 4,4% |
| sed | 30,45 | 30,61 | 30,07 | 30,37 | −10,2% | — | 15,6% |
| leastloaded | 34,81 | 33,05 | 33,67 | 33,84 | — | +11,4% | 11,6% |
| linucb (thuần) | 35,04 | 36,25 | 41,46 | 37,58 | +11,0% | +23,7% | 14,3% |
| p90 | 46,13 | 48,28 | 47,58 | 47,33 | +39,9% | +55,8% | 16,0% |

Chênh tuyệt đối của HG-D so với SED theo từng lần: −1,55 s, −0,22 s, −2,79 s.

## 3. Kết luận P1 theo 4 tiêu chí mục 8

| # | Tiêu chí | Kết quả | Đạt? |
|---|---|---|---|
| 1 | Hợp lệ: `failures.csv` không có dòng, rc = 0 | rc = 0, `failures.csv` rỗng | **Đạt** |
| 2 | Một biến thể HG nhanh hơn LL **và** SED ở cả 3/3 lần | `hybrid-linucb-d-arm-g098`: hơn LL ở 3/3 (−17,0 / −8,1 / −19,0%) và hơn SED ở 3/3 (−5,1 / −0,7 / −9,3%). `hybrid-linucb` thì không: lần 1 chậm hơn SED +4,9% | **Đạt** (nhờ HG-D) |
| 3 | Mức hơn SED trung bình ≥ 5% | HG-D: −5,01% (tỉ số các trung bình: 28,851 s so với 30,373 s) hay −5,04% (trung bình của 3 tỉ lệ từng lần) | **Đạt, sát ngưỡng** (hơn ngưỡng 0,01 điểm %) |
| 4 | Tỉ lệ task vào worker-5 của biến thể HG ≤ một nửa của SED | HG-D 6,6 / 5,9 / 2,0% so với SED 14,5 / 15,1 / 17,1%. Mọi lần đều ≤ một nửa SED (ngưỡng 7,3 / 7,6 / 8,6%); TB 4,8% so với 15,6% | **Đạt** |

**Kết luận: P1 lặp lại thành công theo đúng 4 tiêu chí đã chốt.** Tiêu chí 2 dùng "hoặc", và chỉ HG-D thoả; HG thường không thoả tiêu chí 2 và 3.

**Mức độ chắc chắn của kết quả cần đọc kèm các điểm sau** (mô tả, không phải đánh giá lại tiêu chí):
- Biên của tiêu chí 3 chỉ 0,01 điểm %. Chỉ cần 3 lần lặp lệch nhẹ là kết luận đổi chiều.
- Lần 2: HG-D chỉ hơn SED 0,22 s (−0,7%), gần như hoà. Sự hơn kém "3/3" vì thế mỏng ở lần này.
- So với máy tham chiếu, lợi thế nhỏ hơn: HG-D so với SED −5,0% (máy tham chiếu −9,8%), so với LL −14,7% (−20,6%).
- Thời gian build 2 chỉ ~29–30 s, ngắn hơn ~3 lần máy tham chiếu (87–110 s). Chênh tuyệt đối giữa các scheduler chỉ vài giây.
- Cơ chế (tiêu chí 4) lặp lại rõ: HG né worker-5 (2–7%) trong khi SED, LL, P90 gửi 10–17%. Dù vậy HG-D chỉ nhanh hơn SED một chút, nên cơ chế né đúng nhưng lợi ích về thời gian trên máy này nhỏ.
- `hybrid-linucb` thường chỉ hoà SED (−0,07%). LinUCB thuần không ổn định: round 3 gửi 29,6% task vào worker-5 và chậm nhất trong nhóm có học (41,5 s).
- Số lần lặp chỉ là 3 và pilot này chưa có kiểm định thống kê.

## 4. Bất thường

- **Không có lỗi build, timeout hay lượt chạy lại.** Log chỉ có một dòng khớp "error": `Exception ignored in: <_io.TextIOWrapper name='<stdout>' mode='w' encoding='cp1252'> / OSError: [Errno 22] Invalid argument` (log dòng 1266–1267, ô `linucb` round 1). Đây là Python trên Windows in ra stdout (cp1252) bị lỗi, không ảnh hưởng ô chạy (ô vẫn hoàn tất, không vào `failures.csv`).
- **NVIDIA Overlay** (5 tiến trình) bị app NVIDIA bật lại ngay sau khi tắt, nên vẫn chạy trong lúc pilot. Mức dùng CPU của nó gần như bằng 0 khi đo.
- Phần nền còn lại trong lúc chạy: Antigravity IDE (kèm phiên Claude và java, codex), Docker Desktop, các công cụ truy cập từ xa. Không có công cụ nào khác được chạy trong lúc pilot.
- Không đo nhiệt độ CPU, không đo CPU host trong suốt lúc chạy. Chỉ có thống kê CPU container (`stats_*.csv`) và đo CPU host rảnh trước khi chạy.
- Windows Update đã pause trước khi chạy; hai service `UsoSvc` và `wuauserv` vẫn đang chạy nhưng không tải bản cập nhật.
- Ô khởi động `leastloaded round0` của runner là có chủ đích, không phải lượt chạy lại.
- Compose project `hgexp` dùng chung tên với image và volume cũ của các lượt trước trên máy này. Runner làm mới volume bằng marker; `meta.json` ghi `git_dirty=false`.

## 5. Việc chưa làm và bước tiếp theo

- P2, P3, P4 **chưa chạy**, vì chờ Kien xác nhận (theo kế hoạch, chỉ chạy khi P1 hợp lệ và có đồng ý).
- Power plan đang là High performance và các ứng dụng đã tắt vẫn giữ nguyên; plan cũ (Balanced, `381b4222-f694-41f0-9685-ff5bb260df2e`) lưu ở `C:\bench-work\power-previous.txt` để khôi phục khi xong.
- Không sửa code, không sửa `hypotheses*.md`, không commit hay push, không chạy thí nghiệm chính thức.

## 6. Tệp

- Dữ liệu nén: `C:\bench-work\out\absl_j3_h3.tar.xz` (≈ 353 KB; gồm thư mục `absl_j3_h3`, `absl_j3_h3.log`, `P1-driver.log`, `P1-summary.txt`, `preflight.txt`)
- Dữ liệu gốc: `C:\bench-work\out\pilot-hidden\`
- Script: `C:\bench-work\pilot_hidden.sh`, `C:\bench-work\summarize.py`
