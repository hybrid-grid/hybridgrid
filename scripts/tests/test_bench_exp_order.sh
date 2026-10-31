#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
source <(awk '/^order_for_round\(\) \{/ {copy=1} /^generate_compose\(\) \{/ {copy=0} copy' "$RUNNER")
ARMS='leastloaded sed ewma hybrid-linucb-d-arm-g098 hybrid-linucb-d-arm-g080'
SEED_BASE=3000
ORDER_MODE=latin5
for round in $(seq 1 15); do order_for_round "$round"; done > "$TMP/latin"
for round in $(seq 1 15); do order_for_round "$round"; done > "$TMP/repeated"
diff -u "$TMP/latin" "$TMP/repeated"
python3 - "$TMP/latin" <<'PY'
import collections,sys
rows=[line.split() for line in open(sys.argv[1])]
assert len(rows)==15 and all(len(row)==5 for row in rows)
arms=set(rows[0])
assert all(set(row)==arms for row in rows)
assert all(collections.Counter(row[position] for row in rows)=={arm:3 for arm in arms}
           for position in range(5))
PY
for round in $(seq 1 15); do SEED_BASE=3001 order_for_round "$round"; done > "$TMP/other"
if cmp -s "$TMP/other" "$TMP/latin"; then
    printf 'different seeds produced the same schedule\n' >&2; exit 1
fi
ORDER_MODE=shuffle
for round in $(seq 1 20); do order_for_round "$round"; done > "$TMP/shuffle"
python3 - "$TMP/shuffle" "$ARMS" "$SEED_BASE" <<'PY'
import random,sys
rows=[line.strip() for line in open(sys.argv[1])]
for round_number, actual in enumerate(rows,1):
    arms=sys.argv[2].split()
    random.Random(int(sys.argv[3])+round_number).shuffle(arms)
    assert actual==' '.join(arms)
PY
if ORDER_MODE=latin5 DRY_RUN=1 ARMS='leastloaded sed ewma hybrid-linucb-d-arm-g098' REPS=15 \
    bash "$RUNNER" > "$TMP/out" 2> "$TMP/err"; then
    printf 'expected 4-arm rejection\n' >&2; exit 1
fi
grep -Fq 'requires exactly 5 arms' "$TMP/err"
if ORDER_MODE=latin5 DRY_RUN=1 ARMS="$ARMS" REPS=12 \
    bash "$RUNNER" > "$TMP/out" 2> "$TMP/err"; then
    printf 'expected REPS=12 rejection\n' >&2; exit 1
fi
grep -Fq 'requires REPS divisible by 5' "$TMP/err"
printf 'bench_exp order tests passed\n'
