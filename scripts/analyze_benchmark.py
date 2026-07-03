#!/usr/bin/env python3
"""Analyze scheduler benchmark results from benchmark_statistical.sh.

Usage: python analyze_benchmark.py <OUT_DIR>

Reads <OUT_DIR>/results.csv (scheduler,run,workers,elapsed_s) and the
per-run task logs tasks_<scheduler>_run<N>.jsonl, then writes:

  summary.csv          mean/std/min/max makespan per scheduler
  makespan_boxplot.png makespan distribution per scheduler
  warmup_curve.png     early-phase compile-time comparison (linucb vs hybrid)

and prints Mann-Whitney U significance tests of hybrid-linucb against
each baseline. Statistical "not significant" is a report, not an error:
the script exits non-zero only on I/O problems.
"""

import glob
import os
import re
import sys

import pandas as pd

try:
    from scipy.stats import mannwhitneyu

    HAVE_SCIPY = True
except ImportError:
    HAVE_SCIPY = False

try:
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt

    HAVE_MPL = True
except ImportError:
    HAVE_MPL = False

HYBRID = "hybrid-linucb"
BASELINES = ["leastloaded", "p2c", "linucb"]
ALPHA_LEVEL = 0.05
WARMUP_TASKS = 150
ROLLING_WINDOW = 20


def load_results(out_dir):
    path = os.path.join(out_dir, "results.csv")
    df = pd.read_csv(path)
    df["elapsed_s"] = pd.to_numeric(df["elapsed_s"], errors="coerce")
    df = df.dropna(subset=["elapsed_s"])
    if df.empty:
        raise IOError(f"no usable rows in {path}")
    return df


def summarize_makespan(df, out_dir):
    summary = (
        df.groupby("scheduler")["elapsed_s"]
        .agg(["count", "mean", "std", "min", "median", "max"])
        .round(2)
        .sort_values("mean")
    )
    print("\n=== Makespan (s) per scheduler ===")
    print(summary.to_string())
    summary.to_csv(os.path.join(out_dir, "summary.csv"))
    return summary


def mann_whitney(df, metric_name, groups):
    """One-sided Mann-Whitney U: is hybrid's metric LESS than baseline's?"""
    if not HAVE_SCIPY:
        print(f"[skip] scipy not installed — no significance tests for {metric_name}")
        return
    if HYBRID not in groups:
        print(f"[skip] no {HYBRID} samples for {metric_name}")
        return
    hybrid = groups[HYBRID]
    print(f"\n=== Mann-Whitney U ({metric_name}): {HYBRID} < baseline? ===")
    for base in BASELINES:
        if base not in groups or len(groups[base]) < 2 or len(hybrid) < 2:
            print(f"  vs {base:12s}: insufficient samples")
            continue
        stat, p = mannwhitneyu(hybrid, groups[base], alternative="less")
        med_diff = pd.Series(hybrid).median() - pd.Series(groups[base]).median()
        verdict = "SIGNIFICANT" if p < ALPHA_LEVEL else "not significant"
        print(
            f"  vs {base:12s}: U={stat:8.1f}  p={p:.4f}  "
            f"median diff={med_diff:+.1f}  -> {verdict} (α={ALPHA_LEVEL})"
        )


def boxplot_makespan(df, out_dir):
    if not HAVE_MPL:
        print("[skip] matplotlib not installed — no boxplot")
        return
    order = df.groupby("scheduler")["elapsed_s"].mean().sort_values().index
    data = [df[df["scheduler"] == s]["elapsed_s"] for s in order]
    fig, ax = plt.subplots(figsize=(8, 5))
    ax.boxplot(data, tick_labels=list(order))
    ax.set_ylabel("makespan (s)")
    ax.set_title("CPython 873-task build, 5w-hetero cluster")
    fig.tight_layout()
    path = os.path.join(out_dir, "makespan_boxplot.png")
    fig.savefig(path, dpi=150)
    print(f"wrote {path}")


