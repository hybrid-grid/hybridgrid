#!/usr/bin/env bash
# Randomized complete-block build experiments. No containers are started in DRY_RUN.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STRESS_DIR="$ROOT/test/stress"
COMPOSE_FILE="$STRESS_DIR/docker-compose-exp.yml"
PROJECT=hgexp
ARMS="${ARMS:-leastloaded sed hybrid-linucb hybrid-linucb-d-g098}"
REPS="${REPS:-20}"
ROUND_START="${ROUND_START:-1}"
OUT_DIR="${OUT_DIR:-/tmp/bench_exp_$(date +%Y%m%d_%H%M%S)}"
JOBS="${JOBS:-8}"
CLIENTS="${CLIENTS:-1}"
BUILDER_CPUS="${BUILDER_CPUS:-1.0}"
BUILDER_MEM="${BUILDER_MEM:-1G}"
SESSION_BUILDS="${SESSION_BUILDS:-1}"
CONFIGURE_FLAGS="${CONFIGURE_FLAGS:---disable-test-modules --disable-perf-trampoline}"
CPYTHON_VERSION="${CPYTHON_VERSION:-v3.14.0}"
WORKLOAD="${WORKLOAD:-cpython}"
ABSL_REF="${ABSL_REF:-20250512.1}"
ABSL_COMMIT="${ABSL_COMMIT:-76bb24329e8bf5f39704eb10d21b9a80befa7c81}"
SYN_SEED="${SYN_SEED:-3001}"
SYN_FILES="${SYN_FILES:-300}"
SYN_C_FRACTION="${SYN_C_FRACTION:-0.3}"
SYN_ALPHA="${SYN_ALPHA:-1.5}"
SYN_MEDIAN_UNITS="${SYN_MEDIAN_UNITS:-40}"
SYN_MAX_UNITS="${SYN_MAX_UNITS:-1500}"
SYN_HIDDEN="${SYN_HIDDEN:-0.0}"
SYN_MANIFEST_SHA256=
COOLDOWN="${COOLDOWN:-8}"
SEED_BASE="${SEED_BASE:-1000}"
ALPHA="${ALPHA:-0.5}"
EPSILON="${EPSILON:-0.1}"
WARM_START="${WARM_START:-100}"
P90_THRESHOLD="${P90_THRESHOLD:-}"
# Coordinator per-compile deadline (e.g. 600s). Empty keeps the built-in 120s
# and passes no flag, so Experiments 1-2 commands are unchanged.
REQUEST_TIMEOUT="${REQUEST_TIMEOUT:-}"
DRIFT="${DRIFT:-none}"
DRIFT_AT="${DRIFT_AT:-120}"
DRIFT_TARGET="${DRIFT_TARGET:-worker-5}"
DRIFT_CPU="${DRIFT_CPU:-1}"
DRIFT_DURATION="${DRIFT_DURATION:-30}"
DRIFT_ON="${DRIFT_ON:-10}"
DRIFT_OFF="${DRIFT_OFF:-10}"
MIN_TASKS="${MIN_TASKS:-}"
DRY_RUN="${DRY_RUN:-0}"
EST_CELL_S="${EST_CELL_S:-60}"

