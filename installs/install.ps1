# NOUS Installer for Windows (PowerShell)
# Usage:
#   irm https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.ps1 | iex
#
# What this installs:
#   $env:LOCALAPPDATA\nous\bin\nous.exe — NOUS binary
#   $env:LOCALAPPDATA\nous\skills — skills (AGENTS.md)
#   $HOME\.nous\config — agent configs
#
# To activate a project:
#   cd C:\my-project; nous sdd-init
#   cd C:\my-project; nous sync

$ErrorActionPreference = "SilentlyContinue"

$GITHUB_OWNER = "Danelaton"
$GITHUB_REPO = "NOUS"
$SKILLS_DIR = Join-Path $env:LOCALAPPDATA "nous\skills"
$NOUS_DIR = Join-Path $HOME ".nous"

# Auto-detect latest tag from GitHub API
$VERSION = try {
    $release = Invoke-RestMethod "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases/latest" -UseBasicParsing
    $release.tag_name
} catch {
    Write-Host "[NOUS] Could not fetch latest release — using default version" -ForegroundColor Yellow
    "v2026.4.14-1"
}

function Write-Step  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Cyan }
function Write-Ok    { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Green }
function Write-Warn  { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Yellow }
function Write-Err   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Red }
function Write-Dim   { param($msg) Write-Host "[NOUS] $msg" -ForegroundColor Gray }

Write-Host ""
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   NOUS - AI Ecosystem Configurator"               -ForegroundColor Cyan
Write-Host "[NOUS]   Version: $VERSION"                          -ForegroundColor Cyan
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================================
# PHASE 1: Install NOUS binary
# ============================================================================
Write-Step "Phase 1/5: Installing NOUS binary..."

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

    $installDir = Join-Path $env:LOCALAPPDATA "nous\bin"
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    $exeSrc = Get-ChildItem $tmp -Filter "nous.exe" -Recurse | Select-Object -First 1 -ExpandProperty FullName
    if (-not $exeSrc) {
        Write-Warn "nous.exe not found in archive"
        Remove-Item -Recurse -Force $tmp
        return $false
    }
    Copy-Item $exeSrc (Join-Path $installDir "nous.exe") -Force

    $currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$installDir*") {
        [System.Environment]::SetEnvironmentVariable("PATH", "$installDir;$currentPath", "User")
    }
    $env:PATH = "$installDir;$env:PATH"

    Remove-Item -Recurse -Force $tmp
    Write-Ok "nous $VERSION installed to $installDir"
    return $true
}

function Install-Go {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Err "Go not found — install from https://go.dev/dl/"
        return $false
    }
    $gobin = Join-Path $env:LOCALAPPDATA "nous\bin"
    New-Item -ItemType Directory -Path $gobin -Force | Out-Null
    $env:GOBIN = $gobin
    go install "github.com/Danelaton/NOUS/cmd/nous@$VERSION" 2>$null
    $env:PATH = "$gobin;$env:PATH"
    return [bool](Get-Command nous -ErrorAction SilentlyContinue)
}

if (-not $NOUS_INSTALLED) { Install-Binary; $NOUS_INSTALLED = $? }
if (-not $NOUS_INSTALLED) { Install-Go; $NOUS_INSTALLED = $? }

if (-not $NOUS_INSTALLED) {
    Write-Err "All install methods failed."
    Write-Err "Download manually from: https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases"
    exit 1
}

# ============================================================================
# PHASE 2: Install skills from GitHub
# ============================================================================
Write-Host ""
Write-Step "Phase 2/5: Installing skills..."

New-Item -ItemType Directory -Path $SKILLS_DIR -Force | Out-Null
$AGENTS_URL = "https://raw.githubusercontent.com/$GITHUB_OWNER/$GITHUB_REPO/$VERSION/installs/skeleton/AGENTS.md"
try {
    Invoke-WebRequest -Uri $AGENTS_URL -OutFile (Join-Path $SKILLS_DIR "AGENTS.md") -UseBasicParsing
    Write-Ok "AGENTS.md installed"
} catch {
    Write-Warn "Could not download AGENTS.md — skipping skills"
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
                $fileUrl = "https://raw.githubusercontent.com/$GITHUB_OWNER/$GITHUB_REPO/$VERSION/$($item.path)"
                $destPath = Join-Path $destDir $item.name
                try {
                    Invoke-WebRequest -Uri $fileUrl -OutFile $destPath -UseBasicParsing
                } catch { }
            }
        }
    } catch { }
}

Install-SkillsFolder "installs/skills" $SKILLS_DIR
Write-Ok "Skills folder installed"

# ============================================================================
# PHASE 3: Create ~/.nous/ structure
# ============================================================================
Write-Host ""
Write-Step "Phase 3/5: Creating ~/.nous/ structure..."

$configDir = Join-Path $NOUS_DIR "config"
$nousSkillsDir = Join-Path $NOUS_DIR "skills"
if (-not (Test-Path $configDir)) { New-Item -ItemType Directory -Path $configDir -Force | Out-Null }
if (-not (Test-Path $nousSkillsDir)) { New-Item -ItemType Directory -Path $nousSkillsDir -Force | Out-Null }

if (Test-Path (Join-Path $SKILLS_DIR "AGENTS.md")) {
    Copy-Item (Join-Path $SKILLS_DIR "AGENTS.md") $nousSkillsDir -Force
}
# Copy skill folders (e.g. skill-creator/) to ~/.nous/skills/
Get-ChildItem $SKILLS_DIR -Directory | ForEach-Object {
    $dest = Join-Path $nousSkillsDir $_.Name
    if (Test-Path $dest) { Remove-Item $dest -Recurse -Force }
    Copy-Item $_.FullName $dest -Recurse -Force
}
Write-Ok "~/.nous/ ready"

# ============================================================================
# PHASE 4: Run nous install (detect + inject agent configs)
# ============================================================================
Write-Host ""
Write-Step "Phase 4/5: Detecting agents and configuring..."

$nousExe = Get-Command nous -ErrorAction SilentlyContinue
if ($nousExe) {
    nous install 2>$null
    if ($LASTEXITCODE -ne 0) { Write-Warn "Agent configuration skipped — run 'nous sync' manually" }
} else {
    Write-Warn "nous not in PATH — restart shell then run 'nous sync'"
}

# ============================================================================
# PHASE 5: Summary
# ============================================================================
$nousPath = if ($nousExe) { $nousExe.Source } else { "restart shell to activate" }

Write-Host ""
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host "[NOUS]   NOUS Installation Complete"                      -ForegroundColor Cyan
Write-Host "[NOUS]   Version: $VERSION"                            -ForegroundColor Cyan
Write-Host "[NOUS] ================================================" -ForegroundColor Cyan
Write-Host ("[NOUS]   {0,-20} {1}" -f "nous binary:", $nousPath)   -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "skills:", $SKILLS_DIR)     -ForegroundColor Green
Write-Host ("[NOUS]   {0,-20} {1}" -f "config dir:", $configDir)     -ForegroundColor Green
Write-Host ""
Write-Host "[NOUS]   Next steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "[NOUS]   cd C:\my-project     # go to any project"
Write-Host "[NOUS]   nous sync            # setup dev/ + skills + AGENTS.md in project"
Write-Host ""
Write-Host "[NOUS]   Restart PowerShell for PATH changes to take effect" -ForegroundColor Gray
Write-Host ""
