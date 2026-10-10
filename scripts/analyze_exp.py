#!/usr/bin/env python3
"""Analyze complete randomized blocks from bench_exp.sh (or old rigorous runs).

Usage: python scripts/analyze_exp.py OUT_DIR [--out DIR] [--strict]
"""

import argparse
import json
import math
import re
from pathlib import Path

import numpy as np
import pandas as pd
from scipy.stats import rankdata, wilcoxon


SUFFIX = re.compile(r"^(?P<base>.+?)(?P<arm>-arm)?(?P<gamma>-g\d{3})?(?P<penalty>-l(?:0|\d{3}))?$")
TASK_NAME = re.compile(r"^tasks_(.+)_round(\d+)\.jsonl$")
BASES = {"leastloaded", "simple", "p2c", "epsilon-greedy", "linucb",
         "hybrid-linucb", "hybrid-linucb-d", "heft", "icecc-fastest", "sed"}
DEFAULT_TARGET = "worker-5:50051"
BOOT_SEED = 20260709
BOOT_N = 10000
METRICS = (
    "makespan_s", "skip_idle_all", "skip_idle_given_idle", "strongest_idle",
    "free_slot_skip_all", "free_slot_skip_given_free", "strongest_free_slot",
    "drift_pre_share", "drift_post_share", "time_to_adapt_rank",
    "target_after_recovery_share", "target_after_recovery_tasks",
)


def parse_arm(label):
    """Return the D2 base and resolved options, rejecting malformed suffixes."""
    match = SUFFIX.fullmatch(label)
    if not match:
        raise ValueError(f"invalid arm label: {label}")
    base = match["base"]
    # A suffix-like fragment in the base means the label did not match D2 order.
    if base not in BASES:
        raise ValueError(f"invalid arm label: {label}")
    if re.search(r"(?:-arm|-g\d+|-l\d+)(?:-|$)", base):
        raise ValueError(f"invalid arm label: {label}")
    if match["gamma"] and base != "hybrid-linucb-d":
        raise ValueError(f"discount only applies to hybrid-linucb-d: {label}")
    if match["arm"] and base != "hybrid-linucb-d":
        raise ValueError(f"arm discount only applies to hybrid-linucb-d: {label}")
    if match["penalty"] and base not in ("hybrid-linucb", "hybrid-linucb-d"):
        raise ValueError(f"load penalty does not apply to {base}: {label}")
    gamma = int(match["gamma"][2:]) / 100 if match["gamma"] else 0.98
    penalty = (int(match["penalty"][2:]) / 100
               if match["penalty"] and match["penalty"] != "-l0" else
               (0.0 if match["penalty"] else 0.5))
    if not 0 < gamma <= 1 or not 0 <= penalty <= 1:
        raise ValueError(f"suffix out of range: {label}")
    return {"base": base, "discount_mode": "arm" if match["arm"] else "global",
            "discount": gamma, "load_penalty": penalty}


def holm(p_values):
    """Adjusted p values, preserving input order and monotonicity."""
    order = sorted(range(len(p_values)), key=lambda i: (p_values[i], i))
    adjusted = [0.0] * len(order)
    running = 0.0
    for rank, index in enumerate(order):
        running = max(running, min(1.0, (len(order) - rank) * p_values[index]))
        adjusted[index] = running
    return adjusted


def rank_biserial(differences):
    """Matched-pairs rank-biserial; negative means treatment is smaller."""
    nonzero = np.asarray([float(x) for x in differences if x != 0], dtype=float)
    if len(nonzero) == 0:
        return 0.0
    ranks = rankdata(np.abs(nonzero), method="average")
    plus = float(ranks[nonzero > 0].sum())
    minus = float(ranks[nonzero < 0].sum())
    return (plus - minus) / (plus + minus)