die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
positive_int() { [[ $2 =~ ^[0-9]+$ ]] && (( 10#$2 > 0 )) || die "$1 must be a positive integer"; }
nonnegative_int() { [[ $2 =~ ^[0-9]+$ ]] || die "$1 must be a nonnegative integer"; }
decimal() { [[ $2 =~ ^[0-9]+([.][0-9]+)?$ ]] || die "$1 must be a nonnegative decimal"; }
now() { python3 -c 'import time; print(time.time())'; }
compose() { docker compose -p "$PROJECT" -f "$COMPOSE_FILE" "$@"; }

positive_int REPS "$REPS"; positive_int ROUND_START "$ROUND_START"
positive_int JOBS "$JOBS"; positive_int SESSION_BUILDS "$SESSION_BUILDS"
nonnegative_int WARM_START "$WARM_START"; positive_int DRIFT_CPU "$DRIFT_CPU"
positive_int DRIFT_DURATION "$DRIFT_DURATION"; positive_int DRIFT_ON "$DRIFT_ON"
positive_int DRIFT_OFF "$DRIFT_OFF"; positive_int EST_CELL_S "$EST_CELL_S"
nonnegative_int COOLDOWN "$COOLDOWN"; positive_int DRIFT_AT "$DRIFT_AT"
(( DRIFT_AT > 100 )) || die 'DRIFT_AT must be >= 101'
(( ROUND_START <= REPS )) || die 'ROUND_START must not exceed REPS'
[[ $CLIENTS == 1 || $CLIENTS == 2 ]] || die 'CLIENTS must be 1 or 2'
[[ $DRIFT == none || $DRIFT == permanent || $DRIFT == transient || $DRIFT == onoff ]] || die 'DRIFT must be none, permanent, transient, or onoff'
[[ -z $REQUEST_TIMEOUT || $REQUEST_TIMEOUT =~ ^[1-9][0-9]*(s|m)$ ]] || die 'REQUEST_TIMEOUT must look like 600s or 10m'
# The client-side compile deadline (HG_TIMEOUT for hgcc/hgcxx) must outlast the
# coordinator deadline; otherwise the client cancels and retries long compiles
# while the worker is still running them (pilot 2026-10-22).
HG_TIMEOUT_ENV=()
if [[ -n $REQUEST_TIMEOUT ]]; then
    case $REQUEST_TIMEOUT in *m) rt_s=$(( ${REQUEST_TIMEOUT%m} * 60 )) ;; *) rt_s=${REQUEST_TIMEOUT%s} ;; esac
    HG_TIMEOUT_ENV=("HG_TIMEOUT=$(( rt_s + 60 ))s")
fi
[[ $DRIFT_TARGET =~ ^worker-[1-5]$ ]] || die 'DRIFT_TARGET must be worker-1..worker-5'
[[ $DRY_RUN == 0 || $DRY_RUN == 1 ]] || die 'DRY_RUN must be 0 or 1'
decimal ALPHA "$ALPHA"; decimal EPSILON "$EPSILON"; decimal BUILDER_CPUS "$BUILDER_CPUS"
python3 -c 'import sys; a,e=map(float,sys.argv[1:]); sys.exit(not (0 <= a <= 10 and 0 <= e <= 1))' "$ALPHA" "$EPSILON" || die 'ALPHA must be in [0, 10] and EPSILON in [0, 1]'
python3 -c 'import sys; sys.exit(float(sys.argv[1]) <= 0)' "$BUILDER_CPUS" || die 'BUILDER_CPUS must be > 0'
[[ -z $MIN_TASKS ]] || positive_int MIN_TASKS "$MIN_TASKS"
[[ $SEED_BASE =~ ^-?[0-9]+$ ]] || die 'SEED_BASE must be an integer'
[[ $CPYTHON_VERSION =~ ^[a-zA-Z0-9._-]+$ ]] || die 'invalid CPYTHON_VERSION'
[[ $CONFIGURE_FLAGS =~ ^[-a-zA-Z0-9_=./\ ]+$ ]] || die 'CONFIGURE_FLAGS contains unsupported characters'
[[ $WORKLOAD == cpython || $WORKLOAD == absl || $WORKLOAD == syn ]] || die 'WORKLOAD must be cpython, absl, or syn'
[[ $ABSL_REF =~ ^[a-zA-Z0-9._-]+$ ]] || die 'invalid ABSL_REF'
[[ $ABSL_COMMIT =~ ^[a-fA-F0-9]{40}$ ]] || die 'ABSL_COMMIT must be a 40-character hex commit'
[[ $BUILDER_MEM =~ ^[1-9][0-9]*([kKmMgG][bB]?)?$ ]] || die 'invalid BUILDER_MEM'
nonnegative_int SYN_SEED "$SYN_SEED"; positive_int SYN_FILES "$SYN_FILES"
positive_int SYN_MEDIAN_UNITS "$SYN_MEDIAN_UNITS"; positive_int SYN_MAX_UNITS "$SYN_MAX_UNITS"
decimal SYN_C_FRACTION "$SYN_C_FRACTION"; decimal SYN_ALPHA "$SYN_ALPHA"; decimal SYN_HIDDEN "$SYN_HIDDEN"
python3 -c 'import math,sys; c,a,h=map(float,sys.argv[1:]); sys.exit(not (0 <= c <= 1 and math.isfinite(a) and a > 0 and 0 <= h <= 1))' "$SYN_C_FRACTION" "$SYN_ALPHA" "$SYN_HIDDEN" || die 'SYN_C_FRACTION and SYN_HIDDEN must be in [0, 1], SYN_ALPHA must be > 0'
(( 10#$SYN_MAX_UNITS >= 10#$SYN_MEDIAN_UNITS )) || die 'SYN_MAX_UNITS must be >= SYN_MEDIAN_UNITS'

declare -a arm_list=() arm_bases=() arm_flags=()
declare -A seen=()
read -r -a requested_arms <<< "$ARMS"
(( ${#requested_arms[@]} > 0 )) || die 'ARMS is empty'
for label in "${requested_arms[@]}"; do
    if [[ ! $label =~ ^(leastloaded|simple|p2c|epsilon-greedy|linucb|hybrid-linucb-d|hybrid-linucb|heft|icecc-fastest|sed-p90|sed|p90)(-arm)?(-g[0-9]{3})?(-l(0|[0-9]{3}))?$ ]]; then
        die "invalid arm '$label': expected <base>[-arm][-g<ddd>][-l<ddd>]"
    fi
    base=${BASH_REMATCH[1]}; mode=${BASH_REMATCH[2]}; gamma=${BASH_REMATCH[3]}; penalty=${BASH_REMATCH[4]}
    if [[ $base == p90 || $base == sed-p90 ]]; then
        [[ -n $P90_THRESHOLD ]] || die "P90_THRESHOLD is required for arm '$label'"
        positive_int P90_THRESHOLD "$P90_THRESHOLD"
    fi
    [[ -z ${seen[$label]+x} ]] || die "duplicate arm '$label'"
    seen[$label]=1
    if [[ -n $mode || -n $gamma ]] && [[ $base != hybrid-linucb-d ]]; then
        die "arm '$label': -arm and -g require hybrid-linucb-d"
    fi
    if [[ -n $penalty && $base != hybrid-linucb && $base != hybrid-linucb-d ]]; then
        die "arm '$label': -l requires hybrid-linucb or hybrid-linucb-d"
    fi
    discount=0.98; discount_mode=global; load_penalty=0.5
    [[ -z $mode ]] || discount_mode=arm
    if [[ -n $gamma ]]; then
        digits=${gamma#-g}; (( 10#$digits > 0 && 10#$digits <= 100 )) || die "arm '$label': discount must be in (0, 1]"
        discount=$(printf '%d.%02d' "$((10#$digits / 100))" "$((10#$digits % 100))")
    fi
    if [[ -n $penalty ]]; then
        digits=${penalty#-l}; (( 10#$digits <= 500 )) || die "arm '$label': load penalty exceeds 5.0"
        load_penalty=$(printf '%d.%02d' "$((10#$digits / 100))" "$((10#$digits % 100))")
    fi
    flags="--scheduler=$base --alpha=$ALPHA --epsilon=$EPSILON --warm-start=$WARM_START --load-penalty=$load_penalty"
    [[ $base != p90 && $base != sed-p90 ]] || flags+=" --size-threshold=$P90_THRESHOLD"
    [[ $base != hybrid-linucb-d ]] || flags+=" --discount=$discount --discount-mode=$discount_mode"
    [[ -z $REQUEST_TIMEOUT ]] || flags+=" --request-timeout=$REQUEST_TIMEOUT"
    flags+=' --task-log=/tmp/tasks.jsonl'
    arm_list+=("$label"); arm_bases+=("$base"); arm_flags+=("$flags")
done

order_for_round() {
    SEED_BASE="$SEED_BASE" ARMS="$ARMS" ROUND="$1" python3 - <<'PY'
import os, random
a = os.environ['ARMS'].split()
random.Random(int(os.environ['SEED_BASE']) + int(os.environ['ROUND'])).shuffle(a)
print(' '.join(a))
PY
}

generate_compose() {
    local flags=$1 i cpus mem parallel
    {
        cat <<EOF
services:
  coordinator:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-coord serve --grpc-port=9000 --http-port=8080 $flags
    volumes:
      - task-logs:/tmp
    ports:
      - "19000:9000"
      - "18080:8080"
    networks: [hgnet]
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M
EOF
        for i in 1 2 3 4 5; do
            case $i in
                1) cpus=0.5; mem=512M; parallel=1 ;;
                2) cpus=0.6; mem=614M; parallel=2 ;;
                3) cpus=0.8; mem=820M; parallel=2 ;;
                4) cpus=1.0; mem=1024M; parallel=2 ;;
                5) cpus=1.1; mem=1126M; parallel=3 ;;
            esac
            cat <<EOF
  worker-$i:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-$i:50051 --max-parallel=$parallel
    depends_on: [coordinator]
    networks: [hgnet]
    deploy:
      resources:
        limits:
          cpus: '$cpus'
          memory: $mem
EOF
        done
        for i in $(seq 1 "$CLIENTS"); do
            service=builder; cache=build-cache; source=cpython-src
            if (( i == 2 )); then service=builder2; cache=build-cache2; source=cpython-src2; fi
            cat <<EOF
  $service:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: sleep infinity
    environment:
      HG_COORDINATOR: coordinator:9000
    depends_on: [coordinator]
    networks: [hgnet]
    volumes:
      - $cache:/root/.hybridgrid/cache
      - $source:/workspace
      - gate:/gate
    deploy:
      resources:
        limits:
          cpus: '$BUILDER_CPUS'
          memory: $BUILDER_MEM
EOF
        done
        printf 'networks:\n  hgnet:\n    driver: bridge\nvolumes:\n  task-logs:\n  build-cache:\n  cpython-src:\n  gate:\n'
        if (( CLIENTS == 2 )); then printf '  build-cache2:\n  cpython-src2:\n'; fi
    } > "$COMPOSE_FILE"
}

if [[ $DRY_RUN == 1 ]]; then
    generate_compose "${arm_flags[0]}"
    printf 'compose: %s\n' "$COMPOSE_FILE"
    for i in "${!arm_list[@]}"; do printf 'arm %s: hg-coord %s\n' "${arm_list[$i]}" "${arm_flags[$i]}"; done
    for (( round=1; round<=REPS; round++ )); do
        if (( round < ROUND_START )); then
            printf 'round %d (already completed): %s\n' "$round" "$(order_for_round "$round")"
        else
            printf 'round %d: %s\n' "$round" "$(order_for_round "$round")"
        fi
    done
    cells=$(( (REPS - ROUND_START + 1) * ${#arm_list[@]} + (ROUND_START == 1 ? 1 : 0) ))
    printf 'estimated wall time: %d s (%d cells x %d builds x %d s, plus cooldown; excludes setup)\n' "$((cells * (EST_CELL_S * SESSION_BUILDS + COOLDOWN)))" "$cells" "$SESSION_BUILDS" "$EST_CELL_S"
    case $WORKLOAD in
        cpython)
            printf 'configure: cd /workspace/cpython && ./configure %s\n' "$CONFIGURE_FLAGS"
            printf 'build: cd /workspace/cpython && hgbuild make -j"$1"\n'
            ;;
        absl)
            printf 'configure: HG_COORDINATOR=127.0.0.1:1 CC=hgcc CXX=hgcxx cmake -S /workspace/abseil-cpp -B /workspace/abseil-cpp-build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17 -DABSL_BUILD_TESTING=OFF -DABSL_PROPAGATE_CXX_STD=ON\n'
            printf 'build: cd /workspace/abseil-cpp-build && hgbuild ninja --no-fallback -j"$1"\n'
            ;;
        syn)
            printf 'generate: python3 scripts/gen_synthetic_cpp.py --out DIR --seed %s --files %s --c-fraction %s --pareto-alpha %s --median-units %s --max-units %s --hidden-cost-fraction %s (host)\n' "$SYN_SEED" "$SYN_FILES" "$SYN_C_FRACTION" "$SYN_ALPHA" "$SYN_MEDIAN_UNITS" "$SYN_MAX_UNITS" "$SYN_HIDDEN"
            printf 'configure: HG_COORDINATOR=127.0.0.1:1 CC=hgcc CXX=hgcxx cmake -S /workspace/syn -B /workspace/syn-build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17\n'
            printf 'build: cd /workspace/syn-build && hgbuild ninja --no-fallback -j"$1"\n'
            ;;
    esac
    printf 'meta keys: WORKLOAD ABSL_REF ABSL_COMMIT SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 BUILDER_MEM P90_THRESHOLD\n'
    exit 0
fi

command -v docker >/dev/null || die 'docker is required'
command -v curl >/dev/null || die 'curl is required'
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"
if (( ROUND_START == 1 )); then
    printf 'arm,round,order_pos,build_idx,client,elapsed_s,cell_makespan_s,start_ts,end_ts\n' > "$OUT_DIR/results.csv"
    printf 'round,order\n' > "$OUT_DIR/order_log.csv"
    printf 'arm,round,attempt,reason\n' > "$OUT_DIR/failures.csv"
else
    [[ -f $OUT_DIR/results.csv && -f $OUT_DIR/order_log.csv && -f $OUT_DIR/meta.json ]] || die 'resume requires results.csv, order_log.csv, and meta.json'
    # Never mix configurations in one directory: every setting that defines the
    # experiment must match meta.json. Keys absent from an older meta.json
    # (written before WORKLOAD existed) must still have their default value.
    python3 - "$OUT_DIR/meta.json" \
        "ARMS=$ARMS" "JOBS=$JOBS" "CLIENTS=$CLIENTS" "SESSION_BUILDS=$SESSION_BUILDS" "BUILDER_CPUS=$BUILDER_CPUS" \
        "DRIFT=$DRIFT" "DRIFT_AT=$DRIFT_AT" "DRIFT_TARGET=$DRIFT_TARGET" "SEED_BASE=$SEED_BASE" "WARM_START=$WARM_START" \
        "WORKLOAD=$WORKLOAD" "BUILDER_MEM=$BUILDER_MEM" "ABSL_REF=$ABSL_REF" "ABSL_COMMIT=$ABSL_COMMIT" \
        "SYN_SEED=$SYN_SEED" "SYN_FILES=$SYN_FILES" "SYN_C_FRACTION=$SYN_C_FRACTION" "SYN_ALPHA=$SYN_ALPHA" \
        "SYN_MEDIAN_UNITS=$SYN_MEDIAN_UNITS" "SYN_MAX_UNITS=$SYN_MAX_UNITS" "SYN_HIDDEN=$SYN_HIDDEN" \
        "P90_THRESHOLD=$P90_THRESHOLD" "REQUEST_TIMEOUT=$REQUEST_TIMEOUT" "ALPHA=$ALPHA" "EPSILON=$EPSILON" "CPYTHON_VERSION=$CPYTHON_VERSION" \
        "CONFIGURE_FLAGS=$CONFIGURE_FLAGS" "COOLDOWN=$COOLDOWN" "DRIFT_CPU=$DRIFT_CPU" \
        "DRIFT_DURATION=$DRIFT_DURATION" "DRIFT_ON=$DRIFT_ON" "DRIFT_OFF=$DRIFT_OFF" <<'PY' || die 'resume configuration differs from meta.json'
import json, sys
meta = json.load(open(sys.argv[1]))
defaults = {"WORKLOAD": "cpython", "BUILDER_MEM": "1G", "P90_THRESHOLD": ""}
bad = []
for item in sys.argv[2:]:
    key, value = item.split("=", 1)
    if key in meta:
        if str(meta[key]) != value:
            bad.append(f"{key}: meta.json={meta[key]!r}, now={value!r}")
    elif key in defaults and value != defaults[key]:
        bad.append(f"{key}: absent from meta.json (default {defaults[key]!r}), now={value!r}")
if bad:
    print("resume mismatch:\n  " + "\n  ".join(bad), file=sys.stderr)
    sys.exit(1)
PY
    if [[ -z $MIN_TASKS ]]; then MIN_TASKS=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["MIN_TASKS"])' "$OUT_DIR/meta.json"); fi
    # A crashed round may have some successful cells already. Rerun the whole block.
    python3 - "$OUT_DIR/results.csv" "$OUT_DIR/order_log.csv" "$ROUND_START" <<'PY'
import csv, os, sys
for path in sys.argv[1:3]:
    with open(path, newline='') as f:
        rows = list(csv.reader(f))
    round_col = rows[0].index('round')
    kept = [rows[0]] + [row for row in rows[1:] if int(row[round_col]) < int(sys.argv[3])]
    with open(path, 'w', newline='') as f:
        csv.writer(f, lineterminator='\n').writerows(kept)
PY
fi

injector_pid=; stats_pid=; cell_marker=; events_file=; syn_host_dir=; SYN_MARKER=
stop_stress() {
    local running=0 stopped=0 count
    if compose exec -T "$DRIFT_TARGET" pgrep -x stress-ng >/dev/null 2>&1; then
        running=1
    fi
    if compose exec -T "$DRIFT_TARGET" pkill -x stress-ng >/dev/null 2>&1; then stopped=1; fi
    if (( running == 1 )) && [[ $DRIFT != none ]]; then
        if (( stopped == 0 )); then
            [[ -z $cell_marker ]] || printf 'stress-ng cleanup failed\n' > "$cell_marker.error"
            return 1
        fi
        count=$(dispatch_count 2>/dev/null || printf '0')
        post_event drift_off "$count" 'cell_end' || { [[ -z $cell_marker ]] || printf 'drift_off POST failed\n' > "$cell_marker.error"; return 1; }
    fi
}
stop_background() {
    [[ -z $cell_marker ]] || touch "$cell_marker"
    if [[ -n $injector_pid ]]; then kill "$injector_pid" 2>/dev/null || true; wait "$injector_pid" 2>/dev/null || true; injector_pid=; fi
    if [[ -n $stats_pid ]]; then kill "$stats_pid" 2>/dev/null || true; wait "$stats_pid" 2>/dev/null || true; stats_pid=; fi
    stop_stress || true
}
cleanup() {
    stop_background
    compose down >/dev/null 2>&1 || true
    [[ -z $syn_host_dir ]] || rm -rf "$syn_host_dir"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

dispatch_count() {
    curl -fsS --max-time 2 http://localhost:18080/dispatch-count |
        python3 -c 'import json,sys; print(int(json.load(sys.stdin)["dispatches"]))'
}
post_event() {
    local kind=$1 count=$2 detail=$3
    python3 -c 'import json,sys; print(json.dumps(dict(kind=sys.argv[1], target=sys.argv[2], dispatch_count=int(sys.argv[3]), detail=sys.argv[4])))' \
        "$kind" "$DRIFT_TARGET:50051" "$count" "$detail" |
        curl -fsS --max-time 3 -X POST -H 'Content-Type: application/json' --data-binary @- http://localhost:18080/events >/dev/null || return 1
    printf '%s,%s,%s,%s,%s\n' "$(now)" "$kind" "$DRIFT_TARGET:50051" "$count" "$detail" >> "$events_file"
}
start_stress() {
    local count=$1 try
    compose exec -d -T "$DRIFT_TARGET" stress-ng --cpu "$DRIFT_CPU" || return 1
    for try in 1 2 3 4 5 6 7 8 9 10; do
        if compose exec -T "$DRIFT_TARGET" pgrep -x stress-ng >/dev/null 2>&1; then
            # Onset = counter once stress-ng is confirmed running; docker exec
            # takes ~0.5-1 s on Docker Desktop, so the trigger reading is early.
            post_event drift_on "$(dispatch_count 2>/dev/null || printf '%s' "$count")" "stress_started trigger=$count" || return 1
            return 0
        fi
        sleep 0.2
    done
    return 1
}
stop_stress_phase() {
    local count=$1
    if compose exec -T "$DRIFT_TARGET" pgrep -x stress-ng >/dev/null 2>&1; then
        compose exec -T "$DRIFT_TARGET" pkill -x stress-ng || return 1
        post_event drift_off "$(dispatch_count 2>/dev/null || printf '%s' "$count")" "stress_stopped trigger=$count" || return 1
    fi
}
inject_drift() {
    local count
    while [[ ! -e $cell_marker ]]; do
        count=$(dispatch_count 2>/dev/null || true)
        if [[ $count =~ ^[0-9]+$ ]] && (( count >= DRIFT_AT )); then break; fi
        sleep 0.2
    done
    [[ ! -e $cell_marker ]] || return 0
    if ! start_stress "$count"; then printf 'start_stress failed\n' > "$cell_marker.error"; return 1; fi
    case $DRIFT in
        permanent) while [[ ! -e $cell_marker ]]; do sleep 0.2; done ;;
        transient)
            if wait_build_seconds "$DRIFT_DURATION"; then
                stop_stress_phase "$(dispatch_count || printf '0')" || printf 'stop_stress failed\n' > "$cell_marker.error"
            fi
            ;;
        onoff)
            while [[ ! -e $cell_marker ]]; do
                wait_build_seconds "$DRIFT_ON" || break
                stop_stress_phase "$(dispatch_count || printf '0')" || { printf 'stop_stress failed\n' > "$cell_marker.error"; return 1; }
                wait_build_seconds "$DRIFT_OFF" || break
                count=$(dispatch_count || printf '0')
                start_stress "$count" || { printf 'start_stress failed\n' > "$cell_marker.error"; return 1; }
            done
            ;;
    esac
}

