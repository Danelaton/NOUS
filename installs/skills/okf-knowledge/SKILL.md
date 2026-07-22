---
name: okf-knowledge
description: Maintains durable project knowledge using Open Knowledge Format (OKF v0.1). Use when resuming a project, recording architecture or decisions, documenting verified runbooks, preserving solved problems, or retrieving context across sessions.
---

# OKF Project Knowledge

Maintains the project-local knowledge bundle in `.agents/OKF/` through progressive disclosure.

## When to use

- At the start of a session or before modifying existing code
- After discovering architecture, constraints, dependencies, or conventions
- After making a durable decision or solving a difficult problem
- After verifying setup, test, deploy, migration, or rollback commands
- When prior project context is needed

## Session start

1. Read `.agents/MEMORY.md`.
2. Read `.agents/OKF/index.md`.
3. Follow only links relevant to the current task.
4. Read referenced source documents when a claim must be verified.
5. If legacy `.agents/docs_index.md` or `dev/docs/` content is relevant, read it and migrate durable knowledge incrementally.

Do not load the entire bundle unless the task requires it.

## Persistence routing

| Knowledge | Destination |
|-----------|-------------|
| Current task, blockers, next action | `.agents/MEMORY.md` |
| System structure and boundaries | `.agents/OKF/architecture.md` |
| Durable decision and rationale | `.agents/OKF/decisions/<slug>.md` |
| Verified operational procedure | `.agents/OKF/workflows/<slug>.md` |
| Diagnosed failure and verified fix | `.agents/OKF/troubleshooting/<slug>.md` |
| Curated external or project reference | `.agents/OKF/references/<slug>.md` |
| Major knowledge milestone | `.agents/OKF/log.md` |

Do not duplicate durable knowledge in `MEMORY.md`.

## Concept format

Every non-reserved markdown document must begin with parseable YAML frontmatter containing a non-empty `type`.

```markdown
---
type: Decision
title: Use PostgreSQL advisory locks
description: Coordinates singleton jobs without introducing another service.
tags: [database, concurrency]
timestamp: 2026-07-21T12:00:00Z
---

# Use PostgreSQL advisory locks

## Context

## Decision

## Rationale

## Consequences

# Citations
```

Recommended fields are `title`, `description`, `resource`, `tags`, and an ISO 8601 `timestamp`. Preserve unknown frontmatter fields when updating a document.

## Reserved files

- `index.md` is a progressive-disclosure directory listing.
- `log.md` is a newest-first history grouped under ISO dates (`YYYY-MM-DD`).
- Reserved files normally have no frontmatter.
- Only the bundle-root `.agents/OKF/index.md` may use frontmatter to declare `okf_version: "0.1"`.

## Write protocol

1. Search the relevant index and concepts before creating a file.
2. Update an existing concept when it represents the same knowledge.
3. Otherwise create a lowercase kebab-case filename.
4. Add or update the concept link in the nearest `index.md`.
5. Use bundle-relative links beginning with `/` when practical; relative links are also valid.
6. Add citations for externally sourced claims.
7. Add a concise entry to `.agents/OKF/log.md` for significant additions, corrections, deprecations, or migrations.
8. Update `.agents/MEMORY.md` only with current state and the route to durable knowledge.

## Legacy migration

- Never delete `.agents/docs_index.md` or `dev/docs/` automatically.
- Migrate information only when it becomes relevant or is verified.
- Preserve source history through links or citations.
- Do not copy low-value session chatter into OKF.
- Once migrated, treat the OKF concept as the durable source and the legacy file as historical input.

## Validation

- Every non-reserved `.md` has YAML frontmatter.
- Every concept has a non-empty `type`.
- Every new concept is linked from an `index.md`.
- `log.md` dates use `YYYY-MM-DD` and newest entries come first.
- Links resolve when targets exist; broken links are tolerated but should be intentional.
- Facts are verified, concise, and not duplicated.