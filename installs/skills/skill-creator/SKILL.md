---
name: skill-creator
description: Creates new skills following the Antigravity format. Use when you need to create a new skill for a specific task, workflow, or domain. Helps design skill structure, write SKILL.md with proper YAML frontmatter, and organize supporting files.
---

# Skill Creator

Creates new skills following the Antigravity format.

## When to use this skill

- When you need to create a new skill for a specific task or workflow
- When a team needs a shared skill for repeated patterns
- When integrating a new process or tool that deserves structured guidance

## Skill Structure

Every skill is a folder with at least one required file:

```
<skill-name>/
├── SKILL.md              ← REQUIRED: main instructions with YAML frontmatter
├── scripts/              ← optional: helper scripts
├── examples/             ← optional: reference implementations
└── resources/            ← optional: templates and assets
```

## How to create a skill

### Step 1: Choose a name

- Lowercase
- Hyphens for spaces
- Descriptive of the capability
- Examples: `api-design`, `bug-triaging`, `code-review`

### Step 2: Create the folder

Place it inside `.agent/skills/` in your project:

```
.agent/skills/<skill-name>/
```

For global skills (shared across projects), place in `~/.nous/skills/<skill-name>/`.

### Step 3: Write SKILL.md

SKILL.md has two parts:

#### YAML Frontmatter (required at top)

```yaml
---
name: my-skill
description: Brief description in third person. Include keywords for discoverability. Explain what the skill does and when an agent should use it.
---
```

**Required fields:**

| Field | Details |
|-------|---------|
| `name` | Unique identifier. Defaults to folder name if omitted. Lowercase, hyphens. |
| `description` | **Required**. Third person. Include keywords. Explain what and when. |

#### Markdown Content (required after frontmatter)

```markdown
# My Skill

## Overview
Brief description of what this skill does.

## When to use
- Use when...
- Use this skill when...

## How to use
Step-by-step instructions...

## Examples
Brief examples of the skill in action.
```

### Step 4: Add supporting files (optional)

#### scripts/
Helper scripts the skill uses. Keep them focused and small.

```
scripts/
├── setup.sh
└── validate.sh
```

#### examples/
Reference implementations showing the skill in action.

```
examples/
├── example-1.md
└── example-2.md
```

#### resources/
Templates, configs, and assets.

```
resources/
├── template.md
└── config.yaml
```

## Best practices

1. **Description in third person**: "Helps with..." not "I help with..."
2. **Keywords**: Include domain-specific terms for discoverability
3. **Clear When to use**: List specific situations, not generic ones
4. **Step-by-step How to use**: Agents follow sequential instructions well
5. **Examples**: Real examples beat abstract descriptions
6. **Minimal scripts**: Scripts should be small helpers, not complex programs
7. **Consistent naming**: Use `kebab-case` for folder and file names

## Validation checklist

Before finishing a skill, verify:

- [ ] `SKILL.md` exists in the skill folder
- [ ] YAML frontmatter has `name` and `description`
- [ ] Description is in third person and includes keywords
- [ ] Content explains when to use and how to use
- [ ] Folder name matches the skill name or is a valid alias
- [ ] All linked resources (scripts, examples) actually exist

## Output format

When creating a skill, produce:

1. Skill folder path: `.agent/skills/<name>/`
2. SKILL.md with frontmatter + content
3. Any supporting files (scripts/, examples/, resources/) as needed

## Notes

- Skills are discovered progressively: agent sees name + description first, reads full content when relevant
- You don't need to explicitly invoke skills — the agent decides based on context, but you can reference a skill by name to ensure it's loaded
- Global skills in `~/.nous/skills/` apply to all projects
- Project skills in `.agent/skills/` apply only to that project