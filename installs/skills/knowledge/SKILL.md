---
name: knowledge
description: Compatibility alias for the OKF project knowledge workflow. Use when users ask to ingest, retrieve, consolidate, or preserve project knowledge; routes all durable memory operations to `.agents/OKF/`.
---

# Knowledge

Uses the project-local Open Knowledge Format bundle at `.agents/OKF/`.

This skill preserves the familiar `knowledge` trigger while delegating the storage model and operating rules to the `okf-knowledge` skill.

## When to use

- When asked to remember project information
- When querying prior architecture, decisions, runbooks, or solved problems
- When consolidating duplicate or disconnected project knowledge
- When migrating legacy `.agents/knowledge/`, `.agents/docs_index.md`, or `dev/docs/`

## How to use

1. Read `.agents/skills/okf-knowledge/SKILL.md`.
2. Read `.agents/MEMORY.md`, then `.agents/OKF/index.md`.
3. For ingestion, route the information to the appropriate OKF concept type and nearest index.
4. For queries, follow relevant index links and cite the concepts used.
5. For consolidation, merge overlapping concepts, preserve citations, update indexes, and record a milestone in `.agents/OKF/log.md`.
6. Keep `.agents/MEMORY.md` limited to current work, blockers, next action, and knowledge routes.

## Legacy migration

- Treat `.agents/knowledge/` as historical input, not the active store.
- Do not bulk-copy or delete legacy files.
- Migrate verified durable information when it becomes relevant.
- Preserve provenance with links or citations.
- Treat the migrated OKF concept as the durable source.

## Output

All new durable knowledge must be written under `.agents/OKF/` in OKF v0.1 format.