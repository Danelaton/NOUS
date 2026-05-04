---
name: adk-memory-agent
description: Guides the full implementation of an always-on persistent memory system using Google ADK and Gemini. Creates a background agent that continuously ingests multimodal files, consolidates patterns, and answers queries — all backed by SQLite with no vector database. Use when building a project-level memory layer that runs 24/7 as a lightweight process.
---

# ADK Memory Agent

Implements an always-on memory agent using Google ADK + Gemini. The system watches a folder for new files (text, images, audio, video, PDFs), ingests them into SQLite, consolidates patterns on a timer, and serves queries via HTTP.

**Architecture:** Single `agent.py` with three specialist sub-agents (ingest, consolidate, query) routed by an orchestrator. All logic lives in one file — no microservices, no separate packages.

**Reference implementation:** [GoogleCloudPlatform/generative-ai — always-on-memory-agent](https://github.com/GoogleCloudPlatform/generative-ai/tree/main/gemini/agents/always-on-memory-agent)

> This skill replicates the reference repo exactly. Follow the steps in order and paste each block into `agent.py` — the result will be byte-for-byte equivalent to the original.

## When to use

- When a project needs persistent memory that survives across sessions
- When building an agent that must process multimodal files automatically (images, audio, PDFs)
- When the user asks for a "memory layer", "knowledge base", or "always-on agent"
- When replacing ad-hoc note-taking with structured, queryable memory
- When you need a self-hosted alternative to vector databases for agent memory

## How to use

Follow the steps below in order. All code belongs in a single `agent.py` — append each block in sequence. The optional `dashboard.py` is separate.

---

## Steps

### Step 1 — Scaffold

```bash
mkdir -p memory-agent/inbox
cd memory-agent
touch agent.py dashboard.py requirements.txt .env.example
```

Expected layout after all steps:

```
memory-agent/
├── agent.py          # Everything: DB, tools, agents, API, watcher
├── dashboard.py      # Streamlit UI (optional, Step 8)
├── requirements.txt
├── .env.example
├── inbox/            # Drop files here for auto-ingestion
└── memory.db         # Created automatically on first run
```

---

### Step 2 — Imports, config, and database

Paste this at the top of `agent.py`:

```python
"""
Agent Memory Layer — Always-On ADK Agent

A lightweight, cost-effective background agent that continuously processes, consolidates, and serves memory. Runs 24/7 on Gemini 3.1 Flash-Lite.

Usage:
    python agent.py                          # watch ./inbox, serve on :8888
    python agent.py --watch ./docs --port 9000
    python agent.py --consolidate-every 15   # consolidate every 15 min

Query:
    curl "http://localhost:8888/query?q=what+do+you+know"
    curl -X POST http://localhost:8888/ingest -d '{"text": "some info"}'
"""
import argparse
import asyncio
import json
import logging
import mimetypes
import os
import shutil
import signal
import sqlite3
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

from aiohttp import web
from google.adk.agents import Agent
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService
from google.genai import types

# ─── Config ────────────────────────────────────────────────────

MODEL = os.getenv("MODEL", "gemini-3.1-flash-lite-preview")
DB_PATH = os.getenv("MEMORY_DB", "memory.db")

# Supported file types for multimodal ingestion
TEXT_EXTENSIONS = {".txt", ".md", ".json", ".csv", ".log", ".xml", ".yaml", ".yml"}
MEDIA_EXTENSIONS = {
    # Images
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".jpeg": "image/jpeg",
    ".gif": "image/gif",
    ".webp": "image/webp",
    ".bmp": "image/bmp",
    ".svg": "image/svg+xml",
    # Audio
    ".mp3": "audio/mpeg",
    ".wav": "audio/wav",
    ".ogg": "audio/ogg",
    ".flac": "audio/flac",
    ".m4a": "audio/mp4",
    ".aac": "audio/aac",
    # Video
    ".mp4": "video/mp4",
    ".webm": "video/webm",
    ".mov": "video/quicktime",
    ".avi": "video/x-msvideo",
    ".mkv": "video/x-matroska",
    # Documents
    ".pdf": "application/pdf",
}
ALL_SUPPORTED = TEXT_EXTENSIONS | set(MEDIA_EXTENSIONS.keys())

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(message)s",
    datefmt="[%H:%M]",
)
log = logging.getLogger("memory-agent")

# ─── Database ──────────────────────────────────────────────────

def get_db() -> sqlite3.Connection:
    db = sqlite3.connect(DB_PATH)
    db.row_factory = sqlite3.Row
    db.executescript("""
        CREATE TABLE IF NOT EXISTS memories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT NOT NULL DEFAULT '',
            raw_text TEXT NOT NULL,
            summary TEXT NOT NULL,
            entities TEXT NOT NULL DEFAULT '[]',
            topics TEXT NOT NULL DEFAULT '[]',
            connections TEXT NOT NULL DEFAULT '[]',
            importance REAL NOT NULL DEFAULT 0.5,
            created_at TEXT NOT NULL,
            consolidated INTEGER NOT NULL DEFAULT 0
        );
        CREATE TABLE IF NOT EXISTS consolidations (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source_ids TEXT NOT NULL,
            summary TEXT NOT NULL,
            insight TEXT NOT NULL,
            created_at TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS processed_files (
            path TEXT PRIMARY KEY,
            processed_at TEXT NOT NULL
        );
    """)
    return db
```

**Schema notes:**
- `connections` column stores a JSON array populated by `store_consolidation` (cross-references between memories)
- `consolidated` is `INTEGER` (0/1), not `BOOLEAN` — SQLite has no native boolean
- `processed_files` tracks ingested paths so files are never re-ingested on restart

---

### Step 3 — ADK Tools (all seven)

Append to `agent.py`. Note: `delete_memory` and `clear_all_memories` are required — they back the `/delete` and `/clear` API endpoints.

```python
# ─── ADK Tools ─────────────────────────────────────────────────


def store_memory(
    raw_text: str,
    summary: str,
    entities: list[str],
    topics: list[str],
    importance: float,
    source: str = "",
) -> dict:
    """Store a processed memory in the database.

    Args:
        raw_text: The original input text.
        summary: A concise 1-2 sentence summary.
        entities: Key people, companies, products, or concepts.
        topics: 2-4 topic tags.
        importance: Float 0.0 to 1.0 indicating importance.
        source: Where this memory came from (filename, URL, etc).

    Returns:
        dict with memory_id and confirmation.
    """
    db = get_db()
    now = datetime.now(timezone.utc).isoformat()
    cursor = db.execute(
        """INSERT INTO memories (source, raw_text, summary, entities, topics, importance, created_at)
           VALUES (?, ?, ?, ?, ?, ?, ?)""",
        (source, raw_text, summary, json.dumps(entities), json.dumps(topics), importance, now),
    )
    db.commit()
    mid = cursor.lastrowid
    db.close()
    log.info(f"📥 Stored memory #{mid}: {summary[:60]}...")
    return {"memory_id": mid, "status": "stored", "summary": summary}


def read_all_memories() -> dict:
    """Read all stored memories from the database, most recent first.

    Returns:
        dict with list of memories and count.
    """
    db = get_db()
    rows = db.execute("SELECT * FROM memories ORDER BY created_at DESC LIMIT 50").fetchall()
    memories = []
    for r in rows:
        memories.append({
            "id": r["id"], "source": r["source"], "summary": r["summary"],
            "entities": json.loads(r["entities"]), "topics": json.loads(r["topics"]),
            "importance": r["importance"], "connections": json.loads(r["connections"]),
            "created_at": r["created_at"], "consolidated": bool(r["consolidated"]),
        })
    db.close()
    return {"memories": memories, "count": len(memories)}


def read_unconsolidated_memories() -> dict:
    """Read memories that haven't been consolidated yet.

    Returns:
        dict with list of unconsolidated memories and count.
    """
    db = get_db()
    rows = db.execute(
        "SELECT * FROM memories WHERE consolidated = 0 ORDER BY created_at DESC LIMIT 10"
    ).fetchall()
    memories = []
    for r in rows:
        memories.append({
            "id": r["id"], "summary": r["summary"],
            "entities": json.loads(r["entities"]), "topics": json.loads(r["topics"]),
            "importance": r["importance"], "created_at": r["created_at"],
        })
    db.close()
    return {"memories": memories, "count": len(memories)}


def store_consolidation(
    source_ids: list[int],
    summary: str,
    insight: str,
    connections: list[dict],
) -> dict:
    """Store a consolidation result and mark source memories as consolidated.

    Args:
        source_ids: List of memory IDs that were consolidated.
        summary: A synthesized summary across all source memories.
        insight: One key pattern or insight discovered.
        connections: List of dicts with 'from_id', 'to_id', 'relationship'.

    Returns:
        dict with confirmation.
    """
    db = get_db()
    now = datetime.now(timezone.utc).isoformat()
    db.execute(
        "INSERT INTO consolidations (source_ids, summary, insight, created_at) VALUES (?, ?, ?, ?)",
        (json.dumps(source_ids), summary, insight, now),
    )
    for conn in connections:
        from_id, to_id = conn.get("from_id"), conn.get("to_id")
        rel = conn.get("relationship", "")
        if from_id and to_id:
            for mid in [from_id, to_id]:
                row = db.execute("SELECT connections FROM memories WHERE id = ?", (mid,)).fetchone()
                if row:
                    existing = json.loads(row["connections"])
                    existing.append({"linked_to": to_id if mid == from_id else from_id, "relationship": rel})
                    db.execute("UPDATE memories SET connections = ? WHERE id = ?", (json.dumps(existing), mid))
    placeholders = ",".join("?" * len(source_ids))
    db.execute(f"UPDATE memories SET consolidated = 1 WHERE id IN ({placeholders})", source_ids)
    db.commit()
    db.close()
    log.info(f"🔄 Consolidated {len(source_ids)} memories. Insight: {insight[:80]}...")
    return {"status": "consolidated", "memories_processed": len(source_ids), "insight": insight}


def read_consolidation_history() -> dict:
    """Read past consolidation insights.

    Returns:
        dict with list of consolidation records.
    """
    db = get_db()
    rows = db.execute("SELECT * FROM consolidations ORDER BY created_at DESC LIMIT 10").fetchall()
    result = [{"summary": r["summary"], "insight": r["insight"], "source_ids": r["source_ids"]} for r in rows]
    db.close()
    return {"consolidations": result, "count": len(result)}


def get_memory_stats() -> dict:
    """Get current memory statistics.

    Returns:
        dict with counts of memories, consolidations, etc.
    """
    db = get_db()
    total = db.execute("SELECT COUNT(*) as c FROM memories").fetchone()["c"]
    unconsolidated = db.execute("SELECT COUNT(*) as c FROM memories WHERE consolidated = 0").fetchone()["c"]
    consolidations = db.execute("SELECT COUNT(*) as c FROM consolidations").fetchone()["c"]
    db.close()
    return {
        "total_memories": total,
        "unconsolidated": unconsolidated,
        "consolidations": consolidations,
    }


def delete_memory(memory_id: int) -> dict:
    """Delete a memory by ID.

    Args:
        memory_id: The ID of the memory to delete.

    Returns:
        dict with status.
    """
    db = get_db()
    row = db.execute("SELECT 1 FROM memories WHERE id = ?", (memory_id,)).fetchone()
    if not row:
        db.close()
        return {"status": "not_found", "memory_id": memory_id}
    db.execute("DELETE FROM memories WHERE id = ?", (memory_id,))
    db.commit()
    db.close()
    log.info(f"🗑️  Deleted memory #{memory_id}")
    return {"status": "deleted", "memory_id": memory_id}


def clear_all_memories(inbox_path: str | None = None) -> dict:
    """Delete all memories, consolidations, and inbox files. Full reset."""
    db = get_db()
    mem_count = db.execute("SELECT COUNT(*) as c FROM memories").fetchone()["c"]
    db.execute("DELETE FROM memories")
    db.execute("DELETE FROM consolidations")
    db.execute("DELETE FROM processed_files")
    db.commit()
    db.close()

    # Also clear the inbox folder so files aren't re-ingested
    files_deleted = 0
    if inbox_path:
        folder = Path(inbox_path)
        if folder.is_dir():
            for f in folder.iterdir():
                if f.name.startswith("."):
                    continue  # keep hidden files like .gitkeep
                try:
                    if f.is_file():
                        f.unlink()
                        files_deleted += 1
                    elif f.is_dir():
                        shutil.rmtree(f)
                        files_deleted += 1
                except OSError as e:
                    log.error(f"Failed to delete {f.name}: {e}")

    log.info(f"🗑️  Cleared all {mem_count} memories, deleted {files_deleted} inbox files")
    return {"status": "cleared", "memories_deleted": mem_count, "files_deleted": files_deleted}
```

---

### Step 4 — Sub-agents and orchestrator

Append to `agent.py`. **Copy agent instructions verbatim — ADK routes based on exact wording.**

```python
# ─── ADK Agents ────────────────────────────────────────────────


def build_agents():
    ingest_agent = Agent(
        name="ingest_agent",
        model=MODEL,
        description="Processes raw text or media into structured memory. Call this when new information arrives.",
        instruction=(
            "You are a Memory Ingest Agent. You handle ALL types of input — text, images,\n"
            "audio, video, and PDFs. For any input you receive:\n"
            "1. Thoroughly describe what the content contains\n"
            "2. Create a concise 1-2 sentence summary\n"
            "3. Extract key entities (people, companies, products, concepts, objects, locations)\n"
            "4. Assign 2-4 topic tags\n"
            "5. Rate importance from 0.0 to 1.0\n"
            "6. Call store_memory with all extracted information\n\n"
            "For images: describe the scene, objects, text, people, and any visual details.\n"
            "For audio/video: describe the spoken content, sounds, scenes, and key moments.\n"
            "For PDFs: extract and summarize the document content.\n\n"
            "Use the full description as raw_text in store_memory so the context is preserved.\n"
            "Always call store_memory. Be concise and accurate.\n"
            "After storing, confirm what was stored in one sentence."
        ),
        tools=[store_memory],
    )

    consolidate_agent = Agent(
        name="consolidate_agent",
        model=MODEL,
        description="Merges related memories and finds patterns. Call this periodically.",
        instruction=(
            "You are a Memory Consolidation Agent. You:\n"
            "1. Call read_unconsolidated_memories to see what needs processing\n"
            "2. If fewer than 2 memories, say nothing to consolidate\n"
            "3. Find connections and patterns across the memories\n"
            "4. Create a synthesized summary and one key insight\n"
            "5. Call store_consolidation with source_ids, summary, insight, and connections\n\n"
            "Connections: list of dicts with 'from_id', 'to_id', 'relationship' keys.\n"
            "Think deeply about cross-cutting patterns."
        ),
        tools=[read_unconsolidated_memories, store_consolidation],
    )

    query_agent = Agent(
        name="query_agent",
        model=MODEL,
        description="Answers questions using stored memories.",
        instruction=(
            "You are a Memory Query Agent. When asked a question:\n"
            "1. Call read_all_memories to access the memory store\n"
            "2. Call read_consolidation_history for higher-level insights\n"
            "3. Synthesize an answer based ONLY on stored memories\n"
            "4. Reference memory IDs: [Memory 1], [Memory 2], etc.\n"
            "5. If no relevant memories exist, say so honestly\n\n"
            "Be thorough but concise. Always cite sources."
        ),
        tools=[read_all_memories, read_consolidation_history],
    )

    orchestrator = Agent(
        name="memory_orchestrator",
        model=MODEL,
        description="Routes memory operations to specialist agents.",
        instruction=(
            "You are the Memory Orchestrator for an always-on memory system.\n"
            "Route requests to the right sub-agent:\n"
            "- New information -> ingest_agent\n"
            "- Consolidation request -> consolidate_agent\n"
            "- Questions -> query_agent\n"
            "- Status check -> call get_memory_stats and report\n\n"
            "After the sub-agent completes, give a brief summary."
        ),
        sub_agents=[ingest_agent, consolidate_agent, query_agent],
        tools=[get_memory_stats],
    )

    return orchestrator
```

**Routing rules (exact wording matters):**
- `ingest_agent` — triggered by "New information" (not "New information or files")
- `consolidate_agent` — triggered by "Consolidation request"
- `query_agent` — triggered by "Questions"
- Orchestrator also handles status via `get_memory_stats` tool directly

---

### Step 5 — MemoryAgent runner class

Append to `agent.py`:

```python
# ─── Agent Runner ──────────────────────────────────────────────


class MemoryAgent:
    def __init__(self):
        self.agent = build_agents()
        self.session_service = InMemorySessionService()
        self.runner = Runner(
            agent=self.agent,
            app_name="memory_layer",
            session_service=self.session_service,
        )

    async def run(self, message: str) -> str:
        session = await self.session_service.create_session(
            app_name="memory_layer", user_id="agent",
        )
        content = types.Content(role="user", parts=[types.Part.from_text(text=message)])
        return await self._execute(session, content)

    async def run_multimodal(self, text: str, file_bytes: bytes, mime_type: str) -> str:
        """Send a multimodal message with both text and a media file."""
        session = await self.session_service.create_session(
            app_name="memory_layer", user_id="agent",
        )
        parts = [
            types.Part.from_text(text=text),
            types.Part.from_bytes(data=file_bytes, mime_type=mime_type),
        ]
        content = types.Content(role="user", parts=parts)
        return await self._execute(session, content)

    async def _execute(self, session, content: types.Content) -> str:
        """Run the agent with the given content and return the text response."""
        response = ""
        async for event in self.runner.run_async(
            user_id="agent", session_id=session.id, new_message=content,
        ):
            if event.content and event.content.parts:
                for part in event.content.parts:
                    if hasattr(part, "text") and part.text:
                        response += part.text
        return response

    async def ingest(self, text: str, source: str = "") -> str:
        msg = f"Remember this information (source: {source}):\n\n{text}" if source else f"Remember this information:\n\n{text}"
        return await self.run(msg)

    async def ingest_file(self, file_path: Path) -> str:
        """Ingest a media file (image, audio, video, PDF) via multimodal."""
        suffix = file_path.suffix.lower()
        mime_type = MEDIA_EXTENSIONS.get(suffix)
        if not mime_type:
            # Fallback to mimetypes module
            mime_type, _ = mimetypes.guess_type(str(file_path))
            mime_type = mime_type or "application/octet-stream"

        file_bytes = file_path.read_bytes()
        size_mb = len(file_bytes) / (1024 * 1024)

        # Gemini has a ~20MB inline limit; skip very large files
        if size_mb > 20:
            log.warning(f"⚠️  Skipping {file_path.name} ({size_mb:.1f}MB) — exceeds 20MB limit")
            return f"Skipped: file too large ({size_mb:.1f}MB)"

        prompt = (
            f"Remember this file (source: {file_path.name}, type: {mime_type}).\n\n"
            f"Thoroughly analyze the content of this {mime_type.split('/')[0]} file and "
            f"extract all meaningful information for memory storage."
        )
        log.info(f"🔮 Ingesting {mime_type.split('/')[0]}: {file_path.name} ({size_mb:.1f}MB)")
        return await self.run_multimodal(prompt, file_bytes, mime_type)

    async def consolidate(self) -> str:
        return await self.run("Consolidate unconsolidated memories. Find connections and patterns.")

    async def query(self, question: str) -> str:
        return await self.run(f"Based on my memories, answer: {question}")

    async def status(self) -> str:
        return await self.run("Give me a status report on my memory system.")
```

Key details:
- `ingest()` uses `"Remember this information"` (not `"Remember this"`) — exact wording for orchestrator routing
- `ingest_file()` logs with `🔮` emoji and includes size in MB
- `status()` method exists alongside `consolidate()` and `query()`

---

### Step 6 — File watcher

Append to `agent.py`:

```python
# ─── File Watcher ──────────────────────────────────────────────


async def watch_folder(agent: MemoryAgent, folder: Path, poll_interval: int = 5):
    """Watch a folder for new files and ingest them (text, images, audio, video, PDFs)."""
    folder.mkdir(parents=True, exist_ok=True)
    db = get_db()
    log.info(f"👁️  Watching: {folder}/  (supports: text, images, audio, video, PDFs)")

    while True:
        try:
            for f in sorted(folder.iterdir()):
                if f.name.startswith("."):
                    continue  # skip hidden files
                suffix = f.suffix.lower()
                if suffix not in ALL_SUPPORTED:
                    continue
                row = db.execute("SELECT 1 FROM processed_files WHERE path = ?", (str(f),)).fetchone()
                if row:
                    continue

                try:
                    if suffix in TEXT_EXTENSIONS:
                        # Text-based files — read as string
                        log.info(f"📄 New text file: {f.name}")
                        text = f.read_text(encoding="utf-8", errors="replace")[:10000]
                        if text.strip():
                            await agent.ingest(text, source=f.name)
                    else:
                        # Media files — send as multimodal bytes
                        log.info(f"🖼️  New media file: {f.name}")
                        await agent.ingest_file(f)
                except Exception as file_err:
                    log.error(f"Error ingesting {f.name}: {file_err}")

                db.execute(
                    "INSERT INTO processed_files (path, processed_at) VALUES (?, ?)",
                    (str(f), datetime.now(timezone.utc).isoformat()),
                )
                db.commit()
        except Exception as e:
            log.error(f"Watch error: {e}")

        await asyncio.sleep(poll_interval)
```

---

### Step 7 — Consolidation timer

Append to `agent.py`:

```python
# ─── Consolidation Timer ──────────────────────────────────────


async def consolidation_loop(agent: MemoryAgent, interval_minutes: int = 30):
    """Run consolidation periodically, like sleep cycles."""
    log.info(f"🔄 Consolidation: every {interval_minutes} minutes")
    while True:
        await asyncio.sleep(interval_minutes * 60)
        try:
            db = get_db()
            count = db.execute("SELECT COUNT(*) as c FROM memories WHERE consolidated = 0").fetchone()["c"]
            db.close()
            if count >= 2:
                log.info(f"🔄 Running consolidation ({count} unconsolidated memories)...")
                result = await agent.consolidate()
                log.info(f"🔄 {result[:100]}")
            else:
                log.info(f"🔄 Skipping consolidation ({count} unconsolidated memories)")
        except Exception as e:
            log.error(f"Consolidation error: {e}")
```

Note: the comment says `"like sleep cycles"` — this is the repo's analogy for the human brain's memory consolidation during sleep.

---

### Step 8 — HTTP API (7 endpoints)

Append to `agent.py`. Note: `build_http` takes `watch_path` parameter — used by `/clear` to also wipe the inbox folder.

```python
# ─── HTTP API ──────────────────────────────────────────────────


def build_http(agent: MemoryAgent, watch_path: str = "./inbox"):
    app = web.Application()

    async def handle_query(request: web.Request):
        q = request.query.get("q", "").strip()
        if not q:
            return web.json_response({"error": "missing ?q= parameter"}, status=400)
        answer = await agent.query(q)
        return web.json_response({"question": q, "answer": answer})

    async def handle_ingest(request: web.Request):
        try:
            data = await request.json()
        except Exception:
            return web.json_response({"error": "invalid JSON"}, status=400)
        text = data.get("text", "").strip()
        if not text:
            return web.json_response({"error": "missing 'text' field"}, status=400)
        source = data.get("source", "api")
        result = await agent.ingest(text, source=source)
        return web.json_response({"status": "ingested", "response": result})

    async def handle_consolidate(request: web.Request):
        result = await agent.consolidate()
        return web.json_response({"status": "done", "response": result})

    async def handle_status(request: web.Request):
        stats = get_memory_stats()
        return web.json_response(stats)

    async def handle_memories(request: web.Request):
        data = read_all_memories()
        return web.json_response(data)

    async def handle_delete(request: web.Request):
        try:
            data = await request.json()
        except Exception:
            return web.json_response({"error": "invalid JSON"}, status=400)
        memory_id = data.get("memory_id")
        if not memory_id:
            return web.json_response({"error": "missing 'memory_id' field"}, status=400)
        result = delete_memory(int(memory_id))
        return web.json_response(result)

    async def handle_clear(request: web.Request):
        result = clear_all_memories(inbox_path=watch_path)
        return web.json_response(result)

    app.router.add_get("/query", handle_query)
    app.router.add_post("/ingest", handle_ingest)
    app.router.add_post("/consolidate", handle_consolidate)
    app.router.add_get("/status", handle_status)
    app.router.add_get("/memories", handle_memories)
    app.router.add_post("/delete", handle_delete)
    app.router.add_post("/clear", handle_clear)

    return app
```

All 7 endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/query?q=...` | GET | Query memory with a question |
| `/ingest` | POST | Ingest text `{"text": "...", "source": "..."}` |
| `/consolidate` | POST | Trigger manual consolidation |
| `/status` | GET | Memory statistics |
| `/memories` | GET | List all stored memories (max 50) |
| `/delete` | POST | Delete one memory `{"memory_id": 5}` |
| `/clear` | POST | Full reset — wipes DB + inbox files |

---

### Step 9 — main() and entry point

Append to `agent.py` — this is the final block:

```python
# ─── Main ──────────────────────────────────────────────────────


async def main_async(args):
    agent = MemoryAgent()

    log.info("🧠 Agent Memory Layer starting")
    log.info(f"   Model: {MODEL}")
    log.info(f"   Database: {DB_PATH}")
    log.info(f"   Watch: {args.watch}")
    log.info(f"   Consolidate: every {args.consolidate_every}m")
    log.info(f"   API: http://localhost:{args.port}")
    log.info("")

    # Start background tasks
    tasks = [
        asyncio.create_task(watch_folder(agent, Path(args.watch))),
        asyncio.create_task(consolidation_loop(agent, args.consolidate_every)),
    ]

    # Start HTTP server
    app = build_http(agent, watch_path=args.watch)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, "0.0.0.0", args.port)
    await site.start()

    log.info(f"✅ Agent running. Drop files in {args.watch}/ or POST to http://localhost:{args.port}/ingest")
    log.info(f"   Supported: text, images, audio, video, PDFs")
    log.info("")

    # Wait forever
    try:
        await asyncio.gather(*tasks)
    except asyncio.CancelledError:
        pass
    finally:
        await runner.cleanup()


