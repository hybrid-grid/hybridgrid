#!/bin/bash
# Compare the machine and campaign state with the snapshot taken at the last
# pause (/d/bench-work/pause_snapshots/LAST_PAUSE) before resuming. Read-only.
# Exit 0 = safe to resume (review WARN lines), 1 = do not resume.
# Usage: resume_check.sh [snapshot_dir]
set -u -o pipefail
. /d/bench-work/env.sh
O=$BENCH_OUT/official
P=/d/bench-work/pause_snapshots
S=${1:-$(cat "$P/LAST_PAUSE" 2>/dev/null)}
[[ -n $S && -d $S ]] || { echo "no snapshot directory (LAST_PAUSE missing?)"; exit 1; }
T=$(mktemp -d /d/bench-work/tmp-tests/resume.XXXXXX) || exit 1
trap 'rm -rf "$T"' EXIT
fail=0; warn=0
ok()   { echo "OK    $*"; }
bad()  { echo "FAIL  $*"; fail=1; }
note() { echo "WARN  $*"; warn=1; }
echo "comparing with snapshot $S (taken $(cat "$S/taken_at.txt"))"

# 1. Same physical machine, CPU, RAM, Windows build.
powershell.exe -NoProfile -Command "\$c=Get-CimInstance Win32_ComputerSystemProduct; \$p=Get-CimInstance Win32_Processor; \$o=Get-CimInstance Win32_OperatingSystem; \$s=Get-CimInstance Win32_ComputerSystem; 'UUID=' + \$c.UUID; 'CPU=' + ((\$p | ForEach-Object { \$_.Name.Trim() }) -join ' | '); 'Cores=' + ((\$p | Measure-Object NumberOfCores -Sum).Sum) + ' Threads=' + ((\$p | Measure-Object NumberOfLogicalProcessors -Sum).Sum); 'RAM=' + \$s.TotalPhysicalMemory; 'Windows=' + \$o.Version + ' build ' + \$o.BuildNumber" \
    | tr -d '\r' > "$T/host.txt"
if diff -q "$T/host.txt" "$S/host.txt" >/dev/null; then ok "same host: $(grep CPU= "$T/host.txt"), $(grep Windows= "$T/host.txt")"
else bad "host differs (machine, CPU, RAM or Windows build):"; diff "$S/host.txt" "$T/host.txt"; fi

# 2. Data integrity: every official data file unchanged, no unexpected file.
if (cd "$O" && sha256sum --quiet -c "$S/manifest_official.sha256") > "$T/manifest.txt" 2>&1; then
    ok "all $(cat "$S/manifest_count.txt") official data files unchanged"
else bad "official data changed or missing:"; head -n 20 "$T/manifest.txt"; fi
(cd "$O" && find . -type f ! -path './supervisor/*' ! -name 'stdout.log' ! -name 'supervisor_stdout.log' \
    ! -name campaign.log ! -name '*.pid' ! -name STOP_REQUESTED ! -name pause.log | sort) > "$T/files_now.txt"
sed 's/^[0-9a-f]* [ *]//' "$S/manifest_official.sha256" | sort > "$T/files_then.txt"
extra=$(comm -23 "$T/files_now.txt" "$T/files_then.txt")
[[ -z $extra ]] && ok "no new files in out/official" || { bad "files added to out/official since the pause:"; echo "$extra" | head -n 20; }
cmp -s <(head -c "$(stat -c %s "$S/campaign.log.copy")" "$O/campaign.log") "$S/campaign.log.copy" \
    && ok "campaign.log unchanged up to the pause (appended lines: $(( $(wc -l < "$O/campaign.log") - $(wc -l < "$S/campaign.log.copy") )))" \
    || bad "campaign.log was modified before the pause point"

