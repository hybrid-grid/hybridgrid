#!/bin/bash
# Self-contained orchestration tests; all outputs live under tmp-tests.
set -u
mkdir -p /d/bench-work/tmp-tests
TEST_ROOT=$(mktemp -d /d/bench-work/tmp-tests/campaign.XXXXXX) || exit 1
export TEST_ROOT
trap 'rm -rf -- "$TEST_ROOT"' EXIT
mkdir -p "$TEST_ROOT/repo/scripts" "$TEST_ROOT/repo/test/stress" "$TEST_ROOT/out"
printf 'fake analyzer version 1\n' > "$TEST_ROOT/repo/scripts/analyze_exp.py"
printf 'services: {}\n' > "$TEST_ROOT/repo/test/stress/docker-compose-exp.yml"
cat > "$TEST_ROOT/env.sh" <<'EOF'
export BENCH_REPO="$TEST_ROOT/repo"
export BENCH_OUT="$TEST_ROOT/out"
export PATH="/d/bench-work/bin:$PATH"
EOF
cat > "$TEST_ROOT/powershell.exe" <<'EOF'
#!/bin/bash
[[ $* != *getactivescheme* ]] || echo 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c
EOF
cat > "$TEST_ROOT/preflight.sh" <<'EOF'
#!/bin/bash
printf 'preflight evidence\n' > "$1"
if [[ -f $TEST_ROOT/preflight-fail-on-third ]]; then
    n=$(cat "$TEST_ROOT/preflight.count" 2>/dev/null || printf 0)
    n=$((n+1)); printf '%s\n' "$n" > "$TEST_ROOT/preflight.count"
    if (( n == 3 )); then rm -f "$TEST_ROOT/preflight-fail-on-third"; exit 1; fi
