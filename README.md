# NOUS — AI Ecosystem Configurator

> "One command. Any agent. Any OS."

NOUS is a CLI that gives your AI coding agents **persistent memory**, a
**Spec-Driven Development (SDD)** workflow, and curated **skills** — all fully
local, zero APIs, zero cloud dependencies.

---

## How it works

NOUS installs **once on your machine** and enhances every project you work on:

```
curl | bash / irm | iex
        │
        ▼
~/.nous/                     ← global runtime, never touches your projects
  venv/                      ← mempalace + chromadb (installed from PyPI via uv)
  hooks/                     ← auto-save hooks for every supported agent
  config/                    ← agent configs, injected only for agents detected

        │  (opt-in, per project)
        ▼
cd ~/my-project && nous sdd-init
  openspec/
    specs/SPEC.md            ← write your spec before coding
    changes/CHG_001.md       ← propose changes here
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

All methods install the same thing: binary + `~/.nous/` runtime.

---

## What gets installed

```
~/.local/bin/nous       ← binary added to your PATH

~/.nous/
  venv/                 ← isolated Python env (mempalace 3.x, chromadb 1.x)
  hooks/
    mempal_save_hook.sh          ← triggered by agent Stop event
    mempal_save_hook.ps1
    mempal_precompact_hook.sh    ← triggered before context compression
    mempal_precompact_hook.ps1
  config/
    claude/config.json           ← if Claude Code is installed
    cursor/settings.json         ← if Cursor is installed
    opencode/settings.json       ← if OpenCode is installed
    kiro/config.json             ← if Kiro is installed
    roo/config.json              ← if Roo is installed
  skills/                        ← global skill registry
```

**Nothing is placed in your projects** until you explicitly run `nous sdd-init`.

---

## Project Setup

Run these **once per project** inside your project directory:

| Command | What it creates | When to re-run |
|---------|----------------|----------------|
| `nous sdd-init` | `openspec/specs/SPEC.md` + `openspec/changes/` | First time in a new project |
| `nous skill-registry` | Scans project and registers conventions in MemPalace | After changing stack or dependencies |

```bash
cd ~/my-project
nous sdd-init
# → creates openspec/ here, nothing else
```

> `openspec/` is the only thing NOUS ever writes inside a project. Add it to your
> `.gitignore` if you prefer not to track it, or commit it as part of your workflow.

---

## Commands

| Command | Scope | Description |
|---------|-------|-------------|
| `nous install` | Global (`~/.nous/`) | Install or repair the NOUS runtime |
| `nous status` | Global | Show runtime health, mempalace version, detected agents |
| `nous sync` | Global | Re-inject agent configs (run after installing a new agent) |
| `nous sdd-init` | Project | Create `openspec/` in the current directory |
| `nous skill-registry` | Project | Scan project conventions into MemPalace |
| `nous profile --name NAME` | Global | Switch OpenCode model routing profile |

---

## Supported Agents

NOUS detects agents by their config directories and injects its configuration automatically.

| Agent | Detected via | Config written to |
|-------|-------------|-------------------|
| Claude Code | `~/.claude/` | `~/.nous/config/claude/config.json` |
| Cursor | `~/.cursor/` | `~/.nous/config/cursor/settings.json` |
| OpenCode | `~/.opencode/` | `~/.nous/config/opencode/settings.json` |
| Kiro | `~/.kiro/` | `~/.nous/config/kiro/config.json` |
| Roo | `~/.roo/` | `~/.nous/config/roo/config.json` |

Run `nous sync` after installing a new agent to inject its config.

---

## Features

### Persistent Memory — MemPalace
- Installed from PyPI (`pip install mempalace`) into an isolated venv at `~/.nous/venv/`
- Uses **ChromaDB** for semantic search and **SQLite** for persistent storage — 100% local
- **29 MCP tools** available to any agent: search, store, knowledge graph, navigation
- **Auto-save hooks** fire on agent Stop and PreCompact events — zero context loss
- 96.6% recall rate across sessions

### Spec-Driven Development (SDD)
- Enforces spec-first discipline: no code without a spec
- `openspec/specs/` — project specifications
- `openspec/changes/` — change proposals (`CHG_XXX_proposal.md`)
- Agents reference the spec before every implementation step

### Agent Config Injection
- NOUS writes a `config.json` / `settings.json` for each detected agent
- Points the agent to `~/.nous/venv/` for MemPalace access
- Registers the auto-save hooks with the agent's event system

### Model Routing Profiles (OpenCode)
Select a profile to control which model handles each SDD phase:

| Profile | Design | Implementation | Verification |
|---------|--------|---------------|-------------|
| `fast` | Haiku | Haiku | Haiku |
| `balanced` | Sonnet | Sonnet | GPT-4o |
| `quality` | Opus | Sonnet | GPT-4o |

```bash
nous profile --name balanced
```

---

## Architecture

```
NOUS CLI (Go 1.24)
├── cmd/nous/cli/          # cobra commands: install, status, sync, sdd-init, skill-registry, profile
├── cmd/nous/install/      # orchestrator, agent detection, openspec generator, skill registry
├── pkg/config/            # per-agent adapters (inject config files)
├── pkg/memory/            # MemPalace wrapper + MCP client + wake-up prompt
└── pkg/workflow/          # SDD workflow manager

MemPalace (Python, PyPI package)
├── ChromaDB               # semantic vector search — fully local
├── SQLite                 # persistent drawer storage — fully local
└── MCP Server             # 29 tools exposed to any MCP-compatible agent

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

**Running tests (when added)**
```bash
go test ./...
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

Required CI secrets: `HOMEBREW_TAP_TOKEN`, `SCOOP_BUCKET_TOKEN`, `GITHUB_TOKEN`.

---

## Documentation

| File | Description |
|------|-------------|
| [`AGENTS.md`](./AGENTS.md) | Full agent identity, wake-up prompt, AAAK dialect, MCP tools reference |
| [`docs/ADR_001_Architecture_Decisions.md`](./docs/ADR_001_Architecture_Decisions.md) | Architecture decision records |
| [`installs/install.sh`](./installs/install.sh) | macOS/Linux installer source |
| [`installs/install.ps1`](./installs/install.ps1) | Windows installer source |

---

## License

MIT
