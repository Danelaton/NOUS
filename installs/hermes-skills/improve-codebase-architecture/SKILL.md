---
name: improve-codebase-architecture
description: "Use when code feels shallow — scan modules, present HTML architecture report."
version: 1.1.0
author: Hermes Agent (adapted from Matt Pocock/skills)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [architecture, refactoring, modules, depth, code-quality, html-report]
    related_skills: [codebase-design, domain-modeling, grilling]
---

# Improve Codebase Architecture

Surface architectural friction and propose **deepening opportunities** — refactors that turn shallow modules into deep ones. The aim is testability and AI-navigability.

This command is _informed_ by the project's domain model and built on a shared design vocabulary from the `codebase-design` skill.

## Process

### 1. Explore

**Scope before you scan — YAGNI.** Deepening a module pays off by making future changes to it easier, so put extra weight on the parts of the codebase that have recently changed.

- If the user named a direction — a module, subsystem, or pain point — take it and skip the inference below.
- Otherwise, walk back the commit history to find hot spots — files and areas that keep coming up.

Read the project's domain glossary (`CONTEXT.md`) and any ADRs in the area first.

Then explore the codebase. Don't follow rigid heuristics — explore organically and note where you experience friction:
- Where does understanding one concept require bouncing between many small modules?
- Where are modules **shallow** — interface nearly as complex as the implementation?
- Where have pure functions been extracted just for testability, but the real bugs hide in how they're called (no **locality**)?
- Where do tightly-coupled modules leak across their seams?
- Which parts of the codebase are untested, or hard to test through their current interface?

Apply the **deletion test**: would deleting it concentrate complexity, or just move it? A "yes, concentrates" is the signal you want.

### 2. Present candidates as an HTML report

Write a self-contained HTML file to the OS temp directory. Open it for the user.

The report uses **Tailwind via CDN** for layout and styling, and **Mermaid via CDN** for diagrams. Mix Mermaid with hand-crafted CSS/SVG visuals. Each candidate gets a **before/after visualisation**.

For each candidate, render a card with:
- **Files** — which files/modules are involved
- **Problem** — why the current architecture is causing friction
- **Solution** — plain English description of what would change
- **Benefits** — explained in terms of locality and leverage
- **Before / After diagram** — side-by-side, custom-drawn
- **Recommendation strength** — `Strong`, `Worth exploring`, or `Speculative` badge

End with a **Top recommendation** section.

Use CONTEXT.md vocabulary and `codebase-design` vocabulary for architecture terms.

### 3. Grill the candidates

After the report is written and opened, ask: "Which of these would you like to explore?" Then run a `grilling` session on the chosen candidate.

## File output

Write to `<tmpdir>/architecture-review-<timestamp>.html`:
- Linux/Mac: `$TMPDIR` → `/tmp/`
- Windows: `%TEMP%` or `$TEMP`

Open with:
- Linux: `xdg-open <path>`
- Mac: `open <path>`
- Windows: `start <path>`

## Hermes Integration

- Use `delegate_task` to walk the codebase in parallel
- Use `terminal` for `git log --oneline` hot-spot analysis
- Use `search_files` / `read_file` for codebase exploration
- Use `write_file` to create the HTML report
- Use `terminal` with `start` to open the report on Windows
