---
name: opencode-memory
description: Guides installation and configuration of opencode-mem — a persistent memory plugin for OpenCode that uses local SQLite + USearch vector search, HuggingFace local embeddings, and auto-capture. Use when adding cross-session memory to an OpenCode environment without cloud dependencies.
---

# OpenCode Memory Plugin

Installs and configures [opencode-mem](https://github.com/tickernelz/opencode-mem) — an OpenCode plugin that gives AI coding agents persistent memory using a local vector database (SQLite + USearch). No cloud, no external API required for storage.

**What it provides:**
- Memories persist across OpenCode sessions
- Semantic search — find past context by meaning, not exact words
- Auto-capture — extracts memories automatically when a session ends
- User profile learning — adapts to your coding preferences over time
- Web dashboard at `http://127.0.0.1:4747` for browsing and managing memories
- Smart deduplication and 30-day auto-cleanup

**Stack:** TypeScript + Bun, `@opencode-ai/plugin`, SQLite, USearch, HuggingFace Transformers (local)

**npm:** [`opencode-mem`](https://www.npmjs.com/package/opencode-mem)

## When to use

- When OpenCode sessions need persistent memory that survives restarts
- When you want the agent to remember architectural decisions, preferences, and past context
- When you want semantic memory search without a cloud backend
- When you want per-project memory scoping plus optional cross-project search

## How to use

Add `opencode-mem` to your OpenCode plugin list, configure `~/.config/opencode/opencode-mem.jsonc` with your AI provider, and restart OpenCode.

---

## Steps

### Step 1 — Add plugin to OpenCode

Edit `~/.config/opencode/opencode.json` and add `opencode-mem` to the plugin array:

```jsonc
{
  "plugin": ["opencode-mem"]
}
```

OpenCode downloads the plugin automatically on next startup. No `npm install` needed — OpenCode handles distribution.

**Note:** If you want a specific version, use:

```jsonc
{
  "plugin": ["opencode-mem@2.13.0"]
}
```

---

### Step 2 — Configure the plugin

Create (or edit) `~/.config/opencode/opencode-mem.jsonc`:

```jsonc
{
  // --- Required: choose ONE of these two approaches ---

  // Option A: Use an OpenCode-connected provider (recommended, no extra key needed)
  "opencodeProvider": "anthropic",
  "opencodeModel": "claude-haiku-4-5",
  // or: "opencodeProvider": "openai", "opencodeModel": "gpt-4o-mini"

  // Option B: Use a direct AI API
  // "memoryProvider": "openai-chat",
  // "memoryModel": "gpt-4o-mini",
  // "memoryApiUrl": "https://api.openai.com/v1",
  // "memoryApiKey": "$OPENAI_API_KEY",  // reads from env var

  // --- Identity (helps with per-user memory scoping) ---
  "userEmailOverride": "you@example.com",
  "userNameOverride": "Your Name",

  // --- Memory scope ---
  "memory": {
    "defaultScope": "project"  // "project" or "all-projects"
  },

  // --- Auto-capture (extracts memories when session ends) ---
  "autoCaptureEnabled": true,
  "autoCaptureMaxIterations": 5,
  "autoCaptureLanguage": "en",  // optional: "es", "fr", etc.

  // --- Web dashboard ---
  "webServerEnabled": true,
  "webServerPort": 4747,
  "webServerHost": "127.0.0.1",

  // --- Cleanup ---
  "autoCleanupEnabled": true,
  "autoCleanupRetentionDays": 30,

  // --- Toasts ---
  "showAutoCaptureToasts": true,
  "showUserProfileToasts": true
}
```

**Minimal config (copy this to get started fast):**

```jsonc
{
  "opencodeProvider": "anthropic",
  "opencodeModel": "claude-haiku-4-5",
  "userEmailOverride": "you@example.com",
  "userNameOverride": "Your Name"
}
```

---

### Step 3 — Restart OpenCode

Close and reopen OpenCode. On first start with the plugin:

1. The embedding model downloads (~50MB for `Xenova/nomic-embed-text-v1`) — takes ~30 seconds
2. A toast appears: "Memory Explorer — Web UI started at http://127.0.0.1:4747"
3. Storage directory created at `~/.opencode-mem/data/`

If no toast appears, check the OpenCode logs. If the download fails, ensure you have internet access.

---

### Step 4 — Verify the plugin works

In an OpenCode session, use the `memory` tool:

```typescript
// Add a memory
memory({ mode: "add", content: "This project uses PostgreSQL 15 with pgvector" })

// Search memories
memory({ mode: "search", query: "database technology" })

// List recent memories
memory({ mode: "list", limit: 10 })

// View your user profile (learned preferences)
memory({ mode: "profile" })

// Search across all projects
memory({ mode: "search", query: "architecture decisions", scope: "all-projects" })

// Show help
memory({ mode: "help" })
```

**Expected behavior:**
- `add` → returns `{ success: true, id: "..." }`
- `search` → returns ranked results with similarity scores
- `list` → returns recent memories with timestamps
- `profile` → returns learned coding preferences (populated after a few sessions)

---

### Step 5 — Configure auto-capture behavior

Auto-capture runs automatically when a session completes. It reads the session transcript and extracts key decisions, patterns, and context.

**Control it in your config:**

```jsonc
{
  "autoCaptureEnabled": true,          // on/off
  "autoCaptureMaxIterations": 5,       // how many extraction passes (more = thorough, slower)
  "autoCaptureIterationTimeout": 30000 // ms per iteration
}
```

**When auto-capture fires:**
- On every `session.completed` event
- Only if `isConfigured()` returns true (valid AI provider set)
- Only on the "web server owner" instance (prevents duplicate runs if multiple sessions active)

To see auto-capture in action: complete a meaningful coding session, then check the web dashboard at `http://127.0.0.1:4747`.

---

### Step 6 — Use the web dashboard

Open `http://127.0.0.1:4747` in a browser.

**Features:**
- **Memories tab** — browse all stored memories, filter by project/date, delete individual memories
- **Consolidations tab** — view synthesized insights across memories
- **Profile tab** — view learned user preferences and patterns
- **Query** — test semantic search from the browser
- **Upload** — ingest files directly via the dashboard

**Port conflict?** Change the port in config:

```jsonc
{
  "webServerPort": 4748
}
```

---

### Step 7 — Memory scoping

Memories are scoped by project (based on the project directory path hash):

```typescript
// Only search current project memories (default)
memory({ mode: "search", query: "auth implementation" })
memory({ mode: "search", query: "auth implementation", scope: "project" })

// Search across ALL projects
memory({ mode: "search", query: "auth implementation", scope: "all-projects" })
```

**Project memory** = memories tagged with the current directory's hash.
**All-projects** = search across every project you've ever used memory in.

Change the default scope:

```jsonc
{
  "memory": {
    "defaultScope": "all-projects"
  }
}
```

---

### Step 8 — Configure the embedding model (optional)

The default embedding model is `Xenova/nomic-embed-text-v1` (768 dimensions, ~50MB). It runs locally — no API calls.

**Change the model:**

```jsonc
{
  "embeddingModel": "Xenova/all-MiniLM-L6-v2",   // smaller, faster (384 dims)
  "embeddingDimensions": 384                       // must match the model
}
```

**Available local models (all free, no API key):**

| Model | Dimensions | Size | Notes |
|-------|-----------|------|-------|
| `Xenova/nomic-embed-text-v1` | 768 | ~50MB | Default, good quality |
| `Xenova/all-MiniLM-L6-v2` | 384 | ~25MB | Smaller, faster |
| `Xenova/bge-base-en-v1.5` | 768 | ~50MB | Strong for code |
| `Xenova/all-mpnet-base-v2` | 768 | ~50MB | Higher quality |

**Use OpenAI embedding API instead:**

```jsonc
{
  "embeddingApiUrl": "https://api.openai.com/v1",
  "embeddingApiKey": "$OPENAI_API_KEY",
  "embeddingModel": "text-embedding-3-small",
  "embeddingDimensions": 1536
}
```

**Warning:** Changing the embedding model invalidates existing vectors — stored memories won't be searchable until re-ingested.

---

### Step 9 — Adjust cleanup and deduplication

**Retention (default: 30 days):**

```jsonc
{
  "autoCleanupEnabled": true,
  "autoCleanupRetentionDays": 90  // keep for 90 days instead
}
```

Pinned memories are never deleted by cleanup. Pin a memory from the web dashboard.

**Deduplication (default: enabled, threshold 0.9):**

```jsonc
{
  "deduplicationEnabled": true,
  "deduplicationSimilarityThreshold": 0.9  // 0.0-1.0, higher = more aggressive
}
```

Deduplication runs automatically on `session.completed`. Exact duplicates are deleted; near-duplicates are reported but not auto-deleted.

---

### Step 10 — Chat message injection

By default, memories are injected as a hidden `<memory>` block on the **first message** of each session. The agent sees past context automatically.

**Control this:**

```jsonc
{
  "chatMessage": {
    "enabled": true,
    "maxMemories": 3,              // how many memories to inject
    "excludeCurrentSession": true, // don't inject memories from THIS session
    "maxAgeDays": 30,              // only inject memories newer than 30 days
    "injectOn": "first"            // "first" (first message only) or "always"
  }
}
```

---

## Validation checklist

- [ ] Plugin added to `~/.config/opencode/opencode.json`
- [ ] `~/.config/opencode/opencode-mem.jsonc` exists with valid provider config
- [ ] OpenCode restarted — embedding model downloaded (check `~/.opencode-mem/data/.cache/`)
- [ ] Toast appears: "Memory Explorer — Web UI started at http://127.0.0.1:4747"
- [ ] `memory({ mode: "add", content: "test" })` returns `{ success: true }`
- [ ] `memory({ mode: "search", query: "test" })` returns results
- [ ] Web UI accessible at `http://127.0.0.1:4747`
- [ ] After completing a session, new memories appear in the dashboard (auto-capture)

## Notes

- **Bun is required** — OpenCode itself uses Bun. The plugin runs in the same Bun process. No setup needed.
- **Storage location:** `~/.opencode-mem/data/` — contains `registry.db` (shard index) + per-project `.db` shard files
- **Config file format:** JSONC — supports comments (`//`) and trailing commas. Standard JSON also works (`.json` extension).
- **`$ENV_VAR` syntax** in config values — e.g., `"memoryApiKey": "$OPENAI_API_KEY"` reads from environment. Use this to keep secrets out of the config file.
- **`.agent/knowledge/` vs opencode-mem:** If your project uses the `knowledge` skill (file-based markdown memory), opencode-mem is independent — it's a vector DB layer on top of OpenCode sessions, not a replacement.
- **Multiple projects:** Each project gets its own memory shard based on directory hash. Running `nous sync` in a project just copies the skill — memory storage stays in `~/.opencode-mem/` (global, not per-project).
- **Pinned memories** survive cleanup. Pin important architectural decisions from the web dashboard.
- **Source code:** https://github.com/tickernelz/opencode-mem — for debugging, contributing, or building on top of it.
