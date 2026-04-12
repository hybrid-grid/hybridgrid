# Vấn đề lập lịch phân phối công việc trong hệ thống Hybrid-Grid Build

## 1. Tổng quan hệ thống hiện tại

Hybrid-Grid Build là hệ thống build phân tán đa nền tảng được phát triển trong khuôn khổ đồ án, cho phép chia sẻ công việc biên dịch (compilation) và đóng gói (build) từ máy của lập trình viên sang một cụm các máy worker trong mạng LAN hoặc WAN. Mục tiêu là rút ngắn thời gian build các dự án phần mềm lớn bằng cách tận dụng tài nguyên rảnh của nhiều máy tính khác nhau.

Kiến trúc hệ thống gồm ba thành phần chính:

- **hgbuild (CLI)**: công cụ dòng lệnh đóng vai trò thay thế trực tiếp cho `gcc`, `g++`, `make`, `ninja`, `flutter`. Công cụ này chạy trên máy lập trình viên, tiếp nhận lệnh build, tiền xử lý (preprocess) và gửi yêu cầu tới coordinator thông qua gRPC.
- **hg-coord (coordinator)**: máy chủ trung tâm có nhiệm vụ tiếp nhận yêu cầu build, duy trì danh sách các worker qua cơ chế đăng ký có thời gian sống (TTL) kết hợp tự động khám phá qua mDNS, ra quyết định phân phối task tới worker (phần scheduler), và quản lý bộ nhớ đệm kết quả theo nội dung (content-addressable cache).
- **hg-worker**: các máy thực thi build. Mỗi worker tự khai báo năng lực (capabilities) với coordinator khi đăng ký: kiến trúc CPU (x86_64, arm64), số nhân xử lý, dung lượng bộ nhớ, các trình biên dịch sẵn có (gcc, clang, MSVC), khả năng cross-compile qua Docker, và loại build system được hỗ trợ (C/C++, Flutter Android, ...).

Tại thời điểm hiện tại, hệ thống đã hỗ trợ biên dịch phân tán C/C++ (các phiên bản v0.2 đến v0.3) và build Flutter Android (v0.4) trên cụm các máy có cấu hình không đồng nhất (heterogeneous), bao gồm máy desktop kiến trúc x86_64, laptop kiến trúc arm64 (Apple M1), và máy Raspberry Pi kiến trúc arm64. Hệ thống đã có các tính năng bổ trợ cần thiết cho một hệ phân tán: chịu lỗi qua circuit breaker và cơ chế local fallback, quan sát qua Prometheus metrics và OpenTelemetry tracing, bảo mật qua TLS/mTLS, và cache nội dung dùng hàm băm xxhash.

## 2. Bài toán lập lịch trong hệ thống

Thành phần scheduler trong coordinator có nhiệm vụ trả lời một câu hỏi xuất hiện mỗi khi có task build tới: **nên giao task này cho worker nào?** Đây là bài toán lập lịch trực tuyến (online scheduling) trên tập máy không đồng nhất (unrelated heterogeneous machines), với hai ràng buộc quan trọng: quyết định phải được đưa ra ngay khi task đến (không biết trước các task tương lai) và độ trễ quyết định phải rất nhỏ so với thời gian thực thi task.

Phiên bản hiện tại sử dụng thuật toán **Power of Two Choices (P2C)** kết hợp với một hàm tính điểm (scoring function) có trọng số tĩnh. Cơ chế hoạt động như sau:

1. Lọc ra các worker có đủ năng lực thực thi task (capability matching).
2. Lấy ngẫu nhiên hai worker trong tập đã lọc.
3. Tính điểm mỗi worker theo tổng có trọng số của các đại lượng: mức độ khớp kiến trúc (native match hay cross-compile), số nhân CPU, dung lượng RAM, số task đang chạy, độ trễ mạng, và một điểm thưởng nếu worker nằm cùng LAN.
4. Gán task cho worker có điểm cao hơn.

Cách tiếp cận này cho kết quả chấp nhận được trên cụm worker đồng nhất nhưng bộc lộ năm hạn chế khi cluster có tính heterogeneous cao, vốn là trường hợp sử dụng chính mà hệ thống hướng tới:

