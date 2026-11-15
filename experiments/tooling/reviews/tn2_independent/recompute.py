"""Independent, read-only recomputation of Experiment 2 confirmatory tests.

Reads raw campaign CSV/JSONL only. Writes all artifacts beside this script.
"""

import csv
import json
from pathlib import Path

import numpy as np
import pandas as pd
from scipy.stats import rankdata, wilcoxon


DATA = Path(r"D:\bench-work\out\official")
OUT = Path(__file__).resolve().parent
SCENARIOS = ("E2-perm", "E2-trans", "E2-onoff", "E2-lambda")
MAIN = SCENARIOS[:3]
TARGET = "worker-5:50051"
HG = "hybrid-linucb"
D_ARMS = (
    "hybrid-linucb-d-g095",
    "hybrid-linucb-d-g098",
    "hybrid-linucb-d-g099",
    "hybrid-linucb-d-arm-g098",
)
BASELINES = ("leastloaded", "sed")


def parse_tasks(path):
    tasks, events = [], []
    with path.open(encoding="utf-8") as handle:
        for line in handle:
            record = json.loads(line)
            if record.get("event") == "task_completed":
                tasks.append((int(record["dispatch_seq"]), record["worker_address"],
                              record["success"] is True))
            elif record.get("event") == "injected_event":
                events.append((record["kind"], int(record["dispatch_count"]),
                               record.get("detail", "")))
    tasks.sort(key=lambda t: t[0])
    return tasks, events


def parse_event_csv(path):
    with path.open(newline="", encoding="utf-8") as handle:
        return [(row["kind"], int(row["dispatch_count"]), row["detail"])
                for row in csv.DictReader(handle)]


def share(tasks):
    return sum(t[1] == TARGET for t in tasks) / len(tasks) if tasks else np.nan


def adaptation_rank(post, reference):
    if reference < 0.05:
        return np.nan, False
    chosen = np.fromiter((t[1] == TARGET for t in post), dtype=int)
    for c in range(20, len(chosen) + 1):
        if chosen[c-20:c].sum() / 20 <= 0.5 * reference:
            return c, False
    return len(chosen) + 1, True


def paired_test(frame, metric, treatment, baseline):
    wide = frame.pivot(index="round", columns="arm", values=metric)
    paired = wide[[treatment, baseline]].dropna()
    differences = paired[treatment].to_numpy() - paired[baseline].to_numpy()
    nonzero = differences[differences != 0]
    if not len(nonzero):
        p_raw, r_rb = 1.0, 0.0
    else:
        ranks = rankdata(abs(nonzero))
        plus = ranks[nonzero > 0].sum()
        minus = ranks[nonzero < 0].sum()
        r_rb = float((plus - minus) / (plus + minus))
        no_ties = len(np.unique(abs(nonzero))) == len(nonzero)
        method = "exact" if len(nonzero) <= 25 and no_ties else "approx"
        p_raw = float(wilcoxon(nonzero, zero_method="wilcox",
                               alternative="two-sided", method=method).pvalue)
    return {
        "n": len(paired),
        "treatment_median": float(paired[treatment].median()),
        "baseline_median": float(paired[baseline].median()),
        "median_diff": float(np.median(differences)),
        "p_raw": p_raw,
        "r_rb": r_rb,
    }


def holm(rows):
    # Holm step-down adjusted p values, separately for each scenario/family/metric.
    order = sorted(range(len(rows)), key=lambda i: rows[i]["p_raw"])
    running = 0.0
    for position, i in enumerate(order):
        running = max(running, min(1.0, (len(rows) - position) * rows[i]["p_raw"]))
        rows[i]["p_holm"] = running
    return rows


