# Thực nghiệm bổ sung HG-LinUCB: báo cáo tiến độ

**Thí nghiệm 1 đã xong · Thí nghiệm 2 đã sẵn sàng** · 11/10/2026

*(Mỗi mục ngăn bởi `---` là một slide. Tệp này để trình chiếu; kịch bản nói nằm ở `bao-cao-tien-do-261011.md`.)*

---

## 1. Thầy yêu cầu gì

- HG-LinUCB **hòa** LeastLoaded vì môi trường thử quá dễ: luôn có máy rảnh, task đều, không có biến động.
- Dựng các kịch bản thực tế hơn, nơi LeastLoaded có điểm yếu, rồi xem HG-LinUCB có thắng không.
- Thắng hay thua đều báo cáo đầy đủ.

**Quy định chung:** ≥ 4 scheduler (LeastLoaded, SED, HG-LinUCB, HG-LinUCB-D) · 20 block ngẫu nhiên · viết `hypotheses.md` trước, không sửa sau · sự kiện bơm vào sau dispatch 100 · ghi giờ sự kiện vào log.

**Thứ tự:** thí nghiệm 1 → 2 → gửi thầy xem → 3 → 4 → 5 → 6.

---

## 2. LeastLoaded yếu ở đâu

> Như người **luôn chọn quầy thu ngân ít khách nhất**.

| Điểm mù | Hậu quả |
|---|---|
| Chỉ **đếm** khách, không biết quầy nào làm nhanh hay chậm | Quầy yếu và quầy mạnh được xếp như nhau |
| Không thấy quầy nào **đột nhiên chậm đi** | Vẫn gửi việc cho quầy đó |

→ **Thí nghiệm 1:** tải vừa phải (gần hết máy rảnh).
→ **Thí nghiệm 2:** một máy bị chậm đi mà hệ thống không biết.

---

## 3. Đã cài thêm hai thành phần

**SED** (baseline): chọn máy có `(số task đang chạy + 1) / công suất` nhỏ nhất, công suất đọc từ cgroup.

| Máy | Công suất | Task đang chạy | Điểm |
|---|---|---|---|
| A (mạnh gấp đôi) | 2 | 1 | 2/2 = **1** (được chọn) |
| B | 1 | 1 | 2/1 = 2 |

**HG-LinUCB-D** (bản "quên dần"):

```
A ← γA + xxᵀ + (1−γ)I        b ← γb + r·x
```

Dữ liệu cũ nhẹ dần, nên khi một máy đổi tốc độ, bandit đổi ý nhanh hơn.
Cài **hai cách đọc** công thức: **mọi máy cùng quên** (bản chính) và **chỉ máy vừa nhận kết quả** (đọc đúng chữ).

Kiểm tra: so kết quả cũ từng bit · 31 lỗi cố ý đều bị test bắt · race detector sạch.

---

## 4. Đo cho đúng

- **Log chi tiết:** số thứ tự dispatch, giờ dispatch, ảnh chụp trạng thái cả cụm lúc quyết định, giờ của từng sự kiện bơm vào.
- **20 block ngẫu nhiên:** mỗi block chạy đủ 4 scheduler; so sánh **trong cùng block** nên nhiễu máy ảnh hưởng đều.
- **Thống kê:** median, hiệu số ghép cặp, Wilcoxon, hiệu chỉnh Holm, rank-biserial ghép cặp.
- **Đăng ký trước:** `hypotheses.md` đóng băng ở commit `1b8339d`, trước lượt chạy đầu tiên.
- **Giám sát độc lập:** Codex kiểm tra dữ liệu 90 phút một lần (12 lần; không có vấn đề ảnh hưởng đến dữ liệu).

---

## 5. Thí nghiệm 1: thiết kế và số lượng

| | |
|---|---|
| Kịch bản | **5**: `-j6`, `-j7`, `-j8`, `-j9`, 2 client `-j8` |
| Scheduler | **4**: LeastLoaded, SED, HG-LinUCB, HG-LinUCB-D (γ=0,98) |
| Block | **20** mỗi kịch bản |
| **Số lượt chạy** | **400 cell = 480 build = 141.600 task** |
| Thời gian | **14,5 giờ** (một cell ≈ 110–196 giây) |
| Cell hỏng / chạy lại / block bị loại | **0 / 0 / 0** |

