# Dữ liệu và công cụ thực nghiệm bổ sung (TN1–TN3)

Thư mục này chứa kết quả chính thức của Thí nghiệm 1–2 và bộ công cụ điều phối đã dùng. Trạng thái tổng
thể và quy ước làm việc: `docs/thesis/TIEN-DO-THUC-NGHIEM.md`.

## `tn1-tn2/`: kết quả chính thức Thí nghiệm 1 và 2

Đăng ký trước: `docs/thesis/hypotheses.md` (đóng băng ở commit `1b8339d`). Dữ liệu chạy trên máy tham
chiếu (2 × Xeon E5-2686 v4, Docker Desktop 72 CPU / 31 GiB), từ 2026-10-10 đến 2026-10-21.

| Thư mục | Nội dung |
|---|---|
| `E1-j6`, `E1-j7`, `E1-j8`, `E1-j9`, `E1-2c` | TN1: tải trung gian (4 scheduler × 20 block) |
| `E2-perm`, `E2-trans`, `E2-onoff` | TN2: drift ẩn (7 scheduler × 20 block, phiên 2 build) |
| `E2-lambda` | TN2: quét λ (10 scheduler × 20 block) |

Trong mỗi thư mục kịch bản:
- `results.csv`, `order_log.csv`, `failures.csv`, `meta.json`, `preflight_r*.txt`, và `reruns.csv` /
  `interruptions.csv` nếu có;
- `analysis/`, `analysis_*` (`pairs.csv`, `cells.csv`, `summary.md`) cùng log và dấu `analysis.done`:
  kết quả `scripts/analyze_exp.py` tại `1b8339d`;
- `raw.tar.xz`: **mọi file còn lại** của thư mục gốc (task log JSONL của coordinator, sự kiện drift,
  `docker stats`, log build, log runner, file đánh dấu, `archive/` của block dở), nén xz.

File chung:
- `campaign.log`: nhật ký điều phối, kể cả lần tạm dừng 19→21/10.
- `PHASE2_APPROVED`, `PHASE2_DECISION`: quyết định chạy E2-lambda (RUN), khoá một lần.
- `manifest_E1.sha256`, `manifest_E2-*.sha256`: SHA-256 của mọi file gốc.
- `supervisor/`: 39 audit của giám sát viên Codex cùng `supervisor.log`.
- `pause-snapshot/`: trạng thái máy lúc tạm dừng (dùng cho `resume_check.sh`).

### Kiểm tra toàn vẹn (Git Bash)

```bash
cd experiments/tn1-tn2
s=E2-perm; mkdir -p /tmp/v && cp -r $s /tmp/v/ && (cd /tmp/v/$s && tar -xJf raw.tar.xz && rm raw.tar.xz)
(cd /tmp/v && sha256sum -c "$OLDPWD/manifest_$s.sha256")   # E1-*: dùng manifest_E1.sha256, lọc dòng của kịch bản
```

Đã kiểm tra trước khi commit: 9/9 kịch bản khớp manifest (ví dụ E1-j8 502/502 file, E2-lambda 1628/1628
file).

### Phân tích lại

```bash
python3 scripts/analyze_exp.py <thư mục kịch bản đã giải nén> --strict --out <thư mục ra>
```

Phân tích phải chạy trên thư mục đã giải nén `raw.tar.xz` (analyzer cần task log). Analyzer của nhánh
`feat/exp3` cho kết quả giống hệt bản `1b8339d` trên E1-j8 và E2-perm (đã kiểm tra).

## `tooling/`: công cụ điều phối (bản sao từ `D:\bench-work`)

Các script này viết cho máy tham chiếu, với đường dẫn cứng `D:\bench-work`, `/d/bench-work` và Git Bash
trên Windows. Chúng được lưu ở đây để tái lập và để máy khác hiểu quy trình, **không** nhằm chạy
nguyên trạng trên máy khác.

| File | Vai trò |
|---|---|
| `campaign.sh` | điều phối TN1–2 (thứ tự cố định, sổ chạy lại, lưu block dở, khoá λ, dừng an toàn) |
| `campaign_tn3.sh`, `supervisor_tn3.sh`, `supervisor_prompt_tn3.md` | điều phối và giám sát TN3 |
| `supervisor.sh`, `supervisor_prompt.md` | giám sát viên Codex chỉ-đọc (audit 90 phút) |
| `preflight.sh`, `power.ps1`, `env.sh`, `bin/python3` | kiểm tra máy, power plan, môi trường Python |
| `stop_campaign.sh`, `snapshot_state.sh`, `pause_at_block_end.sh`, `resume_check.sh` | dừng an toàn, chụp và so trạng thái khi tạm dừng/chạy tiếp |
| `analysis_marker.sh` | dấu hoàn tất phân tích (không phân tích lại kết quả đã có) |
| `tests/` | test cho campaign và supervisor (runner giả) |
| `smoke_*.sh`, `check_smoke.sh`, `pilot_*.sh`, `tn3-smoke/` | smoke/pilot (dữ liệu pilot không chính thức, không nằm trong repo) |
| `analysis/` | script phân tích bổ sung cho báo cáo TN2 (trích bảng, kiểm tra sự kiện, hình, kiểm toán) |
| `reviews/` | yêu cầu giao việc và phản biện của Codex (worker `gpt-6-sol`, reviewer `gpt-5.6-sol`) |
| `CHANGELOG-phase2.md` | lịch sử thay đổi công cụ kèm SHA-256 |
