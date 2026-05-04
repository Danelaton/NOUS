---
name: adk-memory-agent
description: Guides the full implementation of an always-on persistent memory system using Google ADK and Gemini. Creates a background agent that continuously ingests multimodal files, consolidates patterns, and answers queries — all backed by SQLite with no vector database. Use when building a project-level memory layer that runs 24/7 as a lightweight process.
---

# ADK Memory Agent

Implements an always-on memory agent using Google ADK + Gemini. The system watches a folder for new files (text, images, audio, video, PDFs), ingests them into SQLite, consolidates patterns on a timer, and serves queries via HTTP.

**Architecture:** Single `agent.py` with three specialist sub-agents (ingest, consolidate, query) routed by an orchestrator. All logic lives in one file — no microservices.

**Reference implementation:** [GoogleCloudPlatform/generative-ai — always-on-memory-agent](https://github.com/GoogleCloudPlatform/generative-ai/tree/main/gemini/agents/always-on-memory-agent)

## When to use

- When a project needs persistent memory that survives across sessions
- When building an agent that must process multimodal files automatically (images, audio, PDFs)
- When the user asks for a "memory layer", "knowledge base", or "always-on agent"
- When replacing ad-hoc note-taking with structured, queryable memory
- When you need a self-hosted alternative to vector databases for agent memory

## How to use

Follow the steps below in order. Each step is self-contained — you can run and test after each one. The final result is a single `agent.py` + optional `dashboard.py`.

---

## Steps

### Step 1 — Scaffold

Create the project structure:

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

### Step 2 — Memory store (SQLite tools)

At the top of `agent.py`, add the database setup and all five ADK tools. The schema has two tables:

```python
import argparse
import asyncio
import json
import logging
import mimetypes
import os
import shutil
import signal
import sqlite3
from datetime import datetime, timezone
from pathlib import Path

from aiohttp import web
from google.adk.agents import Agent
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService
from google.genai import types

# ─── Config ────────────────────────────────────────────────────

MODEL = os.getenv("MODEL", "gemini-2.0-flash-lite")
DB_PATH = os.getenv("MEMORY_DB", "memory.db")

TEXT_EXTENSIONS = {".txt", ".md", ".json", ".csv", ".log", ".xml", ".yaml", ".yml"}
MEDIA_EXTENSIONS = {
    ".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
    ".gif": "image/gif", ".webp": "image/webp", ".bmp": "image/bmp",
    ".svg": "image/svg+xml", ".mp3": "audio/mpeg", ".wav": "audio/wav",
    ".ogg": "audio/ogg", ".flac": "audio/flac", ".m4a": "audio/mp4",
    ".aac": "audio/aac", ".mp4": "video/mp4", ".webm": "video/webm",
    ".mov": "video/quicktime", ".avi": "video/x-msvideo",
    ".mkv": "video/x-matroska", ".pdf": "application/pdf",
}
ALL_SUPPORTED = TEXT_EXTENSIONS | set(MEDIA_EXTENSIONS.keys())

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s", datefmt="[%H:%M]")
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

Then add the five tools (exact signatures — ADK reads docstrings as tool descriptions):

```python
# ─── ADK Tools ─────────────────────────────────────────────────

def store_memory(
    raw_text: str, summary: str, entities: list[str],
    topics: list[str], importance: float, source: str = "",
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
        "INSERT INTO memories (source, raw_text, summary, entities, topics, importance, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
        (source, raw_text, summary, json.dumps(entities), json.dumps(topics), importance, now),
    )
    db.commit()
    mid = cursor.lastrowid
    db.close()
    log.info(f"📥 Stored memory #{mid}: {summary[:60]}...")
    return {"memory_id": mid, "status": "stored", "summary": summary}


def read_all_memories() -> dict:
    """Read all stored memories, most recent first (max 50).

    Returns:
        dict with list of memories and count.
    """
    db = get_db()
    rows = db.execute("SELECT * FROM memories ORDER BY created_at DESC LIMIT 50").fetchall()
    memories = [{
        "id": r["id"], "source": r["source"], "summary": r["summary"],
        "entities": json.loads(r["entities"]), "topics": json.loads(r["topics"]),
        "importance": r["importance"], "connections": json.loads(r["connections"]),
        "created_at": r["created_at"], "consolidated": bool(r["consolidated"]),
    } for r in rows]
    db.close()
    return {"memories": memories, "count": len(memories)}


def read_unconsolidated_memories() -> dict:
    """Read memories that haven't been consolidated yet (max 10).

    Returns:
        dict with list of unconsolidated memories and count.
    """
    db = get_db()
    rows = db.execute(
        "SELECT * FROM memories WHERE consolidated = 0 ORDER BY created_at DESC LIMIT 10"
    ).fetchall()
    memories = [{
        "id": r["id"], "summary": r["summary"],
        "entities": json.loads(r["entities"]), "topics": json.loads(r["topics"]),
        "importance": r["importance"], "created_at": r["created_at"],
    } for r in rows]
    db.close()
    return {"memories": memories, "count": len(memories)}


def store_consolidation(
    source_ids: list[int], summary: str, insight: str, connections: list[dict],
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
    log.info(f"🔄 Consolidated {len(source_ids)} memories: {insight[:80]}...")
    return {"status": "consolidated", "memories_processed": len(source_ids), "insight": insight}


def read_consolidation_history() -> dict:
    """Read past consolidation insights (max 10).

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
    return {"total_memories": total, "unconsolidated": unconsolidated, "consolidations": consolidations}
```

---

### Step 3 — Sub-agents (ADK definitions)

Add the three specialist agents. **Important:** ADK reads docstrings and `description` fields — make them precise.

```python
# ─── ADK Agents ────────────────────────────────────────────────

def build_agents():
    ingest_agent = Agent(
        name="ingest_agent",
        model=MODEL,
        description="Processes raw text or media into structured memory. Call this when new information arrives.",
        instruction=(
            "You are a Memory Ingest Agent. You handle ALL types of input — text, images, "
            "audio, video, and PDFs. For any input you receive:\n"
            "1. Thoroughly describe what the content contains\n"
            "2. Create a concise 1-2 sentence summary\n"
            "3. Extract key entities (people, companies, products, concepts, locations)\n"
            "4. Assign 2-4 topic tags\n"
            "5. Rate importance from 0.0 to 1.0\n"
            "6. Call store_memory with all extracted information\n\n"
            "For images: describe scene, objects, text, people, and visual details.\n"
            "For audio/video: describe spoken content, sounds, scenes, key moments.\n"
            "For PDFs: extract and summarize the document content.\n"
            "Use the full description as raw_text. Always call store_memory."
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
            "Think deeply about cross-cutting patterns and contradictions."
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
            "5. If no relevant memories exist, say so honestly"
        ),
        tools=[read_all_memories, read_consolidation_history],
    )
```

---

### Step 4 — Orchestrator (root agent)

Add the orchestrator and `MemoryAgent` runner class. The orchestrator receives all requests and routes to the right sub-agent.

```python
    orchestrator = Agent(
        name="memory_orchestrator",
        model=MODEL,
        description="Routes memory operations to specialist agents.",
        instruction=(
            "You are the Memory Orchestrator. Route requests to the right sub-agent:\n"
            "- New information or files -> ingest_agent\n"
            "- Consolidation request -> consolidate_agent\n"
            "- Questions about stored info -> query_agent\n"
            "- Status check -> call get_memory_stats and report\n"
            "After the sub-agent completes, give a brief summary."
        ),
        sub_agents=[ingest_agent, consolidate_agent, query_agent],
        tools=[get_memory_stats],
    )

    return orchestrator


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
        msg = f"Remember this (source: {source}):\n\n{text}" if source else f"Remember this:\n\n{text}"
        return await self.run(msg)

    async def ingest_file(self, file_path: Path) -> str:
        suffix = file_path.suffix.lower()
        mime_type = MEDIA_EXTENSIONS.get(suffix) or mimetypes.guess_type(str(file_path))[0] or "application/octet-stream"
        file_bytes = file_path.read_bytes()
        size_mb = len(file_bytes) / (1024 * 1024)
        if size_mb > 20:
            log.warning(f"⚠️  Skipping {file_path.name} ({size_mb:.1f}MB) — exceeds 20MB limit")
            return f"Skipped: file too large ({size_mb:.1f}MB)"
        prompt = (
            f"Remember this file (source: {file_path.name}, type: {mime_type}).\n\n"
            f"Analyze the content of this {mime_type.split('/')[0]} file and extract all meaningful information."
        )
        return await self.run_multimodal(prompt, file_bytes, mime_type)

    async def consolidate(self) -> str:
        return await self.run("Consolidate unconsolidated memories. Find connections and patterns.")

    async def query(self, question: str) -> str:
        return await self.run(f"Based on my memories, answer: {question}")
```

---

### Step 5 — File watcher

The watcher polls `inbox/` every 5 seconds. It tracks processed files in SQLite (`processed_files` table) to avoid double-ingestion on restart.

```python
# ─── File Watcher ──────────────────────────────────────────────

async def watch_folder(agent: MemoryAgent, folder: Path, poll_interval: int = 5):
    """Watch a folder for new files and ingest them automatically."""
    folder.mkdir(parents=True, exist_ok=True)
    db = get_db()
    log.info(f"👁️  Watching: {folder}/")

    while True:
        try:
            for f in sorted(folder.iterdir()):
                if f.name.startswith("."):
                    continue
                suffix = f.suffix.lower()
                if suffix not in ALL_SUPPORTED:
                    continue
                if db.execute("SELECT 1 FROM processed_files WHERE path = ?", (str(f),)).fetchone():
                    continue
                try:
                    if suffix in TEXT_EXTENSIONS:
                        log.info(f"📄 New text file: {f.name}")
                        text = f.read_text(encoding="utf-8", errors="replace")[:10000]
                        if text.strip():
                            await agent.ingest(text, source=f.name)
                    else:
                        log.info(f"🖼️  New media file: {f.name}")
                        await agent.ingest_file(f)
                except Exception as e:
                    log.error(f"Error ingesting {f.name}: {e}")

                db.execute(
                    "INSERT INTO processed_files (path, processed_at) VALUES (?, ?)",
                    (str(f), datetime.now(timezone.utc).isoformat()),
                )
                db.commit()
        except Exception as e:
            log.error(f"Watch error: {e}")

        await asyncio.sleep(poll_interval)
```

Key design choices:
- Processed files are tracked in DB (not moved) so the `inbox/` stays clean
- Text files are read as string (max 10k chars); media files are sent as bytes
- 20MB per-file limit matches Gemini inline upload limit

---

### Step 6 — API server (aiohttp)

```python
# ─── HTTP API ──────────────────────────────────────────────────

def build_http(agent: MemoryAgent):
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
        result = await agent.ingest(text, source=data.get("source", "api"))
        return web.json_response({"status": "ingested", "response": result})

    async def handle_consolidate(request: web.Request):
        result = await agent.consolidate()
        return web.json_response({"status": "done", "response": result})

    async def handle_status(request: web.Request):
        return web.json_response(get_memory_stats())

    async def handle_memories(request: web.Request):
        return web.json_response(read_all_memories())

    app.router.add_get("/query", handle_query)
    app.router.add_post("/ingest", handle_ingest)
    app.router.add_post("/consolidate", handle_consolidate)
    app.router.add_get("/status", handle_status)
    app.router.add_get("/memories", handle_memories)

    return app
```

API endpoints summary:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/query?q=...` | GET | Query memory with a question |
| `/ingest` | POST | Ingest text `{"text": "...", "source": "..."}` |
| `/consolidate` | POST | Trigger manual consolidation |
| `/status` | GET | Memory statistics |
| `/memories` | GET | List all stored memories |

---

### Step 7 — Consolidation loop + main

```python
# ─── Consolidation Loop ──────────────────────────────────────

async def consolidation_loop(agent: MemoryAgent, interval_minutes: int = 30):
    """Run consolidation periodically. Skips if fewer than 2 unconsolidated memories."""
    log.info(f"🔄 Consolidation: every {interval_minutes} minutes")
    while True:
        await asyncio.sleep(interval_minutes * 60)
        try:
            db = get_db()
            count = db.execute("SELECT COUNT(*) as c FROM memories WHERE consolidated = 0").fetchone()["c"]
            db.close()
            if count >= 2:
                log.info(f"🔄 Running consolidation ({count} unconsolidated memories)...")
                await agent.consolidate()
            else:
                log.info(f"🔄 Skipping ({count} unconsolidated memories — need at least 2)")
        except Exception as e:
            log.error(f"Consolidation error: {e}")


# ─── Main ──────────────────────────────────────────────────────

async def main_async(args):
    agent = MemoryAgent()

    log.info(f"🧠 Memory agent starting — model: {MODEL}, db: {DB_PATH}")
    log.info(f"   Watch: {args.watch} | Port: {args.port} | Consolidate every: {args.consolidate_every}m")

    tasks = [
        asyncio.create_task(watch_folder(agent, Path(args.watch))),
        asyncio.create_task(consolidation_loop(agent, args.consolidate_every)),
    ]

    app = build_http(agent)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, "0.0.0.0", args.port)
    await site.start()

    log.info(f"✅ Running. Drop files in {args.watch}/ or POST to http://localhost:{args.port}/ingest")

    try:
        await asyncio.gather(*tasks)
    except asyncio.CancelledError:
        pass
    finally:
        await runner.cleanup()


def main():
    parser = argparse.ArgumentParser(description="Always-On ADK Memory Agent")
    parser.add_argument("--watch", default=os.getenv("INBOX_PATH", "./inbox"))
    parser.add_argument("--port", type=int, default=int(os.getenv("API_PORT", "8888")))
    parser.add_argument("--consolidate-every", type=int,
                        default=int(os.getenv("CONSOLIDATION_INTERVAL_MINUTES", "30")))
    args = parser.parse_args()

    loop = asyncio.new_event_loop()

    def shutdown(sig):
        log.info(f"\n👋 Shutting down...")
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


if __name__ == "__main__":
    main()
```

---

### Step 8 — Dashboard (optional, Streamlit)

The dashboard connects to the running agent via HTTP — it does NOT import `agent.py`. Start the agent first, then the dashboard in a separate terminal.

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
                color = "#4ade80" if importance >= 0.7 else "#fbbf24" if importance >= 0.4 else "#aaa"
                st.markdown(
                    f"**[#{m['id']}]** {m['summary']}  \n"
                    f"`{'  '.join(m.get('topics', []))}` | importance: {importance} | {m.get('source','')}"
                )
                st.divider()

if __name__ == "__main__":
    main()
```

Start: `streamlit run dashboard.py` (agent must be running at `:8888`)

---

### Step 9 — Configuration

**`requirements.txt`:**

```
google-adk>=1.0.0
google-genai>=1.0.0
aiohttp>=3.9.0
streamlit>=1.40.0
requests>=2.31.0
```

**`.env.example`:**

```bash
# Required
GOOGLE_API_KEY=your-gemini-api-key

# Optional — all have defaults
MODEL=gemini-2.0-flash-lite
MEMORY_DB=./memory.db
INBOX_PATH=./inbox
API_PORT=8888
CONSOLIDATION_INTERVAL_MINUTES=30
```

**Install and configure:**

```bash
pip install -r requirements.txt
cp .env.example .env
# Edit .env and add your GOOGLE_API_KEY
# Get key from: https://aistudio.google.com/
export GOOGLE_API_KEY="your-key-here"
```

---

### Step 10 — Test and validate

**Start the agent:**

```bash
python agent.py
# Expected output:
# [HH:MM] 🧠 Memory agent starting — model: gemini-2.0-flash-lite, db: memory.db
# [HH:MM]    Watch: ./inbox | Port: 8888 | Consolidate every: 30m
# [HH:MM] 👁️  Watching: inbox/
# [HH:MM] 🔄 Consolidation: every 30 minutes
# [HH:MM] ✅ Running. Drop files in ./inbox/ or POST to http://localhost:8888/ingest
```

**Test ingest via API:**

```bash
curl -X POST http://localhost:8888/ingest \
  -H "Content-Type: application/json" \
  -d '{"text": "The project uses PostgreSQL for persistence and Redis for caching.", "source": "architecture-notes"}'
# Expected: {"status": "ingested", "response": "..."}
```

**Test ingest via file drop:**

```bash
echo "Team meeting: decided to use TypeScript for the frontend rewrite." > inbox/meeting.txt
# Wait 5-10 seconds — agent picks it up automatically
```

**Test query:**

```bash
curl "http://localhost:8888/query?q=what+database+are+we+using"
# Expected: answer citing memory IDs
```

**Test status:**

```bash
curl http://localhost:8888/status
# Expected: {"total_memories": N, "unconsolidated": N, "consolidations": N}
```

**Test consolidation (manual trigger):**

```bash
# Ingest at least 2 items first, then:
curl -X POST http://localhost:8888/consolidate -d '{}'
```

**Validate checklist:**

- [ ] `python agent.py` starts without errors
- [ ] `GET /status` returns `{"total_memories": 0, "unconsolidated": 0, "consolidations": 0}` on fresh DB
- [ ] `POST /ingest` returns `{"status": "ingested"}` with a response from the LLM
- [ ] `GET /memories` returns the ingested entry
- [ ] Dropping a `.txt` file in `inbox/` causes it to appear in `/memories` within ~10 seconds
- [ ] `GET /query?q=...` returns a coherent answer citing memory IDs
- [ ] After 2+ ingests: `POST /consolidate` produces an insight and sets `consolidated=1` on source memories
- [ ] `memory.db` file is created automatically on first run

## Validation checklist

- [ ] `agent.py` is a single file — no module imports from `agents/` or `tools/` subdirectories
- [ ] All five tools have docstrings with `Args:` and `Returns:` sections (ADK uses these)
- [ ] Orchestrator has `sub_agents=[ingest_agent, consolidate_agent, query_agent]`
- [ ] `MemoryAgent._execute()` iterates `runner.run_async()` to collect streamed response parts
- [ ] File watcher uses `processed_files` table to prevent re-ingestion on restart
- [ ] Consolidation loop checks `count >= 2` before running (avoids empty consolidations)
- [ ] `GOOGLE_API_KEY` is set in environment before starting

## Notes

- The reference repo uses a **single `agent.py`** file — not `agents/` and `tools/` subdirectories. This is intentional: simpler deployment, no import complexity.
- Model is `gemini-2.0-flash-lite` by default (cost-effective for 24/7 background use). Override with `MODEL=gemini-2.0-flash` env var for better quality.
- The `processed_files` table persists across restarts — files in `inbox/` that were already ingested will not be re-processed even if the agent restarts.
- Gemini inline upload limit is ~20MB per file. Larger files are skipped with a warning.
- The dashboard (`dashboard.py`) is optional. The agent works fully headless via the HTTP API.
- ADK tool docstrings are **not optional** — ADK uses them to describe tools to the LLM. Missing or vague docstrings lead to incorrect tool calls.
- If you see `google.api_core.exceptions.InvalidArgument`, check that your `GOOGLE_API_KEY` is valid and has Gemini API access enabled.
- To integrate with the `knowledge` skill: after running this agent, you can ingest `memory.db` exports or the `/memories` API output into `.agent/knowledge/` as structured entries.