wait_build_seconds() {
    local ticks=0 goal=$(( $1 * 5 ))
    while (( ticks < goal )); do
        [[ ! -e $cell_marker ]] || return 1
        if [[ -e $cell_marker.gap ]]; then sleep 0.2; continue; fi
        sleep 0.2
        [[ -e $cell_marker.gap ]] || ticks=$((ticks + 1))
    done
    while [[ -e $cell_marker.gap ]]; do
        [[ ! -e $cell_marker ]] || return 1
        sleep 0.2
    done
    [[ ! -e $cell_marker ]]
}

sample_stats() {
    local service cid sample name cpu mem
    local -a ids=()
    for service in coordinator builder builder2; do
        [[ $service != builder2 || $CLIENTS == 2 ]] || continue
        cid=$(compose ps -q "$service" 2>/dev/null || true)
        [[ -z $cid ]] || ids+=("$cid")
    done
    if (( ${#ids[@]} != CLIENTS + 1 )); then
        printf 'missing stats containers\n' > "$cell_marker.error"
        return 1
    fi
    while [[ ! -e $cell_marker ]]; do
        sample=$(docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' "${ids[@]}" 2>/dev/null || true)
        while IFS=, read -r name cpu mem; do
            [[ -n $name ]] || continue
            case $name in
                *coordinator*) service=coordinator ;;
                *builder2*) service=builder2 ;;
                *builder*) service=builder ;;
                *) continue ;;
            esac
            printf '%s,%s,%s,%s,%s\n' "$(now)" "$build_idx" "$service" "$cpu" "$mem" >> "$stats_file"
        done <<< "$sample"
        sleep 1
    done
}

