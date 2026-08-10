---
name: openspec
description: "Use when a feature needs spec-first development — proposal.md + tasks.md via OpenSpec."
version: 2.1.0
author: Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [openspec, spec-driven, proposals, tasks, nous, okf, wayfinding, tdd]
    related_skills:
      - nous-agent
      - project-map
      - architecture-review
      - grill-with-docs
      - domain-modeling
      - wayfinder
      - writing-plans
      - subagent-driven-development
      - test-driven-development
      - codebase-design
      - code-review
      - okf-knowledge
---

# OpenSpec × NOUS — Full Synergy Pipeline

Spec-driven development with OpenSpec CLI (`@fission-ai/openspec` v1.8.0+) fully woven into the NOUS ecosystem. Two profiles, 8 layers, zero loose ends.

## Prerequisites

> **Auto-checked on every load.** If the CLI is missing, Hermes will guide the installation before proceeding.

| Dependency | Required | Check | Install |
|---|---|---|---|
| `openspec` CLI | ≥ v1.8.0 | `openspec --version` | `npm install -g @fission-ai/openspec@latest` |
| Node.js | ≥ 18 | `node --version` | https://nodejs.org |

### Verification Protocol

When this skill loads, Hermes MUST verify the CLI is available before executing any openspec commands:

```bash
# Step 1: Check Node.js
node --version 2>&1 || echo "MISSING"

# Step 2: Check openspec
openspec --version 2>&1 || echo "MISSING"
```

**If Node.js is missing → try auto-install, then fall back:**

| Platform | Auto-install attempt | Fallback |
|---|---|---|
| **Windows** | `winget install OpenJS.NodeJS.LTS` | https://nodejs.org |
| **macOS** | `brew install node` | https://nodejs.org |
| **Linux** | `apt install nodejs` or `dnf install nodejs` | https://nodejs.org |
| **Any** | `nvm install --lts` (if nvm is present) | https://github.com/nvm-sh/nvm |

Protocol:
1. Detect OS and try the platform-specific package manager first
2. If that fails, check for `nvm`, `fnm`, or `volta` and try those
3. If all fail → inform user with the direct download link and stop
4. After successful Node.js install, continue to openspec check

**If Node.js is present but openspec is missing:**
1. Run `npm install -g @fission-ai/openspec@latest`
2. Run `openspec --version` to verify

**If openspec is present but version < 1.8.0:**
1. Run `npm install -g @fission-ai/openspec@latest` to upgrade
2. Run `openspec --version` to verify ≥ 1.8.0

**If everything is present:** proceed with the skill normally. Do not re-verify on subsequent calls in the same session.

## Two Profiles

| Profile | Trigger | Flow | When |
|---------|---------|------|------|
| **Quick** | `/opsx:explore` → `/opsx:propose` → `/opsx:apply` → `/opsx:archive` | OpenSpec nativo, mínima fricción | Cambios pequeños, bien acotados, un solo módulo |
| **Deep** | Pipeline NOUS completo (8 capas) | Máximo control, TDD, 2-stage review, OKF | Features multi-módulo, refactors, decisiones arquitectónicas |

Hermes decide cuál usar según el scope del cambio. El usuario puede forzar con "quick" o "deep".

---

## Layer 0 — NOUS Session Protocol (always)

Antes de cualquier trabajo, aplicar `nous-agent`:

1. Leer `.agents/MEMORY.md` — contexto activo, blockers
2. Leer `~/.hermes/OKF/index.md` — catálogo global
3. Crear backup en `dev/backups/YYYYMMDD_HHMMSS_<file>.ext` antes de mutar archivos fuera de `dev/sandbox/`
4. No commit/push sin "YES" explícito del usuario tras mostrar `git diff`
5. No mutar APIs/Cloud/DB sin plan detallado y aprobación

---

## Layer 1 — Project Map + Architecture Review (deep profile)

Antes de proponer cambios en deep profile, entender el terreno:

```bash
# Cargar skills
skill_view(name='project-map')
skill_view(name='architecture-review')
```

1. Si `PROJECT_MAP.md` no existe o tiene >7 días → generar con `project-map`
2. Si `ARCHITECTURE_REVIEW.md` no existe o tiene >30 días → generar con `architecture-review`
3. Cross-reference: ¿el cambio toca shallow modules? ¿Módulos con fan-out >8?
4. Actualizar `.agents/OKF/architecture.md` con hallazgos duraderos

