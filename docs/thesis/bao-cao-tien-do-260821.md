# Báo cáo tiến độ: kết quả đo hai mức tải và đính chính lập luận trong báo cáo trước

*Ngày 21/08/2026*

Kính gửi thầy,

Báo cáo ngày 09/08 có nêu một kế hoạch đo lại ở bốn mức chiếm dụng cụm. Chúng em đã chạy xong phần thực hiện được của kế hoạch đó. Kết quả buộc chúng em phải **đính chính một lập luận trung tâm** của chính báo cáo trước, đồng thời cho thấy một giới hạn kiến trúc của hệ thống mà trước đây chúng em chưa biết.

Báo cáo này trình bày ba nội dung: kết quả đo, phần đính chính, và giới hạn phát hiện được.

---

## 1. Kết quả đo

Thiết kế giữ nguyên như quy trình đã báo cáo: khối ngẫu nhiên hoá đầy đủ, 10 khối mỗi mức tải, mỗi khối chạy đủ năm bộ lập lịch theo thứ tự xáo trộn có seed riêng, kiểm định Wilcoxon ghép cặp một phía, hiệu chỉnh Holm–Bonferroni. Cụm 5 worker giới hạn 0,5 / 0,6 / 0,8 / 1,0 / 1,1 lõi, tổng 10 vị trí thực thi. Cả hai nguyên nhân nêu ở mục 1.1 và 1.5 báo cáo trước đều đã được khắc phục trước khi đo.

### 1.1. Mức `-j5` — chiếm dụng 50%

| Bộ lập lịch | Makespan trung vị | Chênh lệch | p (sau Holm) | Effect size |
|---|---:|---:|---:|---|
| **Hybrid-LinUCB** | **48,12 s** | — | — | — |
| LeastLoaded | 49,61 s | −1,11 s | **0,042** | −0,58 (lớn) |
| icecream | 49,73 s | −1,68 s | **0,020** | −0,78 (lớn) |
| P2C | 53,37 s | −5,02 s | **0,004** | −1,00 (lớn) |
| LinUCB thuần | 57,72 s | −9,32 s | **0,004** | −1,00 (lớn) |

Kiểm định Friedman: χ² = 34,48; p < 0,00001 — các bộ lập lịch **khác nhau có ý nghĩa thống kê**.

Bộ lập lịch do chúng em đề xuất thắng **cả bốn** đối chứng, trong đó có LeastLoaded.

### 1.2. Mức `-j10` — chiếm dụng 100%

| Bộ lập lịch | Makespan trung vị |
|---|---:|
| LinUCB thuần | 51,18 s |
| Hybrid-LinUCB | 51,61 s |
| LeastLoaded | 51,84 s |
| P2C | 51,90 s |
| icecream | 52,31 s |

Kiểm định Friedman: χ² = 0,64; p = 0,959 — **không phân biệt được**. Mọi phép so sánh ghép cặp đều cho p ≥ 0,86 sau hiệu chỉnh. Toàn bộ năm bộ lập lịch nằm gọn trong dải 1,1 giây.

Đáng chú ý, LinUCB thuần từ chỗ chậm nhất ở `-j5` (57,72 s) trở thành nhanh nhất theo trung vị ở `-j10` (51,18 s) — tuy chênh lệch không đạt ý nghĩa thống kê.

---

## 2. Đính chính lập luận ở mục 1.2 báo cáo ngày 09/08

Báo cáo trước khẳng định:

> *"Khi cụm luôn dư năng lực, mọi bộ lập lịch có ưu tiên máy rảnh đều đạt cùng một makespan tối ưu. Trần hiệu năng của các thuật toán vì thế bằng nhau, và về nguyên tắc không bộ lập lịch nào có thể vượt LeastLoaded."*

**Khẳng định này sai, và dữ liệu ở mục 1.1 bác bỏ trực tiếp.** Tại đúng mức chiếm dụng 50% mà báo cáo trước mô tả là "bài toán quá dễ, không có gì để tối ưu", bộ lập lịch đề xuất vượt LeastLoaded 1,11 giây với p = 0,042.

Chỗ sai của lập luận là **đánh đồng "máy rảnh" với "máy tương đương"**. Năm worker trong cụm chênh nhau tới 2,2 lần về năng lực. Khi có nhiều máy rảnh, vẫn còn một quyết định thực sự phải ra: chọn máy rảnh **nhanh** hay chọn máy rảnh bất kỳ. LeastLoaded chọn theo số tác vụ đang chạy, không theo năng lực — nên nó trả lời câu hỏi đó chỉ ở mức tình cờ. Bộ lập lịch có ngữ cảnh, sau khi đặc trưng CPU được sửa, trả lời tốt hơn.

Hệ quả với cách diễn giải kết quả:

- Thế hoà quan sát được trước đây **không phải** do cụm dư tải (mục 1.1 báo cáo trước), mà do **đặc trưng CPU bị thoái hoá** (mục 1.5). Trong hai nguyên nhân đã nêu, chỉ nguyên nhân thứ hai là thật.
- Cụm dư tải không làm bài toán mất hết nội dung; nó chỉ làm mất *một phần* nội dung — phần liên quan tới xếp hàng — trong khi phần liên quan tới **tính không đồng nhất** vẫn còn nguyên.

