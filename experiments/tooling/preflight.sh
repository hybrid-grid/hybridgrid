#!/bin/bash
# Read-only environment check before a benchmark session. It never kills a
# process and never changes a setting.
#
# Usage:  . /d/bench-work/env.sh
#         bash /d/bench-work/preflight.sh [report-file]
#   report-file   optional; the full output is also written there (keep it
#                 next to the results, e.g. "$OUT_DIR/preflight.txt").
# Environment:
#   EXPECT_COMMIT   commit prefix the repo must be at (official runs: set it)
#   STRICT=1        official-run mode: every WARN counts as FAIL
# Exit code: 0 = no FAIL, 1 = at least one FAIL.

if [[ -n ${1:-} && -z ${PREFLIGHT_TEE:-} ]]; then
    PREFLIGHT_TEE=1 exec bash "$0" "$@" > >(tee "$1") 2>&1
fi

fail=0
STRICT=${STRICT:-0}
ok()   { printf 'OK    %s\n' "$1"; }
info() { printf 'INFO  %s\n' "$1"; }
bad()  { printf 'FAIL  %s\n' "$1"; fail=1; }
warn() { if [[ $STRICT == 1 ]]; then printf 'FAIL  %s (strict)\n' "$1"; fail=1; else printf 'WARN  %s\n' "$1"; fi; }

REPO=${BENCH_REPO:-/d/bench-work/hybridgrid}
echo "preflight $(date -Is)  STRICT=$STRICT  EXPECT_COMMIT=${EXPECT_COMMIT:-<unset>}"

# 1. python3 must be the venv shim with the exact library versions of the pilots
py=$(command -v python3 || true)
if [[ $py == */bench-work/bin/python3 ]]; then ok "python3 -> $py"; else bad "python3 resolves to '${py:-nothing}', expected the bench-work shim (source env.sh)"; fi
ver=$(python3 -c 'import pandas,scipy,numpy,sys; print(pandas.__version__, scipy.__version__, numpy.__version__, sys.version.split()[0])' 2>/dev/null || true)
if [[ $ver == "2.3.3 1.17.1 2.4.6 "* ]]; then ok "pandas/scipy/numpy/python: $ver"
elif [[ -z $ver ]]; then bad "python3 cannot import pandas/scipy/numpy"
else bad "library versions '$ver' differ from the pilot pins (pandas 2.3.3, scipy 1.17.1, numpy 2.4.6); see /d/bench-work/requirements.txt"; fi

# 2. repo: commit, branch, clean tree, LF endings
if [[ -d $REPO/.git ]]; then
    head=$(git -C "$REPO" rev-parse HEAD) || head=
    info "repo HEAD: $head  $(git -C "$REPO" log --oneline -1 | cut -d' ' -f2-)"
    if [[ -z ${EXPECT_COMMIT:-} ]]; then warn "EXPECT_COMMIT not set: the code under test is not pinned"
    elif [[ -n $head && $head == "$EXPECT_COMMIT"* ]]; then ok "commit matches EXPECT_COMMIT"
    else bad "HEAD ${head:-?} does not match EXPECT_COMMIT=$EXPECT_COMMIT"; fi
    br=$(git -C "$REPO" rev-parse --abbrev-ref HEAD)
    expect_branch=${EXPECT_BRANCH:-feat/exp-readiness}   # Experiment 3 runs set EXPECT_BRANCH=feat/exp3
    [[ $br == "$expect_branch" ]] && ok "branch $br" || warn "branch is $br (expected $expect_branch)"
    st=$(git -C "$REPO" status --porcelain) || st=ERR
    [[ -z $st ]] && ok "repo working tree clean" || bad "repo working tree is dirty or unreadable (meta.json would record dirty=true)"
    [[ $(git -C "$REPO" config core.autocrlf) == false ]] && ok "core.autocrlf=false" || bad "core.autocrlf is not false"
    # grep $'\r' is unreliable in Git Bash; ask git for the working-tree EOL.
    if eol=$(git -C "$REPO" ls-files --eol -- scripts test/stress); then
        if [[ -n $eol ]] && ! grep -q 'w/crlf' <<<"$eol"; then ok "no CRLF in scripts/ and test/stress/"
        else bad "CRLF found (or nothing listed) in scripts/ and test/stress/"; fi
    else bad "git ls-files --eol failed"; fi
