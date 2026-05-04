---
name: knowledge
description: Manages accumulated project knowledge through three unified modes — ingest new information, consolidate patterns across entries, and query using full memory. Use when the agent needs to store, connect, or retrieve information about the project without external APIs or vector databases.
---

# Knowledge

Manages the knowledge base stored in `.agent/knowledge/` using only file reads, LLM reasoning, and YAML frontmatter — no external APIs, no vector DB.

Automatically detects which mode to use based on context. The user does not need to say "ingest" or "query" explicitly.

## When to use

- When the user provides new information (file, URL, text, screenshot, transcript) to be remembered
- When asked a question that requires accumulated project memory
- When there are 5+ unprocessed entries or disconnected information that should be connected
- When the user asks "what do we know about X?" or "have we discussed Y before?"
- When onboarding to a project that has existing `.agent/knowledge/` entries

## How to use

Detect the mode from context, then follow the corresponding steps below.

| Signal | Mode |
|--------|------|
| User provides new info or a file/URL to process | **Ingest** |
| User asks a question about the project | **Query** |
| 5+ entries with `consolidated: false`, or user asks to consolidate | **Consolidate** |

Before any mode: ensure `.agent/knowledge/entries/`, `.agent/knowledge/consolidations/`, and `.agent/knowledge/index.md` exist. Create them if not.

```bash
mkdir -p .agent/knowledge/entries .agent/knowledge/consolidations
touch .agent/knowledge/index.md 2>/dev/null || true
```

---

## Mode 1: Ingest

### Step 1 — Process input

Read or receive the input:
- File: read its content
- URL: fetch and extract text
- Text / screenshot / transcript: use as-is

### Step 2 — Extract structured info

Using LLM reasoning, extract:
- **Summary** — 2–3 sentences describing what this info is about
- **Key Points** — concrete facts, decisions, or data points
- **Entities** — people, services, systems, concepts mentioned (e.g. `UserService`, `PostgreSQL`, `rate-limiter`)
- **Topics** — 2–4 tags (e.g. `backend`, `performance`, `auth`)
- **Importance** — score 0.0–1.0 based on how critical this is to the project

Importance guide:
| Score | Meaning |
|-------|---------|
| 0.9–1.0 | Critical decision, architectural change, security issue |
| 0.7–0.8 | Important context, key dependency, major feature |
| 0.4–0.6 | Useful reference, minor detail |
| 0.0–0.3 | Background info, low relevance |

### Step 3 — Save entry

Write to `.agent/knowledge/entries/YYYY-MM-DD-<slug>.md`:

```markdown
---
source: <filename or URL>
date: YYYY-MM-DD
importance: 0.8
entities: [UserService, PostgreSQL, rate-limiter]
topics: [backend, performance, database]
consolidated: false
---
# <Descriptive title>

## Summary
<2–3 sentences>

## Key Points
- point 1
- point 2

## Raw Notes
<relevant details extracted from source>
```

Slug rules: lowercase, hyphens, max 40 chars, derived from title (e.g. `postgres-connection-pooling`).

### Step 4 — Update index.md

Append to `.agent/knowledge/index.md`:

```markdown
| YYYY-MM-DD | <slug>.md | <one-line summary> | [topic1, topic2] | 0.8 |
```

If index.md is empty, create it with this header first:

```markdown
# Knowledge Index

| Date | Entry | Summary | Topics | Importance |
|------|-------|---------|--------|------------|
```

---

## Mode 2: Consolidate

### Step 1 — Load unprocessed entries

```bash
grep -rl "consolidated: false" .agent/knowledge/entries/ 2>/dev/null
```

Read all entries returned. Also read:

```bash
cat .agent/MEMORY.md 2>/dev/null
cat PROJECT_MAP.md 2>/dev/null
git log --oneline -20 2>/dev/null
```

### Step 2 — Find cross-entry patterns

Using LLM reasoning, identify:
- Entities that appear in multiple entries
- Recurring topics
- Contradictions between entries
- Implicit dependencies (A depends on B, discovered separately)
- Temporal patterns (things that changed over time)

If `PROJECT_MAP.md` was loaded in Step 1, use it as architectural context: check whether entries touch modules that are related or disconnected according to the map. For example, if two entries both mention a module that PROJECT_MAP.md identifies as a high-coupling area, flag that connection explicitly in the consolidation.

### Step 3 — Save consolidation

Write to `.agent/knowledge/consolidations/YYYY-MM-DD.md`:

```markdown
---
date: YYYY-MM-DD
sources: [entry-1.md, entry-2.md, entry-3.md]
---
# Consolidation: <pattern title>

## Pattern Discovered
<what was found across entries>

## Connections
- entry-A + entry-B: <relationship>
- entry-C contradicts entry-A on: <what>

## Implications
- <what this means for the project>
- <suggested action>
```

### Step 4 — Mark entries as consolidated

For each source entry, update its frontmatter:

```
consolidated: false  →  consolidated: true
```

### Step 5 — Update index.md

Append a consolidation row:

```markdown
| YYYY-MM-DD | consolidations/<file>.md | <pattern title> | [topics] | — |
```

---

## Mode 3: Query

### Step 1 — Load index

```bash
cat .agent/knowledge/index.md
```

### Step 2 — Identify relevant entries

From the index, select entries whose topics, entities, or summary match the question. Prioritize:
- Entries with `importance >= 0.9` — never skip these
- Recent consolidations (they summarize multiple entries)
- Entries matching entities or topics in the question

### Step 3 — Read relevant entries in detail

```bash
cat .agent/knowledge/entries/<relevant-entry>.md
cat .agent/knowledge/consolidations/<relevant-consolidation>.md
```

### Step 4 — Synthesize answer

Using LLM reasoning, compose a response that:
- Cites specific sources: "According to [entry-name.md]: ..."
- Connects information across multiple entries when relevant
- Flags contradictions if found
- States confidence level if uncertain

### Step 5 — Handle missing info

If the index has no relevant entries:
- Say so explicitly: "There is no recorded knowledge about X."
- Suggest: "You can ingest [specific source] to add this to the knowledge base."

---

## Validation checklist

Before saving any entry or consolidation:

- [ ] `.agent/knowledge/entries/` and `.agent/knowledge/consolidations/` exist
- [ ] Every entry has complete YAML frontmatter: `source`, `date`, `importance`, `entities`, `topics`, `consolidated`
- [ ] `index.md` is updated after every ingest and every consolidation
- [ ] Consolidation never deletes original entries — only marks `consolidated: true`
- [ ] Query always cites specific source files
- [ ] Entries with `importance >= 0.9` are never skipped in queries
- [ ] Slug is kebab-case, max 40 chars

## Notes

- All three modes use LLM reasoning — for extraction (ingest), pattern detection (consolidate), and synthesis (query)
- No external APIs, no vector DB — only file reads + reasoning
- The agent should auto-detect mode from context; the user should never need to say "ingest" or "query" explicitly
- If `PROJECT_MAP.md` exists, consolidate must read it as additional context
- Entries are append-only — never edit or delete past entries, only add new ones or mark consolidated
- `index.md` is the fast lookup layer; always read it before scanning individual entries
- Knowledge persists across sessions because it lives in `.agent/knowledge/` (not tracked by git by default)
- If entries contain sensitive information (private specs, internal meeting notes, unreleased decisions), add `.agent/knowledge/` to `.gitignore`. If the project already has a `.gitignore`, suggest appending it automatically: `echo '.agent/knowledge/' >> .gitignore`