Cluster: 5 worker (0,5 · 0,6 · 0,8 · 1,0 · 1,1 CPU), 10 chỗ chạy song song · CPython v3.14.0 (295 file).

---

## 6. Thí nghiệm 1: makespan (giây, median của 20 block)

| Kịch bản | LeastLoaded | SED | HG-LinUCB | HG-LinUCB-D |
|---|---:|---:|---:|---:|
| `-j6` | 91,2 | 96,3 | **89,6** | 92,9 |
| `-j7` | 89,8 | **86,4** | 87,5 | 90,7 |
| `-j8` | 87,4 | **85,7** | 86,4 | 87,5 |
| `-j9` | **84,5** | 84,9 | 85,0 | 84,8 |
| 2 client | 165,2 | **164,8** | 165,3 | 165,6 |

→ **Tải càng cao, các scheduler càng giống nhau:** chênh lệch giữa scheduler nhanh nhất và chậm nhất: 6,7 s (`-j6`) → 0,5 s (`-j9`) → 0,8 s (2 client, build dài 165 s).

---

## 7. Kết quả theo giả thuyết đã đăng ký

| Giả thuyết | Kết quả | Chi tiết (p đã hiệu chỉnh Holm) |
|---|---|---|
| **H1.1** HG-LinUCB nhanh hơn LeastLoaded ở `-j8`, `-j9`, 2 client | ❌ **Không ủng hộ** | `-j8`: −0,84 s, p = 0,098 · `-j9`: +0,62 s, p = 0,87 · 2 client: +0,22 s, p = 1,0 |
| **H1.2** HG-LinUCB chọn đúng máy rảnh mạnh nhất nhiều hơn LeastLoaded, ở mọi kịch bản | ◐ **Một phần** (3/5) | Đạt ở `-j6`, `-j8`, `-j9` (p < 0,004). Không đạt ở `-j7` (p = 0,080) và 2 client (p = 0,86) |
| **H1.3** SED bỏ qua máy rảnh nhiều hơn LeastLoaded | ✅ **Ủng hộ** (5/5) | p < 0,0001 ở mọi kịch bản |

---

## 8. Chọn đúng máy rảnh mạnh nhất (median)

| Kịch bản | LeastLoaded | SED | HG-LinUCB | HG-LinUCB-D |
|---|---:|---:|---:|---:|
| `-j6` | 0,84 | 0,52 | **0,90** | 0,82 |
| `-j7` | 0,88 | 0,86 | **0,90** | 0,86 |
| `-j8` | 0,91 | 0,81 | **0,94** | 0,91 |
| `-j9` | 0,90 | 0,87 | 0,94 | **0,94** |
| 2 client | 0,96 | 0,94 | **0,97** | 0,93 |

**SED bỏ qua máy rảnh** (tỉ lệ quyết định): `-j6` 48% · `-j7` 9% · `-j8` 8% · `-j9` 5% · 2 client 1% (LeastLoaded: 0% ở mọi kịch bản).

*"Máy rảnh" = không có task nào đang chạy (định nghĩa của Bảng 5 trong bài).*

---

## 9. Vì sao tải cao thì các scheduler giống nhau?

Tỉ lệ quyết định còn **ít nhất một máy hoàn toàn rảnh** (LeastLoaded):

| `-j6` | `-j7` | `-j8` | `-j9` | 2 client |
|---:|---:|---:|---:|---:|
| 79% | 61% | 43% | 25% | **15%** |

- Khi gần hết máy rảnh, LeastLoaded vốn đã chọn đúng việc cần làm là **tránh máy bận**.
- Còn rất ít chỗ để scheduler thông minh hơn tạo khác biệt.
- *Đây là giả thuyết giải thích, chưa kiểm chứng riêng.*

→ Đây là kết quả **hòa hoặc thua** của HG-LinUCB, báo cáo đúng như thầy dặn.

---

## 10. Thí nghiệm 2 (drift ẩn): đã sẵn sàng, chưa chạy