---

## Layer 2 — Grill with Docs + Domain Modeling (deep profile)

Alinear con el usuario ANTES de abrir cualquier change:

```bash
skill_view(name='grill-with-docs')
skill_view(name='domain-modeling')
```

1. Cargar `CONTEXT.md` si existe — hablar el lenguaje del proyecto
2. Grill al usuario: árbol de diseño, rondas de frontier
3. A medida que los términos cristalizan:
   - Desafiar contra el glosario existente
   - Afilar lenguaje difuso → términos canónicos
   - Actualizar `CONTEXT.md` inline (no batchear)
   - Ofrecer ADRs SOLO cuando: difícil de revertir + sorprendente sin contexto + resultado de un trade-off real
4. Solo proceder cuando el frontier está vacío

---

## Layer 2.5 — Wayfinder (for large work)

Cuando el cambio es demasiado grande para un solo proposal (múltiples módulos, decisiones no resueltas):

```bash
skill_view(name='wayfinder')
```

1. Crear un **map issue** en el tracker (label `wayfinder:map`)
2. Crear child tickets para decisiones especificables (labels `wayfinder:research|prototype|grilling|task`)
3. Resolver tickets uno a uno — cada resolución reduce la niebla
4. Cuando el camino está claro → volver a Layer 3 (OpenSpec change)

---

## Layer 3 — OpenSpec Change (core)

### Quick Profile: `/opsx:*` commands

```bash
# Explorar sin compromiso — el agente lee tu codebase y pesa opciones
/opsx:explore

# Crear proposal + specs + design + tasks automáticamente
/opsx:propose <kebab-case-name>

# Implementar todas las tasks
/opsx:apply

# Archivar → mueve a archive/, actualiza specs principales
/opsx:archive
```

Comandos expandidos (profile `expanded`): `/opsx:new`, `/opsx:continue`, `/opsx:ff`, `/opsx:verify`, `/opsx:bulk-archive`, `/opsx:onboard`

Para activar el perfil expandido:
```bash
openspec config profile expanded
openspec update
```

### Deep Profile: manual + NOUS

Crear el directorio y archivos:

```bash
mkdir -p openspec/changes/<kebab-case-name>
```

#### `proposal.md`

```markdown
# Proposal: <Short title>

> **Change**: `<kebab-case-name>`
> **Status**: Proposed | In Progress | Completed
> **Depends on**: <base commit or change>
> **Estimate**: XS | S | M | L | XL
> **Profile**: deep

## Summary

One paragraph. What changes.

## Motivation

Why. What problem. Link to wayfinder map if applicable.

## Architecture Context

- **Modules touched**: <list from PROJECT_MAP.md>
- **Shallow modules affected**: <list from ARCHITECTURE_REVIEW.md>
- **Fan-out risk**: <none | module names with fan-out >8>

## Scope

### 1. <Change item>
**File**: `path/to/file.tsx`
**Specs impacted**: `openspec/specs/<name>/spec.md`

## Non-goals

What is explicitly NOT included.

## Domain Decisions

- **Term**: definition (from CONTEXT.md alignment)
- **ADR**: `docs/adr/NNNN-title.md` (if created during grilling)
```

#### `tasks.md`

```markdown
# Tasks: <Short title>

> **Change**: `<kebab-case-name>`
> **Profile**: deep

### Task 1: <Descriptive Name>
- [ ] Write failing test: `tests/test_module.py::test_name`
- [ ] Run: `pytest tests/test_module.py::test_name -v` → FAIL
- [ ] Implement: `src/path/to/file.py`
- [ ] Run: `pytest tests/test_module.py::test_name -v` → PASS
- [ ] Run full suite: `pytest tests/ -q` → PASS
- [ ] Commit: `feat: <description>`

### Task 2: ...
```

#### Validate

```bash
openspec validate <change-name>
```

---

## Layer 4 — Writing Plans (deep profile)

```bash
skill_view(name='writing-plans')
```

Expandir `tasks.md` en un plan de implementación bite-sized:

- Cada task = 2-5 minutos
- Paths exactos de archivos
- Código completo (copy-pasteable)
- Comandos exactos con expected output
- DRY, YAGNI, TDD

Guardar en `docs/plans/YYYY-MM-DD-<feature>.md`

---

## Layer 5 — Subagent-Driven Development + TDD (deep profile)

