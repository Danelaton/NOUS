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

## Steps

### Step 1 — Prerequisites

Verify you have Python 3.10+ and a Google AI API key:

```bash
python --version   # must be 3.10+
echo $GOOGLE_API_KEY  # must be set
```

Get a Gemini API key at: https://aistudio.google.com/app/apikey

---

### Step 2 — Clone the repo

Clone only the subdirectory you need (sparse checkout avoids downloading the full 40GB repo):

```bash
# Option A — sparse checkout (recommended, fast)
git clone --no-checkout --depth=1 https://github.com/GoogleCloudPlatform/generative-ai.git tmp-google-ai
cd tmp-google-ai
git sparse-checkout init --cone
git sparse-checkout set gemini/agents/always-on-memory-agent
git checkout main
cp -r gemini/agents/always-on-memory-agent ../memory-agent
cd ..
rm -rf tmp-google-ai
cd memory-agent

# Option B — direct download (if git sparse checkout unavailable)
curl -L "https://github.com/GoogleCloudPlatform/generative-ai/archive/refs/heads/main.tar.gz" | \
  tar -xz --strip-components=3 "generative-ai-main/gemini/agents/always-on-memory-agent"
```

Verify the contents:

```bash
ls -la
# Should see: agent.py  requirements.txt  README.md
```

---

### Step 3 — Install dependencies

```bash
pip install -r requirements.txt
```

Core packages installed: `google-adk`, `google-genai`, `aiohttp`, `watchdog`, `streamlit`, `python-dotenv`.

---

### Step 4 — Configure environment

Create `.env` in the `memory-agent/` directory:

```bash
cat > .env << 'EOF'
GOOGLE_API_KEY=your-gemini-api-key-here
CONSOLIDATION_INTERVAL_MINUTES=30
API_PORT=8888
INBOX_PATH=./inbox
DB_PATH=./memory.db
EOF
```

Then set it in your shell session as well:

```bash
export GOOGLE_API_KEY="your-gemini-api-key-here"
```

**Note:** Never commit `.env` to git. Add it to `.gitignore`:

```bash
echo ".env" >> .gitignore
echo "memory.db" >> .gitignore
echo "inbox/" >> .gitignore
```

---

### Step 5 — Create inbox folder

```bash
mkdir -p inbox
```

Files dropped here are automatically picked up by the file watcher and ingested into memory.

---

### Step 6 — Run the agent

```bash
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
