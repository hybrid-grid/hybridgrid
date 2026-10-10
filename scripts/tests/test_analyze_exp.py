"""Synthetic contract tests for scripts/analyze_exp.py."""

import csv
import importlib.util
import json
import math
import tempfile
import unittest
from pathlib import Path

import numpy as np


MODULE_PATH = Path(__file__).resolve().parents[1] / "analyze_exp.py"
SPEC = importlib.util.spec_from_file_location("analyze_exp", MODULE_PATH)
analyze = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(analyze)
CHECK_SPEC = importlib.util.spec_from_file_location("exp_checks", MODULE_PATH.with_name("exp_checks.py"))
checks = importlib.util.module_from_spec(CHECK_SPEC)
CHECK_SPEC.loader.exec_module(checks)


def task(address="w1", sequence=1, candidates=None, ts=None):
    return {"event": "task_completed", "ts": ts or f"2026-01-01T00:00:{sequence:02d}Z",
            "success": True, "worker_address": address, "dispatch_seq": sequence,
            "worker_cpu_millis": 1100, "candidates": candidates}


class StatisticsTests(unittest.TestCase):
    def test_rank_biserial_and_holm_hand_calculation(self):
        # Absolute ranks 1, 2, 3: W+ = 2, W- = 4.
        self.assertAlmostEqual(analyze.rank_biserial([-1, 2, -3, 0]), -1 / 3)
        self.assertEqual(analyze.holm([0.03, 0.01, 0.04]), [0.06, 0.03, 0.06])

    def test_all_zero_wilcoxon(self):
        result = analyze.paired_stats([2, 2, 2], [2, 2, 2], np.random.default_rng(1))
        self.assertEqual(result["p_raw"], 1)
        self.assertEqual(result["r_rb"], 0)
        self.assertEqual(result["n_zero"], 3)

    def test_exact_wilcoxon_without_ties(self):
        result = analyze.paired_stats([1, 2, 3, 4, 5], [0] * 5, np.random.default_rng(1))
        self.assertAlmostEqual(result["p_raw"], 0.0625)


