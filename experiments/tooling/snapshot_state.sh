#!/bin/bash
# Record the machine and campaign state into a directory OUTSIDE out/official,
# so that a later resume can be checked against it (resume_check.sh).
# Read-only except for the output directory. Exit non-zero if any part fails.
# Usage: snapshot_state.sh <out_dir>
set -u -o pipefail
. /d/bench-work/env.sh
D=${1:?usage: snapshot_state.sh <out_dir>}
O=$BENCH_OUT/official
case $D in "$O"|"$O"/*) echo "snapshot must be outside $O" >&2; exit 2 ;; esac
mkdir -p "$D" || exit 1
err=0
run() { "$@" || { echo "snapshot step failed: $*" >&2; err=1; }; }
date '+%F %T %z' > "$D/taken_at.txt"

# Host identity: same physical machine, CPU, RAM and Windows build.
run powershell.exe -NoProfile -Command "\$c=Get-CimInstance Win32_ComputerSystemProduct; \$p=Get-CimInstance Win32_Processor; \$o=Get-CimInstance Win32_OperatingSystem; \$s=Get-CimInstance Win32_ComputerSystem; 'UUID=' + \$c.UUID; 'CPU=' + ((\$p | ForEach-Object { \$_.Name.Trim() }) -join ' | '); 'Cores=' + ((\$p | Measure-Object NumberOfCores -Sum).Sum) + ' Threads=' + ((\$p | Measure-Object NumberOfLogicalProcessors -Sum).Sum); 'RAM=' + \$s.TotalPhysicalMemory; 'Windows=' + \$o.Version + ' build ' + \$o.BuildNumber" > "$D/host.txt.tmp"
tr -d '\r' < "$D/host.txt.tmp" > "$D/host.txt"; rm -f "$D/host.txt.tmp"

# Processes (name, count, working set MB) and a 10 s CPU-load sample.
run powershell.exe -NoProfile -Command "Get-Process | Group-Object ProcessName | ForEach-Object { '{0},{1},{2}' -f \$_.Name, \$_.Count, [math]::Round((\$_.Group | Measure-Object WorkingSet64 -Sum).Sum/1MB) } | Sort-Object" > "$D/processes.tmp"
tr -d '\r' < "$D/processes.tmp" > "$D/processes.csv"; rm -f "$D/processes.tmp"
run powershell.exe -NoProfile -Command "[math]::Round(((Get-Counter '\Processor(_Total)\% Processor Time' -SampleInterval 2 -MaxSamples 5).CounterSamples | Measure-Object CookedValue -Average).Average, 1)" > "$D/cpu_load.tmp"
tr -d '\r' < "$D/cpu_load.tmp" > "$D/cpu_load_percent.txt"; rm -f "$D/cpu_load.tmp"

# Docker engine, containers, images and volumes used by the runner.
run docker info --format 'NCPU={{.NCPU}} MemTotal={{.MemTotal}} ServerVersion={{.ServerVersion}} OS={{.OperatingSystem}} Kernel={{.KernelVersion}}' > "$D/docker_info.txt"
run docker ps -a --format '{{.Names}} {{.Status}}' > "$D/docker_containers.txt"
docker images --no-trunc --format '{{.Repository}}:{{.Tag}} {{.ID}}' > "$D/docker_images.tmp" || err=1
sort "$D/docker_images.tmp" > "$D/docker_images.txt"; rm -f "$D/docker_images.tmp"
docker volume ls --format '{{.Name}}' > "$D/docker_volumes.tmp" || err=1
sort "$D/docker_volumes.tmp" > "$D/docker_volumes.txt"; rm -f "$D/docker_volumes.tmp"

# Power plan, sleep settings, Windows Update pause (informational).
powershell.exe -NoProfile -Command "powercfg /getactivescheme; powercfg /query SCHEME_CURRENT SUB_SLEEP STANDBYIDLE | Select-String 'Current AC'" | tr -d '\r' > "$D/power.txt"
powershell.exe -NoProfile -Command "(Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\WindowsUpdate\UX\Settings').PauseUpdatesExpiryTime" | tr -d '\r' > "$D/windows_update.txt"

# Code, tooling, Python identity, WSL configuration.
{ git -C "$BENCH_REPO" rev-parse HEAD && git -C "$BENCH_REPO" status --short | wc -l; } > "$D/clone.txt" || err=1
(cd /d/bench-work && sha256sum campaign.sh stop_campaign.sh analysis_marker.sh preflight.sh power.ps1 env.sh requirements.txt bin/python3) > "$D/tooling.sha256" || err=1
(cd /d/bench-work && sha256sum supervisor.sh supervisor_prompt.md) > "$D/tooling_audit_only.sha256" || err=1
{ command -v python3; python3 -c 'import sys; print(sys.executable); print(sys.version)'; sha256sum /d/bench-work/venv/Scripts/python.exe /d/bench-work/venv/pyvenv.cfg; python3 -m pip freeze; } > "$D/python.txt" || err=1
if [[ -f "$USERPROFILE/.wslconfig" ]]; then sha256sum "$USERPROFILE/.wslconfig" > "$D/wslconfig.sha256"; else echo "no .wslconfig" > "$D/wslconfig.sha256"; fi

# Campaign flags and an integrity manifest of every official data file
# (supervisor audits, stdout logs, campaign.log, pid/stop flags excluded).
cp -p "$O/campaign.log" "$D/campaign.log.copy" || err=1
for f in PHASE2_APPROVED PHASE2_DECISION STOP_REQUESTED campaign.pid runner.pid; do
    if [[ -e $O/$f ]]; then echo "$f present: $(head -n 1 "$O/$f" 2>/dev/null)"; else echo "$f absent"; fi
done > "$D/campaign_flags.txt"
(cd "$O" && find . -type f ! -path './supervisor/*' ! -name 'stdout.log' ! -name 'supervisor_stdout.log' \
    ! -name campaign.log ! -name '*.pid' ! -name STOP_REQUESTED ! -name pause.log -print0 | sort -z | xargs -0 sha256sum) \
    > "$D/manifest_official.sha256.tmp" || err=1
mv "$D/manifest_official.sha256.tmp" "$D/manifest_official.sha256"
wc -l < "$D/manifest_official.sha256" > "$D/manifest_count.txt"
for f in host.txt processes.csv docker_info.txt clone.txt tooling.sha256 python.txt manifest_official.sha256; do
    [[ -s $D/$f ]] || { echo "snapshot file empty: $f" >&2; err=1; }
done
(( err == 0 )) || { echo "SNAPSHOT INCOMPLETE: $D" >&2; exit 1; }
echo "snapshot written to $D ($(cat "$D/manifest_count.txt") data files)"