**Cách làm chậm:** `stress-ng --cpu 1` chạy **bên trong** container của máy 1,1 CPU, ở khoảng dispatch 120 (không đổi hạn mức CPU).

| Biến thể | Mô tả |
|---|---|
| Vĩnh viễn | Chậm đến hết phiên |
| 30 giây | Chậm 30 s rồi hồi phục |
| Bật/tắt | Dao động 10 s bật, 10 s tắt |

**Scheduler (7):** LeastLoaded · SED · HG-LinUCB · HG-LinUCB-D γ = 0,95 / 0,98 / 0,99 · bản per-arm γ = 0,98.
**Quét λ** ∈ {0, 0,25, 0,5, 1,0} (10 scheduler, chỉ với drift vĩnh viễn).
**Chỉ số:** makespan · tỉ lệ gửi vào máy bị chậm trước/sau · số quyết định để thích nghi · máy chạy task cuối.

**Phát hiện từ thử nghiệm:** một build chỉ còn ~45 quyết định sau khi hồi phục → dùng **phiên hai build liên tiếp** (~340 quyết định).

---

## 11. Cơ sở lý thuyết: Russac, Vernade, Cappé (NeurIPS 2019)

**Weighted Linear Bandits for Non-Stationary Environments (D-LinUCB)**

- Bandit tuyến tính mà tham số thật **thay đổi theo thời gian**.
- Dùng bình phương tối thiểu có **trọng số mũ γ^s**: quên dần quá khứ, **không cần lưu lại dữ liệu cũ**.
- Hối tiếc động O(d^{2/3} · B_T^{1/3} · T^{2/3}), khớp cận dưới (cần biết ngân sách biến động B_T).

| Trong bài | Trong hệ thống |
|---|---|
| Tham số θ đổi theo thời gian | Tốc độ thật của một worker đổi |
| V ← γV + AAᵀ + (1−γ)λI | `hybrid-linucb-d`, γ ∈ {0,95; 0,98; 0,99} |

**Khác bài** (nên không nêu cận hối tiếc): thống kê **riêng từng worker**; thăm dò dạng đơn giản; phản hồi đến **muộn**; có thêm hệ số phạt tải λ.

*Ý nghĩa của γ:* nửa trọng số mất sau **14** quyết định (γ=0,95), **34** (γ=0,98), **69** (γ=0,99).

---

## 12. Thời gian chạy

> Tổng = số kịch bản × số scheduler × số block × thời gian một cell

| | Số cell | Thời gian một cell | Tổng |
|---|---:|---:|---:|
| **TN1** (đã chạy) | 5 × 4 × 20 = 400 | 110–196 giây | **14,5 giờ** (đo thật) |
| TN2, 3 biến thể | 3 × 7 × 20 = 420 | ≈ 218 giây (2 build) | ≈ 25,4 giờ |
| TN2, quét λ | 10 × 20 = 200 | ≈ 218 giây | ≈ 12,1 giờ |
| **Tổng** | **1.020** | | **≈ 52 giờ** |

Một cell ≈ 93 s build thật + 25 s dựng cluster và dọn. Cell phải chạy **tuần tự**, không song song, vì chạy cùng lúc sẽ cản nhau và làm sai số đo.

---

## 13. Cần thầy cho ý kiến

1. **Hai bản "quên dần":** mọi máy cùng quên (bản chính hiện tại) hay chỉ máy vừa nhận kết quả (đọc đúng chữ công thức)?
2. **Thí nghiệm 2 dùng phiên hai build liên tiếp** (khác giao thức cũ: reset mỗi build): có hợp lý không?
3. **Log thô** ~400–500 MB (≈ 50 MB khi nén): đưa lên repo dạng nén, dùng Git LFS, hay gửi riêng?

---

## 14. Bước tiếp theo

- [x] Cài SED và HG-LinUCB-D
- [x] Công cụ đo, đăng ký giả thuyết, giám sát độc lập
- [x] **Thí nghiệm 1** (400 cell, không lỗi)
- [ ] **Thí nghiệm 2** (≈ 25 giờ) → báo cáo thầy
- [ ] Quét λ (≈ 12 giờ), nếu thầy đồng ý
- [ ] Thí nghiệm 3–6 (sau khi thầy xem 1 và 2)