fi
exit 0
EOF
cat > "$TEST_ROOT/analyze.sh" <<'EOF'
#!/bin/bash
out=$1/analysis; shift
while (( $# )); do
    if [[ $1 == --out ]]; then out=$2; shift; fi
    shift
done
mkdir -p "$out"
printf 'analysis result\n' > "$out/sample.txt"
printf '%s\n' "$out" >> "$TEST_ROOT/analyze.calls"
EOF
cat > "$TEST_ROOT/bench_exp.sh" <<'EOF'
#!/bin/bash
printf 'SEED_BASE=%s DRY_RUN=%s\n' "${SEED_BASE-<unset>}" "${DRY_RUN-<unset>}" >> "$TEST_ROOT/runner.env"
if [[ ${FAKE_MODE:-} == sleep ]]; then
    trap 'exit 0' TERM
    mkdir -p "$OUT_DIR"
    printf 'partial\n' > "$OUT_DIR/tasks_leastloaded_round${ROUND_START}.jsonl"
    while :; do sleep 1; done
fi
count=$(cat "$TEST_ROOT/runner.count" 2>/dev/null || printf 0)
count=$((count + 1)); printf '%s\n' "$count" > "$TEST_ROOT/runner.count"
rc=0
if [[ -f $TEST_ROOT/runner.codes ]]; then
    rc=$(sed -n "${count}p" "$TEST_ROOT/runner.codes")
    rc=${rc:-0}
fi
python3 - "$OUT_DIR" "$ROUND_START" "$REPS" "$ARMS" "$JOBS" "$CLIENTS" "$SESSION_BUILDS" "$DRIFT" "$BUILDER_CPUS" "$rc" <<'PY'
import csv, json, os, sys
out, start, reps, arms, jobs, clients, sb, drift, bcpu, rc = sys.argv[1:]
start, reps, clients, sb, rc = map(int, (start, reps, clients, sb, rc))
os.makedirs(out, exist_ok=True)
meta = dict(ARMS=arms, JOBS=jobs, CLIENTS=str(clients), SESSION_BUILDS=str(sb),
            DRIFT=drift, BUILDER_CPUS=bcpu, REPS=str(reps), git_commit='1b8339d')
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
                 f'stats_{a}_round{start}.csv', f'build_{a}_round{start}_b1_c1.log.status',
                 f'tasks_{a}_round10.jsonl'):
        with open(os.path.join(out, name), 'w') as f: f.write('old\n')
    if start == 1:
        with open(os.path.join(out, f'events_{a}_round0.csv'), 'w') as f: f.write('warm\n')
    with open(os.path.join(out, 'failures.csv'), 'w') as f: f.write(f'old-failure-{rc}\n')
with open(res, 'w', newline='') as f:
    w = csv.DictWriter(f, fieldnames=('arm','round','build_idx','client'))
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
    -e 's#bash /d/bench-work/analysis_marker.sh#bash /d/bench-work/analysis_marker.sh#g' \
    -e 's#powershell.exe#"$TEST_ROOT/powershell.exe"#g' \
    -e 's/^REPS=20$/REPS=2/' \
    -e 's/^A4=.*/A4="leastloaded sed"/' \
    -e 's/^A7=.*/A7="leastloaded sed"/' \
    -e 's/^AL=.*/AL="leastloaded sed"/' \
    /d/bench-work/campaign.sh > "$TEST_ROOT/campaign.sh"
pass=0; fail=0
check() { if eval "$2"; then echo "PASS $1"; pass=$((pass+1)); else echo "FAIL $1"; fail=$((fail+1)); fi; }
run_campaign() { EXPECT_COMMIT=1b8339d BUILDER_CPUS=2.0 bash "$TEST_ROOT/campaign.sh" >/dev/null 2>&1; }
new_case() { CAMPAIGN="$TEST_ROOT/$1"; export CAMPAIGN; mkdir -p "$CAMPAIGN"; rm -f "$TEST_ROOT/runner.count" "$TEST_ROOT/runner.codes" "$TEST_ROOT/runner.env" "$TEST_ROOT/analyze.calls" "$TEST_ROOT/preflight.count"; }

new_case marker
run_campaign
first=$(wc -l < "$TEST_ROOT/analyze.calls")
run_campaign
second=$(wc -l < "$TEST_ROOT/analyze.calls")
rm -f "$CAMPAIGN/E1-j6/analysis/sample.txt"
run_campaign
third=$(wc -l < "$TEST_ROOT/analyze.calls")
check 'T4 five E1 markers skip and missing output reruns analysis' '[[ $first == 10 && $second == 10 && $third == 12 && -f $CAMPAIGN/E1-j6/analysis.done ]]'
cp "$TEST_ROOT/campaign.sh" "$TEST_ROOT/campaign.original.sh"
sed 's/analysis=--strict/analysis=--strict --changed/g' "$TEST_ROOT/campaign.original.sh" > "$TEST_ROOT/campaign.sh"
run_campaign
fourth=$(wc -l < "$TEST_ROOT/analyze.calls")
check 'T4 changed analyzer argv triggers re-analysis' '[[ $fourth == 22 ]]'
mv "$TEST_ROOT/campaign.original.sh" "$TEST_ROOT/campaign.sh"

new_case env_pinning
SEED_BASE=stray DRY_RUN=1 STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'runner inherited seed and dry-run settings reset' '[[ $rc == 0 ]] && grep -q "SEED_BASE=<unset> DRY_RUN=<unset>" "$TEST_ROOT/runner.env"'

new_case gate_missing
run_campaign
check 'T3 missing approval waits before E2' 'grep -q "WAITING FOR PHASE2_APPROVED" "$CAMPAIGN/campaign.log" && [[ ! -d $CAMPAIGN/E2-perm ]]'
for kind in duplicate bad empty; do
    new_case "gate_$kind"
    case $kind in duplicate) printf 'E2_LAMBDA=RUN\nE2_LAMBDA=RUN\n' > "$CAMPAIGN/PHASE2_APPROVED";; bad) printf 'E2_LAMBDA=MAYBE\n' > "$CAMPAIGN/PHASE2_APPROVED";; empty) : > "$CAMPAIGN/PHASE2_APPROVED";; esac
    run_campaign; rc=$?
    check "T3 $kind approval rejected before runner" '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]]'
