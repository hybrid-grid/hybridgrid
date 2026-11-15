# Báo cáo kết quả Thí nghiệm 2 (drift ẩn)

*Ngày 21/10/2026. Giả thuyết, metric và cách kiểm định được đăng ký trước trong
`docs/thesis/hypotheses.md` (commit `1b8339d`, đóng băng trước lượt chạy đầu tiên của TN1). Kế hoạch
vận hành TN2: `docs/thesis/ke-hoach-thi-nghiem-2.md`.*

*Nguồn số liệu:*
- *Bảng kiểm định và bảng mô tả: lấy từ `analyze_exp.py` (commit `1b8339d`). Một script viết riêng đã
  tính lại 70 dòng so sánh (median, hiệu ghép cặp, p, p_Holm, r) và khớp hoàn toàn; khoảng tin cậy
  bootstrap không được tính lại.*
- *Số lượng, thời gian chạy và sự cố: lấy từ `campaign.log` và `failures.csv`.*
- *Kiểm tra sự kiện và hình: lấy từ hai script khám phá có định nghĩa chốt trước khi mở kết quả
  (mục 9).*

---

## 1. Tóm tắt

- **TN2 chạy đủ:** 3 kịch bản × 7 scheduler × 20 block + 1 kịch bản × 10 scheduler × 20 block =
  **620 cell** (80 block, 1.240 build, 365.800 task), khoảng **37 giờ máy**. Có 1 cell hỏng, đã chạy lại
  theo quy tắc đăng ký; không block nào bị loại.
- **Drift có thật:** sau khi bật `stress-ng`, thời gian biên dịch trên mỗi 100 kB ở worker 1,1 CPU tăng
  **1,87 lần**, trong khi worker 1,0 CPU (đối chứng) không đổi (0,99 lần).
- **Theo các giả thuyết đăng ký:** không thành phần nào trong năm thành phần H2.1a–H2.3b đạt
  p_Holm < 0,05 theo hướng giả thuyết. *(Đây là tóm tắt để trình bày, không phải quy tắc phân loại đã
  đăng ký; từng thành phần được báo cáo riêng ở mục 3.)* H2.3 đúng hướng (HG-LinUCB-D nhanh hơn
  LeastLoaded khoảng 2 s) nhưng không có ý nghĩa sau Holm.
- **Quan sát ngoài giả thuyết** (so sánh có sẵn trong họ kiểm định đăng ký, đã hiệu chỉnh Holm, nhưng
  không phải giả thuyết):
  - **HG-LinUCB bản gốc nhanh hơn LeastLoaded** ở cả ba biến thể drift, hiệu ghép cặp 2,0–6,3 s,
    p_Holm ≤ 0,006. So với SED thì không phát hiện khác biệt có ý nghĩa.
  - Các bản HG-LinUCB-D **"mọi máy cùng quên"** với γ = 0,95 chậm hơn HG-LinUCB 5–6 s ở cả ba kịch bản.
  - Bản **"chỉ máy vừa nhận kết quả quên"** thì không phát hiện khác biệt có ý nghĩa so với HG-LinUCB.
- **Diễn giải chính (chưa kiểm chứng riêng):** LeastLoaded không "mù" với drift như giả định. Máy bị
  chậm giữ task lâu hơn, nên LeastLoaded tự gửi ít đi. Ở E2-perm, mọi scheduler gửi 18,7–20,5 % quyết
  định vào máy bị chậm sau sự kiện, gần mức tham chiếu theo tỷ lệ công suất (16,9 %). Còn ít chỗ để
  "học thích nghi" tạo khác biệt.

---

## 2. Thí nghiệm 2 đã chạy những gì

