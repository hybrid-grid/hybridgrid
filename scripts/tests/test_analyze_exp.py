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
    def test_exp3_top5_ceiling_ties_strength_and_last_finisher(self):
        candidates = [
            {"address": "w1", "active": 0, "max_parallel": 1, "cpu_millis": 600},
            {"address": "w2", "active": 0, "max_parallel": 1, "cpu_millis": 1100},
            {"address": "w3", "active": 0, "max_parallel": 1, "cpu_millis": 1000},
        ]
        records = []
        for sequence in range(1, 22):
            address = "w2" if sequence == 1 else "w3" if sequence == 2 else "w1"
            snapshot = candidates if sequence != 2 else [dict(candidates[0], active=1),
                                                            dict(candidates[1], active=1),
                                                            candidates[2]]
            record = task(address, sequence, snapshot,
                          "2026-01-01T00:02:10Z" if sequence == 3
                          else "2026-01-01T00:02:00Z")
            record.update(preprocessed_size_bytes=1000 if sequence <= 3 else 1,
                          worker_cpu_millis={"w1": 600, "w2": 1100, "w3": 1000}[address],
                          dispatch_ts="2026-01-01T00:01:45Z" if sequence == 21
                          else "2026-01-01T00:01:40Z")
            records.append(record)
        rows = analyze.pd.DataFrame([{"start_ts": 1767225700, "end_ts": 1767225718},
                                     {"start_ts": 1767225700, "end_ts": 1767225720}])
        metrics = analyze.exp3_build_metrics(records, rows)
        self.assertEqual(metrics["top5_strong_share"], 1)
        self.assertEqual(metrics["top5_mean_cpu_millis"], 1050)
        self.assertEqual(metrics["top5_multi_choice_share"], 0.5)
        self.assertEqual(metrics["tail_s"], 15)
        self.assertEqual(metrics["last_finisher_size_bytes"], 1000)
        self.assertAlmostEqual(metrics["last_finisher_rank_pct"], 100 * 20 / 21)
        self.assertTrue(math.isnan(analyze.exp3_build_metrics(
            records, rows.drop(columns=["start_ts"]))["tail_s"]))

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

    def test_had_idle_and_free_shares(self):
        tasks = [task(candidates=[{"address": "w1", "active": active,
                                   "max_parallel": 2, "cpu_millis": 1000}])
                 for active in (0, 1, 1, 2)]
        metrics = analyze.candidate_metrics(tasks)
        self.assertEqual(metrics["had_idle_share"], 0.25)
        self.assertEqual(metrics["had_free_share"], 0.75)

    def test_missing_candidates_has_na_shares(self):
        metrics = analyze.candidate_metrics([{"worker_address": "w1"}])
        self.assertTrue(math.isnan(metrics["had_idle_share"]))
        self.assertTrue(math.isnan(metrics["had_free_share"]))

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
        self.assertEqual(metrics["had_idle_share"], 0)
        self.assertEqual(metrics["had_free_share"], 1)
        self.assertTrue(math.isnan(metrics["skip_idle_given_idle"]))
        self.assertEqual(metrics["strongest_free_slot"], 1)
        candidates[1]["healthy"] = True
        metrics = analyze.candidate_metrics([task("w1", candidates=candidates)])
        self.assertEqual(metrics["skip_idle_all"], 1)
        self.assertEqual(metrics["had_idle_share"], 1)
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
    def test_exp3_multi_build_clients_and_default_baselines(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            arms = ("leastloaded", "sed", "p90", "sed-p90", "hybrid-linucb")
            (root / "meta.json").write_text(json.dumps({"ARMS": list(arms), "REPS": 1,
                "CLIENTS": 2, "SESSION_BUILDS": 2}), encoding="utf-8")
            with (root / "results.csv").open("w", newline="", encoding="utf-8") as handle:
                writer = csv.writer(handle)
                writer.writerow(["arm", "round", "build_idx", "client", "elapsed_s",
                                 "cell_makespan_s", "start_ts", "end_ts"])
                for arm in arms:
                    for build_idx, start in ((1, 1767225600), (2, 1767225700)):
                        for client, end in ((1, start + 18), (2, start + 20)):
                            writer.writerow([arm, 1, build_idx, client, 20, 20, start, end])
                    records = []
                    for build_idx, size, address, cpu, finish, dispatch in (
                            (1, 100, "w1", 600, "2026-01-01T00:00:15Z",
                             "2026-01-01T00:00:05Z"),
                            (1, 200, "w2", 1100, "2026-01-01T00:00:10Z",
                             "2026-01-01T00:00:04Z"),
                            (2, 50, "w1", 600, "2026-01-01T00:02:10Z",
                             "2026-01-01T00:01:45Z"),
                            (2, 300, "w2", 1100, "2026-01-01T00:02:00Z",
                             "2026-01-01T00:01:40Z")):
                        sequence = len(records) + 1
                        candidates = [
                            {"address": "w1", "active": 0, "max_parallel": 1,
                             "cpu_millis": 600},
                            {"address": "w2", "active": 0, "max_parallel": 1,
                             "cpu_millis": 1100},
                            {"address": "w3", "active": 0, "max_parallel": 1,
                             "cpu_millis": 1000},
                        ]
                        record = task(address, sequence, candidates, finish)
                        record.update(build_id=f"{arm}-r1-b{build_idx}-c{1 if sequence % 2 else 2}",
                                      preprocessed_size_bytes=size,
                                      worker_cpu_millis=cpu, dispatch_ts=dispatch)
                        records.append(record)
                    (root / f"tasks_{arm}_round1.jsonl").write_text(
                        "\n".join(json.dumps(record) for record in records) + "\n",
                        encoding="utf-8")
            cells, dropped, _ = analyze.load_cells(root)
            self.assertFalse(dropped)
            row = cells[cells["arm"] == "p90"].iloc[0]
            self.assertEqual(row["build_1_top5_mean_cpu_millis"], 1100)
            self.assertEqual(row["build_1_last_finisher_size_bytes"], 100)
            self.assertEqual(row["build_1_tail_s"], 15)
            self.assertEqual(row["build_2_tail_s"], 15)
            self.assertEqual(row["tail_s"], 15)
            self.assertEqual(row["top5_mean_cpu_millis"], 1100)
            self.assertEqual(row["last_finisher_size_bytes"], 50)
            self.assertEqual(row["last_finisher_rank_pct"], 50)
            pairs = analyze.compare(cells, None, None)
            baseline_pairs = pairs[pairs["metric"] == "tail_s"]
            self.assertEqual(set(baseline_pairs["baseline"]), set(arms[:-1]))
            summary = root / "summary.md"
            analyze.write_summary(summary, cells, pairs, [], [], 20, 0.5,
                                  "pre_all", None)
            self.assertIn("cell-level metrics report the last build",
                          summary.read_text(encoding="utf-8"))
            self.assertIn("leastloaded, sed, p90, and sed-p90",
                          summary.read_text(encoding="utf-8"))

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
            for name in ("had_idle_share", "had_free_share"):
                self.assertIn(name, analyze.METRICS)
            with (root / "analysis" / "cells.csv").open(newline="", encoding="utf-8") as handle:
                self.assertTrue({"had_idle_share", "had_free_share"}.issubset(
                    csv.DictReader(handle).fieldnames))
            summary = (root / "analysis" / "summary.md").read_text(encoding="utf-8")
            self.assertIn("## had_idle_share", summary)
            self.assertIn("## had_free_share", summary)

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

    def test_legacy_csv_columns_and_values_remain_unchanged(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "results.csv").write_text(
                "arm,round,elapsed_s\nleastloaded,1,10\nhybrid-linucb,1,8\n",
                encoding="utf-8")
            for arm in ("leastloaded", "hybrid-linucb"):
                (root / f"tasks_{arm}_round1.jsonl").write_text(json.dumps({
                    "event": "task_completed", "success": True,
                    "ts": "2026-01-01T00:00:01Z", "worker_address": "w1",
                    "worker_cpu_millis": 1100, "dispatch_seq": 1}) + "\n",
                    encoding="utf-8")
            self.assertEqual(analyze.main([str(root), "--out", str(root / "analysis")]), 0)
            with (root / "analysis" / "cells.csv").open(newline="", encoding="utf-8") as handle:
                reader = csv.DictReader(handle)
                expected_columns = [
                    "round", "arm", "n_tasks", "makespan_s",
                    "skip_idle_all", "skip_idle_given_idle", "strongest_idle",
                    "free_slot_skip_all", "free_slot_skip_given_free", "strongest_free_slot",
                    "had_idle_share", "had_free_share", "drift_pre_share",
                    "drift_pre_all_share", "drift_post_share", "drift_pre_n",
                    "drift_pre_all_n", "drift_post_n", "drift_post_early_share",
                    "drift_post_late_share", "excess_target_share_pre",
                    "excess_target_share_post", "time_to_adapt", "time_to_adapt_rank",
                    "time_to_adapt_censored", "target_after_recovery_share",
                    "target_after_recovery_tasks", "drift_on_phases_share",
                    "drift_on_phases_n", "drift_on_phases_target_n",
                    "drift_off_phases_share", "drift_off_phases_n",
                    "drift_off_phases_target_n", "drift_build_boundary_phases",
                    "drift_window_flag", "last_finisher_address",
                    "last_finisher_cpu_millis",
                ]
                self.assertEqual(reader.fieldnames, expected_columns)
                rows = list(reader)
            self.assertEqual(len(rows), 2)
            for row in rows:
                expected = dict.fromkeys(expected_columns, "NA")
                expected.update(round="1", arm=row["arm"], n_tasks="1",
                                makespan_s="8.0" if row["arm"] == "hybrid-linucb" else "10.0",
                                drift_pre_n="0", drift_pre_all_n="0", drift_post_n="0",
                                time_to_adapt="NA", time_to_adapt_censored="False",
                                drift_build_boundary_phases="", drift_window_flag="",
                                last_finisher_address="w1", last_finisher_cpu_millis="1100")
                self.assertEqual(row, expected)
            with (root / "analysis" / "pairs.csv").open(newline="", encoding="utf-8") as handle:
                reader = csv.DictReader(handle)
                self.assertEqual(reader.fieldnames, [
                    "metric", "treatment", "baseline", "n", "treatment_median",
                    "baseline_median", "median_diff", "ci_lo", "ci_hi", "p_raw",
                    "r_rb", "n_zero", "p_holm"])
                pair_rows = list(reader)
            self.assertEqual([row["metric"] for row in pair_rows], [
                "makespan_s", "drift_pre_n", "drift_pre_all_n", "drift_post_n"])
            for row in pair_rows:
                expected = {"metric": row["metric"], "treatment": "hybrid-linucb",
                            "baseline": "leastloaded", "n": "1", "p_raw": "1.0",
                            "p_holm": "1.0"}
                if row["metric"] == "makespan_s":
                    expected.update(treatment_median="8.0", baseline_median="10.0",
                                    median_diff="-2.0", ci_lo="-2.0", ci_hi="-2.0",
                                    r_rb="-1.0", n_zero="0")
                else:
                    expected.update(treatment_median="0.0", baseline_median="0.0",
                                    median_diff="0.0", ci_lo="0.0", ci_hi="0.0",
                                    r_rb="0.0", n_zero="1")
                self.assertEqual(row, expected)

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
        for label in ("p90", "sed-p90"):
            self.assertEqual(analyze.parse_arm(label)["base"], label)
        for bad in ("sed-g095", "leastloaded-l025", "hybrid-linucb-d-l025-g095",
                    "hybrid-linucb-d-g999", "hybrid-linucb-d-unknown",
                    "p90-arm", "p90-g095", "p90-l025", "sed-p90-arm",
                    "sed-p90-g095", "sed-p90-l025"):
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                analyze.parse_arm(bad)


if __name__ == "__main__":
    unittest.main()