def paired_stats(treatment, baseline, rng, boot_n=BOOT_N):
    """Wilcoxon uses nonzero pairs; exact only with <=25 untied ranks."""
    a, b = np.asarray(treatment, dtype=float), np.asarray(baseline, dtype=float)
    if len(a) != len(b) or not len(a):
        raise ValueError("paired samples must have equal nonzero length")
    diff = a - b
    nonzero = diff[diff != 0]
    n_zero = len(diff) - len(nonzero)
    if not len(nonzero):
        p_raw = 1.0
    else:
        ties = len(np.unique(np.abs(nonzero))) != len(nonzero)
        method = "exact" if len(nonzero) <= 25 and not ties else "approx"
        p_raw = float(wilcoxon(nonzero, alternative="two-sided",
                               zero_method="wilcox", method=method).pvalue)
    indices = rng.integers(0, len(diff), size=(boot_n, len(diff)))
    medians = np.median(diff[indices], axis=1)
    ci_lo, ci_hi = np.percentile(medians, [2.5, 97.5])
    return {"n": len(diff), "treatment_median": float(np.median(a)),
            "baseline_median": float(np.median(b)),
            "median_diff": float(np.median(diff)), "ci_lo": float(ci_lo),
            "ci_hi": float(ci_hi), "p_raw": p_raw,
            "r_rb": rank_biserial(diff), "n_zero": n_zero}


def _number(value):
    try:
        number = float(value)
        return number if math.isfinite(number) else None
    except (TypeError, ValueError):
        return None


def _share(tasks, target):
    if not tasks or any(not t.get("worker_address") for t in tasks):
        return math.nan
    return sum(t["worker_address"] == target for t in tasks) / len(tasks)


def candidate_metrics(tasks):
    """Exact snapshot-based idle and free-slot rates, or NA for old logs."""
    names = ("skip_idle_all", "skip_idle_given_idle", "strongest_idle",
             "free_slot_skip_all", "free_slot_skip_given_free", "strongest_free_slot")
    na = dict.fromkeys(names, math.nan)
    if not tasks:
        return na
    idle_n = free_n = skipped_idle = skipped_free = strongest_idle = strongest_free = 0
    for task in tasks:
        candidates = task.get("candidates")
        address = task.get("worker_address")
        if not isinstance(candidates, list) or not candidates or not address:
            return na
        if any(_number(c.get("active")) is None or
               _number(c.get("max_parallel")) is None or
               _number(c.get("cpu_millis")) is None or not c.get("address")
               for c in candidates):
            return na
        chosen = next((c for c in candidates if c["address"] == address), None)
        if chosen is None:
            return na
        idle = [c for c in candidates if c["active"] == 0]
        free = [c for c in candidates if c["active"] < c["max_parallel"]]
        if idle:
            idle_n += 1
            skipped_idle += chosen["active"] > 0
            strongest_idle += (chosen["active"] == 0 and chosen["cpu_millis"] ==
                               max(c["cpu_millis"] for c in idle))
        if free:
            free_n += 1
            skipped_free += chosen["active"] >= chosen["max_parallel"]
            strongest_free += (chosen["active"] < chosen["max_parallel"] and
                               chosen["cpu_millis"] == max(c["cpu_millis"] for c in free))
    count = len(tasks)
    return {"skip_idle_all": skipped_idle / count,
            "skip_idle_given_idle": skipped_idle / idle_n if idle_n else math.nan,
            "strongest_idle": strongest_idle / idle_n if idle_n else math.nan,
            "free_slot_skip_all": skipped_free / count,
            "free_slot_skip_given_free": skipped_free / free_n if free_n else math.nan,
            "strongest_free_slot": strongest_free / free_n if free_n else math.nan}


def time_to_adapt(post, target, pre_share, window=20, factor=0.5):
    """Return display value, numeric rank, and censor flag.

    The first complete W-decision window is evaluated at decision W.
    """
    if not math.isfinite(pre_share) or pre_share < 0.05 or not post or any(
            not task.get("worker_address") for task in post):
        return "NA", math.nan, False
    hits = [int(task["worker_address"] == target) for task in post]
    for count in range(window, len(hits) + 1):
        if sum(hits[count - window:count]) / window <= factor * pre_share:
            return str(count), float(count), False
    return f"> {len(post)}", float(len(post) + 1), True


