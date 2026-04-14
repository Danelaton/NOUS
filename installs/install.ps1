# NOUS Installer for Windows (PowerShell)
# Usage:
#   irm https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.ps1 | iex
#
# What this installs (all global, nothing in your projects):
#   %LOCALAPPDATA%\nous\bin\nous.exe  — NOUS binary (added to user PATH)
#   ~/.nous/venv/                      — Python venv with mempalace + chromadb
#   ~/.nous/hooks/                     — auto-save hooks for agents
#   ~/.nous/config/                    — agent configs (only detected agents)
#
# To activate SDD in a project (run AFTER install, inside your project):
#   cd C:\my-project; nous sdd-init

$ErrorActionPreference = "SilentlyContinue"

$GITHUB_OWNER = "nous-cli"
$GITHUB_REPO  = "nous"
$NOUS_DIR     = Join-Path $HOME ".nous"
$NOUS_VENV    = Join-Path $NOUS_DIR "venv"
$PYTHON_EXE   = Join-Path $NOUS_VENV "Scripts\python.exe"

function Write-Step  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Cyan }
function Write-Ok    { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Green }
function Write-Warn  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Yellow }
function Write-Err   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Red }
function Write-Dim   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "[NOUS] =================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   NOUS - AI Ecosystem Configurator"               -ForegroundColor Cyan
Write-Host "[NOUS] =================================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================================
# PHASE 1: Install NOUS binary
# ============================================================================
Write-Step "Phase 1/6: Installing NOUS binary..."

$NOUS_INSTALLED = $false

function Install-ViaScoop {
    if (-not (Get-Command scoop -ErrorAction SilentlyContinue)) { return $false }
    Write-Dim "Installing via Scoop..."
    scoop bucket add nous-cli https://github.com/nous-cli/scoop-bucket 2>$null
    scoop install nous 2>$null
    return [bool](Get-Command nous -ErrorAction SilentlyContinue)
}

function Install-ViaBinary {
    $arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $tmp  = Join-Path $env:TEMP "nous-install-$([System.IO.Path]::GetRandomFileName())"
    New-Item -ItemType Directory -Path $tmp -Force | Out-Null

    try {
        $release = Invoke-RestMethod "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases/latest" -UseBasicParsing
        $version = $release.tag_name
    } catch {
        Write-Warn "Could not fetch latest release"
        Remove-Item -Recurse -Force $tmp
        return $false
    }

    $ver     = $version -replace '^v', ''
    $archive = "nous_${ver}_windows_${arch}.zip"
    $url     = "https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases/download/$version/$archive"
    $zipPath = Join-Path $tmp $archive

    Write-Dim "Downloading nous $version (windows/$arch)..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
    } catch {
        Write-Warn "Binary download failed"
        Remove-Item -Recurse -Force $tmp
        return $false
    }

    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    $installDir = Join-Path $env:LOCALAPPDATA "nous\bin"
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    $exeSrc = Get-ChildItem $tmp -Filter "nous.exe" -Recurse | Select-Object -First 1 -ExpandProperty FullName
    if (-not $exeSrc) {
        Write-Warn "nous.exe not found in archive"
        Remove-Item -Recurse -Force $tmp
        return $false
    }
    Copy-Item $exeSrc (Join-Path $installDir "nous.exe") -Force

    # Add to user PATH permanently
    $currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$installDir*") {
        [System.Environment]::SetEnvironmentVariable("PATH", "$installDir;$currentPath", "User")
    }
    $env:PATH = "$installDir;$env:PATH"

    Remove-Item -Recurse -Force $tmp
    Write-Ok "nous $version installed to $installDir"
    return $true
}

function Install-ViaGo {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Err "Go not found — install from https://go.dev/dl/"
        return $false
    }
    $gobin = Join-Path $env:LOCALAPPDATA "nous\bin"
    New-Item -ItemType Directory -Path $gobin -Force | Out-Null
    $env:GOBIN = $gobin
    go install "github.com/$GITHUB_OWNER/$GITHUB_REPO/cmd/nous@latest" 2>$null
    $env:PATH = "$gobin;$env:PATH"
    return [bool](Get-Command nous -ErrorAction SilentlyContinue)
}

if (Install-ViaScoop)  { $NOUS_INSTALLED = $true; Write-Ok "Installed via Scoop" }
if (-not $NOUS_INSTALLED -and (Install-ViaBinary)) { $NOUS_INSTALLED = $true }
if (-not $NOUS_INSTALLED -and (Install-ViaGo))     { $NOUS_INSTALLED = $true }

if (-not $NOUS_INSTALLED) {
    Write-Err "All install methods failed."
    Write-Err "Download manually from: https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases"
    exit 1
}

# ============================================================================
# PHASE 2: Install uv
# ============================================================================
Write-Host ""
Write-Step "Phase 2/6: Installing uv (Python package manager)..."

$UV_PATH = $null
$uvCmd   = Get-Command uv -ErrorAction SilentlyContinue
$uvLocal = Join-Path $HOME ".local\bin\uv.exe"

