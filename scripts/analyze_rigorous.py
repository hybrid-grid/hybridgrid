#!/usr/bin/env python3
"""Paired statistical analysis for the randomized-block scheduler benchmark.

Reads the output of scripts/benchmark_rigorous.sh, where every round runs
all schedulers once (a statistical block). This pairing lets us use tests
that control for round-level nuisance (thermal drift, background load):

  - Friedman test           omnibus: do schedulers differ at all?
  - Wilcoxon signed-rank     paired, one-sided: hybrid-linucb < baseline?
  - Holm-Bonferroni          family-wise correction over the 3 pairwise tests
  - Cliff's delta            non-parametric effect size per pair
  - Bootstrap 95% CI         for per-scheduler median and paired differences

Two metrics: makespan (wall-clock, from results.csv) and per-round P99
compile time (from tasks_<sched>_round<N>.jsonl). Statistical
non-significance is a reported result, not an error; the script exits
non-zero only on I/O problems.

Usage: python analyze_rigorous.py <OUT_DIR>
"""

import glob
import os
import re
import sys

import numpy as np
import pandas as pd

try:
    from scipy.stats import friedmanchisquare, wilcoxon

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
BOOT_N = 10000
BOOT_SEED = 20260709


# --------------------------------------------------------------------------
# effect size + bootstrap helpers
# --------------------------------------------------------------------------
def cliffs_delta(a, b):
    """P(a>b) - P(a<b). Negative => a tends to be smaller (better for time)."""
    a = np.asarray(a, float)
    b = np.asarray(b, float)
    gt = sum((x > b).sum() for x in a)
    lt = sum((x < b).sum() for x in a)
    return (gt - lt) / (len(a) * len(b))


def cliff_magnitude(d):
    ad = abs(d)
    if ad < 0.147:
        return "negligible"
    if ad < 0.33:
        return "small"
    if ad < 0.474:
        return "medium"
    return "large"


def boot_ci_median(x, rng, n=BOOT_N, lo=2.5, hi=97.5):
    x = np.asarray(x, float)
    if len(x) < 2:
        return (float("nan"), float("nan"))
    idx = rng.integers(0, len(x), size=(n, len(x)))
    meds = np.median(x[idx], axis=1)
    return (float(np.percentile(meds, lo)), float(np.percentile(meds, hi)))


def boot_ci_paired_diff(a, b, rng, n=BOOT_N, lo=2.5, hi=97.5):
    """CI for median(a - b) on paired samples (a,b aligned by block)."""
    d = np.asarray(a, float) - np.asarray(b, float)
    return boot_ci_median(d, rng, n, lo, hi)


def holm(pvals):
    """Holm-Bonferroni: return adjusted p-values in original order."""
    m = len(pvals)
    order = sorted(range(m), key=lambda i: pvals[i])
    adj = [0.0] * m
    running = 0.0
    for rank, i in enumerate(order):
        val = (m - rank) * pvals[i]
        running = max(running, val)
        adj[i] = min(running, 1.0)
    return adj


# --------------------------------------------------------------------------
# data loading
# --------------------------------------------------------------------------
def load_makespan(out_dir):
    df = pd.read_csv(os.path.join(out_dir, "results.csv"))
    df["elapsed_s"] = pd.to_numeric(df["elapsed_s"], errors="coerce")
    df = df.dropna(subset=["elapsed_s"])
    if df.empty:
        raise IOError("no usable rows in results.csv")
    return df


def load_p99(out_dir):
    """Return long DataFrame: scheduler, round, p99_ms (one row per cell)."""
    rx = re.compile(r"tasks_(.+)_round(\d+)\.jsonl$")
    rows = []
    for path in sorted(glob.glob(os.path.join(out_dir, "tasks_*_round*.jsonl"))):
        m = rx.search(os.path.basename(path))
        if not m:
            continue
        sched, rnd = m.group(1), int(m.group(2))
        try:
            d = pd.read_json(path, lines=True)
        except ValueError:
            continue
        if d.empty or "compile_time_ms" not in d.columns:
            continue
        rows.append(
            {"scheduler": sched, "round": rnd,
             "p99_ms": float(d["compile_time_ms"].quantile(0.99))}
        )
    return pd.DataFrame(rows)


def to_blocks(long_df, value_col):
    """Wide matrix: index=round, columns=scheduler, values=value_col.
    Drops rounds missing any scheduler so every column is fully paired."""
    wide = long_df.pivot_table(index="round", columns="scheduler",
                               values=value_col, aggfunc="mean")
    return wide.dropna(axis=0, how="any")


