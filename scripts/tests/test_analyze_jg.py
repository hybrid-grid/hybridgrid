"""Known-answer fixtures for the frozen j x gamma decision procedure."""

import contextlib
import csv
import io
import json
import subprocess
import sys
import unittest
from pathlib import Path
from unittest import mock

from scripts import analyze_jg as jg


ROOT = Path(__file__).resolve().parents[2]
FIELDS = ["arm", "round", "order_pos", "build_idx", "client", "elapsed_s",
          "cell_makespan_s", "start_ts", "end_ts"]


FILES = {}


def write_dir(root, name, rounds, arms, values, omit=None):
    directory = root / name
    with io.StringIO() as handle:
        writer = csv.DictWriter(handle, FIELDS)
        writer.writeheader()
        for rnd in range(1, rounds + 1):
            for pos, arm in enumerate(arms, 1):
                if omit == (rnd, arm):
                    continue
                value = values(arm, rnd) if callable(values) else values[arm]
                for build in (1, 2):
                    writer.writerow({"arm": arm, "round": rnd, "order_pos": pos,
                                     "build_idx": build, "client": "builder",
                                     "elapsed_s": value if build == 2 else 999.123456,
                                     "cell_makespan_s": value if build == 2 else 999.123456,
                                     "start_ts": "2026-10-26T00:00:00Z",
                                     "end_ts": "2026-10-26T00:02:00Z"})
        FILES[str(directory / "results.csv")] = handle.getvalue()
    return directory


def p1_arms():
    return [*jg.BASELINES, *(jg.hg_arm(g) for g in jg.COARSE)]


def p2_arms(gammas=(0.80, 0.82, 0.84)):
    return [*jg.BASELINES, *(jg.hg_arm(g) for g in gammas)]


