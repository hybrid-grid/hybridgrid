# Đề xuất Nâng cấp Cơ chế Lập lịch AI: Hybrid-LinUCB

Tài liệu này đóng vai trò là "Bản Kế hoạch Triển khai" (Implementation Plan) nhằm nâng cấp trực tiếp mã nguồn của hệ thống Hybrid-Grid. 

Mục tiêu cốt lõi: Giải quyết những điểm yếu chí mạng của thuật toán học tăng cường hiện tại, nâng cao hiệu năng thực tế và tạo ra "đóng góp khoa học" (Contribution) vững chắc cho báo cáo đồ án tốt nghiệp.

## 1. Tình trạng hiện tại của Dự án

- **Kiến trúc:** Đã tích hợp thành công thuật toán Contextual Bandit (LinUCB - Algorithm 1) vào luồng lập lịch của Coordinator (`internal/coordinator/scheduler/linucb.go`).
- **Điểm mạnh:** Đã tối ưu hóa bước cập nhật ma trận nghịch đảo bằng công thức Sherman-Morrison (đạt độ phức tạp $\mathcal{O}(d^2)$ thay vì $\mathcal{O}(d^3)$). Đã khắc phục lỗi rò rỉ dữ liệu (target leakage) bằng cơ chế `pendingX`.
- **Thực trạng hiệu năng:** Trong các bài Benchmark thực tế, thuật toán AI (LinUCB) đang tỏ ra **lép vế hoặc chỉ ngang bằng** với thuật toán Heuristic truyền thống (như `LeastLoaded`).

## 2. Phát hiện Nhược điểm của LinUCB thuần túy

Việc áp dụng máy móc nguyên bản bài báo khoa học vào hệ thống thời gian thực (Real-time System) làm nảy sinh 3 nhược điểm lớn:

> **1. Vấn đề "Khởi động lạnh" (Cold Start)**
> Ma trận trí nhớ $A_a$ khởi tạo là Ma trận đơn vị. AI tốn khoảng 50-100 tác vụ đầu tiên chỉ để đưa ra các quyết định ngẫu nhiên mang tính "thử sai" (Exploration). Trong các project code nhỏ, lúc AI học xong thì việc biên dịch cũng vừa kết thúc.

> **2. Sự lệch pha do "Phản hồi trễ" (Delayed Feedback)**
> AI quyết định phân công dựa trên điểm số. Nhưng phần thưởng (thời gian biên dịch) phải mất vài giây mới trả về. Trong vài giây đó, hệ thống mù quáng nhồi nhét hàng chục file vào cùng một máy chủ nhanh nhất, gây ra hiện tượng **quá tải cục bộ**. Thuật toán `LeastLoaded` không bị lỗi này vì nó đếm số task theo mili-giây.

> **3. Bộ nhận thức (Feature Vector) chưa đủ sâu**
> Véc-tơ đặc trưng 9 chiều hiện tại chỉ xoay quanh cấu hình phần cứng tĩnh (CPU, RAM, OS). AI đang thiếu manh mối về chính "món hàng" cần xử lý (ví dụ: file C++ nặng hơn file C rất nhiều).

## 3. Đề xuất Giải pháp Tiếp theo: Kiến trúc Lai tạo (Hybrid-LinUCB)

Để vượt qua các thuật toán truyền thống, AI không được phép hoạt động độc lập mà phải **lai tạo** với các Heuristic chống nghẽn mạng.

### Giải pháp 1: Warm-start (Khởi động ấm)
Sử dụng quy tắc "Bảo mẫu": 100 tác vụ đầu tiên, ép hệ thống sử dụng thuật toán `LeastLoaded` để chia việc. LinUCB không được phép quyết định, mà chỉ được **quan sát và học ngầm** (Update $A$ và $b$). Sau 100 task, AI "đủ lông đủ cánh" mới chính thức tiếp quản.

### Giải pháp 2: Hybrid Q-Value (Tính điểm lai tạo)
Điểm số quyết định máy tính (Q-Value) sẽ được điều chỉnh bởi một hình phạt quá tải (Penalty):
$$Q_{hybrid} = \underbrace{(\hat{\theta}^\top x + \alpha \sqrt{x^\top A^{-1} x})}_{\text{Điểm AI}} - \underbrace{\lambda \cdot \text{LoadRatio}}_{\text{Phạt quá tải (Heuristic)}}$$
*Cơ chế này ngăn chặn tuyệt đối việc AI nhồi nhét file vào một máy đang bận.*

### Giải pháp 3: Nâng cấp Véc-tơ Đặc trưng (d=12)
Mở rộng thêm 3 chiều: 
1. Tính chất ngôn ngữ (Là C hay C++)
2. Kích thước file gốc
3. Tỉ lệ hoàn thành thành công gần đây của Worker

## 4. Kế hoạch Implement chi tiết vào mã nguồn

Dưới đây là các file sẽ bị can thiệp và các bước code cụ thể:

### Bước 1: Sửa đổi Feature Vector (Mở rộng trí não)
**File:** `internal/coordinator/scheduler/linucb.go`
- [MODIFY] Cập nhật hàm `featureDim()` từ trả về `9` thành `12`.
- [MODIFY] Sửa đổi hàm `featureVector()` để trích xuất thêm tính chất từ `TaskContext` (ví dụ: dựa vào phần mở rộng của `TaskID` hoặc kích thước).
- *Lưu ý: Việc này yêu cầu xóa các bài test cũ đang hardcode `dim=9`.*

### Bước 2: Cài đặt biến đếm tổng (Track Total Dispatches)
**File:** `internal/coordinator/scheduler/linucb.go`
- [MODIFY] Thêm trường `totalDispatches int64` vào struct `LinUCBScheduler`.
- [MODIFY] Tăng biến này mỗi khi hàm `SelectWithDispatchInfo` được gọi (dùng `atomic.AddInt64`).

### Bước 3: Cài đặt logic Hybrid-LinUCB trong hàm Select
**File:** `internal/coordinator/scheduler/linucb.go`
- [MODIFY] Trong hàm `SelectWithDispatchInfo`, nếu `totalDispatches < 100`:
  - Chuyển hướng quyết định sang chạy thuật toán `LeastLoaded` (tìm máy ít việc nhất).
  - Trả về máy đó, nhưng vẫn tính feature vector để lưu vào `pendingX` cho AI học sau này.
- [MODIFY] Nếu `totalDispatches >= 100`:
  - Khi tính `bestP`, áp dụng công thức trừ đi hình phạt: `p = (mean + bonus) - (LoadRatio * 0.5)`.

### Bước 4: Viết bộ Benchmark khoa học (Rigorous Benchmark)
**Thư mục:** `scripts/` hoặc `test/stress/`
- [NEW] Viết một file Bash script `benchmark_statistical.sh`.
- Kịch bản: Tự động chạy 10 lần thuật toán `LeastLoaded`, 10 lần `LinUCB cũ`, 10 lần `Hybrid-LinUCB`. Lưu kết quả ra file `results.csv`.
- Dùng Python để vẽ biểu đồ và tính chỉ số chứng minh thuật toán Hybrid ưu việt hơn.