def drift_metrics(tasks, events, target, window, factor):
    result = {"drift_pre_share": math.nan, "drift_post_share": math.nan,
              "time_to_adapt": "NA", "time_to_adapt_rank": math.nan,
              "time_to_adapt_censored": False,
              "target_after_recovery_share": math.nan,
              "target_after_recovery_tasks": math.nan}
    on = next((e for e in events if e.get("kind") == "drift_on"), None)
    if on is None or _number(on.get("dispatch_count")) is None or not tasks or any(
            _number(t.get("dispatch_seq")) is None for t in tasks):
        return result
    # Event is posted after injection; the counter is the last booked dispatch.
    onset = int(on["dispatch_count"]) + 1
    target = on.get("target") or target
    ordered = sorted(tasks, key=lambda t: int(t["dispatch_seq"]))
    pre = [t for t in ordered if 101 <= int(t["dispatch_seq"]) < onset]
    post = [t for t in ordered if int(t["dispatch_seq"]) >= onset]
    result["drift_pre_share"] = _share(pre, target)
    result["drift_post_share"] = _share(post, target)
    display, rank, censored = time_to_adapt(
        post, target, result["drift_pre_share"], window, factor)
    result.update(time_to_adapt=display, time_to_adapt_rank=rank,
                  time_to_adapt_censored=censored)
    phases = [e for e in events if e.get("kind") in ("drift_on", "drift_off")
              and _number(e.get("dispatch_count")) is not None]
    phases.sort(key=lambda e: (int(e["dispatch_count"]), str(e.get("ts", ""))))
    for index, event in enumerate(phases):
        start = int(event["dispatch_count"]) + 1
        end = int(phases[index + 1]["dispatch_count"]) + 1 if index + 1 < len(phases) else math.inf
        phase_tasks = [t for t in ordered if start <= int(t["dispatch_seq"]) < end]
        result[f"drift_phase_{index + 1}_{event['kind'][6:]}_share"] = _share(phase_tasks, target)
    off = phases[1] if (len(phases) == 2 and phases[0]["kind"] == "drift_on"
                        and phases[1]["kind"] == "drift_off") else None
    if off:
        recovered = [t for t in ordered if int(t["dispatch_seq"]) > int(off["dispatch_count"])]
        result["target_after_recovery_share"] = _share(recovered, target)
        if recovered and all(t.get("worker_address") for t in recovered):
            result["target_after_recovery_tasks"] = sum(
                t["worker_address"] == target for t in recovered)
    return result


def load_log(path):
    tasks, events = [], []
    with path.open(encoding="utf-8") as handle:
        for line_number, line in enumerate(handle, 1):
            if not line.strip():
                continue
            try:
                record = json.loads(line)
            except json.JSONDecodeError as error:
                raise ValueError(f"{path.name}:{line_number}: {error}") from error
            if record.get("event", "task_completed") == "task_completed":
                tasks.append(record)
            elif record.get("event") == "injected_event":
                events.append(record)
    return tasks, events


def _meta_arms(meta):
    arms = meta.get("ARMS") or meta.get("arms")
    if isinstance(arms, dict):
        return sorted(arms)
    if isinstance(arms, list):
        return sorted(x if isinstance(x, str) else x.get("label", "") for x in arms)
    if isinstance(arms, str):
        return sorted(arms.split())
    return []


