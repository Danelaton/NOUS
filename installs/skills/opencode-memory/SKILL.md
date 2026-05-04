---
name: opencode-memory
description: Guides the full implementation of opencode-mem — a persistent memory plugin for OpenCode that uses local SQLite + USearch vector indexing, HuggingFace local embeddings, auto-capture, user profile learning, smart deduplication, and a web UI dashboard. Use when building a memory layer for AI coding agents in an OpenCode environment with cross-session context retention.
---

# OpenCode Memory Plugin

Implements `opencode-mem` — a TypeScript plugin for [OpenCode](https://opencode.ai) that gives coding agents persistent memory using a local vector database (SQLite + USearch). No cloud required.

**Reference repo:** [tickernelz/opencode-mem](https://github.com/tickernelz/opencode-mem) (v2.13.0)

**Stack:** TypeScript + Bun, `@opencode-ai/plugin`, SQLite, USearch, HuggingFace Transformers (local), aiohttp-style web server, Streamlit-free Web UI at `localhost:4747`.

**Architecture overview:**

```
opencode-mem/
├── src/
│   ├── plugin.ts           ← Entry point — exports PluginModule
│   ├── index.ts            ← OpenCodeMemPlugin: chat.message + session.completed hooks
│   ├── config.ts           ← JSONC config loader + all defaults
│   ├── types/index.ts      ← MemoryType, MemoryMetadata, AIProviderType
│   ├── services/
│   │   ├── client.ts              ← LocalMemoryClient (search, add, delete, list)
│   │   ├── embedding.ts           ← EmbeddingService (local HF + API fallback)
│   │   ├── auto-capture.ts        ← performAutoCapture() — idle session analysis
│   │   ├── deduplication-service.ts
│   │   ├── cleanup-service.ts     ← 30-day retention, pinned memory protection
│   │   ├── user-memory-learning.ts ← User profile preference learning
│   │   ├── web-server.ts          ← HTTP server at :4747
│   │   ├── web-server-worker.ts
│   │   ├── migration-service.ts
│   │   ├── api-handlers.ts        ← REST API for web UI
│   │   ├── context.ts             ← formatContextForPrompt()
│   │   ├── privacy.ts             ← stripPrivateContent()
│   │   ├── logger.ts              ← log()
│   │   ├── tags.ts                ← getTags() — project + user hash
│   │   ├── jsonc.ts               ← JSONC comment stripper
│   │   ├── secret-resolver.ts
│   │   ├── language-detector.ts
│   │   ├── sqlite/
│   │   │   ├── connection-manager.ts
│   │   │   ├── shard-manager.ts
│   │   │   ├── vector-search.ts
│   │   │   ├── types.ts
│   │   │   └── migration-runner.ts
│   │   ├── vector-backends/
│   │   │   ├── usearch-backend.ts
│   │   │   └── exact-scan-backend.ts
│   │   ├── user-profile/
│   │   │   ├── user-profile-manager.ts
│   │   │   └── user-profile-types.ts
│   │   ├── user-prompt/
│   │   │   └── user-prompt-manager.ts
│   │   └── ai/
│   │       └── opencode-provider.ts
│   └── web/
│       ├── index.html
│       ├── app.js
│       ├── styles.css
│       └── i18n.js
├── package.json
├── tsconfig.json
└── dist/           ← built output (bunx tsc)
```

## When to use

- When OpenCode sessions need persistent memory that survives across restarts
- When building a coding agent that should remember architectural decisions, preferences, and past context
- When you want semantic memory search (vector similarity) without a cloud backend
- When you need per-project memory scoping plus cross-project search
- When you want an automatic memory capture system that runs on session completion

## How to use

Follow the steps in order. Each step builds on the previous. The full project is 25+ files — copy each snippet exactly as shown.

---

## Steps

### Step 1 — Scaffold

```bash
mkdir opencode-mem
cd opencode-mem
git init
```

Create the full directory structure:

```bash
mkdir -p src/services/sqlite
mkdir -p src/services/vector-backends
mkdir -p src/services/user-profile
mkdir -p src/services/user-prompt
mkdir -p src/services/ai
mkdir -p src/types
mkdir -p src/web
mkdir -p dist
```

**`package.json`** (exact from repo):

```json
{
  "name": "opencode-mem",
  "version": "2.13.0",
  "description": "OpenCode plugin that gives coding agents persistent memory using local vector database",
  "type": "module",
  "main": "dist/plugin.js",
  "types": "dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/plugin.js",
      "types": "./dist/index.d.ts"
    },
    "./server": {
      "import": "./dist/plugin.js",
      "types": "./dist/index.d.ts"
    }
  },
  "scripts": {
    "build": "bunx tsc && mkdir -p dist/web && cp -r src/web/* dist/web/",
    "dev": "tsc --watch",
    "typecheck": "tsc --noEmit",
    "format": "prettier --write \"src/**/*.{ts,js,css,html}\""
  },
  "keywords": ["opencode", "plugin", "memory", "vector-database", "ai", "coding-agent"],
  "author": "opencode-mem",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/tickernelz/opencode-mem.git"
  },
  "publishConfig": { "access": "public" },
  "dependencies": {
    "@ai-sdk/anthropic": "^3.0.58",
    "@ai-sdk/openai": "^3.0.41",
    "@huggingface/transformers": "^4.0.1",
    "@opencode-ai/plugin": "^1.3.0",
    "@opencode-ai/sdk": "^1.3.0",
    "ai": "^6.0.116",
    "franc-min": "^6.2.0",
    "iso-639-3": "^3.0.1",
    "usearch": "^2.21.4",
    "zod": "^4.3.6"
  },
  "devDependencies": {
    "@types/bun": "^1.3.8",
    "husky": "^9.1.7",
    "lint-staged": "^16.2.7",
    "prettier": "^3.4.2",
    "typescript": "^5.7.3"
  },
  "opencode": {
    "type": "plugin",
    "hooks": ["chat.message", "event"]
  },
  "files": ["dist", "package.json"]
}
```

**`tsconfig.json`** (exact from repo):

```json
{
  "compilerOptions": {
    "lib": ["ESNext"],
    "target": "ESNext",
    "module": "ESNext",
    "moduleDetection": "force",
    "allowJs": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": false,
    "verbatimModuleSyntax": true,
    "esModuleInterop": true,
    "resolveJsonModule": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true,
    "strict": true,
    "skipLibCheck": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noPropertyAccessFromIndexSignature": false
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

Install dependencies:

```bash
bun install
```

---

### Step 2 — Types and plugin entry point

**`src/types/index.ts`** (exact from repo):

```typescript
export type MemoryType = string;

export interface MemoryMetadata {
  type?: MemoryType;
  source?: "manual" | "auto-capture" | "import" | "api";
  tool?: string;
  sessionID?: string;
  reasoning?: string;
  captureTimestamp?: number;
  promptId?: string;
  displayName?: string;
  userName?: string;
  userEmail?: string;
  projectPath?: string;
  projectName?: string;
  gitRepoUrl?: string;
  [key: string]: unknown;
}

export type AIProviderType = "openai-chat" | "openai-responses" | "anthropic";
```

**`src/plugin.ts`** (exact from repo):

```typescript
import type { PluginModule } from "@opencode-ai/plugin";
import pkg from "../package.json";
const { OpenCodeMemPlugin } = await import("./index.js");

export const id =
  typeof pkg.name === "string" && pkg.name.trim() ? pkg.name.trim() : "opencode-mem";
export { OpenCodeMemPlugin };
export default { id, server: OpenCodeMemPlugin } satisfies PluginModule;
```

Key points:
- Uses top-level `await import()` for dynamic loading — required by Bun/ESM
- `id` is derived from `package.json` `name` field
- Exports both named and default export for `PluginModule` compatibility

---

### Step 3 — Configuration system

**`src/config.ts`** — Full JSONC config loader with all defaults. Config is read from `~/.config/opencode/opencode-mem.jsonc`, then merged with project-level `.opencode/opencode-mem.jsonc`:

```typescript
import { existsSync, readFileSync, mkdirSync } from "node:fs";
import { join } from "node:path";
import { homedir } from "node:os";
import { stripJsoncComments } from "./services/jsonc.js";
import { resolveSecretValue } from "./services/secret-resolver.js";

const CONFIG_DIR = join(homedir(), ".config", "opencode");
const DATA_DIR = join(homedir(), ".opencode-mem");
const CONFIG_FILES = [
  join(CONFIG_DIR, "opencode-mem.jsonc"),
  join(CONFIG_DIR, "opencode-mem.json"),
];

if (!existsSync(CONFIG_DIR)) mkdirSync(CONFIG_DIR, { recursive: true });
if (!existsSync(DATA_DIR)) mkdirSync(DATA_DIR, { recursive: true });

interface OpenCodeMemConfig {
  storagePath?: string;
  userEmailOverride?: string;
  userNameOverride?: string;
  memory?: { defaultScope?: "project" | "all-projects" };
  embeddingModel?: string;
  embeddingDimensions?: number;
  embeddingApiUrl?: string;
  embeddingApiKey?: string;
  similarityThreshold?: number;
  maxMemories?: number;
  maxProfileItems?: number;
  injectProfile?: boolean;
  containerTagPrefix?: string;
  autoCaptureEnabled?: boolean;
  autoCaptureMaxIterations?: number;
  autoCaptureIterationTimeout?: number;
  autoCaptureLanguage?: string;
  memoryProvider?: "openai-chat" | "openai-responses" | "anthropic";
  memoryModel?: string;
  memoryApiUrl?: string;
  memoryApiKey?: string;
  memoryTemperature?: number | false;
  memoryExtraParams?: Record<string, unknown>;
  opencodeProvider?: string;
  opencodeModel?: string;
  vectorBackend?: "usearch-first" | "usearch" | "exact-scan";
  aiSessionRetentionDays?: number;
  webServerEnabled?: boolean;
  webServerPort?: number;
  webServerHost?: string;
  maxVectorsPerShard?: number;
  autoCleanupEnabled?: boolean;
  autoCleanupRetentionDays?: number;
  deduplicationEnabled?: boolean;
  deduplicationSimilarityThreshold?: number;
  userProfileAnalysisInterval?: number;
  userProfileMaxPreferences?: number;
  userProfileMaxPatterns?: number;
  userProfileMaxWorkflows?: number;
  userProfileConfidenceDecayDays?: number;
  userProfileChangelogRetentionCount?: number;
  showAutoCaptureToasts?: boolean;
  showUserProfileToasts?: boolean;
  showErrorToasts?: boolean;
  compaction?: { enabled?: boolean; memoryLimit?: number };
  chatMessage?: {
    enabled?: boolean;
    maxMemories?: number;
    excludeCurrentSession?: boolean;
    maxAgeDays?: number;
    injectOn?: "first" | "always";
  };
}

const DEFAULTS = {
  storagePath: join(DATA_DIR, "data"),
  embeddingModel: "Xenova/nomic-embed-text-v1",
  embeddingDimensions: 768,
  similarityThreshold: 0.6,
  maxMemories: 10,
  maxProfileItems: 5,
  injectProfile: true,
  containerTagPrefix: "opencode",
  autoCaptureEnabled: true,
  autoCaptureMaxIterations: 5,
  autoCaptureIterationTimeout: 30000,
  vectorBackend: "usearch-first" as const,
  aiSessionRetentionDays: 7,
  webServerEnabled: true,
  webServerPort: 4747,
  webServerHost: "127.0.0.1",
  maxVectorsPerShard: 50000,
  autoCleanupEnabled: true,
  autoCleanupRetentionDays: 30,
  deduplicationEnabled: true,
  deduplicationSimilarityThreshold: 0.9,
  userProfileAnalysisInterval: 10,
  userProfileMaxPreferences: 20,
  userProfileMaxPatterns: 15,
  userProfileMaxWorkflows: 10,
  userProfileConfidenceDecayDays: 30,
  userProfileChangelogRetentionCount: 5,
  showAutoCaptureToasts: true,
  showUserProfileToasts: true,
  showErrorToasts: true,
  memory: { defaultScope: "project" as const },
  compaction: { enabled: true, memoryLimit: 10 },
  chatMessage: {
    enabled: true,
    maxMemories: 3,
    excludeCurrentSession: true,
    maxAgeDays: undefined as number | undefined,
    injectOn: "first" as const,
  },
};

function expandPath(path: string): string {
  if (path.startsWith("~/")) return join(homedir(), path.slice(2));
  if (path === "~") return homedir();
  return path;
}

function loadConfigFromPaths(paths: string[]): OpenCodeMemConfig {
  for (const path of paths) {
    if (existsSync(path)) {
      try {
        const content = readFileSync(path, "utf-8");
        const json = stripJsoncComments(content);
        return JSON.parse(json) as OpenCodeMemConfig;
      } catch {}
    }
  }
  return {};
}

// Embedding dimension map — 30+ models supported
function getEmbeddingDimensions(model: string): number {
  const dimensionMap: Record<string, number> = {
    "Xenova/nomic-embed-text-v1": 768,
    "Xenova/nomic-embed-text-v1-unsupervised": 768,
    "Xenova/nomic-embed-text-v1-ablated": 768,
    "Xenova/jina-embeddings-v2-base-en": 768,
    "Xenova/jina-embeddings-v2-base-zh": 768,
    "Xenova/jina-embeddings-v2-base-de": 768,
    "Xenova/jina-embeddings-v2-small-en": 512,
    "Xenova/all-MiniLM-L6-v2": 384,
    "Xenova/all-MiniLM-L12-v2": 384,
    "Xenova/all-mpnet-base-v2": 768,
    "Xenova/bge-base-en-v1.5": 768,
    "Xenova/bge-small-en-v1.5": 384,
    "Xenova/gte-small": 384,
    "Xenova/GIST-small-Embedding-v0": 384,
    "Xenova/text-embedding-ada-002": 1536,
    "text-embedding-3-small": 1536,
    "text-embedding-3-large": 3072,
    "text-embedding-ada-002": 1536,
    "embed-english-v3.0": 1024,
    "embed-multilingual-v3.0": 1024,
    "embed-english-light-v3.0": 384,
    "embed-multilingual-light-v3.0": 384,
    "text-embedding-004": 768,
    "text-multilingual-embedding-002": 768,
    "voyage-3": 1024,
    "voyage-3-lite": 512,
    "voyage-code-3": 1024,
  };
  return dimensionMap[model] ?? 768;
}

function buildConfig(fileConfig: OpenCodeMemConfig) {
  return {
    storagePath: expandPath(fileConfig.storagePath ?? DEFAULTS.storagePath),
    userEmailOverride: fileConfig.userEmailOverride,
    userNameOverride: fileConfig.userNameOverride,
    embeddingModel: fileConfig.embeddingModel ?? DEFAULTS.embeddingModel,
    embeddingDimensions:
      fileConfig.embeddingDimensions ??
      getEmbeddingDimensions(fileConfig.embeddingModel ?? DEFAULTS.embeddingModel),
    embeddingApiUrl: fileConfig.embeddingApiUrl,
    embeddingApiKey: fileConfig.embeddingApiUrl
      ? resolveSecretValue(fileConfig.embeddingApiKey ?? process.env["OPENAI_API_KEY"])
      : undefined,
    similarityThreshold: fileConfig.similarityThreshold ?? DEFAULTS.similarityThreshold,
    maxMemories: fileConfig.maxMemories ?? DEFAULTS.maxMemories,
    maxProfileItems: fileConfig.maxProfileItems ?? DEFAULTS.maxProfileItems,
    injectProfile: fileConfig.injectProfile ?? DEFAULTS.injectProfile,
    containerTagPrefix: fileConfig.containerTagPrefix ?? DEFAULTS.containerTagPrefix,
    autoCaptureEnabled: fileConfig.autoCaptureEnabled ?? DEFAULTS.autoCaptureEnabled,
    autoCaptureMaxIterations:
      fileConfig.autoCaptureMaxIterations ?? DEFAULTS.autoCaptureMaxIterations,
    autoCaptureIterationTimeout:
      fileConfig.autoCaptureIterationTimeout ?? DEFAULTS.autoCaptureIterationTimeout,
    autoCaptureLanguage: fileConfig.autoCaptureLanguage,
    memoryProvider: (fileConfig.memoryProvider ?? "openai-chat") as
      | "openai-chat"
      | "openai-responses"
      | "anthropic",
    memoryModel: fileConfig.memoryModel,
    memoryApiUrl: fileConfig.memoryApiUrl,
    memoryApiKey: resolveSecretValue(fileConfig.memoryApiKey),
    memoryTemperature: fileConfig.memoryTemperature,
    memoryExtraParams: fileConfig.memoryExtraParams,
    opencodeProvider: fileConfig.opencodeProvider,
    opencodeModel: fileConfig.opencodeModel,
    vectorBackend: (fileConfig.vectorBackend ?? "usearch-first") as
      | "usearch-first"
      | "usearch"
      | "exact-scan",
    aiSessionRetentionDays:
      fileConfig.aiSessionRetentionDays ?? DEFAULTS.aiSessionRetentionDays,
    webServerEnabled: fileConfig.webServerEnabled ?? DEFAULTS.webServerEnabled,
    webServerPort: fileConfig.webServerPort ?? DEFAULTS.webServerPort,
    webServerHost: fileConfig.webServerHost ?? DEFAULTS.webServerHost,
    maxVectorsPerShard: fileConfig.maxVectorsPerShard ?? DEFAULTS.maxVectorsPerShard,
    autoCleanupEnabled: fileConfig.autoCleanupEnabled ?? DEFAULTS.autoCleanupEnabled,
    autoCleanupRetentionDays:
      fileConfig.autoCleanupRetentionDays ?? DEFAULTS.autoCleanupRetentionDays,
    deduplicationEnabled: fileConfig.deduplicationEnabled ?? DEFAULTS.deduplicationEnabled,
    deduplicationSimilarityThreshold:
      fileConfig.deduplicationSimilarityThreshold ??
      DEFAULTS.deduplicationSimilarityThreshold,
    userProfileAnalysisInterval:
      fileConfig.userProfileAnalysisInterval ?? DEFAULTS.userProfileAnalysisInterval,
    userProfileMaxPreferences:
      fileConfig.userProfileMaxPreferences ?? DEFAULTS.userProfileMaxPreferences,
    userProfileMaxPatterns:
      fileConfig.userProfileMaxPatterns ?? DEFAULTS.userProfileMaxPatterns,
    userProfileMaxWorkflows:
      fileConfig.userProfileMaxWorkflows ?? DEFAULTS.userProfileMaxWorkflows,
    userProfileConfidenceDecayDays:
      fileConfig.userProfileConfidenceDecayDays ?? DEFAULTS.userProfileConfidenceDecayDays,
    userProfileChangelogRetentionCount:
      fileConfig.userProfileChangelogRetentionCount ??
      DEFAULTS.userProfileChangelogRetentionCount,
    showAutoCaptureToasts:
      fileConfig.showAutoCaptureToasts ?? DEFAULTS.showAutoCaptureToasts,
    showUserProfileToasts:
      fileConfig.showUserProfileToasts ?? DEFAULTS.showUserProfileToasts,
    showErrorToasts: fileConfig.showErrorToasts ?? DEFAULTS.showErrorToasts,
    memory: {
      defaultScope:
        fileConfig.memory?.defaultScope ?? DEFAULTS.memory.defaultScope,
    },
    compaction: {
      enabled: fileConfig.compaction?.enabled ?? DEFAULTS.compaction.enabled,
      memoryLimit:
        fileConfig.compaction?.memoryLimit ?? DEFAULTS.compaction.memoryLimit,
    },
    chatMessage: {
      enabled: fileConfig.chatMessage?.enabled ?? DEFAULTS.chatMessage.enabled,
      maxMemories:
        fileConfig.chatMessage?.maxMemories ?? DEFAULTS.chatMessage.maxMemories,
      excludeCurrentSession:
        fileConfig.chatMessage?.excludeCurrentSession ??
        DEFAULTS.chatMessage.excludeCurrentSession,
      maxAgeDays: fileConfig.chatMessage?.maxAgeDays,
      injectOn: (fileConfig.chatMessage?.injectOn ?? DEFAULTS.chatMessage.injectOn) as
        | "first"
        | "always",
    },
  };
}

let _globalFileConfig = loadConfigFromPaths(CONFIG_FILES);
export let CONFIG = buildConfig(_globalFileConfig);

export function initConfig(directory: string): void {
  const projectPaths = [
    join(directory, ".opencode", "opencode-mem.jsonc"),
    join(directory, ".opencode", "opencode-mem.json"),
  ];
  const globalConfig = loadConfigFromPaths(CONFIG_FILES);
  const projectConfig = loadConfigFromPaths(projectPaths);
  const merged: OpenCodeMemConfig = { ...globalConfig, ...projectConfig };
  _globalFileConfig = merged;
  CONFIG = buildConfig(merged);
}

export function isConfigured(): boolean {
  return (
    !!(
      CONFIG.memoryProvider &&
      CONFIG.memoryModel &&
      CONFIG.memoryApiUrl &&
      CONFIG.memoryApiKey
    ) || !!(CONFIG.opencodeProvider && CONFIG.opencodeModel)
  );
}
```

Config loading order (highest priority wins):
1. Project-level: `<project>/.opencode/opencode-mem.jsonc`
2. Global: `~/.config/opencode/opencode-mem.jsonc`

`isConfigured()` returns true if either a full `memoryProvider` config exists OR an `opencodeProvider + opencodeModel` pair is set.

---

### Step 4 — Utility services

These small services are imported by nearly everything else. Create them first.

**`src/services/logger.ts`:**

```typescript
export function log(message: string, data?: Record<string, unknown>): void {
  const timestamp = new Date().toISOString();
  if (data) {
    console.error(`[opencode-mem ${timestamp}] ${message}`, JSON.stringify(data));
  } else {
    console.error(`[opencode-mem ${timestamp}] ${message}`);
  }
}
```

**`src/services/jsonc.ts`** — strips `//` and `/* */` comments from JSONC:

```typescript
export function stripJsoncComments(jsonc: string): string {
  let result = "";
  let i = 0;
  let inString = false;
  let escaped = false;

  while (i < jsonc.length) {
    const char = jsonc[i]!;

    if (escaped) {
      result += char;
      escaped = false;
      i++;
      continue;
    }

    if (char === "\\" && inString) {
      result += char;
      escaped = true;
      i++;
      continue;
    }

    if (char === '"') {
      inString = !inString;
      result += char;
      i++;
      continue;
    }

    if (!inString) {
      // Line comment
      if (char === "/" && jsonc[i + 1] === "/") {
        while (i < jsonc.length && jsonc[i] !== "\n") i++;
        continue;
      }
      // Block comment
      if (char === "/" && jsonc[i + 1] === "*") {
        i += 2;
        while (i < jsonc.length && !(jsonc[i] === "*" && jsonc[i + 1] === "/")) i++;
        i += 2;
        continue;
      }
      // Trailing comma before } or ]
      if (char === ",") {
        let j = i + 1;
        while (j < jsonc.length && /\s/.test(jsonc[j]!)) j++;
        if (jsonc[j] === "}" || jsonc[j] === "]") {
          i++;
          continue;
        }
      }
    }

    result += char;
    i++;
  }

  return result;
}
```

**`src/services/secret-resolver.ts`** — resolves `$ENV_VAR` syntax in config values:

```typescript
export function resolveSecretValue(value: string | undefined): string | undefined {
  if (!value) return undefined;
  if (value.startsWith("$")) {
    const envKey = value.slice(1);
    return process.env[envKey];
  }
  return value;
}
```

**`src/services/tags.ts`** — generates deterministic container tags from project path and user info:

```typescript
import { createHash } from "node:crypto";
import { CONFIG } from "../config.js";

export interface Tags {
  project: { tag: string; hash: string };
  user: { tag: string; hash: string; userEmail: string | null; userName: string | null };
}

function hashString(input: string): string {
  return createHash("sha256").update(input).digest("hex").slice(0, 8);
}

export function getTags(directory: string): Tags {
  const userEmail = CONFIG.userEmailOverride || process.env["USER_EMAIL"] || null;
  const userName = CONFIG.userNameOverride || process.env["USER_NAME"] || null;
  const userIdentifier = userEmail || userName || "unknown";
  const userHash = hashString(userIdentifier);
  const projectHash = hashString(directory);

  return {
    project: {
      hash: projectHash,
      tag: `${CONFIG.containerTagPrefix}_project_${projectHash}`,
    },
    user: {
      hash: userHash,
      tag: `${CONFIG.containerTagPrefix}_user_${userHash}`,
      userEmail,
      userName,
    },
  };
}
```

**`src/services/language-detector.ts`** — maps BCP-47 codes to human names:

```typescript
export function getLanguageName(code: string): string {
  const names: Record<string, string> = {
    en: "English", es: "Spanish", fr: "French", de: "German",
    it: "Italian", pt: "Portuguese", ru: "Russian", zh: "Chinese",
    ja: "Japanese", ko: "Korean", ar: "Arabic", nl: "Dutch",
    pl: "Polish", tr: "Turkish", sv: "Swedish", da: "Danish",
  };
  return names[code] ?? code;
}
```

---

### Step 5 — Embedding service

**`src/services/embedding.ts`** — local HuggingFace transformers + optional API backend:

```typescript
import { CONFIG } from "../config.js";
import { log } from "./logger.js";
import { join } from "node:path";

const TIMEOUT_MS = 30000;
const GLOBAL_EMBEDDING_KEY = Symbol.for("opencode-mem.embedding.instance");
const MAX_CACHE_SIZE = 100;

let _transformers: {
  pipeline: (typeof import("@huggingface/transformers"))["pipeline"];
  env: (typeof import("@huggingface/transformers"))["env"];
} | null = null;

async function ensureTransformersLoaded(): Promise<typeof _transformers> {
  if (_transformers !== null) return _transformers;
  const mod = await import("@huggingface/transformers");
  mod.env.allowLocalModels = true;
  mod.env.allowRemoteModels = true;
  mod.env.cacheDir = join(CONFIG.storagePath, ".cache");
  _transformers = mod;
  return _transformers!;
}

function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  return Promise.race([
    promise,
    new Promise<T>((_, reject) =>
      setTimeout(() => reject(new Error(`Timeout after ${ms}ms`)), ms)
    ),
  ]);
}

export class EmbeddingService {
  private pipe: any = null;
  private initPromise: Promise<void> | null = null;
  public isWarmedUp: boolean = false;
  private cache: Map<string, Float32Array> = new Map();
  private cachedModelName: string | null = null;

  static getInstance(): EmbeddingService {
    if (!(globalThis as any)[GLOBAL_EMBEDDING_KEY]) {
      (globalThis as any)[GLOBAL_EMBEDDING_KEY] = new EmbeddingService();
    }
    return (globalThis as any)[GLOBAL_EMBEDDING_KEY];
  }

  async warmup(progressCallback?: (progress: any) => void): Promise<void> {
    if (this.isWarmedUp) return;
    if (this.initPromise) return this.initPromise;
    this.initPromise = this.initializeModel(progressCallback);
    return this.initPromise;
  }

  private async initializeModel(
    progressCallback?: (progress: any) => void
  ): Promise<void> {
    try {
      if (CONFIG.embeddingApiUrl && CONFIG.embeddingApiKey) {
        this.isWarmedUp = true;
        return;
      }
      const { pipeline } = await ensureTransformersLoaded();
      this.pipe = await pipeline("feature-extraction", CONFIG.embeddingModel, {
        progress_callback: progressCallback,
      });
      this.isWarmedUp = true;
    } catch (error) {
      this.initPromise = null;
      log("Failed to initialize embedding model", { error: String(error) });
      throw error;
    }
  }

  async embed(text: string): Promise<Float32Array> {
    if (this.cachedModelName !== CONFIG.embeddingModel) {
      this.clearCache();
      this.cachedModelName = CONFIG.embeddingModel;
    }
    const cached = this.cache.get(text);
    if (cached) return cached;

    if (!this.isWarmedUp && !this.initPromise) await this.warmup();
    if (this.initPromise) await this.initPromise;

    let result: Float32Array;

    if (CONFIG.embeddingApiUrl && CONFIG.embeddingApiKey) {
      const response = await fetch(`${CONFIG.embeddingApiUrl}/embeddings`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${CONFIG.embeddingApiKey}`,
        },
        body: JSON.stringify({ input: text, model: CONFIG.embeddingModel }),
      });
      if (!response.ok) {
        throw new Error(`API embedding failed: ${response.statusText}`);
      }
      const data: any = await response.json();
      result = new Float32Array(data.data[0].embedding);
    } else {
      const output = await this.pipe(text, { pooling: "mean", normalize: true });
      result = new Float32Array(output.data);
    }

    if (this.cache.size >= MAX_CACHE_SIZE) {
      const firstKey = this.cache.keys().next().value;
      if (firstKey !== undefined) this.cache.delete(firstKey);
    }
    this.cache.set(text, result);
    return result;
  }

  async embedWithTimeout(text: string): Promise<Float32Array> {
    return withTimeout(this.embed(text), TIMEOUT_MS);
  }

  clearCache(): void {
    this.cache.clear();
  }
}

