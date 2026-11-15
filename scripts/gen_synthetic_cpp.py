#!/usr/bin/env python3
"""Generate a deterministic, heavy-tailed C/C++ CMake workload.

The manifest's sha256 is SHA-256 of the concatenated, sorted TU SHA-256 hex
digests. Generated text is ASCII with LF line endings on every platform.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import random
import statistics
from pathlib import Path


VERSION = "1.0"
HIDDEN_CHAIN_LENGTH = 64


def write_text(path: Path, content: str) -> str:
    data = content.encode("ascii")
    path.write_bytes(data)
    return hashlib.sha256(data).hexdigest()


def common_header() -> str:
    return (
        "#ifndef HG_SYNTH_COMMON_H\n"
        "#define HG_SYNTH_COMMON_H\n"
        "#include <stdint.h>\n"
        "#if defined(__GNUC__) || defined(__clang__)\n"
        "#define HG_NOINLINE __attribute__((noinline))\n"
        "#elif defined(_MSC_VER)\n"
        "#define HG_NOINLINE __declspec(noinline)\n"
        "#else\n"
        "#define HG_NOINLINE\n"
        "#endif\n"
        "#endif\n"
    )


def c_source(index: int, units: int) -> str:
    lines = ['#include "common.h"', ""]
    for unit in range(units):
        name = f"hg_part_{index:04d}_{unit:04d}"
        lines.extend(
            [
                f"static HG_NOINLINE uint32_t {name}(uint32_t value) {{",
                "    for (uint32_t step = 0; step < 8u; ++step) {",
                f"        value = (value ^ (step + {unit + 1}u)) * 1664525u + 1013904223u;",
                "        value ^= value >> 13u;",
                "    }",
                "    return value;",
                "}",
            ]
        )
    lines.extend(["", f"uint32_t hg_tu_{index:04d}(void) {{", f"    uint32_t value = {index + 1}u;"])
    for unit in range(units):
        lines.append(f"    value ^= hg_part_{index:04d}_{unit:04d}(value);")
    lines.extend(["    return value;", "}", ""])
    return "\n".join(lines)


def cpp_source(index: int, units: int, hidden_units: int) -> str:
    visible_units = units - hidden_units
    lines = ['#include "common.h"', ""]
    for unit in range(visible_units):
        name = f"HgWork_{index:04d}_{unit:04d}"
        lines.extend(
            [
                "template <uint32_t Salt>",
                f"struct {name} {{",
                "    static HG_NOINLINE uint32_t run(uint32_t value) {",
                "        for (uint32_t step = 0; step < 8u; ++step) {",
                "            value = (value ^ (step + Salt)) * 1664525u + 1013904223u;",
                "            value ^= value >> 13u;",
                "        }",
                "        return value;",
                "    }",
                "};",
                f"template struct {name}<{unit + 1}u>;",
            ]
        )
    if hidden_units:
        lines.extend(
            [
                "template <int Depth, uint32_t Salt>",
                "struct HgHidden {",
                "    static constexpr uint32_t value =",
                "        (HgHidden<Depth - 1, Salt>::value * 1664525u) ^",
                "        (Salt + static_cast<uint32_t>(Depth) * 1013904223u);",
                "};",
                "template <uint32_t Salt>",
                "struct HgHidden<0, Salt> {",
                "    static constexpr uint32_t value = Salt;",
                "};",
            ]
        )
    lines.extend(["", f'extern "C" uint32_t hg_tu_{index:04d}(void) {{', f"    uint32_t value = {index + 1}u;"])
    for unit in range(visible_units):
        lines.append(f"    value ^= HgWork_{index:04d}_{unit:04d}<{unit + 1}u>::run(value);")
    for chain, start in enumerate(range(0, hidden_units, HIDDEN_CHAIN_LENGTH)):
        depth = min(HIDDEN_CHAIN_LENGTH, hidden_units - start)
        salt = index * (units + 1) + chain + 1
        lines.append(f"    value ^= HgHidden<{depth}, {salt}u>::value;")
    lines.extend(["    return value;", "}", ""])
    return "\n".join(lines)


def main_source(indices: list[int]) -> str:
    lines = ['#include "common.h"', "#include <stdio.h>", ""]
    for index in indices:
        lines.append(f"uint32_t hg_tu_{index:04d}(void);")
    lines.extend(["", "int main(void) {", "    uint32_t checksum = 0u;"])
    for index in indices:
        lines.append(f"    checksum ^= hg_tu_{index:04d}();")
    lines.extend(['    printf("%u\\n", checksum);', "    return 0;", "}", ""])
    return "\n".join(lines)


def cmake_source(filenames: list[str]) -> str:
    lines = [
        "cmake_minimum_required(VERSION 3.16)",
        "project(hg_synth C CXX)",
        "set(CMAKE_CXX_STANDARD 17)",
        "set(CMAKE_CXX_STANDARD_REQUIRED ON)",
        "add_library(hg_synth STATIC",
    ]
    lines.extend(f"    {name}" for name in filenames)
    lines.extend(
        [
            ")",
            "add_executable(hg_synth_main main.c)",
            "target_link_libraries(hg_synth_main PRIVATE hg_synth)",
            "",
        ]
    )
    return "\n".join(lines)


def generate(args: argparse.Namespace) -> dict:
    rng = random.Random(args.seed)
    output = Path(args.out)
    # Stale files from an earlier run (e.g. unit_0001.c next to a new
    # unit_0001.cc) would silently join the build, so require an empty tree.
    if output.exists() and any(output.iterdir()):
        raise SystemExit(f"output directory is not empty: {output}")
    output.mkdir(parents=True, exist_ok=True)

    c_count = math.floor(args.files * args.c_fraction + 0.5)
    c_indices = set(rng.sample(range(args.files), c_count))
    scale = args.median_units / (2.0 ** (1.0 / args.pareto_alpha))
    entries = []
    for index in range(args.files):
        sample = scale / (1.0 - rng.random()) ** (1.0 / args.pareto_alpha)
        units = min(args.max_units, max(1, math.floor(sample + 0.5)))
        is_c = index in c_indices
        hidden_units = 0 if is_c else math.floor(units * args.hidden_cost_fraction + 0.5)
        filename = f"unit_{index:04d}.{'c' if is_c else 'cc'}"
        source = c_source(index, units) if is_c else cpp_source(index, units, hidden_units)
        digest = write_text(output / filename, source)
        entries.append(
            {
                "file": filename,
                "language": "C" if is_c else "C++",
                "units": units,
                "hidden_units": hidden_units,
                "sha256": digest,
            }
        )

    order = list(range(args.files))
    if args.shuffle_order:
        rng.shuffle(order)
    for position, index in enumerate(order):
        entries[index]["position"] = position
    ordered_names = [entries[index]["file"] for index in order]

    write_text(output / "common.h", common_header())
    write_text(output / "main.c", main_source(order[: min(3, args.files)]))
    write_text(output / "CMakeLists.txt", cmake_source(ordered_names))

    aggregate = hashlib.sha256("".join(sorted(entry["sha256"] for entry in entries)).encode("ascii")).hexdigest()
    manifest = {
        "generator_version": VERSION,
        "seed": args.seed,
        "files": args.files,
        "c_fraction": args.c_fraction,
        "pareto_alpha": args.pareto_alpha,
        "median_units": args.median_units,
        "max_units": args.max_units,
        "hidden_cost_fraction": args.hidden_cost_fraction,
        "shuffle_order": args.shuffle_order,
        "manifest_sha256": aggregate,
        "translation_units": entries,
    }
    write_text(output / "manifest.json", json.dumps(manifest, sort_keys=True, indent=2, ensure_ascii=True) + "\n")

    costs = sorted(entry["units"] for entry in entries)
    median = statistics.median(costs)
    p95 = costs[math.ceil(0.95 * len(costs)) - 1]
    top_count = math.ceil(0.05 * len(costs))
    top_ratio = statistics.mean(costs[-top_count:]) / median
    print(
        f"files={args.files} C={c_count} C++={args.files - c_count} "
        f"median={median:g} p95={p95} max={costs[-1]} "
        f"top5%/median={top_ratio:.2f}"
    )
    return manifest


def positive_int(value: str) -> int:
    number = int(value)
    if number < 1:
        raise argparse.ArgumentTypeError("must be at least 1")
    return number


def finite_positive_float(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or number <= 0:
        raise argparse.ArgumentTypeError("must be a finite positive number")
    return number


def fraction(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or not 0 <= number <= 1:
        raise argparse.ArgumentTypeError("must be between 0 and 1")
    return number


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", required=True, type=Path)
    parser.add_argument("--seed", required=True, type=int)
    parser.add_argument("--files", type=positive_int, default=300)
    parser.add_argument("--c-fraction", type=fraction, default=0.3)
    parser.add_argument("--pareto-alpha", type=finite_positive_float, default=1.5)
    parser.add_argument("--median-units", type=positive_int, default=40)
    parser.add_argument("--max-units", type=positive_int, default=1500)
    parser.add_argument("--hidden-cost-fraction", type=fraction, default=0.0)
    parser.add_argument("--shuffle-order", action=argparse.BooleanOptionalAction, default=True)
    args = parser.parse_args()
    if args.max_units < args.median_units:
        parser.error("--max-units must be at least --median-units")
    return args


if __name__ == "__main__":
    generate(parse_args())