def load_cells(out_dir, window=20, factor=0.5):
    """Return complete cells, drops, and notes. One valid cell per arm/round."""
    meta_path = out_dir / "meta.json"
    meta = json.loads(meta_path.read_text(encoding="utf-8")) if meta_path.exists() else {}
    results = pd.read_csv(out_dir / "results.csv")
    if "arm" not in results.columns and "scheduler" in results.columns:
        results = results.rename(columns={"scheduler": "arm"})
    needed = {"arm", "round", "elapsed_s"}
    if not needed.issubset(results.columns):
        raise ValueError(f"results.csv missing {sorted(needed - set(results.columns))}")
    if results.empty:
        raise ValueError("results.csv is empty")
    results["round"] = pd.to_numeric(results["round"], errors="coerce")
    arms = _meta_arms(meta) or sorted(str(x) for x in results["arm"].dropna().unique())
    for arm in arms:
        parse_arm(arm)
    rounds = sorted(int(x) for x in results["round"].dropna().unique())
    log_keys = {}
    for path in sorted(out_dir.glob("tasks_*_round*.jsonl")):
        match = TASK_NAME.fullmatch(path.name)
        if match:
            log_keys[(match[1], int(match[2]))] = path
    if meta.get("REPS"):
        rounds = list(range(int(meta.get("ROUND_START", 1)), int(meta["REPS"]) + 1))
    else:
        rounds = sorted(set(rounds) | {rnd for _, rnd in log_keys})
        if rounds and min(rounds) >= 1:
            rounds = list(range(min(rounds), max(rounds) + 1))
    expected_clients = int(meta.get("CLIENTS", meta.get("clients", 0)) or 0)
    expected_builds = int(meta.get("SESSION_BUILDS", meta.get("session_builds", 0)) or 0)
    if "client" in results and not expected_clients:
        expected_clients = int(results["client"].nunique())
    if "build_idx" in results and not expected_builds:
        expected_builds = int(pd.to_numeric(results["build_idx"], errors="coerce").max())
    expected_clients = expected_clients or 1
    expected_builds = expected_builds or 1
    target = meta.get("DRIFT_TARGET", DEFAULT_TARGET)
    if target and ":" not in str(target):
        target = f"{target}:50051"
    drift_required = str(meta.get("DRIFT", "none")).lower() != "none"
    cells, dropped, notes = [], [], set()
    if results["round"].isna().any():
        dropped.append("results.csv: rows with invalid round")
    for extra in sorted(set(results["arm"].dropna()) - set(arms)):
        dropped.append(f"results.csv: unlisted arm {extra}")
    for arm, rnd in sorted(log_keys):
        if arm not in arms or rnd not in rounds:
            dropped.append(f"{log_keys[(arm, rnd)].name}: task log outside listed arms/rounds")
    min_tasks = int(meta.get("MIN_TASKS", 1) or 1)
    for rnd in rounds:
        round_cells = []
        round_reasons = []
        for arm in arms:
            rows = results[(results["round"] == rnd) & (results["arm"] == arm)]
            reason = None
            if rows.empty:
                reason = "missing result"
            elif rows["elapsed_s"].map(_number).isna().any() or (rows["elapsed_s"].astype(float) <= 0).any():
                reason = "invalid elapsed_s"
            elif "cell_makespan_s" in rows and (rows["cell_makespan_s"].map(_number).isna().any() or
                                                 (rows["cell_makespan_s"].astype(float) <= 0).any()):
                reason = "invalid cell_makespan_s"
            elif "build_idx" in rows:
                idx = pd.to_numeric(rows["build_idx"], errors="coerce")
                clients = rows["client"] if "client" in rows else pd.Series([1] * len(rows), index=rows.index)
                keys = list(zip(idx, clients))
                if idx.isna().any() or len(keys) != len(set(keys)) or len(rows) != expected_clients * expected_builds or set(idx) != set(range(1, expected_builds + 1)):
                    reason = "missing or duplicate client/build result"
            elif len(rows) != 1:
                reason = "duplicate result"
            path = out_dir / f"tasks_{arm}_round{rnd}.jsonl"
            tasks, events = [], []
            if reason is None:
                if not path.exists():
                    reason = "missing task log"
                else:
                    try:
                        tasks, events = load_log(path)
                    except ValueError as error:
                        reason = str(error)
                    if reason is None and len(tasks) < min_tasks:
                        reason = f"only {len(tasks)} completed tasks; need {min_tasks}"
                    elif reason is None and any(t.get("success") is False for t in tasks):
                        reason = "failed task"
                    elif reason is None and drift_required and not any(e.get("kind") == "drift_on" for e in events):
                        reason = "missing drift_on event"
                    elif reason is None and drift_required and any(
                            e.get("kind") == "drift_on" and
                            (_number(e.get("dispatch_count")) is None or
                             _number(e.get("dispatch_count")) < 101) for e in events):
                        reason = "drift_on before dispatch 101"
            if reason:
                round_reasons.append(f"{arm}: {reason}")
                continue
            row = {"round": rnd, "arm": arm, "n_tasks": len(tasks)}
            row["makespan_s"] = (float(rows.groupby("build_idx")["cell_makespan_s"].max().sum())
                                  if expected_builds > 1 and "cell_makespan_s" in rows else
                                  float(rows["cell_makespan_s"].max())
                                  if "cell_makespan_s" in rows else float(rows["elapsed_s"].max()))
            if "build_idx" in rows:
                for idx, group in rows.groupby("build_idx"):
                    row[f"build_{int(idx)}_makespan_s"] = (float(group["cell_makespan_s"].max())
                                                             if "cell_makespan_s" in group else float(group["elapsed_s"].max()))
                    if "client" in group:
                        for _, client_row in group.iterrows():
                            row[f"build_{int(idx)}_client_{client_row['client']}_elapsed_s"] = float(client_row["elapsed_s"])
                if "client" in rows:
                    for client, group in rows.groupby("client"):
                        row[f"client_{client}_elapsed_s"] = float(group["elapsed_s"].sum())
            row.update(candidate_metrics(tasks))
            row.update(drift_metrics(tasks, events, target, window, factor))
            finishers = [t for t in tasks if t.get("ts")]
            last = max(finishers, key=lambda t: pd.Timestamp(t["ts"])) if finishers else {}
            row["last_finisher_address"] = last.get("worker_address") or "NA"
            row["last_finisher_cpu_millis"] = last.get("worker_cpu_millis", math.nan)
            if math.isnan(row["skip_idle_all"]):
                notes.add("Candidate snapshot metrics unavailable for some cells (old-format task logs).")
            if math.isnan(row["drift_pre_share"]):
                notes.add("Drift metrics unavailable for cells without dispatch sequence and drift events.")
            round_cells.append(row)
        if round_reasons:
            dropped.append(f"round {rnd} dropped (all arms excluded): " + "; ".join(round_reasons))
        else:
            cells.extend(round_cells)
    return pd.DataFrame(cells), dropped, sorted(notes)