def main():
    parser = argparse.ArgumentParser(description="Agent Memory Layer - Always-On ADK Agent")
    parser.add_argument("--watch", default="./inbox", help="Folder to watch for new files (default: ./inbox)")
    parser.add_argument("--port", type=int, default=8888, help="HTTP API port (default: 8888)")
    parser.add_argument("--consolidate-every", type=int, default=30, help="Consolidation interval in minutes (default: 30)")
    args = parser.parse_args()

    # Handle graceful shutdown
    loop = asyncio.new_event_loop()

    def shutdown(sig):
        log.info(f"\n👋 Shutting down (signal {sig})...")
        for task in asyncio.all_tasks(loop):
            task.cancel()

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, shutdown, sig)

    try:
        loop.run_until_complete(main_async(args))
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
    finally:
        loop.close()
        log.info("🧠 Agent stopped.")


if __name__ == "__main__":
    main()
```

**Important:** `main()` does NOT read from env vars for defaults — hardcodes `"./inbox"`, `8888`, `30`. Override via CLI flags: `python agent.py --watch ./docs --port 9000 --consolidate-every 15`.

However, `MODEL` and `DB_PATH` ARE read from env vars at module level (see Step 2).

---

### Step 10 — Configuration

**`requirements.txt`** (exact from repo, order preserved):

```
streamlit>=1.40.0
google-genai>=1.0.0
google-adk>=1.0.0
aiohttp>=3.9.0
requests>=2.31.0
```

**`.env.example`:**

```bash
# Required
GOOGLE_API_KEY=your-gemini-api-key