export const embeddingService = EmbeddingService.getInstance();
```

Key design choices:
- Singleton via `Symbol.for()` — safe across module reloads in Bun
- LRU-style cache (evicts oldest at 100 entries)
- Model cached in `~/.opencode-mem/data/.cache/` (respects `storagePath`)
- API fallback: if `embeddingApiUrl` + `embeddingApiKey` set, skips local model entirely

---

### Step 6 — SQLite backend

The SQLite layer has 4 files. The shard system splits memories into shards by scope (`user` vs `project`) and hash, with `maxVectorsPerShard` (default 50,000) per file.

**`src/services/sqlite/types.ts`:**

```typescript
export interface MemoryRecord {
  id: string;
  container_tag: string;
  content: string;
  summary: string;
  vector: Buffer | null;
  metadata: string | null;
  created_at: number;
  updated_at: number;
  is_pinned: number;
}

export interface ShardInfo {
  id: string;
  dbPath: string;
  scope: "user" | "project";
  hash: string;
  vectorCount: number;
}
```

**`src/services/sqlite/connection-manager.ts`** — manages SQLite connections per shard file:

```typescript
import { Database } from "bun:sqlite";
import { existsSync, mkdirSync } from "node:fs";
import { dirname } from "node:path";
import { log } from "../logger.js";