def compare(cells, treatments, baselines):
    rows = []
    rng = np.random.default_rng(BOOT_SEED)
    available = sorted(cells["arm"].unique())
    treatments = treatments or [arm for arm in available if parse_arm(arm)["base"] in
                                ("hybrid-linucb", "hybrid-linucb-d")]
    baselines = baselines or [arm for arm in available if parse_arm(arm)["base"] in
                              ("leastloaded", "sed")]
    extra_metrics = tuple(sorted(c for c in cells if c.startswith(("build_", "client_", "drift_phase_"))))
    for metric in METRICS + extra_metrics:
        if metric not in cells:
            continue
        family = []
        for treatment in sorted(treatments):
            for baseline in sorted(baselines):
                if treatment not in available or baseline not in available or treatment == baseline:
                    continue
                pivot = cells.pivot(index="round", columns="arm", values=metric)
                aligned = pivot[[treatment, baseline]].apply(pd.to_numeric, errors="coerce").dropna()
                if aligned.empty:
                    continue
                stats = paired_stats(aligned[treatment], aligned[baseline], rng)
                family.append({"metric": metric, "treatment": treatment,
                               "baseline": baseline, **stats})
        adjusted = holm([entry["p_raw"] for entry in family])
        for entry, p_holm in zip(family, adjusted):
            entry["p_holm"] = p_holm
        rows.extend(family)
    return pd.DataFrame(rows)


def _fmt(value):
    return "NA" if pd.isna(value) else f"{value:.4g}"


