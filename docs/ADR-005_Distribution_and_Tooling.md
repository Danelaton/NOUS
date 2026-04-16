# ADR-005: Distribution and Tooling

## Status
Accepted

## Date
2026-04-14

## Context
NOUS must be installable with a single command on macOS, Linux, and Windows. Distribution must be reliable, update-friendly, and require no build step from the user.

## Decisions

### 1. Bootstrap Installation: One Command

**Decision**: The primary installation method is a single command that bootstraps everything:

```
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash

# Windows
irm https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.ps1 | iex
```

The script downloads the pre-built binary from GitHub Releases, places it in PATH, and runs `nous install`.

**Rationale**:
- Zero build dependencies for the user
- Works immediately after download
- Detects OS/arch automatically

### 2. Binary Distribution via GitHub Releases

**Decision**: Pre-built binaries are distributed via GitHub Releases.

Targets:
| OS | Arch | File |
|----|------|------|
| Linux | amd64 | `nous_vVERSION_linux_amd64.tar.gz` |
| Linux | arm64 | `nous_vVERSION_linux_arm64.tar.gz` |
| macOS | amd64 | `nous_vVERSION_darwin_amd64.tar.gz` |
| macOS | arm64 | `nous_vVERSION_darwin_arm64.tar.gz` |
| Windows | amd64 | `nous_vVERSION_windows_amd64.zip` |
| Windows | arm64 | `nous_vVERSION_windows_arm64.zip` |

**Rationale**:
- GitHub Releases provides reliable, permanent storage
- Checksums verify integrity
- No CDN or external hosting dependency

### 3. GoReleaser for Build Automation

**Decision**: GoReleaser handles all build and release automation.

**Workflow** (`.github/workflows/release.yml`):
```yaml
on:
  push:
    tags:
      - 'v*'
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - uses: goreleaser/goreleaser-action@v6
```

On `git push tag vX.Y.Z.W`:
1. Compiles for all 6 targets
2. Uploads to GitHub Release
3. Updates Homebrew tap formula
4. Updates Scoop bucket manifest

**Rationale**:
- Single source of truth: git tag
- Automates all manual release steps
- No release engineer intervention

### 4. Homebrew Tap (macOS/Linux)

**Decision**: `Danelaton/homebrew-tap` hosts the Homebrew formula.

Install:
```bash
brew tap nous-cli/tap
brew install nous
```

GoReleaser writes `Formula/nous.rb` on each release.

**Rationale**:
- Homebrew manages updates: `brew upgrade nous`
- Standard macOS package manager
- One command install with update management

### 5. Scoop Bucket (Windows)

**Decision**: `Danelaton/scoop-bucket` hosts the Scoop manifest.

Install:
```powershell
scoop bucket add nous-cli https://github.com/Danelaton/scoop-bucket
scoop install nous
```

GoReleaser writes `nous.json` on each release.

**Rationale**:
- Scoop manages updates: `scoop update nous`
- Standard Windows package manager
- Portable (no admin rights required)

### 6. nous install — Runtime Setup

**Decision**: After downloading the binary, `nous install` runs to set up the runtime:

1. Create `~/.nous/config/` (agent configs)
2. Copy `AGENTS.md` to `~/.nous/skills/`
3. Detect installed agents
4. Inject configs to `~/.nous/config/<agent>/`

### 7. Project Setup: nous sdd-init + nous sync

**Decision**: Project setup is two commands:

```bash
nous sdd-init   # Creates openspec/specs/ + openspec/changes/
nous sync       # Creates dev/ + copies AGENTS.md + injects configs
```

**Rationale**:
- Separation: SDD workflow (sdd-init) vs. full setup (sync)
- User chooses when to activate NOUS in a project
- No automatic project modification on install

## Consequences

### Positive
- One-command bootstrap install
- Multi-platform (linux/darwin/windows × amd64/arm64)
- Homebrew and Scoop for managed updates
- GoReleaser automates entire release pipeline
- Project setup is explicit and opt-in

### Negative
- Homebrew/Scoop require repos to be pre-created
- GitHub token required for GoReleaser to update taps
- Binary size larger than interpreted alternatives
- Two installer scripts to maintain (bash + PowerShell)