prepare_builder() {
    local service=$1
    compose exec -T "$service" bash -s -- "$CPYTHON_VERSION" <<'SH'
set -euo pipefail
version=$1
if [[ -d /workspace/cpython ]] && [[ $(cat /workspace/cpython/.hg-bench-version 2>/dev/null || true) != "$version" ]]; then
    rm -rf /workspace/cpython
fi
if [[ ! -d /workspace/cpython ]]; then
    git clone --depth=1 --branch="$version" https://github.com/python/cpython.git /workspace/cpython
    printf '%s\n' "$version" > /workspace/cpython/.hg-bench-version
fi
SH
    local -a configure_args
    read -r -a configure_args <<< "$CONFIGURE_FLAGS"
    compose exec -T "$service" bash -s -- "$CONFIGURE_FLAGS" "${configure_args[@]}" <<'SH'
set -euo pipefail
cd /workspace/cpython
want=$1; shift
have=$(cat .hg-configure-flags 2>/dev/null || true)
if [[ ! -f Makefile || $want != "$have" ]]; then
    make clean >/dev/null 2>&1 || true
    ./configure "$@" > /tmp/hg-configure.log 2>&1 || { tail -30 /tmp/hg-configure.log; exit 1; }
    printf '%s\n' "$want" > .hg-configure-flags
fi
SH
}
prepare_absl_builder() {
    local service=$1
    compose exec -T "$service" bash -s -- "$ABSL_REF" "$ABSL_COMMIT" <<'SH'
set -euo pipefail
ref=$1; expected=$2
source=/workspace/abseil-cpp
marker="$ref $expected"
if [[ -d $source ]] && { [[ $(cat "$source/.hg-bench-source" 2>/dev/null || true) != "$marker" ]] || [[ $(git -C "$source" rev-parse HEAD 2>/dev/null || true) != "$expected" ]]; }; then
    rm -rf "$source" /workspace/abseil-cpp-build
fi
if [[ ! -d $source ]]; then
    git clone --depth=1 --branch="$ref" https://github.com/abseil/abseil-cpp.git "$source"
    actual=$(git -C "$source" rev-parse HEAD)
    [[ $actual == "$expected" ]] || { printf 'Abseil commit mismatch: expected %s, got %s\n' "$expected" "$actual" >&2; exit 1; }
    printf '%s\n' "$marker" > "$source/.hg-bench-source"
fi
SH
    configure_cmake "$service" abseil-cpp -DABSL_BUILD_TESTING=OFF -DABSL_PROPAGATE_CXX_STD=ON
}
prepare_syn_builder() {
    local service=$1 have
    # Trust the marker only if the tree it describes is actually present (an
    # interrupted copy must not leave a valid-looking marker behind).
    have=$(compose exec -T "$service" sh -c 'test -f /workspace/syn/CMakeLists.txt && test -f /workspace/syn/manifest.json && cat /workspace/syn/.hg-bench-source 2>/dev/null || true') || return 1
    if [[ $have != "$SYN_MARKER" ]]; then
        compose exec -T "$service" sh -c 'rm -rf /workspace/syn /workspace/syn-build && mkdir -p /workspace/syn' || return 1
        # Stream the tree with tar: `docker compose cp` cannot tell a Windows
        # host path (C:\...) from SERVICE:PATH and fails with "unknown copy direction".
        tar -C "$syn_host_dir/syn" -cf - . | compose exec -T "$service" sh -c 'tar -C /workspace/syn -xf -' || return 1
        compose exec -T "$service" sh -c 'test -f /workspace/syn/CMakeLists.txt' || return 1
        compose exec -T "$service" sh -c 'printf "%s\n" "$1" > /workspace/syn/.hg-bench-source' _ "$SYN_MARKER" || return 1
    fi
    configure_cmake "$service" syn
}
# Pass the source directory name relative to /workspace: Git Bash rewrites
# arguments that start with "/" before they reach docker compose exec.
configure_cmake() {
    local service=$1 source=$2
    shift 2
    compose exec -T "$service" bash -s -- "$source" "$@" <<'SH'
set -euo pipefail
source=/workspace/$1; shift
build="${source}-build"
want=$(cat "$source/.hg-bench-source")
have=$(cat "$build/.hg-bench-configured-source" 2>/dev/null || true)
if [[ ! -f $build/build.ninja || $want != "$have" ]]; then
    # Compiler probes during configure must not reach the measured coordinator
    # (they would train the scheduler and advance the dispatch counter), so
    # point hgcc at an unreachable address: the probes fall back to local gcc.
    HG_COORDINATOR=127.0.0.1:1 CC=hgcc CXX=hgcxx cmake -S "$source" -B "$build" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17 "$@" > /tmp/hg-configure.log 2>&1 || { tail -30 /tmp/hg-configure.log; exit 1; }
    printf '%s\n' "$want" > "$build/.hg-bench-configured-source"
fi
SH
}
prepare_workload_builder() {
    case $WORKLOAD in
        cpython) prepare_builder "$1" ;;
        absl) prepare_absl_builder "$1" ;;
        syn) prepare_syn_builder "$1" ;;
    esac
}
wait_registration() {
    local n waited
    for (( waited=0; waited<=90; waited+=2 )); do
        n=$(compose logs coordinator 2>&1 | grep -c 'Worker registered' || true)
        (( n >= 5 )) && return 0
        sleep 2
    done
    FAIL_REASON="only $n/5 workers registered"
    return 1
}
clean_builders() {
    local service
    for service in builder builder2; do
        [[ $service != builder2 || $CLIENTS == 2 ]] || continue
        if [[ $WORKLOAD == cpython ]]; then
            compose exec -T "$service" bash -c 'cd /workspace/cpython && make clean >/dev/null 2>&1 && rm -rf /root/.hybridgrid/cache/*' || return 1
        else
            # Relative to /workspace: Git Bash would rewrite a leading "/".
            local build=abseil-cpp-build
            [[ $WORKLOAD != syn ]] || build=syn-build
            compose exec -T "$service" bash -c 'ninja -C "/workspace/$1" clean >/dev/null 2>&1 && rm -rf /root/.hybridgrid/cache/*' _ "$build" || return 1
        fi
    done
}
collect_worker_limits() {
    local i values
    printf 'worker,cpu_max,nproc\n' > "$OUT_DIR/.worker_limits.csv"
    for i in 1 2 3 4 5; do
        values=$(compose exec -T "worker-$i" sh -c 'cat /sys/fs/cgroup/cpu.max; nproc') || return 1
        printf 'worker-%s,%s,%s\n' "$i" "$(printf '%s\n' "$values" | head -1)" "$(printf '%s\n' "$values" | tail -1)" >> "$OUT_DIR/.worker_limits.csv"
    done
}
write_meta() {
    local image_ids docker_info git_commit git_dirty host_nproc service
    image_ids=
    for service in coordinator worker-1 worker-2 worker-3 worker-4 worker-5 builder; do
        image_ids+="$service=$(docker inspect --format '{{.Image}}' "$(compose ps -q "$service")")|"
    done
    if (( CLIENTS == 2 )); then image_ids+="builder2=$(docker inspect --format '{{.Image}}' "$(compose ps -q builder2)")|"; fi
    docker_info=$(docker info --format '{{.NCPU}},{{.MemTotal}}')
    git_commit=$(git -C "$ROOT" rev-parse HEAD)
    git_dirty=false
    [[ -z $(git -C "$ROOT" status --porcelain) ]] || git_dirty=true
    host_nproc=$(nproc)
    export ARMS REPS ROUND_START OUT_DIR JOBS CLIENTS BUILDER_CPUS SESSION_BUILDS CONFIGURE_FLAGS CPYTHON_VERSION COOLDOWN SEED_BASE ALPHA EPSILON WARM_START DRIFT DRIFT_AT DRIFT_TARGET DRIFT_CPU DRIFT_DURATION DRIFT_ON DRIFT_OFF MIN_TASKS DRY_RUN EST_CELL_S
    export WORKLOAD ABSL_REF ABSL_COMMIT SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 BUILDER_MEM P90_THRESHOLD REQUEST_TIMEOUT
    META_IMAGES="$image_ids" META_DOCKER="$docker_info" META_COMMIT="$git_commit" META_DIRTY="$git_dirty" META_NPROC="$host_nproc" \
        META_LABELS="$(printf '%s|' "${arm_list[@]}")" META_FLAGS="$(printf '%s|' "${arm_flags[@]}")" \
        python3 - "$OUT_DIR/.worker_limits.csv" "$OUT_DIR/meta.json" <<'PY'
import csv, json, os, sys
e = os.environ
keys = 'ARMS REPS ROUND_START OUT_DIR JOBS CLIENTS BUILDER_CPUS SESSION_BUILDS CONFIGURE_FLAGS CPYTHON_VERSION COOLDOWN SEED_BASE ALPHA EPSILON WARM_START DRIFT DRIFT_AT DRIFT_TARGET DRIFT_CPU DRIFT_DURATION DRIFT_ON DRIFT_OFF MIN_TASKS DRY_RUN EST_CELL_S WORKLOAD ABSL_REF ABSL_COMMIT SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 BUILDER_MEM P90_THRESHOLD REQUEST_TIMEOUT'.split()
meta = {key: e[key] for key in keys}
labels = e['META_LABELS'].rstrip('|').split('|')
flags = e['META_FLAGS'].rstrip('|').split('|')
images = dict(item.split('=', 1) for item in e['META_IMAGES'].rstrip('|').split('|'))
meta.update(git_commit=e['META_COMMIT'], git_dirty=e['META_DIRTY'] == 'true', image_ids=images, host={'nproc': int(e['META_NPROC']), 'docker_NCPU': int(e['META_DOCKER'].split(',')[0]), 'docker_MemTotal': int(e['META_DOCKER'].split(',')[1])}, resolved_arms=[{'label': a, 'flags': b} for a,b in zip(labels, flags)])
with open(sys.argv[1], newline='') as f:
    meta['workers'] = list(csv.DictReader(f))
with open(sys.argv[2], 'w') as f:
    json.dump(meta, f, indent=2)
PY
}

