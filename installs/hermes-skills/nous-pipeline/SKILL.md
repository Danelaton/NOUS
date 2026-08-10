---
name: nous-pipeline
description: "Run NOUS SDD pipeline: grill→openspec→plan→TDD→review."
version: 1.0.0
author: Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [nous, pipeline, sdd, orchestration, delegate]
    related_skills: [grilling, grill-with-docs, openspec, writing-plans, subagent-driven-development, test-driven-development, requesting-code-review, nous-agent]
---

# NOUS Pipeline (SDD Orchestrator)

Umbrella skill that runs the full spec-driven development pipeline in order,
using `delegate_task` subagents and the underlying NOUS skills. Invoke when the
user asks for the complete flow ("pipeline completo", "grill to review",
"nous pipeline") or when a task needs spec-first delivery.

## Pipeline Stages

```
1. GRILL     → align requirements, build CONTEXT.md + ADRs
2. OPENSPEC  → proposal.md + tasks.md in openspec/changes/
3. PLAN      → bite-sized plan (2-5 min per task)
4. EXECUTE   → TDD implementation (RED-GREEN-REFACTOR)
5. REVIEW    → verify diff before push
```

## When To Use

- User asks for the full spec-driven flow on a feature/change.
- A task is complex enough to need requirements alignment before code.
- User says "continua" or "pipeline" — execute the next stage immediately.

## Execution Rules

1. **Load stage skills as you go** — each stage loads its own skill:
   - Stage 1: `grill-with-docs` (or `grilling` for quick alignment)
   - Stage 2: `openspec`
   - Stage 3: `writing-plans`
   - Stage 4: `test-driven-development` (+ `subagent-driven-development` for
     delegation with 2-stage review)
   - Stage 5: `requesting-code-review`
2. **Delegate heavy stages** — use `delegate_task` for research, code review,
   and testing. Always pass full context; subagents have no memory of the
   conversation.
3. **Never skip stages** — each stage's output feeds the next. If the user
   says "continua", proceed to the next stage, don't restart.
4. **Progress disclosure** — report stage transitions briefly (what stage is
   done, what's next), in Spanish for conversation, English for artifacts.
5. **Git safety** — no commit/push without explicit user confirmation after
   showing `git diff`. Hooks blocked: use `--no-verify`.
6. **Backups** — pre-mutation copy to `dev/backups/YYYYMMDD_HHMMSS_filename.ext`
   for edits outside `dev/sandbox/`.

## Stage Handoffs

| Stage | Output | Consumed By |
|-------|--------|-------------|
| GRILL | `CONTEXT.md` + ADRs in `docs/` | OPENSPEC |
| OPENSPEC | `openspec/changes/<id>/proposal.md` + `tasks.md` | PLAN |
| PLAN | plan file with bite-sized tasks | EXECUTE |
| EXECUTE | code + tests (TDD) | REVIEW |
| REVIEW | verified diff, review comments | USER |

## Pitfalls

- **Don't load all stage skills at once** — context bloat. Load per stage.
- **Don't skip grill** for ambiguous tasks — the ADRs save rework later.
- **Subagents need full context** — always include paths, error messages,
  conventions (Spanish conversation / English artifacts) in the `context`
  field of `delegate_task`.
- **TDD is non-negotiable in EXECUTE** — RED-GREEN-REFACTOR, tests before code.
- **REVIEW is not optional** — verify the actual diff (`git diff`) before
  declaring done; never rely on self-reports alone.

## Verification

- After each stage, confirm the output file exists and is non-empty.
- After REVIEW, show the user the diff summary and ask for explicit "YES"
  before any commit/push.
