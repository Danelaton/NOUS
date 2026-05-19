# NOUS — AI Skills Installer

> "One command. Any project. Any agent."

NOUS is a CLI that installs AI agent skills into your projects — fully local, zero APIs required.

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
    knowledge/
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
| `nous sync` | Project | Setup project: `dev/` + `.agent/` + `AGENTS.md` + skills |
| `nous skills` | Project | Install/update skills from `~/.nous/skills/` into `.agent/skills/` |

### Project setup

Run **once per project** inside your project directory:

```bash
cd ~/my-project
nous sync
```

`nous sync` creates:
- `dev/` — local working state (sandbox, tmp-repos, docs, scripts, tests, backups)
- `.agent/` — agent memory system (MEMORY.md, docs_index.md, skills/)
- `AGENTS.md` — agent identity and workflow instructions (project root)

> Add `dev/` and `.agent/` to your `.gitignore` — they are local state, not tracked.

---

## Included Skills

| Skill | What it does |
|-------|-------------|
| `adk-memory-agent` | Google ADK memory agent integration guide |
| `architecture-review` | Architecture review workflow for AI agents |
| `knowledge` | Knowledge base and documentation patterns |
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

---

## License

MIT
