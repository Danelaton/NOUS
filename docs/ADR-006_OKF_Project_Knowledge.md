# ADR-006: Project Knowledge with Open Knowledge Format

## Status

Accepted

## Date

2026-07-21

## Context

NOUS previously used `.agents/MEMORY.md`, `.agents/docs_index.md`, tracked ADRs, and `dev/docs/` as separate memory layers. The model offered continuity but duplicated indexes, encouraged a growing monolithic memory file, and relied on `nous sync` code that overwrote the memory templates on every run.

Open Knowledge Format (OKF) v0.1 provides a portable, human-readable, agent-readable hierarchy of Markdown concepts with YAML frontmatter and progressive-disclosure indexes.

## Decision

### 1. Project-local knowledge bundle

Each project uses `.agents/OKF/` as its durable agent knowledge library:

```text
.agents/
├── MEMORY.md
├── OKF/
│   ├── index.md
│   ├── log.md
│   ├── architecture.md
│   ├── decisions/
│   ├── workflows/
│   ├── troubleshooting/
│   └── references/
└── skills/
```

The bundle is local and ignored by Git by default. Teams may explicitly track or export it if they want shared knowledge.

### 2. MEMORY.md becomes a logical router

`.agents/MEMORY.md` contains only active work, blockers, next actions, and links into `.agents/OKF/`. Durable architecture, decisions, procedures, and solutions belong in OKF concepts.

AAAK remains optional as a compact notation for active state but is no longer the storage model for durable knowledge.

### 3. Progressive disclosure

At session start, agents:

1. Read `.agents/MEMORY.md`.
2. Read `.agents/OKF/index.md`.
3. Follow only links relevant to the task.
4. Verify claims against code, project documents, or cited sources.

Agents do not load the entire knowledge bundle by default.

### 4. OKF v0.1 conformance

Every non-reserved Markdown concept contains parseable YAML frontmatter with a non-empty `type`. `index.md` and `log.md` follow their reserved structures. Only the root `index.md` may include frontmatter to declare `okf_version: "0.1"`.

### 5. Non-destructive synchronization

`nous sync` creates memory and OKF templates only when they do not exist. It never overwrites project memory or OKF concepts.

Existing `.agents/docs_index.md` and `dev/docs/` files are preserved. Agents migrate verified durable knowledge incrementally when it is relevant; NOUS performs no destructive automatic migration.

### 6. Persistence triggers

Agents persist knowledge when they:

- Establish or revise system architecture
- Make a durable technical or product decision
- Verify a setup, test, deployment, migration, or rollback procedure
- Diagnose a failure and verify its solution
- Curate a reference that supports future work

Routine conversation and transient execution output are not persisted.

## Supersedes

This ADR supersedes the active memory architecture and maintenance procedures in ADR-002 and ADR-003. Those ADRs remain immutable historical records.

## Consequences

### Positive

- Memory survives repeated `nous sync` runs.
- Knowledge is portable, diffable, and progressively discoverable.
- The router remains small while the project library can grow.
- Legacy memory can be migrated without data loss.
- Concepts are interoperable with OKF-aware agents and tools.

### Negative

- Agents must maintain indexes and logs consistently.
- Existing projects temporarily contain both legacy and OKF structures.
- Local-by-default knowledge is not shared across machines unless explicitly tracked or exported.

## References

- [Open Knowledge Format v0.1 specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)