# ADR_001: Architecture Decisions

## Status
Accepted

## Date
2026-04-14

## Context
Building NOUS, a CLI ecosystem configurator that enhances AI coding agents with persistent memory (MemPalace), Spec-Driven Development workflow (OpenSpec), and automatic agent configuration injection.

## Decisions

### 1. Language: Go for CLI, Python for MemPalace

**Decision**: Use Go 1.24+ for CLI core (multiplatform binaries, fast compilation) and Python 3.9+ for MemPalace integration (existing codebase, rich ecosystem).

**Rationale**:
- Go produces single static binaries ideal for CLI distribution
- MemPalace is Python-native; wrapping avoids reimplementation
- Communication via venv + subprocess or IPC

### 2. MemPalace Integration: Clone and Adapt

**Decision**: Clone github.com/MemPalace/mempalace and adapt with minimal modifications.

**Rationale**:
- MemPalace achieves 96.6% recall (best-in-class for local-only)
- 29 MCP tools already implemented
- Wings/Rooms/Closets/Drawers architecture fits NOUS requirements
- Zero API dependencies (local ChromaDB + SQLite)

### 3. OpenSpec Structure

**Decision**: Replicate OpenSpec structure: `openspec/specs/` and `openspec/changes/`.

**Rationale**:
- Spec-First is core to SDD workflow
- Natural fit with gentle-ai inspiration (Gentleman-Programming/gentle-ai)
- Separating specs from changes enforces disciplined development

### 4. Agent Adapters: Pluggable Architecture

**Decision**: Create adapter interface for each supported agent (OpenCode, Claude Code, Cursor, Kiro, Roo).

**Rationale**:
- Different agents have different config formats (JSON, TOML, etc.)
- Easy to add new agents without modifying core logic
- Each adapter handles detection and injection independently

### 5. Installers: Shell for Unix, PowerShell for Windows

**Decision**: Separate `install.sh` (bash) and `install.ps1` (PowerShell).

**Rationale**:
- Windows-first development environment
- PowerShell has modern symlink support without admin privileges
- Unix systems can use standard shell with same functionality

### 6. Directories: Follow Agent Rules

**Decision**: Use standard directory structure:
- `dev/` - Scratchpad (NOT tracked in git)
- `.agent/skills/<skill-name>/` - Agent skills only
- `docs/` - ADRs only (tracked in git)

**Rationale**: Matches user-specified agent rules to avoid polluting project root.

## Consequences

### Positive
- Clear separation of concerns (CLI vs Memory vs Workflow)
- Easy to test individual components
- Extensible for new agents
- All storage local (no external API dependencies)

### Negative
- Python venv adds complexity for Windows users
- Two installer scripts to maintain
- Adapter implementations needed for each agent

## References
- gentle-ai: https://github.com/Gentleman-Programming/gentle-ai
- MemPalace: https://github.com/MemPalace/mempalace
- OpenSpec: https://openspec.dev/