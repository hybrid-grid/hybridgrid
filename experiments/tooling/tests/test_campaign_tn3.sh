#!/bin/bash
# Self-contained TN3 orchestration tests. No real runner, Docker, or preflight.
export PATH="/d/bench-work/bin:$PATH"  # the Windows Store python3 alias is a stub
set -u
mkdir -p /d/bench-work/tmp-tests
TEST_ROOT=$(mktemp -d /d/bench-work/tmp-tests/campaign_tn3.XXXXXX) || exit 1
export TEST_ROOT
trap 'rm -rf -- "$TEST_ROOT"' EXIT
mkdir -p "$TEST_ROOT/repo/scripts" "$TEST_ROOT/out"
git -C "$TEST_ROOT/repo" init -q || exit 1
printf 'fake analyzer version 1\n' > "$TEST_ROOT/repo/scripts/analyze_exp.py"
git -C "$TEST_ROOT/repo" add scripts/analyze_exp.py
git -C "$TEST_ROOT/repo" -c user.name=Test -c user.email=test@example.invalid commit -qm test || exit 1
EXPECT_TEST=$(git -C "$TEST_ROOT/repo" rev-parse HEAD)
export EXPECT_TEST
cat > "$TEST_ROOT/env.sh" <<'EOF'
export BENCH_OUT="$TEST_ROOT/out"
export BENCH_REPO="$TEST_ROOT/incorrect-default-repo"
export PATH="/d/bench-work/bin:$PATH"
EOF
cat > "$TEST_ROOT/powershell.exe" <<'EOF'
#!/bin/bash
[[ $* != *getactivescheme* ]] || echo 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c
EOF
cat > "$TEST_ROOT/preflight.sh" <<'EOF'
#!/bin/bash
printf 'STRICT=%s BENCH_REPO=%s EXPECT_COMMIT=%s\n' "${STRICT-}" "${BENCH_REPO-}" "${EXPECT_COMMIT-}" > "$1"
[[ $STRICT == 1 && $BENCH_REPO == "$TEST_ROOT/repo" && $EXPECT_COMMIT == "$EXPECT_TEST" ]]
EOF
cat > "$TEST_ROOT/analyze.sh" <<'EOF'
#!/bin/bash
out=$1/analysis
printf '%s\n' "$*" >> "$TEST_ROOT/analyze.calls"
while (( $# )); do
    if [[ $1 == --out ]]; then out=$2; shift; fi
    shift
done
mkdir -p "$out"
printf 'analysis result\n' > "$out/sample.txt"
EOF
cat > "$TEST_ROOT/bench_exp.sh" <<'EOF'
#!/bin/bash
printf '%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n' \
    "${OUT_DIR##*/}" "$ARMS" "$REPS" "$JOBS" "$CLIENTS" "$SESSION_BUILDS" \
    "$BUILDER_CPUS" "$BUILDER_MEM" "$SEED_BASE" "$WORKLOAD" "$P90_THRESHOLD" \
    "$DRIFT" "$SYN_SEED" "$SYN_FILES" "$SYN_C_FRACTION" "$SYN_ALPHA" \
    "$SYN_MEDIAN_UNITS" "$SYN_MAX_UNITS" "$SYN_HIDDEN" "$ABSL_REF" "$ABSL_COMMIT" \
    "${DRY_RUN-<unset>}" "${DRIFT_AT-<unset>}" "${SYN_MANIFEST_SHA256-<unset>}" \
    "$ROUND_START" "${ALPHA-<unset>}" "${MIN_TASKS-<unset>}" "${OUT_DIR}" "${CONFIGURE_FLAGS-<unset>}" >> "$TEST_ROOT/runner.env"
count=$(cat "$TEST_ROOT/runner.count" 2>/dev/null || printf 0)
count=$((count + 1)); printf '%s\n' "$count" > "$TEST_ROOT/runner.count"
rc=0
if [[ -f $TEST_ROOT/runner.codes ]]; then
    rc=$(sed -n "${count}p" "$TEST_ROOT/runner.codes"); rc=${rc:-0}
fi
python3 - "$OUT_DIR" "$ROUND_START" "$REPS" "$ARMS" "$JOBS" "$CLIENTS" "$SESSION_BUILDS" "$DRIFT" "$BUILDER_CPUS" "$BUILDER_MEM" "$SEED_BASE" "$WORKLOAD" "$P90_THRESHOLD" "$ABSL_REF" "$ABSL_COMMIT" "$SYN_SEED" "$SYN_FILES" "$SYN_C_FRACTION" "$SYN_ALPHA" "$SYN_MEDIAN_UNITS" "$SYN_MAX_UNITS" "$SYN_HIDDEN" "$EXPECT_TEST" "$rc" <<'PY'
import csv, json, os, sys
out, start, reps, arms, jobs, clients, sb, drift, bcpu, bmem, seed, workload, p90, absl_ref, absl_commit, syn_seed, syn_files, syn_fraction, syn_alpha, syn_median, syn_max, syn_hidden, commit, rc = sys.argv[1:]
start, reps, clients, sb, rc = map(int, (start, reps, clients, sb, rc))
os.makedirs(out, exist_ok=True)
meta = dict(ARMS=arms, JOBS=jobs, CLIENTS=str(clients), SESSION_BUILDS=str(sb),
            DRIFT=drift, BUILDER_CPUS=bcpu, BUILDER_MEM=bmem, REPS=str(reps),
            SEED_BASE=seed, WORKLOAD=workload, P90_THRESHOLD=p90, REQUEST_TIMEOUT=os.environ.get("REQUEST_TIMEOUT", ""),
            ABSL_REF=absl_ref, ABSL_COMMIT=absl_commit, SYN_SEED=syn_seed,
            SYN_FILES=syn_files, SYN_C_FRACTION=syn_fraction, SYN_ALPHA=syn_alpha,
            SYN_MEDIAN_UNITS=syn_median, SYN_MAX_UNITS=syn_max, SYN_HIDDEN=syn_hidden,
            git_commit=commit)
with open(os.path.join(out, 'meta.json'), 'w') as f: json.dump(meta, f)
res = os.path.join(out, 'results.csv')
rows = []
if start > 1 and os.path.isfile(res):
    with open(res, newline='') as f: rows = [r for r in csv.DictReader(f) if int(r['round']) < start]
if rc == 0:
    for k in range(start, reps + 1):
        for a in arms.split():
            for b in range(1, sb + 1):
                for c in range(1, clients + 1):
                    rows.append(dict(arm=a, round=k, build_idx=b, client=c))
            with open(os.path.join(out, f'tasks_{a}_round{k}.jsonl'), 'w') as f: f.write('{}\n')
else:
    a = arms.split()[0]
    for name in (f'tasks_{a}_round{start}.jsonl', f'events_{a}_round{start}.csv',
                 f'stats_{a}_round{start}.csv', f'build_{a}_round{start}_b1_c1.log.status'):
        with open(os.path.join(out, name), 'w') as f: f.write('partial\n')
    if start == 1:
        with open(os.path.join(out, f'events_{a}_round0.csv'), 'w') as f: f.write('warmup\n')
    with open(os.path.join(out, 'failures.csv'), 'w') as f: f.write('old failure\n')
with open(res, 'w', newline='') as f:
    w = csv.DictWriter(f, fieldnames=('arm', 'round', 'build_idx', 'client'))
    w.writeheader(); w.writerows(rows)
with open(os.path.join(out, 'order_log.csv'), 'w') as f: f.write('round,order\n')
PY
[[ ${FAKE_MODE:-} != request_stop ]] || touch "$CAMPAIGN/STOP_REQUESTED"
exit "$rc"
EOF
chmod +x "$TEST_ROOT"/*.sh "$TEST_ROOT/powershell.exe"
sed -e 's#\. /d/bench-work/env.sh#. "$TEST_ROOT/env.sh"#' \
    -e 's#bash /d/bench-work/preflight.sh#bash "$TEST_ROOT/preflight.sh"#' \
    -e 's#bash scripts/bench_exp.sh#bash "$TEST_ROOT/bench_exp.sh"#' \
    -e 's#python3 scripts/analyze_exp.py#bash "$TEST_ROOT/analyze.sh"#g' \
    -e 's#powershell.exe#"$TEST_ROOT/powershell.exe"#g' \
    -e 's/^REPS=20 JOBS=/REPS=2 JOBS=/' \
    /d/bench-work/campaign_tn3.sh > "$TEST_ROOT/campaign.sh"

pass=0; fail=0
check() { if eval "$2"; then echo "PASS $1"; pass=$((pass+1)); else echo "FAIL $1"; fail=$((fail+1)); fi; }
run_campaign() { TN3_REPO="$TEST_ROOT/repo" EXPECT_COMMIT="$EXPECT_TEST" P90_SYN=150189 P90_ABSL=2356237 bash "$TEST_ROOT/campaign.sh" > "$TEST_ROOT/campaign.output" 2>&1; }
new_case() {
    CAMPAIGN="$TEST_ROOT/$1"; export CAMPAIGN
    mkdir -p "$CAMPAIGN"
    rm -f "$TEST_ROOT/runner.count" "$TEST_ROOT/runner.codes" "$TEST_ROOT/runner.env" "$TEST_ROOT/analyze.calls"
}

unset CAMPAIGN
STOP_AFTER=E3-syn run_campaign; rc=$?
check 'default output is official-tn3, isolated from TN1-2' '[[ $rc == 0 && -f $TEST_ROOT/out/official-tn3/E3-syn/analysis.done && ! -e $TEST_ROOT/out/official ]]'
CAMPAIGN="$TEST_ROOT/out/official"; export CAMPAIGN
run_campaign; rc=$?
check 'TN1-2 official directory override refused' '[[ $rc != 0 && ! -e $TEST_ROOT/out/official ]]'

new_case order
DRY_RUN=1 ALPHA=9 MIN_TASKS=999 DRIFT_AT=999 SYN_MANIFEST_SHA256=bogus CONFIGURE_FLAGS=bogus run_campaign; rc=$?
check 'order, two blocks, and two analysis families' '[[ $rc == 0 && $(wc -l < "$TEST_ROOT/runner.env") == 3 && $(wc -l < "$TEST_ROOT/analyze.calls") == 6 ]] && cut -d"|" -f1 "$TEST_ROOT/runner.env" | paste -sd, - | grep -Fxq "E3-syn,E3-absl,E3-syn-drift"'
check 'fixed runner inputs and cleared inherited inputs' 'awk -F"|" "{if (\$2 != \"leastloaded sed p90 sed-p90 linucb hybrid-linucb hybrid-linucb-d-arm-g098\" || \$3 != 2 || \$4 != 8 || \$5 != 1 || \$6 != 2 || \$7 != \"2.0\" || \$8 != \"2G\" || \$9 != 3000 || \$13 != 3001 || \$14 != 300 || \$15 != \"0.3\" || \$16 != \"1.5\" || \$17 != 120 || \$18 != 4500 || \$19 != \"0.0\" || \$20 != \"20250512.1\" || \$22 != \"<unset>\" || \$23 != \"<unset>\" || \$24 != \"<unset>\" || \$26 != \"<unset>\" || \$27 != \"<unset>\" || \$29 != \"<unset>\") exit 1} END {if (NR != 3) exit 1}" "$TEST_ROOT/runner.env"'
check 'workload, P90, drift and absl commit per scenario' 'awk -F"|" "NR==1 && (\$10!=\"syn\" || \$11!=150189 || \$12!=\"none\") {exit 1} NR==2 && (\$10!=\"absl\" || \$11!=2356237 || \$12!=\"none\") {exit 1} NR==3 && (\$10!=\"syn\" || \$11!=150189 || \$12!=\"permanent\") {exit 1} {if (\$21!=\"76bb24329e8bf5f39704eb10d21b9a80befa7c81\") exit 1} END {if (NR!=3) exit 1}" "$TEST_ROOT/runner.env" && grep -q "p90_syn=150189 p90_absl=2356237" "$CAMPAIGN/campaign.log"'
check 'STRICT preflight uses TN3 clone for every scenario' '[[ $(find "$CAMPAIGN" -name preflight_r1.txt | wc -l) == 3 ]] && ! grep -L "STRICT=1 BENCH_REPO=$TEST_ROOT/repo EXPECT_COMMIT=$EXPECT_TEST" "$CAMPAIGN"/E3-*/preflight_r1.txt | grep -q .'
check 'both analysis families and argv are marked' 'python3 - "$CAMPAIGN" <<"PY"
import json, pathlib, sys
root=pathlib.Path(sys.argv[1]); scenarios=("E3-syn","E3-absl","E3-syn-drift")
for name in scenarios:
    m=json.loads((root/name/"analysis.done").read_text())
    assert m["families"]==["analysis","analysis_p90_vs_ll"]
    assert m["argv"]["analysis"]=="--strict"
    assert m["argv"]["analysis_p90_vs_ll"]=="--strict --treatments p90 sed-p90 --baselines leastloaded --out analysis_p90_vs_ll"
PY'
run_campaign; rc=$?
check 'complete rerun skips runner and analysis' '[[ $rc == 0 && $(wc -l < "$TEST_ROOT/runner.env") == 3 && $(wc -l < "$TEST_ROOT/analyze.calls") == 6 ]]'

new_case missing_p90
env -u P90_SYN TN3_REPO="$TEST_ROOT/repo" EXPECT_COMMIT="$EXPECT_TEST" P90_ABSL=2356237 bash "$TEST_ROOT/campaign.sh" > "$TEST_ROOT/campaign.output" 2>&1; rc=$?
check 'missing P90_SYN refused before runner' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]]'
new_case wrong_commit
TN3_REPO="$TEST_ROOT/repo" EXPECT_COMMIT=0000000 P90_SYN=150189 P90_ABSL=2356237 bash "$TEST_ROOT/campaign.sh" > "$TEST_ROOT/campaign.output" 2>&1; rc=$?
check 'EXPECT_COMMIT mismatch refused before runner' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]] && grep -q "does not match" "$TEST_ROOT/campaign.output"'
new_case wrong_threshold
TN3_REPO="$TEST_ROOT/repo" EXPECT_COMMIT="$EXPECT_TEST" P90_SYN=150000 P90_ABSL=2356237 bash "$TEST_ROOT/campaign.sh" > "$TEST_ROOT/campaign.output" 2>&1; rc=$?
check 'unregistered P90 threshold refused before runner' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]] && grep -q "registered thresholds" "$TEST_ROOT/campaign.output"'
new_case wrong_repo
TN3_REPO="$TEST_ROOT/missing" EXPECT_COMMIT="$EXPECT_TEST" P90_SYN=1 P90_ABSL=1 bash "$TEST_ROOT/campaign.sh" > "$TEST_ROOT/campaign.output" 2>&1; rc=$?
check 'missing TN3_REPO refused before runner' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]]'