| | |
|---|---|
| Sự kiện | `stress-ng --cpu 1` chạy **bên trong** worker 1,1 CPU; cấu hình kích hoạt ở `DRIFT_AT=120`; sự kiện ghi nhận thực tế ở dispatch 127–132; hạn mức CPU không đổi |
| Phiên đo | 2 build CPython liên tiếp trên cùng coordinator (590 quyết định), `-j8`, 1 client |
| E2-perm | chậm vĩnh viễn đến hết phiên |
| E2-trans | chậm 30 s rồi hồi phục |
| E2-onoff | bật/tắt liên tục: danh nghĩa 10 s / 10 s; thời lượng quan sát được (median) 14,1 s bật, 14,1 s tắt; 5–6 lần bật mỗi phiên |
| Scheduler (E2-perm/trans/onoff) | LeastLoaded, SED, HG-LinUCB, HG-LinUCB-D γ = 0,95 / 0,98 / 0,99 ("mọi máy cùng quên"), HG-LinUCB-D γ = 0,98 ("chỉ máy vừa nhận kết quả quên") |
| E2-lambda | drift vĩnh viễn; HG-LinUCB và HG-LinUCB-D (γ = 0,98) với λ ∈ {0; 0,25; 0,5; 1,0}, cùng LeastLoaded và SED |
| Block | 20 mỗi kịch bản; thứ tự scheduler xáo trộn có seed trong từng block |

| Kịch bản | Cell | Thời gian chạy | Cell hỏng / block chạy lại | Ghi chú |
|---|---:|---:|---|---|
| E2-perm | 140 | 8,5 h | 0 / 0 | |
| E2-trans | 140 | 8,1 h | 0 / 0 | đúng 1 cặp bật/tắt ở 140/140 cell |
| E2-onoff | 140 | 8,3 h | 0 / 0 | |
| E2-lambda | 200 | 12,2 h | 1 / 0 | tạm dừng theo yêu cầu sau block 14 (19 → 21/10); cell đầu tiên sau khi chạy tiếp hỏng 1 lần rồi chạy lại thành công |
| **Tổng** | **620** | **≈ 37 h** | **1 / 0** | đăng ký ước tính ~37 h |

Mọi cell đủ 590 task thành công; sự kiện drift bật ở dispatch 127–132, trong khoảng hợp lệ
[101, 160]; không có makespan bất thường. Một giám sát viên độc lập (Codex) kiểm tra dữ liệu 90 phút một
lần (27 lần trong TN2), không phát hiện vấn đề ảnh hưởng dữ liệu.

---

## 3. Kết quả theo giả thuyết đăng ký (mẫu Table 3)

Cách tính:
- Wilcoxon signed-rank hai phía trên 20 block ghép cặp; p hiệu chỉnh Holm trong họ đã đăng ký (số so
  sánh trong họ ghi trong ngoặc).
- r là matched-pairs rank-biserial, âm khi scheduler thứ nhất có giá trị nhỏ hơn.
- Δ là median của hiệu ghép cặp (scheduler thứ nhất − scheduler thứ hai), kèm CI 95 % bootstrap.
- `n_zero` là số block có hiệu bằng 0 (Wilcoxon bỏ qua các block này).

Ký hiệu: D = HG-LinUCB-D γ = 0,98 "mọi máy cùng quên"; HG = HG-LinUCB; LL = LeastLoaded.

| GT | Kịch bản | Metric | So sánh (hướng dự đoán) | n / n_zero | Median | Δ [CI 95 %] | p thô | p_Holm (họ) | r | Hướng quan sát |
|---|---|---|---|---|---|---|---:|---:|---:|---|
| H2.1a | E2-perm | time-to-adapt (số quyết định) | D < HG | 20 / 1 | 45,5 vs 50 | +2,5 [−17,5; 11,5] | 0,794 | 1,00 (4) | −0,07 | lẫn lộn: median từng nhóm và r nghiêng về D, hiệu ghép cặp thì không |
| H2.1b | E2-perm | tỷ lệ gửi vào worker bị chậm sau sự kiện | D < HG | 20 / 0 | 0,190 vs 0,205 | −0,009 [−0,037; 0,016] | 0,409 | 0,82 (4) | −0,22 | đúng hướng, không có ý nghĩa |
| H2.2 | E2-trans | tỷ lệ gửi vào worker sau khi hồi phục | D > HG | 20 / 0 | 0,270 vs 0,270 | −0,003 [−0,046; 0,056] | 0,985 | 1,00 (4) | −0,01 | ngược hướng (rất nhỏ), không có ý nghĩa |
| H2.3a | E2-perm | makespan phiên (s) | D < LL | 20 / 0 | 188,5 vs 189,5 | −2,25 [−3,73; −0,40] | 0,036 | 0,22 (10) | −0,53 | đúng hướng, không có ý nghĩa sau Holm |
| H2.3b | E2-onoff | makespan phiên (s) | D < LL | 20 / 0 | 183,8 vs 184,6 | −1,72 [−2,62; 0,16] | 0,058 | 0,23 (10) | −0,49 | đúng hướng, không có ý nghĩa |
| H2.4 | E2-lambda | tỷ lệ sau sự kiện theo λ | mô tả | | xem mục 5 | | | không kiểm định | | khám phá |

