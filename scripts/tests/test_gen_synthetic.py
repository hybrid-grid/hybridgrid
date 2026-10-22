"""Tests for the deterministic C/C++ workload generator (no compiler needed)."""

import hashlib
import json
import math
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


GENERATOR = Path(__file__).resolve().parents[1] / "gen_synthetic_cpp.py"


def run_generator(output, seed=2026, *options):
    result = subprocess.run(
        [sys.executable, str(GENERATOR), "--out", str(output), "--seed", str(seed), *options],
        capture_output=True,
        text=True,
    )
    if result.returncode:
        raise RuntimeError(f"generator failed: {result.stderr}")
    return json.loads((output / "manifest.json").read_text(encoding="ascii")), result.stdout


class TestSyntheticGenerator(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temp = tempfile.TemporaryDirectory()
        cls.addClassCleanup(cls.temp.cleanup)
        cls.output = Path(cls.temp.name) / "default"
        cls.manifest, cls.summary = run_generator(cls.output)

    def test_determinism_and_seed(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            first, _ = run_generator(root / "first", 91, "--files", "48")
            second, _ = run_generator(root / "second", 91, "--files", "48")
            different, _ = run_generator(root / "different", 92, "--files", "48")
            first_files = {p.name: p.read_bytes() for p in (root / "first").iterdir()}
            second_files = {p.name: p.read_bytes() for p in (root / "second").iterdir()}
            self.assertEqual(first_files, second_files)
            self.assertEqual(first["manifest_sha256"], second["manifest_sha256"])
            self.assertNotEqual(first["manifest_sha256"], different["manifest_sha256"])

    def test_file_counts_split_and_cap(self):
        entries = self.manifest["translation_units"]
        self.assertEqual(len(entries), 300)
        self.assertEqual(len({entry["file"] for entry in entries}), 300)
        c_count = sum(entry["language"] == "C" for entry in entries)
        self.assertLessEqual(abs(c_count - 300 * 0.3), 1)
        self.assertEqual(sum(entry["language"] == "C++" for entry in entries), 300 - c_count)
        self.assertTrue(all(1 <= entry["units"] <= 1500 for entry in entries))
        self.assertTrue(all(entry["hidden_units"] == 0 for entry in entries))

    def test_pareto_tail(self):
        units = sorted(entry["units"] for entry in self.manifest["translation_units"])
        median = (units[149] + units[150]) / 2
        top_count = math.ceil(len(units) * 0.05)
        top_mean = sum(units[-top_count:]) / top_count
        self.assertGreaterEqual(top_mean / median, 5.0)
        self.assertIn("top5%/median=", self.summary)
        self.assertEqual(len(self.summary.splitlines()), 1)

    def test_hidden_cost_changes_only_cpp_sources(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root)
            visible, _ = run_generator(root / "visible", 17, "--files", "32")
            hidden, _ = run_generator(
                root / "hidden", 17, "--files", "32", "--hidden-cost-fraction", "0.5"
            )
            self.assertEqual(
                [(e["file"], e["units"], e["position"]) for e in visible["translation_units"]],
                [(e["file"], e["units"], e["position"]) for e in hidden["translation_units"]],
            )
            for original, changed in zip(visible["translation_units"], hidden["translation_units"]):
                name = original["file"]
                before = (root / "visible" / name).read_bytes()
                after = (root / "hidden" / name).read_bytes()
                if original["language"] == "C":
                    self.assertEqual(before, after)
                    self.assertEqual(changed["hidden_units"], 0)
                else:
                    self.assertNotEqual(before, after)
                    self.assertGreater(changed["hidden_units"], 0)
                    self.assertIn(b"HgHidden<", after)

    def test_ascii_newlines_and_hashes(self):
        hashes = []
        for entry in self.manifest["translation_units"]:
            data = (self.output / entry["file"]).read_bytes()
            self.assertTrue(data.endswith(b"\n"))
            self.assertNotIn(b"\r", data)
            self.assertTrue(data.isascii())
            digest = hashlib.sha256(data).hexdigest()
            self.assertEqual(entry["sha256"], digest)
            hashes.append(digest)
        expected = hashlib.sha256("".join(sorted(hashes)).encode("ascii")).hexdigest()
        self.assertEqual(self.manifest["manifest_sha256"], expected)
        for name in ("common.h", "main.c", "CMakeLists.txt", "manifest.json"):
            data = (self.output / name).read_bytes()
            self.assertTrue(data.endswith(b"\n"))
            self.assertNotIn(b"\r", data)
            self.assertTrue(data.isascii())

    def test_cmake_lists_exact_translation_units_in_manifest_order(self):
        lines = (self.output / "CMakeLists.txt").read_text(encoding="ascii").splitlines()
        start = lines.index("add_library(hg_synth STATIC") + 1
        end = lines.index(")", start)
        listed = [line.strip() for line in lines[start:end]]
        expected = [entry["file"] for entry in sorted(self.manifest["translation_units"], key=lambda e: e["position"])]
        self.assertEqual(listed, expected)
        self.assertEqual(len(set(listed)), len(listed))
        self.assertEqual(sorted(entry["position"] for entry in self.manifest["translation_units"]), list(range(300)))
        self.assertIn("add_executable(hg_synth_main main.c)", lines)

    def test_no_shuffle_uses_filename_order(self):
        with tempfile.TemporaryDirectory() as root:
            manifest, _ = run_generator(Path(root) / "ordered", 5, "--files", "20", "--no-shuffle-order")
            self.assertEqual([e["position"] for e in manifest["translation_units"]], list(range(20)))


if __name__ == "__main__":
    unittest.main()
