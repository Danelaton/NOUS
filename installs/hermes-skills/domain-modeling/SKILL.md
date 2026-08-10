---
name: domain-modeling
description: "Use when domain terms are fuzzy — build and sharpen the project glossary and ADRs."
version: 1.1.0
author: Hermes Agent (adapted from Matt Pocock/skills)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [domain-modeling, glossary, adr, context, ubiquitous-language]
    related_skills: [grilling, grill-with-docs, codebase-design]
---

# Domain Modeling

Actively build and sharpen the project's domain model as you design and implement. This is the *active* discipline — challenging terms, inventing edge-case scenarios, and writing the glossary and decisions down the moment they crystallise. Merely *reading* `CONTEXT.md` for vocabulary is not this skill — that's a one-line habit any skill can do. This skill is for when you're changing the model, not just consuming it.

## File structure

Most repos have a single context:

```
/
├── CONTEXT.md
├── docs/
│   └── adr/
│       ├── 0001-event-sourced-orders.md
│       └── 0002-postgres-for-write-model.md
└── src/
```

If a `CONTEXT-MAP.md` exists at the root, the repo has multiple contexts. The map points to where each one lives. Create files lazily — only when you have something to write. If no `CONTEXT.md` exists, create one when the first term is resolved. If no `docs/adr/` exists, create it when the first ADR is needed.

## During the session

### Challenge against the glossary

When the user uses a term that conflicts with the existing language in `CONTEXT.md`, call it out immediately. "Your glossary defines 'cancellation' as X, but you seem to mean Y — which is it?"

### Sharpen fuzzy language

When the user uses vague or overloaded terms, propose a precise canonical term. "You're saying 'account' — do you mean the Customer or the User? Those are different things."

### Discuss concrete scenarios

When domain relationships are being discussed, stress-test them with specific scenarios. Invent scenarios that probe edge cases and force the user to be precise about the boundaries between concepts.

### Cross-reference with code

When the user states how something works, check whether the code agrees. If you find a contradiction, surface it: "Your code cancels entire Orders, but you just said partial cancellation is possible — which is right?"

### Update CONTEXT.md inline

When a term is resolved, update `CONTEXT.md` right there. Don't batch these up — capture them as they happen.

`CONTEXT.md` should be totally devoid of implementation details. Do not treat `CONTEXT.md` as a spec, a scratch pad, or a repository for implementation decisions. It is a glossary and nothing else.

Use this format:

```markdown
## Language

**Term**: definition. One or two sentences max.
_Avoid_: alternative terms to not use

## Relationships

- An X has many Ys

## Flagged ambiguities

- "term" was previously used to mean both X and Y — resolved: it means X.
```

### Offer ADRs sparingly

Only offer to create an ADR when all three are true:

1. **Hard to reverse** — the cost of changing your mind later is meaningful
2. **Surprising without context** — a future reader will wonder "why did they do it this way?"
3. **The result of a real trade-off** — there were genuine alternatives and you picked one for specific reasons

If any of the three is missing, skip the ADR.

ADR format:

```markdown
# ADR-NNNN: <title>

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Superseded

## Context

What forced this decision? What were the alternatives?

## Decision

What did we choose? Why?

## Consequences

What becomes easier? What becomes harder? What follow-up work is needed?
```

## Hermes Integration

- Use `read_file` to load existing `CONTEXT.md` and ADRs
- Use `write_file` / `patch` to update files inline
- Use `search_files` to cross-reference terms against the codebase
- Use `session_search("CONTEXT.md domain-modeling")` to recall past domain decisions
