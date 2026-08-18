# Báo cáo tiến độ: chẩn đoán nguyên nhân thế hoà và hướng chỉnh sửa bài báo

*Ngày 09/08/2026*

Kính gửi thầy,

Sau buổi làm việc trước, chúng em đã rà soát lại toàn bộ dữ liệu đo và xác định được nguyên nhân vì sao bộ lập lịch do chúng em đề xuất chỉ đạt kết quả ngang bằng với heuristic LeastLoaded. Báo cáo này trình bày kết quả chẩn đoán, một đính chính quan trọng trong phần khảo sát, và hướng viết lại bài báo theo góp ý của thầy.

---

## 1. Chẩn đoán nguyên nhân thế hoà

### 1.1. Cụm thí nghiệm chưa đủ tải

Nguyên nhân của thế hoà không nằm ở thuật toán mà nằm ở cấu hình thí nghiệm.

Cụm gồm 5 worker với tổng cộng **10 vị trí thực thi** (phân bổ 3–1–2–2–2 theo giới hạn cgroup), trong khi client biên dịch chạy `make -j5`, tức tối đa **5 tác vụ đồng thời**. Mức chiếm dụng trần của cụm vì vậy chỉ đạt **50%**: tại mọi thời điểm luôn dư ít nhất 5 vị trí trống.

Chúng em xác định đây là thiếu sót trong thiết kế thí nghiệm của mình. Trước khi chạy, chúng em đã không kiểm tra quan hệ giữa tham số `-j` của client và tổng công suất cụm, dẫn tới việc đo một bài toán lập lịch đã bị làm cho dễ đi.

### 1.2. Hệ quả: trần hiệu năng bằng nhau, sàn hiệu năng khác nhau

Khi cụm luôn dư năng lực, mọi bộ lập lịch có ưu tiên máy rảnh đều đạt cùng một makespan tối ưu. **Trần hiệu năng của các thuật toán vì thế bằng nhau**, và về nguyên tắc không bộ lập lịch nào có thể vượt LeastLoaded. Thế hoà quan sát được là kết quả đúng của cấu hình này, không phải dấu hiệu thuật toán kém.

Tuy nhiên **sàn hiệu năng thì không bằng nhau**. Trong điều kiện luôn có máy rảnh, một bộ lập lịch vẫn có thể chủ động chọn sai bằng cách bỏ qua máy rảnh để giao việc vào máy đang bận. Đây là loại lỗi duy nhất còn khả năng xảy ra, và nó đo được trực tiếp từ nhật ký thi hành.

Hai nhận định trên giải thích trọn vẹn hai kết quả tưởng như mâu thuẫn: bộ lập lịch học **không thắng được** LeastLoaded (do trần bằng nhau), nhưng lại **thắng rõ rệt** LinUCB thuần (do sàn khác nhau).

### 1.3. Kết quả đo

Chúng em định nghĩa **tỉ lệ bỏ phí máy rảnh** là tỉ lệ số lần giao việc vào một worker đã có tác vụ đang chạy. Đo trên workload nhẹ (10 khối × 293 tác vụ, tương ứng 2.930 lượt giao việc cho mỗi bộ lập lịch):

| Bộ lập lịch | Tỉ lệ bỏ phí máy rảnh | Makespan trung vị |
|---|---:|---:|
| LeastLoaded | 0,0% | 40,24 s |
| **Hybrid-LinUCB** | **0,2%** | **40,52 s** |
| P2C | 20,0% | 45,60 s |
| LinUCB thuần | 30,8% | 47,48 s |

Con số 0,0% của LeastLoaded là hệ quả tất yếu của định nghĩa thuật toán — nó luôn chọn máy ít tải nhất — nên chúng em sử dụng nó như phép kiểm tra tính đúng đắn của công cụ đo, không xem là phát hiện.

Hồi quy giữa tỉ lệ bỏ phí và makespan trên toàn bộ 40 lần chạy cho **R² = 0,82** (r = 0,906); hệ số góc tương ứng khoảng **2,2 giây makespan cho mỗi 10 điểm phần trăm bỏ phí**. Chênh lệch 6,80 giây giữa Hybrid-LinUCB và LinUCB thuần được giải thích gần như trọn vẹn bằng cơ chế này.

