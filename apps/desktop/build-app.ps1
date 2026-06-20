<#
.SYNOPSIS
  Build the IAD desktop app for Windows and bundle the iad-agent scanner beside it.

.DESCRIPTION
  Produces a self-contained folder under build/bin/ that contains:
    - iad-console.exe      (Wails desktop UI, release build)
    - iad-console-dev.exe  (Wails desktop UI, debug build with devtools)  [Mode dev/both]
    - iad-agent.exe        (the network scanner the UI invokes)
    - rules/               (access-detection rules the agent needs)
  so the desktop app's "Run Scan" works out of the box (app.go resolveAgentBin
  finds the agent next to the executable).

.PARAMETER Mode
  prod  (default) -> only the release iad-console.exe
  dev             -> only the debug iad-console-dev.exe (devtools / right-click inspect)
  both            -> both consoles

.PARAMETER SkipAgent
  Skip rebuilding iad-agent.exe and re-copying rules/ (faster UI-only iterations).

.EXAMPLE
  .\build-app.ps1                 # release console + agent + rules
  .\build-app.ps1 -Mode dev       # debug console + agent + rules
  .\build-app.ps1 -Mode both      # both consoles + agent + rules
  .\build-app.ps1 -SkipAgent      # release console only (no agent rebuild)

  Requires: Go 1.22+, Node 18+, and the Wails CLI
  (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).
#>
[CmdletBinding()]
param(
    [ValidateSet('prod', 'dev', 'both')]
    [string]$Mode = 'prod',
    [switch]$SkipAgent
)

$ErrorActionPreference = 'Stop'
$desktop = $PSScriptRoot
$repo = Resolve-Path (Join-Path $desktop '..\..')
$agent = Join-Path $repo 'agent'
$binDir = Join-Path $desktop 'build\bin'

function Invoke-WailsBuild {
    param([string]$OutputName, [switch]$Debug)

    $buildArgs = @('build', '-o', $OutputName)
    if ($Debug) { $buildArgs += @('-debug', '-devtools') }

    Push-Location $desktop
    try {
        # We deliberately avoid `-clean`: deleting build/bin fails if an editor /
        # indexer (gopls, VS Code, Search) holds a handle on it, common on Windows.
        & wails @buildArgs
        if ($LASTEXITCODE -ne 0) { throw "wails build ($OutputName) failed ($LASTEXITCODE)" }
    }
    finally { Pop-Location }
}

if ($Mode -eq 'prod' -or $Mode -eq 'both') {
    Write-Host '==> Building iad-console.exe (release)...' -ForegroundColor Cyan
    Invoke-WailsBuild -OutputName 'iad-console.exe'
}
if ($Mode -eq 'dev' -or $Mode -eq 'both') {
    Write-Host '==> Building iad-console-dev.exe (debug + devtools)...' -ForegroundColor Cyan
    Invoke-WailsBuild -OutputName 'iad-console-dev.exe' -Debug
}

if (-not $SkipAgent) {
    Write-Host '==> Building iad-agent.exe (scanner) into bundle...' -ForegroundColor Cyan
    Push-Location $agent
    try {
        New-Item -ItemType Directory -Force -Path $binDir | Out-Null
        & go build -o (Join-Path $binDir 'iad-agent.exe') ./cmd/iad-agent
        if ($LASTEXITCODE -ne 0) { throw "agent build failed ($LASTEXITCODE)" }
    }
    finally { Pop-Location }

    # The agent needs the rules/ directory for --classify / --full scans. Ship it
    # next to the binary so resolveRulesDir finds it (exe-dir/rules).
    Write-Host '==> Bundling rules/ ...' -ForegroundColor Cyan
    Copy-Item -Recurse -Force (Join-Path $repo 'rules') (Join-Path $binDir 'rules')
}

Write-Host "==> Done. Bundle in: $binDir" -ForegroundColor Green
Get-ChildItem $binDir -File | Select-Object Name, @{N = 'Size(KB)'; E = { [int]($_.Length / 1KB) } }, LastWriteTime | Format-Table -AutoSize