Nhận xét:
- **Thích nghi:** không có bằng chứng thống kê rằng HG-LinUCB-D (γ = 0,98) thích nghi nhanh hơn, hay gửi
  ít hơn vào máy bị chậm, so với HG-LinUCB.
- **Hồi phục:** sau khi máy hồi phục (E2-trans), không phát hiện khác biệt có ý nghĩa giữa hai bản về tỷ
  lệ quay lại dùng máy đó. Median của cả hai đều là 27 %. Đây không phải kiểm định tương đương.
- **So với LeastLoaded:** HG-LinUCB-D nhanh hơn khoảng 2 s mỗi phiên (≈ 1 %) ở cả hai kịch bản, đúng
  hướng dự đoán. Nhưng H2.3 nằm trong họ 10 so sánh, nên sau Holm không đạt. Sức mạnh thống kê thấp của
  H2.3 đã được ghi là hạn chế từ trước khi chạy.
- **Kiểm duyệt time-to-adapt:** 1/140 cell ở E2-perm (γ = 0,95) không đạt ngưỡng trước khi hết phiên.
  Mọi dòng giả thuyết có n = 20, không có giá trị thiếu, số so sánh trong họ đúng như đăng ký.

---

## 4. Các so sánh khác trong họ đăng ký (quan sát mô tả, không phải giả thuyết)

Makespan phiên (s), median của 20 block:

| Scheduler | E2-perm | E2-trans | E2-onoff |
|---|---:|---:|---:|
| LeastLoaded | 189,5 | 179,0 | 184,6 |
| SED | 186,3 | 175,0 | 180,2 |
| HG-LinUCB | 184,1 | 174,2 | 181,2 |
| HG-LinUCB-D γ = 0,95 | 190,2 | 181,0 | 186,7 |
| HG-LinUCB-D γ = 0,98 | 188,5 | 178,9 | 183,8 |
| HG-LinUCB-D γ = 0,99 | 186,6 | 175,9 | 182,1 |
| HG-LinUCB-D γ = 0,98, "chỉ máy vừa nhận kết quả quên" | 184,1 | 175,2 | 181,1 |

Một số hiệu ghép cặp (Δ, p_Holm); **in đậm khi p_Holm < 0,05**:

| So sánh | E2-perm | E2-trans | E2-onoff |
|---|---|---|---|
| HG-LinUCB vs LL | **−6,3 s, 0,0007** | **−3,5 s, 0,003** | **−2,0 s, 0,006** |
| HG-LinUCB-D "chỉ máy vừa nhận kết quả quên" vs LL | **−5,2 s, 0,0002** | **−4,6 s, 0,002** | **−3,8 s, 0,001** |
| HG-LinUCB vs SED | −0,7 s, 0,40 | 0,0 s, 1,00 | +1,3 s, 0,35 |
| HG-LinUCB-D γ = 0,95 vs HG-LinUCB | **+6,3 s, 0,0002** | **+5,9 s, 0,0004** | **+5,0 s, 0,0003** |
| HG-LinUCB-D γ = 0,98 vs HG-LinUCB | **+4,0 s, 0,003** | **+2,8 s, 0,007** | +1,6 s, 0,08 |
| HG-LinUCB-D "chỉ máy vừa nhận kết quả quên" vs HG-LinUCB | +0,3 s, 0,90 | +0,4 s, 0,84 | −0,2 s, 0,74 |

