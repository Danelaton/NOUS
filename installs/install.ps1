# NOUS Installer for Windows (PowerShell)
# Usage:
#   irm https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.ps1 | iex
#
# What this installs:
#   $env:LOCALAPPDATA\nous\bin\nous.exe — NOUS binary
#   ~/.nous/skills/                     — predefined skills
#
# To use:
#   nous sync        # setup project (dev/ + .agents/OKF/ + AGENTS.md + skills)
#   nous skills      # install skills into current project

$ErrorActionPreference = "SilentlyContinue"

$GITHUB_OWNER = "Danelaton"
$GITHUB_REPO = "NOUS"
$NOUS_DIR = Join-Path $HOME ".nous"
$SKILLS_DIR = Join-Path $NOUS_DIR "skills"
$INSTALL_DIR = Join-Path $env:LOCALAPPDATA "nous\bin"

# Auto-detect latest tag from GitHub API
$VERSION = try {
    $release = Invoke-RestMethod "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases/latest" -UseBasicParsing
    $release.tag_name
} catch {
    Write-Host "[NOUS] Could not fetch latest release — using default version" -ForegroundColor Yellow
    "v2026.5.19"
}

function Write-Step  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Cyan }
function Write-Ok    { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Green }
function Write-Warn  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Yellow }
function Write-Err   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Red }
function Write-Dim   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   NOUS — AI Skills Installer"                      -ForegroundColor Cyan
Write-Host "[NOUS]   Version: $VERSION"                               -ForegroundColor Cyan
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================================
# PHASE 1: Remove previous installation (clean upgrade)
# ============================================================================
Write-Step "Phase 1/4: Removing previous installation..."

$oldExe = Join-Path $INSTALL_DIR "nous.exe"
if (Test-Path $oldExe) {
    Remove-Item -Force $oldExe
    Write-Dim "Previous binary removed"
} else {
    Write-Dim "No previous binary found"
}

# Also clean up any go-installed binary in the same dir
$goInstalled = Get-Command nous -ErrorAction SilentlyContinue
if ($goInstalled -and $goInstalled.Source -ne $oldExe) {
    Write-Dim "Previous binary at $($goInstalled.Source) will be replaced"
}

# ============================================================================
# PHASE 2: Install NOUS binary
# ============================================================================
Write-Host ""
Write-Step "Phase 2/4: Installing NOUS binary..."

$NOUS_INSTALLED = $false

function Install-Binary {
    $arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $tmp  = Join-Path $env:TEMP "nous-install-$([System.IO.Path]::GetRandomFileName())"
    New-Item -ItemType Directory -Path $tmp -Force | Out-Null

    $ver = $VERSION -replace '^v', ''
    $archive = "nous_${ver}_windows_${arch}.zip"
    $url = "https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases/download/$VERSION/$archive"
    $zipPath = Join-Path $tmp $archive

    Write-Dim "Downloading nous $VERSION (windows/$arch)..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
    } catch {
        Write-Warn "Binary download failed"
        Remove-Item -Recurse -Force $tmp
        return $false
    }

    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null

    $exeSrc = Get-ChildItem $tmp -Filter "nous.exe" -Recurse | Select-Object -First 1 -ExpandProperty FullName
    if (-not $exeSrc) {
        Write-Warn "nous.exe not found in archive"
        Remove-Item -Recurse -Force $tmp
        return $false
    }
    Copy-Item $exeSrc (Join-Path $INSTALL_DIR "nous.exe") -Force

    $currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$INSTALL_DIR*") {
        [System.Environment]::SetEnvironmentVariable("PATH", "$INSTALL_DIR;$currentPath", "User")
    }
    $env:PATH = "$INSTALL_DIR;$env:PATH"

    Remove-Item -Recurse -Force $tmp
    Write-Ok "nous $VERSION installed to $INSTALL_DIR"
    return $true
}

function Install-Go {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Err "Go not found — install from https://go.dev/dl/"
        return $false
    }
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    $env:GOBIN = $INSTALL_DIR
    go install "github.com/Danelaton/NOUS/cmd/nous@$VERSION" 2>$null
    $env:PATH = "$INSTALL_DIR;$env:PATH"
    return [bool](Get-Command nous -ErrorAction SilentlyContinue)
}

$NOUS_INSTALLED = Install-Binary
if (-not $NOUS_INSTALLED) { $NOUS_INSTALLED = Install-Go }

if (-not $NOUS_INSTALLED) {
    Write-Err "All install methods failed."
    Write-Err "Download manually from: https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases"
    exit 1
}

# ============================================================================
# PHASE 3: Download skills to ~/.nous/skills/
# ============================================================================
Write-Host ""
Write-Step "Phase 3/4: Downloading skills to ~/.nous/skills/..."

New-Item -ItemType Directory -Path $SKILLS_DIR -Force | Out-Null

# Download AGENTS.md
$AGENTS_URL = "https://raw.githubusercontent.com/$GITHUB_OWNER/$GITHUB_REPO/main/installs/skeleton/AGENTS.md"
try {
    Invoke-WebRequest -Uri $AGENTS_URL -OutFile (Join-Path $SKILLS_DIR "AGENTS.md") -UseBasicParsing
    Write-Ok "AGENTS.md downloaded"
} catch {
    Write-Warn "Could not download AGENTS.md — skipping"
}

# Download skills folders from installs/skills/
function Install-SkillsFolder($repoPath, $destDir) {
    $apiUrl = "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/contents/$repoPath"
    try {
        $items = Invoke-RestMethod $apiUrl -UseBasicParsing
        foreach ($item in $items) {
            if ($item.type -eq "dir") {
                $subDest = Join-Path $destDir $item.name
                New-Item -ItemType Directory -Path $subDest -Force | Out-Null
                Install-SkillsFolder $item.path $subDest
            } elseif ($item.type -eq "file") {
                $fileUrl = "https://raw.githubusercontent.com/$GITHUB_OWNER/$GITHUB_REPO/main/$($item.path)"
                $destPath = Join-Path $destDir $item.name
                try {
                    Invoke-WebRequest -Uri $fileUrl -OutFile $destPath -UseBasicParsing
                } catch { }
            }
        }
    } catch { }
}

Install-SkillsFolder "installs/skills" $SKILLS_DIR
Write-Ok "Skills downloaded to $SKILLS_DIR"

# ============================================================================
# PHASE 4: Summary
# ============================================================================
$nousExe = Get-Command nous -ErrorAction SilentlyContinue
$nousPath = if ($nousExe) { $nousExe.Source } else { "restart shell to activate" }

Write-Host ""
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   Installation Complete"                           -ForegroundColor Cyan
Write-Host "[NOUS]   Version: $VERSION"                               -ForegroundColor Cyan
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host ("[NOUS]   {0,-20} {1}" -f "nous binary:", $nousPath)     -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "skills:", $SKILLS_DIR)        -ForegroundColor Green
Write-Host ""
Write-Host "[NOUS]   Usage:" -ForegroundColor Cyan
Write-Host ""
Write-Host "[NOUS]   cd C:\my-project"
Write-Host "[NOUS]   nous sync            # setup project: dev/ + .agents/OKF/ + AGENTS.md + skills"
Write-Host "[NOUS]   nous skills          # install/update skills in current project"
Write-Host ""
Write-Host "[NOUS]   Restart PowerShell for PATH changes to take effect" -ForegroundColor Gray
Write-Host ""
