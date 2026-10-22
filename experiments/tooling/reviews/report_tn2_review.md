The main confirmatory statistics are numerically correct: medians, paired deltas, CIs, \(p_{\text{Holm}}\), \(r\), counts, runtimes, lambda medians, and manipulation-check ratios match the supplied data. H2.4 is correctly presented as exploratory and untested.

Required corrections:

1. **Incorrect provenance claim** ([lines 5–6](</D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/bao-cao-tn2-261021.md:5>))  
   Wrong: “Mọi con số … lấy từ `analyze_exp.py` … và đã được tính lại độc lập.”  
   Correct: Analyzer supplies the inferential tables; `campaign.log`/`failures.csv` supply counts, incidents, and hours; `tn2_manipulation.py` supplies the exploratory manipulation check. The independent recomputation verified 70 rows of medians, paired deltas, \(p\), \(p_{\text{Holm}}\), and \(r\), but did **not** recompute bootstrap CIs.

2. **Summary arithmetic/range wording**  
   Wrong: “4 kịch bản × 20 block = 620 cell.”  
   Correct: \(3\times7\times20 + 10\times20 = 620\) measured cells across 80 blocks.  
   Wrong ([line 20](</D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/bao-cao-tn2-261021.md:20>)): HG versus LL improved “3–7 s.”  
   Correct paired-median range: **2.0–6.3 s** (6.2705, 3.4575, 1.9900 s).

3. **Confirmatory table does not fully follow the plan’s reporting rule**  
   Add `n_zero`, `p_raw`, and explicit observed direction to every registered component:

   - H2.1a: `n_zero=1`, `p_raw=0.793527`; direction is mixed—marginal medians and \(r=-0.0684\) favor D, but paired median \(\Delta=+2.5\) does not.
   - H2.1b: `n_zero=0`, `p_raw=0.409098`; hypothesized direction.
   - H2.2: `n_zero=0`, `p_raw=0.985435`; paired \(\Delta=-0.002662\), opposite the hypothesized positive direction.
   - H2.3a: `n_zero=0`, `p_raw=0.036234`; hypothesized direction.
   - H2.3b: `n_zero=0`, `p_raw=0.058258`; hypothesized direction.

   Replace “Cả ba giả thuyết … không được ủng hộ” with:  
   **“Tóm tắt trình bày, không phải quy tắc phân loại đăng ký: không thành phần nào trong năm thành phần H2.1a–H2.3b đạt \(p_{\text{Holm}}<0.05\) theo hướng giả thuyết.”**

4. **Non-significance is repeatedly described as equivalence**  
   Replace “ngang SED,” “ngang HG-LinUCB,” “tỷ lệ như nhau,” and “không mất gì” with **“không phát hiện khác biệt có ý nghĩa; không có kiểm định tương đương/non-inferiority.”**  
   Likewise, replace “không học nhanh hơn” with **“không có bằng chứng thống kê rằng thích nghi nhanh hơn.”**  
   The heading “Hiệu ghép cặp có ý nghĩa sau Holm” is also wrong because its table contains nonsignificant rows. Use **“Một số hiệu ghép cặp; in đậm khi \(p_{\text{Holm}}<0.05\).”**  
   “Discount … có hại” should be explicitly labelled a **descriptive, non-hypothesis observation**, as required by the plan.

5. **Interpretation corrections**  
   The capacity arithmetic is correct: \(1.1/1.87=0.588\), total \(=3.488\), share \(=16.86\%\). However:

   - Exact E2-perm scheduler medians are **18.72–20.46%**, not “19–21%.”
   - “Gần mức tối ưu” is unsupported; call 16.9% a **capacity-proportional heuristic benchmark**, not a demonstrated optimum.
   - LL and SED are indeed fastest in E2-perm: medians **26 and 23** decisions; the other arms are 32–50.
   - Nominal on/off phases were not approximately 15 s: observed median durations were **14.116 s on** and **14.085 s off**.
   - “khi dispatch thứ 120” should say **configured at `DRIFT_AT=120`; recorded `drift_on` was 127–132**.
   - “median đổi dưới 0,6 s” is false: the maximum makespan-median change after dropping block 15 is **0.620 s**, for `hybrid-linucb-d-g098-l100`.
   - The 204.550 s resumed LL cell was **6.45%** above its 192.1475 s median; “do máy nguội” must be labelled an interpretation, not a demonstrated cause.

6. **Required plan §6 outputs are missing**  
   The report omits the mandated per-scenario descriptive tables for build-1/build-2 makespan, pre/post/early/late shares, censoring rates, excess share, recovery/phase metrics, and last finisher. It also omits E2-onoff event dispatch counts, task counts per phase, build-boundary phases, and the registered sliding-share figure. Add that E2-trans had exactly one non-cleanup on/off pair in **140/140 cells**. The manipulation appendix should report **0 dropped cells and 0 dropped tasks**.

7. **Incident disclosure is incomplete**  
   Replace the current “6/140 cell” disclosure with the full result: **19/140 E2-onoff cells had an ordinary `drift_off` followed by another cleanup `drift_off` at dispatch 590; six of those additionally had a task-JSONL `drift_on` at 590 absent from the events CSV.** No task followed and cleanup succeeded.  
   For literal completeness, disclose the harmless stats-sampler shutdown exceptions: **102 total** (`19/26/24/33` in perm/trans/onoff/lambda), which affected only the final stats sample, not task or outcome data.

NEEDS CORRECTIONS