Các quan sát dưới đây là mô tả, không phải giả thuyết đã đăng ký:
1. Khi có drift, HG-LinUCB nhanh hơn LeastLoaded một cách nhất quán (khoảng 1–3 %). Trong TN1, không có
   drift, giả thuyết HG-LinUCB nhanh hơn LeastLoaded không được ủng hộ.
2. Với cách "mọi máy cùng quên", γ càng nhỏ (quên càng nhanh) thì makespan càng lớn. Các bản γ = 0,95 và
   γ = 0,98 chậm hơn HG-LinUCB, có ý nghĩa sau Holm (riêng γ = 0,98 ở E2-onoff thì không).
3. Bản "chỉ máy vừa nhận kết quả quên", tức cách đọc nguyên văn công thức thầy giao: không phát hiện khác
   biệt có ý nghĩa so với HG-LinUCB (đây không phải kiểm định tương đương), và nhanh hơn LeastLoaded có ý
   nghĩa.

---

## 5. Quét λ (E2-lambda, khám phá H2.4)

Median của 20 block, drift vĩnh viễn. "Sau sự kiện" là tỷ lệ quyết định gửi vào worker bị chậm.

| λ | HG-LinUCB: makespan | HG-LinUCB: sau sự kiện | HG-LinUCB-D (0,98): makespan | HG-LinUCB-D (0,98): sau sự kiện |
|---|---:|---:|---:|---:|
| 0 | 188,7 | 0,189 | 193,4 | 0,184 |
| 0,25 | 186,1 | 0,205 | 189,3 | 0,186 |
| 0,5 (mặc định) | 185,2 | 0,215 | 188,8 | 0,192 |
| 1,0 | 186,3 | 0,191 | 188,4 | 0,208 |
| *LeastLoaded / SED* | *192,1 / 186,6* | *0,187 / 0,202* | | |

- Với HG-LinUCB, tỷ lệ sau sự kiện không thay đổi đều theo λ. Với HG-LinUCB-D, tỷ lệ tăng dần theo λ
  (0,184 → 0,208), đúng chiều mà H2.4 mô tả. Nhưng chênh lệch nhỏ và H2.4 không có kiểm định, nên không
  rút kết luận xác nhận.
- λ = 0 (bỏ hẳn phạt tải) có makespan median lớn nhất ở cả hai bản.
- **Độ nhạy:** bỏ block 15 (xem mục 7) thì median makespan thay đổi tối đa 0,62 s
  (`hybrid-linucb-d-g098-l100`), median tỷ lệ sau sự kiện thay đổi tối đa 0,007. Khi bỏ block này, xu
  hướng tăng dần theo λ của HG-LinUCB-D không còn đơn điệu ở hai mức đầu (λ = 0: 0,183; λ = 0,25: 0,182),
  càng cho thấy không nên đọc H2.4 thành một xu hướng.

---

## 6. Diễn giải: vì sao giả thuyết không đứng vững

Các ý dưới đây là diễn giải từ số liệu mô tả, chưa được kiểm chứng bằng thí nghiệm riêng.

1. **LeastLoaded tự điều chỉnh qua số task đang chạy.** Máy bị chậm giữ task lâu hơn, trông "bận" hơn,
   nên LeastLoaded gửi ít đi mà không cần học. Điều này đã được ghi là mối đe dọa trong đăng ký (mục 4
   của `hypotheses.md`).
   - Mức tham chiếu theo tỷ lệ công suất sau drift: 1,1 CPU ÷ 1,87 ≈ 0,59 CPU trên tổng ≈ 3,49 CPU, tức
     **16,9 %**. Đây là mức tham chiếu thô, không phải mức tối ưu đã chứng minh.
   - Ở E2-perm, các scheduler gửi 18,7–20,5 % (median) vào máy bị chậm sau sự kiện, tức đã khá gần mức
     này.
2. **Thời gian thích nghi đều ngắn so với phần còn lại của phiên.** Ở E2-perm, median là 23 (SED) và 26
   (LeastLoaded) quyết định; các bản học máy từ 32 đến 50, trên khoảng 460 quyết định sau sự kiện. Theo
   thước đo này LeastLoaded và SED đạt ngưỡng nhanh nhất, nên discount có ít việc để làm.