done
new_case gate_skip
printf 'E2_LAMBDA=SKIP\r\n' > "$CAMPAIGN/PHASE2_APPROVED"
run_campaign; rc=$?
check 'T3 SKIP stops after onoff' '[[ $rc == 0 && -d $CAMPAIGN/E2-onoff && ! -d $CAMPAIGN/E2-lambda ]] && grep -q "STOP_AFTER reached after E2-onoff" "$CAMPAIGN/campaign.log"'
before_calls=$(cat "$TEST_ROOT/runner.count")
printf 'E2_LAMBDA=RUN\n' > "$CAMPAIGN/PHASE2_APPROVED"
run_campaign; rc=$?
check 'T3 recorded SKIP decision rejects later RUN' '[[ $rc != 0 && $(cat "$TEST_ROOT/runner.count") == "$before_calls" ]] && grep -q "decision changed after it was recorded" "$CAMPAIGN/campaign.log"'
new_case gate_conflict
printf 'E2_LAMBDA=RUN\n' > "$CAMPAIGN/PHASE2_APPROVED"
STOP_AFTER=E2-onoff run_campaign; rc=$?
check 'T3 RUN and STOP_AFTER conflict rejected' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]]'

new_case reruns
printf '1\n1\n1\n' > "$TEST_ROOT/runner.codes"
touch "$TEST_ROOT/preflight-fail-on-third"
STOP_AFTER=E1-j6 run_campaign; rc=$?
STOP_AFTER=E1-j6 run_campaign; rc2=$?
ledger=$CAMPAIGN/E1-j6/reruns.csv
archives=$(find "$CAMPAIGN/E1-j6/archive" -mindepth 1 -maxdepth 1 -type d | wc -l)
check 'T1 persistent ledger across restart and third abort cap' '[[ $rc != 0 && $rc2 != 0 && $(wc -l < "$ledger") == 4 ]] && grep -q ",2,3$" "$ledger"'
check 'T2 exact block archive plus warmup and old failures' '[[ $archives == 3 ]] && find "$CAMPAIGN/E1-j6/archive" -name "events_*_round0.csv" | grep -q . && find "$CAMPAIGN/E1-j6/archive" -name "build_*_round1_*.status" | grep -q . && find "$CAMPAIGN/E1-j6/archive" -name failures.csv | xargs grep -q old-failure && [[ -f $CAMPAIGN/E1-j6/tasks_leastloaded_round10.jsonl ]] && ! find "$CAMPAIGN/E1-j6/archive" -name "*round10*" | grep -q .'
check 'preflight report copied into partial-block archive' 'find "$CAMPAIGN/E1-j6/archive" -name preflight_r1.txt | grep -q .'

new_case abandoned
mkdir -p "$CAMPAIGN/E1-j6"
printf '1,2026-10-17 00:00:00,1234\n' > "$CAMPAIGN/E1-j6/.active_attempt"
printf 'partial\n' > "$CAMPAIGN/E1-j6/tasks_leastloaded_round1.jsonl"
printf 'old preflight\n' > "$CAMPAIGN/E1-j6/preflight_r1.txt"
STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'abandoned attempt recorded and archived without rerun ledger' '[[ $rc == 0 && ! -f $CAMPAIGN/E1-j6/reruns.csv && ! -f $CAMPAIGN/E1-j6/.active_attempt ]] && grep -q "2026-10-17 00:00:00,E1-j6,1,.*abandoned attempt (no runner exit observed)" "$CAMPAIGN/E1-j6/interruptions.csv" && find "$CAMPAIGN/E1-j6/archive" -name tasks_leastloaded_round1.jsonl | grep -q . && find "$CAMPAIGN/E1-j6/archive" -name preflight_r1.txt | xargs grep -q "old preflight"'

new_case damaged
mkdir -p "$CAMPAIGN/E1-j6"
printf 'later\n' > "$CAMPAIGN/E1-j6/tasks_leastloaded_round2.jsonl"
STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'damaged directory with later round refuses fresh start' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]] && grep -q "missing results.csv or meta.json" "$CAMPAIGN/campaign.log"'
new_case damaged_archive
mkdir -p "$CAMPAIGN/E1-j6/archive/old"
printf 'evidence\n' > "$CAMPAIGN/E1-j6/archive/old/runner.log"
STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'archive from earlier block-1 attempt allows block-1 recovery' '[[ $rc == 0 && -f $TEST_ROOT/runner.count && -f $CAMPAIGN/E1-j6/archive/old/runner.log ]]'

