---
name: skill-creator
description: Create new skills following the Antigravity format standard.
version: 1.0.0
author: Hermes Agent (adapted from NOUS)
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [skill-creator, skills, authoring, templates]
    related_skills: [project-map, okf-knowledge]
---

# Skill Creator

Creates new skills following the Antigravity format. For Hermes skills, use `skill_manage(action='create')` instead — this skill is for creating project-level skills in `.agents/skills/`.

## When to use

- Need a new skill for a specific task or workflow
- Team needs a shared skill for repeated patterns
- Integrating a new process or tool that deserves structured guidance

## Skill Structure

```
<skill-name>/
├── SKILL.md              ← REQUIRED: YAML frontmatter + markdown
├── scripts/              ← optional: helper scripts
├── examples/             ← optional: reference implementations
└── resources/            ← optional: templates and assets
```

## How to create a skill

### Step 1: Choose a name

- Lowercase, hyphens for spaces
- Descriptive of the capability
- Examples: `api-design`, `bug-triaging`, `code-review`

### Step 2: Create the folder

For project-specific skills: `.agents/skills/<skill-name>/`

For skills distributed with NOUS: `installs/skills/<skill-name>/` (NOUS format) or `installs/hermes-skills/<skill-name>/` (Hermes format).

### Step 3: Write SKILL.md

#### YAML Frontmatter (required)

NOUS format:
```yaml
---
name: my-skill
description: Third-person description with keywords. What and when.
---
```

Hermes format:
```yaml
---
name: my-skill
description: Short description under 60 chars.
version: 1.0.0
author: Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [tag1, tag2]
    related_skills: [other-skill]
---
```

#### Markdown Content

```markdown
# My Skill

Brief description.

## When to use
- Use when...

## How to use
Step-by-step instructions...

## Examples
Brief examples in action.
```

### Step 4: Add supporting files (optional)

- `scripts/` — focused helper scripts
- `examples/` — reference implementations
- `resources/` — templates, configs, assets

## Best practices

1. **Description in third person**: "Helps with..." not "I help with..."
2. **Keywords**: Include domain-specific terms
3. **Clear When to use**: List specific situations
4. **Step-by-step How to use**: Agents follow sequential instructions
5. **Examples**: Real examples beat abstract descriptions
6. **Minimal scripts**: Small helpers, not complex programs

## Validation checklist

- [ ] `SKILL.md` exists in the skill folder
- [ ] YAML frontmatter has `name` and `description`
- [ ] Description is third person
- [ ] Content explains when and how to use
- [ ] Folder name matches skill name
- [ ] All linked resources exist

## Notes

- Skills are discovered progressively: name + description first, full content when relevant
- For Hermes: use `skill_manage(action='create')` to create skills in `~/.hermes/skills/`
- For NOUS distribution: place in `installs/skills/` (NOUS format) or `installs/hermes-skills/` (Hermes format)
- Project-only skills: `.agents/skills/<name>/`
