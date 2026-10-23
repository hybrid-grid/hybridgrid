# Tiến độ thực nghiệm bổ sung HG-LinUCB (đồng bộ giữa các máy)

*Cập nhật: 2026-10-22. File này để người và các phiên Claude Code trên máy khác nắm nhanh trạng thái,
quy ước và việc tiếp theo. Khi trạng thái thay đổi, cập nhật file này trong cùng commit.*

## 1. Trạng thái

| TN | Nội dung | Trạng thái | Tài liệu |
|---|---|---|---|
| 1 | Tải trung gian (LL, SED, HG-LinUCB, HG-LinUCB-D; 5 kịch bản × 20 block) | ✅ xong 2026-10-11, 400 cell, 0 lỗi | `bao-cao-tien-do-261011.md` (+ `-slides.md`) |
| 2 | Drift ẩn (perm/trans/onoff × 7 scheduler) + quét λ (10 scheduler) | ✅ xong 2026-10-21, 620 cell, 1 cell chạy lại hợp lệ | `ke-hoach-thi-nghiem-2.md`, `bao-cao-tn2-261021.md`, `figures/tn2-rolling-share.svg` |
| 3 | C++ đuôi nặng: workload tổng hợp, abseil-cpp, tổng hợp + drift; 7 scheduler × 20 block | 🔧 chuẩn bị xong, pilot đạt; **chưa đóng băng, chưa chạy chính thức** | `ke-hoach-thi-nghiem-3.md` (mục 10 là kế hoạch chạy), `hypotheses-tn3.md` (bản nháp) |
| 4–6 | Thiếu RAM, cache affinity, cụm vật lý | chưa bắt đầu; theo thứ tự thầy giao | `phan-tich-leastloaded-va-kich-ban.md` |

Kết quả chính đến nay:
- **TN1:** HG-LinUCB không nhanh hơn LeastLoaded (H1.1 không được ủng hộ). HG-LinUCB chọn máy rảnh mạnh
  nhất tốt hơn ở 3/5 kịch bản. SED bỏ qua máy rảnh nhiều hơn LeastLoaded (5/5 kịch bản).
- **TN2:** không giả thuyết H2.x nào đạt sau hiệu chỉnh Holm. Ngoài giả thuyết, HG-LinUCB nhanh hơn
  LeastLoaded 2–6 s khi có drift. Bản "mọi máy cùng quên" chậm hơn HG-LinUCB; bản "chỉ máy vừa nhận kết
  quả quên" thì không phát hiện khác biệt có ý nghĩa. LeastLoaded tự né máy chậm qua số task đang chạy.

## 2. Nhánh và commit

| Nhánh / commit | Ý nghĩa |
|---|---|
| `feat/exp-readiness` @ `1b8339d` | **đóng băng đăng ký TN1–2** (`docs/thesis/hypotheses.md`). TN1–2 chạy đúng commit này |
| `feat/exp3` | công việc TN3, tách từ `1b8339d`; có thêm dữ liệu TN1–2 (`experiments/`), báo cáo, công cụ, và code TN3 |
| (sắp tới) commit đóng băng TN3 trên `feat/exp3` | sẽ chứa bản cuối `hypotheses-tn3.md`; TN3 chạy đúng commit đó, từ một clone LF sạch |

Code TN3 (đã test, chưa đóng băng): sửa phân tán CMake + Ninja (`internal/compiler`, `cmd/hgbuild`),
`HG_TIMEOUT`, cờ `--request-timeout` của coordinator, thang phần thưởng cố định, scheduler `p90` /
`sed-p90`, runner đa workload (`WORKLOAD=cpython|absl|syn`), bộ sinh `scripts/gen_synthetic_cpp.py`,
metric mới trong `scripts/analyze_exp.py`.

## 3. Quy ước bắt buộc

