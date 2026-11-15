#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
source <(awk '/^order_for_round\(\) \{/ {copy=1} /^generate_compose\(\) \{/ {copy=0} copy' "$RUNNER")

ARMS='leastloaded sed hybrid-linucb-d-arm-g080 hybrid-linucb-d-arm-g086 hybrid-linucb-d-arm-g092'
SEED_BASE=4004
ORDER_MODE=latin5
for round in $(seq 1 10); do order_for_round "$round"; done > "$TMP/latin5"
ORDER_MODE=latinN
for round in $(seq 1 10); do order_for_round "$round"; done > "$TMP/latinN"
diff -u "$TMP/latin5" "$TMP/latinN"

for arms in \
    'leastloaded sed' \
    'leastloaded sed hybrid-linucb-d-arm-g080' \
    'leastloaded sed hybrid-linucb-d-arm-g080 hybrid-linucb-d-arm-g086'; do
    ARMS=$arms
    n=$(wc -w <<< "$ARMS")
    for round in $(seq 1 "$((n * 2))"); do order_for_round "$round"; done > "$TMP/latinN"
    python3 - "$TMP/latinN" "$ARMS" "$n" <<'PY'
import collections, random, sys
rows = [line.split() for line in open(sys.argv[1])]
arms = sys.argv[2].split()
n = int(sys.argv[3])
assert len(rows) == 2*n
assert all(set(row) == set(arms) for row in rows)
for group in range(2):
    rng = random.Random(4004*1000 + group)
    r = rng.sample(range(n), n)
    c = rng.sample(range(n), n)
    s = rng.sample(arms, n)
    for row in range(n):
        assert rows[group*n+row] == [s[(r[row]+c[pos]) % n] for pos in range(n)]
    for pos in range(n):
        assert collections.Counter(rows[group*n+row][pos] for row in range(n)) == collections.Counter(arms)
PY
done

invalid() {
    local arms=$1 reps=$2 expected=$3
    if DRY_RUN=1 ORDER_MODE=latinN ARMS="$arms" REPS="$reps" bash "$RUNNER" > "$TMP/out" 2> "$TMP/err"; then
        printf 'expected latinN validation failure\n' >&2; exit 1
    fi
    grep -Fq "$expected" "$TMP/err" || { cat "$TMP/err" >&2; exit 1; }
}
invalid leastloaded 2 'requires at least 2 arms'
invalid 'leastloaded sed hybrid-linucb-d-arm-g080' 10 'requires REPS divisible by number of arms'
printf 'bench_exp latinN tests passed\n'