# Optional — override defaults
MODEL=gemini-3.1-flash-lite-preview
MEMORY_DB=./memory.db
```

**Install and run:**

```bash
pip install -r requirements.txt

# Set API key (required)
export GOOGLE_API_KEY="your-key-here"
# Get key from: https://aistudio.google.com/

# Start the agent
python agent.py

# Optional: start dashboard in a separate terminal
streamlit run dashboard.py
```

---

### Step 11 — Dashboard (optional, Streamlit)

The dashboard connects to the running agent via HTTP — it does **not** import `agent.py`. Start the agent first.

```python
# dashboard.py
import json
import time
from pathlib import Path
import requests
import streamlit as st

AGENT_URL = "http://localhost:8888"
INBOX_DIR = Path("./inbox")

def api_get(path):
    try:
        return requests.get(f"{AGENT_URL}{path}", timeout=30).json()
    except Exception as e:
        st.error(f"Agent not reachable: {e}")
        return None

def api_post(path, data):
    try:
        return requests.post(f"{AGENT_URL}{path}", json=data, timeout=60).json()
    except Exception as e:
        st.error(f"Agent not reachable: {e}")
        return None

def main():
    st.set_page_config(page_title="Memory Agent", page_icon="🧠", layout="wide")
    st.title("🧠 Memory Agent Dashboard")

    stats = api_get("/status")
    if stats:
        col1, col2, col3 = st.columns(3)
        col1.metric("Total Memories", stats.get("total_memories", 0))
        col2.metric("Unconsolidated", stats.get("unconsolidated", 0))
        col3.metric("Consolidations", stats.get("consolidations", 0))
    else:
        st.warning("Agent offline. Run: `python agent.py`")
        return

    tab_ingest, tab_query, tab_memories = st.tabs(["📥 Ingest", "🔍 Query", "🧠 Memories"])

    with tab_ingest:
        text = st.text_area("Paste text to ingest", height=150)
        if st.button("Ingest", type="primary") and text.strip():
            with st.spinner("Processing..."):
                result = api_post("/ingest", {"text": text, "source": "dashboard"})
            if result:
                st.success(result.get("response", "Ingested"))

        st.markdown("---")
        st.markdown("**Upload files** (saved to inbox/, auto-ingested by agent)")
        UPLOAD_TYPES = ["txt","md","json","csv","png","jpg","jpeg","gif","webp","mp3","wav","mp4","pdf"]
        files = st.file_uploader("Drop files", type=UPLOAD_TYPES, accept_multiple_files=True)
        if files:
            INBOX_DIR.mkdir(exist_ok=True)
            for f in files:
                dest = INBOX_DIR / f.name
                if not dest.exists():
                    dest.write_bytes(f.getvalue())
                    st.success(f"Saved {f.name} to inbox/")

        st.markdown("---")
        if st.button("🔄 Run Consolidation"):
            with st.spinner("Consolidating..."):
                result = api_post("/consolidate", {})
            if result:
                st.success(result.get("response", "Done"))

        st.markdown("---")
        if st.button("🗑️ Clear All Memories", type="secondary"):
            with st.spinner("Clearing..."):
                result = api_post("/clear", {})
            if result:
                st.warning(f"Cleared {result.get('memories_deleted', 0)} memories")

    with tab_query:
        q = st.text_input("Ask your memory anything")
        if q:
            with st.spinner("Querying..."):
                result = api_get(f"/query?q={q}")
            if result:
                st.markdown(result.get("answer", ""))

    with tab_memories:
        data = api_get("/memories")
        if data and data.get("memories"):
            for m in data["memories"]:
                importance = m.get("importance", 0.5)
                st.markdown(
                    f"**[#{m['id']}]** {m['summary']}  \n"
                    f"`{'  '.join(m.get('topics', []))}` | importance: {importance} | {m.get('source','')}"
                )
                if st.button(f"Delete #{m['id']}", key=f"del_{m['id']}"):
                    api_post("/delete", {"memory_id": m["id"]})
                    st.rerun()
                st.divider()