def load_task_logs(out_dir):
    """Return {scheduler: [per-run DataFrame, ...]} sorted by timestamp."""
    logs = {}
    pattern = os.path.join(out_dir, "tasks_*_run*.jsonl")
    rx = re.compile(r"tasks_(.+)_run(\d+)\.jsonl$")
    for path in sorted(glob.glob(pattern)):
        m = rx.search(os.path.basename(path))
        if not m:
            continue
        sched = m.group(1)
        try:
            df = pd.read_json(path, lines=True)
        except ValueError as e:
            print(f"[warn] unreadable {path}: {e}")
            continue
        if df.empty or "compile_time_ms" not in df.columns:
            continue
        if "ts" in df.columns:
            df = df.sort_values("ts").reset_index(drop=True)
        logs.setdefault(sched, []).append(df)
    return logs


def tail_latency(logs, out_dir):
    p99_groups = {}
    rows = []
    for sched, runs in logs.items():
        p99s = [run["compile_time_ms"].quantile(0.99) for run in runs]
        p99_groups[sched] = p99s
        rows.append(
            {
                "scheduler": sched,
                "runs": len(p99s),
                "p99_mean_ms": round(pd.Series(p99s).mean(), 0),
                "p99_std_ms": round(pd.Series(p99s).std(), 0),
            }
        )
    if not rows:
        print("[skip] no task logs found — no tail-latency analysis")
        return
    table = pd.DataFrame(rows).sort_values("p99_mean_ms").set_index("scheduler")
    print("\n=== Per-run P99 compile time (ms) ===")
    print(table.to_string())
    table.to_csv(os.path.join(out_dir, "p99_summary.csv"))
    mann_whitney(None, "per-run P99 compile_time_ms", p99_groups)


def warmup_curve(logs, out_dir):
    """Early-phase behaviour: does warm-start avoid the cold-start spike?"""
    if not HAVE_MPL:
        print("[skip] matplotlib not installed — no warm-up curve")
        return
    targets = [s for s in ("linucb", HYBRID) if s in logs]
    if len(targets) < 2:
        print("[skip] need both linucb and hybrid-linucb logs for warm-up curve")
        return
    fig, axes = plt.subplots(2, 1, figsize=(9, 8), sharex=True)
    for sched in targets:
        curves_ct, curves_q = [], []
        for run in logs[sched]:
            head = run.head(WARMUP_TASKS)
            curves_ct.append(
                head["compile_time_ms"].rolling(ROLLING_WINDOW, min_periods=1).mean()
            )
            if "q_value_at_dispatch" in head.columns:
                curves_q.append(
                    head["q_value_at_dispatch"]
                    .rolling(ROLLING_WINDOW, min_periods=1)
                    .mean()
                )
        axes[0].plot(
            pd.concat(curves_ct, axis=1).mean(axis=1), label=sched, linewidth=2
        )
        if curves_q:
            axes[1].plot(
                pd.concat(curves_q, axis=1).mean(axis=1), label=sched, linewidth=2
            )
    axes[0].set_ylabel(f"compile_time_ms (rolling {ROLLING_WINDOW})")
    axes[0].set_title(f"Warm-up: first {WARMUP_TASKS} tasks, averaged across runs")
    axes[0].axvline(100, color="gray", linestyle="--", label="warm-start end (N=100)")
    axes[0].legend()
    axes[1].set_ylabel(f"q_value_at_dispatch (rolling {ROLLING_WINDOW})")
    axes[1].set_xlabel("task index")
    axes[1].axvline(100, color="gray", linestyle="--")
    axes[1].legend()
    fig.tight_layout()
    path = os.path.join(out_dir, "warmup_curve.png")
    fig.savefig(path, dpi=150)
    print(f"wrote {path}")


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(2)
    out_dir = sys.argv[1]
    if not os.path.isdir(out_dir):
        print(f"ERROR: not a directory: {out_dir}")
        sys.exit(1)

    df = load_results(out_dir)
    summarize_makespan(df, out_dir)

    makespan_groups = {
        s: g["elapsed_s"].tolist() for s, g in df.groupby("scheduler")
    }
    mann_whitney(df, "makespan", makespan_groups)
    boxplot_makespan(df, out_dir)

    logs = load_task_logs(out_dir)
    tail_latency(logs, out_dir)
    warmup_curve(logs, out_dir)

    print("\nDone. Artifacts in", out_dir)


if __name__ == "__main__":
    main()
