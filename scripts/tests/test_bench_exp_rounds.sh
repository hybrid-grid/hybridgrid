#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Keep the runner's startup, resume, metadata, and round loop. Replace only a
# measured cell and Docker, so this test never starts a container.
python3 - "$RUNNER" "$TMP/runner.sh" <<'PY'
import sys
text = open(sys.argv[1], encoding='utf-8').read()
text = text.replace('ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"', 'ROOT="$TEST_REPO_ROOT"', 1)
text = text.replace('COMPOSE_FILE="$STRESS_DIR/docker-compose-exp.yml"', 'COMPOSE_FILE="$TEST_COMPOSE_FILE"', 1)
start = text.index('run_cell() {\n')
end = text.index('if [[ $WORKLOAD == syn ]]; then\n', start)
text = text[:start] + '''run_cell() {
    local arm=$1 round=$2 pos=$3 record=$4
    if [[ ! -f $OUT_DIR/.worker_limits.csv ]]; then
        printf 'worker,cpu_max,nproc\\nworker-1,1,1\\n' > "$OUT_DIR/.worker_limits.csv"
    fi
    if (( record == 0 )); then MIN_TASKS=1; return 0; fi
    if [[ ${FAKE_FAIL_AFTER_ONE:-0} == 1 && $round == 2 && -f $OUT_DIR/.fake_round2_started ]]; then
        FAIL_REASON='simulated interruption'
        return 1
    fi
    if [[ ${FAKE_FAIL_AFTER_ONE:-0} == 1 && $round == 2 ]]; then
        touch "$OUT_DIR/.fake_round2_started"
    fi
    printf '%s,%s,%s,1,1,1.000,1.000,1.000,2.000\\n' "$arm" "$round" "$pos" >> "$OUT_DIR/results.csv"
}

''' + text[end:]
with open(sys.argv[2], 'w', encoding='utf-8', newline='\n') as f:
    f.write(text)
PY
cat > "$TMP/docker" <<'SH'
#!/usr/bin/env bash
if [[ $1 == compose ]]; then
    case $6 in
        config) printf '%s\n' hgexp-coordinator:latest hgexp-worker-{1..5}:latest hgexp-builder:latest ;;
        ps) printf '%s\n' "$8" ;;
        exec) [[ $9 != pgrep ]] ;;
        *) : ;;
    esac
elif [[ $1 == inspect ]]; then
    printf 'sha256:%s\n' "$4"