3. **Đồng hồ discount toàn cục làm mô hình quên cả các máy không đổi**, có thể gây thêm thăm dò thừa. Điều
   này phù hợp với việc các bản "mọi máy cùng quên" chậm hơn HG-LinUCB, còn bản "chỉ máy vừa nhận kết quả
   quên" thì không.
4. **Ưu thế của HG-LinUCB so với LeastLoaded khi có drift** có thể đến từ việc nó phân biệt máy mạnh/yếu
   và nhận biết máy 1,1 CPU chậm đi qua phần thưởng, chứ không phải từ discount.

---

## 7. Kiểm tra tính hợp lệ, sự cố và hạn chế

**Quy trình:**
- **Đăng ký trước:** `hypotheses.md` đóng băng ở `1b8339d` trước TN1, không sửa. Quyết định chạy E2-lambda
  (`E2_LAMBDA=RUN`) được ghi một lần trước TN2 và bị khoá bằng file.
- **Bịt mắt:** không mở kết quả so sánh scheduler cho tới khi chạy xong cả bốn kịch bản.
- **Kiểm toán trước khi suy luận:** mọi dòng giả thuyết có n = 20, không có giá trị thiếu, số so sánh
  trong họ đúng (4 và 10).
- **Tính lại độc lập:** 70 dòng so sánh khớp với analyzer (mục 9).
- **Toàn vẹn dữ liệu:** có manifest SHA-256 cho từng kịch bản; dữ liệu TN1 không đổi suốt TN2.
- **Thay đổi công cụ vận hành** (ngoài repo, không đổi runner hay analyzer): sổ chạy lại lưu ra file, lưu
  trữ block dở, khoá quyết định λ, dừng an toàn, giám sát chỉ cảnh báo. Xem
  `D:\bench-work\CHANGELOG-phase2.md`; đã test 21 + 5 ca và review ba vòng.

**Sự cố và sai lệch:**
- **Tạm dừng theo yêu cầu:** sau block 14 của E2-lambda (19/10 → 21/10). Trước khi chạy tiếp đã so khớp
  trạng thái máy: cùng máy, Docker, image, volume, code, Python; 7.253 file dữ liệu không đổi. Block dở
  được lưu riêng rồi chạy lại nguyên block.
- **Cell đầu tiên sau khi chạy tiếp** (LeastLoaded, block 15):
  - Hỏng một lần vì worker chưa kịp đăng ký; runner chạy lại theo quy tắc đăng ký.
  - Makespan 204,6 s, cao hơn median của LeastLoaded (192,1 s) 6,45 %. Có thể do máy vừa nghỉ lâu, nhưng
    đây là phỏng đoán, chưa chứng minh.
  - Cell không bị loại, vì không vi phạm tiêu chí loại nào; phân tích độ nhạy ở mục 5.
- **Thời lượng pha bật/tắt** khoảng 14,1 s thay vì 10 s danh nghĩa, do độ trễ `docker exec`. Đăng ký đã
  lường trước; phân tích dùng mốc ghi trong log.
- **Sự kiện thừa ở E2-onoff:** ở 19/140 cell, sau một `drift_off` thường có thêm `drift_off` dọn dẹp ở
  dispatch 590. Trong 6 cell đó, task log còn có thêm một `drift_on` ở dispatch 590 mà file events CSV
  không ghi. Các sự kiện này nằm sau task cuối cùng, không chứa quyết định nào, và việc dọn dẹp thành
  công.
- **Bộ lấy mẫu `docker stats`** báo lỗi khi tắt ở 102 cell (19 / 26 / 24 / 33 cell ở perm / trans /
  onoff / lambda). Lỗi này chỉ làm mất mẫu thống kê CPU cuối cùng của cell, không ảnh hưởng task hay kết
  quả.
- **RAM Docker** ghi nhận ở TN2 thấp hơn TN1 0,002 %, do khởi động lại Docker Desktop giữa hai đợt; thiết
  lập vẫn là 31 GiB, 72 CPU.
- **Model của giám sát viên Codex** đổi từ `gpt-5.5` (TN1) sang `gpt-5.6-sol` (TN2), vì bản cũ không còn
  được hỗ trợ.

