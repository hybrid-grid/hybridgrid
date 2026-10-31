#!/usr/bin/env python3
"""Prespecified selection and confirmation analysis for the j x gamma campaign."""

import argparse
import csv
import json
import math
import re
import sys
from pathlib import Path

import numpy as np


X = 0.10
TMAX = 84.24
BOOT = 50000
BOOT_SEED = 20261026
BAND_Q = 0.90
WINDOW = 1.02
GATE_M = 0.05
GATE_T = 1.10
SIGN_ALPHA = 0.05

BASELINES = ("leastloaded", "sed")
COARSE = (0.80, 0.86, 0.92, 0.98)
GAMMA_MAP = {
    0.80: (0.80, 0.82, 0.84),
    0.86: (0.84, 0.86, 0.88),
    0.92: (0.90, 0.92, 0.94),
    0.98: (0.94, 0.96, 0.98),
}
ARM_RE = re.compile(r"hybrid-linucb-d-arm-g(0(?:8[02468]|9[02468]))\Z")
REQUIRED = {"arm", "round", "order_pos", "build_idx", "client", "elapsed_s",
            "cell_makespan_s", "start_ts", "end_ts"}


class AnalysisError(ValueError):
    """Invalid or incomplete campaign input."""


def arm_gamma(arm):
    match = ARM_RE.fullmatch(arm)
    if match is None:
        raise AnalysisError(f"unknown arm label: {arm}")
    return int(match.group(1)) / 100


def hg_arm(gamma):
    return f"hybrid-linucb-d-arm-g{round(gamma * 100):03d}"


def load_dir(path):
    """Return build-2 times keyed by round and arm, plus build-2 row counts."""
    path = Path(path)
    csv_path = path / "results.csv"
    if not csv_path.is_file():
        raise AnalysisError(f"missing file: {csv_path}")
    try:
        with csv_path.open(newline="", encoding="utf-8-sig") as handle:
            reader = csv.DictReader(handle)
            if reader.fieldnames is None:
                raise AnalysisError(f"empty results.csv: {csv_path}")
            missing = REQUIRED - set(reader.fieldnames)
            if missing:
                raise AnalysisError(f"missing columns in {csv_path}: {', '.join(sorted(missing))}")
            rounds = {}
            counts = {}
            for line, row in enumerate(reader, 2):
                arm = row["arm"]
                if arm not in BASELINES:
                    arm_gamma(arm)
                try:
                    rnd = int(row["round"])
                    build = int(row["build_idx"])
                except (TypeError, ValueError) as exc:
                    raise AnalysisError(f"invalid round/build_idx in {csv_path}:{line}") from exc
                if rnd < 1 or build not in (1, 2):
                    raise AnalysisError(f"invalid round/build_idx in {csv_path}:{line}")
                if build != 2:
                    continue
                try:
                    elapsed = float(row["elapsed_s"])
                except (TypeError, ValueError) as exc:
                    raise AnalysisError(f"invalid elapsed_s in {csv_path}:{line}") from exc
                if not math.isfinite(elapsed) or elapsed <= 0:
                    raise AnalysisError(f"invalid elapsed_s in {csv_path}:{line}")
                cells = rounds.setdefault(rnd, {})
                if arm in cells:
                    raise AnalysisError(f"duplicate build-2 row for {arm}, round {rnd} in {csv_path}")
                cells[arm] = elapsed
                counts[arm] = counts.get(arm, 0) + 1
    except OSError as exc:
        raise AnalysisError(f"failed to read {csv_path}: {exc}") from exc
    return rounds, counts


def require_rounds(path, arms, expected):
    rounds, counts = load_dir(path)
    if set(counts) != set(arms):
        raise AnalysisError(
            f"wrong arm count/labels in {path}: expected {sorted(arms)}, got {sorted(counts)}"
        )
    wanted = set(range(1, expected + 1))
    if set(rounds) != wanted:
        raise AnalysisError(
            f"wrong round count in {path}: expected rounds 1..{expected}, "
            f"got {sorted(rounds)}"
        )
    for rnd in range(1, expected + 1):
        missing = set(arms) - set(rounds[rnd])
        if missing:
            raise AnalysisError(f"incomplete round {rnd} in {path}: missing {sorted(missing)}")
    return rounds