Các kiểm định ghép cặp trên cùng bộ dữ liệu:

| So sánh | Chênh lệch trung vị | p |
|---|---:|---:|
| Hybrid-LinUCB vs LinUCB thuần | +6,80 s | 0,003 (sau Holm) |
| Hybrid-LinUCB vs P2C | +5,26 s | 0,003 (sau Holm) |
| Hybrid-LinUCB vs LeastLoaded | −0,09 s | 0,65 — hoà |

Toàn bộ số liệu trong mục này lấy từ **cùng một thí nghiệm** (workload nhẹ, thiết kế khối ngẫu nhiên hoá 10 khối), nhằm tránh lặp lại lỗi ghép số liệu chéo giai đoạn đo mà chúng em đã mắc trước đây.

### 1.4. Giới hạn của kết luận

Trên workload nặng (10 khối × 371 tác vụ), quan hệ giữa tỉ lệ bỏ phí và makespan yếu đi rõ rệt: **R² chỉ còn 0,07**, và nếu loại một lần chạy ngoại lai (101,9 giây, gấp đôi trung vị) thì đạt 0,31. Phương sai giữa các lần chạy lấn át tín hiệu cơ chế.

Chúng em vì vậy **chỉ khẳng định cơ chế này trên workload nhẹ** và không suy rộng sang workload nặng. Làm rõ nguyên nhân phương sai lớn ở workload nặng là một hạng mục còn tồn đọng, ghi ở mục 5.

---

### 1.5. Nguyên nhân thứ hai: đặc trưng CPU bị thoái hoá

Trong quá trình chuẩn bị phần so sánh với icecream, chúng em phát hiện thêm một lỗi độc lập, ảnh hưởng còn trực tiếp hơn.

- Khi worker chạy dưới giới hạn CPU của Docker, trường `cpu_cores` mà worker báo về **là số lõi của máy chủ**, không phải phần được cấp. Giới hạn `--cpus` không được phản ánh.
- Cụm thí nghiệm giới hạn 0,5 / 0,6 / 0,8 / 1,0 / 1,1 lõi, nên **cả 5 worker đều báo cùng một giá trị**.
- Hệ quả: thành phần CPU trong vector đặc trưng của LinUCB là một **hằng số không phân biệt được worker**. Bộ học không hề quan sát được sự không đồng nhất mà nó được thiết kế để khai thác. Hàm chấm điểm năng lực của HEFT mắc cùng lỗi này.

Lỗi này đã được xác nhận trực tiếp trên cụm thật ngày 18/08/2026. Dựng cụm 5 worker với hạn ngạch 0,5 / 0,6 / 0,8 / 1,0 / 1,1 lõi rồi chạy 190 tác vụ biên dịch, kết quả ghi trong nhật ký tác vụ:

| Hạn ngạch cgroup | `cpu_cores` báo về | `cpu_millis` báo về |
|---|---:|---:|
| 0,5 lõi | 8 | 500 |
| 0,6 lõi | 8 | 600 |
| 0,8 lõi | 8 | 800 |
| 1,0 lõi | 8 | 1000 |
| 1,1 lõi | 8 | 1100 |

Cột `cpu_cores` cho giá trị **8 ở cả năm worker** — đúng bằng số lõi của máy chủ, không phải phần được cấp — trong khi năng lực thật giữa chúng chênh nhau tới 2,2 lần. Bộ nhớ không dính lỗi này, `memory_gb` báo đúng theo giới hạn container. Bằng chứng đầy đủ lưu tại `.sisyphus/evidence/cgroup-smoke-260818/`.

