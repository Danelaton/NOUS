# ADR-001: Core Architecture

## Status
Accepted

## Date
2026-04-14

## Context
NOUS is an AI Ecosystem Configurator CLI that gives coding agents a Spec-Driven Development (SDD) workflow and project structure — with zero external API dependencies.

## Decisions

### 1. Language: Go for CLI Only

**Decision**: Use Go 1.24+ for the entire CLI. No Python runtime bundled.

**Rationale**:
- Go produces single static binaries ideal for CLI distribution
- Multi-platform (linux/darwin/windows × amd64/arm64) with one toolchain
- No external dependencies for the agent runtime
- Fast compilation and startup

### 2. Global ~/.nous/ Installation Model

**Decision**: The NOUS runtime lives in `~/.nous/` globally — never inside any project directory.

```
~/.nous/
  config/           — agent configs injected by nous install
  skills/           — skills source (AGENTS.md, etc.)
```

Agent configs are injected to `~/.nous/config/<agent>/` (claude/config.json, cursor/settings.json, etc.)

**Rationale**:
- One installation serves all projects
- Agent configs point to `~/.nous/` paths
- Project directories remain clean

### 3. Project Structure: openspec/ + dev/ + .agent/

**Decision**: Project directories contain three distinct areas:

| Directory | Tracked | Purpose |
|-----------|---------|---------|
| `openspec/` | No | SDD specs and change proposals |
| `dev/` | No | Working memory: sandbox, scripts, tests, docs, backups |
| `.agent/` | No | Agent memory: MEMORY.md, docs_index.md, skills |
| `docs/` | Yes | ADRs (Architectural Decision Records) |

**Rationale**:
- Clear separation between working state (dev/) and persistent project memory (.agent/)
- `dev/` is explicitly NOT tracked — scratchpad for experiments
- `docs/` is tracked — formal architectural decisions
- `.agent/` persists context across agent sessions

### 4. OpenSpec as SDD Workflow Foundation

**Decision**: NOUS uses OpenSpec structure for Spec-Driven Development:

```
openspec/specs/SPEC.md       — project specification
openspec/changes/           — change proposals (CHG_XXX_proposal.md)
```

The agent must read SPEC.md before writing any code.

**Rationale**:
- Enforces disciplined, spec-first development
- Change proposals create audit trail
- Separates specification from implementation

### 5. Agent Adapters: Pluggable Architecture

**Decision**: Create an adapter interface for each supported agent (OpenCode, Claude Code, Cursor, Kiro, Roo).

Each adapter:
- Detects if the agent is installed (checks `~/.opencode/`, `~/.claude/`, etc.)
- Injects a config file pointing to `~/.nous/` paths
- Sets `openspec.enabled: true` in the agent config

**Rationale**:
- Different agents use different config formats (JSON, TOML, etc.)
- Easy to add new agents without modifying core logic
- Each adapter is independent and testable

### 6. Installation Distribution: Scripts + Binaries

**Decision**: Distribution via shell scripts that bootstrap the binary from GitHub Releases:

```
curl -fsSL https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.sh | bash
irm https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.ps1 | iex
```

The script downloads the pre-built binary from the GitHub Release and runs `nous install`.

**Rationale**:
- One command installation — no build required
- Scripts detect OS/arch and download the correct binary
- Works on any system with curl/PowerShell
- Future: Homebrew tap + Scoop bucket for managed updates

## Consequences

### Positive
- Single static Go binary — no runtime dependencies
- Global installation — one setup, all projects
- Clean project directories — no pollution
- Extensible adapters for new agents
- Zero API dependencies

### Negative
- Agent must be manually configured to read AGENTS.md
- Two installer scripts to maintain (bash + PowerShell)
- Each new agent requires a new adapter

## References
- OpenSpec: https://openspec.dev/