def write_summary(path, cells, pairs, drops, notes, window, factor):
    lines = ["# Experiment analysis", "", f"Complete blocks: {cells['round'].nunique() if not cells.empty else 0}.",
             f"Time-to-adapt: W={window}, threshold={factor:g} x pre-share; pre-share < 0.05 is NA. "
             "Censored values display as > n and enter paired tests as n_post + 1.",
             "Holm family: every treatment x baseline pair for one metric in this directory; "
             "gamma/lambda sweep arms enlarge the family.", ""]
    if drops:
        lines += ["## Dropped blocks", ""] + [f"- {x}" for x in drops] + [""]
    if notes:
        lines += ["## Availability notes", ""] + [f"- {x}" for x in notes] + [""]
    metrics = [c for c in cells if c in METRICS or c.startswith(("build_", "client_", "drift_phase_"))]
    for metric in metrics:
        lines += [f"## {metric}", "", "| Arm | Median | n |", "|---|---:|---:|"]
        for arm, group in cells.groupby("arm", sort=True):
            values = pd.to_numeric(group[metric], errors="coerce").dropna()
            lines.append(f"| {arm} | {_fmt(values.median()) if not values.empty else 'NA'} | {len(values)} |")
        lines += ["", "| Treatment | Baseline | n | Median diff [95% CI] | p_raw | p_holm | r_rb | n_zero |",
                  "|---|---|---:|---:|---:|---:|---:|---:|"]
        subset = pairs[pairs["metric"] == metric] if not pairs.empty else []
        for _, pair in subset.iterrows():
            delta = f"{_fmt(pair['median_diff'])} [{_fmt(pair['ci_lo'])}, {_fmt(pair['ci_hi'])}]"
            lines.append(f"| {pair['treatment']} | {pair['baseline']} | {int(pair['n'])} | {delta} | "
                         f"{_fmt(pair['p_raw'])} | {_fmt(pair['p_holm'])} | {_fmt(pair['r_rb'])} | {int(pair['n_zero'])} |")
        lines.append("")
    if not cells.empty:
        lines += ["## Last finisher", "", "| Round | Arm | Worker address | CPU millicores |",
                  "|---:|---|---|---:|"]
        for _, row in cells.sort_values(["round", "arm"]).iterrows():
            lines.append(f"| {int(row['round'])} | {row['arm']} | {row['last_finisher_address']} | "
                         f"{_fmt(row['last_finisher_cpu_millis'])} |")
        lines.append("")
    path.write_text("\n".join(lines), encoding="utf-8")


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("out_dir", type=Path)
    parser.add_argument("--out", type=Path, help="output directory (default OUT_DIR/analysis)")
    parser.add_argument("--treatments", nargs="+", help="treatment arm labels")
    parser.add_argument("--baselines", nargs="+", help="baseline arm labels")
    parser.add_argument("--window", type=int, default=20)
    parser.add_argument("--factor", type=float, default=0.5)
    parser.add_argument("--strict", action="store_true")
    args = parser.parse_args(argv)
    if args.window < 1 or not 0 <= args.factor <= 1:
        parser.error("--window must be positive and --factor between 0 and 1")
    cells, dropped, notes = load_cells(args.out_dir, args.window, args.factor)
    for reason in dropped:
        print("DROPPED:", reason)
    for note in notes:
        print("NOTE:", note)
    pairs = compare(cells, args.treatments, args.baselines) if not cells.empty else pd.DataFrame()
    output = args.out or args.out_dir / "analysis"
    output.mkdir(parents=True, exist_ok=True)
    cells.to_csv(output / "cells.csv", index=False, na_rep="NA")
    pairs.to_csv(output / "pairs.csv", index=False, na_rep="NA")
    write_summary(output / "summary.md", cells, pairs, dropped, notes, args.window, args.factor)
    print(f"Wrote {output} ({cells['round'].nunique() if not cells.empty else 0} complete blocks)")
    return 2 if args.strict and dropped else 0


if __name__ == "__main__":
    raise SystemExit(main())
