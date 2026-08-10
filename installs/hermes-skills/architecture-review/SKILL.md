---
name: architecture-review
description: "Use after project-map — analyze PROJECT_MAP.md, produce ARCHITECTURE_REVIEW.md."
version: 1.1.0
author: Hermes Agent (adapted from NOUS)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [architecture, review, code-quality, refactoring, coupling]
    related_skills: [project-map, codebase-design, improve-codebase-architecture]
---

# Architecture Review

Analyzes project architecture using `PROJECT_MAP.md` and produces actionable recommendations in `ARCHITECTURE_REVIEW.md`. Requires `PROJECT_MAP.md` — generates it first via `project-map` if missing or stale (>7 days).

## When to use

- Evaluating codebase quality before a major feature or refactor
- Onboarding and need to understand structural risks
- Something "feels wrong" but the problem isn't clearly localized
- Preparing a technical proposal or ADR touching multiple modules

## Steps

### Step 1 — Load context

Use `read_file` to load:
- `PROJECT_MAP.md`
- `AGENTS.md` (project conventions)
- `.agents/MEMORY.md` (active context)
- `.agents/OKF/index.md` (durable knowledge catalog)
- `docs/ADR_*.md` (architectural decisions)

If `PROJECT_MAP.md` is missing or stale, run `project-map` skill first.

### Step 2 — Analyze for common issues

Use the `codebase-design` skill vocabulary (module, interface, depth, seam, adapter, leverage, locality).

#### Shallow modules
A module is shallow when its interface is nearly as complex as its implementation. Look for:
- Wrapper classes/functions that only delegate
- Modules with 1–2 exports called from only one place
- `utils/`, `helpers/` folders that have grown too large

Apply the **deletion test**: would deleting it concentrate complexity, or just move it?

#### Excessive coupling
Modules depending on too many others become fragile. Measure fan-out via import counts:
```bash
# Go
MODULE=$(head -1 go.mod 2>/dev/null | awk '{print $2}')
[ -n "$MODULE" ] && grep -r "\"$MODULE/" --include="*.go" -l 2>/dev/null | \
  while read f; do count=$(grep -c "\"$MODULE/" "$f"); echo "$count $f"; done | sort -rn | head -20
```

Flag modules with fan-out > 8 or fan-in > 15.

#### Circular dependencies
```bash
go build ./... 2>&1 | grep "import cycle" || echo "No import cycles detected"
```

#### Dead code and duplication
```bash
grep -r "TODO\|FIXME\|HACK\|XXX" --include="*.go" --include="*.ts" -n 2>/dev/null | head -20
```

#### Separation of concerns
- Are HTTP handlers mixed with business logic?
- Does persistence code leak into service layer?
- Are config values hardcoded inside domain code?

#### Test coverage gaps
```bash
total=$(find . -name "*.go" -o -name "*.ts" 2>/dev/null | grep -v test | wc -l)
tests=$(find . -name "*_test.*" -o -name "*.spec.*" 2>/dev/null | wc -l)
echo "Source: $total | Tests: $tests"
```

### Step 3 — Evaluate strengths

Before listing problems, identify what the architecture does well.

### Step 4 — Prioritize issues

| Priority | Criteria |
|----------|---------|
| P0 | Breaks correctness or unmaintainable now |
| P1 | Causes frequent bugs or significant friction |
| P2 | Technical debt that compounds over time |
| P3 | Nice to have, low urgency |

### Step 5 — Generate ARCHITECTURE_REVIEW.md

Write to project root:

```markdown
# ARCHITECTURE_REVIEW

Generated: <date>
Based on: PROJECT_MAP.md

## Summary
<2–3 sentences: overall health, main risk, one concrete recommendation>

## Strengths
- <specific strength with example>

## Issues Found
### P0 — <Issue>
**What:** <description>
**Where:** `path/to/file.go:line`
**Why it matters:** <impact>
**How to fix:** <specific steps>

## Recommendations
1. **<Action>** — `<file>` — <what to do>

## Priority Actions
- [ ] <Action 1> — effort: S/M/L
- [ ] <Action 2>
- [ ] <Action 3>
```

## Validation

- [ ] `PROJECT_MAP.md` was read before starting
- [ ] Every issue cites a specific file, module, or path
- [ ] Strengths are honest and specific
- [ ] Priority Actions are concrete and executable
- [ ] Use `codebase-design` vocabulary for architectural findings
- [ ] Update `.agents/OKF/architecture.md` with verified durable findings
