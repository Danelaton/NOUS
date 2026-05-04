---
name: project-map
description: Scans a project's file system, detects stack and architecture, and generates a structured PROJECT_MAP.md. Use when starting work on an unfamiliar codebase, onboarding to a project, or needing a quick architectural overview without external tools or APIs.
---

# Project Map

Generates a structured PROJECT_MAP.md by scanning the file system, reading dependency manifests, and analyzing imports — zero LLM calls for extraction, zero external APIs.

## When to use

- When starting work on an unfamiliar codebase
- When onboarding to a project for the first time
- When PROJECT_MAP.md does not exist or is outdated
- Before running `architecture-review` (this skill feeds it)
- When asked "what does this project do?" or "how is this structured?"

## How to use

Execute the steps below sequentially. All extraction uses file system commands and grep only — no LLM inference during extraction. Use LLM reasoning only to format the final output.

## Steps

### Step 1 — Scan folder structure

```bash
# Get top-level structure (depth 2–3)
find . -maxdepth 3 -not -path '*/.git/*' -not -path '*/node_modules/*' \
       -not -path '*/__pycache__/*' -not -path '*/vendor/*' \
       -not -path '*/.agent/*' -not -path '*/dev/*' | sort
```

Identify project type from structure:
| Pattern | Type |
|---------|------|
| Multiple `*/cmd/` or `*/service-*/` dirs | Microservices / monorepo |
| Single `src/` or `app/` with `controllers/`, `models/`, `views/` | MVC |
| `packages/` or `apps/` at root | Monorepo (npm/pnpm workspaces) |
| Single `main.*` or `index.*` + flat structure | Single app |

### Step 2 — Detect stack

Read dependency manifests — do NOT infer from file names alone:

```bash
# Node / JS / TS
cat package.json 2>/dev/null
cat yarn.lock 2>/dev/null | head -5   # detect yarn vs npm vs pnpm

# Go
cat go.mod 2>/dev/null

# Python
cat pyproject.toml 2>/dev/null
cat requirements.txt 2>/dev/null
cat setup.py 2>/dev/null

# Rust
cat Cargo.toml 2>/dev/null

# Ruby
cat Gemfile 2>/dev/null

# Java / Kotlin
cat pom.xml 2>/dev/null
cat build.gradle 2>/dev/null
```

Count files by extension to confirm primary language:

```bash
find . -not -path '*/.git/*' -not -path '*/node_modules/*' -not -path '*/vendor/*' \
  -type f | sed 's/.*\.//' | sort | uniq -c | sort -rn | head -20
```

### Step 3 — Identify entry points

```bash
# Common entry point patterns
ls main.go cmd/*/main.go 2>/dev/null        # Go
ls main.py app.py server.py 2>/dev/null     # Python
ls index.js src/index.ts 2>/dev/null        # JS/TS
ls main.rs src/main.rs 2>/dev/null          # Rust
ls Program.cs src/main/*/Application.java 2>/dev/null  # C# / Java

# Detect API routes
grep -r "router\.\|app\.get\|app\.post\|@app\.route\|http\.HandleFunc\|\.Handle(" \
  --include="*.go" --include="*.py" --include="*.js" --include="*.ts" \
  -l 2>/dev/null | head -10
```

### Step 4 — Detect architecture relationships

```bash
# Find internal imports/requires to map module relationships
# Go
grep -r '"github.com/' --include="*.go" -h 2>/dev/null | \
  sed 's/.*"\(.*\)".*/\1/' | sort | uniq -c | sort -rn | head -20

# JS/TS — internal imports
grep -r "from '\.\." --include="*.ts" --include="*.js" -h 2>/dev/null | \
  sed "s/.*from '\\(\\.[^']*\\)'.*/\\1/" | sort | uniq -c | sort -rn | head -20

# Python
grep -r "^from \.\|^import \." --include="*.py" -h 2>/dev/null | \
  sort | uniq -c | sort -rn | head -20
```

### Step 5 — Detect conventions

```bash
# Naming: check test file patterns
find . -name "*_test.*" -o -name "*.test.*" -o -name "*.spec.*" 2>/dev/null | head -5

# Commit style: sample last 10 commits
git log --oneline -10 2>/dev/null

# Config files: detect CI/CD and tooling
ls .github/workflows/ .circleci/ .gitlab-ci.yml Makefile Dockerfile \
   docker-compose.yml 2>/dev/null
```

### Step 6 — Generate PROJECT_MAP.md

Using the data gathered above, write `PROJECT_MAP.md` to the project root with this structure:

```markdown
# PROJECT_MAP

Generated: <date>
Tool: nous project-map

## Overview

- **Type:** <monorepo|microservices|MVC|single-app|library>
- **Primary language:** <language> (<file count> files)
- **Stack:** <framework(s)>, <runtime version>
- **Entry points:** <list>

## Architecture

<ASCII or text diagram of main modules and their relationships>

Key modules:
| Module | Path | Responsibility |
|--------|------|----------------|

## Key Files

| File | Purpose |
|------|---------|

## Dependencies

### Runtime
| Package | Version | Purpose |
|---------|---------|---------|

### Dev
| Package | Version | Purpose |
|---------|---------|---------|

## Conventions

- **Tests:** <pattern — e.g., *_test.go, *.spec.ts>
- **Commits:** <style — e.g., conventional commits, free-form>
- **Linting:** <tools detected>
- **CI/CD:** <tools detected>

## Relationships

<How modules depend on each other — derived from import analysis>
```

## Validation checklist

Before saving PROJECT_MAP.md:

- [ ] Step 1–5 ran using file system commands only (no LLM inference for extraction)
- [ ] Stack detected from actual manifest files, not guessed from extensions
- [ ] Entry points verified to actually exist
- [ ] PROJECT_MAP.md written to project root
- [ ] All sections populated — leave "N/A" if genuinely not applicable, never skip a section

## Notes

- Never use an LLM to infer what a file does without reading it first
- If `node_modules/` or `vendor/` is huge, skip scanning inside them
- PROJECT_MAP.md is input for the `architecture-review` skill — keep it factual, not analytical
- Re-run this skill when major structural changes happen (new modules, refactors)
- If the project has an existing `README.md`, read it first as a starting point