```bash
skill_view(name='subagent-driven-development')
skill_view(name='test-driven-development')
```

Por cada task del plan:

1. **Dispatch implementer subagent** — subagente fresco con contexto completo, sin memoria de la conversación. Sigue TDD estricto:
   - RED: test falla primero
   - GREEN: código mínimo para pasar
   - REFACTOR: limpiar sin romper tests
2. **Spec compliance review** — ¿implementado según spec? ¿nada extra?
3. **Code quality review** — ¿convenciones? ¿edge cases? ¿seguridad?
4. Solo proceder cuando AMBAS revisiones aprueban

```python
delegate_task(
    goal="Implement Task 1: <name> using strict TDD",
    context="""<full task text, paths, code, expected output>""",
)
```

---

## Layer 6 — Code Review + Codebase Design (deep profile)

```bash
skill_view(name='codebase-design')
```

Al final de todas las tasks:

1. **Integration review** — ¿consistencia entre tasks?
2. **Codebase design review** usando vocabulario estándar:
   - ¿Nuevos módulos son **deep** (mucha implementación, poca interfaz)?
   - ¿Se crearon **shallow modules** (pass-through, wrappers)?
   - **Deletion test**: ¿si borro este módulo, la complejidad desaparece o se dispersa?
   - ¿Interfaces expuestas en **seams** correctos?
3. `ARCHITECTURE_REVIEW.md` se actualiza si la estructura cambió significativamente

---

## Layer 7 — OKF Knowledge Persistence (always)

```bash
skill_view(name='okf-knowledge')
```

Al archivar un cambio (`openspec archive`):

1. **Decisiones duraderas** → `~/.hermes/OKF/<project>/decisions/<kebab-case>.md`
   ```yaml
   ---
   type: Decision
   title: <title>
   description: <one-liner>
   tags: [<relevant>]
   timestamp: <ISO>
   ---
   ```
2. **Arquitectura verificada** → actualizar `~/.hermes/OKF/<project>/architecture.md`
3. **Workflows verificados** (comandos de build/test/deploy) → `~/.hermes/OKF/<project>/workflows/`
4. **Troubleshooting** (si hubo errores y se resolvieron) → `~/.hermes/OKF/<project>/troubleshooting/`
5. **Milestone** → `~/.hermes/OKF/log.md` (YYYY-MM-DD, newest first)
6. **Hermes memory** — solo hechos compactos y frecuentemente accedidos

### Persistence routing

| Knowledge | Destination |
|-----------|-------------|
| Current blockers, next action | `.agents/MEMORY.md` |
| Frequently-needed facts | Hermes `memory` tool |
| Architecture & boundaries | `~/.hermes/OKF/<project>/architecture.md` |
| Durable decisions & rationale | `~/.hermes/OKF/<project>/decisions/` |
| Verified runbooks | `~/.hermes/OKF/<project>/workflows/` |
| Diagnosed failures & fixes | `~/.hermes/OKF/<project>/troubleshooting/` |
| Cross-project conventions | `~/.hermes/OKF/_system/conventions.md` |
| Major milestones | `~/.hermes/OKF/log.md` |

### Forbidden

- NUNCA duplicar conocimiento durable entre MEMORY.md y OKF
- NUNCA migrar chatter de sesión no verificado a OKF
- NUNCA saltar verificación de fuente antes de persistir
- NUNCA sobrescribir memoria de proyecto durante sync

---

## Directory Structure (full)

```
project/
├── PROJECT_MAP.md              ← Layer 1: generated by project-map
├── ARCHITECTURE_REVIEW.md      ← Layer 1: generated by architecture-review
├── CONTEXT.md                  ← Layer 2: domain glossary
├── AGENTS.md                   ← AI agent instructions
├── .agents/
│   ├── MEMORY.md               ← Layer 0: active context, blockers
│   └── OKF/
│       ├── index.md
│       ├── architecture.md
│       └── decisions/
├── docs/
│   ├── adr/
│   │   └── NNNN-title.md       ← Layer 2: architectural decisions
│   └── plans/
│       └── YYYY-MM-DD-feature.md ← Layer 4: implementation plans
├── openspec/
│   ├── config.yaml
│   ├── specs/                  ← canonical specs (what IS)
│   │   └── <domain>/spec.md
│   └── changes/                ← active proposals (what SHOULD CHANGE)
│       ├── <change-name>/
│       │   ├── proposal.md
│       │   ├── design.md       ← (opsx:propose auto-generates)
│       │   └── tasks.md
│       └── archive/            ← completed
├── dev/
│   ├── backups/                ← Layer 0: pre-mutation backups
│   └── sandbox/                ← safe experimentation zone
└── src/
```

