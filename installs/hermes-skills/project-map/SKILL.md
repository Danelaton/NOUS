---
name: project-map
description: "Use to map an unknown codebase — generate structured PROJECT_MAP.md."
version: 1.1.0
author: Hermes Agent (adapted from NOUS)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [project-map, codebase, onboarding, architecture, scanning]
    related_skills: [architecture-review, codebase-design, okf-knowledge]
---

# Project Map

Generates a structured `PROJECT_MAP.md` by scanning the file system, reading dependency manifests, and analyzing imports. Zero LLM calls for extraction, zero external APIs.

## When to use

- Starting work on an unfamiliar codebase
- Onboarding to a project for the first time
- `PROJECT_MAP.md` does not exist or is outdated
- Before running `architecture-review` (this skill feeds it)
- Asked "what does this project do?" or "how is this structured?"

## Steps

### Step 1 — Scan folder structure

Use `search_files` with `target='files'` to get top-level structure. Identify project type:

| Pattern | Type |
|---------|------|
| Multiple `*/cmd/` or `*/service-*/` dirs | Microservices / monorepo |
| Single `src/` or `app/` with `controllers/`, `models/`, `views/` | MVC |
| `packages/` or `apps/` at root | Monorepo workspaces |
| Single `main.*` or `index.*` + flat structure | Single app |

### Step 2 — Detect stack

Read dependency manifests via `read_file` — do NOT infer from file names alone:

- `package.json` / `yarn.lock` → Node/JS/TS
- `go.mod` → Go
- `pyproject.toml` / `requirements.txt` → Python
- `Cargo.toml` → Rust
- `Gemfile` → Ruby
- `pom.xml` / `build.gradle` → Java/Kotlin

Count files by extension via `search_files`:

```bash
find . -not -path '*/.git/*' -not -path '*/node_modules/*' \
  -type f | sed 's/.*\.//' | sort | uniq -c | sort -rn | head -20
```

### Step 3 — Identify entry points

Use `search_files` to find common entry patterns:
- `main.go`, `cmd/*/main.go`
- `main.py`, `app.py`, `server.py`
- `index.js`, `src/index.ts`
- `main.rs`, `src/main.rs`

Search for API routes:
```bash
grep -r "router\.\|app\.get\|app\.post\|@app\.route\|http\.HandleFunc" \
  --include="*.go" --include="*.py" --include="*.ts" -l 2>/dev/null | head -10
```

### Step 4 — Detect architecture relationships

Use `search_files` with regex to find internal imports and map module relationships.

### Step 5 — Detect conventions

```bash
# Test file patterns
find . -name "*_test.*" -o -name "*.test.*" -o -name "*.spec.*" 2>/dev/null | head -5

# Commit style
git log --oneline -10 2>/dev/null

# CI/CD
ls .github/workflows/ .circleci/ Makefile Dockerfile 2>/dev/null
```

### Step 6 — Generate PROJECT_MAP.md

Write to project root with this structure:

```markdown
# PROJECT_MAP

Generated: <date>
Tool: nous project-map

## Overview
- **Type:** <monorepo|microservices|MVC|single-app|library>
- **Primary language:** <language>
- **Stack:** <frameworks, runtime>
- **Entry points:** <list>

## Architecture
Key modules:
| Module | Path | Responsibility |
|--------|------|----------------|

## Key Files
| File | Purpose |
|------|---------|

## Dependencies
### Runtime / Dev
| Package | Version | Purpose |
|---------|---------|---------|

## Conventions
- **Tests:** <pattern>
- **Commits:** <style>
- **CI/CD:** <tools>
```

## Validation

- [ ] Steps 1–5 ran using file system commands
- [ ] Stack detected from manifest files, not guessed
- [ ] Entry points verified to exist
- [ ] `PROJECT_MAP.md` written to project root
- [ ] All sections populated — use "N/A" if genuinely not applicable

## Notes

- Read `README.md` first as a starting point if it exists
- `PROJECT_MAP.md` is input for `architecture-review` — keep it factual
- Re-run when major structural changes happen
- After generating, update `.agents/OKF/architecture.md` with verified durable structure