if ($uvCmd) {
    $UV_PATH = $uvCmd.Source
    Write-Dim "uv already installed: $UV_PATH"
} elseif (Test-Path $uvLocal) {
    $UV_PATH = $uvLocal
    Write-Dim "uv found at $UV_PATH"
} else {
    Write-Dim "Downloading uv..."
    Invoke-Expression (Invoke-WebRequest -Uri https://astral.sh/uv/install.ps1 -UseBasicParsing).Content
    $UV_PATH = $uvLocal
}

$uvDir = [System.IO.Path]::GetDirectoryName($UV_PATH)
if ($env:PATH -notlike "*$uvDir*") { $env:PATH = "$uvDir;$env:PATH" }

$uvVersion = & $UV_PATH --version 2>&1
Write-Ok "uv ready: $uvVersion"

# ============================================================================
# PHASE 3: Create ~/.nous/ structure
# ============================================================================
Write-Host ""
Write-Step "Phase 3/6: Creating ~/.nous/ structure..."

@($NOUS_DIR, (Join-Path $NOUS_DIR "config"), (Join-Path $NOUS_DIR "hooks"), (Join-Path $NOUS_DIR "skills")) |
    ForEach-Object { if (-not (Test-Path $_)) { New-Item -ItemType Directory -Path $_ -Force | Out-Null } }
# Note: no mempalace\ subdir — mempalace lives inside the venv

Write-Ok "~/.nous/ ready"

# ============================================================================
# PHASE 4: Create venv + install mempalace from PyPI
# ============================================================================
Write-Host ""
Write-Step "Phase 4/6: Installing mempalace into ~/.nous/venv/ ..."

# Create venv if python not present
if (-not (Test-Path $PYTHON_EXE)) {
    Write-Dim "Creating virtual environment..."
    & $UV_PATH venv $NOUS_VENV --python python --quiet 2>$null
}

# Check if mempalace already installed
$mpCheck = & $PYTHON_EXE -c "import mempalace" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Dim "mempalace already installed — skipping"
} else {
    Write-Dim "Running: uv pip install mempalace"
    & $UV_PATH pip install mempalace --python $PYTHON_EXE --quiet 2>$null
}

# Verify
$MP_VERSION     = (& $PYTHON_EXE -m mempalace --version 2>&1 | Select-Object -First 1) -join ""
$CHROMA_VERSION = (& $PYTHON_EXE -c "import chromadb; print(chromadb.__version__)" 2>&1) -join ""

if ($LASTEXITCODE -eq 0 -or $MP_VERSION) {
    Write-Ok "mempalace: $MP_VERSION"
} else {
    Write-Warn "mempalace CLI not responding — run: nous status"
}
if ($CHROMA_VERSION) { Write-Dim "chromadb: $CHROMA_VERSION" }

# ============================================================================
# PHASE 5: Write auto-save hooks
# ============================================================================
Write-Host ""
Write-Step "Phase 5/6: Writing auto-save hooks..."

$psHookSave = @"
# NOUS auto-save hook - triggered by agent on Stop
`$python = "$PYTHON_EXE"
if (Test-Path `$python) { & `$python -m mempalace save --checkpoint 2>`$null }
"@

$psHookPrecompact = @"
# NOUS pre-compact hook - triggered before context compression
`$python = "$PYTHON_EXE"
if (Test-Path `$python) { & `$python -m mempalace save --emergency 2>`$null }
"@

$bashHookSave = "#!/bin/bash`n# NOUS auto-save hook`n`"$($PYTHON_EXE -replace '\\','/')`" -m mempalace save --checkpoint 2>/dev/null || true"
$bashHookPrecompact = "#!/bin/bash`n# NOUS pre-compact hook`n`"$($PYTHON_EXE -replace '\\','/')`" -m mempalace save --emergency 2>/dev/null || true"

$psHookSave       | Set-Content (Join-Path $NOUS_DIR "hooks\mempal_save_hook.ps1")       -Encoding UTF8
$psHookPrecompact | Set-Content (Join-Path $NOUS_DIR "hooks\mempal_precompact_hook.ps1") -Encoding UTF8
$bashHookSave     | Set-Content (Join-Path $NOUS_DIR "hooks\mempal_save_hook.sh")        -Encoding UTF8
$bashHookPrecompact | Set-Content (Join-Path $NOUS_DIR "hooks\mempal_precompact_hook.sh") -Encoding UTF8

Write-Ok "Hooks installed in $NOUS_DIR\hooks\"

# ============================================================================
# PHASE 6: Run nous install to inject agent configs
# ============================================================================
Write-Host ""
Write-Step "Phase 6/6: Configuring detected agents..."

$nousExe = Get-Command nous -ErrorAction SilentlyContinue
if ($nousExe) {
    nous install 2>$null
    if ($LASTEXITCODE -ne 0) { Write-Warn "Agent config injection skipped — run 'nous sync' manually" }
} else {
    Write-Warn "nous not in PATH yet — restart shell then run 'nous sync'"
}

# ============================================================================
# Summary
# ============================================================================
$nousPath = (Get-Command nous -ErrorAction SilentlyContinue)?.Source ?? "restart shell to activate"

Write-Host ""
Write-Host "[NOUS] =================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   NOUS Installation Complete"                      -ForegroundColor Cyan
Write-Host "[NOUS] =================================================" -ForegroundColor Cyan
Write-Host ("[NOUS]   {0,-20} {1}" -f "nous binary:", $nousPath)     -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "runtime dir:", $NOUS_DIR)     -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "mempalace:", $MP_VERSION)     -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "chromadb:", $CHROMA_VERSION)  -ForegroundColor Green
Write-Host ""
Write-Host "[NOUS]   Next steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "[NOUS]   cd C:\my-project     # go to any project"
Write-Host "[NOUS]   nous sdd-init        # activate SDD workflow there (creates openspec/)"
Write-Host "[NOUS]   nous status          # verify installation"
Write-Host ""
Write-Host "[NOUS]   Restart PowerShell for PATH changes to take effect" -ForegroundColor Gray
Write-Host ""