Và kết quả ở mục 1.2 cho thấy điều ngược với dự đoán của báo cáo trước: **ở mức bão hoà 100%, lập lịch mới thật sự trở nên vô nghĩa.** Khi mọi worker đều bận liên tục, tác vụ đi đâu là do vị trí nào trống trước, không còn gì để chọn. Giá trị của lập lịch có ngữ cảnh nằm ở vùng **có dư năng lực và máy không đồng nhất**, chứ không phải vùng bão hoà.

---

## 3. Cơ chế giải thích chênh lệch

Chúng em tái dựng trạng thái cụm tại đúng thời điểm ra mỗi quyết định giao việc: mỗi worker bận trong khoảng từ lúc nhận việc tới lúc hoàn tất, giao các khoảng đó với từng mốc quyết định thì biết được bao nhiêu máy đang rảnh. Định nghĩa **tỉ lệ bỏ qua máy rảnh** là tỉ lệ số quyết định giao việc vào một máy đang bận trong khi vẫn còn ít nhất một máy hoàn toàn rảnh.

Tại `-j5`, **100% quyết định đều có sẵn máy rảnh** dưới cả năm bộ lập lịch. Trong chế độ đó, bỏ qua máy rảnh là loại lỗi duy nhất còn khả năng xảy ra:

| Bộ lập lịch | Tỉ lệ bỏ qua máy rảnh | Makespan trung vị |
|---|---:|---:|
| LeastLoaded | 0,0% | 49,61 s |
| Hybrid-LinUCB | 0,3% | 48,12 s |
| icecream | 7,4% | 49,73 s |
| P2C | 23,1% | 53,37 s |
| LinUCB thuần | 34,7% | 57,72 s |

Hồi quy makespan theo tỉ lệ này, ở hai cấp độ tổng hợp:

| Cấp độ | Hệ số góc | R² |
|---|---:|---:|
| Từng lần chạy (n = 50) | 0,227 s / 1% | **0,860** |
| Trung bình theo chính sách (n = 5) | 0,247 s / 1% | 0,954 |

Hai cấp độ cho cùng hệ số góc, nên quan hệ này không phải sản phẩm của cách gộp số liệu. Nhân hệ số góc với dải biến thiên 34,7% được 8,6 giây, khớp với chênh lệch 9,32 giây đo được giữa Hybrid-LinUCB và LinUCB thuần.

**Một cảnh báo về phương pháp.** Tại `-j10`, nếu chỉ hồi quy trên 5 giá trị trung bình theo chính sách thì được r = −0,82 (R² = 0,674), tức có vẻ như *càng bỏ qua máy rảnh càng nhanh*. Nhưng hồi quy trên toàn bộ 50 lần chạy cho **R² = 0,005** — không có quan hệ nào. Con số 5 điểm kia là ảo giác sinh ra từ việc gộp trung bình trên một dải makespan chỉ rộng 1,1 giây và thuần nhiễu (Friedman p = 0,959). Chúng em nêu ra vì đây đúng là loại bẫy mà quy trình đo nghiêm ngặt được lập ra để tránh, và vì nó cho thấy **không nên tin một hệ số tương quan tính trên số điểm quá ít**, kể cả khi nó ủng hộ kết luận của mình.

---

## 4. Giới hạn kiến trúc: không đo được vùng quá tải

Kế hoạch ở mục 4 báo cáo trước dự kiến đo thêm hai mức `-j16` (160%) và `-j24` (240%). **Cả hai đều không chạy được**, và nguyên nhân không phải cấu hình sai mà là một giới hạn thiết kế của hệ thống.

Khi mọi worker đều đã đầy vị trí, bộ điều phối thử lại việc giao tác vụ theo một ngân sách có giới hạn: tối đa 8 lần, mỗi lần chờ thêm 25 mili-giây, tổng cộng khoảng **700 mili-giây**, sau đó báo lỗi và `make` dừng. Chú thích trong mã nguồn cho thấy ngân sách này được đặt dựa trên giả định thời gian biên dịch nằm trong khoảng 10–500 mili-giây. Nhưng phân vị P99 đo được thực tế là **6,6–8,4 giây ở `-j5` và 12,4–13,2 giây ở `-j10`** (trung vị trên mười khối, tuỳ bộ lập lịch) — lớn hơn giả định từ mười tới vài trăm lần.

Ở `-j16` với 10 vị trí, luôn có 6 tác vụ phải chờ; ngân sách 700 mili-giây cạn ngay trong giai đoạn khởi động và bản build hỏng.

Nói cách khác, hệ thống hiện **chưa có cơ chế điều tiết ngược thực sự** — chỉ có một vòng thử lại ngắn. Nó phục vụ được tải tới đúng tổng năng lực cụm, nhưng không chịu được quá tải bền vững.

