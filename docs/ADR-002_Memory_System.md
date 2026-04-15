# ADR-002: Memory System

## Status
Accepted

## Date
2026-04-14

## Context
NOUS agents need persistent memory across sessions without relying on external APIs. The system must be resilient to context loss and enable fast retrieval of past decisions.

## Decisions

### 1. 4-Layer Memory Architecture

**Decision**: Memory is organized in 4 layers, from dense to narrative:

| Layer | File | Type | Tracked | Purpose |
|-------|------|------|---------|---------|
| 1 | `.agent/MEMORY.md` | AAAK index | No | Fast lookup: entities, decisions, work status |
| 2 | `.agent/docs_index.md` | Document map | No | Locate relevant documentation fast |
| 3 | `docs/ADR_*.md` | Narratives | Yes | Formal architectural decisions |
| 4 | `dev/docs/*.md` | Logs/references | No | Technical context, migrations, team |

**Rule**: Search deeper only when needed. Always search Layer 1 first.

### 2. MEMORY.md — Dense AAAK Index

**Decision**: `.agent/MEMORY.md` is the primary memory index, encoded in AAAK dialect.

**Structure**:
```
# NOUS Memory Index

## Meta
last_updated: YYYY-MM-DD HH:MM
session_count: N
agent_version: NOUS v2026.4.14.18

## Entities (CODED — AAAK)
DRIFT  — Driftwood Analytics, TS monorepo, 2024-present
K      — Kai, backend lead
AUTH-MIG — Session auth migration project

## Decisions Log
YYYY-MM-DD: PROJECT → decision made (rationale)

## Current Work
CODENAME | OWNER | %done | short description

## Open Issues
CODENAME | description | blocked by | ETA

## Session Log
YYYY-MM-DD HH:MM: action taken

## Notes (free)
Any context that doesn't fit above
```

**Rationale**:
- AAAK encoding compresses context into dense, retrievable tokens
- Fixed structure enables fast scanning
- Session count and timestamps track continuity

### 3. AAAK Dialect (Abstractive Abbreviated Annotated Knowledge)

**Decision**: All entity references in MEMORY.md use AAAK encoding.

**Format Rules**:
- Project names → CODES: `"Driftwood Analytics"` → `"DRIFT"`
- People → abbreviations: `"Kai"` → `"K"`, `"Priya"` → `"PRI"`
- `→` arrow means "focused on" or "related to"
- `|` separates independent facts

**Example**:
```
English:
  "Kai has been working on the Driftwood Analytics project for 3 months"

AAAK:
  "K|WORK→DRIFT|3mo"
```

**Rationale**:
- Dense format fits in context window
- Every token carries meaning
- Human-readable when retrieved

### 4. Auto-Persist Protocol

**Decision**: The agent updates MEMORY.md automatically every 15 messages without asking.

**Update Triggers**:
| Trigger | Action |
|---------|---------|
| New person mentioned | → Add to Entities with AAAK code |
| Architectural decision made | → Add to Decisions Log with date |
| New task started | → Add to Current Work |
| Milestone completed | → Update % in Current Work |
| Blocker detected | → Add to Open Issues |
| Blocker resolved | → Move from Open Issues to Notes |
| Session ends | → Add to Session Log + update last_updated |

**Rationale**:
- No user intervention required
- Continuous memory accumulation
- 15-message interval balances persistence vs. interruption

### 5. Session Start Protocol

**Decision**: At the start of every session, the agent reads MEMORY.md and docs_index.md before doing anything else.

**Sequence**:
1. READ `.agent/MEMORY.md` completely
2. READ `.agent/docs_index.md`
3. If session_count > 1 → review Session Log from last session
4. Load relevant ADRs into context
5. If Open Issues exist → check if `dev/docs/` has updates

**Rationale**:
- Guarantees continuity across sessions
- Agent always has full context before acting
- No assumption — always retrieve first

### 6. Memory Search Protocol

**Decision**: When uncertain, the agent always searches MEMORY.md before asking the user.

**Sequence**:
1. Read `.agent/MEMORY.md` fully
2. Search for the entity or keyword
3. If found → use that context
4. If NOT found → ask the user, then add to MEMORY.md

**Rule**: NEVER assume. NEVER guess past decisions. ALWAYS search MEMORY.md first.

### 7. docs_index.md — Document Map

**Decision**: `.agent/docs_index.md` maps all documentation for fast location.

**Structure**:
```
# NOUS Document Index

## docs/ (TRACKED — ADRs)
| ADR | Topic | Summary | Date |

## dev/docs/ (NOT TRACKED — Logs & References)
| File | Content | Last Updated |
```

The agent maintains docs_index.md automatically:
- New ADR created → add entry
- New `dev/docs/` file → add entry
- `dev/docs/` file updated → update "last updated"

## Consequences

### Positive
- 4 layers cover both fast recall and deep context
- AAAK encoding maximizes context window efficiency
- Auto-persist ensures no context loss
- Session start protocol guarantees continuity

### Negative
- Agent must self-govern memory discipline (enforced by AGENTS.md rules)
- MEMORY.md can grow large without periodic compaction
- AAAK encoding requires agent training/adherence
