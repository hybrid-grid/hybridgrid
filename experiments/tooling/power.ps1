# Switch the Windows power plan for a benchmark session and restore it afterwards.
#   powershell -File D:\bench-work\power.ps1 on       (remember current plan, switch to High performance)
#   powershell -File D:\bench-work\power.ps1 off      (restore the remembered plan)
#   powershell -File D:\bench-work\power.ps1 status
# 'off' restores only when High performance is still the active plan. If the
# user switched to another plan in between, that choice is left untouched.
param([Parameter(Mandatory = $true)][ValidateSet('on', 'off', 'status')][string]$Action)

$ErrorActionPreference = 'Stop'
$high = '8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c'
$state = 'D:\bench-work\power-previous.txt'

function Get-ActiveGuid {
    $line = powercfg /getactivescheme
    if ($line -match '([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})') { return $Matches[1].ToLower() }
    throw "cannot parse active power scheme: $line"
}

function Test-SchemeExists([string]$guid) {
    return [bool](powercfg /list | Select-String -SimpleMatch $guid)
}

switch ($Action) {
    'status' {
        powercfg /getactivescheme
        if (Test-Path $state) { "remembered previous plan: $((Get-Content $state -Raw).Trim())" } else { 'no remembered plan' }
    }
    'on' {
        if (-not (Test-SchemeExists $high)) { throw "High performance plan ($high) does not exist on this machine" }
        $cur = Get-ActiveGuid
        if ($cur -ne $high) {
            # a stale file from an earlier session is replaced by the plan active right now
            Set-Content -Path $state -Value $cur -NoNewline
            powercfg /setactive $high
        }
        powercfg /getactivescheme
    }
    'off' {
        if (-not (Test-Path $state)) { 'nothing to restore (no remembered plan)'; exit 0 }
        $prev = (Get-Content $state -Raw).Trim()
        $cur = Get-ActiveGuid
        if ($cur -ne $high) {
            "active plan is no longer High performance ($cur); leaving it unchanged, forgetting the remembered plan"
        } elseif (-not (Test-SchemeExists $prev)) {
            "remembered plan $prev no longer exists; leaving High performance active"
        } else {
            powercfg /setactive $prev
        }
        Remove-Item $state
        powercfg /getactivescheme
    }
}
