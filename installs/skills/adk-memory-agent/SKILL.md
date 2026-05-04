---
name: adk-memory-agent
description: Guides installation and configuration of the Google ADK always-on memory agent — a background process that ingests, consolidates, and serves project knowledge via HTTP API using Gemini. Use when setting up persistent memory infrastructure for AI agents in a Python/Gemini environment.
---

# ADK Memory Agent

Installs and configures the [always-on-memory-agent](https://github.com/GoogleCloudPlatform/generative-ai/tree/main/gemini/agents/always-on-memory-agent) from Google's Generative AI repo. This is a ready-to-run background service — not something to reimplement.

**What it provides:**
- HTTP API for ingesting text, files, and queries
- Automatic memory consolidation (periodic background task)
- SQLite persistence + multi-modal input support (text, images, audio, video, PDF)
- Streamlit dashboard (optional)
- `inbox/` folder watcher — drop files, they get auto-ingested

**Stack:** Python, Google ADK, Gemini 2.0 Flash, SQLite, aiohttp

## When to use

- When you need a persistent memory backend for AI agents in a Python + Gemini project
- When you want to ingest files/docs automatically via a folder watcher
- When you want a standalone memory service that other tools can query via HTTP

## How to use

Clone the Google repo, configure `.env`, run `python agent.py`. The agent serves a REST API on `localhost:8888`.

---

## What this installs in your project

The memory agent runs as a **sidecar service** alongside your existing project. It adds:

```
your-project/
└── memory-agent/          ← new subfolder (isolated, does not touch existing files)
    ├── agent.py
    ├── requirements.txt
    ├── .env                ← NEW (add to .gitignore)
    ├── memory.db           ← NEW (add to .gitignore)
    └── inbox/              ← NEW (add to .gitignore)
```

**Nothing in your existing project is modified.** No existing `requirements.txt`, `.env`, `package.json`, or `.gitignore` is touched. The agent installs into its own subfolder and communicates with the rest of the project via HTTP API.

---

## Rollback / Uninstall

To remove completely:

```bash
# Stop the agent if running
kill $(cat memory-agent/memory-agent.pid 2>/dev/null) 2>/dev/null || true

# Remove the subfolder and all its data
rm -rf memory-agent/

# If you registered a system service:
# macOS: launchctl unload ~/Library/LaunchAgents/com.memory-agent.plist && rm ~/Library/LaunchAgents/com.memory-agent.plist
# Linux: sudo systemctl stop memory-agent && sudo systemctl disable memory-agent && sudo rm /etc/systemd/system/memory-agent.service
```

No rollback needed for your existing project files — nothing was changed there.

---

## Steps

### Step 1 — Prerequisites

Verify you have Python 3.10+ and a Google AI API key:

```bash
python --version   # must be 3.10+
echo $GOOGLE_API_KEY  # must be set
```

Get a Gemini API key at: https://aistudio.google.com/app/apikey

---

### Step 2 — Clone the repo into a subfolder

The agent lives in its own `memory-agent/` subfolder — completely separate from your existing code.

```bash
# From the root of your project:

# Option A — sparse checkout (recommended, fast — avoids downloading the full 40GB repo)
git clone --no-checkout --depth=1 https://github.com/GoogleCloudPlatform/generative-ai.git _tmp_google_ai
cd _tmp_google_ai
git sparse-checkout init --cone
git sparse-checkout set gemini/agents/always-on-memory-agent
git checkout main
cd ..
cp -r _tmp_google_ai/gemini/agents/always-on-memory-agent memory-agent
rm -rf _tmp_google_ai

# Option B — direct download (if git sparse checkout unavailable)
curl -L "https://github.com/GoogleCloudPlatform/generative-ai/archive/refs/heads/main.tar.gz" | \
  tar -xz --strip-components=3 -C memory-agent "generative-ai-main/gemini/agents/always-on-memory-agent"
```

Verify — you should now have a `memory-agent/` subfolder in your project root:

```bash
ls memory-agent/
# Should see: agent.py  requirements.txt  README.md
```

---

### Step 3 — Install dependencies

The agent has its own `requirements.txt` inside `memory-agent/`. Install into a virtual environment to avoid conflicts with your project's existing packages:

```bash
cd memory-agent/

# Create a dedicated virtualenv (recommended — does not affect your project's deps)
python -m venv .venv
source .venv/bin/activate       # macOS/Linux
# .venv\Scripts\activate        # Windows

pip install -r requirements.txt
```

Core packages installed: `google-adk`, `google-genai`, `aiohttp`, `watchdog`, `streamlit`, `python-dotenv`.

**Do NOT merge these dependencies into your project's existing `requirements.txt`** — the agent runs as a separate process and doesn't need to share the same Python environment.

---

### Step 4 — Configure environment

The agent needs its own `.env` inside `memory-agent/`. **Do not modify your project's existing `.env`** — this is a separate file scoped to the agent subfolder.

```bash
# Inside memory-agent/ — create a NEW .env file for the agent only
cat > memory-agent/.env << 'EOF'
GOOGLE_API_KEY=your-gemini-api-key-here
CONSOLIDATION_INTERVAL_MINUTES=30
API_PORT=8888
INBOX_PATH=./inbox
DB_PATH=./memory.db
EOF
```

**If your project already has a root-level `.env`**, add the key there as an alternative — the agent reads from its own directory first:

```bash
# Option: append to existing project .env instead
echo "" >> .env
echo "# ADK Memory Agent" >> .env
echo "GOOGLE_API_KEY=your-gemini-api-key-here" >> .env
```

**Update `.gitignore`** — append these lines without overwriting the file:

```bash
# Check what's already ignored first
cat .gitignore | grep -E "\.env|memory\.db|inbox"

# Append only what's missing (safe — >> never truncates)
echo "" >> .gitignore
echo "# ADK Memory Agent" >> .gitignore
grep -qxF "memory-agent/.env" .gitignore     || echo "memory-agent/.env" >> .gitignore
grep -qxF "memory-agent/memory.db" .gitignore || echo "memory-agent/memory.db" >> .gitignore
grep -qxF "memory-agent/inbox/" .gitignore   || echo "memory-agent/inbox/" >> .gitignore
grep -qxF "memory-agent/.venv/" .gitignore   || echo "memory-agent/.venv/" >> .gitignore
```

The `grep -qxF ... ||` pattern only adds a line if it doesn't already exist — safe to run multiple times.

---

### Step 5 — Create inbox folder

```bash
mkdir -p memory-agent/inbox
```

Files dropped here are automatically picked up by the file watcher and ingested into memory.

---

### Step 6 — Run the agent

```bash
cd memory-agent/
source .venv/bin/activate   # activate the virtualenv created in Step 3
python agent.py
```

Expected output on success:

```
INFO: Starting memory agent on port 8888
INFO: File watcher started on ./inbox
INFO: Consolidation loop started (every 30 minutes)
INFO: Agent ready
```

Keep this terminal open. The agent runs as a foreground process.

---

### Step 7 — Verify it works

In a new terminal, run the smoke tests:

```bash
# Check status
curl http://localhost:8888/status
# Expected: {"memories_count": 0, "consolidations_count": 0, "last_consolidation": null}

# Ingest a memory
curl -X POST http://localhost:8888/ingest \
  -H "Content-Type: application/json" \
  -d '{"text": "The project uses PostgreSQL 15 with pgvector extension"}'
# Expected: {"status": "ok", "id": <number>}

# Query memory
curl "http://localhost:8888/query?q=database+technology"
# Expected: response referencing PostgreSQL

# List all memories
curl http://localhost:8888/memories
# Expected: {"memories": [...], "count": 1}
```

---

### Step 8 — Ingest a file (optional)

Drop any supported file into `inbox/`:

```bash
cp /path/to/your/docs/spec.pdf inbox/
# The file watcher detects it, ingests it, and moves it to inbox/processed/
```

Supported formats: `.txt`, `.md`, `.json`, `.csv`, `.log`, `.xml`, `.yaml`, `.png`, `.jpg`, `.gif`, `.webp`, `.mp3`, `.wav`, `.mp4`, `.webm`, `.pdf`

---

### Step 9 — Run as background service (optional)

**macOS/Linux — using nohup:**

```bash
nohup python agent.py > memory-agent.log 2>&1 &
echo $! > memory-agent.pid
echo "Agent started with PID $(cat memory-agent.pid)"
```

Stop it:

```bash
kill $(cat memory-agent.pid)
```

**macOS — using launchd (persistent across reboots):**

Create `~/Library/LaunchAgents/com.memory-agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.memory-agent</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/bin/python3</string>
    <string>/absolute/path/to/memory-agent/agent.py</string>
  </array>
  <key>WorkingDirectory</key><string>/absolute/path/to/memory-agent</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>GOOGLE_API_KEY</key><string>your-key-here</string>
  </dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>/tmp/memory-agent.log</string>
  <key>StandardErrorPath</key><string>/tmp/memory-agent.err</string>
</dict>
</plist>
```

Load it:

```bash
launchctl load ~/Library/LaunchAgents/com.memory-agent.plist
```

**Linux — using systemd:**

Create `/etc/systemd/system/memory-agent.service`:

```ini
[Unit]
Description=ADK Memory Agent
After=network.target

[Service]
Type=simple
WorkingDirectory=/absolute/path/to/memory-agent
ExecStart=/usr/bin/python3 /absolute/path/to/memory-agent/agent.py
EnvironmentFile=/absolute/path/to/memory-agent/.env
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable memory-agent
sudo systemctl start memory-agent
sudo systemctl status memory-agent
```

---

### Step 10 — Streamlit dashboard (optional)

The agent includes an optional web dashboard:

```bash
pip install streamlit  # if not already installed via requirements.txt
streamlit run dashboard.py  # check if dashboard.py exists in the repo
```

Dashboard shows: all memories, consolidation history, query interface, file upload.

---

### Step 11 — Integrate with your project

Once the agent is running, connect your app to it via HTTP:

```python
import httpx

MEMORY_API = "http://localhost:8888"

# Store a memory
httpx.post(f"{MEMORY_API}/ingest", json={"text": "User prefers dark mode"})

# Query memory
result = httpx.get(f"{MEMORY_API}/query", params={"q": "user preferences"})
memories = result.json()

# Get all memories
all_mem = httpx.get(f"{MEMORY_API}/memories").json()
```

**Environment variable pattern (recommended):**

```python
import os
MEMORY_API = os.environ.get("MEMORY_API_URL", "http://localhost:8888")
```

---

## Validation checklist

- [ ] `python agent.py` starts without errors
- [ ] `curl http://localhost:8888/status` returns JSON with `memories_count`
- [ ] `POST /ingest` returns `{"status": "ok"}`
- [ ] `GET /query?q=<text>` returns relevant results after ingesting
- [ ] File dropped in `inbox/` moves to `inbox/processed/` after processing
- [ ] `.env` is in `.gitignore` — no credentials in git

## Notes

- **Repo location:** `GoogleCloudPlatform/generative-ai` → `gemini/agents/always-on-memory-agent/`
- **Model used:** Gemini 2.0 Flash (orchestrator) + Gemini 2.0 Flash Lite (sub-agents) — requires valid `GOOGLE_API_KEY`
- **Port:** 8888 by default, configurable via `API_PORT` env var
- **Consolidation:** Runs every 30 minutes automatically. Connects patterns across memories using Gemini.
- **File limit:** 20MB per file in inbox/
- **Persistence:** `memory.db` SQLite file — back it up periodically if data is important
- **The agent is stateful:** Restarting it keeps all memories (SQLite persists). Only `DB_PATH` deletion wipes memory.
- If `PROJECT_MAP.md` exists in your project, consider ingesting it: `curl -X POST http://localhost:8888/ingest -d '{"text": "...contents..."}'`