Global (cross-project):
```
~/.hermes/
├── OKF/
│   ├── index.md                ← global catalog
│   ├── log.md                  ← milestones
│   ├── _system/
│   │   └── conventions.md
│   └── <project>/
│       ├── index.md
│       ├── architecture.md
│       ├── decisions/
│       ├── workflows/
│       └── troubleshooting/
└── skills/
    └── ...
```

---

## The Complete Flow (Deep Profile)

```
Session start (L0)
  │
  ├─ Read .agents/MEMORY.md + ~/.hermes/OKF/index.md
  │
  ├─ Project scan (L1)
  │   ├─ project-map → PROJECT_MAP.md
  │   └─ architecture-review → ARCHITECTURE_REVIEW.md
  │
  ├─ Alignment (L2)
  │   ├─ grill-with-docs → CONTEXT.md updated
  │   └─ domain-modeling → ADRs if qualifying
  │
  ├─ [If large] Wayfinding (L2.5)
  │   └─ wayfinder → map issue + decision tickets
  │
  ├─ Spec (L3)
  │   └─ openspec change → proposal.md + tasks.md
  │
  ├─ Plan (L4)
  │   └─ writing-plans → docs/plans/<plan>.md
  │
  ├─ Implement (L5)
  │   └─ subagent-driven-dev + TDD → per-task subagents
  │
  ├─ Review (L6)
  │   ├─ code-review → integration + quality
  │   └─ codebase-design → depth/seam analysis
  │
  ├─ Archive (L3)
  │   └─ openspec archive → archive/ + specs updated
  │
  └─ Persist (L7)
      └─ okf-knowledge → decisions, architecture, workflows, log
```

---

## The Complete Flow (Quick Profile)

```
/opsx:explore        → AI reads codebase, weighs options
/opsx:propose <name> → generates proposal + specs + design + tasks
/opsx:apply          → implements all tasks
/opsx:archive        → archives + updates specs
                      → L7: minimal OKF (decisions + log only)
```

---

## Core CLI Commands

```bash
openspec init [path]            # Initialize in a project
openspec list                   # List active changes
openspec list --specs           # List all specs
openspec view                   # Interactive dashboard
openspec show <name>            # Show a change or spec
openspec validate <name>        # Validate a change
openspec archive <name>         # Archive completed change → updates specs
openspec status <name>          # Task completion status
openspec update                 # Refresh agent instructions + slash commands
openspec config profile <name>  # Switch profile (default | expanded)
```

---

## Hermes Tool Mapping

| Action | Tool |
|--------|------|
| `openspec` CLI commands | `terminal` |
| Create proposal.md / tasks.md | `write_file` / `patch` |
| Load existing specs | `read_file` on `openspec/specs/` |
| Find specs by name/content | `search_files` |
| Parallel research during wayfinding | `delegate_task` |
| Per-task implementation (deep) | `delegate_task` + `terminal` |
| Session context recall | `session_search` |
| Durable facts | `memory` |
| OKF persistence | `write_file` / `patch` on `~/.hermes/OKF/` |

---

## Profile Selection Heuristic

| Signal | Profile |
|--------|---------|
| Single file, well-understood module | Quick |
| 2-3 files, same module | Quick |
| Multi-module, new abstractions | Deep |
| Architecture decision required | Deep |
| User says "quick" or "fast" | Quick |
| User says "deep", "thorough", "con cuidado" | Deep |
| New project, unfamiliar codebase | Deep (L1 mandatory) |
| Bug fix with clear root cause | Quick |

When uncertain, default to Deep. The overhead is justified by the safety net.

---

## Conventions

- **Schema**: `spec-driven`
- **Language**: propuestas en español, code/docs/commits en inglés
- **Proposal states**: Proposed → In Progress → Completed
- **Task format**: Markdown checkboxes (`- [ ]`) with exact file paths
- **Spec updates**: via `openspec archive`, not manual edits
- **Backups**: `dev/backups/YYYYMMDD_HHMMSS_<file>.ext` before mutating
- **Git**: no commit/push without explicit user "YES" after diff review
- **OKF**: no durable knowledge duplication between MEMORY.md and OKF concepts
