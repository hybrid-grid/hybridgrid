#!/usr/bin/env python3
"""Idle-skip analysis: why the bandit ties LeastLoaded but beats plain LinUCB.

Reads the same randomized-block output as scripts/analyze_rigorous.py and
answers a narrower question: when the cluster is under-subscribed, the only
mistake a scheduler can still make is passing over an idle worker to queue
work behind a busy one. This script counts that mistake and regresses it
against makespan.

Definitions
-----------
idle-skip rate
    Fraction of dispatches sent to a worker that already had at least one
    task running (worker_active_tasks_at_dispatch > 0).

    LeastLoaded scores 0 by construction -- it always picks the least-loaded
    worker -- so its value is a correctness check on the instrumentation,
    not a finding. Treat it that way when reading the output.

occupancy ceiling
    Concurrent client tasks (make -j) divided by total worker slots. Below
    100% the cluster never saturates, so an idle worker is almost always
    available and the only mistake left is passing one over.

    An earlier version of this docstring went further and claimed no
    scheduler can beat LeastLoaded below 100% occupancy. Measurement at
    -j5 (50% occupancy) refuted it: Hybrid-LinUCB won by 1.11 s, p=0.042.
    The reason is that idle does not mean equivalent -- these workers
    differ by 2.2x in capacity, so "which idle worker" is still a real
    decision,
    and LeastLoaded answers it by task count rather than by capacity.
    Use scripts/analyze_dispatch_availability.py to measure availability
    across the whole cluster rather than only the chosen worker.

The regression is reported per workload. A high R-squared means the metric
explains makespan; a low one means run-to-run variance dominates and the
mechanism claim does not carry for that workload. Both outcomes are results.

Usage: python analyze_idle_skip.py <OUT_DIR>

    OUT_DIR is a rigorous-benchmark directory containing results.csv and
    tasks_<sched>_round<N>.jsonl, or a parent holding light/ and heavy/.
"""

import csv
import glob
import json
import os
import statistics as st
import sys
from collections import defaultdict

try:
    from scipy.stats import wilcoxon
except ImportError:  # analysis still runs; only the paired tests drop out
    wilcoxon = None


def _scheduler_of(path):
    """tasks_hybrid-linucb_round3.jsonl -> ('hybrid-linucb', 3)"""
    stem = os.path.basename(path)[len("tasks_"):-len(".jsonl")]
    sched, tail = stem.rsplit("_", 1)
    digits = "".join(c for c in tail if c.isdigit())
    return sched, int(digits) if digits else 0


def idle_skip(path):
    """(dispatches, dispatches onto an already-busy worker) for one run."""
    total = busy = 0
    with open(path) as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            active = json.loads(line).get("worker_active_tasks_at_dispatch")
            if active is None:
                continue
            total += 1
            busy += active > 0
    return total, busy


def slot_capacity(path):
    """Total execution slots across the workers seen in one run."""
    slots = {}
    with open(path) as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            rec = json.loads(line)
            if rec.get("worker_max_parallel"):
                slots[rec["worker_id"]] = rec["worker_max_parallel"]
    return slots


def r_squared(xs, ys):
    mx, my = st.mean(xs), st.mean(ys)
    sxy = sum((a - mx) * (b - my) for a, b in zip(xs, ys))
    sxx = sum((a - mx) ** 2 for a in xs)
    syy = sum((b - my) ** 2 for b in ys)
    if sxx == 0 or syy == 0:
        return float("nan"), float("nan")
    r = sxy / (sxx * syy) ** 0.5
    return r, sxy / sxx


