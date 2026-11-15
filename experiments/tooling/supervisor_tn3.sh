#!/bin/bash
# Read-only Codex supervision of the running campaign. Never touches the
# campaign: it only writes audit files and a one-line summary per audit.
# Usage: nohup bash supervisor.sh [interval_seconds] &   (default 5400 = 90 min)
# Ends when the campaign log shows CAMPAIGN DONE / STOP_AFTER reached / STOPPED
# (after one final audit) or when a file STOP_SUPERVISOR exists next to it.
. /d/bench-work/env.sh
INTERVAL=${1:-5400}
[[ $INTERVAL =~ ^[0-9]+$ ]] && (( INTERVAL > 0 )) || { echo 'interval must be a positive integer' >&2; exit 2; }
OFF=${CAMPAIGN:-$BENCH_OUT/official-tn3}
SUP=$OFF/supervisor
mkdir -p "$SUP"
SLOG=$SUP/supervisor.log
n=0
for old_audit in "$SUP"/audit_*.md; do
    [[ -f $old_audit ]] || continue
    old_number=${old_audit##*/audit_}; old_number=${old_number%.md}
    [[ $old_number =~ ^[0-9]+$ ]] || continue
    if (( 10#$old_number > n )); then n=$((10#$old_number)); fi
done
lock=$SUP/supervisor.lock
if [[ -e $lock ]]; then
    old_pid=$(cat "$lock" 2>/dev/null || true)
    if [[ $old_pid =~ ^[0-9]+$ ]] && kill -0 "$old_pid" 2>/dev/null; then
        echo "supervisor already running (PID $old_pid)" >&2; exit 1
    fi
    rm -f "$lock"
fi
( set -C; printf '%s\n' "$$" > "$lock" ) 2>/dev/null || { echo 'supervisor lock exists' >&2; exit 1; }
trap 'rm -f "$lock"' EXIT
start_lines=$(wc -l < "$OFF/campaign.log" 2>/dev/null || printf '0')
finished() {
    [[ -f $OFF/campaign.log ]] || return 1
    current_lines=$(wc -l < "$OFF/campaign.log" 2>/dev/null || printf '0')
    if (( current_lines < start_lines )); then
        echo "$(date '+%F %T') campaign.log shrank from $start_lines to $current_lines lines; resetting start offset" >> "$SLOG"
        start_lines=0
    fi
    tail -n "+$((start_lines + 1))" "$OFF/campaign.log" |
        grep -a -q -E 'CAMPAIGN DONE|STOP_AFTER reached|CAMPAIGN STOPPED|WAITING FOR PHASE2_APPROVED|MANUAL STOP|stopped by STOP_REQUESTED'
}
notify_operator() {
    local message=$1
    if [[ -n ${NOTIFY_CMD:-} ]]; then
        "$NOTIFY_CMD" "$message" >/dev/null 2>&1 &
    else
        powershell.exe -NoProfile -Command 'Add-Type -AssemblyName System.Windows.Forms; $n=New-Object System.Windows.Forms.NotifyIcon; $n.Icon=[System.Drawing.SystemIcons]::Warning; $n.Visible=$true; $n.BalloonTipTitle="Benchmark supervisor"; $n.BalloonTipText=$args[0]; $n.ShowBalloonTip(10000); Start-Sleep -Seconds 11; $n.Dispose()' "$message" >/dev/null 2>&1 &
    fi
}
echo "$(date '+%F %T') supervisor start interval=${INTERVAL}s" >> "$SLOG"
while :; do
    n=$((n + 1))
    f=$SUP/audit_$(printf '%03d' $n).md
    if [[ -n ${CODEX_CMD:-} ]]; then
        timeout 2700 "$CODEX_CMD" -o "$f" - < /d/bench-work/supervisor_prompt_tn3.md > "$SUP/audit_$(printf '%03d' "$n").log" 2>&1
    else
        timeout 2700 codex exec --skip-git-repo-check -m gpt-5.6-sol -c model_reasoning_effort=high -s read-only \
            -C 'D:\bench-work' --add-dir 'D:\Đồ án tốt nghiệp\hybridgrid' \
            -o "$f" - < /d/bench-work/supervisor_prompt_tn3.md > "$SUP/audit_$(printf '%03d' "$n").log" 2>&1
    fi
    rc=$?
    sev=$(head -n 1 "$f" 2>/dev/null | tr -d '\r')
    [[ -n $sev ]] || sev="SEVERITY: UNKNOWN (codex rc=$rc, no output)"
    echo "$(date '+%F %T') audit $n: $sev -> $f" >> "$SLOG"
    if [[ $sev == 'SEVERITY: ALERT' || $sev == 'SEVERITY: WARN' ]]; then
        echo "$(date '+%F %T') NOTIFY audit $n: $sev" >> "$SLOG"
        notify_operator "Benchmark supervisor audit $n: $sev"
    fi
    if finished; then
        if [[ ${last_final:-0} == 1 ]]; then break; fi
        echo "$(date '+%F %T') terminal line observed after audit $n; one final audit" >> "$SLOG"
        last_final=1; continue
    fi
    [[ -e $SUP/STOP_SUPERVISOR ]] && { echo "$(date '+%F %T') stopped by STOP_SUPERVISOR" >> "$SLOG"; break; }
    sleep "$INTERVAL"
done
echo "$(date '+%F %T') supervisor end" >> "$SLOG"