Chúng em đã sửa: bổ sung trường `cpu_millis` đọc trực tiếp hạn ngạch cgroup (giữ được phần lẻ dưới một lõi), và chuyển chuẩn hoá đặc trưng CPU cùng bộ nhớ từ tuyến tính sang thang log. Cột `cpu_millis` trong bảng trên xác nhận bản sửa hoạt động đúng trên toàn bộ chuỗi: worker đọc cgroup, báo qua gRPC, coordinator ghi nhận. Lý do đổi sang log: với thang tuyến tính trên 16 lõi, 0,5 và 1,1 lõi rơi vào 0,03 và 0,07 — chênh lệch 0,04, gần như tan trong sai số dấu phẩy động khi cập nhật ma trận theo công thức Sherman–Morrison. Trong thang log, hai giá trị này thành khoảng 0,64 và 0,72, tức khoảng cách rộng gấp tám lần.

### 1.6. Hệ quả chung của hai nguyên nhân

Hai lỗi trên độc lập với nhau và cùng làm sai lệch kết quả theo hướng bất lợi cho bộ lập lịch học:

- Bài toán quá dễ, không có gì để tối ưu (mục 1.1).
- Bộ học không nhìn thấy dữ liệu cần thiết để ra quyết định (mục 1.5).

Vì vậy chúng em cho rằng kết luận *"bộ lập lịch học chỉ đạt ngang heuristic"* **chưa đủ căn cứ** với tư cách một phát biểu về thuật toán: nó được đo trên một bộ học bị che mất tín hiệu đầu vào, trong một bài toán không còn quyết định nào để tối ưu.

Toàn bộ thí nghiệm cần chạy lại sau khi hai lỗi này được khắc phục. Kết quả có thể khác so với những gì đã báo cáo trước đây.

---

## 2. Đính chính một khẳng định sai trong phần khảo sát

Khi rà soát lại phần công trình liên quan, chúng em phát hiện bài báo đang khẳng định sai rằng chưa có hệ thống build phân tán nào học trực tuyến. Thực tế có ít nhất hai hệ thống:

- **icecream** ước lượng tốc độ từng máy một cách trực tuyến qua cửa sổ trượt 200 công việc gần nhất (hàm `server_speed()` trong `scheduler.cpp`), có điều tiết theo độ bão hoà hàng đợi, khuếch đại các máy còn ít mẫu, và ưu tiên gửi việc tới máy chưa từng được đo. Diễn đạt theo ngôn ngữ bandit, đây là một bộ ước lượng giá trị không ngữ cảnh, hiệu chỉnh thủ công, kèm cơ chế thăm dò.
- **Buildbarn** lưu kết quả các lần thực thi trước và xếp hạng máy bằng thuật toán PageRank.

Khẳng định sai này ảnh hưởng trực tiếp tới phát biểu đóng góp của bài báo, nên chúng em sẽ sửa lại và đồng thời bổ sung icecream làm mốc so sánh, thay vì chỉ đối chiếu với các heuristic đơn giản.

Phần cài đặt đã hoàn thành. Chúng em đã chuyển quy tắc chọn máy của icecream vào hệ thống của mình, đặt sau cùng một giao diện với các bộ lập lịch còn lại. Cách làm này giúp chênh lệch makespan quy được về quy tắc quyết định, không lẫn với khác biệt về giao thức truyền tải, nén dữ liệu và cách vận chuyển bộ công cụ biên dịch — vốn sẽ xuất hiện nếu đo trực tiếp phần mềm icecream.

Hai điểm buộc phải lệch so với bản gốc, đã ghi chú ngay tại vị trí phát sinh trong mã nguồn:

1. Hệ thống của chúng em không ghi nhận kích thước tệp đối tượng, nên dùng kích thước mã nguồn sau tiền xử lý làm đại lượng thay thế cho khối lượng công việc.
2. Daemon không báo cáo tải nhân hệ điều hành, nên bỏ hệ số `(1000 − load)/1000`.

Hệ số bão hoà hàng đợi được giữ nguyên và mang cùng ý nghĩa như bản gốc.

---

## 3. Hướng viết lại bài báo

Theo đúng góp ý của thầy — bài báo nên bám vào một vấn đề cụ thể gặp phải trong hệ thống, rồi trình bày cách giải và chỉ rõ điểm mới — chúng em viết lại theo mạch: hệ thống gặp vấn đề gì, nguyên nhân từ đâu, giải quyết bằng cách nào, điểm mới nằm ở đâu.

