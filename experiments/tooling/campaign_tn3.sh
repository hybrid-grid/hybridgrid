#!/bin/bash
# Official Experiment 3 campaign. Run only after hypotheses-tn3.md is frozen
# in a clean LF clone at TN3_REPO and both pilot P90 thresholds are fixed.
#
# Usage (Git Bash):
#   TN3_REPO=/d/bench-work/hybridgrid-tn3 EXPECT_COMMIT=<frozen-sha> \
#     P90_SYN=<positive-byte-threshold> P90_ABSL=<positive-byte-threshold> \
#     bash /d/bench-work/campaign_tn3.sh
# Supervisor (separate process):
#   CAMPAIGN=/d/bench-work/out/official-tn3 nohup bash /d/bench-work/supervisor_tn3.sh &
# Request a clean stop with: touch /d/bench-work/out/official-tn3/STOP_REQUESTED
# (or the CAMPAIGN override). Delete that file before resuming. The TN1-2
# stop_campaign.sh checks for a process named campaign.sh and is not a TN3 stop tool.
# STRICT preflight runs with EXPECT_BRANCH=feat/exp3 (the frozen TN3 clone's branch).
# The scenario order is fixed. STOP_AFTER ends after a scenario. An incomplete
# block is rerun as a whole, with its partial files archived first.
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
    DRY_RUN EST_CELL_S ARMS REPS ROUND_START JOBS CLIENTS SESSION_BUILDS DRIFT OUT_DIR \
    WORKLOAD SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS \
    SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 ABSL_REF ABSL_COMMIT \
    BUILDER_CPUS BUILDER_MEM P90_THRESHOLD REQUEST_TIMEOUT
: "${TN3_REPO:?set TN3_REPO to the clean LF clone of the frozen TN3 commit}"
: "${EXPECT_COMMIT:?set EXPECT_COMMIT to the frozen TN3 commit prefix}"
: "${P90_SYN:?set P90_SYN to the positive integer syn pilot threshold}"
: "${P90_ABSL:?set P90_ABSL to the positive integer absl pilot threshold}"
[[ $EXPECT_COMMIT =~ ^[a-fA-F0-9]{7,40}$ ]] || { echo 'EXPECT_COMMIT must be a Git hex prefix (7-40 characters)' >&2; exit 2; }
# The P90 thresholds are registered in hypotheses-tn3.md; any other value would
# run a different experiment, so the operator must pass exactly these.
REGISTERED_P90_SYN=150189 REGISTERED_P90_ABSL=2356237
[[ $P90_SYN == "$REGISTERED_P90_SYN" && $P90_ABSL == "$REGISTERED_P90_ABSL" ]] || {
    echo "P90_SYN/P90_ABSL must equal the registered thresholds $REGISTERED_P90_SYN/$REGISTERED_P90_ABSL" >&2; exit 2; }
for threshold in "$P90_SYN" "$P90_ABSL"; do
    [[ $threshold =~ ^[0-9]+$ ]] && [[ $threshold =~ [1-9] ]] || { echo 'P90_SYN and P90_ABSL must be positive integers' >&2; exit 2; }