Chúng em xem đây là một **phát hiện có giá trị** chứ không phải một trở ngại phải giấu: nó xác định chính xác biên mà kết quả của chúng em có hiệu lực, và chỉ ra hạng mục kỹ thuật cần làm trước nếu muốn nghiên cứu vùng quá tải — thay vòng thử lại bằng một hàng đợi nhận việc thật, đặt tham số dựa trên P99 đo được.

Vì vậy nghiên cứu này chỉ báo cáo hai điểm vận hành: 50% và 100%. Chúng em không đưa ra khẳng định nào về vùng trên 100%.

---

## 5. Kiểm chứng chéo trong nhóm

Bạn Nguyễn Trung Kiên đã chạy một loạt đo độc lập trên máy khác và viết lại bài báo dựa trên đó. Chúng em đối chiếu hai bộ số liệu:

| So với | Số liệu của em | Số liệu của bạn Kiên |
|---|---:|---:|
| LinUCB thuần | +9,32 s | +8,95 s |
| P2C | +5,02 s | +4,66 s |
| LeastLoaded | +1,11 s (p = 0,042) | +1,12 s (p = 0,053) |

Makespan tuyệt đối lệch khá nhiều giữa hai máy (48–50 s so với 40–42 s), nhưng vì thiết kế khối ngẫu nhiên hoá dùng kiểm định **ghép cặp**, chênh lệch nền bị triệt tiêu và độ lớn hiệu ứng trùng nhau trong khoảng 0,4 giây. Đây là một lần **lặp lại độc lập thành công**, có giá trị hơn một lần đo đơn lẻ dù nhiều mẫu tới đâu.

Riêng phép so sánh với LeastLoaded rơi vào hai bên ngưỡng 0,05 — hai bộ số liệu cách nhau **đúng một bậc hạng** trong kiểm định Wilcoxon. Chúng em vì vậy sẽ phát biểu kết quả này một cách thận trọng là *"chưa được xác lập chắc chắn"* thay vì tuyên bố đã chứng minh, và trình bày cả hai bộ dữ liệu trong bài.

Chúng em cũng đã tái lập độc lập bảng cơ chế của bạn Kiên và thu được cùng kết quả (100% quyết định có máy rảnh tại `-j5`; tỉ lệ bỏ qua 0,0 / 0,3 / 23,1 / 34,7%). Công cụ tái dựng được kiểm chứng bằng cách so với bộ đếm mà bộ điều phối ghi trực tiếp lúc chạy: **khớp 99,8–100%**.

---

## 6. Các hạng mục còn tồn đọng

1. ~~Kiểm tra trực tiếp trên cụm rằng 5 worker báo đúng 500 / 600 / 800 / 1000 / 1100 milli-core.~~ **Đã xong 18/08/2026.**
2. ~~Chạy lại thí nghiệm ở các mức tải sau khi khắc phục hai nguyên nhân.~~ **Đã xong 21/08/2026** — hai mức 50% và 100%; hai mức còn lại không khả thi, lý do ở mục 4.
3. Viết lại phần đánh giá của bài báo theo hai điểm vận hành, thay cho cách chia workload nhẹ/nặng trước đây.
4. Bổ sung icecream vào bảng kết quả chính. Hiện thuật toán đã được cài đặt và đã chạy trong mọi khối đo, nhưng bản thảo hiện tại chưa báo cáo — trong khi đây là đối chứng có ý nghĩa nhất vì nó là hệ thống đang được dùng trong thực tế.
5. Nếu muốn nghiên cứu vùng quá tải: thay vòng thử lại bằng hàng đợi nhận việc thật ở bộ điều phối. Đây là thay đổi thiết kế, không phải chỉnh tham số.

**Về tính tái lập.** Toàn bộ số liệu trong báo cáo này sinh ra từ hai công cụ đã lưu trong kho mã nguồn, chạy trên dữ liệu thô của 100 lần chạy:

```
python scripts/analyze_rigorous.py <thư-mục-dữ-liệu>/j5
python scripts/analyze_dispatch_availability.py <thư-mục-dữ-liệu>/j5 --warm-start 100
```

Công cụ thứ hai in kèm một cột đối chiếu giữa phép tái dựng ngoại tuyến và bộ đếm ghi trực tiếp lúc chạy, để người đọc kiểm chứng được phép tái dựng thay vì phải tin.

---

Chúng em nhận thấy việc phải đính chính một lập luận mình vừa trình bày cách đây mười hai ngày là điều đáng tiếc. Tuy nhiên chúng em cho rằng báo cáo đúng vẫn tốt hơn báo cáo nhất quán, và bản thân quá trình từ giả thuyết tới bác bỏ bằng dữ liệu cũng là một phần kết quả đáng ghi lại trong luận văn.

Kính mong thầy cho ý kiến, đặc biệt về hai điểm: cách phát biểu kết quả so với LeastLoaded khi hai bộ dữ liệu nằm hai bên ngưỡng ý nghĩa, và việc có nên đưa giới hạn kiến trúc ở mục 4 vào bài báo như một đóng góp hay chỉ ghi ở phần hạn chế.

Em cảm ơn thầy ạ.