**Hạn chế (đã ghi trước khi có kết quả):**
- `time_to_adapt` là chỉ số thứ bậc bị chặn, không phải phân tích sống sót.
- H2.3 nằm trong họ 10 so sánh với 20 block, nên sức mạnh thống kê thấp.
- Thứ tự kịch bản cố định; thứ tự arm trong block dùng cùng seed ở mọi kịch bản.
- Chỉ một máy và một workload (CPython).
- Phiên 2 build khác với giao thức của bài báo và của TN1.

---

## 8. Câu hỏi cho thầy và đề xuất bước tiếp

1. Thầy có muốn đổi **bản HG-LinUCB-D chính** sang cách "chỉ máy vừa nhận kết quả quên" cho các thí
   nghiệm sau không? Trong TN2, so với HG-LinUCB, bản này không cho thấy khác biệt có ý nghĩa (không có
   kiểm định tương đương hay non-inferiority). Ngược lại, các bản "mọi máy cùng quên" với γ = 0,95 và
   0,98 chậm hơn HG-LinUCB, có ý nghĩa.
2. Muốn thật sự kiểm tra khả năng thích nghi, cần một kịch bản mà LeastLoaded **không** tự điều chỉnh được
   qua số task đang chạy. Ví dụ: máy vẫn xử lý nhanh task nhỏ nhưng rất chậm với task lớn (TN4, thiếu RAM),
   hoặc drift làm sai phần thưởng chứ không chỉ làm chậm. Nếu thầy đồng ý, phần này sẽ làm thành **kịch bản
   bổ sung có đăng ký riêng**, không sửa TN2.
3. Thầy duyệt để bắt đầu TN3 (workload C++ đuôi nặng) theo thứ tự đã giao.
4. Log thô khoảng 710 MB (TN2) + 273 MB (TN1): thầy muốn lưu bằng Git LFS, file nén, hay gửi riêng?

---

## 9. Phụ lục

### 9.1 Bảng mô tả theo kịch bản (median của 20 block)

Cột: b1, b2 = makespan build 1, build 2 (s); trước = tỷ lệ gửi vào worker bị chậm trong [1, onset);
trước¹⁰¹ = trong [101, onset); sau / sớm / muộn = sau sự kiện / 50 quyết định đầu sau sự kiện / phần còn
lại; dư = phần gửi vào worker bị chậm vượt mức tỷ lệ công suất trong các slot còn trống (âm = ít hơn);
TTA = time-to-adapt (tỷ lệ cell bị kiểm duyệt); cuối = tỷ lệ block mà task hoàn thành cuối cùng chạy trên
worker bị chậm. D = HG-LinUCB-D "mọi máy cùng quên"; D-arm = "chỉ máy vừa nhận kết quả quên".

**E2-perm**

| Scheduler | b1 | b2 | trước | trước¹⁰¹ | sau | sớm | muộn | dư | TTA (kd) | cuối |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| LeastLoaded | 92,8 | 97,5 | 0,238 | 0,217 | 0,188 | 0,20 | 0,195 | −0,278 | 26 (0 %) | 0,15 |
| SED | 89,9 | 94,3 | 0,245 | 0,207 | 0,203 | 0,17 | 0,203 | −0,247 | 23 (0 %) | 0,35 |
| HG-LinUCB | 91,5 | 93,1 | 0,254 | 0,293 | 0,205 | 0,23 | 0,200 | −0,230 | 50 (0 %) | 0,10 |
| D γ = 0,95 | 94,2 | 97,7 | 0,287 | 0,393 | 0,196 | 0,22 | 0,191 | −0,179 | 50 (5 %) | 0,30 |
| D γ = 0,98 | 92,8 | 97,3 | 0,235 | 0,258 | 0,190 | 0,22 | 0,188 | −0,224 | 45,5 (0 %) | 0,10 |
| D γ = 0,99 | 91,3 | 94,5 | 0,273 | 0,266 | 0,187 | 0,22 | 0,189 | −0,236 | 32 (0 %) | 0,10 |
| D-arm γ = 0,98 | 90,0 | 93,4 | 0,273 | 0,344 | 0,194 | 0,23 | 0,192 | −0,244 | 41 (0 %) | 0,20 |