done
[[ -d $TN3_REPO/.git ]] || { echo "TN3_REPO is not a Git clone: $TN3_REPO" >&2; exit 2; }
clone_head=$(git -C "$TN3_REPO" rev-parse HEAD 2>/dev/null) || { echo 'cannot read TN3_REPO HEAD' >&2; exit 2; }
[[ $clone_head == "$EXPECT_COMMIT"* ]] || { echo "TN3_REPO HEAD $clone_head does not match EXPECT_COMMIT=$EXPECT_COMMIT" >&2; exit 2; }
export BENCH_REPO=$TN3_REPO
CAMPAIGN=${CAMPAIGN:-$BENCH_OUT/official-tn3}
[[ ${CAMPAIGN%/} != "$BENCH_OUT/official" && $CAMPAIGN != "$BENCH_OUT/official/"* ]] || {
    echo 'CAMPAIGN must not point into the TN1-2 official directory' >&2; exit 2;
}
REPS=20 JOBS=8 CLIENTS=1 SESSION_BUILDS=2 BUILDER_CPUS=2.0 BUILDER_MEM=2G SEED_BASE=3000
ARMS='leastloaded sed p90 sed-p90 linucb hybrid-linucb hybrid-linucb-d-arm-g098'
SYN_SEED=3001 SYN_FILES=300 SYN_C_FRACTION=0.3 SYN_ALPHA=1.5
SYN_MEDIAN_UNITS=120 SYN_MAX_UNITS=4500 SYN_HIDDEN=0.0
ABSL_REF=20250512.1 ABSL_COMMIT=76bb24329e8bf5f39704eb10d21b9a80befa7c81
# Pilot: heavy TUs reached 113 s against the coordinator's 120 s default when
# P90 stacked them on worker-5; a timeout would fail cells for some arms only.
REQUEST_TIMEOUT=600s
mkdir -p "$CAMPAIGN"
LOG=$CAMPAIGN/campaign.log
log_reset='CONFIGURE_FLAGS CPYTHON_VERSION COOLDOWN SEED_BASE ALPHA EPSILON WARM_START DRIFT_AT DRIFT_TARGET DRIFT_CPU DRIFT_DURATION DRIFT_ON DRIFT_OFF MIN_TASKS DRY_RUN EST_CELL_S ARMS REPS ROUND_START JOBS CLIENTS SESSION_BUILDS DRIFT OUT_DIR WORKLOAD SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 ABSL_REF ABSL_COMMIT BUILDER_CPUS BUILDER_MEM P90_THRESHOLD REQUEST_TIMEOUT'

ORDER=(E3-syn E3-absl E3-syn-drift)

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
    S_DRIFT=none
    case $1 in
        E3-syn|E3-syn-drift)
            S_WORKLOAD=syn; S_P90=$P90_SYN
            [[ $1 != E3-syn-drift ]] || S_DRIFT=permanent ;;
        E3-absl)
            S_WORKLOAD=absl; S_P90=$P90_ABSL
            ;;
        *) return 1 ;;
    esac
}

# Prints the first incomplete block (1..REPS), REPS+1 when all are complete,
# or "MISMATCH <reason>" when the directory belongs to another configuration.
next_round() {
    python3 - "$1" "$REPS" "$ARMS" "$JOBS" "$CLIENTS" "$SESSION_BUILDS" "$S_DRIFT" \
        "$BUILDER_CPUS" "$BUILDER_MEM" "$SEED_BASE" "$S_WORKLOAD" "$S_P90" \
        "$ABSL_REF" "$ABSL_COMMIT" "$SYN_SEED" "$SYN_FILES" \
        "$SYN_C_FRACTION" "$SYN_ALPHA" "$SYN_MEDIAN_UNITS" \
        "$SYN_MAX_UNITS" "$SYN_HIDDEN" "$EXPECT_COMMIT" "$REQUEST_TIMEOUT" <<'PY'
import csv, json, os, sys
out, reps, arms, jobs, clients, sb, drift, bcpu, bmem, seed, workload, p90, absl_ref, absl_commit, syn_seed, syn_files, syn_c_fraction, syn_alpha, syn_median, syn_max, syn_hidden, commit, request_timeout = sys.argv[1:]
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
        "DRIFT": drift, "BUILDER_CPUS": bcpu, "BUILDER_MEM": bmem,
        "REPS": str(reps), "SEED_BASE": seed, "WORKLOAD": workload,
        "P90_THRESHOLD": p90, "REQUEST_TIMEOUT": request_timeout,
        "ABSL_REF": absl_ref, "ABSL_COMMIT": absl_commit,
        "SYN_SEED": syn_seed, "SYN_FILES": syn_files,
        "SYN_C_FRACTION": syn_c_fraction, "SYN_ALPHA": syn_alpha,
        "SYN_MEDIAN_UNITS": syn_median, "SYN_MAX_UNITS": syn_max,
        "SYN_HIDDEN": syn_hidden}
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
    python3 scripts/analyze_exp.py "$out" --strict --treatments p90 sed-p90 --baselines leastloaded \
        --out "$out/analysis_p90_vs_ll" > "$out/analysis_p90_vs_ll.log" 2>&1
    rc=$?
    log "  analysis P90 vs LL rc=$rc -> $out/analysis_p90_vs_ll"
    (( rc == 0 )) || stop "$(basename "$out"): P90-vs-LL analysis failed (see $out/analysis_p90_vs_ll.log)"
}