FAIL_REASON=
run_cell() {
    local arm=$1 round=$2 pos=$3 record=$4 flags=$5 attempt=$6
    local i service status start end elapsed makespan tasks_path counts successes failures first_on post_completed stats_before stats_after
    local token ready_entries ready_all client log id pid time_tag build_status min_start max_end
    FAIL_REASON=; stop_background
    compose down >/dev/null 2>&1 || true
    generate_compose "$flags"
    local -a services=(coordinator builder worker-1 worker-2 worker-3 worker-4 worker-5)
    if (( CLIENTS == 2 )); then services+=(builder2); fi
    if ! compose up -d "${services[@]}"; then FAIL_REASON='compose up failed'; return 1; fi
    compose exec -T coordinator sh -c 'truncate -s 0 /tmp/tasks.jsonl' || { FAIL_REASON='task log truncate failed'; return 1; }
    if [[ $WORKLOAD != cpython ]]; then wait_registration || return 1; fi
    for i in $(seq 1 "$CLIENTS"); do
        service=builder; (( i == 1 )) || service=builder2
        prepare_workload_builder "$service" || { FAIL_REASON="prepare $service failed"; return 1; }
    done
    if [[ $WORKLOAD == cpython ]]; then
        wait_registration || return 1
    else
        # CMake's compiler probes use hgcc, so discard them before the cell.
        compose exec -T coordinator sh -c 'truncate -s 0 /tmp/tasks.jsonl' || { FAIL_REASON='task log truncate failed'; return 1; }
    fi
    if [[ ! -f $OUT_DIR/.worker_limits.csv ]]; then collect_worker_limits || { FAIL_REASON='worker limits failed'; return 1; }; fi
    printf 'cell %s round %s attempt %s: all 5 workers registered\n' "$arm" "$round" "$attempt"
    sleep "$COOLDOWN"
    local tmp_results="$OUT_DIR/.cell_results.csv"
    : > "$tmp_results"
    events_file="$OUT_DIR/events_${arm}_round${round}.csv"
    stats_file="$OUT_DIR/stats_${arm}_round${round}.csv"
    printf 'ts,kind,target,dispatch_count,detail\n' > "$events_file"
    printf 'ts,build_idx,service,cpu_percent,mem_usage\n' > "$stats_file"
    cell_marker="$OUT_DIR/.cell_${arm}_${round}.done"
    rm -f "$cell_marker" "$cell_marker.error" "$cell_marker.gap"
    for (( build_idx=1; build_idx<=SESSION_BUILDS; build_idx++ )); do
        clean_builders || { FAIL_REASON='make clean or cache clear failed'; return 1; }
        token="${arm}-r${round}-b${build_idx}-a${attempt}"
        compose exec -T builder sh -c 'rm -f /gate/*' || { FAIL_REASON="gate clear failed for build $build_idx"; return 1; }
        declare -a pids=()
        for (( client=1; client<=CLIENTS; client++ )); do
            service=builder; (( client == 1 )) || service=builder2
            id="${arm}-r${round}-b${build_idx}-c${client}"
            log="$OUT_DIR/build_${arm}_round${round}_b${build_idx}_c${client}.log"
            (
                set +e
                compose exec -T "$service" env "HG_BUILD_ID=$id" "${HG_TIMEOUT_ENV[@]}" bash -c '
                    token=$2; client=$3
                    touch "/gate/ready-c${client}-${token}" || exit 1
                    deadline=$((SECONDS + 120))
                    while (( SECONDS < deadline )); do
                        [[ ! -e /gate/go-${token} ]] || break
                        sleep 0.01
                    done
                    [[ -e /gate/go-${token} ]] || { printf "gate timeout\n" >&2; exit 1; }
                    if [[ $4 == cpython ]]; then
                        cd /workspace/cpython || exit 1
                    elif [[ $4 == absl ]]; then
                        cd /workspace/abseil-cpp-build || exit 1
                    else
                        cd /workspace/syn-build || exit 1
                    fi
                    start=$(date +%s.%N)
                    # CMake workloads must never compile a measured TU locally:
                    # --no-fallback makes a failed remote compile fail the build.
                    if [[ $4 == cpython ]]; then hgbuild make -j"$1"; else hgbuild ninja --no-fallback -j"$1"; fi
                    result=$?
                    end=$(date +%s.%N)
                    printf "HGTIME %s %s %s %s\n" "$start" "$end" "$result" "$(hostname)"
                    exit "$result"
                ' _ "$JOBS" "$token" "$client" "$WORKLOAD" > "$log" 2>&1
                status=$?
                printf '%s\n' "$status" > "$log.status"
            ) &
            pids+=("$!")
        done
        ready_all=0
        local gate_deadline=$((SECONDS + 60))
        while (( SECONDS < gate_deadline )); do
            # sh -c keeps Git Bash from rewriting /gate into a Windows path.
            ready_entries=$(compose exec -T builder sh -c 'ls -1 /gate' 2>/dev/null) || ready_entries=
            ready_all=1
            for (( client=1; client<=CLIENTS; client++ )); do
                if ! grep -Fxq "ready-c${client}-${token}" <<< "$ready_entries"; then ready_all=0; break; fi
            done
            (( ready_all == 0 )) || break
            sleep 0.01
        done
        (( ready_all == 1 )) || { FAIL_REASON="gate readiness timeout for build $build_idx: expected $CLIENTS clients"; return 1; }
        if (( build_idx == 1 )); then
            [[ $DRIFT == none ]] || { inject_drift & injector_pid=$!; }
        fi
        stats_before=$(wc -l < "$stats_file")
        compose exec -T builder sh -c 'touch "/gate/go-$1"' _ "$token" || { FAIL_REASON="gate release failed for build $build_idx"; return 1; }
        rm -f "$cell_marker.gap"
        if [[ $DRIFT != none ]]; then
            count=$(dispatch_count) || { FAIL_REASON="dispatch count failed after gate release for build $build_idx"; return 1; }
            post_event drift_note "$count" "build_start b=$build_idx" || { FAIL_REASON="build_start event failed for build $build_idx"; return 1; }
        fi
        sample_stats & stats_pid=$!
        for pid in "${pids[@]}"; do wait "$pid" || true; done
        # The next build's cleanup, gate wait, and result processing are outside
        # build time; leave stress-ng in its current state until gate release.
        if (( build_idx < SESSION_BUILDS )); then touch "$cell_marker.gap"; fi
        # Stop the drift injector as soon as the last build ends, so it cannot
        # start another on-phase while the cell is being torn down.
        if (( build_idx == SESSION_BUILDS )); then touch "$cell_marker"; fi
        if [[ -n $stats_pid ]]; then kill "$stats_pid" 2>/dev/null || true; wait "$stats_pid" 2>/dev/null || true; stats_pid=; fi
        stats_after=$(wc -l < "$stats_file")
        (( stats_after > stats_before )) || { FAIL_REASON="no docker stats sample for build $build_idx"; return 1; }
        min_start=; max_end=
        local build_rows="$OUT_DIR/.build_results.csv"
        : > "$build_rows"
        local -a hostnames=()
        for (( client=1; client<=CLIENTS; client++ )); do
            log="$OUT_DIR/build_${arm}_round${round}_b${build_idx}_c${client}.log"
            [[ -f $log.status ]] || { FAIL_REASON="client $client missing exit status"; return 1; }
            status=$(cat "$log.status")
            [[ $status == 0 ]] || { FAIL_REASON="client $client build exit $status (see $log)"; return 1; }
            if [[ $WORKLOAD != cpython ]] && grep -Fq 'FAILED:' "$log"; then
                FAIL_REASON="client $client ninja reported FAILED: (see $log)"; return 1
            fi
            read -r time_tag start end build_status hostname extra <<< "$(tail -n 1 "$log")"
            [[ $time_tag == HGTIME && $build_status == 0 && -n $hostname && -z ${extra:-} && $start =~ ^[0-9]+([.][0-9]+)?$ && $end =~ ^[0-9]+([.][0-9]+)?$ ]] || {
                FAIL_REASON="client $client missing or failed HGTIME for build $build_idx (see $log)"; return 1;
            }
            hostnames+=("$hostname")
            elapsed=$(python3 -c 'import sys; s,e=map(float,sys.argv[1:]); assert e>=s; print(f"{e-s:.3f}")' "$start" "$end") || { FAIL_REASON="invalid HGTIME for client $client"; return 1; }
            if [[ -z $min_start ]]; then min_start=$start; max_end=$end; else
                read -r min_start max_end <<< "$(python3 -c 'import sys; a,b,c,d=map(float,sys.argv[1:]); print(min(a,c),max(b,d))' "$min_start" "$max_end" "$start" "$end")"
            fi
            printf '%s,%s,%s,%s,%s,%s,PLACEHOLDER,%s,%s\n' "$arm" "$round" "$pos" "$build_idx" "$client" "$elapsed" "$start" "$end" >> "$build_rows"
        done
        if (( CLIENTS == 2 )) && [[ ${hostnames[0]} == "${hostnames[1]}" ]]; then
            FAIL_REASON="build $build_idx clients share HGTIME hostname ${hostnames[0]}"; return 1
        fi
        makespan=$(python3 -c 'import sys; print(f"{float(sys.argv[2])-float(sys.argv[1]):.3f}")' "$min_start" "$max_end")
        python3 -c 'import sys; p=sys.argv[1]; s=open(p).read().replace("PLACEHOLDER",sys.argv[2]); open(p,"w").write(s)' "$build_rows" "$makespan"
        cat "$build_rows" >> "$tmp_results"
    done
    stop_background
    [[ ! -f $cell_marker.error ]] || { FAIL_REASON=$(cat "$cell_marker.error"); return 1; }
    tasks_path="$OUT_DIR/tasks_${arm}_round${round}.jsonl"
    compose cp coordinator:/tmp/tasks.jsonl "$tasks_path" || { FAIL_REASON='task log copy failed'; return 1; }
    span_reason=$(python3 "$ROOT/scripts/exp_checks.py" "$tasks_path" "$tmp_results" "$arm" "$round") || {
        FAIL_REASON="HGTIME span check: $span_reason"; return 1;
    }
    counts=$(python3 - "$tasks_path" <<'PY'
import json,sys
ok=bad=0
post=0
first_on=None
completed=[]
for line in open(sys.argv[1]):
    row=json.loads(line)
    if row.get('event') == 'task_completed':
        if row.get('success') is True: ok+=1
        else: bad+=1
        completed.append(int(row.get('dispatch_seq',0)))
    elif row.get('event') == 'injected_event' and row.get('kind') == 'drift_on' and first_on is None:
        first_on=int(row.get('dispatch_count',0))
if first_on is not None:
    post=sum(seq > first_on for seq in completed)
print(ok,bad,first_on if first_on is not None else -1,post)
PY
) || { FAIL_REASON='task log parse failed'; return 1; }
    read -r successes failures first_on post_completed <<< "$counts"
    if [[ $record == 0 ]]; then
        MIN_TASKS=${MIN_TASKS:-$successes}
        (( successes > 0 && failures == 0 )) || { FAIL_REASON="warm-up tasks: $successes successful, $failures failed"; return 1; }
    else
        (( successes >= MIN_TASKS )) || { FAIL_REASON="only $successes successful tasks; need $MIN_TASKS"; return 1; }
        (( failures == 0 )) || { FAIL_REASON="$failures failed tasks"; return 1; }
        if [[ $DRIFT != none ]]; then
            (( first_on >= 101 )) || { FAIL_REASON="missing valid drift_on event (first dispatch_count=$first_on)"; return 1; }
            (( first_on <= DRIFT_AT + 40 )) || { FAIL_REASON="drift_on too late: dispatch_count=$first_on exceeds $((DRIFT_AT + 40))"; return 1; }
            (( post_completed >= 100 )) || { FAIL_REASON="only $post_completed completed tasks after drift_on; need 100"; return 1; }
        fi
        cat "$tmp_results" >> "$OUT_DIR/results.csv"
    fi
    return 0
}