**E2-trans** (thêm cột "hồi phục" = tỷ lệ sau khi máy hồi phục; 140/140 cell có đúng một cặp bật/tắt)

| Scheduler | b1 | b2 | trước | trước¹⁰¹ | sau | sớm | muộn | dư | hồi phục | TTA (kd) | cuối |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| LeastLoaded | 91,5 | 86,7 | 0,220 | 0,222 | 0,246 | 0,17 | 0,253 | −0,228 | 0,272 | 40,5 (0 %) | 0,05 |
| SED | 89,7 | 85,1 | 0,244 | 0,193 | 0,250 | 0,13 | 0,258 | −0,199 | 0,283 | 20 (0 %) | 0,40 |
| HG-LinUCB | 89,9 | 84,1 | 0,256 | 0,273 | 0,249 | 0,20 | 0,246 | −0,192 | 0,270 | 27 (0 %) | 0,20 |
| D γ = 0,95 | 92,8 | 87,5 | 0,210 | 0,164 | 0,245 | 0,18 | 0,257 | −0,158 | 0,268 | 26 (0 %) | 0,25 |
| D γ = 0,98 | 90,6 | 87,1 | 0,229 | 0,158 | 0,250 | 0,17 | 0,266 | −0,170 | 0,270 | 27 (0 %) | 0,20 |
| D γ = 0,99 | 91,1 | 84,9 | 0,258 | 0,333 | 0,240 | 0,21 | 0,252 | −0,184 | 0,258 | 42,5 (0 %) | 0,30 |
| D-arm γ = 0,98 | 90,4 | 84,2 | 0,225 | 0,237 | 0,256 | 0,19 | 0,266 | −0,169 | 0,296 | 26,5 (0 %) | 0,20 |

**E2-onoff** (thêm tỷ lệ gộp trong các pha bật và pha tắt)

| Scheduler | b1 | b2 | trước | trước¹⁰¹ | sau | sớm | muộn | dư | pha bật | pha tắt | TTA (kd) | cuối |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| LeastLoaded | 91,2 | 92,8 | 0,253 | 0,271 | 0,234 | 0,21 | 0,236 | −0,241 | 0,203 | 0,278 | 43,5 (0 %) | 0,20 |
| SED | 89,4 | 90,5 | 0,301 | 0,246 | 0,227 | 0,16 | 0,241 | −0,224 | 0,182 | 0,278 | 20 (0 %) | 0,45 |
| HG-LinUCB | 90,7 | 90,1 | 0,243 | 0,207 | 0,254 | 0,22 | 0,264 | −0,180 | 0,203 | 0,302 | 59 (5 %) | 0,25 |
| D γ = 0,95 | 91,2 | 94,9 | 0,264 | 0,283 | 0,235 | 0,22 | 0,238 | −0,164 | 0,210 | 0,270 | 33 (0 %) | 0,05 |
| D γ = 0,98 | 91,3 | 91,2 | 0,292 | 0,379 | 0,254 | 0,22 | 0,255 | −0,176 | 0,213 | 0,288 | 38 (0 %) | 0,15 |
| D γ = 0,99 | 90,2 | 91,3 | 0,233 | 0,259 | 0,252 | 0,17 | 0,264 | −0,180 | 0,207 | 0,305 | 37 (0 %) | 0,15 |
| D-arm γ = 0,98 | 89,9 | 90,8 | 0,283 | 0,263 | 0,240 | 0,21 | 0,238 | −0,190 | 0,195 | 0,288 | 29,5 (0 %) | 0,30 |

Cấu trúc pha của E2-onoff, tính trên 140 cell, không kể sự kiện dọn dẹp `cell_end`:
- **Median `dispatch_count` của các sự kiện bật/tắt, theo thứ tự:** 129 (bật), 176, 237, 289, 327,
  358, 392, 444,5, 499, 550. Sự kiện thứ 11 (bật, ở dispatch 590) chỉ có ở 98/140 cell và nằm sau task
  cuối cùng.
