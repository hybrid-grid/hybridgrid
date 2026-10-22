#!/bin/bash
# Source this file (". /d/bench-work/env.sh") in Git Bash before any benchmark.
# Puts the python3 shim (venv, pandas < 3) first on PATH and defines paths.
# Pilot/exploratory data and official data both go under $BENCH_OUT, never
# inside the repository.
export BENCH_ROOT=/d/bench-work
export BENCH_REPO=$BENCH_ROOT/hybridgrid
export BENCH_OUT=$BENCH_ROOT/out
# Do NOT export MSYS_NO_PATHCONV here: native tools (git, docker) need Git
# Bash to convert /d/... paths. Set it per command when a container path such
# as /gate must stay untouched (the runner already handles this itself).
case ":$PATH:" in
    *":$BENCH_ROOT/bin:"*) ;;
    *) PATH="$BENCH_ROOT/bin:$PATH"; export PATH ;;
esac
mkdir -p "$BENCH_OUT"
# The runner does NOT read BENCH_OUT: it defaults to /tmp/bench_exp_*. Always
# pass the output directory explicitly, one directory per scenario, e.g.
#   OUT_DIR=$BENCH_OUT/E1-j8 bash scripts/bench_exp.sh
# and keep the preflight report next to the results:
#   EXPECT_COMMIT=<sha> STRICT=1 bash /d/bench-work/preflight.sh "$OUT_DIR/preflight.txt"
