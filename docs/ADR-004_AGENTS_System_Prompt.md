# ADR-004: AGENTS.md System Prompt

## Status
Accepted

## Date
2026-04-14

## Context
NOUS agents need a consistent identity, rules, and memory protocol across all sessions. The system prompt must be self-contained, unambiguous, and govern all agent behavior.

## Decisions

### 1. AGENTS.md as Static System Prompt

**Decision**: `AGENTS.md` is the single source of truth for agent behavior. It is:

- Stored in `~/.nous/skills/AGENTS.md` (global installation)
- Copied to every project by `nous sync`
- Loaded by the agent at the start of every session

**Agent Identity**: NOUS — Autonomous Systems Architect & DevOps Expert

### 2. AGENTS.md Sections

**Decision**: AGENTS.md is organized into numbered, mandatory sections:

| # | Section | Purpose |
|---|---------|---------|
| 1 | Identity & Role | Name, role, reasoning model |
| 2 | Communication Protocol | Language rules (Spanish reports, English code) |
| 3 | Directory Topology | Mandatory structure (dev/, openspec/, .agent/, docs/) |
| 4 | OpenSpec & SDD Workflow | SPEC-first protocol |
| 5 | Memory System | MEMORY.md, auto-persist, AAAK dialect |
| 6 | AAAK Dialect | Encoding rules for memory compression |
| 7 | Operational Protocols | Git safety, backup/rollback, pre-mutation protocol |
| 8 | Security Standards | uv only, .env secrets, SSL, OWASP sanitization |
| 9 | Updates & Reports | Formal reporting structure |
| 10 | Document Knowledge System | docs/ + dev/docs/ integration |

### 3. Self-Governance Without Intervention

**Decision**: The agent updates MEMORY.md and docs_index.md **automatically without asking**.

- Every 15 messages: evaluate if context needs recording
- Session end: update Session Log + last_updated
- New entity mentioned: encode and add to MEMORY.md

The agent is trusted to maintain its own memory. Asking for permission to update memory would break flow.

### 4. Communication Protocol

**Decision**: Language rules are strict and non-negotiable:

- **Spanish (es)**: Explanations, reports, user-facing communication
- **English (en)**: Code, comments, git logs, technical documentation
- **No preamble**: Direct communication, no fluff

### 5. Pre-Mutation Backup Protocol

**Decision**: Before editing any file outside `dev/sandbox/`, the agent creates a backup.

**Format**: `dev/backups/YYYYMMDD_HHMMSS_filename.ext`

**Process**:
1. Create backup before edit
2. Register in Thought (ReAct reasoning)
3. If failure detected → propose rollback with diff
4. User approves rollback before execution

### 6. Git Safety: No Silent Mutations

**Decision**: git commit and git push require explicit user approval.

1. Show `git diff`
2. Wait for "YES" from user
3. Only then execute mutation

### 7. Forbidden Actions

**Decision**: These are explicitly forbidden and non-negotiable:

| Forbidden | Reason |
|-----------|--------|
| Modifying tracked ADR | Create new ADR instead |
| Deleting `dev/` files without logging | Institutional knowledge loss |
| Hardcoding credentials | Security violation |
| Skipping ADR review before architectural advice | Repeating past mistakes |
| Using pip instead of uv | Dependency consistency |
| Context abandonment | Memory loss |

## Consequences

### Positive
- Single source of truth for all agent behavior
- Self-governing memory — no user intervention needed
- Consistent agent behavior across all sessions
- Strict safety rules prevent data loss

### Negative
- AGENTS.md must be kept in sync with implemented features
- Agent behavior is governed by rules, not code — relies on enforcement
- Length of AGENTS.md grows with feature additions
