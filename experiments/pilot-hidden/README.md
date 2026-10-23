# Pilot thăm dò "máy mạnh bị chiếm ngầm" (không phải dữ liệu chính thức)

Dữ liệu pilot P1 (`absl_j3_h3`: absl, -j3, worker-5 bị chiếm 3 luồng từ task 101, 3 lần lặp, 6 scheduler),
chạy 2026-10-23 trên máy phát triển (i5-14600KF, Docker 20 CPU / 23,4 GiB), commit `6e0fd92`.
Kế hoạch: `docs/thesis/ke-hoach-pilot-hg-loi-the.md`. Báo cáo: `docs/thesis/bao-cao-pilot-p1-hg-loi-the.md`.

- `absl_j3_h3.tar.xz`: thư mục kết quả của runner (`results.csv`, `failures.csv`, `meta.json`, task log,
  sự kiện drift, `docker stats`, log build), log runner, đầu ra `summarize.py` và báo cáo preflight.
- Đây là pilot: không đăng ký trước, không dùng làm kết quả luận văn. Kết quả chính thức nằm ở `tn1-tn2/`.