if __name__ == "__main__":
    main()
```

---

### Step 12 — Test and validate

**Start the agent:**

```bash
python agent.py
# Expected output:
# [HH:MM] 🧠 Agent Memory Layer starting
# [HH:MM]    Model: gemini-3.1-flash-lite-preview
# [HH:MM]    Database: memory.db
# [HH:MM]    Watch: ./inbox
# [HH:MM]    Consolidate: every 30m
# [HH:MM]    API: http://localhost:8888
# [HH:MM]
# [HH:MM] 👁️  Watching: inbox/  (supports: text, images, audio, video, PDFs)
# [HH:MM] 🔄 Consolidation: every 30 minutes
# [HH:MM] ✅ Agent running. Drop files in ./inbox/ or POST to http://localhost:8888/ingest
# [HH:MM]    Supported: text, images, audio, video, PDFs
```

**Smoke test sequence:**

```bash
# 1. Check status (fresh DB)
curl http://localhost:8888/status
# → {"total_memories": 0, "unconsolidated": 0, "consolidations": 0}

# 2. Ingest via API
curl -X POST http://localhost:8888/ingest \
  -H "Content-Type: application/json" \
  -d '{"text": "The project uses PostgreSQL for persistence and Redis for caching.", "source": "architecture-notes"}'
