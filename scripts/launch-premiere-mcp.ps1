<#
    Launch Premiere Pro with a project open, then start the backend services.

    You usually do not need this script. Quitting Premiere normally and
    reopening a project works on its own: the extension registers as Trusted,
    the panel auto-starts (its manifest listens for ApplicationActivate) and
    ts-bridge connects by itself. All this script really adds is doing those
    steps in one command, plus one recovery step.

    The recovery step is clearing orphaned CEPHtmlEngine processes. If Premiere
    is force-killed or crashes, its CEP host processes can survive; the next
    launch then fails extension registration with "Signature verification
    failed" and the panel is unavailable for the whole session -- including
    from the Extensions menu, which is what makes it look intermittent. A
    stale cep_cache directory is NOT the cause; verified by launching with the
    cache intact and no orphans, which registers as Trusted every time.

    A project must be open for the panel to stay alive: CEP unloads the
    extension on the welcome screen, and ApplicationActivate only fires once
    at startup, so there is no second auto-start. Hence -Project.
#>

[CmdletBinding()]
param(
    # Project to open. Strongly recommended: CEP unloads the extension while
    # Premiere sits on the welcome screen, so the panel dies within seconds of
    # starting if no project is open. Defaults to the last one used.
    [string]$Project,
    # Skip starting the rust/python/ts-bridge services.
    [switch]$NoServices,
    # Kill a running Premiere so the cache clear can take effect.
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$repoRoot   = Split-Path -Parent $PSScriptRoot
$premiere   = 'C:\Program Files\Adobe\Adobe Premiere Pro 2026\Adobe Premiere Pro.exe'
$extensionId = 'com.premierpro.mcp.bridge'

function Write-Step($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-Warn($msg) { Write-Host "    $msg" -ForegroundColor Yellow }

# --- 1. PlayerDebugMode must be a string, not a DWORD ---------------------
# CEP reads this as REG_SZ on Windows; a DWORD is ignored and the extension
# falls back to a signature check it cannot pass.
Write-Step 'Checking PlayerDebugMode'
foreach ($v in 10, 11, 12, 13) {
    $key = "HKCU:\Software\Adobe\CSXS.$v"
    if (-not (Test-Path $key)) { New-Item -Path $key -Force | Out-Null }
    $existing = Get-ItemProperty -Path $key -Name PlayerDebugMode -ErrorAction SilentlyContinue
    if ($null -eq $existing -or $existing.PlayerDebugMode -ne '1') {
        Remove-ItemProperty -Path $key -Name PlayerDebugMode -ErrorAction SilentlyContinue
        New-ItemProperty -Path $key -Name PlayerDebugMode -Value '1' -PropertyType String -Force | Out-Null
        Write-Warn "CSXS.$v PlayerDebugMode set to REG_SZ 1"
    }
}

# --- 2. Premiere must be closed: registration happens at startup ----------
$running = Get-Process -Name 'Adobe Premiere Pro' -ErrorAction SilentlyContinue
if ($running) {
    if (-not $Force) {
        Write-Warn 'Premiere Pro is already running.'
        Write-Warn 'Extensions register at startup, so quit Premiere and run this'
        Write-Warn 'again (or pass -Force to close it for you).'
        Write-Warn 'Save your work first -- Force does not prompt.'
        exit 1
    }
    Write-Step 'Closing Premiere Pro (-Force)'
    $running | Stop-Process -Force
    Start-Sleep -Seconds 3
}

# --- 3. Clear orphaned CEP host processes ---------------------------------
# This is the actual recovery step. Survivors of a crash or a force-kill make
# the next launch fail extension registration.
$orphans = Get-Process -Name 'CEPHtmlEngine' -ErrorAction SilentlyContinue
if ($orphans) {
    Write-Step "Clearing $($orphans.Count) orphaned CEPHtmlEngine process(es)"
    $orphans | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
}

# The cache is left alone on purpose: it is not the cause, and deleting it
# only forces the panel to rebuild it on the next launch.

# --- 4. Launch Premiere and wait for the panel's WebSocket ---------------
# Open a project along with Premiere where possible. The panel's manifest
# starts it on ApplicationActivate, which fires once at startup -- and CEP
# unloads the extension while no project is open, so launching to the welcome
# screen means the panel appears and then dies, with no second auto-start.
if ($Project) {
    if (-not (Test-Path $Project)) { throw "Project not found: $Project" }
    # Open via the .prproj file association rather than passing the path as an
    # argument to the exe: PowerShell hands Premiere an extended-length
    # "\\?\C:\..." path that way, and Premiere rejects it with "This file path
    # does not exist on disk at this location."
    $projectFile = (Get-Item -LiteralPath $Project).FullName
    Write-Step "Launching Premiere Pro with $(Split-Path -Leaf $projectFile)"
    Start-Process -FilePath $projectFile
} else {
    Write-Step 'Launching Premiere Pro'
    Write-Warn 'No -Project given. Open a project promptly: the panel does not'
    Write-Warn 'survive on the welcome screen.'
    Start-Process $premiere
}

Write-Step 'Waiting for the MCP Bridge panel on port 9801'
$deadline = (Get-Date).AddMinutes(3)
$panelUp = $false
while ((Get-Date) -lt $deadline) {
    if (Get-NetTCPConnection -LocalPort 9801 -State Listen -ErrorAction SilentlyContinue) {
        $panelUp = $true
        break
    }
    Start-Sleep -Seconds 2
}

if (-not $panelUp) {
    Write-Warn 'Panel did not come up within 3 minutes. In order of likelihood:'
    Write-Warn '  1. No project open -- open one, then Window > Extensions >'
    Write-Warn '     PremierPro MCP Bridge (auto-start only fires at launch).'
    Write-Warn "  2. Check $env:LOCALAPPDATA\Temp\CEP12-PPRO.log for"
    Write-Warn '     "Signature verification failed". If present, quit Premiere,'
    Write-Warn '     confirm no CEPHtmlEngine processes remain, and rerun.'
    Write-Warn '  3. Panel opened but rendered blank -- close it and reopen it'
    Write-Warn "     from that menu; if it persists, delete"
    Write-Warn "     $env:LOCALAPPDATA\Temp\cep_cache\PPRO_*_$extensionId."
} else {
    Write-Host '    Panel is listening on 9801.' -ForegroundColor Green
}

# --- 5. Start the backend services --------------------------------------
if ($NoServices) {
    Write-Step 'Skipping services (-NoServices)'
    return
}

Write-Step 'Starting backend services'
& "$PSScriptRoot\start-services-win.bat"