**Hạn chế 1 — Trọng số được ấn định cố định trong mã nguồn.** Các hệ số cho CPU, RAM, số task đang chạy, độ trễ đều là hằng số được mã hoá cứng. Không có cơ sở chứng minh tỉ lệ này tối ưu cho mọi workload. Một workload nặng về compile template C++ và một workload nặng về link LTO có đặc trưng (profile) khác nhau đáng kể, nhưng scheduler hiện tại dùng chung một bộ trọng số cho cả hai.

**Hạn chế 2 — Không phân biệt loại task.** Scheduler coi task biên dịch một file `helloworld.c` kích thước khoảng 200 mili-giây và một file chứa template Boost kích thước khoảng 45 giây là như nhau về mặt quyết định phân phối. Hậu quả trực tiếp là task nặng có thể bị gán cho worker yếu (Raspberry Pi), trong khi task nhẹ lại bị gán cho worker mạnh (desktop), trái ngược với phân bổ tối ưu.

**Hạn chế 3 — Số task đang chạy là đại lượng không phản ánh tải thực tế.** Giá trị `active_tasks = 3` có thể biểu diễn ba task nhỏ sắp hoàn thành, hoặc ba task lớn còn thực thi rất lâu. Scheduler không phân biệt được hai trạng thái "bận sắp rảnh" và "bận kéo dài" mặc dù chúng dẫn đến quyết định phân phối khác nhau.

**Hạn chế 4 — Không học từ dữ liệu lịch sử.** Scheduler hiện tại ra quyết định hoàn toàn dựa trên trạng thái tức thời. Nó không ghi nhớ được thông tin có giá trị như "worker-B đã từng compile file `chrono.hpp` trong 12 giây trong khi worker-A chỉ mất 3 giây cho cùng file đó". Các sai lầm phân phối có xu hướng lặp lại nhiều lần.

**Hạn chế 5 — Phạt cross-compile là nhị phân.** Hàm tính điểm chỉ có hai mức: khớp kiến trúc native (cộng điểm cao) hoặc phải cross-compile qua Docker (cộng điểm thấp). Trong thực tế, mức độ chậm khi cross-compile tuỳ thuộc vào loại task: một file C nhỏ chỉ chậm khoảng 1.5 lần, trong khi một file C++ nặng template có thể chậm 3 đến 5 lần. Một hằng số duy nhất không biểu diễn được tính liên tục này.

Năm hạn chế trên cộng hưởng trong môi trường cluster heterogeneous thực tế (kết hợp máy workstation mạnh, laptop arm, máy SBC yếu, và node Docker cross-compile), dẫn tới hai hiện tượng đã quan sát được trong các lần đo hiệu năng sơ bộ: mất cân bằng tải rõ rệt giữa các worker, và sự xuất hiện của straggler — tức là một worker đột nhiên trở thành nút cổ chai và kéo dài đáng kể tổng thời gian hoàn thành (makespan) của cả batch build.

## 3. Hướng tiếp cận đề xuất

Tác giả đề xuất thay thế hàm tính điểm heuristic tĩnh hiện tại của P2C bằng một **cost model có khả năng dự đoán** (predictive cost model). Với mỗi cặp (worker, task) đang được xem xét, cost model ước lượng tổng thời gian hoàn thành dự kiến nếu task đó được giao cho worker đó. Scheduler vẫn giữ cơ chế lấy mẫu P2C (chọn hai worker ngẫu nhiên rồi chọn cái tốt hơn) nhằm tránh hiệu ứng herd behavior và giữ chi phí quyết định ở mức thấp, nhưng hàm tính điểm bên trong được thay thế hoàn toàn bằng cost model.

Cost model dự kiến bao gồm bốn thành phần:

1. **Thời gian thực thi dự đoán** — học từ lịch sử thông qua exponential moving average (EMA) cho từng cặp (worker, task class). Task class được xác định qua phân loại nhẹ dựa trên loại file nguồn, kích thước, và một số đặc trưng tĩnh có thể lấy được nhanh (số lượng include, có sử dụng template hay không).
2. **Thời gian chờ trong hàng đợi** — tính bằng tổng thời gian thực thi dự đoán của tất cả task đang xếp hàng trên worker đó, thay cho đại lượng "số task đang chạy" đơn thuần.
3. **Chi phí truyền mạng** — tính bằng kích thước dữ liệu đầu vào chia cho băng thông khả dụng của worker, cộng hai lần thời gian round-trip (RTT) đo liên tục qua các lời gọi gRPC.
4. **Hệ số phạt** — bao gồm phạt cross-compile được hiệu chỉnh từ dữ liệu lịch sử (thay cho phạt nhị phân hiện tại), phạt theo tỉ lệ thất bại gần đây của worker, và chi phí vô hạn nếu worker không đủ năng lực thực thi.

Bản chất khác biệt của hướng tiếp cận này so với P2C hiện tại nằm ở câu hỏi mà scheduler cố gắng trả lời: P2C hiện tại trả lời câu hỏi "worker nào đang trông khoẻ hơn?", trong khi cost model đề xuất trả lời câu hỏi "worker nào sẽ kết thúc task này sớm nhất?". Chuyển từ heuristic sang predictive là thay đổi quan trọng về mặt khoa học.

Phương pháp đánh giá dự kiến bao gồm việc so sánh trực tiếp scheduler mới với P2C hiện tại và các baseline khác (Least-Loaded, Round-Robin) trên cùng một cluster heterogeneous, cùng tập workload thực tế (ví dụ Redis, CPython, một subset của LLVM). Các metric đánh giá: makespan (thời gian hoàn thành tổng), tail latency ở phân vị 95 và 99, độ mất cân bằng tải theo thời gian, và khả năng phục hồi khi chủ động đưa vào một straggler nhân tạo.

Tác giả đã rà soát sơ bộ một số công trình liên quan để định vị hướng tiếp cận: lý thuyết P2C của Mitzenmacher (2001) là nền tảng của scheduler hiện tại; Sparrow (Ousterhout và cộng sự, 2013 — SOSP) áp dụng P2C vào môi trường data center trực tuyến; bài toán scheduling trên unrelated machines có các thuật toán xấp xỉ cổ điển của Lenstra, Shmoys và Tardos (1990); Firmament (Gog và cộng sự, 2016 — OSDI) sử dụng mô hình min-cost max-flow. Tuy nhiên, trong phạm vi rà soát ban đầu, tác giả chưa tìm thấy công trình nào kết hợp đồng thời bốn đặc tính: lập lịch trực tuyến, cluster heterogeneous, học từ lịch sử thực thi, và hỗ trợ nhiều loại build system. Đây chính là khoảng trống mà đồ án kỳ vọng khai thác như một đóng góp khoa học.

## 4. Các nội dung xin ý kiến hướng dẫn

Tác giả xin được thầy/cô cho ý kiến định hướng về các nội dung sau trước khi chuyển sang giai đoạn cài đặt:

1. Hướng tiếp cận "cost model học từ dữ liệu lịch sử thay thế P2C heuristic" có phù hợp với kỳ vọng về mức độ đóng góp mới của đồ án không? Nếu chưa phù hợp, hướng nghiên cứu nào nên được ưu tiên hơn?

2. Về mức độ lý thuyết cần trình bày trong luận văn: đồ án cần chứng minh tính đúng đắn và phân tích competitive ratio của thuật toán đề xuất, hay có thể tập trung chủ yếu vào phần đánh giá thực nghiệm?

3. Về baseline so sánh trong phần đánh giá: có yêu cầu so sánh với các hệ thống công nghiệp đã tồn tại như distcc, icecream, hay Bazel Remote Execution không, hay chỉ cần so sánh nội bộ giữa các biến thể thuật toán (P2C, Least-Loaded, Round-Robin, Cost-based) là đã đủ?

4. Về benchmark corpus: có yêu cầu dùng bộ dữ liệu chuẩn (SPEC CPU, build Chromium, build LLVM self-host) hay tác giả có thể tự chọn tập corpus đại diện và công bố cùng với kết quả?

5. Về phạm vi phần học trong scheduler: phương pháp EMA đơn giản có đủ làm đóng góp, hay cần so sánh với các phương pháp phức tạp hơn như regression, contextual bandit, hoặc reinforcement learning?
