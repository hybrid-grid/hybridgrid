#!/usr/bin/env python3
"""Validate container HGTIME durations against coordinator task timestamps."""

import csv
import json
import re
import sys
from datetime import datetime
from pathlib import Path


def _timestamp(value):
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


def check_task_spans(tasks, rows, arm, round_number):
    """Return reasons for client durations outside their task receipt/completion span."""
    by_build = {}
    for task in tasks:
        if task.get("event") != "task_completed":
            continue
        build_id = task.get("build_id")
        if build_id:
            by_build.setdefault(build_id, []).append(task)
    reasons = []
    for row in rows:
        build_idx, client = row["build_idx"], row["client"]
        build_id = f"{arm}-r{round_number}-b{build_idx}-c{client}"
        selected = by_build.get(build_id, [])
        if not selected:
            reasons.append(f"build {build_idx} client {client}: no tasks with build_id {build_id}")
            continue
        try:
            span = (max(_timestamp(task["ts"]) for task in selected) -
                    min(_timestamp(task["received_ts"]) for task in selected)).total_seconds()
            elapsed = float(row["elapsed_s"])
        except (KeyError, TypeError, ValueError) as error:
            reasons.append(f"build {build_idx} client {client}: invalid task timestamps or elapsed_s ({error})")
            continue
        if elapsed < span - 0.5 or elapsed > span + 20:
            reasons.append(f"build {build_idx} client {client}: elapsed_s={elapsed:.3f}, "
                           f"task_span_s={span:.3f}; expected [{span - 0.5:.3f}, "
                           f"{span + 20:.3f}]")
    return reasons


def check_task_cardinality(tasks, expected_per_build, session_builds):
    """Return reasons for incomplete builds or noncontinuous cell dispatch sequence."""
    completed = [task for task in tasks if task.get("event") == "task_completed"]
    counts = {index: 0 for index in range(1, session_builds + 1)}
    failures = 0
    ungrouped = 0
    sequences = []
    for task in completed:
        match = re.search(r"-b([0-9]+)-", str(task.get("build_id", "")))
        if match is None or int(match.group(1)) not in counts:
            ungrouped += 1
        elif task.get("success") is True:
            counts[int(match.group(1))] += 1
        if task.get("success") is not True:
            failures += 1
        sequences.append(task.get("dispatch_seq"))
    reasons = []
    for index, count in counts.items():
        if count != expected_per_build:
            reasons.append(f"build {index}: {count} successful tasks; expected {expected_per_build}")
    if failures:
        reasons.append(f"{failures} failed tasks; expected zero")
    if ungrouped:
        reasons.append(f"{ungrouped} tasks have missing or out-of-range build_id")
    total = expected_per_build * session_builds
    if (len(sequences) != total or
            any(type(sequence) is not int for sequence in sequences) or
            set(sequences) != set(range(1, total + 1)) or
            len(sequences) != len(set(sequences))):
        reasons.append(f"dispatch_seq must be unique and exactly 1..{total} "
                       f"({len(sequences)} task_completed events found)")
    return reasons


def main(argv=None):
    args = sys.argv[1:] if argv is None else argv
    if args and args[0] == "cardinality":
        if len(args) != 4:
            raise SystemExit("usage: exp_checks.py cardinality TASKS_JSONL EXPECT_TASKS_PER_BUILD SESSION_BUILDS")
        _, tasks_path, expected, builds = args
        with Path(tasks_path).open(encoding="utf-8") as handle:
            tasks = [json.loads(line) for line in handle if line.strip()]
        reasons = check_task_cardinality(tasks, int(expected), int(builds))
        if reasons:
            print("; ".join(reasons))
            return 1
        return 0
    if len(args) != 4:
        raise SystemExit("usage: exp_checks.py TASKS_JSONL RESULTS_CSV ARM ROUND")
    tasks_path, results_path, arm, round_number = args
    with Path(tasks_path).open(encoding="utf-8") as handle:
        tasks = [json.loads(line) for line in handle if line.strip()]
    with Path(results_path).open(newline="", encoding="utf-8") as handle:
        rows = list(csv.DictReader(handle, fieldnames=["arm", "round", "order_pos",
                                                      "build_idx", "client", "elapsed_s",
                                                      "cell_makespan_s"]))
    reasons = check_task_spans(tasks, rows, arm, round_number)
    if reasons:
        print("; ".join(reasons))
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
