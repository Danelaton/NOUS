# NOUS — AI Skills Installer

> "One command. Any project. Any agent."

NOUS is a local-first CLI that equips coding agents with reusable skills and durable project memory — one command, zero required APIs, no vendor lock-in.

## Why NOUS

- **Permanent project context** — a concise memory router plus an OKF v0.1 knowledge bundle survive across agent sessions.
- **Progressive disclosure** — agents load the project catalog first and follow only task-relevant links instead of flooding their context window.
- **Non-destructive sync** — repeated `nous sync` runs update instructions and skills without overwriting project knowledge.
- **Portable knowledge** — architecture, decisions, runbooks, troubleshooting, and references remain human-readable Markdown.
- **Any coding agent** — behavior is distributed through `AGENTS.md` and self-contained skills rather than a proprietary runtime.
- **Local by default** — `.agents/` stays in the project workspace and is ignored by Git unless a team explicitly chooses to share it.

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

Both scripts download the latest binary and install predefined skills to `~/.nous/skills/`.

To update: re-run the same command.

---

## What gets installed

```
~/.local/bin/nous              ← binary (macOS/Linux)
$env:LOCALAPPDATA\nous\bin\nous.exe  ← binary (Windows)

~/.nous/
  skills/                      ← predefined skills (downloaded from GitHub)
    AGENTS.md
    adk-memory-agent/
    architecture-review/
    knowledge/                 ← compatibility alias for OKF
    okf-knowledge/             ← durable project knowledge workflow
    opencode-memory/
    project-map/
    skill-creator/
```

---

## Commands

| Command | Scope | Description |
|---------|-------|-------------|
| `nous install` | Global | Initialize `~/.nous/` directory structure |
| `nous status` | Global | Show NOUS status and installed skills |
| `nous sync` | Project | Setup project: `dev/` + `.agents/OKF/` + `AGENTS.md` + skills |
| `nous skills` | Project | Install/update skills from `~/.nous/skills/` into `.agents/skills/` |

### Project setup

Run **once per project** inside your project directory:

```bash
cd ~/my-project
nous sync
```

`nous sync` creates:
- `dev/` — local working state (sandbox, tmp-repos, docs, scripts, tests, backups)
- `.agents/` — local agent state: concise `MEMORY.md`, durable `OKF/` knowledge bundle, and `skills/`
- `AGENTS.md` — agent identity and workflow instructions (project root)

> Add `dev/` and `.agents/` to your `.gitignore` — they are local state, not tracked.

### Durable project memory

```text
.agents/
├── MEMORY.md                 ← active context and routes
├── OKF/
│   ├── index.md              ← progressive-disclosure catalog
│   ├── log.md                ← major knowledge milestones
│   ├── architecture.md       ← verified system structure
│   ├── decisions/            ← durable decisions and rationale
│   ├── workflows/            ← verified runbooks
│   ├── troubleshooting/      ← diagnosed failures and fixes
│   └── references/           ← curated sources
└── skills/
```

Every non-reserved OKF concept is Markdown with YAML frontmatter and a required `type`. Existing knowledge is preserved when `nous sync` runs again.

---

## Included Skills

| Skill | What it does |
|-------|-------------|
| `adk-memory-agent` | Google ADK memory agent integration guide |
| `architecture-review` | Architecture review workflow for AI agents |
| `knowledge` | Compatibility alias that routes memory operations to OKF |
| `okf-knowledge` | Durable project knowledge using Open Knowledge Format v0.1 |
| `opencode-memory` | OpenCode memory plugin implementation guide |
| `project-map` | Project structure mapping and analysis |
| `skill-creator` | Guide for creating new agent skills |

---

## Architecture

```
NOUS CLI (Go 1.24)
├── cmd/nous/cli/          # commands: install, status, sync, skills
├── cmd/nous/install/      # orchestrator, detector, skills manager
└── pkg/config/            # per-agent adapters

Distribution
└── GoReleaser             # cross-compile: linux/darwin/windows × amd64/arm64
                            # publishes to GitHub Releases
```

---

## Development

```bash
git clone https://github.com/Danelaton/NOUS
cd NOUS

go build -o nous ./cmd/nous    # or nous.exe on Windows
./nous status
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
| [`docs/ADR-001_Core_Architecture.md`](./docs/ADR-001_Core_Architecture.md) | Core architecture decisions |
| [`docs/ADR-002_Memory_System.md`](./docs/ADR-002_Memory_System.md) | Memory system design |
| [`docs/ADR-003_Document_Knowledge_System.md`](./docs/ADR-003_Document_Knowledge_System.md) | Document knowledge layers |
| [`docs/ADR-004_AGENTS_System_Prompt.md`](./docs/ADR-004_AGENTS_System_Prompt.md) | System prompt structure |
| [`docs/ADR-005_Distribution_and_Tooling.md`](./docs/ADR-005_Distribution_and_Tooling.md) | Distribution and tooling |
| [`docs/ADR-006_OKF_Project_Knowledge.md`](./docs/ADR-006_OKF_Project_Knowledge.md) | Durable project knowledge with OKF v0.1 |

---

## License

MIT
