---
name: nous-agent
description: "Use at session start or in NOUS projects — protocols, backups, git safety, OKF maintenance."
version: 1.1.0
author: Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [nous, protocols, safety, backups, okf, session]
    related_skills: [okf-knowledge, domain-modeling, openspec]
---

# NOUS Agent Protocols

Cross-project operational protocols injected by `nous sync hermes`. These apply to ALL projects, regardless of whether the project has an AGENTS.md.

## Session Start Protocol

Every session, in order:

1. Read `.agents/MEMORY.md` — active context, blockers, next action
2. Read `~/.hermes/OKF/index.md` — global catalog
3. Follow only links relevant to the current task (progressive disclosure — don't load everything)
4. Use Hermes `memory` tool for frequently-needed facts (auto-injected every turn)
5. Use `session_search("query")` for historical conversation context

## Backup Protocol

**Golden rule: never mutate external state without a backup.**

- Before editing any file outside `dev/sandbox/`, create a copy in `dev/backups/` with format: `YYYYMMDD_HHMMSS_filename.ext`
- Notify the backup creation to the user
- If failures are detected post-edit, analyze differences with the backup
- Propose rollback with a diff — never execute rollback without explicit user confirmation

## Git Safety

- **No silent mutations**: Forbidden to `git commit` or `git push` without explicit "YES" from user after showing `git diff`
- **External impact**: Actions on APIs, Cloud, or CI/CD require a detailed plan and prior human approval
- **Data protection**: Forbidden to delete databases or root directories without triple confirmation
- **Husky hooks**: Use `--no-verify` or `HUSKY=0` when hooks block legitimate operations

## OKF Maintenance Protocol

After meaningful work, update the knowledge base:

1. Update or create the relevant OKF concept in `~/.hermes/OKF/<project>/`
2. Ensure valid YAML frontmatter with non-empty `type`
3. Link from nearest `index.md`
4. Add milestones to `~/.hermes/OKF/log.md` (YYYY-MM-DD, newest first)
5. Use Hermes `memory` tool for compact, frequently-accessed facts
6. Keep `.agents/MEMORY.md` limited to active state and routes

### Persistence routing

| Knowledge | Destination |
|-----------|-------------|
| Current work, blockers, next action | `.agents/MEMORY.md` |
| Frequently-needed facts | Hermes `memory` tool |
| Architecture and boundaries | `~/.hermes/OKF/<project>/architecture.md` |
| Durable decisions and rationale | `~/.hermes/OKF/<project>/decisions/` |
| Verified runbooks | `~/.hermes/OKF/<project>/workflows/` |
| Diagnosed failures and fixes | `~/.hermes/OKF/<project>/troubleshooting/` |
| Cross-project conventions | `~/.hermes/OKF/_system/conventions.md` |
| Major milestones | `~/.hermes/OKF/log.md` |

### Forbidden

- NEVER duplicate durable knowledge across MEMORY.md and OKF concepts
- NEVER migrate unverified or low-value session chatter into OKF
- NEVER skip source verification before persisting a claim
- NEVER overwrite project memory during sync

## Hermes Tool Usage

- **`delegate_task`**: For parallel subtasks (research, code review, testing). Always include full context — subagents have no memory of your conversation.
- **`session_search`**: For recalling past decisions. Use FTS5 queries: `session_search("topic keyword")`.
- **`memory`**: For durable facts. Write declarative (what IS), not imperative (what TO DO). Consolidate when near the character limit.
- **`cronjob`**: For recurring tasks. Use `action='create'` with schedule, prompt, and deliver target.
- **`clarify`**: For ambiguous decisions. Prefer making a reasonable default choice for low-stakes decisions.
