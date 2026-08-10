---
name: wayfinder
description: Plan huge work as a map of decision tickets on the tracker.
version: 1.0.0
author: Hermes Agent (adapted from Matt Pocock/skills)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [planning, wayfinding, tickets, decisions, large-work, roadmap]
    related_skills: [grilling, domain-modeling, grill-with-docs, writing-plans]
---

# Wayfinder

A loose idea has arrived — too big for one agent session, and wrapped in fog: the way from here to the **destination** isn't visible yet. Wayfinding is about finding that way, not charging at the destination.

This skill charts the way as a **shared map** on the repo's issue tracker, then works its **decision tickets** — questions whose resolution is a decision, not slices of a build to execute — one at a time until the route is clear.

## Plan, don't do

Wayfinder is **planning** by default: each ticket resolves a decision, and the map is done when the way is clear — nothing left to decide before someone goes and does the thing. The pull to just do the work is usually the signal you've reached the edge of the map and it's time to hand off.

## The Map

The map is a single issue on the repo's issue tracker, labelled `wayfinder:map`. Its tickets are child issues of the map.

### Map body format

```markdown
## Destination
<what reaching the end of this map looks like — the spec, decision, or change>

## Notes
<domain; skills every session should consult; standing preferences>

## Decisions so far
- [closed ticket title](link) — one-line gist of the answer

## Not yet specified
<!-- in-scope fog you can't ticket yet; graduates as the frontier advances -->

## Out of scope
<!-- work ruled beyond the destination; closed, never graduates -->
```

## Ticket Types

Every ticket carries a `wayfinder:<type>` label:
- **Research** (AFK): Reading docs, APIs, or KBs. Resolved by a sub-agent via `delegate_task`.
- **Prototype** (HITL): Cheap, rough artifact to react to — outline, stub, or UI/logic code.
- **Grilling** (HITL): Conversation using the `grilling` and `domain-modeling` skills.
- **Task** (HITL or AFK): Manual work that must happen before a decision can be made.

## Process

### Chart the map (first session)
1. **Name the destination** — run a `grilling` and `domain-modeling` session
2. **Map the frontier** — grill breadth-first across the space
3. **Create the map** issue (label `wayfinder:map`)
4. **Create child tickets** for specifiable decisions, wire blocking edges
5. **Fire research sub-agents** in parallel for research tickets
6. **Stop** — charting is one session's work; it hand-resolves nothing

### Work through the map (subsequent sessions)
1. Load the map
2. Pick the next frontier ticket (open, unblocked, unclaimed)
3. Claim it (assign to yourself)
4. Resolve it using indicated skills
5. Record resolution, close ticket, update map's Decisions-so-far
6. Graduate fog into new tickets
7. Update map body

## Fog of war

The map is _deliberately_ incomplete: don't chart what you can't yet see. Beyond the live tickets lies the **fog of war** — decisions and investigations you can tell are coming but can't yet pin down. Resolving a ticket clears the fog ahead of it, graduating whatever's now specifiable into fresh tickets.

**Ticket when** the question is already sharp — even if it's blocked. **Not yet specified when** you can't yet phrase it that sharply.

## Hermes Integration

- Use `gh issue create` / `gh issue list` for GitHub-based maps
- Use `delegate_task` for parallel research tickets
- Use `writing-plans` when the map graduates into buildable work
- Use `grilling` for decision tickets labeled `wayfinder:grilling`
