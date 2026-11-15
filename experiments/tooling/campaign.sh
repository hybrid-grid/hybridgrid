#!/bin/bash
# Official campaign for the advisor's Experiments 1-2, following section 2b of
# docs/thesis/hypotheses.md (fixed scenario order, 20 blocks each, resume from
# the interrupted block, never mix configurations).
#
# Usage (Git Bash):
#   EXPECT_COMMIT=<sha of the commit containing hypotheses.md> BUILDER_CPUS=<decision C> \
#     [STOP_AFTER=<scenario>] bash /d/bench-work/campaign.sh
# The scenario order is fixed (no arguments). Phase 2 (E2-*) additionally needs
# the file $CAMPAIGN/PHASE2_APPROVED, created by the operator after Experiment 1.
# STOP_AFTER ends the campaign cleanly after that scenario. The phase gate
# decides whether E2-lambda is run. Re-running continues where it stopped: complete
# scenarios are skipped, an incomplete one resumes at its first incomplete
# block, which is rerun as a whole (all arms, same seed and order).
#
# Stops the whole campaign (exit 1) when: the power plan cannot be switched,
# preflight (STRICT) fails, a scenario directory holds data from another
# configuration, the runner aborts more than twice in one
# scenario, or the analysis of a finished scenario reports dropped blocks.
# The power plan is switched to High performance and always restored.

set -u
. /d/bench-work/env.sh
unset CONFIGURE_FLAGS CPYTHON_VERSION COOLDOWN SEED_BASE ALPHA EPSILON WARM_START \
    DRIFT_AT DRIFT_TARGET DRIFT_CPU DRIFT_DURATION DRIFT_ON DRIFT_OFF MIN_TASKS \
    DRY_RUN EST_CELL_S ARMS REPS ROUND_START JOBS CLIENTS SESSION_BUILDS DRIFT OUT_DIR
: "${EXPECT_COMMIT:?set EXPECT_COMMIT to the commit that contains docs/thesis/hypotheses.md}"
: "${BUILDER_CPUS:?set BUILDER_CPUS (decision C)}"
CAMPAIGN=${CAMPAIGN:-$BENCH_OUT/official}
REPS=20
mkdir -p "$CAMPAIGN"
LOG=$CAMPAIGN/campaign.log
log_reset='CONFIGURE_FLAGS CPYTHON_VERSION COOLDOWN SEED_BASE ALPHA EPSILON WARM_START DRIFT_AT DRIFT_TARGET DRIFT_CPU DRIFT_DURATION DRIFT_ON DRIFT_OFF MIN_TASKS DRY_RUN EST_CELL_S ARMS REPS ROUND_START JOBS CLIENTS SESSION_BUILDS DRIFT OUT_DIR'

A4="leastloaded sed hybrid-linucb hybrid-linucb-d-g098"
A7="leastloaded sed hybrid-linucb hybrid-linucb-d-g095 hybrid-linucb-d-g098 hybrid-linucb-d-g099 hybrid-linucb-d-arm-g098"
AL="leastloaded sed hybrid-linucb hybrid-linucb-l0 hybrid-linucb-l025 hybrid-linucb-l100 hybrid-linucb-d-g098 hybrid-linucb-d-g098-l0 hybrid-linucb-d-g098-l025 hybrid-linucb-d-g098-l100"
D_ARMS="hybrid-linucb-d-g095 hybrid-linucb-d-g098 hybrid-linucb-d-g099 hybrid-linucb-d-arm-g098"
ORDER=(E1-j6 E1-j7 E1-j8 E1-j9 E1-2c E2-perm E2-trans E2-onoff E2-lambda)