Phần khảo sát sáu thuật toán được hạ xuống thành công cụ thí nghiệm, không còn giữ vai trò đóng góp chính. Đóng góp chúng em dự kiến phát biểu lại như sau:

> Chỉ ra rằng giá trị thực tế của lập lịch học trong hệ thống build phân tán bị quyết định bởi **mức chiếm dụng cụm**; đề xuất một tiêu chí đo được để chẩn đoán điều kiện đó; và định vị kết quả trong tương quan với icecream — bộ lập lịch có học trực tuyến đang được dùng trong thực tế.

Cách phát biểu này giữ được kết quả âm — bộ lập lịch học chỉ hoà với heuristic khi cụm dư tải — như một phát hiện có giá trị khoa học, thay vì trình bày nó như một thất bại của phương pháp.

---

## 4. Kế hoạch thí nghiệm bổ sung

Để trả lời câu hỏi *khi nào mới thật sự cần đến thuật toán học*, chúng em sẽ đo lại ở các mức chiếm dụng tăng dần:

| Mức tải | `-j` của client | Tác vụ đồng thời / 10 vị trí | Mục đích |
|---|---:|---:|---|
| Hiện tại | 5 | 50% | mốc đối chiếu, đã có dữ liệu |
| Vừa đủ | 10 | 100% | cụm bắt đầu bão hoà |
| Quá tải | 16 | 160% | hàng đợi thực sự hình thành |
| Quá tải nặng | 24 | 240% | phân biệt rõ chất lượng quyết định |

Giữ nguyên thiết kế khối ngẫu nhiên hoá với 10 khối cho mỗi mức tải, kiểm định ghép cặp một phía và hiệu chỉnh Holm như quy trình hiện hành. Nhóm bộ lập lịch đưa vào so sánh gồm LeastLoaded, P2C, icecream, LinUCB thuần và Hybrid-LinUCB.

Chúng em **chưa chạy** phần thí nghiệm này nên chưa khẳng định điều gì về kết quả. Chúng em cũng lường trước khả năng LeastLoaded vẫn thắng ở mọi mức tải, bởi về bản chất nó là chính sách join-shortest-queue — một chính sách có nền tảng lý thuyết vững và nổi tiếng là khó vượt trong môi trường dừng. Nếu kịch bản đó xảy ra, kết luận của bài báo sẽ là xác định rõ ranh giới áp dụng của lập lịch học, và hướng khai thác còn lại sẽ chuyển sang cache-affinity cùng môi trường không dừng.

---

## 5. Các hạng mục còn tồn đọng

1. ~~Kiểm tra trực tiếp trên cụm rằng 5 worker báo đúng 500 / 600 / 800 / 1000 / 1100 milli-core.~~ **Đã xong ngày 18/08/2026** — kết quả ở mục 1.5, bằng chứng tại `.sisyphus/evidence/cgroup-smoke-260818/`.
2. Chạy lại toàn bộ thí nghiệm ở bốn mức tải nêu ở mục 4, sau khi đã khắc phục cả hai nguyên nhân. Việc này cần một máy đủ rảnh: máy dùng cho smoke test đang có tải nền cao (load average ~35 trên 8 lõi do các tiến trình đồng bộ của hệ điều hành), không đạt yêu cầu về nhiễu nêu ở §7.4 của bài báo.
3. Làm rõ nguyên nhân phương sai lớn trên workload nặng, đặc biệt lần chạy ngoại lai 101,9 giây.

Về tính tái lập: toàn bộ số liệu trong báo cáo này sinh ra từ `scripts/analyze_idle_skip.py`, chạy trên dữ liệu thô đã lưu trong kho mã nguồn. Thầy có thể kiểm chứng bằng lệnh:

```
python scripts/analyze_idle_skip.py .sisyphus/evidence/rigorous-v3.14.0
```

Chúng em sẽ hoàn thiện bản viết lại và gửi thầy trong thời gian sớm nhất. Nếu thầy thấy hướng đi này chưa phù hợp, kính mong thầy góp ý thêm để chúng em kịp điều chỉnh.

Em cảm ơn thầy ạ.