class MetricTests(unittest.TestCase):
    def test_idle_skip_strongest_and_free_slot(self):
        first = [
            {"address": "w1", "active": 1, "max_parallel": 2, "cpu_millis": 600},
            {"address": "w2", "active": 0, "max_parallel": 1, "cpu_millis": 1100},
            {"address": "w3", "active": 0, "max_parallel": 1, "cpu_millis": 1100},
        ]
        second = [dict(first[0], active=0), dict(first[1], active=1),
                  dict(first[2], active=1)]
        no_idle = [dict(first[0], active=2), dict(first[1], active=1),
                   dict(first[2], active=1)]
        metrics = analyze.candidate_metrics([
            task("w1", candidates=first), task("w1", candidates=second),
            task("w1", candidates=no_idle), task("w3", candidates=first),
        ])
        self.assertEqual(metrics["skip_idle_all"], 0.25)
        self.assertAlmostEqual(metrics["skip_idle_given_idle"], 1 / 3)
        self.assertAlmostEqual(metrics["strongest_idle"], 2 / 3)  # w3 ties w2
        self.assertEqual(metrics["free_slot_skip_all"], 0.0)

    def test_no_idle_candidate_conditional_is_na(self):
        candidates = [{"address": "w1", "active": 1, "max_parallel": 2,
                       "cpu_millis": 1000}]
        metrics = analyze.candidate_metrics([task(candidates=candidates)])
        self.assertEqual(metrics["skip_idle_all"], 0)
        self.assertTrue(math.isnan(metrics["skip_idle_given_idle"]))
        self.assertTrue(math.isnan(metrics["strongest_idle"]))

    def test_unhealthy_candidates_do_not_count_as_idle_or_free(self):
        candidates = [
            {"address": "w1", "active": 1, "max_parallel": 2, "cpu_millis": 600},
            {"address": "w2", "active": 0, "max_parallel": 1,
             "cpu_millis": 1100, "healthy": False},
        ]
        metrics = analyze.candidate_metrics([task("w1", candidates=candidates)])
        self.assertEqual(metrics["skip_idle_all"], 0)
        self.assertTrue(math.isnan(metrics["skip_idle_given_idle"]))
        self.assertEqual(metrics["strongest_free_slot"], 1)
        candidates[1]["healthy"] = True
        metrics = analyze.candidate_metrics([task("w1", candidates=candidates)])
        self.assertEqual(metrics["skip_idle_all"], 1)
        self.assertEqual(metrics["strongest_free_slot"], 0)
        candidates[0]["healthy"] = False
        metrics = analyze.candidate_metrics([task("w1", candidates=candidates)])
        self.assertEqual(metrics["strongest_free_slot"], 0)
        self.assertEqual(metrics["free_slot_skip_all"], 1)

    def test_time_to_adapt_censor_and_floor(self):
        post = [task("target" if i < 3 else "other", i + 1) for i in range(6)]
        self.assertEqual(analyze.time_to_adapt(post, "target", 1.0, window=2),
                         ("4", 4.0, False))
        self.assertEqual(analyze.time_to_adapt(post[:3], "target", 1.0, window=2),
                         ("> 3", 4.0, True))
        display, rank, _ = analyze.time_to_adapt(post, "target", 0.049, window=2)
        self.assertEqual(display, "NA")
        self.assertTrue(math.isnan(rank))

    def test_drift_windows_phases_and_recovery(self):
        tasks = [task("target" if i <= 103 else "other", i) for i in range(101, 108)]
        events = [{"event": "injected_event", "kind": "drift_on",
                   "target": "target", "dispatch_count": 102},
                  {"event": "injected_event", "kind": "drift_off",
                   "target": "target", "dispatch_count": 105}]
        values = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5)
        self.assertEqual(values["drift_pre_share"], 1)
        self.assertAlmostEqual(values["drift_post_share"], 1 / 5)
        self.assertEqual(values["drift_phase_1_on_share"], 1 / 3)
        self.assertEqual(values["drift_phase_2_off_share"], 0)
        self.assertEqual(values["target_after_recovery_share"], 0)

    def test_drift_references_counts_early_late_and_excess(self):
        candidates = [{"address": "target", "active": 0, "max_parallel": 1,
                       "cpu_millis": 1000},
                      {"address": "other", "active": 0, "max_parallel": 1,
                       "cpu_millis": 1000}]
        no_target_slot = [dict(candidates[0], active=1), candidates[1]]
        tasks = [task("target" if i <= 100 else "other", i,
                      candidates=candidates) for i in range(1, 102)]
        tasks += [task("target", 102, candidates=candidates),
                  task("other", 103, candidates=no_target_slot)]
        tasks += [task("target" if i == 104 else "other", i,
                       candidates=candidates) for i in range(104, 154)]
        event = [{"kind": "drift_on", "dispatch_count": 102, "target": "target"}]
        values = analyze.drift_metrics(tasks, event, "fallback", 2, 0.5)
        self.assertEqual((values["drift_pre_n"], values["drift_pre_all_n"],
                          values["drift_post_n"]), (2, 102, 51))
        self.assertEqual(values["drift_pre_share"], 0.5)
        self.assertAlmostEqual(values["drift_pre_all_share"], 101 / 102)
        self.assertAlmostEqual(values["drift_post_share"], 1 / 51)
        self.assertEqual(values["drift_post_early_share"], 1 / 50)
        self.assertEqual(values["drift_post_late_share"], 0)
        self.assertEqual(values["excess_target_share_pre"], 0)
        self.assertAlmostEqual(values["excess_target_share_post"],
                               (0 + 0.5 - 49 * 0.5) / 51)

    def test_tta_reference_selection_floor_and_fixed_threshold(self):
        tasks = [task("other", i) for i in range(1, 101)]
        tasks += [task("target" if i == 101 else "other", i)
                  for i in range(101, 111)]
        tasks += [task("other", i) for i in range(111, 114)]
        events = [{"kind": "drift_on", "dispatch_count": 110, "target": "target"}]
        default = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5)
        pre = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5, "pre")
        fixed = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5,
                                      "fixed", 0)
        self.assertEqual(default["time_to_adapt"], "NA")
        self.assertEqual(pre["time_to_adapt"], "2")
        self.assertEqual(fixed["time_to_adapt"], "2")
        self.assertEqual(analyze.time_to_adapt(tasks[-3:], "target", 0,
                                               window=2, fixed=True)[0], "2")
        post = [task("target", 1), task("target", 2), task("other", 3)]
        self.assertEqual(analyze.time_to_adapt(post, "target", 0.5,
                                               window=2, factor=0.1,
                                               fixed=True)[0], "3")

    def test_cell_end_off_is_neither_recovery_nor_new_phase(self):
        tasks = [task("target", i) for i in range(101, 106)]
        events = [{"kind": "drift_on", "dispatch_count": 102, "target": "target"},
                  {"kind": "drift_off", "dispatch_count": 104,
                   "detail": "cell_end cleanup"}]
        values = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5)
        self.assertTrue(math.isnan(values["target_after_recovery_share"]))
        self.assertNotIn("drift_phase_2_off_share", values)
        self.assertEqual(values["drift_phase_1_on_share"], 1)
        events[1]["detail"] = "stress_stopped"
        values = analyze.drift_metrics(tasks, events, "fallback", 2, 0.5)
        self.assertEqual(values["target_after_recovery_tasks"], 1)
        self.assertEqual(values["drift_phase_2_off_share"], 1)

    def test_onoff_aggregates_and_build_boundary_flag(self):
        tasks = [task("target" if i in (102, 103, 107) else "other", i)
                 for i in range(101, 109)]
        events = [
            {"kind": "drift_on", "dispatch_count": 101, "target": "target"},
            {"kind": "drift_note", "detail": "build_start b=2", "dispatch_count": 102},
            {"kind": "drift_off", "dispatch_count": 103},
            {"kind": "drift_on", "dispatch_count": 105},
            {"kind": "drift_off", "dispatch_count": 107},
            {"kind": "drift_off", "detail": "cell_end", "dispatch_count": 108},
        ]
        values = analyze.drift_metrics(tasks, events, "target", 2, 0.5)
        self.assertEqual(values["drift_on_phases_n"], 4)
        self.assertEqual(values["drift_on_phases_target_n"], 3)
        self.assertEqual(values["drift_on_phases_share"], 0.75)
        self.assertEqual(values["drift_off_phases_n"], 3)
        self.assertEqual(values["drift_off_phases_target_n"], 0)
        self.assertEqual(values["drift_off_phases_share"], 0)
        self.assertEqual(values["drift_build_boundary_phases"], "1_on")
        self.assertTrue(math.isnan(values["target_after_recovery_share"]))
        self.assertEqual(values["drift_phase_2_off_share"], 0)
        without_note = analyze.drift_metrics(tasks, [e for e in events if e["kind"] != "drift_note"],
                                             "target", 2, 0.5)
        self.assertEqual(values["drift_on_phases_share"], without_note["drift_on_phases_share"])

    def test_permanent_has_no_onoff_aggregate(self):
        values = analyze.drift_metrics([task("target", 102)],
                                       [{"kind": "drift_on", "dispatch_count": 101},
                                        {"kind": "drift_off", "dispatch_count": 102,
                                         "detail": "cell_end"}], "target", 2, 0.5)
        self.assertTrue(math.isnan(values["drift_on_phases_share"]))
        self.assertTrue(math.isnan(values["target_after_recovery_tasks"]))

    def test_hgtime_span_rule(self):
        records = [{"event": "task_completed", "build_id": "arm-r1-b1-c1",
                    "received_ts": "2026-01-01T00:00:01Z", "ts": "2026-01-01T00:00:11Z"}]
        row = {"build_idx": "1", "client": "1", "elapsed_s": "9.5"}
        self.assertEqual(checks.check_task_spans(records, [row], "arm", "1"), [])
        for elapsed in ("9.499", "30.001"):
            row["elapsed_s"] = elapsed
            self.assertIn("task_span_s=10.000", checks.check_task_spans(
                records, [row], "arm", "1")[0])