log() { printf '%s %s\n' "$(date '+%F %T')" "$*" | tee -a "$LOG"; }
log "runner inherited inputs reset: $log_reset"
stop() { log "CAMPAIGN STOPPED: $*"; exit 1; }
interruption() {
    local block=$1 detail=$2 ts=${3:-$(date '+%F %T')} ledger=$out/interruptions.csv
    [[ -f $ledger ]] || printf 'ts,scenario,block,detected_at,detail\n' > "$ledger"
    printf '%s,%s,%s,%s,%s\n' "$ts" "$sc" "$block" "$(date '+%F %T')" "$detail" >> "$ledger"
}
# The runner covers blocks rs..REPS in one invocation, so the block that was
# running when it exited is the first incomplete one, not necessarily rs.
current_block() {
    local b
    b=$(next_round "$out")
    if [[ $b =~ ^[0-9]+$ ]] && (( b <= REPS )); then printf '%s\n' "$b"; else printf '%s\n' "${1:-$rs}"; fi
}
stopped() {
    if [[ -e $CAMPAIGN/STOP_REQUESTED ]]; then
        if [[ ${1:-} == after_runner ]]; then
            cur=$(current_block)
            interruption "$cur" 'manual stop'
            printf '%s\n' "$cur" > "$out/.pending_archive"
        fi
        log "stopped by STOP_REQUESTED (delete $CAMPAIGN/STOP_REQUESTED to resume)"
        exit 0
    fi
}
phase_decision() {
    local -a lines=()
    local value
    mapfile -t lines < "$CAMPAIGN/PHASE2_APPROVED" || stop 'cannot read PHASE2_APPROVED'
    (( ${#lines[@]} == 1 )) || stop 'PHASE2_APPROVED must contain exactly one E2_LAMBDA line'
    value=${lines[0]%$'\r'}
    [[ $value =~ ^E2_LAMBDA=(RUN|SKIP)$ ]] || stop 'PHASE2_APPROVED must be E2_LAMBDA=RUN or E2_LAMBDA=SKIP'
    PHASE2_LAMBDA=${BASH_REMATCH[1]}
    local decision=$CAMPAIGN/PHASE2_DECISION recorded
    if [[ -e $decision ]]; then
        recorded=$(cat "$decision" 2>/dev/null) || stop 'cannot read PHASE2_DECISION'
        [[ $recorded == "E2_LAMBDA=$PHASE2_LAMBDA" ]] || stop 'decision changed after it was recorded'
    fi
    if [[ $PHASE2_LAMBDA == SKIP ]]; then
        [[ ${STOP_AFTER:-} != E2-lambda ]] || stop 'STOP_AFTER=E2-lambda conflicts with E2_LAMBDA=SKIP'
        [[ -n ${STOP_AFTER:-} && ${STOP_AFTER:-} == E1-* ]] || STOP_AFTER=E2-onoff
    elif [[ ${STOP_AFTER:-} == E2-perm || ${STOP_AFTER:-} == E2-trans || ${STOP_AFTER:-} == E2-onoff ]]; then
        stop "STOP_AFTER=$STOP_AFTER conflicts with E2_LAMBDA=RUN"
    fi
    if [[ ! -e $decision ]]; then
        ( set -C; printf 'E2_LAMBDA=%s\n' "$PHASE2_LAMBDA" > "$decision" ) 2>/dev/null || true
    fi
    recorded=$(cat "$decision" 2>/dev/null) || stop 'cannot read PHASE2_DECISION'
    [[ $recorded == "E2_LAMBDA=$PHASE2_LAMBDA" ]] || stop 'decision changed after it was recorded'
    log "PHASE2_APPROVED E2_LAMBDA=$PHASE2_LAMBDA effective_stop_after=${STOP_AFTER:-<none>}"
}
rerun_count() {
    local ledger=$1
    if [[ -f $ledger ]]; then
        awk 'NR > 1 && NF { n++ } END { print n+0 }' "$ledger"
    else
        printf '0\n'
    fi
}
archive_partial() {
    local out=$1 round=$2 n=$3 file base dest r found=0
    local -a rounds=("$round")
    (( round != 1 )) || rounds+=(0)
    local -a files=()
    shopt -s nullglob
    for file in "$out"/*; do
        [[ -f $file ]] || continue
        base=${file##*/}
        for r in "${rounds[@]}"; do
            if [[ $base =~ ^(tasks_.*_round${r}\.jsonl|events_.*_round${r}\.csv|stats_.*_round${r}\.csv|build_.*_round${r}_.+)$ ]]; then
                files+=("$file"); found=1; break
            fi
        done
    done
    shopt -u nullglob
    [[ ! -f $out/.pending_archive ]] || [[ $(cat "$out/.pending_archive") != "$round" ]] || found=1
    [[ ! -f $out/preflight_r${round}.txt ]] || found=1
    (( found )) || return 0
    dest="$out/archive/rerun_${n}_$(date '+%Y%m%d-%H%M%S')"
    while [[ -e $dest ]]; do dest="${dest}_$RANDOM"; done
    mkdir -p "$dest" || stop "cannot create archive $dest"
    for file in results.csv order_log.csv failures.csv meta.json runner.log "preflight_r${round}.txt"; do
        [[ ! -f $out/$file ]] || cp -p "$out/$file" "$dest/" || stop "cannot copy $file to archive"
    done
    if (( ${#files[@]} )); then
        mv -- "${files[@]}" "$dest/" || stop "cannot move round $round files to archive"
    fi
    rm -f "$out/.pending_archive"
    log "$(basename "$out"): archived partial block $round to $dest"
}

set_scenario() {
    S_JOBS=8; S_CLIENTS=1; S_SB=1; S_DRIFT=none; S_D_FAMILY=0; S_SED_FAMILY=0
    case $1 in
        E1-j6|E1-j7|E1-j8|E1-j9) S_ARMS=$A4; S_JOBS=${1#E1-j}; S_SED_FAMILY=1 ;;
        E1-2c)     S_ARMS=$A4; S_CLIENTS=2; S_SED_FAMILY=1 ;;
        E2-perm)   S_ARMS=$A7; S_SB=2; S_DRIFT=permanent; S_D_FAMILY=1 ;;
        E2-trans)  S_ARMS=$A7; S_SB=2; S_DRIFT=transient; S_D_FAMILY=1 ;;
        E2-onoff)  S_ARMS=$A7; S_SB=2; S_DRIFT=onoff;     S_D_FAMILY=1 ;;
        E2-lambda) S_ARMS=$AL; S_SB=2; S_DRIFT=permanent ;;
        *) return 1 ;;
    esac
}

