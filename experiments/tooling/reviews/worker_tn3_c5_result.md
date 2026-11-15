Changed [scripts/analyze_exp.py](</D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py>) and [scripts/tests/test_analyze_exp.py](</D:/Đồ án tốt nghiệp/hybridgrid/scripts/tests/test_analyze_exp.py>). No other files were edited or committed.

The analyzer accepts `p90` and `sed-p90` as default baselines. For each build, it computes the top 5% placement share on the two strongest workers, mean assigned worker CPU, share with at least two free-slot candidates, tail time, and the last finisher’s size and percentile rank. Multi-client tasks are pooled by build; cell-level values come from the last build. Results without the new timestamp columns retain their legacy CSV layout.

The 18 tests that do not write temporary files passed, as did syntax and diff checks. The full file-based suite remains to be run in Claude’s environment:

```bash
python3 -m unittest scripts/tests/test_analyze_exp.py
```