if [[ $WORKLOAD == syn ]]; then
    # The builder image has no Python. Generate once on the host, then copy the
    # project into each persistent builder volume when its marker differs.
    syn_host_dir=$(mktemp -d)
    python3 "$ROOT/scripts/gen_synthetic_cpp.py" --out "$syn_host_dir/syn" --seed "$SYN_SEED" --files "$SYN_FILES" \
        --c-fraction "$SYN_C_FRACTION" --pareto-alpha "$SYN_ALPHA" --median-units "$SYN_MEDIAN_UNITS" \
        --max-units "$SYN_MAX_UNITS" --hidden-cost-fraction "$SYN_HIDDEN" || die 'synthetic project generation failed'
    SYN_MANIFEST_SHA256=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["manifest_sha256"])' "$syn_host_dir/syn/manifest.json") || die 'synthetic manifest is missing manifest_sha256'
    [[ $SYN_MANIFEST_SHA256 =~ ^[a-fA-F0-9]{64}$ ]] || die 'invalid synthetic manifest_sha256'
    SYN_MARKER="$SYN_MANIFEST_SHA256 $SYN_SEED $SYN_FILES $SYN_C_FRACTION $SYN_ALPHA $SYN_MEDIAN_UNITS $SYN_MAX_UNITS $SYN_HIDDEN"