def analyze(out_dir):
    results = os.path.join(out_dir, "results.csv")
    runs = sorted(glob.glob(os.path.join(out_dir, "tasks_*.jsonl")))
    if not os.path.exists(results) or not runs:
        return False

    print(f"\n=== {out_dir} ===")

    slots = slot_capacity(runs[0])
    if slots:
        total = sum(slots.values())
        print(f"cluster: {len(slots)} workers, {total} slots "
              f"({'+'.join(str(v) for v in sorted(slots.values(), reverse=True))})")
        print(f"         at make -j5 the occupancy ceiling is {500 // total}%"
              f" -- below 100%, an idle worker is almost always available,"
              f"\n         but idle != equivalent: which idle worker still matters"
              f" on a heterogeneous cluster")

    makespan = {}
    with open(results) as fh:
        for row in csv.DictReader(fh):
            makespan[(row["scheduler"], int(row["round"]))] = float(row["elapsed_s"])

    per_sched = defaultdict(lambda: [0, 0])
    points = []
    for path in runs:
        sched, rnd = _scheduler_of(path)
        total, busy = idle_skip(path)
        if not total:
            continue
        per_sched[sched][0] += total
        per_sched[sched][1] += busy
        if (sched, rnd) in makespan:
            points.append((busy / total, makespan[(sched, rnd)]))

    by_median = {}
    for sched in per_sched:
        vals = [v for (s, _), v in makespan.items() if s == sched]
        by_median[sched] = st.median(vals) if vals else float("nan")

    print(f"\n{'scheduler':16}{'dispatches':>12}{'idle-skip':>12}{'makespan med':>15}")
    for sched in sorted(per_sched, key=lambda s: by_median[s]):
        total, busy = per_sched[sched]
        note = "  <- 0% by construction" if busy == 0 else ""
        print(f"{sched:16}{total:12}{100 * busy / total:11.1f}%"
              f"{by_median[sched]:14.2f}s{note}")

    if len(points) > 2:
        r, slope = r_squared([p[0] for p in points], [p[1] for p in points])
        print(f"\nidle-skip vs makespan over {len(points)} runs: "
              f"r={r:.3f}  R^2={r * r:.3f}")
        print(f"  slope: +10 points of idle-skip -> {slope * 0.10:+.2f}s makespan")
        if r * r < 0.4:
            print("  R^2 is low: run-to-run variance dominates here, so the "
                  "mechanism does not explain makespan for this workload")

    if wilcoxon is not None:
        rounds = defaultdict(dict)
        for (sched, rnd), val in makespan.items():
            rounds[rnd][sched] = val
        ref = "hybrid-linucb"
        others = sorted({s for v in rounds.values() for s in v} - {ref})
        raw = {}
        for other in others:
            pairs = [(v[other], v[ref]) for v in rounds.values()
                     if other in v and ref in v]
            if len(pairs) < 5:
                continue
            a = [p[0] for p in pairs]
            b = [p[1] for p in pairs]
            raw[other] = (st.median([x - y for x, y in zip(a, b)]),
                          wilcoxon(a, b, alternative="greater").pvalue)
        if raw:
            # Holm-Bonferroni over the pairwise family, matching analyze_rigorous.py
            ordered = sorted(raw.items(), key=lambda kv: kv[1][1])
            n = len(ordered)
            print(f"\npaired Wilcoxon, one-sided, vs {ref} (Holm-corrected):")
            prev = 0.0
            for i, (other, (gap, p)) in enumerate(ordered):
                p_holm = max(prev, min(1.0, p * (n - i)))
                prev = p_holm
                verdict = "faster" if p_holm < 0.05 else "tie"
                print(f"  {other:16}{gap:+7.2f}s   p={p_holm:.3f}  {verdict}")
    else:
        print("\n(scipy not installed -- skipping the paired tests)")
    return True


def main():
    if len(sys.argv) != 2:
        print(__doc__.strip().split("Usage:")[-1].strip(), file=sys.stderr)
        return 2
    root = sys.argv[1]
    found = analyze(root)
    for sub in ("light", "heavy"):
        found |= analyze(os.path.join(root, sub))
    if not found:
        print(f"no results.csv + tasks_*.jsonl under {root}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
