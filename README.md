# NOUS — AI Ecosystem Configurator

> "One command. Any agent. Any OS."

NOUS is a CLI that gives your AI coding agents a **Spec-Driven Development (SDD)** workflow and automatic configuration injection — fully local, zero APIs required.

---

## How it works

NOUS installs **once on your machine** and enhances every project you work on:

```
curl -fsSL .../install.sh | bash
        │
        ▼
~/.local/bin/nous              ← binary added to your PATH
~/.nous/config/                ← agent configs (injected only for detected agents)

        │  (opt-in, per project)
        ▼
cd ~/my-project && nous sdd-init
  openspec/specs/SPEC.md            ← write your spec before coding
  openspec/changes/CHG_001.md     ← propose changes here
```

Your projects stay clean. NOUS never writes inside them unless you explicitly run `sdd-init`.

---

## Quick Start

### macOS / Linux

**Homebrew (recommended — manages updates)**
```bash
brew tap Danelaton/tap
brew install nous
```

**One-liner**
```bash
curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash
```

### Windows

**Scoop (recommended — manages updates)**
```powershell
scoop bucket add nous-cli https://github.com/Danelaton/scoop-bucket
scoop install nous
```

**PowerShell one-liner**
```powershell
irm https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.ps1 | iex
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

**Nothing is placed in your projects** until you explicitly run `sdd-init`.

---

## Project Setup

Run **once per project** inside your project directory:

```bash
cd ~/my-project
nous sdd-init    # → creates openspec/specs/ + openspec/changes/
nous sync        # → creates dev/ + .agent/ + copies AGENTS.md
```

| Command | What it creates | When to re-run |
|---------|----------------|----------------|
| `nous sdd-init` | `openspec/specs/SPEC.md` + `openspec/changes/` | First time in a new project |
| `nous sync` | `dev/` (6 subdirs) + `.agent/` (memory system) + `AGENTS.md` | First time in a new project |

> Add `dev/` and `.agent/` to your `.gitignore` — they are local working state, not tracked.

---

## Commands

| Command | Scope | Description |
|---------|-------|-------------|
| `nous install` | Global | Detect agents and inject NOUS configuration |
| `nous status` | Global | Show system and detected agents |
| `nous sync` | Global | Sync project structure + re-inject agent configs |
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

### Persistent Memory System
- `.agent/MEMORY.md` — AAAK-encoded index of entities, decisions, and work
- `.agent/docs_index.md` — map of all documentation
- Auto-persist every 15 messages without asking
- Session start: reads MEMORY.md + docs_index.md before anything else

### Document Knowledge System
- `docs/` (tracked) — Architectural Decision Records (ADRs)
- `dev/docs/` (not tracked) — session logs, migrations, troubleshooting

---

## Architecture

```
NOUS CLI (Go 1.24)
├── cmd/nous/cli/          # commands: install, status, sync, sdd-init
├── cmd/nous/install/      # orchestrator, detector, openspec generator
└── pkg/config/           # per-agent adapters

Distribution
├── Homebrew tap           # Danelaton/homebrew-tap — Formula/nous.rb
├── Scoop bucket           # Danelaton/scoop-bucket — nous.json
└── GoReleaser            # cross-compile: linux/darwin/windows × amd64/arm64
```

---

## Development

```bash
git clone https://github.com/Danelaton/NOUS
cd NOUS

go build -o nous ./cmd/nous    # or nous.exe on Windows
./nous status
./nous install
```

**Release flow** (maintainers only)
```bash
git tag v0.x.0
git push origin v0.x.0
# GoReleaser CI builds all targets and publishes:
#   → GitHub Releases (binaries + checksums)
#   → Danelaton/homebrew-tap (Formula auto-updated)
#   → Danelaton/scoop-bucket (manifest auto-updated)
```

---

## Documentation

| File | Description |
|------|-------------|
| [`AGENTS.md`](./AGENTS.md) | Agent identity, SDD protocol, AAAK dialect, memory system |
| [`docs/ADR-001_Core_Architecture.md`](./docs/ADR-001_Core_Architecture.md) | Core architecture decisions |
| [`docs/ADR-002_Memory_System.md`](./docs/ADR-002_Memory_System.md) | Memory system design |
| [`docs/ADR-003_Document_Knowledge_System.md`](./docs/ADR-003_Document_Knowledge_System.md) | Document knowledge layers |
| [`docs/ADR-004_AGENTS_System_Prompt.md`](./docs/ADR-004_AGENTS_System_Prompt.md) | System prompt structure |
| [`docs/ADR-005_Distribution_and_Tooling.md`](./docs/ADR-005_Distribution_and_Tooling.md) | Distribution and tooling |

---

## License

MIT
