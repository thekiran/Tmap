<#
.SYNOPSIS
  Build the IAD desktop app for Windows and bundle the iad-agent scanner beside it.

.DESCRIPTION
  Produces a self-contained folder under build/bin/ that contains both:
    - iad-console.exe  (Wails desktop UI)
    - iad-agent.exe    (the network scanner the UI invokes)
  so the desktop app's "Run Scan" works out of the box (app.go resolveAgentBin
  finds the agent next to the executable).

  Requires: Go 1.22+, Node 18+, and the Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).
#>
$ErrorActionPreference = 'Stop'
$desktop = $PSScriptRoot
$repo = Resolve-Path (Join-Path $desktop '..\..')
$agent = Join-Path $repo 'agent'
$binDir = Join-Path $desktop 'build\bin'

# Build the desktop UI first (overwrites build/bin in place). We deliberately
# avoid `-clean`: deleting build/bin fails if an editor/indexer (gopls, VS Code,
# Search) holds a handle on it, which is common on Windows.
Write-Host '==> Building iad-console (desktop UI)...' -ForegroundColor Cyan
Push-Location $desktop
try {
    & wails build
    if ($LASTEXITCODE -ne 0) { throw "wails build failed ($LASTEXITCODE)" }
} finally { Pop-Location }

Write-Host '==> Building iad-agent (scanner) into bundle...' -ForegroundColor Cyan
Push-Location $agent
try {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    & go build -o (Join-Path $binDir 'iad-agent.exe') ./cmd/iad-agent
    if ($LASTEXITCODE -ne 0) { throw "agent build failed ($LASTEXITCODE)" }
} finally { Pop-Location }

# The agent needs the rules/ directory for --classify / --full scans. Ship it
# next to the binary so resolveRulesDir finds it (exe-dir/rules).
Write-Host '==> Bundling rules/ ...' -ForegroundColor Cyan
Copy-Item -Recurse -Force (Join-Path $repo 'rules') (Join-Path $binDir 'rules')

Write-Host "==> Done. Bundle in: $binDir" -ForegroundColor Green
Get-ChildItem $binDir | Select-Object Name, Length | Format-Table -AutoSize
