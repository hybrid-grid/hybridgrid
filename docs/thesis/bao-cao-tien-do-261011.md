# Báo cáo tiến độ: thực nghiệm bổ sung HG-LinUCB (11/10/2026)

Tài liệu gom bốn phần:

- **A.** Kịch bản báo cáo tiến độ cho thầy (nói khoảng 10 phút).
- **B.** Các bài báo được nhắc trong yêu cầu của thầy và liên hệ với hệ thống.
- **C.** Vì sao Thí nghiệm 1 và 2 chạy lâu, kèm tính toán thời gian.
- **D.** Bảng thống kê số lượng của Thí nghiệm 1.
- **E.** Ghi chú cho người trình bày (việc còn mở).

Nguồn số liệu: `D:\bench-work\out\official\` (dữ liệu thô và phân tích), bản đăng ký trước `docs/thesis/hypotheses.md` (commit đóng băng `1b8339d`). Số liệu Thí nghiệm 1 là số đo thật; số liệu Thí nghiệm 2 chỉ là dự tính từ các cell thử.

---

# A. Kịch bản báo cáo tiến độ

*Phần in nghiêng là gợi ý thao tác khi trình bày, không phải lời nói.*

## A1. Mở đầu: nhắc lại thầy đã yêu cầu gì (khoảng 1 phút)

Thưa thầy, em báo cáo tiến độ đợt thực nghiệm bổ sung mà thầy giao.

Em nhắc lại vấn đề gốc. Ở bản trước, HG-LinUCB **hòa** với LeastLoaded. Thầy nhận xét nguyên nhân là môi trường thử quá dễ cho LeastLoaded: lúc nào cũng có máy rảnh, task nhỏ và đều, không có biến động. Thầy yêu cầu dựng các kịch bản thực tế hơn, nơi LeastLoaded có điểm yếu, rồi xem HG-LinUCB có thắng không, và dù thắng hay thua đều báo cáo đầy đủ.

Thầy đặt ra sáu thí nghiệm, làm theo thứ tự, và dặn xong **thí nghiệm 1 và 2 thì gửi thầy xem trước**. Hôm nay em báo cáo đến đúng mốc đó: thí nghiệm 1 đã chạy xong, thí nghiệm 2 đã sẵn sàng nhưng chưa chạy.

## A2. Trước khi làm: phân tích điểm yếu của LeastLoaded (khoảng 1 phút)

Em không làm ngay mà phân tích trước: LeastLoaded yếu ở đâu?

*Ví dụ dễ hiểu:* LeastLoaded giống một người **luôn chọn quầy thu ngân ít khách nhất**. Nghe hợp lý, nhưng có hai chỗ mù:

- Nó chỉ **đếm** số khách, không biết quầy nào **làm nhanh hay chậm**. Một quầy có người mới vào việc vẫn được xếp khách như quầy có người giỏi.
- Nó không thấy trước nếu một quầy **đột nhiên chậm đi** mà số khách vẫn như cũ.

Từ đó em xác định hai thí nghiệm đáng làm trước, đúng như thầy sắp xếp: thí nghiệm 1 (tải vừa phải, khi gần hết máy rảnh) và thí nghiệm 2 (một máy bị chậm đi mà hệ thống không biết).

## A3. Hai thành phần đã cài thêm (khoảng 2 phút)

**Thứ nhất, baseline SED**, theo công thức thầy đưa: chọn máy có (số task đang chạy + 1) chia cho công suất là nhỏ nhất, công suất đọc từ cgroup.

*Ví dụ:* máy A mạnh gấp đôi máy B, mỗi máy đang chạy 1 task. Điểm của A là 2 chia 2 bằng 1, điểm của B là 2 chia 1 bằng 2, nên SED chọn A. LeastLoaded thì thấy hai máy bằng nhau. Trong lúc review em phát hiện một lỗi của chính mình: khi số task bị đếm âm thì SED chọn sai. Em đã sửa và thêm test.

**Thứ hai, HG-LinUCB-D**, bản có "quên dần" của thầy. Ý tưởng: bandit học tốc độ từng máy, nhưng nếu máy đó **đổi tốc độ**, những gì đã học trước đó trở nên lạc hậu. Hệ số γ làm cho dữ liệu cũ nhẹ dần.

*Ví dụ:* một máy chạy nhanh suốt 200 task, rồi bị chậm đi. Không có "quên", bandit vẫn tin 200 task cũ và còn gửi việc cho máy đó rất lâu. Có "quên", vài chục task chậm là đủ làm nó đổi ý.

Công thức thầy ghi có một chỗ mơ hồ, nên em **cài cả hai cách đọc**: (a) mọi máy cùng "quên" theo mỗi quyết định, đây là bản chính; (b) đọc đúng chữ, chỉ máy vừa nhận kết quả mới cập nhật. Em chạy cả hai và đề nghị thầy cho ý kiến ở cuối buổi.

Cả hai phần đều được kiểm tra kỹ: có bộ test so với kết quả cũ từng bit, 31 lỗi cố ý được cài vào đều bị test bắt, và chạy kiểm tra đua luồng (race detector) sạch.

## A4. Chuẩn bị để thí nghiệm đo đúng (khoảng 1 phút)

Em bổ sung công cụ để mọi con số trong báo cáo kiểm chứng được:

- Log ghi **số thứ tự dispatch, giờ dispatch, và ảnh chụp trạng thái cả cụm** ngay lúc quyết định. Nhờ vậy em tính **chính xác** việc "gửi vào máy bận trong khi có máy rảnh", chứ không phải ước đoán như trước.
- Các sự kiện bơm vào (làm chậm máy) cũng được **ghi giờ vào log**, và luôn xảy ra sau dispatch thứ 100, đúng quy định chung của thầy.
- Mỗi thí nghiệm chạy **20 block** theo thứ tự ngẫu nhiên. Mỗi block chạy đủ cả bốn scheduler. Em so sánh các scheduler **trong cùng một block**, nên nhiễu của máy ảnh hưởng đều cho cả bốn.
- Bảng kết quả theo mẫu thầy yêu cầu: median, hiệu số, p-value Holm, và effect size ghép cặp (rank-biserial).

Em cũng **viết trước và khóa** file `hypotheses.md`: giả thuyết, metric, cấu hình, cách xử lý cell lỗi. Nó được đóng băng ở commit `1b8339d` trước khi chạy và sẽ không sửa sau khi có kết quả. Em còn có một Codex độc lập làm "người giám sát" trong lúc chạy, cứ 90 phút kiểm tra dữ liệu một lần.

## A5. Kết quả thí nghiệm 1: tải trung gian (khoảng 3 phút)

*Chiếu bảng này.*

Em chạy CPython với `-j6`, `-j7`, `-j8`, `-j9`, thêm một kịch bản **hai client build cùng lúc**. Tổng cộng 400 lượt chạy (480 build), mất 14,5 giờ, **không lượt nào lỗi**.

Median makespan (giây):

| Kịch bản | HG-LinUCB | HG-LinUCB-D | LeastLoaded | SED |
|---|---|---|---|---|
| `-j6` | 89,6 | 93,0 | 91,2 | 96,3 |
| `-j7` | 87,5 | 90,7 | 89,8 | 86,4 |
| `-j8` | 86,4 | 87,5 | 87,4 | 85,7 |
| `-j9` | 85,0 | 84,9 | 84,5 | 85,0 |
| 2 client | 165,3 | 165,6 | 165,2 | 164,8 |

**Giả thuyết H1.1: "ở tải cao, HG-LinUCB nhanh hơn LeastLoaded". Kết quả: không được ủng hộ.** Ở `-j8`, HG-LinUCB nhanh hơn 0,84 giây nhưng chưa đạt ý nghĩa thống kê (p_Holm = 0,098). Ở `-j9` và hai client thì không khác gì. Điểm đáng chú ý: **khi tải tăng, bốn scheduler càng giống nhau**. Chênh lệch giữa scheduler nhanh nhất và chậm nhất là 6,7 giây ở `-j6`, nhưng chỉ 0,5 giây ở `-j9`.

**Giả thuyết H1.2: "HG-LinUCB chọn đúng máy rảnh mạnh nhất nhiều hơn LeastLoaded". Kết quả: ủng hộ một phần.** Ví dụ ở `-j8`: HG-LinUCB chọn đúng 94% so với 91% của LeastLoaded (p = 0,003). Đạt ở 3 trong 5 kịch bản, không đạt ở `-j7` và hai client, nên chưa thể nói "ở mọi kịch bản" như đã đăng ký.

*Giải thích "máy rảnh":* em tính máy rảnh khi nó **không có task nào đang chạy** (đúng định nghĩa Bảng 5 của bài). Một máy có ba chỗ chạy song song mà đang chạy một task thì em tính là bận. Còn LeastLoaded luôn chọn máy rảnh khi có, nên nó đạt đúng 0% "bỏ qua máy rảnh". Vì vậy chỉ số đó không thể cho thấy HG-LinUCB hơn, mà chỉ cho biết có kém hơn không. Chỉ số "chọn đúng máy rảnh *mạnh nhất*" mới phân biệt được hai bên, vì LeastLoaded không phân biệt máy mạnh hay yếu.

**Giả thuyết H1.3: "SED hay bỏ qua máy rảnh hơn LeastLoaded". Kết quả: ủng hộ ở cả 5 kịch bản.** SED dồn việc vào máy mạnh dù có máy yếu đang rảnh: 48% quyết định ở `-j6`, nhưng chỉ 1% ở hai client.

**Vì sao tải tăng thì các scheduler giống nhau?** Em đưa ra như một giả thuyết, chưa kiểm chứng: tỉ lệ quyết định có máy hoàn toàn rảnh giảm từ 77% ở `-j6` xuống chỉ 15% ở hai client. Còn rất ít chỗ để một scheduler thông minh hơn tạo khác biệt. Em báo cáo đây là kết quả **hòa hoặc thua** đúng như thầy dặn.

## A6. Thí nghiệm 2 và việc tiếp theo (khoảng 1 phút)

Thí nghiệm 2, drift ẩn, mới là chỗ LeastLoaded có điểm yếu rõ hơn. Em đã chạy thử từng biến thể một lần để kiểm tra công cụ:

- **Làm chậm bằng `stress-ng`** chạy *bên trong* container của máy 1,1 CPU, đúng như thầy dặn (không đổi hạn mức CPU).
- Ba biến thể: chậm vĩnh viễn, chậm 30 giây rồi hồi phục, bật/tắt.
- Em phát hiện một điều: nếu chạy một build thì sau khi hồi phục chỉ còn khoảng 45 quyết định để quan sát. Em chọn **hai build liên tiếp** trên cùng coordinator để có khoảng 340 quyết định.

Ước lượng thời gian: thí nghiệm 2 khoảng 25 giờ, cộng thêm quét λ khoảng 12 giờ.

## A7. Những điểm cần thầy cho ý kiến (khoảng 1 phút)

1. **Hai bản "quên dần":** thầy muốn bản (a) mọi máy cùng quên, hay (b) đọc đúng chữ công thức, làm bản chính?
2. **Thí nghiệm 2 chạy hai build liên tiếp**, khác với giao thức cũ (reset sau mỗi build). Thầy thấy có hợp lý không?
3. **Log thô khoảng 400–500 MB** cho cả đợt (khoảng 50 MB khi nén): thầy muốn đưa lên repo ở dạng nén, dùng Git LFS, hay gửi riêng?

## A8. Những điều cần nói thẳng (nếu thầy hỏi)

- *Các con số có so với bài báo cũ được không?* Không trực tiếp. Em chạy trên máy khác chậm hơn khoảng 2,6–3 lần về biên dịch, và nâng builder từ 1,0 lên 2,0 CPU vì 1,0 CPU làm nghẽn phía client (chạm hạn mức ở 25% số mẫu đo). Em dùng so sánh **giữa các scheduler trong cùng thí nghiệm**, không so tuyệt đối với bài.
- *Chênh lệch dưới 1–2 giây có nghĩa gì không?* Hầu hết không đạt ý nghĩa thống kê, em đã nói rõ ở từng kịch bản.
- *Còn hạn chế nào?* Đặc trưng thứ 9 (độ trễ RPC) của HG-LinUCB đang là hằng số, nên coi là hạn chế đã biết. Hệ số phạt tải λ=0,5 có thể làm phần "học" ít có tác dụng, nên cần quét λ để diễn giải.
- *Tình trạng dữ liệu:* 400 lượt chạy (480 build), không lượt nào lỗi, một commit duy nhất, một bộ giám sát độc lập kiểm tra 12 lần.

---

# B. Các bài báo được nhắc trong yêu cầu của thầy

Yêu cầu của thầy chỉ nêu đích danh **một** bài báo: Russac, Vernade, Cappé (NeurIPS 2019), dùng cho HG-LinUCB-D. Các phần khác thầy nhắc ("Section 7.4", "Table 3", "Table 5") là tham chiếu đến bài báo của chính nhóm. SED thầy chỉ nêu công thức, không dẫn nguồn. Phần tóm tắt dưới đây đối chiếu với bản PDF của bài.

## B1. Russac, Vernade, Cappé: "Weighted Linear Bandits for Non-Stationary Environments"

**Bài toán.** Bandit tuyến tính có ngữ cảnh, nhưng tham số thật θ* **thay đổi theo thời gian**. LinUCB thường giả định θ* cố định, nên khi nó đổi thì dữ liệu cũ trở nên lạc hậu mà thuật toán vẫn tin.

**Ý tưởng.** Dùng hồi quy bình phương tối thiểu có **trọng số mũ**: quan sát cách đây s bước có trọng số γ^s, nên quên dần một cách mượt. Thuật toán D-LinUCB cập nhật mỗi bước:

- V ← γV + AAᵀ + (1−γ)λI
- b ← γb + X·A, và θ̂ = V⁻¹b

Đây chính là công thức thầy ghi trong yêu cầu, với λ=1.

**Kết quả lý thuyết.**

- Xử lý được cả môi trường thay đổi chậm lẫn thay đổi đột ngột.
- Chặn trên hối tiếc động bậc O(d^{2/3}·B_T^{1/3}·T^{2/3}), với B_T = Σ‖θ_s − θ_{s+1}‖ là "ngân sách biến động". Bậc này khớp cận dưới đã biết.
- Điều kiện: chọn γ = 1 − (B_T/(dT))^{2/3}, tức phải biết trước B_T và T. Bài thừa nhận đây là điểm mở.
- Ưu điểm so với cửa sổ trượt (SW-LinUCB) hay phát hiện điểm đổi: **không cần lưu lại các hành động cũ**, chạy hoàn toàn trực tuyến.

## B2. Bài nền tảng liên quan: Garivier & Moulines

Thầy không nêu bài này, nhưng Russac trích dẫn nó và nhóm cũng đã tải. Bài này xét bandit nhiều tay (không ngữ cảnh) khi phân phối phần thưởng đổi tại các thời điểm không biết trước, phân tích hai thuật toán **Discounted-UCB** và **Sliding-Window UCB**, và chứng minh cả hai gần chạm cận dưới (sai khác hệ số logarit). Nó là cơ sở cho ý tưởng "mọi tay cùng bị quên theo thời gian".

## B3. Liên hệ với hệ thống của nhóm

| Trong bài | Trong hệ thống |
|---|---|
| Tham số θ* đổi theo thời gian | Tốc độ thật của một worker đổi (stress-ng làm worker 1,1 CPU chậm đi trong khi hạn mức CPU vẫn như cũ) |
| Đổi đột ngột | Biến thể chậm vĩnh viễn, chậm 30 s rồi hồi phục, bật/tắt |
| Công thức V ← γV + AAᵀ + (1−γ)λI | `hybrid-linucb-d`: cùng công thức, γ ∈ {0,95; 0,98; 0,99} |
| Quên theo mỗi bước thời gian | **Đồng hồ global**: mọi worker cùng quên theo mỗi quyết định (bản chính, giống D-UCB của Garivier) |
| Công thức áp dụng cho arm nào? | **Đồng hồ per-arm**: chỉ worker vừa nhận kết quả cập nhật (bản đọc đúng chữ) |
| Không cần lưu quá khứ | Đúng, chỉ giữ A và b, cập nhật theo công thức |

**Ví dụ để hình dung γ.** Một quan sát cách đây 100 quyết định còn trọng số 0,98¹⁰⁰ ≈ 0,13 (γ=0,98), 0,95¹⁰⁰ ≈ 0,006 (γ=0,95), và 0,99¹⁰⁰ ≈ 0,37 (γ=0,99). Nửa trọng số mất sau khoảng 34 quyết định ở γ=0,98, 14 quyết định ở γ=0,95 và 69 quyết định ở γ=0,99. Với phiên thí nghiệm khoảng 590 quyết định và drift ở khoảng quyết định 124, γ=0,95 có thể quên quá nhanh, γ=0,99 có thể quên quá chậm. Đó là lý do quét ba mức.

## B4. Chỗ hệ thống khác bài, nên lý thuyết không áp dụng nguyên vẹn

1. **Mô hình khác.** Bài dùng một θ chung cho mọi hành động; hệ thống dùng thống kê **riêng cho từng worker**.
2. **Thăm dò khác.** Bài có bán kính tin cậy chặt dạng √(aᵀV⁻¹ṼV⁻¹a) với Ṽ xây bằng γ². Hệ thống dùng dạng đơn giản α√(xᵀA⁻¹x) theo yêu cầu của thầy, nên **không tuyên bố cận hối tiếc của bài**.
3. **Phản hồi đến muộn.** Bài giả định phần thưởng nhận ngay sau mỗi bước. Hệ thống có nhiều task chạy song song, kết quả về trễ, nên thêm trọng số γ^(thời điểm nhận − thời điểm quyết định). Đây là mở rộng của nhóm, không có trong bài.
4. **Phần ngoài bandit.** Hệ số phạt tải λ·LoadRatio và 100 quyết định đầu giao cho LeastLoaded không có trong mô hình lý thuyết.
5. **γ không chọn theo công thức của bài** vì không biết B_T. Nhóm quét γ thay vì tính.

## B5. Một chỗ cần sửa trong tài liệu của nhóm

`docs/thesis/theory-notes.md` mục 5.3 ghi rằng cả cửa sổ trượt lẫn **quên theo hàm mũ đều "không có bảo đảm chặt trong trường hợp tuyến tính"**. Điều này không còn đúng sau Russac et al. 2019, vì bài đó chứng minh đúng bảo đảm cho quên theo hàm mũ. Nên cập nhật mục này và mục hạn chế trong luận văn.

## B6. Về SED

SED (Shortest Expected Delay) là một heuristic chia tải kinh điển cho các máy không đồng đều. Phần nguồn gốc (có trong Linux Virtual Server) chỉ là hiểu biết chung, chưa tra trong tài liệu nhóm; nếu cần trích dẫn trong bài thì phải tra nguồn trước.

---

# C. Vì sao Thí nghiệm 1 và 2 chạy lâu

Hai thí nghiệm lâu vì mỗi lượt đo là một lần build CPython thật (khoảng 1,5–3 phút) và phải chạy **tuần tự** hàng trăm lần. Thí nghiệm 1 đã đo thật là **14,5 giờ**. Thí nghiệm 2 chưa chạy, các con số là dự tính từ các cell thử.

## C1. Một "lượt đo" mất bao lâu và mất vào đâu

Một lượt đo (cell) là cho **một scheduler build CPython một lần** trên cluster giả lập.

*Hình dung:* như chạy thử một chiếc xe trên đường đua để bấm giờ. Muốn so công bằng thì xe nào cũng phải chạy một mình trên đường vắng, chạy xong mới đến xe khác. Hai xe chạy cùng lúc sẽ cản nhau và số đo hỏng.

Theo số đo thật của TN1 (một client, `-j8`):

| Phần | Thời gian | Việc gì |
|---|---|---|
| Build | ~85–93 giây | Biên dịch 295 file C của CPython |
| Chuẩn bị và dọn | ~25 giây | Khởi động 7 container, chờ 5 worker đăng ký, xóa cache và `make clean`, lấy log, nghỉ 8 giây giữa hai lượt |
| **Một cell** | **~110–118 giây** | khoảng 79% là build thật |

**Vì sao một build lâu thế?** Cluster cố ý nhỏ: 5 worker có tổng cộng chỉ 4 CPU (0,5 + 0,6 + 0,8 + 1,0 + 1,1) và 10 chỗ chạy song song, để mô phỏng cụm máy không đồng đều. 295 file chia cho 4 CPU thì không thể xong trong vài giây. Máy chạy chính thức cũng biên dịch chậm hơn máy pilot khoảng 2,6–3,1 lần, nên một build ~90 giây ở đây, so với ~37 giây ở máy pilot.

## C2. Vì sao phải chạy hàng trăm lượt

> Tổng thời gian = số kịch bản × số scheduler × số block × thời gian một cell

Số block là **20** theo yêu cầu thầy: mỗi block cho mọi scheduler chạy một lần theo thứ tự ngẫu nhiên, rồi so sánh các scheduler *trong cùng một block*. Có 20 block thì kiểm định thống kê mới đủ sức. Với vài block thôi thì p nhỏ nhất đạt được đã không dưới 0,05 sau hiệu chỉnh.

## C3. Tính cho Thí nghiệm 1 (đã chạy xong, số đo thật)

| Kịch bản | Số cell | Thời gian một cell | Tổng |
|---|---|---|---|
| `-j6` | 4 scheduler × 20 = 80 | ~118 giây | 2,6 giờ |
| `-j7` | 80 | ~114 giây | 2,5 giờ |
| `-j8` | 80 | ~112 giây | 2,5 giờ |
| `-j9` | 80 | ~110 giây | 2,5 giờ |
| 2 client | 80 | ~196 giây | 4,4 giờ |
| **Cộng** | **400 cell** | | **14,5 giờ** (15:56 đến 06:25) |

Kịch bản hai client lâu hơn vì mỗi cell có **hai build chạy cùng lúc** tranh cùng cluster: mỗi build mất ~165 giây thay vì ~85 giây.

## C4. Dự tính cho Thí nghiệm 2 (chưa chạy)

Mỗi cell của TN2 là **hai build liên tiếp** (phiên ~590 quyết định). Cell thử đo được 101 + 98 = 199 giây cộng ~19 giây dọn, tức ~218 giây.

| Nhóm | Phép tính | Số cell | Thời gian |
|---|---|---|---|
| 3 biến thể (chậm vĩnh viễn, 30 giây, bật/tắt) | 3 × 7 scheduler × 20 | 420 | 420 × 218 giây ≈ **25,4 giờ** |
| Quét λ | 10 scheduler × 20 | 200 | 200 × 218 giây ≈ **12,1 giờ** |
| **TN2 cộng** | | **620 cell** | **≈ 37,5 giờ** |

Có thêm khoảng 27 phút cho 9 lần build làm nóng (một lần mỗi kịch bản). Tổng toàn bộ 14,5 + 37,5 ≈ **52 giờ**.

TN2 có nhiều scheduler hơn TN1 (7 so với 4) vì phải thêm 3 mức γ (0,95; 0,98; 0,99) và bản đồng hồ per-arm. Quét λ thêm bốn mức λ cho cả HG-LinUCB lẫn HG-LinUCB-D, thành 10.

## C5. Vì sao không chạy nhanh hơn được

- **Không chạy song song được dễ dàng.** Máy có 72 luồng nhưng cluster chỉ dùng ~6 CPU. Về lý thuyết chạy vài cluster cùng lúc được, nhưng chúng tranh cache và băng thông bộ nhớ, làm lệch số đo. Muốn làm thì phải ghim mỗi cluster vào nhân riêng và kiểm chứng lại, chưa làm.
- **Không cắt bớt block.** 20 block là yêu cầu của thầy.
- **Không tăng CPU cluster.** Đổi cluster là đổi thí nghiệm.
- **Chạy ở máy cũ nhanh hơn khoảng 2,5–3 lần,** tức ~23 giờ cho toàn bộ, nhưng máy đó không còn ở đây.

## C6. Nếu cần nói gọn với thầy

> "Thưa thầy, mỗi lần đo là một lần build CPython thật, mất khoảng một phút rưỡi đến hai phút. Em phải đo bốn đến bảy scheduler, mỗi cái hai mươi lần theo thiết kế block của thầy, ở nhiều kịch bản, và chúng phải chạy lần lượt để không cản nhau. Thí nghiệm 1 gồm 400 lượt, đã chạy 14 tiếng rưỡi. Thí nghiệm 2 gồm khoảng 620 lượt, mỗi lượt hai build liên tiếp, dự tính khoảng 37 tiếng. Nếu thầy muốn rút ngắn, em có thể quét λ sau hoặc bớt một số mức γ, nhưng sẽ mất một phần đối chiếu."

---

# D. Thống kê số lượng của Thí nghiệm 1

Lấy từ dữ liệu thật (`results.csv`, task log, `campaign.log`).

## D1. Thiết kế

| Thông số | Giá trị |
|---|---|
| Số kịch bản | **5** (`-j6`, `-j7`, `-j8`, `-j9`, hai client `-j8`) |
| Scheduler mỗi kịch bản | **4** (LeastLoaded, SED, HG-LinUCB, HG-LinUCB-D γ=0,98) |
| Số block mỗi kịch bản | **20** |
| Workload | CPython v3.14.0, 295 file mỗi build |
| Cluster | 5 worker (0,5 / 0,6 / 0,8 / 1,0 / 1,1 CPU), 10 chỗ chạy song song, builder 2,0 CPU |

## D2. Số lượt chạy thực tế theo từng kịch bản

| Kịch bản | Cấu hình | Scheduler | Block | Cell | Build | Task biên dịch | Thời gian | Giây mỗi cell |
|---|---|---|---|---|---|---|---|---|
| E1-j6 | `-j6`, 1 client | 4 | 20 | 80 | 80 | 23.600 | 2,63 giờ | 118 |
| E1-j7 | `-j7`, 1 client | 4 | 20 | 80 | 80 | 23.600 | 2,52 giờ | 114 |
| E1-j8 | `-j8`, 1 client | 4 | 20 | 80 | 80 | 23.600 | 2,50 giờ | 112 |
| E1-j9 | `-j9`, 1 client | 4 | 20 | 80 | 80 | 23.600 | 2,46 giờ | 111 |
| E1-2c | `-j8`, **2 client** | 4 | 20 | 80 | **160** | 47.200 | 4,35 giờ | 196 |
| **Tổng** | | | | **400** | **480** | **141.600** | **14,45 giờ** | |

Cách đọc:

- **Cell:** một scheduler build một lần (kịch bản × scheduler × block = 5 × 4 × 20 = 400).
- **Build:** ở kịch bản hai client, mỗi cell có hai build chạy cùng lúc, nên 80 cell cho 160 build.
- **Task biên dịch:** mỗi build có 295 task, tổng 480 × 295 = 141.600 quyết định điều phối đã được ghi lại.

```
Số cell     = 5 kịch bản × 4 scheduler × 20 block = 400
Số build    = 4 kịch bản × 80 + 1 kịch bản × 160  = 480
Số task     = 480 build × 295                      = 141.600
Thời gian   = 2,63 + 2,52 + 2,50 + 2,46 + 4,35     = 14,45 giờ
```

## D3. Đối chiếu với kế hoạch

| | Kế hoạch (ước tính) | Thực tế |
|---|---|---|
| Số cell | 400 | 400 |
| Thời gian | ~15 giờ | 14,45 giờ (15:56 ngày 10/10 đến 06:25 ngày 11/10) |
| Cell hỏng / chạy lại | | 0 / 0 |
| Block bị loại | | 0 (cả 5 kịch bản đủ 20/20 block) |
| Lần giám sát của Codex | | 12, trong đó 10 `OK`, 2 `WARN` lành tính, đã điều tra |

---

# E. Ghi chú cho người trình bày (việc còn mở)

- **Chưa push** nhánh `feat/exp-readiness` lên GitHub. Commit đóng băng `1b8339d` hiện chỉ có ở máy chạy thí nghiệm; nếu thầy hỏi bằng chứng về thời điểm đăng ký trước thì nên push trước buổi báo cáo.
- **Đợt 2 (Thí nghiệm 2) chưa chạy.** Muốn chạy phải tạo tệp `PHASE2_APPROVED` và quyết định **một lần** có chạy quét λ (E2-lambda) hay không (nếu không thì đặt `STOP_AFTER=E2-onoff`), trước khi có bất kỳ kết quả TN2 nào.
- **Ba câu hỏi gửi thầy** (mục A7): đồng hồ discount nào là bản chính, phiên hai build cho TN2, và nơi lưu log thô.
- **Cập nhật tài liệu:** sửa `theory-notes.md` mục 5.3 (xem B5) và mục hạn chế trong luận văn.
- Ví dụ "quầy thu ngân" và "máy ba chỗ" chỉ là minh họa. Mọi con số trong bảng kết quả là số đo thật từ dữ liệu.