const GLOBAL_CM_KEY = Symbol.for("opencode-mem.connection-manager");

class ConnectionManager {
  private connections: Map<string, Database> = new Map();

  getConnection(dbPath: string): Database {
    const existing = this.connections.get(dbPath);
    if (existing) return existing;

    const dir = dirname(dbPath);
    if (!existsSync(dir)) mkdirSync(dir, { recursive: true });

    const db = new Database(dbPath);
    db.exec("PRAGMA journal_mode = WAL");
    db.exec("PRAGMA synchronous = NORMAL");
    db.exec("PRAGMA foreign_keys = ON");
    db.exec("PRAGMA cache_size = -8000"); // 8MB cache
    this.connections.set(dbPath, db);
    return db;
  }

  closeAll(): void {
    for (const [, db] of this.connections) {
      try { db.close(); } catch {}
    }
    this.connections.clear();
  }

  checkpointAll(): void {
    for (const [, db] of this.connections) {
      try {
        db.exec("PRAGMA wal_checkpoint(TRUNCATE)");
      } catch (error) {
        log("WAL checkpoint failed", { error: String(error) });
      }
    }
  }
}

function getConnectionManager(): ConnectionManager {
  if (!(globalThis as any)[GLOBAL_CM_KEY]) {
    (globalThis as any)[GLOBAL_CM_KEY] = new ConnectionManager();
  }
  return (globalThis as any)[GLOBAL_CM_KEY];
}