# Prints the first incomplete block (1..REPS), REPS+1 when all are complete,
# or "MISMATCH <reason>" when the directory belongs to another configuration.
next_round() {
    python3 - "$1" "$REPS" "$S_ARMS" "$S_JOBS" "$S_CLIENTS" "$S_SB" "$S_DRIFT" "$BUILDER_CPUS" "$EXPECT_COMMIT" <<'PY'
import csv, json, os, sys
out, reps, arms, jobs, clients, sb, drift, bcpu, commit = sys.argv[1:]
reps, clients, sb = int(reps), int(clients), int(sb)
res, meta = os.path.join(out, "results.csv"), os.path.join(out, "meta.json")
if not (os.path.isfile(res) and os.path.isfile(meta)):
    import re
    substantial = any(re.match(r'^tasks_.*_round([2-9]|[1-9][0-9]+)\.jsonl$', name)
                      for name in os.listdir(out))
    # archive/ alone is not damage: campaign creates it when a block-1 attempt
    # (warm-up included, before meta.json exists) is retried.
    if substantial:
        print('MISMATCH missing results.csv or meta.json with later-round artifacts'); sys.exit()
    print(1); sys.exit()
m = json.load(open(meta))
want = {"ARMS": arms, "JOBS": jobs, "CLIENTS": str(clients), "SESSION_BUILDS": str(sb),
        "DRIFT": drift, "BUILDER_CPUS": bcpu, "REPS": str(reps)}
bad = [f"{k}={m.get(k)!r} (want {v!r})" for k, v in want.items() if str(m.get(k)) != v]
if not str(m.get("git_commit", "")).startswith(commit):
    bad.append(f"git_commit={m.get('git_commit')!r} (want {commit}*)")
if bad:
    print("MISMATCH " + "; ".join(bad)); sys.exit()
# A block is complete only when every (arm, build_idx, client) result row is
# present exactly once and every arm's task log for that round is non-empty.
expected = {(a, b, c) for a in arms.split() for b in range(1, sb + 1) for c in range(1, clients + 1)}
seen = {}
with open(res, newline="") as f:
    for r in csv.DictReader(f):
        key = (r["arm"], int(r["build_idx"]), int(r["client"]))
        seen.setdefault(int(r["round"]), []).append(key)
for k in range(1, reps + 1):
    keys = seen.get(k, [])
    logs_ok = all(os.path.getsize(p) > 0 if os.path.isfile(p) else False
                  for p in (os.path.join(out, f"tasks_{a}_round{k}.jsonl") for a in arms.split()))
    if len(keys) != len(set(keys)) or set(keys) != expected or not logs_ok:
        print(k); sys.exit()
print(reps + 1)
PY
}