# 3. Campaign flags: decision unchanged, stop still requested, nothing running.
[[ $(cat "$O/PHASE2_DECISION" 2>/dev/null) == E2_LAMBDA=RUN ]] && ok "lambda decision unchanged (E2_LAMBDA=RUN)" || bad "PHASE2_DECISION is not E2_LAMBDA=RUN"
[[ -e $O/STOP_REQUESTED ]] && ok "STOP_REQUESTED present (delete it only when resuming)" || note "STOP_REQUESTED already removed"
live=0
if [[ -e $O/campaign.pid ]]; then
    pid=$(cat "$O/campaign.pid" 2>/dev/null)
    if [[ $pid =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null && tr '\0' ' ' < "/proc/$pid/cmdline" 2>/dev/null | grep -q campaign.sh; then live=1; fi
fi
(( live )) && bad "a campaign is already running (PID $pid)" || ok "no campaign running$([[ -e $O/campaign.pid ]] && echo ' (stale campaign.pid will be replaced by campaign.sh)')"
running=$(docker ps -q 2>/dev/null | wc -l); all_hg=$(docker ps -aq --filter label=com.docker.compose.project=hgexp 2>/dev/null | wc -l)
(( all_hg == 0 )) && ok "no hgexp containers" || bad "$all_hg hgexp containers exist (run stop_campaign.sh)"
(( running == 0 )) && ok "no other running containers" || note "$running other containers running (they compete for CPU): $(docker ps --format '{{.Names}}' | tr '\n' ' ')"

# 4. Docker engine, images, volumes.
docker info --format 'NCPU={{.NCPU}} MemTotal={{.MemTotal}} ServerVersion={{.ServerVersion}} OS={{.OperatingSystem}} Kernel={{.KernelVersion}}' > "$T/docker_info.txt" 2>&1
diff -q "$T/docker_info.txt" "$S/docker_info.txt" >/dev/null && ok "Docker: $(cat "$T/docker_info.txt")" \
    || bad "Docker differs: now '$(cat "$T/docker_info.txt")', snapshot '$(cat "$S/docker_info.txt")' (start Docker Desktop / restore its resources)"
docker images --no-trunc --format '{{.Repository}}:{{.Tag}} {{.ID}}' | sort > "$T/images.txt"
diff -q "$T/images.txt" "$S/docker_images.txt" >/dev/null && ok "Docker images identical" \
    || { note "Docker images differ (the runner rebuilds from the same Dockerfiles; check the base image):"; diff "$S/docker_images.txt" "$T/images.txt" | head -n 10; }
docker volume ls --format '{{.Name}}' | sort > "$T/volumes.txt"
diff -q "$T/volumes.txt" "$S/docker_volumes.txt" >/dev/null && ok "Docker volumes identical" || { bad "Docker volumes differ:"; diff "$S/docker_volumes.txt" "$T/volumes.txt"; }

# 5. Code, tooling, Python, WSL.
[[ $(git -C "$BENCH_REPO" rev-parse HEAD) == $(head -n 1 "$S/clone.txt") && $(git -C "$BENCH_REPO" status --short | wc -l) == 0 ]] \
    && ok "benchmark clone at $(git -C "$BENCH_REPO" rev-parse --short HEAD), clean" || bad "benchmark clone moved or dirty"
(cd /d/bench-work && sha256sum --quiet -c "$S/tooling.sha256") >/dev/null 2>&1 && ok "measurement tooling unchanged" || bad "measurement tooling changed since the pause"
(cd /d/bench-work && sha256sum --quiet -c "$S/tooling_audit_only.sha256") >/dev/null 2>&1 && ok "supervisor files unchanged" || note "supervisor.sh or supervisor_prompt.md changed (audit only, not measurement)"
{ command -v python3; python3 -c 'import sys; print(sys.executable); print(sys.version)'; sha256sum /d/bench-work/venv/Scripts/python.exe /d/bench-work/venv/pyvenv.cfg; python3 -m pip freeze; } > "$T/python.txt" 2>&1
diff -q "$T/python.txt" "$S/python.txt" >/dev/null && ok "Python interpreter and packages unchanged" || { bad "Python environment changed:"; diff "$S/python.txt" "$T/python.txt" | head -n 10; }
if [[ -f "$USERPROFILE/.wslconfig" ]]; then w=$(sha256sum "$USERPROFILE/.wslconfig"); else w="no .wslconfig"; fi
[[ $w == "$(cat "$S/wslconfig.sha256")" ]] && ok ".wslconfig unchanged" || bad ".wslconfig changed (WSL2 VM resources may differ)"

# 6. Background load: noisy apps, new heavy processes, CPU load.
powershell.exe -NoProfile -Command "Get-Process | Group-Object ProcessName | ForEach-Object { '{0},{1},{2}' -f \$_.Name, \$_.Count, [math]::Round((\$_.Group | Measure-Object WorkingSet64 -Sum).Sum/1MB) } | Sort-Object" \
    | tr -d '\r' > "$T/processes.csv"
noisy=$(grep -iE '^(msedge|chrome|firefox|rider64|Rider\.Backend|Unity|Unity Hub|devenv|Code|steam|Teams|ms-teams|Discord|OneDrive|Fork),' "$T/processes.csv" | cut -d, -f1 | tr '\n' ' ')
[[ -z $noisy ]] && ok "no known noisy apps" || note "noisy apps running: $noisy(close them before resuming)"
heavy=$(awk -F, 'NR==FNR { seen[$1]=1; next } !($1 in seen) && $3 > 300 { printf "%s(%s MB) ", $1, $3 }' "$S/processes.csv" "$T/processes.csv")
[[ -z $heavy ]] && ok "no new heavy processes compared with the pause" || note "new heavy processes: $heavy"
load=$(powershell.exe -NoProfile -Command "[math]::Round(((Get-Counter '\Processor(_Total)\% Processor Time' -SampleInterval 2 -MaxSamples 5).CounterSamples | Measure-Object CookedValue -Average).Average, 1)" | tr -d '\r')
then_load=$(cat "$S/cpu_load_percent.txt" 2>/dev/null)
awk -v l="$load" 'BEGIN { exit !(l + 0 <= 5) }' && ok "idle CPU load ${load}% (at pause: ${then_load}%)" || note "CPU load ${load}% while idle (at pause: ${then_load}%): find what is running"

# 7. Windows Update pause (the remaining run is about 6.5 h).
exp=$(powershell.exe -NoProfile -Command "(Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\WindowsUpdate\UX\Settings').PauseUpdatesExpiryTime" | tr -d '\r')
if [[ -z $exp ]]; then note "Windows Update is not paused"; else
    left=$(( ($(date -d "$exp" +%s) - $(date +%s)) / 3600 ))
    (( left >= 12 )) && ok "Windows Update paused until $exp (${left} h left)" || note "Windows Update pause ends $exp (${left} h left): extend it before resuming"
fi

echo
if (( fail )); then echo "RESULT: DO NOT RESUME (fix FAIL items first)"; exit 1; fi
(( warn )) && echo "RESULT: resumable after reviewing WARN items" || echo "RESULT: identical state, safe to resume"
cat <<'EOF'
Resume:
  rm /d/bench-work/out/official/STOP_REQUESTED
  cd /d/bench-work && EXPECT_COMMIT=1b8339d BUILDER_CPUS=2.0 nohup bash campaign.sh >> out/official/stdout.log 2>&1 &
  nohup bash /d/bench-work/supervisor.sh 5400 >> /d/bench-work/out/official/supervisor_stdout.log 2>&1 &
campaign.sh switches the power plan to High performance, runs STRICT preflight, archives the
partial block and reruns it from its first arm with the same seed and order.
EOF
