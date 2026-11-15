#!/bin/bash
set -u
mkdir -p /d/bench-work/tmp-tests
TEST_ROOT=$(mktemp -d /d/bench-work/tmp-tests/supervisor.XXXXXX) || exit 1
export TEST_ROOT CAMPAIGN="$TEST_ROOT/official"
trap 'rm -rf -- "$TEST_ROOT"' EXIT
mkdir -p "$CAMPAIGN"
cat > "$TEST_ROOT/env.sh" <<'EOF'
export BENCH_OUT="$TEST_ROOT/out"
EOF
sed 's#\. /d/bench-work/env.sh#. "$TEST_ROOT/env.sh"#' /d/bench-work/supervisor.sh > "$TEST_ROOT/supervisor.sh"
printf 'old CAMPAIGN DONE\nold detail\nold detail\n' > "$CAMPAIGN/campaign.log"
mkdir -p "$CAMPAIGN/supervisor"
printf 'existing audit\n' > "$CAMPAIGN/supervisor/audit_005.md"
cat > "$TEST_ROOT/timeout" <<'EOF'
#!/bin/bash
printf '%s\n' "$1" >> "$TEST_ROOT/timeout.calls"
shift
"$@"
EOF
cat > "$TEST_ROOT/codex-fake" <<'EOF'
#!/bin/bash
while (( $# )); do
    if [[ $1 == -o ]]; then out=$2; shift; fi
    shift
done
n=$(cat "$TEST_ROOT/calls" 2>/dev/null || printf 0)
n=$((n+1)); printf '%s\n' "$n" > "$TEST_ROOT/calls"
if (( n == 1 )); then printf 'SEVERITY: OK\n' > "$out"; else printf 'SEVERITY: ALERT\n' > "$out"; fi
EOF
cat > "$TEST_ROOT/notify-fake" <<'EOF'
#!/bin/bash
printf '%s\n' "$*" >> "$TEST_ROOT/notifications"
EOF
chmod +x "$TEST_ROOT/codex-fake" "$TEST_ROOT/notify-fake" "$TEST_ROOT/timeout"
export PATH="$TEST_ROOT:$PATH"
CODEX_CMD="$TEST_ROOT/codex-fake" NOTIFY_CMD="$TEST_ROOT/notify-fake" bash "$TEST_ROOT/supervisor.sh" 1 >/dev/null 2>&1 &
super_pid=$!
for ((i=0;i<100;i++)); do [[ -f $TEST_ROOT/calls ]] && break; sleep 0.1; done
for ((i=0;i<100;i++)); do [[ -f $CAMPAIGN/supervisor/audit_006.md ]] && break; sleep 0.1; done
for ((i=0;i<100;i++)); do grep -q 'audit 6:' "$CAMPAIGN/supervisor/supervisor.log" 2>/dev/null && break; sleep 0.1; done
pass=0; fail=0
check() { if eval "$2"; then echo "PASS $1"; pass=$((pass+1)); else echo "FAIL $1"; fail=$((fail+1)); fi; }
check 'T6 old terminal line does not end supervisor' 'kill -0 "$super_pid" 2>/dev/null && [[ -f $CAMPAIGN/supervisor/supervisor.lock ]]'
check 'T6 restart continues after audit 005 with timeout wrapper' '[[ -f $CAMPAIGN/supervisor/audit_006.md && $(cat "$CAMPAIGN/supervisor/audit_005.md") == "existing audit" ]] && grep -q "^2700$" "$TEST_ROOT/timeout.calls"'
CODEX_CMD="$TEST_ROOT/codex-fake" NOTIFY_CMD="$TEST_ROOT/notify-fake" bash "$TEST_ROOT/supervisor.sh" 1 >/dev/null 2>&1
lock_rc=$?
check 'T6 second supervisor refused by live lock' '[[ $lock_rc != 0 ]]'
printf 'new MANUAL STOP test\n' > "$CAMPAIGN/campaign.log"
for ((i=0;i<100;i++)); do kill -0 "$super_pid" 2>/dev/null || break; sleep 0.1; done
wait "$super_pid"; supervisor_rc=$?
terminal_audit=$(sed -n 's/.*terminal line observed after audit \([0-9][0-9]*\); one final audit/\1/p' "$CAMPAIGN/supervisor/supervisor.log")
final_audit=$CAMPAIGN/supervisor/audit_$(printf '%03d' "$((terminal_audit + 1))").md
extra_audit=$CAMPAIGN/supervisor/audit_$(printf '%03d' "$((terminal_audit + 2))").md
check 'T6 truncation reset sees terminal line and gets one final audit' '[[ $supervisor_rc == 0 && -f $final_audit && ! -e $extra_audit && ! -f $CAMPAIGN/supervisor/supervisor.lock ]] && grep -q "campaign.log shrank" "$CAMPAIGN/supervisor/supervisor.log"'
for ((i=0;i<20;i++)); do [[ -s $TEST_ROOT/notifications ]] && break; sleep 0.1; done
check 'T6 ALERT logs NOTIFY and calls notifier' 'grep -q "NOTIFY.*SEVERITY: ALERT" "$CAMPAIGN/supervisor/supervisor.log" && [[ -s $TEST_ROOT/notifications ]]'
echo "supervisor tests: $pass passed, $fail failed"
(( fail == 0 ))
