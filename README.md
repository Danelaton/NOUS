# NOUS — AI Ecosystem Configurator

> "One command. Any agent. Any OS."

NOUS is a CLI that gives your AI coding agents a **Spec-Driven Development (SDD)** workflow and automatic configuration injection — fully local, zero APIs required.

---

## Quick Start

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.ps1 | iex
```

After installing, verify with:
```bash
nous status
```

---

## Install Methods

| Command | OS | Notes |
|---------|----|-------|
| `curl -fsSL .../install.sh \| bash` | macOS / Linux | Downloads binary from latest GitHub release |
| `irm .../install.ps1 \| iex` | Windows | PowerShell one-liner |

Both scripts auto-detect the latest release and install to:
- **Binary:** `~/.local/bin/nous` (macOS/Linux) or `$env:LOCALAPPDATA\nous\bin\nous.exe` (Windows)
- **Skills:** `~/.nous/skills/`
- **Config:** `~/.nous/config/`

To update: re-run the same install command.

---

## How it works

NOUS installs **once on your machine** and enhances every project you work on:

```
~/.local/bin/nous              ← binary added to your PATH
~/.nous/config/                ← agent configs (injected only for detected agents)
~/.nous/skills/               ← installed skills
~/.agent/skills/              ← skills available to this project
```

NOUS never writes inside your projects unless you explicitly run `nous sync`.

---

## Project Setup

Run **once per project** inside your project directory:

```bash
cd ~/my-project
nous sync        # → creates dev/ + .agent/ + AGENTS.md + skills
```

| Command | What it creates | When to re-run |
|---------|----------------|----------------|
| `nous sync` | `dev/` (6 subdirs) + `.agent/` (memory system) + `AGENTS.md` + skills | First time in a new project |

> Add `dev/` and `.agent/` to your `.gitignore` — they are local working state, not tracked.

---

## Commands

| Command | Scope | Description |
|---------|-------|-------------|
| `nous install` | Global | Detect agents and inject NOUS configuration |
| `nous status` | Global | Show system and detected agents |
| `nous sync` | Global | Sync project structure + re-inject agent configs |

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

## What gets installed

```
~/.local/bin/nous       ← binary added to your PATH (macOS/Linux)
# or
$env:LOCALAPPDATA\nous\bin\nous.exe  ← binary (Windows)

~/.nous/
  config/               ← agent configs, injected only for detected agents
    claude/config.json           ← if Claude Code is installed
    cursor/settings.json         ← if Cursor is installed
    opencode/settings.json       ← if OpenCode is installed
    kiro/config.json            ← if Kiro is installed
    roo/config.json             ← if Roo is installed
```

---

## Features

### Agent Config Injection
- NOUS writes a `config.json` / `settings.json` for each detected agent
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
├── cmd/nous/cli/          # commands: install, status, sync
├── cmd/nous/install/      # orchestrator, detector, skills manager
└── pkg/config/           # per-agent adapters

Distribution
└── GoReleaser            # cross-compile: linux/darwin/windows × amd64/arm64
                           # publishes to GitHub Releases
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

**Release flow** (automatic on push to main)
```bash
git push origin main
# auto-tag.yml generates a timestamped tag → release.yml runs GoReleaser → GitHub Release published
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