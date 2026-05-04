---
name: architecture-review
description: Analyzes a project's architecture using PROJECT_MAP.md and produces a structured ARCHITECTURE_REVIEW.md with strengths, issues, and prioritized recommendations. Use when evaluating code quality, preparing for a refactor, or identifying structural problems like excessive coupling, shallow modules, circular dependencies, or missing separation of concerns.
---

# Architecture Review

Analyzes project architecture and produces actionable recommendations in ARCHITECTURE_REVIEW.md. Requires PROJECT_MAP.md or generates it first via the `project-map` skill.

## When to use

- When evaluating the quality of an existing codebase before a major feature or refactor
- When onboarding and need to understand structural risks
- When something "feels wrong" but the problem isn't clearly localized
- When preparing a technical proposal or ADR that touches multiple modules
- When asked "what are the biggest structural problems here?"

## How to use

Follow the steps below. This skill uses LLM reasoning to analyze — extraction was already done by `project-map`. Do not re-scan the file system unless PROJECT_MAP.md is missing or stale.

## Steps

### Step 1 — Load context

```bash
# Check if PROJECT_MAP.md exists
cat PROJECT_MAP.md 2>/dev/null
```

If PROJECT_MAP.md does not exist or is older than 7 days, run the `project-map` skill first, then continue.

Also load:
```bash
cat AGENTS.md 2>/dev/null         # project conventions
cat .agent/MEMORY.md 2>/dev/null  # past decisions
ls docs/ADR_*.md 2>/dev/null      # architectural decisions on record
```

### Step 2 — Analyze for common issues

For each category below, evaluate based on PROJECT_MAP.md + targeted file reads. Be specific — cite module names, file paths, and line ranges.

#### 2a. Shallow modules
A module is shallow when its public interface is nearly as complex as its implementation — it adds little abstraction value.

Look for:
- Wrapper classes/functions that do nothing but delegate
- Modules with 1–2 functions exported that are called from only one place
- Utility folders (`utils/`, `helpers/`) that have grown too large

#### 2b. Excessive coupling
Modules that depend on too many others become fragile — one change ripples everywhere.

```bash
# Measure fan-in / fan-out via import counts
# For Go:
grep -r '"<module-path>' --include="*.go" -l 2>/dev/null | wc -l

# For TS/JS:
grep -r "from '.*<module>" --include="*.ts" --include="*.js" -l 2>/dev/null | wc -l
```

Flag modules with fan-out > 8 or fan-in > 15 as candidates.

#### 2c. Circular dependencies

```bash
# Go: use go list
go list -f '{{.ImportPath}}: {{.Imports}}' ./... 2>/dev/null | grep circular || true

# JS/TS: look for obvious cycles via grep
# (full cycle detection requires madge or similar — note if unavailable)
```

#### 2d. Dead code and duplication

```bash
# Find unexported / unexposed symbols (Go)
grep -r "^func [a-z]" --include="*.go" -h 2>/dev/null | sort | uniq -c | sort -rn | head -20

# Find duplicated patterns (any language)
grep -r "TODO\|FIXME\|HACK\|XXX" --include="*.go" --include="*.ts" --include="*.py" \
  -n 2>/dev/null | head -20
```

#### 2e. Separation of concerns

Review top-level module responsibilities from PROJECT_MAP.md:
- Are HTTP handlers mixed with business logic?
- Does persistence code leak into service layer?
- Are config values hardcoded inside domain code?

#### 2f. Test coverage gaps

```bash
# Check test file ratio
total=$(find . -name "*.go" -o -name "*.ts" -o -name "*.py" 2>/dev/null | grep -v test | wc -l)
tests=$(find . -name "*_test.*" -o -name "*.test.*" -o -name "*.spec.*" 2>/dev/null | wc -l)
echo "Source: $total | Tests: $tests"

# Find untested modules (no corresponding test file)
find . -name "*.go" -not -name "*_test.go" -not -path "*/vendor/*" 2>/dev/null | \
  while read f; do
    base="${f%.*}"
    ls "${base}_test.go" 2>/dev/null || echo "UNTESTED: $f"
  done | head -20
```

### Step 3 — Evaluate strengths

Before listing problems, identify what the architecture does well:
- Consistent module boundaries
- Good separation of concerns in specific areas
- Effective use of interfaces/abstractions
- Well-tested critical paths
- Clear dependency direction

### Step 4 — Prioritize issues

Rank each issue found using this matrix:

| Priority | Criteria |
|----------|---------|
| P0 | Breaks correctness or makes the system unmaintainable now |
| P1 | Causes frequent bugs or significant friction for development |
| P2 | Technical debt that compounds over time |
| P3 | Nice to have, low urgency |

### Step 5 — Generate ARCHITECTURE_REVIEW.md

Write to project root with this structure:

```markdown
# ARCHITECTURE_REVIEW

Generated: <date>
Based on: PROJECT_MAP.md (<date of project map>)
Reviewer: NOUS architecture-review skill

## Summary

<2–3 sentences: overall health, main risk, one concrete recommendation>

## Strengths

- <specific strength with example>
- <specific strength with example>

## Issues Found

### P0 — <Issue title>
**What:** <description>
**Where:** `path/to/file.go:line` or `module-name/`
**Why it matters:** <concrete impact — bugs, friction, blocked features>
**How to fix:** <specific steps, not generic advice>

### P1 — <Issue title>
...

### P2 — <Issue title>
...

## Recommendations

Ordered by impact:

1. **<Action>** — `<file or module>` — <what to do and why>
2. **<Action>** — ...

## Priority Actions

The 3 things to do first:

- [ ] <Action 1> — estimated effort: <S/M/L>
- [ ] <Action 2> — estimated effort: <S/M/L>
- [ ] <Action 3> — estimated effort: <S/M/L>
```

## Validation checklist

Before saving ARCHITECTURE_REVIEW.md:

- [ ] PROJECT_MAP.md was read before starting analysis
- [ ] Every issue cites a specific file, module, or path — no generic claims
- [ ] Strengths section is honest and specific (not just filler)
- [ ] Priority Actions are concrete and executable, not vague
- [ ] Recommendations do not propose breaking the current architecture without justification
- [ ] ARCHITECTURE_REVIEW.md written to project root

## Notes

- This skill uses LLM reasoning for analysis — extraction commands are a guide, not a substitute for reading the actual code
- Do not recommend rewrites unless the current structure is genuinely unmaintainable
- If a circular dependency is detected, note it but do not assume it's always a problem (sometimes intentional)
- Revisit ARCHITECTURE_REVIEW.md after major refactors — it goes stale quickly
- Keep recommendations specific to this project's stack and conventions (read from PROJECT_MAP.md)