# → {"status": "ingested", "response": "..."}

# 3. Ingest via file drop
echo "Team meeting: decided to use TypeScript for the frontend rewrite." > inbox/meeting.txt
# Wait 5-10 seconds — watcher picks it up

# 4. Check memories
curl http://localhost:8888/memories
# → {"memories": [...], "count": 2}

# 5. Query
curl "http://localhost:8888/query?q=what+database+are+we+using"
# → {"question": "...", "answer": "... [Memory 1] ..."}

# 6. Trigger manual consolidation (need at least 2 memories)
curl -X POST http://localhost:8888/consolidate -H "Content-Type: application/json" -d '{}'
# → {"status": "done", "response": "..."}

# 7. Delete a specific memory
curl -X POST http://localhost:8888/delete \
  -H "Content-Type: application/json" \
  -d '{"memory_id": 1}'
# → {"status": "deleted", "memory_id": 1}

# 8. Full reset
curl -X POST http://localhost:8888/clear -H "Content-Type: application/json" -d '{}'
# → {"status": "cleared", "memories_deleted": N, "files_deleted": N}
```

**Validate checklist:**

- [ ] `python agent.py` starts and prints the exact 9-line startup log shown above
- [ ] `GET /status` returns `{"total_memories": 0, "unconsolidated": 0, "consolidations": 0}` on fresh DB
- [ ] `POST /ingest` stores a memory and returns `{"status": "ingested"}`
- [ ] Dropping a `.txt` file in `inbox/` causes it to appear in `/memories` within ~10 seconds
- [ ] `GET /query?q=...` returns an answer citing `[Memory N]` IDs
- [ ] `POST /consolidate` with 2+ memories returns an insight and sets `consolidated=1`
- [ ] `POST /delete {"memory_id": N}` removes the memory
- [ ] `POST /clear` resets DB and wipes inbox
- [ ] `memory.db` file created automatically on first run
- [ ] Restarting the agent does NOT re-ingest files already in `inbox/`

## Validation checklist

- [ ] `import sys` and `import time` are present in imports (Step 2)
- [ ] `MODEL` default is `"gemini-3.1-flash-lite-preview"` — not `"gemini-2.0-flash-lite"`
- [ ] All **seven** tools are defined: `store_memory`, `read_all_memories`, `read_unconsolidated_memories`, `store_consolidation`, `read_consolidation_history`, `get_memory_stats`, `delete_memory`, `clear_all_memories`
- [ ] `consolidate_agent` instruction ends with `"Think deeply about cross-cutting patterns."` (no "and contradictions")
- [ ] `query_agent` instruction ends with `"Be thorough but concise. Always cite sources."`
- [ ] `MemoryAgent.ingest()` uses `"Remember this information"` (not `"Remember this"`)
- [ ] `MemoryAgent.ingest_file()` logs with `🔮` emoji
- [ ] `MemoryAgent.status()` method exists
- [ ] `build_http()` registers 7 routes including `/delete` and `/clear`
- [ ] `main()` hardcodes defaults (`"./inbox"`, `8888`, `30`) — does NOT read from env vars
- [ ] `main_async()` logs the exact 6-line startup block: model, db, watch, consolidate, API, then blank line
- [ ] `consolidation_loop()` logs `result[:100]` after successful consolidation

## Notes

- **Model:** `gemini-3.1-flash-lite-preview` is the default — a preview model. If it's unavailable in your region, override with `MODEL=gemini-2.0-flash-lite` env var.
- **Single file architecture:** The reference repo is intentionally one file (`agent.py`). Do not split into `agents/` and `tools/` subdirectories — that would break the import structure.
- **`processed_files` table:** Tracks ingested file paths across restarts. Files dropped in `inbox/` before the agent starts will be ingested on first run; subsequent restarts will skip them.
- **`clear_all_memories()`** also deletes files from the inbox folder. The `/clear` endpoint triggers a full reset — use carefully.
- **ADK docstrings are required:** ADK uses them to describe tools to the LLM. Vague or missing docstrings lead to incorrect tool routing.
- **`sys` and `time` imports:** Present in the original repo even though they are not explicitly called in the main flow — do not remove them.
- **To integrate with the `knowledge` skill:** After running this agent, export memories via `GET /memories` and ingest that JSON into `.agent/knowledge/` as structured entries.