else
    bad "repo not found at $REPO"
fi

# 3. Docker: up, enough resources, nothing else running
if docker info >/dev/null 2>&1; then
    ncpu=$(docker info --format '{{.NCPU}}'); mem=$(( $(docker info --format '{{.MemTotal}}') / 1024 / 1024 / 1024 ))
    info "docker NCPU=$ncpu Mem=${mem} GiB (reference host: 72 CPU / 31 GiB)"
    if [[ $ncpu -lt 16 || $mem -lt 16 ]]; then bad "docker VM too small for a 5-worker cluster plus builders (need >= 16 CPU and >= 16 GiB)"
    elif [[ $ncpu -ne 72 || $mem -ne 31 ]]; then warn "docker VM differs from the reference host (72 CPU / 31 GiB)"
    else ok "docker VM matches the reference host"; fi
    n=$(docker ps -q | wc -l)
    [[ $n -eq 0 ]] && ok "no running containers" || warn "$n container(s) already running (they would share the VM)"
    docker compose version >/dev/null 2>&1 && ok "docker compose available" || bad "docker compose missing"
else
    bad "docker daemon not reachable"
fi

# 4. disk (Docker data lives on C:, results on D:)
free_gb() { df -BG --output=avail "$1" 2>/dev/null | tail -1 | tr -dc '0-9'; }
c=$(free_gb /c); d=$(free_gb /d)
[[ ${c:-0} -ge 15 ]] && ok "C: free ${c} GB" || bad "C: free ${c:-?} GB (< 15 GB, Docker data lives there)"
[[ ${d:-0} -ge 20 ]] && ok "D: free ${d} GB" || bad "D: free ${d:-?} GB (< 20 GB)"

# 5. host description (recorded in the report file) and background apps
info "host: $(powershell.exe -NoProfile -Command '(Get-CimInstance Win32_OperatingSystem).Caption + " | " + (Get-CimInstance Win32_Processor | Select-Object -First 1).Name + " x" + (Get-CimInstance Win32_Processor).Count' 2>/dev/null | tr -d '\r')"
noisy=$(powershell.exe -NoProfile -Command "Get-Process | Where-Object { \$_.Name -match '^(Unity|Unity Hub|rider64|Rider.Backend|msedge|chrome|firefox|Teams|ms-teams|OneDrive|Code|Cursor|Slack|Discord|Spotify)\$' } | Group-Object Name | ForEach-Object { \$_.Name + 'x' + \$_.Count }" 2>/dev/null | tr -d '\r' | tr '\n' ' ')
[[ -z $noisy ]] && ok "no known noisy apps running" || warn "noisy apps running: $noisy (close them; the list is not exhaustive)"

# 6. power plan and sleep (English Windows output is parsed)
plan=$(powershell.exe -NoProfile -Command "powercfg /getactivescheme" 2>/dev/null | tr -d '\r')
info "power plan: ${plan#*: }"
case $plan in
    *"High performance"*|*"Ultimate Performance"*) ok "power plan is a performance plan" ;;
    *) warn "power plan is not High performance (powershell -File D:\\bench-work\\power.ps1 on)" ;;
esac
for sub in STANDBYIDLE HIBERNATEIDLE; do
    v=$(powershell.exe -NoProfile -Command "powercfg /query SCHEME_CURRENT SUB_SLEEP $sub" 2>/dev/null | tr -d '\r' | grep -i 'Current AC Power Setting Index' | sed 's/.*: *//')
    if [[ $v == 0x00000000 ]]; then ok "$sub on AC: never"
    else warn "$sub on AC is '${v:-unknown}' (not 0x00000000): the PC may sleep during a long run"; fi
done
info "not checked by this script: Windows Update restarts (pause updates / set active hours), antivirus scanning of D:\\bench-work and Docker data (add an exclusion), clock sync"

echo
[[ $fail -eq 0 ]] && echo "preflight: no FAIL" || echo "preflight: FAILED"
exit $fail
