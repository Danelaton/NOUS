---
name: grill-with-docs
description: "Use when starting feature work — grill to align requirements, build CONTEXT.md and ADRs."
version: 1.1.0
author: Hermes Agent (adapted from Matt Pocock/skills)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [grilling, domain-modeling, docs, adr, context, planning]
    related_skills: [grilling, domain-modeling, wayfinder]
---

# Grill with Docs

Run a grilling session while simultaneously applying the domain-modeling discipline.

This is the recommended entry point for any codebase change:
1. **Grill** the user to align on what they want
2. **Domain-model** as you go: capture terms in `CONTEXT.md`, offer ADRs for surprising architectural decisions
3. When the grilling session ends (frontier empty), you have a shared understanding AND the domain model is updated

## Workflow

1. Load `CONTEXT.md` if it exists — read it first so you speak the project's language
2. Start the grilling session (design tree, frontier rounds)
3. During each round, as terms are resolved or decisions crystallise, apply domain-modeling:
   - Challenge terms against the existing glossary
   - Sharpen fuzzy language into canonical terms
   - Update `CONTEXT.md` inline (don't batch — capture as they happen)
   - Offer ADRs when a decision is hard-to-reverse, surprising, and the result of a real trade-off
4. When the frontier is empty, confirm the domain model is updated and the user is ready to proceed

## Files touched

- `CONTEXT.md` — created or updated with new/resolved terms
- `docs/adr/NNNN-title.md` — created for qualifying decisions
- No other files are modified — this skill is about understanding, not building

## When NOT to use

- Pure decision-making with no code implications → use `grilling` alone
- Quick clarifications → use `clarify()`
- Already have a spec written → go straight to planning