def expected_dir_names(paths, prefix, js=None):
    names = [Path(path).name for path in paths]
    if len(names) != len(set(names)):
        raise AnalysisError(f"duplicate {prefix} directories")
    pattern = re.compile(rf"{re.escape(prefix)}-j([1-9][0-9]*)\Z")
    found = []
    for name in names:
        match = pattern.fullmatch(name)
        if match is None:
            raise AnalysisError(f"wrong directory name for {prefix}: {name}")
        found.append(int(match.group(1)))
    if js is not None and set(found) != set(js):
        raise AnalysisError(f"wrong directory count/levels for {prefix}: expected {list(js)}, got {found}")
    if len(found) != len(set(found)):
        raise AnalysisError(f"duplicate j level for {prefix}: {found}")
    return dict(zip(found, paths))


def estimate(rounds, gammas, n):
    """Build paired log ratios and full-precision point estimates."""
    arms = [hg_arm(gamma) for gamma in gammas]
    h = np.asarray([[rounds[r][arm] for arm in arms] for r in range(1, n + 1)], dtype=float)
    b = np.asarray([[rounds[r][arm] for arm in BASELINES] for r in range(1, n + 1)], dtype=float)
    z = np.log(b[:, None, :] / h[:, :, None])
    theta = np.median(z, axis=0)
    m_cb = -np.expm1(-theta)
    m = np.min(m_cb, axis=1)
    t = np.median(h, axis=0)
    return h, z, theta, m_cb, m, t


def distance(m, t):
    return max(0.0, X - m) / X + max(0.0, math.log(t / TMAX)) / math.log(1.2)


def cell(j, gamma, m_cb, m, t, d=None):
    item = {"j": j, "gamma": gamma, "m_cb": dict(zip(BASELINES, map(float, m_cb))),
            "m_c": float(m), "T_c": float(t)}
    if d is not None:
        item["D"] = float(d)
    return item


def phase1(paths):
    by_j = expected_dir_names(paths, "p1", (4, 5, 6))
    table = []
    winners = []
    arms = (*BASELINES, *(hg_arm(g) for g in COARSE))
    for j in sorted(by_j):
        rounds = require_rounds(by_j[j], arms, 6)
        _, _, _, margins, m, t = estimate(rounds, COARSE, 6)
        entries = [cell(j, gamma, margins[i], m[i], t[i], distance(m[i], t[i]))
                   for i, gamma in enumerate(COARSE)]
        table.extend(entries)
        winners.append(min(entries, key=lambda c: (c["D"], c["T_c"], -c["m_c"], c["gamma"])))
    kept = sorted(winners, key=lambda c: (c["D"], c["T_c"], -c["m_c"], c["j"]))[:2]
    js = [entry["j"] for entry in kept]
    return {"js": js, "gammas": {str(entry["j"]): list(GAMMA_MAP[entry["gamma"]])
                                   for entry in kept}, "phase": 1, "table": table}


def phase2(paths):
    if len(paths) != 2:
        raise AnalysisError(f"wrong directory count for phase2: expected 2, got {len(paths)}")
    by_j = expected_dir_names(paths, "p2")
    h_parts, z_parts, items = [], [], []
    for j in sorted(by_j):
        rounds, counts = load_dir(by_j[j])
        gammas = sorted(arm_gamma(arm) for arm in counts if arm not in BASELINES)
        if len(gammas) != 3:
            raise AnalysisError(f"wrong arm count in {by_j[j]}: expected 3 gamma arms, got {len(gammas)}")
        # A phase-2 directory is the chosen three-point neighbourhood of one coarse gamma.
        if tuple(gammas) not in GAMMA_MAP.values():
            raise AnalysisError(f"wrong gamma set in {by_j[j]}: {gammas}")
        expected = (*BASELINES, *(hg_arm(g) for g in gammas))
        rounds = require_rounds(by_j[j], expected, 10)
        h, z, _, _, _, _ = estimate(rounds, gammas, 10)
        h_parts.append(h)
        z_parts.append(z)
        items.extend((j, gamma) for gamma in gammas)
    h = np.concatenate(h_parts, axis=1)
    z = np.concatenate(z_parts, axis=1)
    theta = np.median(z, axis=0)
    t = np.median(h, axis=0)
    rng = np.random.default_rng(BOOT_SEED)
    indices = rng.integers(0, 10, size=(BOOT, 10))
    theta_star = np.median(z[indices], axis=1)
    log_t_star = np.log(np.median(h[indices], axis=1))
    error = np.maximum(np.max(theta_star - theta, axis=(1, 2)),
                       np.max(np.log(t) - log_t_star, axis=1))
    q = max(0.0, float(np.quantile(error, BAND_Q, method="linear")))
    m_cb = -np.expm1(-theta)
    m = np.min(m_cb, axis=1)
    lower_cb = -np.expm1(-(theta - q))
    lower = np.min(lower_cb, axis=1)
    upper = np.exp(np.log(t) + q)
    table = []
    for i, (j, gamma) in enumerate(items):
        entry = cell(j, gamma, m_cb[i], m[i], t[i])
        entry.update({"L_c": float(lower[i]), "U_c": float(upper[i]),
                      "D": distance(float(lower[i]), float(upper[i]))})
        table.append(entry)
    feasible = [entry for entry in table if entry["L_c"] >= X and entry["U_c"] <= TMAX]
    if feasible:
        t_min = min(entry["T_c"] for entry in feasible)
        window = [entry for entry in feasible if entry["T_c"] <= WINDOW * t_min]
        chosen = min(window, key=lambda c: (-c["L_c"], c["T_c"], c["j"], c["gamma"]))
        branch = "feasible"
        proceed = True
    else:
        # Nearest ties resolve by D, then smaller T_c, then ascending j and gamma.
        chosen = min(table, key=lambda c: (c["D"], c["T_c"], c["j"], c["gamma"]))
        branch = "nearest"
        proceed = chosen["m_c"] >= GATE_M and chosen["T_c"] <= GATE_T * TMAX
    return {"j": chosen["j"], "gamma": chosen["gamma"], "branch": branch,
            "continue": bool(proceed), "q": q, "selected": chosen, "table": table}