export const connectionManager = getConnectionManager();
```

**`src/services/sqlite/shard-manager.ts`** — manages the shard registry (which DB files exist):

```typescript
import { Database } from "bun:sqlite";
import { existsSync, mkdirSync } from "node:fs";
import { join } from "node:path";
import { CONFIG } from "../../config.js";
import { connectionManager } from "./connection-manager.js";
import type { ShardInfo } from "./types.js";

const GLOBAL_SM_KEY = Symbol.for("opencode-mem.shard-manager");

class ShardManager {
  private registryDb: Database | null = null;
  private shards: Map<string, ShardInfo> = new Map();
  private initialized = false;

  private getRegistry(): Database {
    if (this.registryDb) return this.registryDb;
    const storageDir = CONFIG.storagePath;
    if (!existsSync(storageDir)) mkdirSync(storageDir, { recursive: true });
    const registryPath = join(storageDir, "registry.db");
    this.registryDb = new Database(registryPath);
    this.registryDb.exec(`
      CREATE TABLE IF NOT EXISTS shards (
        id TEXT PRIMARY KEY,
        db_path TEXT NOT NULL,
        scope TEXT NOT NULL,
        hash TEXT NOT NULL,
        vector_count INTEGER NOT NULL DEFAULT 0,
        created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
      );
    `);
    return this.registryDb;
  }

  private loadShards(): void {
    if (this.initialized) return;
    const registry = this.getRegistry();
    const rows = registry.prepare("SELECT * FROM shards").all() as any[];
    for (const row of rows) {
      this.shards.set(row.id, {
        id: row.id,
        dbPath: row.db_path,
        scope: row.scope,
        hash: row.hash,
        vectorCount: row.vector_count,
      });
    }
    this.initialized = true;
  }