- **Pha 4** (bắt đầu từ sự kiện ở median 289) bắc qua ranh giới hai build ở 140/140 cell. Điều này là
  bình thường, vì đồng hồ pha dừng giữa hai build.
- **Mỗi pha** chứa khoảng 30–60 quyết định. Tính gộp mỗi phiên có median 240 quyết định trong pha bật và
  221 trong pha tắt.

**E2-lambda** (drift vĩnh viễn)

| Scheduler | b1 | b2 | trước | sau | sớm | muộn | dư | TTA (kd) | cuối |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| LeastLoaded | 94,8 | 98,2 | 0,217 | 0,187 | 0,16 | 0,191 | −0,276 | 23,5 (0 %) | 0,05 |
| SED | 91,4 | 94,0 | 0,266 | 0,202 | 0,15 | 0,200 | −0,250 | 22 (0 %) | 0,20 |
| HG-LinUCB λ = 0 | 92,7 | 96,3 | 0,288 | 0,189 | 0,25 | 0,179 | −0,209 | 56,5 (0 %) | 0,20 |
| HG-LinUCB λ = 0,25 | 92,1 | 93,5 | 0,247 | 0,205 | 0,16 | 0,202 | −0,218 | 26 (0 %) | 0,20 |
| HG-LinUCB λ = 0,5 | 90,9 | 93,8 | 0,275 | 0,215 | 0,23 | 0,220 | −0,223 | 52,5 (0 %) | 0,15 |
| HG-LinUCB λ = 1,0 | 91,3 | 93,7 | 0,266 | 0,191 | 0,19 | 0,193 | −0,243 | 40 (0 %) | 0,25 |
| D γ = 0,98, λ = 0 | 94,0 | 99,8 | 0,276 | 0,183 | 0,24 | 0,180 | −0,176 | 42,5 (0 %) | 0,35 |
| D γ = 0,98, λ = 0,25 | 92,4 | 96,3 | 0,228 | 0,185 | 0,20 | 0,184 | −0,212 | 25 (0 %) | 0,05 |
| D γ = 0,98, λ = 0,5 | 92,1 | 96,4 | 0,249 | 0,192 | 0,21 | 0,188 | −0,228 | 47 (0 %) | 0,20 |
| D γ = 0,98, λ = 1,0 | 92,2 | 95,6 | 0,301 | 0,208 | 0,20 | 0,206 | −0,222 | 28,5 (0 %) | 0,15 |

### 9.2 Hình khám phá (không đăng ký trước)

`docs/thesis/figures/tn2-rolling-share.svg`: tỷ lệ trượt 20 quyết định gửi vào worker bị chậm theo
`dispatch_seq`, median của 20 block, mỗi scheduler một đường, có vạch onset và vạch hồi phục.

### 9.3 Kiểm tra sự kiện (khám phá; định nghĩa chốt trước khi mở kết quả)

`D:\bench-work\review\tn2_manipulation.py`, chỉ trên E2-perm:
- Tỷ số được đo: `compile_time_ms` trên 100 kB mã đã tiền xử lý.
- Cách tính: lấy median trước và sau onset trong từng cell, rồi so worker 1,1 CPU với worker 1,0 CPU làm
  đối chứng. Dùng 140/140 cell, không bỏ cell nào, không bỏ task nào.

Kết quả:
- Worker 1,1 CPU chậm đi **1,87 lần** (IQR 1,37–2,25).
- Worker 1,0 CPU **0,99 lần** (IQR 0,79–1,26).
- Chỉ số tương đối giữa hai worker **1,80** (IQR 1,29–2,54).

### 9.4 Vị trí dữ liệu và script

- Kết quả analyzer: `D:\bench-work\out\official\E2-*\analysis*\` (`pairs.csv`, `cells.csv`, `summary.md`).
- Tính lại độc lập: `D:\bench-work\review\tn2_independent\` (`recompute.py`, `results.csv`).
- Bảng mô tả, độ nhạy, hình: `D:\bench-work\review\tn2_extract.py`, `tn2_describe.py`, `tn2_describe2.py`,
  `tn2_figure.py`.
- Manifest SHA-256: `D:\bench-work\out\official\manifest_E1.sha256`, `manifest_E2-*.sha256`.