fi

# Build the common image once, before the discarded warm-up and measured cells.
generate_compose "${arm_flags[0]}"
compose build
if (( ROUND_START == 1 )); then
    printf 'warm-up (discarded)\n'
    for attempt in 1 2; do
        if run_cell "${arm_list[0]}" 0 0 0 "${arm_flags[0]}" "$attempt"; then break; fi
        stop_background
        compose logs --no-color --timestamps > "$OUT_DIR/compose_${arm_list[0]}_round0_a${attempt}.log" 2>&1 || true
        printf '%s,0,%s,%s\n' "${arm_list[0]}" "$attempt" "${FAIL_REASON//,/;}" >> "$OUT_DIR/failures.csv"
        (( attempt < 2 )) || die "warm-up failed twice: $FAIL_REASON"
    done
    rm -f "$OUT_DIR/tasks_${arm_list[0]}_round0.jsonl"
    write_meta
fi
for (( round=ROUND_START; round<=REPS; round++ )); do
    order=$(order_for_round "$round")
    printf '%s,%s\n' "$round" "${order// /|}" >> "$OUT_DIR/order_log.csv"
    read -r -a shuffled <<< "$order"
    pos=0
    for arm in "${shuffled[@]}"; do
        pos=$((pos+1))
        for i in "${!arm_list[@]}"; do [[ ${arm_list[$i]} != "$arm" ]] || { flags=${arm_flags[$i]}; break; }; done
        for attempt in 1 2; do
            if run_cell "$arm" "$round" "$pos" 1 "$flags" "$attempt"; then
                printf 'completed %s round %s\n' "$arm" "$round"
                break
            fi
            stop_background
            # Keep the cluster logs of a failed cell for diagnosis; the containers
            # are only removed when the next cell starts.
            compose logs --no-color --timestamps > "$OUT_DIR/compose_${arm}_round${round}_a${attempt}.log" 2>&1 || true
            printf '%s,%s,%s,%s\n' "$arm" "$round" "$attempt" "${FAIL_REASON//,/;}" >> "$OUT_DIR/failures.csv"
            printf 'cell failed: %s (attempt %s): %s\n' "$arm" "$attempt" "$FAIL_REASON" >&2
            (( attempt < 2 )) || die "cell $arm round $round failed twice; resume with ROUND_START=$round"
        done
    done
done
printf 'results: %s/results.csv\n' "$OUT_DIR"