(( $# == 0 )) || { echo "campaign_tn3.sh takes no arguments: the scenario order is fixed (use STOP_AFTER)" >&2; exit 2; }
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
restore() { powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' off >> "$LOG" 2>&1; log "power plan restored"; }
trap 'restore; cleanup' EXIT
powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' on >> "$LOG" 2>&1 || stop "cannot switch the power plan to High performance"
powershell.exe -NoProfile -Command "powercfg /getactivescheme" 2>/dev/null | grep -q 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c \
    || stop "High performance is not the active power plan"

cd "$TN3_REPO" || stop "repo not found"
log "CAMPAIGN START commit=$EXPECT_COMMIT repo=$TN3_REPO builder_cpus=$BUILDER_CPUS builder_mem=$BUILDER_MEM p90_syn=$P90_SYN p90_absl=$P90_ABSL order=${ORDER[*]} stop_after=${STOP_AFTER:-<none>}"

for sc in "${ORDER[@]}"; do
    set_scenario "$sc"
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
        STRICT=1 BENCH_REPO="$TN3_REPO" EXPECT_BRANCH=feat/exp3 EXPECT_COMMIT=$EXPECT_COMMIT bash /d/bench-work/preflight.sh "$out/preflight_r$rs.txt" > /dev/null 2>&1 \
            || stop "$sc: preflight failed before block $rs (see $out/preflight_r$rs.txt)"
        log "$sc: start at block $rs (arms: $ARMS | jobs=$JOBS clients=$CLIENTS session_builds=$SESSION_BUILDS workload=$S_WORKLOAD p90=$S_P90 drift=$S_DRIFT)"
        stopped
        attempt_tmp=$out/.active_attempt.$$
        printf '%s,%s,%s\n' "$rs" "$(date '+%F %T')" "$$" > "$attempt_tmp" || stop "$sc: cannot write active attempt"
        mv -f "$attempt_tmp" "$out/.active_attempt" || stop "$sc: cannot activate attempt"
        ARMS=$ARMS REPS=$REPS ROUND_START=$rs JOBS=$JOBS CLIENTS=$CLIENTS SESSION_BUILDS=$SESSION_BUILDS \
            DRIFT=$S_DRIFT BUILDER_CPUS=$BUILDER_CPUS BUILDER_MEM=$BUILDER_MEM \
            SEED_BASE=$SEED_BASE WORKLOAD=$S_WORKLOAD P90_THRESHOLD=$S_P90 REQUEST_TIMEOUT=$REQUEST_TIMEOUT \
            SYN_SEED=$SYN_SEED SYN_FILES=$SYN_FILES SYN_C_FRACTION=$SYN_C_FRACTION \
            SYN_ALPHA=$SYN_ALPHA SYN_MEDIAN_UNITS=$SYN_MEDIAN_UNITS SYN_MAX_UNITS=$SYN_MAX_UNITS \
            SYN_HIDDEN=$SYN_HIDDEN ABSL_REF=$ABSL_REF ABSL_COMMIT=$ABSL_COMMIT OUT_DIR="$out" \
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
    required=(analysis analysis_p90_vs_ll)
    marker_argv=(--argv 'analysis=--strict'
        'analysis_p90_vs_ll=--strict --treatments p90 sed-p90 --baselines leastloaded --out analysis_p90_vs_ll')
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
