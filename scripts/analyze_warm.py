#!/usr/bin/env python3
"""Learning-curve analysis for the warm-bandit benchmark.

Reads scripts/benchmark_warm.sh output: results.csv with columns
scheduler,session,build_idx,workers,elapsed_s. Each (scheduler, session) is
one learning curve over build_idx = 1..K, with SESSIONS independent curves
per scheduler.

Question: does a persistent bandit get FASTER across sequential builds while a
stateless heuristic stays flat? We test three things per scheduler:

  1. Learning trend — Spearman rank correlation between build_idx and makespan
     (negative => faster over time). Reported per scheduler.
  2. build-1 vs build-K — paired Wilcoxon signed-rank across sessions
     (is the last build faster than the first?), with median drop + bootstrap CI.
  3. Warm vs cold advantage — at the final build index, paired Wilcoxon of each
     learner against leastloaded across sessions (does a WARM bandit finally
     beat the heuristic it only tied when cold?).

Non-significance is a reported result, not an error. Exits non-zero only on I/O.

Usage: python analyze_warm.py <OUT_DIR>
"""

import os
import sys

import numpy as np
import pandas as pd

try:
    from scipy.stats import spearmanr, wilcoxon

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

CONTROL = "leastloaded"
ALPHA_LEVEL = 0.05
BOOT_N = 10000
BOOT_SEED = 20260709


def boot_ci_median(x, rng, n=BOOT_N, lo=2.5, hi=97.5):
    x = np.asarray(x, float)
    if len(x) < 2:
        return (float("nan"), float("nan"))
    idx = rng.integers(0, len(x), size=(n, len(x)))
    meds = np.median(x[idx], axis=1)
    return float(np.percentile(meds, lo)), float(np.percentile(meds, hi))


def load(out_dir):
    df = pd.read_csv(os.path.join(out_dir, "results.csv"))
    df["elapsed_s"] = pd.to_numeric(df["elapsed_s"], errors="coerce")
    df = df.dropna(subset=["elapsed_s"])
    if df.empty:
        raise IOError("no usable rows in results.csv")
    return df


def per_build_table(df):
    """mean/median makespan per (scheduler, build_idx)."""
    g = (df.groupby(["scheduler", "build_idx"])["elapsed_s"]
         .agg(["count", "mean", "std", "median"]).round(2))
    print("\n=== Makespan by build index (learning curve) ===")
    print(g.to_string())
    return g


def learning_trend(df, rng):
    print("\n=== Learning trend (Spearman: build_idx vs makespan) ===")
    print("  negative rho => scheduler gets FASTER across sequential builds")
    for sched in sorted(df["scheduler"].unique()):
        s = df[df["scheduler"] == sched]
        if HAVE_SCIPY and s["build_idx"].nunique() >= 3:
            rho, p = spearmanr(s["build_idx"], s["elapsed_s"])
            tag = "LEARNING" if (rho < 0 and p < ALPHA_LEVEL) else \
                  ("flat/none" if p >= ALPHA_LEVEL else "SLOWER")
            print(f"  {sched:16s} rho={rho:+.3f}  p={p:.4f}  -> {tag}")
        else:
            print(f"  {sched:16s} [skip] need scipy + >=3 build indices")


