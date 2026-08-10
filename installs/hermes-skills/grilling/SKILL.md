---
name: grilling
description: Design-tree interview to sharpen plans and decisions.
version: 1.0.0
author: Hermes Agent (adapted from Matt Pocock/skills)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [grilling, interview, planning, design-review, decision-making]
    related_skills: [domain-modeling, grill-with-docs, wayfinder]
---

# Grilling

Interview the user relentlessly until you reach a shared understanding. Map this as a **design tree**: every decision branches into the decisions that hang off it.

## The Design Tree

Work the tree in **rounds**. The **frontier** is every decision whose prerequisites are already settled — the questions you can ask _now_ without guessing at answers you haven't heard yet. Ask the whole frontier in one round: number each question and give your recommended answer. Then wait for the user's answers before the next round.

Each question should be formatted like so:

```
❓ **Q1** — **<question title>**: <question body, might be multiple paragraphs, including multiple choices>

➡️ <your recommended answer>
```

Each round the user answers reshapes the tree — settled decisions push the frontier outward and unblock questions that depended on them. Recompute the frontier and ask the next round. A question whose answer depends on another question still open in this round belongs to a _later_ round, not this one.

## Rules

### You find facts, the user makes decisions

Finding _facts_ is your job, never the user's. When a frontier question needs a fact from the environment (filesystem, tools, codebase, web, etc.), use your tools to find it — don't ask the user for anything you could look up yourself.

Don't block on it: a running exploration is an unsettled prerequisite, so only the questions downstream of it wait — ask the rest of the frontier now.

The _decisions_ are the user's — put each to them and wait.

### Completion

The session is done when the frontier is empty: every branch of the design tree visited, nothing left silently assumed.

Do not act on the plan until the user confirms you have reached a shared understanding. State explicitly: "Frontier is empty — every branch of the design tree has been visited. Ready to proceed?"

## When to use this vs grill-with-docs

- **grilling**: pure interview — no files created. Use for decisions, plans, designs that don't touch code or when you don't need persistent docs.
- **grill-with-docs**: interview + builds CONTEXT.md and ADRs as you go. Use when working on a codebase that should retain the domain model.

## Hermes Integration

Use `clarify()` for simpler decisions (single multiple-choice). Use grilling when the decision space is a tree — multiple interdependent choices where later questions depend on answers to earlier ones.

To find facts autonomously, use:
- `search_files` / `read_file` — explore the codebase
- `terminal` — run commands, analyze output
- `web_search` / `web_extract` — research external sources
- `session_search` — recall past decisions from conversation history
- `delegate_task` — dispatch sub-agents for parallel fact-finding