new_case stopped
touch "$CAMPAIGN/STOP_REQUESTED"
STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'T5 refuse start with STOP_REQUESTED' '[[ $rc == 0 && ! -f $TEST_ROOT/runner.count && ! -f $CAMPAIGN/E1-j6/reruns.csv ]]'
rm -f "$CAMPAIGN/STOP_REQUESTED"
printf '%s\n' "$$" > "$CAMPAIGN/campaign.pid"
STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'T5 live campaign PID refuses second instance' '[[ $rc != 0 && ! -f $TEST_ROOT/runner.count ]]'
rm -f "$CAMPAIGN/campaign.pid"
printf '1\n' > "$TEST_ROOT/runner.codes"
FAKE_MODE=request_stop STOP_AFTER=E1-j6 run_campaign; rc=$?
check 'T5 runner exit during manual stop has no ledger entry' '[[ $rc == 0 && ! -f $CAMPAIGN/E1-j6/reruns.csv ]]'
rm -f "$CAMPAIGN/STOP_REQUESTED" "$TEST_ROOT/runner.count" "$TEST_ROOT/runner.codes"
new_case manual_stop
cat > "$TEST_ROOT/docker" <<'EOF'
#!/bin/bash
printf '%s\n' "$*" >> "$TEST_ROOT/docker.calls"
exit 0
EOF
chmod +x "$TEST_ROOT/docker"
sed 's#\. /d/bench-work/env.sh#. "$TEST_ROOT/env.sh"#' /d/bench-work/stop_campaign.sh > "$TEST_ROOT/stop_campaign.sh"
FAKE_MODE=sleep STOP_AFTER=E1-j6 EXPECT_COMMIT=1b8339d BUILDER_CPUS=2.0 bash "$TEST_ROOT/campaign.sh" >/dev/null 2>&1 &
cpid=$!
for ((i=0;i<100;i++)); do [[ -f $CAMPAIGN/runner.pid ]] && break; sleep 0.1; done
for ((i=0;i<100;i++)); do [[ -f $CAMPAIGN/E1-j6/tasks_leastloaded_round1.jsonl ]] && break; sleep 0.1; done
DOCKER="$TEST_ROOT/docker" bash "$TEST_ROOT/stop_campaign.sh" 'test manual stop' > "$TEST_ROOT/stop.output" 2>&1
stop_rc=$?
wait "$cpid"; campaign_rc=$?
rm -f "$CAMPAIGN/STOP_REQUESTED"
STOP_AFTER=E1-j6 run_campaign; resume_rc=$?
check 'T5 manual stop exits zero, archives its own partial block, and verifies Docker' '[[ $stop_rc == 0 && $campaign_rc == 0 && $resume_rc == 0 && ! -f $CAMPAIGN/E1-j6/reruns.csv && ! -f $CAMPAIGN/campaign.pid ]] && find "$CAMPAIGN/E1-j6/archive" -name "tasks_*_round1.jsonl" | grep -q . && grep -q "manual stop" "$CAMPAIGN/E1-j6/interruptions.csv" && grep -q "OK: hgexp containers absent" "$TEST_ROOT/stop.output" && grep -q "down --remove-orphans" "$TEST_ROOT/docker.calls" && grep -q "ps -aq" "$TEST_ROOT/docker.calls"'

new_case stale_pid
mkdir -p "$TEST_ROOT/proc/$$"
printf 'unrelated process\0' > "$TEST_ROOT/proc/$$/cmdline"
printf '%s\n' "$$" > "$CAMPAIGN/campaign.pid"
printf '99999999\n' > "$CAMPAIGN/runner.pid"
begin=$SECONDS
DOCKER="$TEST_ROOT/docker" PROC_ROOT="$TEST_ROOT/proc" bash "$TEST_ROOT/stop_campaign.sh" 'stale PID test' > "$TEST_ROOT/stale.output" 2>&1
stale_rc=$?; elapsed=$((SECONDS-begin))
check 'T5 stale or foreign PIDs are removed without waiting' '[[ $stale_rc == 0 && $elapsed -lt 10 && ! -e $CAMPAIGN/campaign.pid && ! -e $CAMPAIGN/runner.pid ]]'

echo "campaign tests: $pass passed, $fail failed"
(( fail == 0 ))