# --strict exits non-zero when any block is dropped; a scenario must have all
# 20 complete blocks, so that stops the campaign for inspection.
analyze() {
    local out=$1 rc
    python3 scripts/analyze_exp.py "$out" --strict > "$out/analysis.log" 2>&1
    rc=$?
    log "  analysis rc=$rc -> $out/analysis"
    (( rc == 0 )) || stop "$(basename "$out"): analysis reports dropped blocks or errors (see $out/analysis.log)"
    if (( S_SED_FAMILY )); then
        # H1.3: SED versus LeastLoaded is its own Holm family (hypotheses.md section 1)
        python3 scripts/analyze_exp.py "$out" --strict --treatments sed --baselines leastloaded             --out "$out/analysis_sed_vs_ll" > "$out/analysis_sed_vs_ll.log" 2>&1
        rc=$?
        log "  analysis SED vs LL rc=$rc -> $out/analysis_sed_vs_ll"
        (( rc == 0 )) || stop "$(basename "$out"): SED-vs-LL analysis failed (see $out/analysis_sed_vs_ll.log)"
    fi
    if (( S_D_FAMILY )); then
        # shellcheck disable=SC2086
        python3 scripts/analyze_exp.py "$out" --strict --treatments $D_ARMS --baselines hybrid-linucb \
            --out "$out/analysis_d_vs_hg" > "$out/analysis_d_vs_hg.log" 2>&1
        rc=$?
        log "  analysis D vs HG rc=$rc -> $out/analysis_d_vs_hg"
        (( rc == 0 )) || stop "$(basename "$out"): D-vs-HG analysis failed (see $out/analysis_d_vs_hg.log)"
    fi
}