1. **Đăng ký trước:**
   - Không sửa `hypotheses.md` (TN1–2) hay `hypotheses-tn3.md` sau khi đã đóng băng. Thay đổi về sau chỉ
     là kịch bản bổ sung ở file khác, kèm lý do.
   - Không chạy thí nghiệm chính thức khi người dùng chưa xác nhận.
   - Không xem so sánh scheduler trước khi chạy xong (bịt mắt).
2. **Máy chạy:**
   - Mọi lượt chạy chính thức dùng **máy tham chiếu**: Windows 10, 2 × Xeon E5-2686 v4, Docker 72 CPU /
     31 GiB, thư mục `D:\bench-work`.
   - Không chia một thí nghiệm ra nhiều máy; máy khác chỉ dùng để phát triển, review và pilot.
3. **Trước mỗi lượt chạy:**
   - Buộc tắt ứng dụng không cần thiết (Edge, Unity, Rider…); giữ IDE đang chứa phiên Claude, `codex` và
     Remote Desktop.
   - Chạy `preflight.sh` với `STRICT=1`; tạm dừng Windows Update.
4. **Codex CLI:**
   - Worker `gpt-6-sol`; reviewer và supervisor `gpt-5.6-sol` (`gpt-5.5` không còn được hỗ trợ từ khoảng
     12/10). Không dùng "astra".
   - Mọi thay đổi code hoặc kế hoạch quan trọng đều qua review của Codex.
5. **Dữ liệu:**
   - Pilot và dữ liệu thăm dò nằm ngoài repo (`D:\bench-work\out\pilot-*`), không dùng làm kết quả.
   - Kết quả chính thức được nén vào `experiments/` (xem `experiments/README.md`).
6. **Git:** chỉ commit hoặc push khi người dùng yêu cầu.

## 4. Lưu ý kỹ thuật trên Windows (máy tham chiếu)

- Bản checkout chính dùng `core.autocrlf=true`, nên lượt chạy dùng một **clone LF riêng**
  (`D:\bench-work\hybridgrid` cho TN1–2, sẽ là `hybridgrid-tn3` cho TN3).
- `python3` của Windows là bản giả (Store alias); dùng shim `D:\bench-work\bin\python3` (venv, xem
  `env.sh`).
- Git Bash tự đổi đối số bắt đầu bằng `/`, nên truyền đường dẫn trong container dưới dạng tương đối
  hoặc dùng `MSYS_NO_PATHCONV=1`.
- `docker compose cp` không nhận đường dẫn Windows; dùng luồng `tar` qua `compose exec`.
- Không có Go/CMake/ninja trên Windows: unit test Go chạy trong `golang:1.25-bookworm`; build C++ chạy
  trong image `test/stress/Dockerfile.base` (đã có cmake, ninja, `hgcc`/`hgcxx`).

## 5. Việc tiếp theo

0. **Cập nhật 2026-10-23.** Phân tích dư địa cho thấy ở -j8 cụm bão hoà, nên TN3 như thiết kế khó phân biệt
   các scheduler. Pilot trên máy tham chiếu tìm được kịch bản HG-LinUCB thắng cả LL lẫn SED (3/3 lần): absl,
   -j3, worker-5 bị chiếm ngầm. Pilot lặp lại trên máy khác: `ke-hoach-pilot-hg-loi-the.md`. Kế hoạch TN3 bên
   dưới đang chờ người dùng quyết định lại.
1. **TN3** (khi người dùng nói bắt đầu): làm theo `ke-hoach-thi-nghiem-3.md` mục 10:
   - rà soát và test lần cuối;
   - đóng băng `hypotheses-tn3.md` bằng commit;
   - tạo clone LF sạch, gia hạn tạm dừng Windows Update, buộc tắt app, chạy preflight;
   - chạy `campaign_tn3.sh` cùng `supervisor_tn3.sh` (khoảng 45 giờ).
2. Gửi thầy báo cáo TN1–2. Câu hỏi còn mở với thầy: bản D chính; kịch bản bổ sung mà LeastLoaded không
   tự né được; cách lưu log thô (hiện đã nén trong repo, khoảng 36 MB).
3. TN4–6 theo thứ tự, sau khi thầy xem kết quả.