elif [[ $1 == image && $2 == inspect ]]; then
    case $5 in
        sha256:*) printf '%s\n' "$5" ;;
        hgexp-*:latest) name=${5#hgexp-}; printf 'sha256:%s\n' "${name%:latest}" ;;
    esac
elif [[ $1 == info ]]; then
    printf '8,1000000\n'
fi
SH
chmod +x "$TMP/docker"
export PATH="$TMP:$PATH" TEST_REPO_ROOT="$ROOT" TEST_COMPOSE_FILE="$TMP/compose.yml"
OUT_DIR="$TMP/out"
export OUT_DIR

run() {
    env -u ROUND_END -u DRIFT_ON_MAX ARMS='leastloaded sed' ORDER_MODE=latinN REPS=4 \
        SESSION_BUILDS=1 COOLDOWN=0 "$@" bash "$TMP/runner.sh"
}
run ROUND_START=1 ROUND_END=1 > "$TMP/stdout" 2> "$TMP/stderr"
grep -Fq 'rounds through 1 of 4 completed; resume with ROUND_START=2' "$TMP/stdout"
python3 - "$OUT_DIR" <<'PY'
import csv, json, os, sys
p = sys.argv[1]
meta = json.load(open(os.path.join(p, 'meta.json')))
assert meta['REPS'] == '4' and meta['ROUND_END'] == '1'
assert 'DRIFT_ON_MAX' not in meta
assert len(list(csv.DictReader(open(os.path.join(p, 'results.csv'))))) == 2
PY

if run ROUND_START=2 ROUND_END=2 FAKE_FAIL_AFTER_ONE=1 > "$TMP/stdout" 2> "$TMP/stderr"; then
    printf 'expected simulated interruption\n' >&2; exit 1
fi
grep -Fq 'resume with ROUND_START=2' "$TMP/stderr"
[[ $(awk -F, '$2 == 2' "$OUT_DIR/results.csv" | wc -l) == 1 ]]
run ROUND_START=2 ROUND_END=2 > "$TMP/stdout" 2> "$TMP/stderr"
grep -Fq 'rounds through 2 of 4 completed; resume with ROUND_START=3' "$TMP/stdout"
run ROUND_START=3 ROUND_END=4 > "$TMP/stdout" 2> "$TMP/stderr"
! grep -Fq 'rounds through' "$TMP/stdout"
python3 - "$OUT_DIR" <<'PY'
import collections, csv, json, os, sys
p = sys.argv[1]
rows = list(csv.DictReader(open(os.path.join(p, 'results.csv'))))
orders = list(csv.DictReader(open(os.path.join(p, 'order_log.csv'))))
assert collections.Counter(row['round'] for row in rows) == {'1': 2, '2': 2, '3': 2, '4': 2}
assert collections.Counter(row['round'] for row in orders) == {'1': 1, '2': 1, '3': 1, '4': 1}
assert json.load(open(os.path.join(p, 'meta.json')))['ROUND_END'] == '4'
PY

if OUT_DIR="$TMP/no-meta" run ROUND_START=2 ROUND_END=2 > "$TMP/stdout" 2> "$TMP/stderr"; then
    printf 'expected resume without meta rejection\n' >&2; exit 1
fi
grep -Fq 'resume requires results.csv, order_log.csv, and meta.json' "$TMP/stderr"

for bounds in '1 0' '3 2' '1 5' '1 abc'; do
    read -r start end <<< "$bounds"
    if DRY_RUN=1 run ROUND_START="$start" ROUND_END="$end" > "$TMP/stdout" 2> "$TMP/stderr"; then
        printf 'expected invalid round bounds rejection\n' >&2; exit 1
    fi
done
DRY_RUN=1 run ROUND_START=2 ROUND_END=2 > "$TMP/stdout" 2> "$TMP/stderr"
grep -Fq '3 cells x 1 builds' "$TMP/stdout" && { printf 'unexpected cell count\n' >&2; exit 1; }
grep -Fq '2 cells x 1 builds' "$TMP/stdout"
! grep -Fq 'round 3:' "$TMP/stdout"

# A configured drift limit is recorded and fixed across resumes. An older
# metadata file with no key uses DRIFT_AT+40 as its comparison value.
DRIFT_OUT="$TMP/drift-out"
if OUT_DIR="$DRIFT_OUT" run ROUND_START=1 ROUND_END=1 DRIFT_ON_MAX=152 > "$TMP/stdout" 2> "$TMP/stderr"; then
    python3 - "$DRIFT_OUT/meta.json" <<'PY'
import json, sys
assert json.load(open(sys.argv[1]))['DRIFT_ON_MAX'] == '152'
PY
else
    cat "$TMP/stderr" >&2; exit 1
fi
if OUT_DIR="$DRIFT_OUT" run ROUND_START=2 ROUND_END=2 DRIFT_ON_MAX=153 > "$TMP/stdout" 2> "$TMP/stderr"; then
    printf 'expected drift limit resume mismatch\n' >&2; exit 1
fi
grep -Fq 'DRIFT_ON_MAX: meta.json=' "$TMP/stderr"
OUT_DIR="$DRIFT_OUT" run ROUND_START=2 ROUND_END=2 DRIFT_ON_MAX=152 > "$TMP/stdout" 2> "$TMP/stderr"
if run ROUND_START=4 ROUND_END=4 DRIFT_ON_MAX=161 > "$TMP/stdout" 2> "$TMP/stderr"; then
    printf 'expected legacy drift limit mismatch\n' >&2; exit 1
fi
grep -Fq 'DRIFT_ON_MAX: absent from meta.json (default' "$TMP/stderr"
run ROUND_START=4 ROUND_END=4 DRIFT_ON_MAX=160 > "$TMP/stdout" 2> "$TMP/stderr"
printf 'bench_exp round range tests passed\n'