  getOrCreateShard(scope: "user" | "project", hash: string): ShardInfo {
    this.loadShards();
    const existing = [...this.shards.values()].find(
      (s) =>
        s.scope === scope &&
        s.hash === hash &&
        s.vectorCount < CONFIG.maxVectorsPerShard
    );
    if (existing) return existing;

    const id = `${scope}_${hash}_${Date.now()}`;
    const dbPath = join(CONFIG.storagePath, `${id}.db`);
    const shard: ShardInfo = { id, dbPath, scope, hash, vectorCount: 0 };

    const registry = this.getRegistry();
    registry.prepare(
      "INSERT OR REPLACE INTO shards (id, db_path, scope, hash, vector_count) VALUES (?, ?, ?, ?, ?)"
    ).run(shard.id, shard.dbPath, shard.scope, shard.hash, shard.vectorCount);

    const db = connectionManager.getConnection(dbPath);
    db.exec(`
      CREATE TABLE IF NOT EXISTS memories (
        id TEXT PRIMARY KEY,
        container_tag TEXT NOT NULL,
        content TEXT NOT NULL,
        summary TEXT,
        vector BLOB,
        metadata TEXT,
        created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
        updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
        is_pinned INTEGER NOT NULL DEFAULT 0
      );
      CREATE INDEX IF NOT EXISTS idx_container_tag ON memories(container_tag);
      CREATE INDEX IF NOT EXISTS idx_created_at ON memories(created_at);
    `);

    this.shards.set(id, shard);
    return shard;
  }

  getAllShards(scope: "user" | "project", hash: string): ShardInfo[] {
    this.loadShards();
    if (hash === "") return [...this.shards.values()].filter((s) => s.scope === scope);
    return [...this.shards.values()].filter((s) => s.scope === scope && s.hash === hash);
  }

  incrementVectorCount(shardId: string): void {
    const shard = this.shards.get(shardId);
    if (!shard) return;
    shard.vectorCount++;
    const registry = this.getRegistry();
    registry
      .prepare("UPDATE shards SET vector_count = ? WHERE id = ?")
      .run(shard.vectorCount, shardId);
  }

  decrementVectorCount(shardId: string): void {
    const shard = this.shards.get(shardId);
    if (!shard || shard.vectorCount <= 0) return;
    shard.vectorCount--;
    const registry = this.getRegistry();
    registry
      .prepare("UPDATE shards SET vector_count = ? WHERE id = ?")
      .run(shard.vectorCount, shardId);
  }
}

function getShardManager(): ShardManager {
  if (!(globalThis as any)[GLOBAL_SM_KEY]) {
    (globalThis as any)[GLOBAL_SM_KEY] = new ShardManager();
  }
  return (globalThis as any)[GLOBAL_SM_KEY];
}

export const shardManager = getShardManager();
```

**`src/services/sqlite/vector-search.ts`** — USearch + ExactScan with automatic fallback:

```typescript
import type { Database } from "bun:sqlite";
import { randomUUID } from "node:crypto";
import { CONFIG } from "../../config.js";
import { log } from "../logger.js";
import type { MemoryRecord, ShardInfo } from "./types.js";

interface SearchResult {
  similarity: number;
  memory: string;
  id: string;
  sessionID?: string;
}

class VectorSearch {
  private useUSearch: boolean | null = null;

  private async tryInitUSearch(): Promise<boolean> {
    if (this.useUSearch !== null) return this.useUSearch;
    if (CONFIG.vectorBackend === "exact-scan") {
      this.useUSearch = false;
      return false;
    }
    try {
      const { Index } = await import("usearch");
      // Test construction
      new Index({ metric: "cos", connectivity: 16, dimensions: CONFIG.embeddingDimensions });
      this.useUSearch = true;
      return true;
    } catch {
      log("USearch unavailable — falling back to ExactScan");
      this.useUSearch = false;
      return false;
    }
  }