def sign_pvalue(k, n=24):
    return sum(math.comb(n, i) for i in range(k, n + 1)) / (2 ** n)


def confirmation_scenario(path, gamma):
    primary = hg_arm(gamma)
    secondary = hg_arm(0.98)
    arms = tuple(dict.fromkeys((primary, *BASELINES, secondary)))
    rounds = require_rounds(path, arms, 24)
    h = np.asarray([rounds[r][primary] for r in range(1, 25)])
    conditions = {}
    for baseline in BASELINES:
        b = np.asarray([rounds[r][baseline] for r in range(1, 25)])
        # Algebraically equivalent to z > -log(1-X); multiplication keeps
        # threshold ties from changing sides through log roundoff.
        exceeds = b * (1 - X) > h
        k = int(np.count_nonzero(exceeds))
        p = sign_pvalue(k)
        conditions[baseline] = {"exceedances": k, "p_value": p,
                                "supported": p <= SIGN_ALPHA}
    k = int(np.count_nonzero(h < TMAX))
    p = sign_pvalue(k)
    conditions["time"] = {"exceedances": k, "p_value": p, "supported": p <= SIGN_ALPHA}
    medians = {arm: float(np.median([rounds[r][arm] for r in range(1, 25)])) for arm in arms}
    margins = {}
    for arm in (*BASELINES, secondary):
        if arm == primary:
            margins[arm] = 0.0
        else:
            b = np.asarray([rounds[r][arm] for r in range(1, 25)])
            margins[arm] = float(-np.expm1(-np.median(np.log(b / h))))
    return {"conditions": conditions, "medians_s": medians, "m": margins}


def confirm(c1, c2, final_path):
    path = Path(final_path)
    if not path.is_file():
        raise AnalysisError(f"missing file: {path}")
    try:
        plan = json.loads(path.read_text(encoding="utf-8"))
        gamma = float(plan["gamma"])
        j = int(plan["j"])
    except (OSError, ValueError, KeyError, TypeError) as exc:
        raise AnalysisError(f"invalid final file: {path}: {exc}") from exc
    if j < 1 or gamma not in [i / 100 for i in range(80, 99, 2)]:
        raise AnalysisError(f"invalid j/gamma in final file: {path}")
    scenarios = {"c1": confirmation_scenario(c1, gamma),
                 "c2": confirmation_scenario(c2, gamma)}
    claim = all(condition["supported"] for scenario in scenarios.values()
                for condition in scenario["conditions"].values())
    return {"j": j, "gamma": gamma, "claim": claim, "scenarios": scenarios}


