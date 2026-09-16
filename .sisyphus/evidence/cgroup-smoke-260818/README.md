# Smoke test: cpu_millis đọc đúng hạn ngạch cgroup

*Ngày 18/08/2026 — chạy trên macOS/arm64, Docker 29.4.0, cụm `test/stress/docker-compose-hetero.yml`*

Mục đích: xác nhận trường `cpu_millis` (thêm ở PR #9) đọc đúng giới hạn CPU của
container, trước khi dùng nó làm đầu vào cho bất kỳ loạt benchmark nào.

## Cách làm

Dựng cụm 5 worker không đồng nhất, chạy 190 tác vụ biên dịch C qua `hgbuild cc`,
rồi đọc `worker_cpu_millis` và `worker_cpu_cores` từ nhật ký tác vụ của
coordinator. Đợt đầu 90 tác vụ tuần tự chỉ chạm được 4 worker — worker yếu nhất
(0,5 lõi, `max-parallel=1`) không nhận tác vụ nào vì luôn có worker mạnh hơn đang
rảnh. Đợt sau chạy 100 tác vụ song song (`xargs -P 12`) để bão hoà cụm mới phủ
đủ 5.

## Kết quả

| Hạn ngạch cgroup | `cpu_millis` báo về | `cpu_cores` báo về | Số tác vụ nhận |
|---|---:|---:|---:|
| 0,5 lõi (`50000/100000`) | **500** | 8 | 9 |
| 0,6 lõi (`60000/100000`) | **600** | 8 | 20 |
| 0,8 lõi (`80000/100000`) | **800** | 8 | 42 |
| 1,0 lõi (`100000/100000`) | **1000** | 8 | 55 |
| 1,1 lõi (`110000/100000`) | **1100** | 8 | 64 |

`cpu_millis` khớp chính xác hạn ngạch ở cả năm worker. Chuỗi đầy đủ hoạt động
đúng: worker đọc cgroup → báo qua gRPC → coordinator ghi vào nhật ký.

## Lỗi cũ được chứng minh trực tiếp

Cột `cpu_cores` cho **8 ở cả năm worker** — đúng bằng số lõi luận lý của máy chủ,
không phải phần được cấp. Đây là bằng chứng sống cho lỗi mô tả ở mục 1.5 của
`docs/thesis/bao-cao-tien-do-260809.md`: trước khi có `cpu_millis`, thành phần CPU
trong vector đặc trưng của LinUCB là một hằng số không phân biệt được worker, dù
năng lực thật giữa chúng chênh nhau tới 2,2 lần.

Bộ nhớ thì không dính lỗi này — `memory_gb` báo đúng 0,50 / 0,60 / 0,80 / 1,00 /
1,10 GB theo giới hạn container.

## Ghi chú cho lần benchmark tới

Việc worker yếu nhất không nhận tác vụ nào trong 90 lượt tuần tự là biểu hiện của
đúng hiện tượng cụm dư tải nêu ở mục 1.1 của báo cáo: khi luôn có máy rảnh, tác vụ
dồn về các worker mạnh và worker yếu gần như không được dùng tới. Chỉ khi ép song
song 12 luồng thì nó mới nhận việc.

Dữ liệu thô: `tasks.jsonl` (190 bản ghi).