def main():
    all_cells, issues = [], []
    for scenario in SCENARIOS:
        source = DATA / scenario
        builds = pd.read_csv(source / "results.csv")
        builds["round"] = builds["round"].astype(int)
        builds["build_idx"] = builds["build_idx"].astype(int)
        for (arm, round_no), group in builds.groupby(["arm", "round"], sort=True):
            cell = {"scenario": scenario, "arm": arm, "round": int(round_no)}
            if len(group) != 2 or set(group.build_idx) != {1, 2}:
                issues.append(f"{scenario}/{arm}/r{round_no}: expected builds 1 and 2; got {group.build_idx.tolist()}")
            cell["makespan"] = float(group.cell_makespan_s.sum())
            stem = f"{arm}_round{round_no}"
            tasks, events = parse_tasks(source / f"tasks_{stem}.jsonl")
            csv_events = parse_event_csv(source / f"events_{stem}.csv")
            if events != csv_events:
                issues.append(f"{scenario}/{arm}/r{round_no}: JSONL and CSV events differ")
            sequences = [task[0] for task in tasks]
            if sequences != list(range(1, len(tasks) + 1)):
                issues.append(f"{scenario}/{arm}/r{round_no}: dispatch sequences are not contiguous 1..N")
            cell["successful_tasks"] = sum(task[2] for task in tasks)
            if cell["successful_tasks"] != 590 or len(tasks) != 590:
                issues.append(f"{scenario}/{arm}/r{round_no}: {cell['successful_tasks']} successful of {len(tasks)} completed tasks")
            on = [e for e in events if e[0] == "drift_on"]
            off = [e for e in events if e[0] == "drift_off" and not e[2].startswith("cell_end")]
            if not on:
                issues.append(f"{scenario}/{arm}/r{round_no}: missing drift_on")
                cell["first_drift_on_dispatch_count"] = np.nan
                cell["drift_pre_all_share"] = np.nan
                cell["drift_post_share"] = np.nan
                cell["time_to_adapt"] = np.nan
                cell["tta_censored"] = False
            else:
                onset = on[0][1] + 1
                pre = [task for task in tasks if 1 <= task[0] < onset]
                post = [task for task in tasks if task[0] >= onset]
                cell["first_drift_on_dispatch_count"] = on[0][1]
                cell["drift_pre_all_share"] = share(pre)
                cell["drift_post_share"] = share(post)
                cell["time_to_adapt"], cell["tta_censored"] = adaptation_rank(
                    post, cell["drift_pre_all_share"])
            cell["target_after_recovery_share"] = np.nan
            if scenario == "E2-trans":
                if len(on) != 1 or len(off) != 1:
                    issues.append(f"{scenario}/{arm}/r{round_no}: expected one drift_on and one non-cell_end drift_off; got {len(on)}, {len(off)}")
                else:
                    recovered = [task for task in tasks if task[0] >= off[0][1] + 1]
                    cell["target_after_recovery_share"] = share(recovered)
            all_cells.append(cell)

    cells = pd.DataFrame(all_cells)
    summary = (cells.groupby(["scenario", "arm"], sort=True)
               .agg(cells=("round", "size"),
                    cells_590_successful=("successful_tasks", lambda s: int((s == 590).sum())),
                    censored_time_to_adapt=("tta_censored", "sum"),
                    first_drift_on_min=("first_drift_on_dispatch_count", "min"),
                    first_drift_on_max=("first_drift_on_dispatch_count", "max"))
               .reset_index())
    summary["censored_time_to_adapt"] = summary["censored_time_to_adapt"].where(
        summary.scenario.eq("E2-perm"), pd.NA)

    results = []
    for scenario in MAIN:
        frame = cells[cells.scenario == scenario]
        actual_arms = set(frame.arm)
        expected = set(D_ARMS) | {HG, *BASELINES}
        if actual_arms != expected:
            issues.append(f"{scenario}: arms differ from expected: {sorted(actual_arms ^ expected)}")
        if any(set(frame[frame.arm == arm]["round"]) != set(range(1, 21)) for arm in expected):
            issues.append(f"{scenario}: incomplete 20-round arm coverage")
        metric_sets = {
            "D vs HG": ["makespan", "drift_post_share", "time_to_adapt"]
                       + (["target_after_recovery_share"] if scenario == "E2-trans" else []),
            "default": ["makespan"],
        }
        for family, metrics in metric_sets.items():
            for metric in metrics:
                comparisons = ([(arm, HG) for arm in D_ARMS] if family == "D vs HG"
                               else [(arm, base) for arm in (HG, *D_ARMS) for base in BASELINES])
                family_rows = []
                for treatment, baseline in comparisons:
                    row = {"scenario": scenario, "family": family, "metric": metric,
                           "treatment": treatment, "baseline": baseline}
                    row.update(paired_test(frame, metric, treatment, baseline))
                    family_rows.append(row)
                results.extend(holm(family_rows))

    fields = ["scenario", "family", "metric", "treatment", "baseline", "n",
              "treatment_median", "baseline_median", "median_diff", "p_raw", "p_holm", "r_rb"]
    with (OUT / "results.csv").open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=fields)
        writer.writeheader()
        writer.writerows(results)
    report = [
        "# Independent Experiment 2 recomputation",
        "",
        "Run: `D:\\bench-work\\venv\\Scripts\\python.exe recompute.py` from this directory.",
        "Inputs are read only. The script reads `results.csv`, task JSONL, and event CSV",
        "for all four E2 scenarios; it writes only `results.csv` and this README.",
        "No project analyzer code is imported or run.",
        "",
        "Cell makespan is the sum of the two build makespans. Dispatch metrics use",
        "completed tasks ordered by dispatch sequence and event positions from task JSONL.",
        "Time to adapt uses the first qualifying 20-task post-onset window; censored",
        "cells take rank value `len(post)+1`, and pre-share below 0.05 gives NA.",
        "Pairs are matched by round; differences are treatment minus baseline.",
        "Two-sided Wilcoxon uses `zero_method=wilcox`, exact with <=25 nonzero",
        "untied absolute differences, otherwise normal approximation, without",
        "continuity correction. Matched-pairs rank-biserial r uses signed rank sums.",
        "Holm adjustment is separate within each scenario, family, and metric.",
        "",
        "## Cell checks",
        "",
        "| Scenario | Arm | Cells | 590 successful | Censored TTA | First drift_on min-max |",
        "| --- | --- | ---: | ---: | ---: | ---: |",
    ]
    for row in summary.itertuples(index=False):
        censored = str(int(row.censored_time_to_adapt)) if row.scenario == "E2-perm" else "—"
        report.append(f"| {row.scenario} | {row.arm} | {row.cells} | "
                      f"{row.cells_590_successful} | {censored} | "
                      f"{int(row.first_drift_on_min)}–{int(row.first_drift_on_max)} |")
    report += ["", "## Data anomalies", ""]
    if issues:
        report += [f"- {issue}" for issue in issues]
        report += ["", "For these six E2-onoff cells, JSONL has an additional `drift_on`",
                   "at dispatch_count 590 after the final task; event CSV omits it.",
                   "The first onset and reported confirmatory metrics are unaffected."]
    else:
        report.append("None found.")
    (OUT / "README.md").write_text("\n".join(report) + "\n", encoding="utf-8")
    print(f"{len(cells)} cells; {len(results)} comparisons; {len(issues)} anomalies")
    for row in results:
        if ((row["scenario"], row["family"], row["metric"], row["treatment"], row["baseline"]) in {
            ("E2-perm", "D vs HG", "time_to_adapt", D_ARMS[1], HG),
            ("E2-perm", "D vs HG", "drift_post_share", D_ARMS[1], HG),
            ("E2-trans", "D vs HG", "target_after_recovery_share", D_ARMS[1], HG),
            ("E2-perm", "default", "makespan", D_ARMS[1], "leastloaded"),
            ("E2-onoff", "default", "makespan", D_ARMS[1], "leastloaded"),
        }):
            print(row)


if __name__ == "__main__":
    main()