def integrity(paths):
    """Report only opaque structural counts; never inspect timing columns."""
    output = []
    for index, path in enumerate(paths, 1):
        directory = Path(path)
        csv_path = directory / "results.csv"
        if not csv_path.is_file():
            raise AnalysisError(f"missing results.csv for dir#{index}")

        name = directory.name
        if re.fullmatch(r"p[15]-j[1-9][0-9]*", name):
            expected, min_arms = 6, 6
        elif re.fullmatch(r"p2-j[1-9][0-9]*", name):
            expected, min_arms = 10, 5
        elif name in ("c1", "c2"):
            expected, min_arms = 24, 3
        else:
            expected, min_arms = 0, 0

        rows = 0
        observed_arms = set()
        build2 = {}
        try:
            with csv_path.open(newline="", encoding="utf-8-sig") as handle:
                reader = csv.reader(handle)
                header = next(reader, [])
                columns = {column: header.index(column) for column in ("arm", "round", "build_idx")
                           if column in header}
                for row in reader:
                    rows += 1
                    if len(columns) != 3 or len(row) <= max(columns.values()):
                        continue
                    arm = row[columns["arm"]]
                    if arm:
                        observed_arms.add(arm)
                    try:
                        rnd = int(row[columns["round"]])
                        build = int(row[columns["build_idx"]])
                    except ValueError:
                        continue
                    if rnd > 0 and build == 2 and arm:
                        build2.setdefault(rnd, set()).add(arm)
        except (OSError, UnicodeError, csv.Error) as exc:
            raise AnalysisError(f"unreadable results.csv for dir#{index}: {exc.__class__.__name__}") from exc

        arms = observed_arms
        meta = directory / "meta.json"
        if meta.is_file():
            try:
                specified = json.loads(meta.read_text(encoding="utf-8")).get("ARMS")
            except (OSError, UnicodeError, ValueError):
                specified = None
            if isinstance(specified, str) and specified.strip():
                arms = set(specified.split())
        elif re.fullmatch(r"p[15]-j[1-9][0-9]*", name):
            arms = set((*BASELINES, *(hg_arm(g) for g in COARSE)))

        if expected == 0:
            expected = max(build2, default=0)
        complete = sum(len(arms) >= min_arms and arms <= build2.get(rnd, set())
                       for rnd in range(1, expected + 1))
        output.append(f"dir#{index} arms={len(arms)} rounds_present={len(build2)} "
                      f"rounds_complete={complete} rounds_missing={expected - complete} rows={rows}")
    output.append(f"total_dirs={len(paths)}")
    return "\n".join(output) + "\n"


def rounded(value):
    if isinstance(value, dict):
        return {key: rounded(item) for key, item in value.items()}
    if isinstance(value, (list, tuple)):
        return [rounded(item) for item in value]
    if isinstance(value, (float, np.floating)):
        return round(float(value), 6)
    return value


def emit(data, out=None):
    rendered = json.dumps(rounded(data), sort_keys=True, ensure_ascii=False,
                          allow_nan=False, separators=(",", ":")) + "\n"
    if out is None:
        sys.stdout.write(rendered)
    else:
        try:
            Path(out).parent.mkdir(parents=True, exist_ok=True)
            Path(out).write_text(rendered, encoding="utf-8", newline="\n")
        except OSError as exc:
            raise AnalysisError(f"failed to write {out}: {exc}") from exc


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    integrity_cmd = sub.add_parser("integrity", help="print counts only")
    integrity_cmd.add_argument("--dirs", nargs="+", required=True)
    phase1_cmd = sub.add_parser("phase1", help="select j levels and gamma neighbourhoods")
    phase1_cmd.add_argument("--dirs", nargs="+", required=True)
    phase1_cmd.add_argument("--out", required=True)
    phase2_cmd = sub.add_parser("phase2", help="select final configuration")
    phase2_cmd.add_argument("--dirs", nargs="+", required=True)
    phase2_cmd.add_argument("--out", required=True)
    confirm_cmd = sub.add_parser("confirm", help="run confirmation sign tests")
    confirm_cmd.add_argument("--c1", required=True)
    confirm_cmd.add_argument("--c2", required=True)
    confirm_cmd.add_argument("--final", required=True)
    confirm_cmd.add_argument("--out", required=True)
    args = parser.parse_args(argv)
    try:
        if args.command == "integrity":
            sys.stdout.write(integrity(args.dirs))
        elif args.command == "phase1":
            emit(phase1(args.dirs), args.out)
        elif args.command == "phase2":
            emit(phase2(args.dirs), args.out)
        else:
            emit(confirm(args.c1, args.c2, args.final), args.out)
    except AnalysisError as exc:
        parser.error(str(exc))


if __name__ == "__main__":
    main()
