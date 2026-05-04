# NOUS Skills Standard

This is the definitive reference for creating, naming, and validating skills in NOUS.

---

## What is a skill?

A skill is a folder with a `SKILL.md` file that provides reusable instructions for an AI agent. Skills are installed globally by `nous install` and copied into projects by `nous sync`.

---

## Installation flow

```
nous install
  → downloads installs/skills/*  →  ~/.nous/skills/
  → downloads installs/skeleton/AGENTS.md  →  ~/.nous/skills/AGENTS.md

nous sync (per project)
  → copies ~/.nous/skills/<skill-name>/  →  .agent/skills/<skill-name>/
  → copies ~/.nous/skills/AGENTS.md     →  project root AGENTS.md
```

Skills in `~/.nous/skills/` are **global** — shared across all projects.
Skills in `.agent/skills/` are **per-project** — local, not tracked in git.

---

## Required structure

Every skill is a folder:

```
<skill-name>/
└── SKILL.md         ← REQUIRED
```

Optional subdirectories:

```
<skill-name>/
├── SKILL.md         ← REQUIRED
├── examples/        ← reference implementations
├── resources/       ← templates, configs, assets
└── scripts/         ← helper scripts (small, focused)
```

---

## SKILL.md format

### 1. YAML frontmatter (required, must be first)

```yaml
---
name: skill-name
description: Third-person description with keywords. Explains what the skill does and when to use it.
---
```

| Field | Required | Rules |
|-------|----------|-------|
| `name` | Yes (recommended) | Lowercase, hyphens. Defaults to folder name if omitted. |
| `description` | **Yes** | Third person ("Helps with..." not "I help with..."). Include keywords for discoverability. Explain **what** and **when**. |

### 2. Markdown content (required, follows frontmatter)

Minimum required sections:

```markdown
# Skill Name

One-line summary of what this skill does.

## When to use

- Specific situation 1
- Specific situation 2

## How to use

Step-by-step instructions the agent will follow.
```

Recommended additional sections:

```markdown
## Examples

Brief examples of the skill in action.

## Notes

Edge cases, limitations, or caveats.
```

---

## Naming conventions

| Rule | Example |
|------|---------|
| Lowercase only | `code-review` not `Code-Review` |
| Hyphens for spaces | `api-design` not `api_design` |
| Descriptive of capability | `bug-triaging` not `util` |
| Folder name = skill name | folder `code-review/` → `name: code-review` |

---

## Description rules

The `description` field is used by agents for **progressive discovery** — they read the description first to decide if the skill is relevant, then read the full `SKILL.md` if needed.

Rules:
1. **Third person** — "Helps with..." / "Creates..." / "Guides the agent through..."
2. **Include keywords** — domain terms that identify when this skill applies
3. **Explain what AND when** — what the skill does + the situation that triggers it
4. **One paragraph** — keep it under 3 sentences

Bad:
```yaml
description: I will help you create skills.
```

Good:
```yaml
description: Creates new skills following the Antigravity format. Use when you need to create a reusable skill for a task, workflow, or domain. Helps design skill structure, write SKILL.md with proper YAML frontmatter, and organize supporting files.
```

---

## Validation checklist

Before committing a skill:

- [ ] Folder name is `kebab-case`
- [ ] `SKILL.md` exists inside the folder
- [ ] YAML frontmatter is at the very top (line 1)
- [ ] `name` field matches the folder name
- [ ] `description` is in third person with keywords
- [ ] `description` explains both what and when
- [ ] Content has at minimum: title, When to use, How to use
- [ ] All referenced files in `examples/`, `resources/`, `scripts/` actually exist
- [ ] No broken paths — uses `.agent/skills/` (not `.agents/skills/`)

---

## Minimal complete example

```
my-skill/
└── SKILL.md
```

```yaml
---
name: my-skill
description: Guides the agent through writing structured commit messages. Use when preparing to commit changes, reviewing diffs, or documenting work in git history.
---

# My Skill

Helps write conventional commit messages following the Angular format.

## When to use

- When about to run `git commit`
- When reviewing a set of file changes that need to be documented

## How to use

1. Read the git diff with `git diff --staged`
2. Identify the type: feat / fix / refactor / docs / chore / ci
3. Write a subject line: `type(scope): brief description`
4. Add bullet points for breaking changes or context if needed

## Examples

```
feat(auth): add JWT refresh token rotation
fix(api): handle null response from upstream service
docs(readme): add macOS install instructions
```
```

---

## Where to add new skills

Add new skills to `installs/skills/<skill-name>/` in the NOUS repo. They will be:
1. Downloaded to `~/.nous/skills/<skill-name>/` by `nous install`
2. Copied to `.agent/skills/<skill-name>/` by `nous sync`

For project-specific skills not meant for distribution, place them directly in `.agent/skills/<skill-name>/` inside your project.

---

## Output conventions

Skills generate output files in two different locations. This is intentional:

| Output | Location | Why |
|--------|----------|-----|
| `PROJECT_MAP.md` | Project root | Public project document — commitable, readable by all contributors |
| `ARCHITECTURE_REVIEW.md` | Project root | Public project document — commitable, shareable with team |
| `.agent/knowledge/` | `.agent/` subdirectory | Private accumulated memory — optionally excluded from git via `.gitignore` |
| `.agent/skills/` | `.agent/` subdirectory | Skills copied by `nous sync` — do not edit manually, re-sync to update |

Root-level outputs (`PROJECT_MAP.md`, `ARCHITECTURE_REVIEW.md`) are meant to be committed and shared. `.agent/` outputs are local working state — they persist across sessions but are not necessarily tracked.

---

## Recommended execution order for new projects

When starting on a new or unfamiliar project, run the skills in this order:

1. **`project-map`** — Scan the codebase and generate `PROJECT_MAP.md`. Establishes the factual baseline: stack, structure, entry points, relationships.

2. **`architecture-review`** — Analyze `PROJECT_MAP.md` and generate `ARCHITECTURE_REVIEW.md`. Identifies structural risks, coupling issues, and priority actions.

3. **`knowledge` (ingest)** — Feed existing docs, specs, ADRs, and the generated files into `.agent/knowledge/`. Builds the queryable memory base.

4. **`knowledge` (consolidate)** — Connect patterns across ingested entries. Surfaces implicit dependencies and contradictions discovered during onboarding.

5. **`skill-creator`** — Once you understand the project, create custom skills for recurring workflows specific to this codebase.

Skills 1–2 are stateless (generate files, no persistent state). Skills 3–4 are stateful (accumulate in `.agent/knowledge/`). Skill 5 extends the system itself.

---

## Known issues (to be fixed)

| # | File | Line | Issue |
|---|------|------|-------|
| 1 | `installs/install.sh` | 131-143 | JSON parser iterates lines instead of objects — GitHub API returns multi-line JSON, so `grep/sed` parsing fails silently on macOS. Skills folder not downloaded. |
| 2 | `installs/skills/skill-creator/SKILL.md` | 42, 140, 149 | `.agents/skills/` typo — should be `.agent/skills/` (no trailing `s`) |
| 3 | `installs/install.sh` | 158-160 | Redundant copy: `SKILLS_DIR` and `NOUS_DIR/skills` resolve to the same path (`~/.nous/skills`). Phase 3 copies files onto themselves. |