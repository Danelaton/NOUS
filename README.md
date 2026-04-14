# NOUS — AI Ecosystem Configurator

> "One command. Any agent. Any OS."

NOUS is a CLI that gives your AI coding agents a **Spec-Driven Development (SDD)** workflow and automatic configuration injection — fully local, zero APIs required.

---

## How it works

NOUS installs **once on your machine** and enhances every project you work on:

```
curl | bash / irm | iex
        │
        ▼
~/.local/bin/nous       ← binary added to your PATH
~/.nous/config/         ← agent configs (injected only for detected agents)

        │  (opt-in, per project)
        ▼
cd ~/my-project && nous sdd-init
  openspec/
    specs/SPEC.md            ← write your spec before coding
    changes/CHG_001.md      ← propose changes here
```

Your projects stay clean. NOUS never writes inside them unless you explicitly run `nous sdd-init`.

---

## Quick Start

### macOS / Linux

**Homebrew (recommended — manages updates)**
```bash
brew tap nous-cli/tap
brew install nous
```

**One-liner**
```bash
curl -fsSL https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.sh | bash
```

### Windows

**Scoop (recommended — manages updates)**
```powershell
scoop bucket add nous-cli https://github.com/nous-cli/scoop-bucket
scoop install nous
```

**PowerShell one-liner**
```powershell
irm https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.ps1 | iex
```

After either method, verify with:
```bash
nous status
```

---

## Install Methods Compared

| Method | OS | Updates | Requires |
|--------|----|---------|---------|
| `brew install nous` | macOS / Linux | `brew upgrade nous` | Homebrew |
| `scoop install nous` | Windows | `scoop update nous` | Scoop |
| `curl \| bash` | macOS / Linux | Re-run the curl command | curl |
| `irm \| iex` | Windows | Re-run the irm command | PowerShell 5+ |

---

## What gets installed

```
~/.local/bin/nous       ← binary added to your PATH

~/.nous/
  config/               ← agent configs, injected only for detected agents
    claude/config.json           ← if Claude Code is installed
    cursor/settings.json         ← if Cursor is installed
    opencode/settings.json       ← if OpenCode is installed
    kiro/config.json            ← if Kiro is installed
    roo/config.json             ← if Roo is installed
```

**Nothing is placed in your projects** until you explicitly run `nous sdd-init`.

---

## Project Setup

Run **once per project** inside your project directory:

```bash
cd ~/my-project
nous sdd-init
# → creates openspec/ here, nothing else
```

| Command | What it creates | When to re-run |
|---------|----------------|----------------|
| `nous sdd-init` | `openspec/specs/SPEC.md` + `openspec/changes/` | First time in a new project |

> Add `openspec/` to your `.gitignore` if you prefer not to track it, or commit it as part of your SDD workflow.

---

## Commands

| Command | Scope | Description |
|---------|-------|-------------|
| `nous install` | Global | Detect agents and inject NOUS configuration |
| `nous status` | Global | Show system and detected agents |
| `nous sync` | Global | Re-inject agent configurations |
| `nous sdd-init` | Project | Create `openspec/` in the current directory |

---

## Supported Agents

NOUS detects agents automatically and injects its configuration.

| Agent | Detected via | Config written to |
|-------|-------------|-------------------|
| Claude Code | `~/.claude/` | `~/.nous/config/claude/config.json` |
| Cursor | `~/.cursor/` | `~/.nous/config/cursor/settings.json` |
| OpenCode | `~/.opencode/` | `~/.nous/config/opencode/settings.json` |
| Kiro | `~/.kiro/` | `~/.nous/config/kiro/config.json` |
| Roo | `~/.roo/` | `~/.nous/config/roo/config.json` |

Run `nous install` after installing a new agent to inject its config.

---

## Features

### Spec-Driven Development (SDD)
- Enforces spec-first discipline: no code without a spec
- `openspec/specs/` — project specifications
- `openspec/changes/` — change proposals (`CHG_XXX_proposal.md`)
- Agents reference the spec before every implementation step

### Agent Config Injection
- NOUS writes a `config.json` / `settings.json` for each detected agent
- Configures the agent to recognize and use the `openspec/` structure
- Works with whatever model the agent already has configured

---

## Architecture

```
NOUS CLI (Go 1.24)
├── cmd/nous/cli/          # cobra commands: install, status, sync, sdd-init
├── cmd/nous/install/      # orchestrator, detector, openspec generator
└── pkg/config/           # per-agent adapters

Distribution
├── Homebrew tap           # nous-cli/homebrew-tap — Formula/nous.rb
├── Scoop bucket           # nous-cli/scoop-bucket — nous.json
└── GoReleaser             # cross-compile: linux/darwin/windows × amd64/arm64
```

---

## Development

```bash
git clone https://github.com/nous-cli/nous
cd nous

go build -o nous ./cmd/nous    # or nous.exe on Windows
./nous status
./nous install
```

**Release flow** (maintainers only)
```bash
git tag v0.x.0 -m "release: v0.x.0"
git push origin v0.x.0
# GoReleaser CI builds all targets and publishes:
#   → GitHub Releases (binaries + checksums)
#   → nous-cli/homebrew-tap (Formula auto-updated)
#   → nous-cli/scoop-bucket (manifest auto-updated)
```

---

## Documentation

| File | Description |
|------|-------------|
| [`AGENTS.md`](./AGENTS.md) | Agent identity, SDD protocol, AAAK dialect, configuration reference |
| [`docs/ADR_001_Architecture_Decisions.md`](./docs/ADR_001_Architecture_Decisions.md) | Architecture decision records |

---

## License

MIT