(( $# == 0 )) || { echo "campaign.sh takes no arguments: the scenario order is fixed (use STOP_AFTER)" >&2; exit 2; }
if [[ -n ${STOP_AFTER:-} ]]; then set_scenario "$STOP_AFTER" || { echo "unknown STOP_AFTER=$STOP_AFTER" >&2; exit 2; }; fi
stopped
if [[ -e $CAMPAIGN/campaign.pid ]]; then
    old_pid=$(cat "$CAMPAIGN/campaign.pid" 2>/dev/null || true)
    if [[ $old_pid =~ ^[0-9]+$ ]] && kill -0 "$old_pid" 2>/dev/null; then
        echo "campaign already running (PID $old_pid)" >&2; exit 1
    fi
    rm -f "$CAMPAIGN/campaign.pid"
fi
( set -C; printf '%s\n' "$$" > "$CAMPAIGN/campaign.pid" ) 2>/dev/null || { echo 'campaign.pid lock exists' >&2; exit 1; }
cleanup() { rm -f "$CAMPAIGN/runner.pid" "$CAMPAIGN/campaign.pid"; }
trap cleanup EXIT
[[ ! -e $CAMPAIGN/PHASE2_APPROVED ]] || phase_decision

restore() { powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' off >> "$LOG" 2>&1; log "power plan restored"; }
trap 'restore; cleanup' EXIT
powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' on >> "$LOG" 2>&1 || stop "cannot switch the power plan to High performance"
powershell.exe -NoProfile -Command "powercfg /getactivescheme" 2>/dev/null | grep -q 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c \
    || stop "High performance is not the active power plan"

cd "$BENCH_REPO" || stop "repo not found"
log "CAMPAIGN START commit=$EXPECT_COMMIT builder_cpus=$BUILDER_CPUS order=${ORDER[*]} stop_after=${STOP_AFTER:-<none>}"

for sc in "${ORDER[@]}"; do
    set_scenario "$sc"
    # Phase gate (hypotheses.md section 2b): Experiment 2 starts only after the
    # operator has seen Experiment 1 and created PHASE2_APPROVED.
    if [[ $sc == E2-* && ! -e $CAMPAIGN/PHASE2_APPROVED ]]; then
        log "WAITING FOR PHASE2_APPROVED before $sc"
        exit 0
    fi
    if [[ $sc == E2-perm ]]; then phase_decision; fi
    out=$CAMPAIGN/$sc
    mkdir -p "$out"
    if [[ -e $out/.active_attempt ]]; then
        IFS=, read -r abandoned_block abandoned_ts abandoned_pid < "$out/.active_attempt"
        [[ $abandoned_block =~ ^[0-9]+$ ]] || stop "$sc: invalid .active_attempt; inspect $out"
        abandoned_block=$(current_block "$abandoned_block")
        interruption "$abandoned_block" 'abandoned attempt (no runner exit observed)' "$abandoned_ts"
        log "$sc: abandoned attempt at block $abandoned_block from $abandoned_ts (PID $abandoned_pid)"
        printf '%s\n' "$abandoned_block" > "$out/.pending_archive"
        rm -f "$out/.active_attempt"
    fi
    while :; do
        rs=$(next_round "$out")
        [[ $rs == MISMATCH* ]] && stop "$sc: directory holds another configuration: ${rs#MISMATCH }"
        [[ $rs =~ ^[0-9]+$ ]] || stop "$sc: cannot read progress ($rs)"
        if (( rs > REPS )); then log "$sc: all $REPS blocks complete"; break; fi
        stopped
        prior_aborts=$(rerun_count "$out/reruns.csv")
        (( prior_aborts <= 2 )) || stop "$sc: rerun ledger already exceeds the cap ($prior_aborts aborts)"
        archive_partial "$out" "$rs" "$(( $(rerun_count "$out/reruns.csv") + 1 ))"
        STRICT=1 EXPECT_COMMIT=$EXPECT_COMMIT bash /d/bench-work/preflight.sh "$out/preflight_r$rs.txt" > /dev/null 2>&1 \
            || stop "$sc: preflight failed before block $rs (see $out/preflight_r$rs.txt)"
        log "$sc: start at block $rs (arms: $S_ARMS | jobs=$S_JOBS clients=$S_CLIENTS session_builds=$S_SB drift=$S_DRIFT)"
        stopped
        attempt_tmp=$out/.active_attempt.$$
        printf '%s,%s,%s\n' "$rs" "$(date '+%F %T')" "$$" > "$attempt_tmp" || stop "$sc: cannot write active attempt"
        mv -f "$attempt_tmp" "$out/.active_attempt" || stop "$sc: cannot activate attempt"
        ARMS=$S_ARMS REPS=$REPS ROUND_START=$rs JOBS=$S_JOBS CLIENTS=$S_CLIENTS SESSION_BUILDS=$S_SB \
            DRIFT=$S_DRIFT BUILDER_CPUS=$BUILDER_CPUS OUT_DIR="$out" \
            bash scripts/bench_exp.sh >> "$out/runner.log" 2>&1 &
        rp=$!
        printf '%s\n' "$rp" > "$CAMPAIGN/runner.pid"
        [[ ! -e $CAMPAIGN/STOP_REQUESTED ]] || kill -TERM "$rp" 2>/dev/null || true
        wait "$rp"
        rc=$?
        rm -f "$CAMPAIGN/runner.pid"
        rm -f "$out/.active_attempt"
        stopped after_runner
        log "$sc: runner exit $rc"
        if (( rc != 0 )); then
            ledger=$out/reruns.csv
            if [[ ! -f $ledger ]]; then printf 'ts,scenario,block,runner_rc,count_before,count_after\n' > "$ledger"; fi
            before=$(rerun_count "$ledger")
            after=$((before + 1))
            cur=$(current_block)
            printf '%s,%s,%s,%s,%s,%s\n' "$(date '+%F %T')" "$sc" "$cur" "$rc" "$before" "$after" >> "$ledger"
            printf '%s\n' "$cur" > "$out/.pending_archive"
            (( after <= 2 )) || stop "$sc: runner aborted $after times (see $out/runner.log, $out/failures.csv)"
            log "$sc: resume $after/2"
        fi
    done
    required=(analysis)
    marker_argv=(--argv 'analysis=--strict')
    (( S_SED_FAMILY == 0 )) || required+=(analysis_sed_vs_ll)
    (( S_D_FAMILY == 0 )) || required+=(analysis_d_vs_hg)
    (( S_SED_FAMILY == 0 )) || marker_argv+=("analysis_sed_vs_ll=--strict --treatments sed --baselines leastloaded --out analysis_sed_vs_ll")
    (( S_D_FAMILY == 0 )) || marker_argv+=("analysis_d_vs_hg=--strict --treatments $D_ARMS --baselines hybrid-linucb --out analysis_d_vs_hg")
    if bash /d/bench-work/analysis_marker.sh check "$out" "${required[@]}" "${marker_argv[@]}"; then
        log "$sc: analysis up to date"
    else
        rm -f "$out/analysis.done"
        analyze "$out"
        bash /d/bench-work/analysis_marker.sh write "$out" "${required[@]}" "${marker_argv[@]}" || stop "$sc: cannot write analysis marker"
    fi
    if [[ ${STOP_AFTER:-} == "$sc" ]]; then log "STOP_AFTER reached after $sc"; exit 0; fi
done
log "CAMPAIGN DONE"
