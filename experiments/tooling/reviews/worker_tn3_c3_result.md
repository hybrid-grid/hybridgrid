Created [gen_synthetic_cpp.py](</D:/Đồ án tốt nghiệp/hybridgrid/scripts/gen_synthetic_cpp.py>) and [test_gen_synthetic.py](</D:/Đồ án tốt nghiệp/hybridgrid/scripts/tests/test_gen_synthetic.py>). No other repository files were changed or committed.

The generator draws seeded, capped Pareto work units. Separate C functions and explicitly instantiated C++ templates make preprocessed size grow with visible work; short recursive C++ template chains add compile cost with little source text when hidden cost is enabled. It writes a shuffled CMake source list and a manifest with per-file and aggregate hashes.

Syntax checks passed. The full unittest run was blocked by this sandbox’s permission error when Python accessed its temporary directory; Claude should run:

```bash
/d/bench-work/bin/python3 -m unittest discover -s scripts/tests -p 'test_gen_synthetic.py' -v
/d/bench-work/bin/python3 scripts/gen_synthetic_cpp.py --out D:/bench-work/exp3-syn --seed 2026 --files 300 --hidden-cost-fraction 0.0
```