# --------------------------------------------------------------------------
# analysis of one metric
# --------------------------------------------------------------------------
def analyze_metric(name, wide, unit, rng, out_dir):
    print(f"\n{'=' * 70}\n{name} — paired analysis ({len(wide)} complete blocks)\n{'=' * 70}")
    scheds = [s for s in [HYBRID] + BASELINES if s in wide.columns]
    if not scheds:
        print("  no schedulers present")
        return

    # Per-scheduler summary + bootstrap CI for the median.
    print(f"\nPer-scheduler {name} ({unit}):")
    print(f"  {'scheduler':16s} {'n':>3s} {'median':>9s} {'mean':>9s} "
          f"{'std':>7s}  95% CI (median)")
    summ = wide[scheds].median().sort_values()
    for s in summ.index:
        col = wide[s].dropna()
        lo, hi = boot_ci_median(col.values, rng)
        print(f"  {s:16s} {len(col):3d} {col.median():9.2f} {col.mean():9.2f} "
              f"{col.std():7.2f}  [{lo:.2f}, {hi:.2f}]")

    # Friedman omnibus (needs >=3 schedulers, >=2 blocks).
    if HAVE_SCIPY and len(scheds) >= 3 and len(wide) >= 2:
        stat, p = friedmanchisquare(*[wide[s].values for s in scheds])
        verdict = "schedulers DIFFER" if p < ALPHA_LEVEL else "no detectable difference"
        print(f"\nFriedman omnibus: chi2={stat:.3f}  p={p:.5f}  -> {verdict}")
    else:
        print("\n[skip] Friedman needs scipy + >=3 schedulers")

    # Pairwise Wilcoxon signed-rank (hybrid < baseline), paired by block.
    if not (HAVE_SCIPY and HYBRID in wide.columns):
        print(f"[skip] no paired Wilcoxon for {name}")
        return
    pairs, raw_p = [], []
    for base in BASELINES:
        if base not in wide.columns:
            continue
        h, b = wide[HYBRID].values, wide[base].values
        try:
            stat, p = wilcoxon(h, b, alternative="less")
        except ValueError:
            stat, p = float("nan"), 1.0  # all-zero differences
        pairs.append((base, stat, p, h, b))
        raw_p.append(p)
    adj_p = holm(raw_p)

    print(f"\nWilcoxon signed-rank ({HYBRID} < baseline), Holm-corrected:")
    print(f"  {'vs baseline':14s} {'W':>7s} {'p_raw':>8s} {'p_holm':>8s} "
          f"{'med diff':>10s}  95% CI diff        Cliff d     verdict")
    for (base, stat, p, h, b), pa in zip(pairs, adj_p):
        md = float(np.median(h - b))
        lo, hi = boot_ci_paired_diff(h, b, rng)
        d = cliffs_delta(h, b)
        verd = "SIGNIFICANT" if pa < ALPHA_LEVEL else "n.s."
        sw = f"{stat:7.1f}" if stat == stat else "    nan"
        print(f"  {base:14s} {sw} {p:8.4f} {pa:8.4f} {md:+10.2f} "
              f"  [{lo:+.2f}, {hi:+.2f}]  {d:+.3f} ({cliff_magnitude(d):10s}) {verd}")

    return scheds


def plot_makespan(wide, out_dir):
    if not HAVE_MPL:
        return
    scheds = [s for s in [HYBRID] + BASELINES if s in wide.columns]
    order = list(wide[scheds].median().sort_values().index)

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(13, 5))
    ax1.boxplot([wide[s].dropna().values for s in order], tick_labels=order)
    ax1.set_ylabel("makespan (s)")
    ax1.set_title("Makespan distribution (randomized block design)")
    ax1.tick_params(axis="x", rotation=15)

    # Paired spaghetti: one line per round across schedulers in `order`.
    x = range(len(order))
    for rnd in wide.index:
        ax2.plot(x, [wide.loc[rnd, s] for s in order], marker="o",
                 alpha=0.5, linewidth=1)
    ax2.set_xticks(list(x))
    ax2.set_xticklabels(order, rotation=15)
    ax2.set_ylabel("makespan (s)")
    ax2.set_title("Per-round paired makespan (each line = one block)")
    fig.tight_layout()
    path = os.path.join(out_dir, "makespan_rigorous.png")
    fig.savefig(path, dpi=150)
    print(f"\nwrote {path}")


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(2)
    out_dir = sys.argv[1]
    if not os.path.isdir(out_dir):
        print(f"ERROR: not a directory: {out_dir}")
        sys.exit(1)
    rng = np.random.default_rng(BOOT_SEED)

    mk = load_makespan(out_dir)
    mk_wide = to_blocks(mk, "elapsed_s")
    analyze_metric("MAKESPAN", mk_wide, "s", rng, out_dir)
    plot_makespan(mk_wide, out_dir)

    p99 = load_p99(out_dir)
    if not p99.empty:
        p99_wide = to_blocks(p99, "p99_ms")
        analyze_metric("P99 COMPILE TIME", p99_wide, "ms", rng, out_dir)
    else:
        print("\n[skip] no task logs — no P99 analysis")

    # Persist the paired makespan matrix for the thesis appendix.
    mk_wide.to_csv(os.path.join(out_dir, "makespan_blocks.csv"))
    print("\nDone. Artifacts in", out_dir)


if __name__ == "__main__":
    main()