class AnalyzeJGTests(unittest.TestCase):
    def setUp(self):
        FILES.clear()
        self.root = ROOT / "virtual-jg-fixtures"
        original_open = Path.open
        original_is_file = Path.is_file
        original_read_text = Path.read_text
        original_write_text = Path.write_text

        def is_file(path):
            return str(path) in FILES or original_is_file(path)

        def open_file(path, *args, **kwargs):
            if str(path) in FILES:
                return io.StringIO(FILES[str(path)])
            return original_open(path, *args, **kwargs)

        def read_text(path, *args, **kwargs):
            if str(path) in FILES:
                return FILES[str(path)]
            return original_read_text(path, *args, **kwargs)

        def write_text(path, data, *args, **kwargs):
            FILES[str(path)] = data
            return len(data)

        patches = [mock.patch.object(Path, "is_file", is_file),
                   mock.patch.object(Path, "open", open_file),
                   mock.patch.object(Path, "read_text", read_text),
                   mock.patch.object(Path, "write_text", write_text),
                   mock.patch.object(Path, "mkdir", lambda *args, **kwargs: None)]
        for patcher in patches:
            patcher.start()
            self.addCleanup(patcher.stop)

    def test_frozen_constants_and_gamma_map(self):
        self.assertEqual((jg.X, jg.TMAX, jg.BOOT, jg.BOOT_SEED, jg.BAND_Q),
                         (0.10, 84.24, 50000, 20261026, 0.90))
        self.assertEqual((jg.WINDOW, jg.GATE_M, jg.GATE_T, jg.SIGN_ALPHA),
                         (1.02, 0.05, 1.10, 0.05))
        self.assertEqual(jg.GAMMA_MAP, {
            0.80: (0.80, 0.82, 0.84), 0.86: (0.84, 0.86, 0.88),
            0.92: (0.90, 0.92, 0.94), 0.98: (0.94, 0.96, 0.98),
        })

    def test_phase1_selection_all_tie_breaks_and_map(self):
        # All cells are feasible (D=0). j4 resolves by T, j5 by m,
        # j6 by gamma; the j-level tie is resolved by T then m then j.
        configs = {4: {0.80: 70, 0.86: 72, 0.92: 74, 0.98: 76},
                   5: {0.80: 71, 0.86: 71, 0.92: 75, 0.98: 76},
                   6: {g: 71 for g in jg.COARSE}}
        paths = []
        for j, cells in configs.items():
            def value(arm, rnd, cells=cells, j=j):
                if arm in jg.BASELINES:
                    return 100 if j != 5 or rnd <= 3 else 120
                if j == 5 and jg.arm_gamma(arm) == 0.86:
                    return 76 if rnd <= 3 else 66
                return cells[jg.arm_gamma(arm)]
            paths.append(write_dir(self.root, f"p1-j{j}", 6, p1_arms(), value))
        answer = jg.phase1(paths)
        self.assertEqual(answer["js"], [4, 5])
        self.assertEqual(answer["gammas"], {"4": [0.80, 0.82, 0.84],
                                              "5": [0.84, 0.86, 0.88]})
        self.assertEqual(len(answer["table"]), 12)
        self.assertAlmostEqual(answer["table"][0]["m_c"], 0.30)
        self.assertEqual(answer["table"][0]["D"], 0)
        # Equal D and T across j5/j6: higher margin wins (j6), then lower j.
        self.assertEqual(min(answer["table"][8:],
                             key=lambda c: (c["D"], c["T_c"], -c["m_c"], c["gamma"]))["gamma"],
                         0.80)

    def test_phase1_distance_and_j_ties(self):
        paths = []
        for j in (4, 5, 6):
            t = {4: 80, 5: 80, 6: 90}[j]
            margins = {4: 0.15, 5: 0.15, 6: 0.20}[j]
            values = {arm: t / (1 - margins) for arm in jg.BASELINES}
            values.update({jg.hg_arm(g): t for g in jg.COARSE})
            paths.append(write_dir(self.root, f"p1-j{j}", 6, p1_arms(), values))
        answer = jg.phase1(paths)
        self.assertEqual(answer["js"], [4, 5])
        self.assertEqual(answer["gammas"]["4"], [0.80, 0.82, 0.84])
        self.assertGreater(answer["table"][-1]["D"], 0)

    def make_phase2(self, cells, baseline=100):
        paths = []
        for j in (4, 5):
            values = {arm: baseline for arm in jg.BASELINES}
            values.update({jg.hg_arm(g): cells[(j, g)] for g in (0.80, 0.82, 0.84)})
            paths.append(write_dir(self.root, f"p2-j{j}", 10, p2_arms(), values))
        return paths

    def test_phase2_feasible_window_inside_and_outside(self):
        for second_time, expected_j in ((71.3, 5), (71.5, 4)):
            with self.subTest(second_time=second_time):
                old_root = self.root
                self.root = old_root / f"window-{second_time}"
                try:
                    cells = {(j, g): 80 for j in (4, 5) for g in (0.80, 0.82, 0.84)}
                    cells[(4, 0.80)] = 70
                    cells[(5, 0.80)] = second_time
                    paths = []
                    for j, baseline in ((4, 100), (5, 120)):
                        values = {arm: baseline for arm in jg.BASELINES}
                        values.update({jg.hg_arm(g): cells[(j, g)]
                                       for g in (0.80, 0.82, 0.84)})
                        paths.append(write_dir(self.root, f"p2-j{j}", 10,
                                               p2_arms(), values))
                    result = jg.phase2(paths)
                    self.assertEqual(result["branch"], "feasible")
                    self.assertTrue(result["continue"])
                    self.assertEqual(result["q"], 0)
                    self.assertEqual(result["j"], expected_j)
                    self.assertEqual(result["gamma"], 0.80)
                finally:
                    self.root = old_root

    def test_phase2_feasible_higher_lower_bound_wins_window(self):
        # j4 has a faster point time. j5's baseline is larger, so its L is
        # larger despite its slightly slower time.
        cells = {(j, g): 82 for j in (4, 5) for g in (0.80, 0.82, 0.84)}
        cells[(4, 0.80)] = 70
        cells[(5, 0.80)] = 71.3
        paths = []
        for j, baseline in ((4, 100), (5, 120)):
            values = {arm: baseline for arm in jg.BASELINES}
            values.update({jg.hg_arm(g): cells[(j, g)] for g in (0.80, 0.82, 0.84)})
            paths.append(write_dir(self.root, f"p2-j{j}", 10, p2_arms(), values))
        result = jg.phase2(paths)
        self.assertEqual((result["j"], result["gamma"]), (5, 0.80))
        self.assertEqual(result["branch"], "feasible")

    def test_phase2_nearest_continue_true_and_false(self):
        for best_time, proceed in ((92, True), (97, False)):
            with self.subTest(best_time=best_time):
                old_root = self.root
                self.root = old_root / f"nearest-{best_time}"
                try:
                    cells = {(j, g): 130 for j in (4, 5)
                             for g in (0.80, 0.82, 0.84)}
                    cells[(4, 0.80)] = best_time
                    result = jg.phase2(self.make_phase2(cells, baseline=105))
                    self.assertEqual(result["branch"], "nearest")
                    self.assertEqual((result["j"], result["gamma"]), (4, 0.80))
                    self.assertIs(result["continue"], proceed)
                finally:
                    self.root = old_root

    def test_phase2_nearest_exact_distance_and_time_ties(self):
        cells = {(j, g): 130 for j in (4, 5) for g in (0.80, 0.82, 0.84)}
        cells[(4, 0.82)] = 92
        cells[(4, 0.84)] = 92
        cells[(5, 0.80)] = 92
        result = jg.phase2(self.make_phase2(cells, baseline=100))
        self.assertEqual(result["branch"], "nearest")
        tied = [entry for entry in result["table"] if entry["T_c"] == 92]
        self.assertEqual(len(tied), 3)
        self.assertEqual(len({(entry["D"], entry["T_c"]) for entry in tied}), 1)
        self.assertEqual((result["j"], result["gamma"]), (4, 0.82))

    def test_confirm_sign_counts_and_ties(self):
        final = self.root / "final.json"
        final.write_text('{"j":4,"gamma":0.8}', encoding="utf-8")
        arms = [jg.hg_arm(0.80), *jg.BASELINES, jg.hg_arm(0.98)]
        scenarios = []
        for name, exceed in (("c1", 17), ("c2", 16)):
            def value(arm, rnd, exceed=exceed):
                if arm == jg.hg_arm(0.80):
                    return 70 if rnd <= exceed else jg.TMAX
                if arm in jg.BASELINES:
                    h = 70 if rnd <= exceed else jg.TMAX
                    return 100 if rnd <= exceed else h / 0.9
                return 75
            scenarios.append(write_dir(self.root, name, 24, arms, value))
        result = jg.confirm(*scenarios, final)
        self.assertFalse(result["claim"])
        for name, expected in (("c1", 17), ("c2", 16)):
            for condition in ("leastloaded", "sed", "time"):
                item = result["scenarios"][name]["conditions"][condition]
                self.assertEqual(item["exceedances"], expected)
                self.assertIs(item["supported"], expected == 17)
        self.assertLessEqual(jg.sign_pvalue(17), jg.SIGN_ALPHA)
        self.assertGreater(jg.sign_pvalue(16), jg.SIGN_ALPHA)
        self.assertEqual(jg.sign_pvalue(24), 1 / 2 ** 24)

    def test_confirm_24_exceedances_and_duplicate_gamma_arm(self):
        final = self.root / "final.json"
        final.write_text('{"j":4,"gamma":0.98}', encoding="utf-8")
        arms = [jg.hg_arm(0.98), *jg.BASELINES]
        values = {jg.hg_arm(0.98): 70, "leastloaded": 90, "sed": 91}
        c1 = write_dir(self.root, "c1", 24, arms, values)
        c2 = write_dir(self.root, "c2", 24, arms, values)
        result = jg.confirm(c1, c2, final)
        self.assertTrue(result["claim"])
        self.assertEqual(result["scenarios"]["c1"]["conditions"]["time"]["exceedances"], 24)
        self.assertEqual(len(result["scenarios"]["c1"]["medians_s"]), 3)

    def test_errors_for_incomplete_wrong_counts_unknown_arm_and_missing_file(self):
        values = {arm: 80 for arm in p1_arms()}
        dirs = [write_dir(self.root, f"p1-j{j}", 6, p1_arms(), values)
                for j in (4, 5, 6)]

        def cli_error(message, paths=dirs):
            error = io.StringIO()
            with contextlib.redirect_stderr(error), self.assertRaises(SystemExit) as raised:
                jg.main(["phase1", "--dirs", *(str(path) for path in paths),
                         "--out", str(self.root / "out.json")])
            self.assertEqual(raised.exception.code, 2)
            self.assertIn(message, error.getvalue())

        cli_error("wrong directory count/levels", dirs[:2])
        write_dir(self.root, "p1-j4", 6, p1_arms(), values,
                  omit=(6, jg.hg_arm(0.98)))
        cli_error("incomplete round 6")
        write_dir(self.root, "p1-j4", 6, p1_arms(), values)
        write_dir(self.root, "p1-j5", 5, p1_arms(), values)
        cli_error("wrong round count")
        write_dir(self.root, "p1-j5", 6, p1_arms(), values)
        write_dir(self.root, "p1-j6", 6, ["ewma"], {"ewma": 80})
        cli_error("unknown arm label: ewma")
        FILES.pop(str(dirs[2] / "results.csv"))
        cli_error("missing file")

        with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit) as raised:
            jg.main(["phase2", "--dirs", str(dirs[0]), "--out", str(self.root / "out.json")])
        self.assertEqual(raised.exception.code, 2)

    def test_integrity_prints_counts_only(self):
        values = {"leastloaded": 123.456789, "sed": "not-a-number",
                  jg.hg_arm(0.80): ""}
        path = write_dir(self.root, "c1", 2, list(values), values,
                         omit=(2, jg.hg_arm(0.80)))
        full = write_dir(self.root, "p1-j4", 6, p1_arms(),
                         {arm: 81.234567 for arm in p1_arms()})
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            jg.main(["integrity", "--dirs", str(path), str(full)])
        raw = output.getvalue()
        self.assertEqual(raw.splitlines(), [
            "dir#1 arms=3 rounds_present=2 rounds_complete=1 rounds_missing=23 rows=10",
            "dir#2 arms=6 rounds_present=6 rounds_complete=6 rounds_missing=0 rows=72",
            "total_dirs=2",
        ])
        for secret in (str(path), path.name, full.name, *values,
                       "g080", "123.456789", "81.234567", "999.123456"):
            self.assertNotIn(secret, raw)
        with self.assertRaisesRegex(jg.AnalysisError, "invalid elapsed_s"):
            jg.load_dir(path)

        error = io.StringIO()
        with contextlib.redirect_stderr(error), self.assertRaises(SystemExit) as raised:
            jg.main(["integrity", "--dirs", str(self.root / "missing")])
        self.assertEqual(raised.exception.code, 2)
        self.assertNotIn(str(self.root), error.getvalue())

    def test_integrity_reads_only_structural_csv_columns(self):
        path = write_dir(self.root, "p1-j4", 6, p1_arms(),
                         {arm: 81 for arm in p1_arms()})
        csv_path = str(path / "results.csv")
        source = csv.DictReader(io.StringIO(FILES[csv_path]))
        with io.StringIO() as handle:
            writer = csv.DictWriter(handle, ["arm", "round", "build_idx"])
            writer.writeheader()
            for row in source:
                writer.writerow({key: row[key] for key in writer.fieldnames})
            FILES[csv_path] = handle.getvalue()
        self.assertEqual(jg.integrity([path]).splitlines(), [
            "dir#1 arms=6 rounds_present=6 rounds_complete=6 rounds_missing=0 rows=72",
            "total_dirs=1",
        ])

    def test_bootstrap_reproducibility_and_json_rounding(self):
        paths = []
        for j in (4, 5):
            def value(arm, rnd, j=j):
                if arm in jg.BASELINES:
                    return 100 + (rnd % 3) + j / 10
                return 75 + (rnd % 4) + j / 10 + 10 * (jg.arm_gamma(arm) - .80)
            paths.append(write_dir(self.root, f"p2-j{j}", 10, p2_arms(), value))
        first = jg.phase2(paths)
        second = jg.phase2(paths)
        self.assertEqual(first, second)
        self.assertGreater(first["q"], 0)
        out = self.root / "final.json"
        jg.emit(first, out)
        self.assertEqual(json.loads(out.read_text(encoding="utf-8"))["q"], round(first["q"], 6))
        self.assertEqual(first["q"], second["q"])

    def test_help_lists_subcommands(self):
        result = subprocess.run([sys.executable, str(ROOT / "scripts/analyze_jg.py"), "--help"],
                                capture_output=True, text=True, cwd=ROOT)
        self.assertEqual(result.returncode, 0)
        for word in ("integrity", "phase1", "phase2", "confirm"):
            self.assertIn(word, result.stdout)


if __name__ == "__main__":
    unittest.main()