new_case mismatch
STOP_AFTER=E3-syn run_campaign
python3 - "$CAMPAIGN/E3-syn/meta.json" <<'PY'
import json,sys
p=sys.argv[1]; m=json.load(open(p)); m['WORKLOAD']='absl'; json.dump(m,open(p,'w'))
PY
run_campaign; rc=$?
check 'MISMATCH on WORKLOAD change' '[[ $rc != 0 && $(cat "$TEST_ROOT/runner.count") == 1 ]] && grep -q "MISMATCH\|directory holds another configuration" "$CAMPAIGN/campaign.log"'

new_case mismatch_p90
STOP_AFTER=E3-syn run_campaign
python3 - "$CAMPAIGN/E3-syn/meta.json" <<'PY'
import json,sys
p=sys.argv[1]; m=json.load(open(p)); m['P90_THRESHOLD']='999'; json.dump(m,open(p,'w'))
PY
run_campaign; rc=$?
check 'MISMATCH on P90_THRESHOLD change' '[[ $rc != 0 && $(cat "$TEST_ROOT/runner.count") == 1 ]] && grep -q "P90_THRESHOLD" "$CAMPAIGN/campaign.log"'

new_case reruns
printf '1\n1\n1\n' > "$TEST_ROOT/runner.codes"
STOP_AFTER=E3-syn run_campaign; rc=$?
ledger=$CAMPAIGN/E3-syn/reruns.csv
check 'two retries followed by cap, append-only ledger' '[[ $rc != 0 && $(wc -l < "$ledger") == 4 && $(cat "$TEST_ROOT/runner.count") == 3 ]] && grep -q ",2,3$" "$ledger"'
check 'partial block and warmup archived on rerun' '[[ $(find "$CAMPAIGN/E3-syn/archive" -mindepth 1 -maxdepth 1 -type d | wc -l) == 2 ]] && find "$CAMPAIGN/E3-syn/archive" -name "events_*_round0.csv" | grep -q . && find "$CAMPAIGN/E3-syn/archive" -name "preflight_r1.txt" | grep -q .'
new_case stop
touch "$CAMPAIGN/STOP_REQUESTED"
STOP_AFTER=E3-syn run_campaign; rc=$?
check 'STOP_REQUESTED before start makes no runner call' '[[ $rc == 0 && ! -f $TEST_ROOT/runner.count ]]'
rm -f "$CAMPAIGN/STOP_REQUESTED"
printf '1\n' > "$TEST_ROOT/runner.codes"
FAKE_MODE=request_stop STOP_AFTER=E3-syn run_campaign; rc=$?
check 'stop after runner records interruption, no abort ledger' '[[ $rc == 0 && ! -f $CAMPAIGN/E3-syn/reruns.csv ]] && grep -q "manual stop" "$CAMPAIGN/E3-syn/interruptions.csv" && [[ -f $CAMPAIGN/E3-syn/.pending_archive ]]'
rm -f "$CAMPAIGN/STOP_REQUESTED" "$TEST_ROOT/runner.codes"
STOP_AFTER=E3-syn run_campaign; rc=$?
check 'resume archives interrupted block and completes it' '[[ $rc == 0 && -f $CAMPAIGN/E3-syn/analysis.done && ! -f $CAMPAIGN/E3-syn/.pending_archive ]] && find "$CAMPAIGN/E3-syn/archive" -name "tasks_leastloaded_round1.jsonl" | grep -q .'

echo "TN3 campaign tests: $pass passed, $fail failed"
(( fail == 0 ))
