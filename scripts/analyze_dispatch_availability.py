#!/usr/bin/env python3
"""Reconstruct cluster state at every dispatch instant.

This is the script behind the mechanism table of the paper (Table
"worker availability at dispatch"). It answers one question per dispatch:
when the coordinator chose a worker, how many workers were sitting
completely idle -- and did it pass over one?

Method
------
The task log records, for every task, the worker that ran it, the completion
timestamp (``ts``) and the total task duration (``total_duration_ms``). A
worker is therefore busy over the half-open interval

    [ts - total_duration_ms, ts)

Intersecting those intervals with each dispatch timestamp gives the number of
workers idle at that instant. No extra instrumentation is needed; the fields
are already in the log.

Definitions
-----------
idle available
    Fraction of dispatches taken while at least one worker held zero tasks.
    This is a property of the *operating point*, not of the scheduler: below
    saturation it approaches 100% for every policy.

idle skip (absolute)
    Fraction of all dispatches sent to a busy worker while at least one
    worker was completely idle. This is the mistake a scheduler can still
    make when the cluster is under-subscribed.

idle skip (conditional)
    The same count, but as a fraction of only those dispatches that actually
    had an idle worker to skip. At saturation the two diverge: the absolute
    rate falls mostly because the *opportunity* to skip becomes rare, not
    because the policy got smarter. Reporting both separates the two effects.

LeastLoaded scores ~0 by construction -- its rule *is* "fewest active tasks"
-- so its value is a correctness check on the reconstruction, not a finding.

Self-validation
---------------
The coordinator independently records ``worker_active_tasks_at_dispatch``:
its own live count for the chosen worker at decision time. The script
recomputes that number offline and reports the agreement rate. Anything below
~99% means the reconstruction is not trustworthy and the rest of the output
should be discarded.

Warm-start split
----------------
With ``--warm-start N`` the first N dispatches of each round are reported
separately. Those are delegated to the least-loaded heuristic by design, so
splitting there shows whether a learner's low skip rate is its own doing or
just an artefact of the delegation window.

Usage
-----
    python analyze_dispatch_availability.py <OUT_DIR> [--warm-start 100]

    OUT_DIR holds tasks_<sched>_round<N>.jsonl files, as produced by
    scripts/benchmark_rigorous.sh.

Complexity is O(n^2) per round, which is immaterial at benchmark scale
(~300 tasks per round) and keeps the interval logic readable.
"""

import argparse
import collections
import glob
import json
import os
import sys
from datetime import datetime


def parse_ts(value):
    """Parse an RFC3339 timestamp with variable-width fractional seconds."""
    text = value.rstrip("Z")
    if "." in text:
        head, frac = text.split(".", 1)
        # datetime.fromisoformat accepts 3 or 6 fractional digits only.
        text = head + "." + (frac + "000000")[:6]
    return datetime.fromisoformat(text).timestamp()


def load_round(path):
    """Return [(start, end, worker, active_at_dispatch)] sorted by dispatch time."""
    tasks = []
    with open(path, encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            try:
                rec = json.loads(line)
            except json.JSONDecodeError:
                continue
            if rec.get("event") != "task_completed":
                continue
            if "ts" not in rec or "total_duration_ms" not in rec:
                continue
            end = parse_ts(rec["ts"])
            start = end - rec["total_duration_ms"] / 1000.0
            tasks.append(
                (start, end, rec["worker_id"], rec.get("worker_active_tasks_at_dispatch"))
            )
    tasks.sort(key=lambda t: t[0])
    return tasks


def analyze_round(tasks, warm_start):
    """Per-round counters. Returns a dict of running totals."""
    workers = {t[2] for t in tasks}
    acc = collections.Counter()
    for index, (start, _end, worker, active_truth) in enumerate(tasks):
        active = collections.Counter()
        for other_index, (o_start, o_end, o_worker, _) in enumerate(tasks):
            if other_index == index:
                continue
            if o_start <= start < o_end:
                active[o_worker] += 1

        idle = [w for w in workers if active[w] == 0]
        phase = "warm" if index < warm_start else "learned"

        acc["n"] += 1
        acc[f"n_{phase}"] += 1
        acc["idle_workers"] += len(idle)
        if idle:
            acc["has_idle"] += 1
            acc[f"has_idle_{phase}"] += 1
            if active[worker] > 0:
                acc["skip"] += 1
                acc[f"skip_{phase}"] += 1

        if active_truth is not None:
            acc["checked"] += 1
            if active[worker] == active_truth:
                acc["agree"] += 1
    return acc


def pct(part, whole):
    return 100.0 * part / whole if whole else 0.0


def main():
    parser = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    parser.add_argument("out_dir", help="benchmark output directory")
    parser.add_argument(
        "--warm-start",
        type=int,
        default=0,
        metavar="N",
        help="dispatches per round delegated to the heuristic (0 disables the split)",
    )
    args = parser.parse_args()

    paths = sorted(glob.glob(os.path.join(args.out_dir, "tasks_*_round*.jsonl")))
    if not paths:
        sys.exit(f"no tasks_*_round*.jsonl under {args.out_dir}")

    totals = collections.defaultdict(collections.Counter)
    rounds = collections.Counter()
    for path in paths:
        name = os.path.basename(path)[len("tasks_"):]
        scheduler = name.rsplit("_round", 1)[0]
        tasks = load_round(path)
        if not tasks:
            continue
        totals[scheduler].update(analyze_round(tasks, args.warm_start))
        rounds[scheduler] += 1

    if not totals:
        sys.exit("no usable task records found")

    print(f"\nDispatch-time worker availability -- {args.out_dir}")
    print(
        f"{'scheduler':16}{'rounds':>7}{'dispatches':>12}{'idle avail':>12}"
        f"{'idle/5':>9}{'skip abs':>10}{'skip cond':>11}{'recon ok':>10}"
    )
    order = sorted(totals, key=lambda s: -pct(totals[s]["skip"], totals[s]["n"]))
    for scheduler in order:
        acc = totals[scheduler]
        print(
            f"{scheduler:16}{rounds[scheduler]:7d}{acc['n']:12d}"
            f"{pct(acc['has_idle'], acc['n']):11.1f}%"
            f"{acc['idle_workers'] / acc['n']:9.2f}"
            f"{pct(acc['skip'], acc['n']):9.1f}%"
            f"{pct(acc['skip'], acc['has_idle']):10.1f}%"
            f"{pct(acc['agree'], acc['checked']):9.1f}%"
        )

    print(
        "\n'recon ok' is the agreement between this offline reconstruction and the"
        "\ncoordinator's own worker_active_tasks_at_dispatch counter. Below ~99%,"
        "\ndiscard the rest of this table."
    )

    if args.warm_start:
        print(
            f"\nSplit at the warm-start boundary (first {args.warm_start} dispatches"
            f" of each round):"
        )
        print(f"{'scheduler':16}{'warm skip':>12}{'learned skip':>15}")
        for scheduler in order:
            acc = totals[scheduler]
            print(
                f"{scheduler:16}"
                f"{pct(acc['skip_warm'], acc['n_warm']):11.2f}%"
                f"{pct(acc['skip_learned'], acc['n_learned']):14.2f}%"
            )
        print(
            "\nA warm-start rate of 0.00% is required, not informative: those"
            "\ndispatches are least-loaded by construction. The learned column is"
            "\nwhere a proposed mechanism has to earn its result."
        )


if __name__ == "__main__":
    main()