def first_vs_last(df, rng):
    """Paired Wilcoxon of build 1 vs build K within each session."""
    kmax = int(df["build_idx"].max())
    print(f"\n=== build 1 vs build {kmax} (paired across sessions) ===")
    if not HAVE_SCIPY:
        print("  [skip] scipy not installed")
        return
    for sched in sorted(df["scheduler"].unique()):
        s = df[df["scheduler"] == sched]
        piv = s.pivot_table(index="session", columns="build_idx",
                            values="elapsed_s", aggfunc="mean")
        if 1 not in piv.columns or kmax not in piv.columns:
            print(f"  {sched:16s} [skip] missing build 1 or {kmax}")
            continue
        pair = piv[[1, kmax]].dropna()
        if len(pair) < 2:
            print(f"  {sched:16s} [skip] <2 paired sessions")
            continue
        b1, bk = pair[1].values, pair[kmax].values
        try:
            stat, p = wilcoxon(bk, b1, alternative="less")  # is build K < build 1?
        except ValueError:
            stat, p = float("nan"), 1.0
        drop = float(np.median(bk - b1))
        lo, hi = boot_ci_median(bk - b1, rng)
        verd = "SIGNIFICANT speedup" if p < ALPHA_LEVEL else "no significant change"
        print(f"  {sched:16s} n={len(pair)}  median Δ(bK-b1)={drop:+.2f}s "
              f"CI[{lo:+.2f},{hi:+.2f}]  p={p:.4f} -> {verd}")


def warm_vs_control(df, rng):
    """At the final build index, is each learner faster than leastloaded?"""
    if CONTROL not in df["scheduler"].unique():
        print(f"\n[skip] no {CONTROL} control for warm-vs-control test")
        return
    kmax = int(df["build_idx"].max())
    print(f"\n=== WARM (build {kmax}) learner vs {CONTROL} (paired across sessions) ===")
    if not HAVE_SCIPY:
        print("  [skip] scipy not installed")
        return
    ctrl = (df[(df["scheduler"] == CONTROL) & (df["build_idx"] == kmax)]
            .set_index("session")["elapsed_s"])
    for sched in sorted(df["scheduler"].unique()):
        if sched == CONTROL:
            continue
        warm = (df[(df["scheduler"] == sched) & (df["build_idx"] == kmax)]
                .set_index("session")["elapsed_s"])
        j = pd.concat([warm.rename("w"), ctrl.rename("c")], axis=1).dropna()
        if len(j) < 2:
            print(f"  {sched:16s} [skip] <2 paired sessions")
            continue
        try:
            stat, p = wilcoxon(j["w"].values, j["c"].values, alternative="less")
        except ValueError:
            stat, p = float("nan"), 1.0
        md = float(np.median(j["w"].values - j["c"].values))
        lo, hi = boot_ci_median(j["w"].values - j["c"].values, rng)
        verd = "WARM BANDIT WINS" if p < ALPHA_LEVEL else "still tie/loss"
        print(f"  {sched:16s} vs {CONTROL}: median Δ={md:+.2f}s "
              f"CI[{lo:+.2f},{hi:+.2f}]  p={p:.4f} -> {verd}")


def plot_curves(df, out_dir):
    if not HAVE_MPL:
        print("[skip] matplotlib not installed — no plot")
        return
    fig, ax = plt.subplots(figsize=(9, 5.5))
    for sched in sorted(df["scheduler"].unique()):
        s = df[df["scheduler"] == sched]
        agg = s.groupby("build_idx")["elapsed_s"].agg(["mean", "std", "count"])
        x = agg.index.values
        y = agg["mean"].values
        se = (agg["std"] / np.sqrt(agg["count"].clip(lower=1))).values
        ax.plot(x, y, marker="o", linewidth=2, label=sched)
        ax.fill_between(x, y - se, y + se, alpha=0.15)
    ax.set_xlabel("sequential build index (bandit warms up →)")
    ax.set_ylabel("makespan (s), mean ± SE across sessions")
    ax.set_title("Warm-bandit learning curves — persistent coordinator")
    ax.legend()
    ax.grid(alpha=0.3)
    fig.tight_layout()
    path = os.path.join(out_dir, "warm_learning_curve.png")
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

    df = load(out_dir)
    per_build_table(df).to_csv(os.path.join(out_dir, "warm_by_build.csv"))
    learning_trend(df, rng)
    first_vs_last(df, rng)
    warm_vs_control(df, rng)
    plot_curves(df, out_dir)
    print("\nDone. Artifacts in", out_dir)


if __name__ == "__main__":
    main()
