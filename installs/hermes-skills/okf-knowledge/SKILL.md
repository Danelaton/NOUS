---
name: okf-knowledge
description: Maintain durable project knowledge in OKF v0.1 format.
version: 1.0.0
author: Hermes Agent (adapted from NOUS)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [okf, knowledge, memory, documentation, durable]
    related_skills: [domain-modeling, memory, openspec]
---

# OKF Project Knowledge

Maintains durable project knowledge using Open Knowledge Format (OKF v0.1). Two-tier architecture:

- **Global**: `~/.hermes/OKF/` — cross-project knowledge library (Hermes reads this)
- **Local**: `.agents/OKF/` — per-project reference (other agents, `nous sync`)

## When to use

- At session start: read `.agents/MEMORY.md` → `~/.hermes/OKF/index.md` → follow relevant links
- After discovering architecture, constraints, or conventions
- After making a durable decision or solving a difficult problem
- After verifying setup, test, deploy, or rollback commands
- When prior project context is needed

## Session start protocol

1. Read `.agents/MEMORY.md`
2. Read `~/.hermes/OKF/index.md` (global catalog)
3. Read `~/.hermes/OKF/<project>/index.md` (project bundle)
4. Follow only links relevant to the current task
5. Use Hermes `memory` tool for frequently-needed facts

Do not load the entire bundle unless the task requires it.

## Persistence routing

| Knowledge | Destination |
|-----------|-------------|
| Current task, blockers, next action | `.agents/MEMORY.md` |
| Cross-project conventions, skills, tools | `~/.hermes/OKF/_system/` |
| Project architecture and boundaries | `~/.hermes/OKF/<project>/architecture.md` |
| Durable decision and rationale | `~/.hermes/OKF/<project>/decisions/` |
| Verified operational procedure | `~/.hermes/OKF/<project>/workflows/` |
| Diagnosed failure and verified fix | `~/.hermes/OKF/<project>/troubleshooting/` |
| Frequently-needed facts | Hermes `memory` tool |
| Major knowledge milestone | `~/.hermes/OKF/log.md` (global) |

## Concept format (OKF v0.1)

Every non-reserved markdown document must begin with YAML frontmatter containing a non-empty `type`:

```yaml
---
type: Decision
title: Use PostgreSQL advisory locks
description: Coordinates singleton jobs without another service.
tags: [database, concurrency]
timestamp: 2026-07-21T12:00:00Z
---
```

## Reserved files

- `index.md` — progressive-disclosure directory listing
- `log.md` — newest-first history under `YYYY-MM-DD` headings
- Only the bundle-root `index.md` may declare `okf_version: "0.1"`

## Write protocol

1. Search relevant indexes before creating a file
2. Update existing concept when it represents the same knowledge
3. Otherwise create `kebab-case.md` file
4. Link from nearest `index.md`
5. Add citations for externally sourced claims
6. Add entry to `~/.hermes/OKF/log.md` for significant changes
7. Use Hermes `memory` tool for compact, frequently-accessed facts

## Hermes Integration

- `read_file` to load OKF concepts
- `write_file` / `patch` to create/update concepts
- `search_files` to find existing knowledge
- `memory` tool for session-persistent facts (auto-injected every turn)
- `session_search` for historical conversation context