class InputTests(unittest.TestCase):
    def make_directory(self, root):
        with (root / "results.csv").open("w", newline="", encoding="utf-8") as handle:
            writer = csv.DictWriter(handle, fieldnames=["arm", "round", "order_pos",
                                                        "build_idx", "client", "elapsed_s",
                                                        "cell_makespan_s"])
            writer.writeheader()
            for rnd in (1, 2):
                for arm in ("leastloaded", "hybrid-linucb"):
                    if rnd == 2 and arm == "hybrid-linucb":
                        continue
                    writer.writerow({"arm": arm, "round": rnd, "order_pos": 1,
                                     "build_idx": 1, "client": "builder", "elapsed_s": 10,
                                     "cell_makespan_s": 10})
                    with (root / f"tasks_{arm}_round{rnd}.jsonl").open("w", encoding="utf-8") as log:
                        log.write(json.dumps(task(candidates=[{"address": "w1", "active": 0,
                            "max_parallel": 1, "cpu_millis": 1000}])) + "\n")
                        log.write(json.dumps({"event": "injected_event", "kind": "drift_note",
                                              "ts": "2026-01-01T00:00:02Z"}) + "\n")

    def test_injected_event_excluded_and_incomplete_block_dropped(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.make_directory(root)
            cells, dropped, _ = analyze.load_cells(root)
            self.assertEqual(len(cells), 2)
            self.assertNotIn("build_1_makespan_s", cells)
            self.assertNotIn("client_builder_elapsed_s", cells)
            self.assertEqual(set(cells["n_tasks"]), {1})
            self.assertEqual(len(dropped), 1)
            self.assertIn("round 2", dropped[0])
            self.assertEqual(analyze.main([str(root), "--out", str(root / "analysis"),
                                           "--strict"]), 2)

    def test_old_format_is_na(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "scheduler,round,order_pos,workers,elapsed_s\nleastloaded,1,1,5,10\n",
                encoding="utf-8")
            (root / "tasks_leastloaded_round1.jsonl").write_text(
                json.dumps({"ts": "2026-01-01T00:00:01Z", "success": True,
                            "worker_id": "old-worker", "worker_cpu_millis": 600}) + "\n",
                encoding="utf-8")
            cells, dropped, notes = analyze.load_cells(root)
            self.assertFalse(dropped)
            self.assertEqual(cells.iloc[0]["makespan_s"], 10)
            self.assertTrue(math.isnan(cells.iloc[0]["strongest_idle"]))
            self.assertTrue(math.isnan(cells.iloc[0]["time_to_adapt_rank"]))
            self.assertTrue(math.isnan(cells.iloc[0]["excess_target_share_post"]))
            self.assertTrue(notes)

    def test_session_build_and_client_makespans(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "meta.json").write_text(json.dumps({"ARMS": "leastloaded",
                "REPS": "1", "CLIENTS": "2", "SESSION_BUILDS": "2", "MIN_TASKS": "1"}),
                encoding="utf-8")
            with (root / "results.csv").open("w", newline="", encoding="utf-8") as handle:
                writer = csv.writer(handle)
                writer.writerow(["arm", "round", "build_idx", "client", "elapsed_s",
                                 "cell_makespan_s"])
                writer.writerows([
                    ["leastloaded", 1, 1, 1, 3, 5],
                    ["leastloaded", 1, 1, 2, 5, 5],
                    ["leastloaded", 1, 2, 1, 7, 8],
                    ["leastloaded", 1, 2, 2, 8, 8],
                ])
            (root / "tasks_leastloaded_round1.jsonl").write_text(
                json.dumps(task(candidates=[{"address": "w1", "active": 0,
                    "max_parallel": 1, "cpu_millis": 1000}])) + "\n", encoding="utf-8")
            cells, dropped, _ = analyze.load_cells(root)
            self.assertFalse(dropped)
            row = cells.iloc[0]
            self.assertEqual(row["makespan_s"], 13)
            self.assertEqual(row["build_1_makespan_s"], 5)
            self.assertEqual(row["build_2_makespan_s"], 8)
            self.assertEqual(row["client_1_elapsed_s"], 10)
            self.assertEqual(row["client_2_elapsed_s"], 13)
            self.assertEqual(row["build_1_client_1_elapsed_s"], 3)

    def test_one_build_two_clients_emits_only_client_metrics(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "arm,round,build_idx,client,elapsed_s,cell_makespan_s\n"
                "leastloaded,1,1,1,3,5\nleastloaded,1,1,2,5,5\n",
                encoding="utf-8")
            (root / "tasks_leastloaded_round1.jsonl").write_text(
                json.dumps(task()) + "\n", encoding="utf-8")
            cells, dropped, _ = analyze.load_cells(root)
            self.assertFalse(dropped)
            self.assertEqual(cells.iloc[0]["client_1_elapsed_s"], 3)
            self.assertNotIn("build_1_makespan_s", cells)
            self.assertNotIn("build_1_client_1_elapsed_s", cells)

    def test_two_builds_one_client_emits_only_build_metrics(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "arm,round,build_idx,client,elapsed_s,cell_makespan_s\n"
                "leastloaded,1,1,1,3,3\nleastloaded,1,2,1,5,5\n",
                encoding="utf-8")
            (root / "tasks_leastloaded_round1.jsonl").write_text(
                json.dumps(task()) + "\n", encoding="utf-8")
            cells, dropped, _ = analyze.load_cells(root)
            self.assertFalse(dropped)
            self.assertEqual(cells.iloc[0]["makespan_s"], 8)
            self.assertEqual(cells.iloc[0]["build_1_makespan_s"], 3)
            self.assertNotIn("client_1_elapsed_s", cells)
            self.assertNotIn("build_1_client_1_elapsed_s", cells)

    def test_missing_client_or_build_invalidates_block(self):
        cases = [
            ([(1, 1), (1, 2), (2, 1)], "missing client"),
            ([(1, 1), (1, 2), (1, 2), (2, 1), (2, 2)], "duplicate client"),
            ([(1, 1), (1, 2), (3, 1), (3, 2)], "wrong build index"),
            ([(1, 1), (1, 3), (2, 1), (2, 3)], "wrong client index"),
        ]
        for entries, label in cases:
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                (root / "meta.json").write_text(json.dumps({"ARMS": "leastloaded",
                    "REPS": 1, "CLIENTS": 2, "SESSION_BUILDS": 2}), encoding="utf-8")
                with (root / "results.csv").open("w", newline="", encoding="utf-8") as handle:
                    writer = csv.writer(handle)
                    writer.writerow(["arm", "round", "build_idx", "client", "elapsed_s",
                                     "cell_makespan_s"])
                    for build_idx, client in entries:
                        writer.writerow(["leastloaded", 1, build_idx, client, 10, 10])
                cells, dropped, _ = analyze.load_cells(root)
                self.assertTrue(cells.empty)
                self.assertIn("client/build result", dropped[0])

    def test_complete_data_sets_expected_shape_without_meta(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "arm,round,build_idx,client,elapsed_s,cell_makespan_s\n"
                "leastloaded,1,1,1,10,10\nleastloaded,1,1,2,10,10\n"
                "leastloaded,1,2,1,10,10\nleastloaded,1,2,2,10,10\n"
                "leastloaded,2,1,1,10,10\nleastloaded,2,1,2,10,10\n"
                "leastloaded,2,2,1,10,10\n", encoding="utf-8")
            for rnd in (1, 2):
                (root / f"tasks_leastloaded_round{rnd}.jsonl").write_text(
                    json.dumps(task()) + "\n", encoding="utf-8")
            cells, dropped, _ = analyze.load_cells(root)
            self.assertEqual(len(cells), 1)
            self.assertIn("round 2", dropped[0])

    def test_short_drift_window_is_flagged_and_kept(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "arm,round,elapsed_s\nleastloaded,1,10\n", encoding="utf-8")
            records = [task("target", i, ts="2026-01-01T00:00:01Z")
                       for i in range(1, 105)]
            records.append({"event": "injected_event", "kind": "drift_on",
                            "dispatch_count": 102, "target": "target"})
            (root / "tasks_leastloaded_round1.jsonl").write_text(
                "\n".join(json.dumps(x) for x in records) + "\n", encoding="utf-8")
            cells, dropped, notes = analyze.load_cells(root)
            self.assertFalse(dropped)
            self.assertEqual(cells.iloc[0]["drift_window_flag"], "pre<15, post<100")
            summary = root / "summary.md"
            analyze.write_summary(summary, cells, analyze.pd.DataFrame(), dropped,
                                  notes, 20, 0.5, "pre_all", None)
            self.assertIn("Flagged drift cells", summary.read_text(encoding="utf-8"))
            self.assertIn("reference=pre_all share", summary.read_text(encoding="utf-8"))
            analyze.write_summary(summary, cells, analyze.pd.DataFrame(), dropped,
                                  notes, 20, 0.5, "fixed", 0.2)
            self.assertIn("reference=fixed share 0.2, threshold=0.2",
                          summary.read_text(encoding="utf-8"))

    def test_build_boundary_phase_appears_in_summary(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            cells = analyze.pd.DataFrame([{"round": 1, "arm": "leastloaded",
                "drift_build_boundary_phases": "1_on", "drift_window_flag": "",
                "last_finisher_address": "target", "last_finisher_cpu_millis": 1100}])
            summary = root / "summary.md"
            analyze.write_summary(summary, cells, analyze.pd.DataFrame(), [], [],
                                  20, 0.5, "pre_all", None)
            self.assertIn("| 1 | leastloaded | 1_on |", summary.read_text(encoding="utf-8"))

    def test_label_parsing(self):
        parsed = analyze.parse_arm("hybrid-linucb-d-arm-g098-l100")
        self.assertEqual(parsed["base"], "hybrid-linucb-d")
        self.assertEqual(parsed["discount_mode"], "arm")
        self.assertEqual(parsed["discount"], 0.98)
        self.assertEqual(parsed["load_penalty"], 1.0)
        self.assertEqual(analyze.parse_arm("hybrid-linucb-l0")["load_penalty"], 0)
        for bad in ("sed-g095", "leastloaded-l025", "hybrid-linucb-d-l025-g095",
                    "hybrid-linucb-d-g999", "hybrid-linucb-d-unknown"):
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                analyze.parse_arm(bad)


if __name__ == "__main__":
    unittest.main()