  async addVector(
    db: Database,
    shard: ShardInfo,
    content: string,
    summary: string,
    vector: Float32Array,
    containerTag: string,
    metadata: string | null
  ): Promise<string> {
    const id = randomUUID();
    const vectorBuffer = Buffer.from(vector.buffer);
    const now = Date.now();
    db.prepare(
      `INSERT INTO memories (id, container_tag, content, summary, vector, metadata, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(id, containerTag, content, summary, vectorBuffer, metadata, now, now);
    return id;
  }

  async searchAcrossShards(
    shards: ShardInfo[],
    queryVector: Float32Array,
    containerTag: string,
    limit: number,
    similarityThreshold: number,
    _query: string
  ): Promise<SearchResult[]> {
    const canUseUSearch = await this.tryInitUSearch();
    const allResults: SearchResult[] = [];

    for (const shard of shards) {
      try {
        const { connectionManager } = await import("./connection-manager.js");
        const db = connectionManager.getConnection(shard.dbPath);
        const query = containerTag
          ? db.prepare("SELECT * FROM memories WHERE container_tag = ?").all(containerTag) as MemoryRecord[]
          : db.prepare("SELECT * FROM memories").all() as MemoryRecord[];

        for (const row of query) {
          if (!row.vector) continue;
          const rowVector = new Float32Array(new Uint8Array(row.vector).buffer);
          const similarity = this.cosineSimilarity(queryVector, rowVector);
          if (similarity >= similarityThreshold) {
            let metadata: any = {};
            try { metadata = JSON.parse(row.metadata ?? "{}"); } catch {}
            allResults.push({
              similarity,
              memory: row.content,
              id: row.id,
              sessionID: metadata.sessionID,
            });
          }
        }
      } catch (error) {
        log("Search shard error", { shardId: shard.id, error: String(error) });
      }
    }

    return allResults.sort((a, b) => b.similarity - a.similarity).slice(0, limit);
  }

  async deleteVector(db: Database, id: string, _shard: ShardInfo): Promise<void> {
    db.prepare("DELETE FROM memories WHERE id = ?").run(id);
  }

  getAllMemories(db: Database): MemoryRecord[] {
    return db.prepare("SELECT * FROM memories").all() as MemoryRecord[];
  }

  private cosineSimilarity(a: Float32Array, b: Float32Array): number {
    if (a.length !== b.length) return 0;
    let dot = 0, normA = 0, normB = 0;
    for (let i = 0; i < a.length; i++) {
      dot += (a[i] ?? 0) * (b[i] ?? 0);
      normA += (a[i] ?? 0) ** 2;
      normB += (b[i] ?? 0) ** 2;
    }
    if (normA === 0 || normB === 0) return 0;
    return dot / (Math.sqrt(normA) * Math.sqrt(normB));
  }
}

export const vectorSearch = new VectorSearch();
```

---

### Step 7 — Memory client

**`src/services/client.ts`** — the core CRUD interface used by all agents and handlers:

```typescript
import { embeddingService } from "./embedding.js";
import { shardManager } from "./sqlite/shard-manager.js";
import { vectorSearch } from "./sqlite/vector-search.js";
import { connectionManager } from "./sqlite/connection-manager.js";
import { CONFIG } from "../config.js";
import { log } from "./logger.js";
import type { MemoryType } from "../types/index.js";
import type { MemoryRecord } from "./sqlite/types.js";

export type MemoryScope = "project" | "all-projects";

function safeToISOString(timestamp: any): string {
  try {
    if (timestamp === null || timestamp === undefined) return new Date().toISOString();
    const numValue = Number(timestamp);
    if (isNaN(numValue) || numValue < 0) return new Date().toISOString();
    return new Date(numValue).toISOString();
  } catch {
    return new Date().toISOString();
  }
}

function extractScopeFromContainerTag(containerTag: string): {
  scope: "user" | "project";
  hash: string;
} {
  const parts = containerTag.split("_");
  if (parts.length >= 3) {
    return { scope: parts[1] as "user" | "project", hash: parts.slice(2).join("_") };
  }
  return { scope: "user", hash: containerTag };
}

function resolveScopeValue(
  scope: MemoryScope,
  containerTag: string
): { scope: "user" | "project"; hash: string } {
  if (scope === "all-projects") return { scope: "project", hash: "" };
  return extractScopeFromContainerTag(containerTag);
}

const GLOBAL_CLIENT_KEY = Symbol.for("opencode-mem.client");

export class LocalMemoryClient {
  private initPromise: Promise<void> | null = null;
  private isInitialized: boolean = false;

  private async initialize(): Promise<void> {
    if (this.isInitialized) return;
    if (this.initPromise) return this.initPromise;
    this.initPromise = (async () => {
      try {
        this.isInitialized = true;
      } catch (error) {
        this.initPromise = null;
        log("SQLite initialization failed", { error: String(error) });
        throw error;
      }
    })();
    return this.initPromise;
  }

  async warmup(progressCallback?: (progress: any) => void): Promise<void> {
    await this.initialize();
    await embeddingService.warmup(progressCallback);
  }

  async isReady(): Promise<boolean> {
    return this.isInitialized && embeddingService.isWarmedUp;
  }

  getStatus() {
    return {
      dbConnected: this.isInitialized,
      modelLoaded: embeddingService.isWarmedUp,
      ready: this.isInitialized && embeddingService.isWarmedUp,
    };
  }

  close(): void {
    connectionManager.closeAll();
  }

  async searchMemories(
    query: string,
    containerTag: string,
    scope: MemoryScope = "project"
  ) {
    try {
      await this.initialize();
      const queryVector = await embeddingService.embedWithTimeout(query);
      const resolved = resolveScopeValue(scope, containerTag);
      const shards = shardManager.getAllShards(resolved.scope, resolved.hash);
      if (shards.length === 0) return { success: true as const, results: [], total: 0, timing: 0 };
      const results = await vectorSearch.searchAcrossShards(
        shards,
        queryVector,
        scope === "all-projects" ? "" : containerTag,
        CONFIG.maxMemories,
        CONFIG.similarityThreshold,
        query
      );
      return { success: true as const, results, total: results.length, timing: 0 };
    } catch (error) {
      return { success: false as const, error: String(error) };
    }
  }

  async addMemory(
    content: string,
    containerTag: string,
    metadata: Record<string, unknown> = {},
    memoryType?: MemoryType
  ) {
    try {
      await this.initialize();
      const vector = await embeddingService.embedWithTimeout(content);
      const resolved = extractScopeFromContainerTag(containerTag);
      const shard = shardManager.getOrCreateShard(resolved.scope, resolved.hash);
      const db = connectionManager.getConnection(shard.dbPath);
      const metaStr = JSON.stringify({ ...metadata, type: memoryType });
      const id = await vectorSearch.addVector(
        db, shard, content, content, vector, containerTag, metaStr
      );
      shardManager.incrementVectorCount(shard.id);
      return { success: true as const, id };
    } catch (error) {
      return { success: false as const, error: String(error) };
    }
  }

  async deleteMemory(id: string, containerTag: string) {
    try {
      await this.initialize();
      const resolved = extractScopeFromContainerTag(containerTag);
      const shards = shardManager.getAllShards(resolved.scope, resolved.hash);
      for (const shard of shards) {
        const db = connectionManager.getConnection(shard.dbPath);
        const row = db.prepare("SELECT 1 FROM memories WHERE id = ?").get(id);
        if (row) {
          await vectorSearch.deleteVector(db, id, shard);
          shardManager.decrementVectorCount(shard.id);
          return { success: true as const };
        }
      }
      return { success: false as const, error: "Memory not found" };
    } catch (error) {
      return { success: false as const, error: String(error) };
    }
  }

  async listMemories(containerTag: string, limit: number = CONFIG.maxMemories) {
    try {
      await this.initialize();
      const resolved = extractScopeFromContainerTag(containerTag);
      const shards = shardManager.getAllShards(resolved.scope, resolved.hash);
      const all: any[] = [];
      for (const shard of shards) {
        const db = connectionManager.getConnection(shard.dbPath);
        const rows = db
          .prepare(
            "SELECT * FROM memories WHERE container_tag = ? ORDER BY updated_at DESC LIMIT ?"
          )
          .all(containerTag, limit) as MemoryRecord[];
        for (const r of rows) {
          let metadata: any = {};
          try { metadata = JSON.parse(r.metadata ?? "{}"); } catch {}
          all.push({
            id: r.id,
            summary: r.content,
            metadata,
            createdAt: safeToISOString(r.created_at),
            updatedAt: safeToISOString(r.updated_at),
            isPinned: r.is_pinned === 1,
          });
        }
      }
      all.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
      return { success: true as const, memories: all.slice(0, limit) };
    } catch (error) {
      return { success: false as const, memories: [], error: String(error) };
    }
  }
}

function getMemoryClient(): LocalMemoryClient {
  if (!(globalThis as any)[GLOBAL_CLIENT_KEY]) {
    (globalThis as any)[GLOBAL_CLIENT_KEY] = new LocalMemoryClient();
  }
  return (globalThis as any)[GLOBAL_CLIENT_KEY];
}

export const memoryClient = getMemoryClient();
```

---

### Step 8 — Deduplication and cleanup

**`src/services/deduplication-service.ts`** (exact from repo):

```typescript
import { shardManager } from "./sqlite/shard-manager.js";
import { vectorSearch } from "./sqlite/vector-search.js";
import { connectionManager } from "./sqlite/connection-manager.js";
import { CONFIG } from "../config.js";
import { log } from "./logger.js";

interface DuplicateGroup {
  representative: { id: string; content: string; containerTag: string; createdAt: number };
  duplicates: Array<{ id: string; content: string; similarity: number }>;
}

interface DeduplicationResult {
  exactDuplicatesDeleted: number;
  nearDuplicateGroups: DuplicateGroup[];
}

export class DeduplicationService {
  private isRunning = false;

  async detectAndRemoveDuplicates(): Promise<DeduplicationResult> {
    if (this.isRunning) throw new Error("Deduplication already running");
    if (!CONFIG.deduplicationEnabled) throw new Error("Deduplication is disabled in config");
    this.isRunning = true;
    try {
      const allShards = [
        ...shardManager.getAllShards("user", ""),
        ...shardManager.getAllShards("project", ""),
      ];
      let exactDeleted = 0;
      const nearDuplicateGroups: DuplicateGroup[] = [];

      for (const shard of allShards) {
        const db = connectionManager.getConnection(shard.dbPath);
        const memories = vectorSearch.getAllMemories(db);
        const contentMap = new Map<string, any[]>();

        for (const memory of memories) {
          const key = `${memory.container_tag}:${memory.content}`;
          if (!contentMap.has(key)) contentMap.set(key, []);
          contentMap.get(key)!.push(memory);
        }

        // Remove exact duplicates — keep newest
        for (const [, duplicates] of contentMap) {
          if (duplicates.length > 1) {
            duplicates.sort((a, b) => Number(b.created_at) - Number(a.created_at));
            for (const dup of duplicates.slice(1)) {
              try {
                await vectorSearch.deleteVector(db, dup.id, shard);
                shardManager.decrementVectorCount(shard.id);
                exactDeleted++;
              } catch (error) {
                log("Deduplication: delete error", { memoryId: dup.id, error: String(error) });
              }
            }
          }
        }

        // Detect near-duplicates via cosine similarity
        const uniqueMemories = Array.from(contentMap.values()).map((arr) => arr[0]);
        const processedIds = new Set<string>();

        for (let i = 0; i < uniqueMemories.length; i++) {
          const mem1 = uniqueMemories[i];
          if (!mem1?.vector || processedIds.has(mem1.id)) continue;
          const vector1 = new Float32Array(new Uint8Array(mem1.vector).buffer);
          const group: DuplicateGroup = {
            representative: {
              id: mem1.id, content: mem1.content,
              containerTag: mem1.container_tag, createdAt: mem1.created_at,
            },
            duplicates: [],
          };

          for (let j = i + 1; j < uniqueMemories.length; j++) {
            const mem2 = uniqueMemories[j];
            if (!mem2?.vector || processedIds.has(mem2.id)) continue;
            if (mem1.container_tag !== mem2.container_tag) continue;
            const vector2 = new Float32Array(new Uint8Array(mem2.vector).buffer);
            const similarity = this.cosineSimilarity(vector1, vector2);
            if (
              similarity >= CONFIG.deduplicationSimilarityThreshold &&
              similarity < 1.0
            ) {
              group.duplicates.push({ id: mem2.id, content: mem2.content, similarity });
              processedIds.add(mem2.id);
            }
          }

          if (group.duplicates.length > 0) nearDuplicateGroups.push(group);
        }
      }

      return { exactDuplicatesDeleted: exactDeleted, nearDuplicateGroups };
    } finally {
      this.isRunning = false;
    }
  }

  private cosineSimilarity(a: Float32Array, b: Float32Array): number {
    if (a.length !== b.length) return 0;
    let dot = 0, normA = 0, normB = 0;
    for (let i = 0; i < a.length; i++) {
      dot += (a[i] ?? 0) * (b[i] ?? 0);
      normA += (a[i] ?? 0) ** 2;
      normB += (b[i] ?? 0) ** 2;
    }
    if (normA === 0 || normB === 0) return 0;
    return dot / (Math.sqrt(normA) * Math.sqrt(normB));
  }

  getStatus() {
    return {
      enabled: CONFIG.deduplicationEnabled,
      similarityThreshold: CONFIG.deduplicationSimilarityThreshold,
      isRunning: this.isRunning,
    };
  }
}

export const deduplicationService = new DeduplicationService();
```

**`src/services/cleanup-service.ts`** (exact from repo):

```typescript
import { shardManager } from "./sqlite/shard-manager.js";
import { vectorSearch } from "./sqlite/vector-search.js";
import { connectionManager } from "./sqlite/connection-manager.js";
import { CONFIG } from "../config.js";
import { log } from "./logger.js";
import { userPromptManager } from "./user-prompt/user-prompt-manager.js";

interface CleanupResult {
  deletedCount: number;
  userCount: number;
  projectCount: number;
  promptsDeleted: number;
  linkedMemoriesDeleted: number;
  pinnedMemoriesSkipped: number;
}

export class CleanupService {
  private lastCleanupTime = 0;
  private isRunning = false;

  async shouldRunCleanup(): Promise<boolean> {
    if (!CONFIG.autoCleanupEnabled || this.isRunning) return false;
    return Date.now() - this.lastCleanupTime >= 24 * 60 * 60 * 1000;
  }

  async runCleanup(): Promise<CleanupResult> {
    if (this.isRunning) throw new Error("Cleanup already running");
    this.isRunning = true;
    this.lastCleanupTime = Date.now();
    try {
      const cutoffTime = Date.now() - CONFIG.autoCleanupRetentionDays * 86400000;
      const allShards = [
        ...shardManager.getAllShards("user", ""),
        ...shardManager.getAllShards("project", ""),
      ];

      const pinnedMemoryIds = new Set<string>();
      for (const shard of allShards) {
        const db = connectionManager.getConnection(shard.dbPath);
        const pinned = db.prepare("SELECT id FROM memories WHERE is_pinned = 1").all() as any[];
        pinned.forEach((row) => pinnedMemoryIds.add(row.id));
      }

      const promptCleanupResult = userPromptManager.deleteOldPrompts(cutoffTime);
      const protectedIds = new Set<string>([
        ...pinnedMemoryIds,
        ...promptCleanupResult.linkedMemoryIds,
      ]);

      let totalDeleted = 0, userDeleted = 0, projectDeleted = 0, pinnedSkipped = 0;

      for (const shard of allShards) {
        const db = connectionManager.getConnection(shard.dbPath);
        const old = db
          .prepare("SELECT id, container_tag, is_pinned FROM memories WHERE updated_at < ?")
          .all(cutoffTime) as any[];

        for (const memory of old) {
          if (memory.is_pinned === 1) { pinnedSkipped++; continue; }
          if (protectedIds.has(memory.id)) continue;
          try {
            await vectorSearch.deleteVector(db, memory.id, shard);
            shardManager.decrementVectorCount(shard.id);
            totalDeleted++;
            if (memory.container_tag?.includes("_user_")) userDeleted++;
            else if (memory.container_tag?.includes("_project_")) projectDeleted++;
          } catch (error) {
            log("Cleanup: delete error", { memoryId: memory.id, error: String(error) });
          }
        }
      }

      return {
        deletedCount: totalDeleted, userCount: userDeleted, projectCount: projectDeleted,
        promptsDeleted: promptCleanupResult.deleted - promptCleanupResult.linkedMemoryIds.size,
        linkedMemoriesDeleted: 0, pinnedMemoriesSkipped: pinnedSkipped,
      };
    } finally {
      this.isRunning = false;
    }
  }

  getStatus() {
    return {
      enabled: CONFIG.autoCleanupEnabled,
      retentionDays: CONFIG.autoCleanupRetentionDays,
      lastCleanupTime: this.lastCleanupTime,
      isRunning: this.isRunning,
    };
  }
}

export const cleanupService = new CleanupService();
```

---

### Step 9 — Context and privacy services

**`src/services/context.ts`** — formats memories as a synthetic prompt prefix:

```typescript
export function formatContextForPrompt(
  userId: string | null,
  searchResult: { results: Array<{ similarity: number; memory: string }>; total: number; timing: number }
): string | null {
  if (!searchResult.results || searchResult.results.length === 0) return null;

  const lines = searchResult.results.map((r) => `- ${r.memory}`);
  return `<memory>\n## Relevant Project Memories\n${lines.join("\n")}\n</memory>`;
}
```

**`src/services/privacy.ts`** — strips private content markers:

```typescript
const PRIVATE_MARKERS = [
  /\bpassword\s*[:=]\s*\S+/gi,
  /\bsecret\s*[:=]\s*\S+/gi,
  /\bapi[_-]?key\s*[:=]\s*\S+/gi,
  /\btoken\s*[:=]\s*\S+/gi,
  /\bprivate[_-]?key\s*[:=]\s*\S+/gi,
  /-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----[\s\S]*?-----END\s+(?:RSA\s+)?PRIVATE\s+KEY-----/g,
];

export function stripPrivateContent(text: string): string {
  let result = text;
  for (const pattern of PRIVATE_MARKERS) {
    result = result.replace(pattern, "[REDACTED]");
  }
  return result;
}

export function isFullyPrivate(text: string): boolean {
  return text.trim().length === 0 || text.replace(/\[REDACTED\]/g, "").trim().length === 0;
}
```

---

### Step 10 — Plugin main (index.ts)

**`src/index.ts`** — the Plugin function that implements `chat.message` and `session.completed` hooks (exact from repo):

```typescript
import type { Plugin, PluginInput } from "@opencode-ai/plugin";
import type { Part } from "@opencode-ai/sdk";

import { memoryClient } from "./services/client.js";
import { formatContextForPrompt } from "./services/context.js";
import { getTags } from "./services/tags.js";
import { stripPrivateContent, isFullyPrivate } from "./services/privacy.js";
import { performAutoCapture } from "./services/auto-capture.js";
import { performUserProfileLearning } from "./services/user-memory-learning.js";
import { userPromptManager } from "./services/user-prompt/user-prompt-manager.js";
import { startWebServer, type WebServer } from "./services/web-server.js";

import { isConfigured, CONFIG, initConfig } from "./config.js";
import { log } from "./services/logger.js";
import type { MemoryScope } from "./services/client.js";

export const OpenCodeMemPlugin: Plugin = async (ctx: PluginInput) => {
  const { directory } = ctx;
  initConfig(directory);
  const tags = getTags(directory);
  let webServer: WebServer | null = null;

  // One-time warmup per process
  const GLOBAL_PLUGIN_WARMUP_KEY = Symbol.for("opencode-mem.plugin.warmedup");
  if (!(globalThis as any)[GLOBAL_PLUGIN_WARMUP_KEY] && isConfigured()) {
    try {
      await memoryClient.warmup();
      (globalThis as any)[GLOBAL_PLUGIN_WARMUP_KEY] = true;
    } catch (error) {
      log("Plugin warmup failed", { error: String(error) });
    }
  }

  // Wire opencode state path and provider list (fire-and-forget)
  (async () => {
    try {
      const { setStatePath, setConnectedProviders } =
        await import("./services/ai/opencode-provider.js");
      const pathResult = await ctx.client.path.get();
      if (pathResult.data?.state) setStatePath(pathResult.data.state);
      const providerResult = await ctx.client.provider.list();
      if (providerResult.data?.connected) setConnectedProviders(providerResult.data.connected);
    } catch (error) {
      log("Failed to initialize opencode provider state", { error: String(error) });
    }
  })();

  // Start web server at :4747 (fire-and-forget, shows toast on success)
  if (CONFIG.webServerEnabled) {
    startWebServer({
      port: CONFIG.webServerPort,
      host: CONFIG.webServerHost,
      enabled: CONFIG.webServerEnabled,
    })
      .then((server) => {
        webServer = server;
        const url = webServer.getUrl();

        webServer.setOnTakeoverCallback(async () => {
          if (ctx.client?.tui) {
            ctx.client.tui
              .showToast({
                body: {
                  title: "Memory Explorer",
                  message: "Took over web server ownership",
                  variant: "success",
                  duration: 3000,
                },
              })
              .catch(() => {});
          }
        });

        if (webServer.isServerOwner() && ctx.client?.tui) {
          ctx.client.tui
            .showToast({
              body: {
                title: "Memory Explorer",
                message: `Web UI started at ${url}`,
                variant: "success",
                duration: 5000,
              },
            })
            .catch(() => {});
        }
      })
      .catch((error) => {
        log("Failed to start web server", { error: String(error) });
      });
  }

  const shutdownHandler = async () => {
    try {
      if (webServer) await webServer.stop();
      memoryClient.close();
      process.exit(0);
    } catch {
      process.exit(1);
    }
  };
  process.on("SIGINT", shutdownHandler);
  process.on("SIGTERM", shutdownHandler);

  return {
    "chat.message": async (input, output) => {
      if (!isConfigured() || !CONFIG.chatMessage.enabled) return;
      try {
        const textParts = output.parts.filter(
          (p): p is Part & { type: "text"; text: string } => p.type === "text"
        );
        if (textParts.length === 0) return;
        const userMessage = textParts.map((p) => p.text).join("\n");
        if (!userMessage.trim()) return;

        userPromptManager.savePrompt(input.sessionID, output.message.id, directory, userMessage);

        const messagesResponse = await ctx.client.session.messages({ path: { id: input.sessionID } });
        const messages = messagesResponse.data || [];
        const hasNonSyntheticUserMessages = messages.some(
          (m) =>
            m.info.role === "user" &&
            !m.parts.every((p) => p.type !== "text" || p.synthetic === true)
        );
        const lastMessage = messages.length > 0 ? messages[messages.length - 1] : null;
        const isAfterCompaction = lastMessage?.info?.summary === true;
        const shouldInject =
          CONFIG.chatMessage.injectOn === "always" ||
          !hasNonSyntheticUserMessages ||
          (isAfterCompaction &&
            messages.filter(
              (m) =>
                m.info.role === "user" &&
                !m.parts.every((p) => p.type !== "text" || p.synthetic === true)
            ).length === 1);

        if (!shouldInject) return;

        const listResult = await memoryClient.listMemories(
          tags.project.tag,
          CONFIG.chatMessage.maxMemories
        );
        let memories = listResult.success ? listResult.memories : [];

        if (CONFIG.chatMessage.excludeCurrentSession)
          memories = memories.filter((m: any) => m.metadata?.sessionID !== input.sessionID);
        if (CONFIG.chatMessage.maxAgeDays) {
          const cutoff = Date.now() - CONFIG.chatMessage.maxAgeDays * 86400000;
          memories = memories.filter((m: any) => new Date(m.createdAt).getTime() > cutoff);
        }
        if (memories.length === 0) return;

        const projectMemories = {
          results: memories.map((m: any) => ({ similarity: 1.0, memory: m.summary })),
          total: memories.length,
          timing: 0,
        };
        const userId = tags.user.userEmail || null;
        const memoryContext = formatContextForPrompt(userId, projectMemories);
        if (memoryContext) {
          const contextPart: Part = {
            id: `prt-memory-context-${Date.now()}`,
            sessionID: input.sessionID,
            messageID: output.message.id,
            type: "text",
            text: memoryContext,
            synthetic: true,
          } as any;
          output.parts.unshift(contextPart);
        }
      } catch (error) {
        log("chat.message: ERROR", { error: String(error) });
        if (ctx.client?.tui && CONFIG.showErrorToasts) {
          await ctx.client.tui
            .showToast({
              body: {
                title: "Memory System Error",
                message: String(error),
                variant: "error",
                duration: 5000,
              },
            })
            .catch(() => {});
        }
      }
    },

    "session.completed": async (input) => {
      if (!isConfigured() || !CONFIG.sessionCompleted?.enabled) return;
      try {
        const sessionID = input.sessionID;
        if (!sessionID) return;
        const messagesResponse = await ctx.client.session.messages({ path: { id: sessionID } });
        const messages = messagesResponse.data || [];
        if (messages.length === 0) return;

        await performAutoCapture(ctx, sessionID, directory);

        if (webServer?.isServerOwner()) {
          await performUserProfileLearning(ctx, directory);
          const { cleanupService } = await import("./services/cleanup-service.js");
          if (await cleanupService.shouldRunCleanup()) await cleanupService.runCleanup();
          const { connectionManager } = await import("./services/sqlite/connection-manager.js");
          connectionManager.checkpointAll();
        }
      } catch (error) {
        log("session.completed: ERROR", { error: String(error) });
      }
    },
  };
};
```

Key architectural points:
- `chat.message` hook: injects relevant memories as a **synthetic `<memory>` part** prepended to the message — the agent sees past context automatically
- `session.completed` hook: triggers auto-capture + profile learning + cleanup (only on web server owner to avoid duplicate runs when multiple sessions are active)
- `injectOn: "first"` means memories only injected on first message or after compaction
- Warmup via `Symbol.for()` prevents re-initializing the embedding model across plugin reloads

---

### Step 11 — Build and configuration

**`~/.config/opencode/opencode-mem.jsonc`** (recommended minimal config):

```jsonc
{
  // Required: choose one provider
  "opencodeProvider": "anthropic",
  "opencodeModel": "claude-haiku-4-5-20251001",

  // Or use direct API:
  // "memoryProvider": "openai-chat",
  // "memoryModel": "gpt-4o-mini",
  // "memoryApiUrl": "https://api.openai.com/v1",
  // "memoryApiKey": "$OPENAI_API_KEY",

  // Optional overrides
  "userEmailOverride": "you@example.com",
  "userNameOverride": "Your Name",
  "webServerEnabled": true,
  "webServerPort": 4747,
  "autoCaptureEnabled": true,
  "showAutoCaptureToasts": true,
  "memory": {
    "defaultScope": "project"
  }
}
```

**Build:**

```bash
bun run build
# Output: dist/plugin.js + dist/web/
```

**Register in OpenCode** (`~/.config/opencode/opencode.json`):

```jsonc
{
  "plugin": ["opencode-mem"]
}
```

For local development, use absolute path:

```jsonc
{
  "plugin": ["/absolute/path/to/opencode-mem/dist/plugin.js"]
}
```

---

### Step 12 — Test and validate

```bash
# Type check
bun run typecheck

# Build
bun run build

# Verify dist/plugin.js was generated
ls -la dist/plugin.js dist/web/
```

**Smoke test in OpenCode:**

1. Start OpenCode — plugin loads and warms up embedding model (~10s first run)
2. Web UI available at `http://127.0.0.1:4747`
3. Add a memory: type `memory({ mode: "add", content: "Project uses PostgreSQL" })`
4. Search: `memory({ mode: "search", query: "database decisions" })`
5. List: `memory({ mode: "list", limit: 5 })`
6. Profile: `memory({ mode: "profile" })` — shows learned user preferences
7. Cross-project: `memory({ mode: "search", query: "architecture", scope: "all-projects" })`

**Memory tool modes:**

| Mode | Usage |
|------|-------|
| `add` | Store a new memory |
| `search` | Vector similarity search |
| `list` | List recent memories with optional limit |
| `profile` | Show learned user profile |
| `forget` | Delete a memory by ID |
| `help` | Show usage |

## Validation checklist

- [ ] `bun run typecheck` passes with 0 errors
- [ ] `bun run build` produces `dist/plugin.js`
- [ ] `src/web/` is copied to `dist/web/` during build
- [ ] `~/.opencode-mem/data/` directory created on first run
- [ ] `~/.config/opencode/opencode-mem.jsonc` config file is readable
- [ ] `isConfigured()` returns true with valid `opencodeProvider` + `opencodeModel`
- [ ] Web UI accessible at `http://127.0.0.1:4747` after plugin loads
- [ ] First run downloads HuggingFace model to `~/.opencode-mem/data/.cache/`
- [ ] `memory.db` created as registry; shard `.db` files created in `storagePath`
- [ ] `chat.message` hook injects `<memory>` block as synthetic part on first message

## Notes

- **Bun is required** — the project uses `bun:sqlite` and Bun-specific APIs. Node.js is not supported.
- **Single process** — `Symbol.for()` singletons prevent duplicate initialization when plugin is hot-reloaded.
- **Shard architecture** — memories are split across multiple SQLite files (`user_<hash>_<ts>.db`, `project_<hash>_<ts>.db`) for performance. The `registry.db` tracks which shards exist.
- **USearch fallback** — `vectorBackend: "usearch-first"` (default) tries USearch first; if unavailable falls back to exact cosine scan. Set `"exact-scan"` to skip USearch entirely.
- **Local embedding model** downloads from HuggingFace on first run (~50MB for `Xenova/nomic-embed-text-v1`). Cached in `storagePath/.cache/`. Subsequent starts are instant.
- **Auto-capture** runs `performAutoCapture()` on `session.completed` — uses the configured AI provider to extract memories from the session transcript.
- **Web UI** is plain HTML/JS in `src/web/` — no build step needed for the UI files, they are copied to `dist/web/` by the build script.
- **Config merging** — project-level config (`.opencode/opencode-mem.jsonc`) overrides global config. Shallow merge (not deep).
- **Services not shown in this skill** (fetch them from the repo): `auto-capture.ts`, `user-memory-learning.ts`, `web-server.ts`, `web-server-worker.ts`, `api-handlers.ts`, `user-profile/`, `user-prompt/`, `ai/opencode-provider.ts` — these are large files (10–36KB each) that require reading the full repo source.
