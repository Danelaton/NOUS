package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentAdapter defines interface for agent configuration
type AgentAdapter interface {
	Inject(nousDir string) error
	Detect() bool
	AgentName() string
}

// BaseAdapter provides common detection logic
type BaseAdapter struct {
	name       string
	configDirs []string
}

func (b *BaseAdapter) AgentName() string { return b.name }

func (b *BaseAdapter) Detect() bool {
	for _, dir := range b.configDirs {
		if _, err := os.Stat(dir); err == nil {
			return true
		}
	}
	return false
}

// ── OpenCode ─────────────────────────────────────────────────────────────────

type OpenCodeAdapter struct{ BaseAdapter }

func NewOpenCodeAdapter() *OpenCodeAdapter {
	home, _ := os.UserHomeDir()
	return &OpenCodeAdapter{BaseAdapter{
		name:       "opencode",
		configDirs: []string{filepath.Join(home, ".opencode")},
	}}
}

func (a *OpenCodeAdapter) Inject(nousDir string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(nousDir, "config", "opencode", "settings.json")
	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
		"paths": map[string]interface{}{
			"nous": nousDir,
			"home": home,
		},
	}
	return writeConfigFile(configPath, config)
}

// ── Claude ────────────────────────────────────────────────────────────────────

type ClaudeAdapter struct{ BaseAdapter }

func NewClaudeAdapter() *ClaudeAdapter {
	home, _ := os.UserHomeDir()
	return &ClaudeAdapter{BaseAdapter{
		name:       "claude",
		configDirs: []string{filepath.Join(home, ".claude")},
	}}
}

func (a *ClaudeAdapter) Inject(nousDir string) error {
	configPath := filepath.Join(nousDir, "config", "claude", "config.json")
	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
	}
	return writeConfigFile(configPath, config)
}

// ── Cursor ────────────────────────────────────────────────────────────────────

type CursorAdapter struct{ BaseAdapter }

func NewCursorAdapter() *CursorAdapter {
	home, _ := os.UserHomeDir()
	return &CursorAdapter{BaseAdapter{
		name:       "cursor",
		configDirs: []string{filepath.Join(home, ".cursor")},
	}}
}

func (a *CursorAdapter) Inject(nousDir string) error {
	configPath := filepath.Join(nousDir, "config", "cursor", "settings.json")
	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
	}
	return writeConfigFile(configPath, config)
}

// ── Kiro ──────────────────────────────────────────────────────────────────────

type KiroAdapter struct{ BaseAdapter }

func NewKiroAdapter() *KiroAdapter {
	home, _ := os.UserHomeDir()
	return &KiroAdapter{BaseAdapter{
		name:       "kiro",
		configDirs: []string{filepath.Join(home, ".kiro")},
	}}
}

func (a *KiroAdapter) Inject(nousDir string) error {
	configPath := filepath.Join(nousDir, "config", "kiro", "config.json")
	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
		"steering": map[string]interface{}{
			"enabled":       true,
			"orchestration": "sdd",
		},
	}
	return writeConfigFile(configPath, config)
}

// ── Roo ───────────────────────────────────────────────────────────────────────

type RooAdapter struct{ BaseAdapter }

func NewRooAdapter() *RooAdapter {
	home, _ := os.UserHomeDir()
	return &RooAdapter{BaseAdapter{
		name:       "roo",
		configDirs: []string{filepath.Join(home, ".roo")},
	}}
}

func (a *RooAdapter) Inject(nousDir string) error {
	configPath := filepath.Join(nousDir, "config", "roo", "config.json")
	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
		"subagents": map[string]interface{}{
			"enabled": true,
		},
	}
	return writeConfigFile(configPath, config)
}

// ── Hermes ────────────────────────────────────────────────────────────────────

type HermesAdapter struct{ BaseAdapter }

func NewHermesAdapter() *HermesAdapter {
	hermesHome := resolveHermesHome()
	return &HermesAdapter{BaseAdapter{
		name:       "hermes",
		configDirs: []string{hermesHome},
	}}
}

// resolveHermesHome returns the Hermes home directory.
// On Windows, this is %APPDATA%/hermes (not ~/.hermes).
func resolveHermesHome() string {
	// 1. HERMES_HOME env var (explicit override)
	if hh := os.Getenv("HERMES_HOME"); hh != "" {
		return hh
	}
	// 2. XDG-compatible (Linux/Mac)
	home, _ := os.UserHomeDir()
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "hermes")
	}
	// 3. Windows: %APPDATA%/hermes
	if appData := os.Getenv("APPDATA"); appData != "" {
		p := filepath.Join(appData, "hermes")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// 4. Fallback: ~/.hermes
	return filepath.Join(home, ".hermes")
}

func (a *HermesAdapter) Inject(nousDir string) error {
	hermesDir := resolveHermesHome()

	fmt.Println("[NOUS] Injecting Hermes configuration...")

	// 1. Ensure ~/.hermes/skills/ directories exist
	engDir := filepath.Join(hermesDir, "skills", "engineering")
	nousSkDir := filepath.Join(hermesDir, "skills", "nous")
	for _, d := range []string{engDir, nousSkDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", d, err)
		}
	}

	// 2. Copy hermes-skills from ~/.nous/hermes-skills/ to ~/.hermes/skills/
	// Engineering skills (8): codebase-design, domain-modeling, grilling,
	//   grill-with-docs, improve-codebase-architecture, openspec, wayfinder,
	//   nous-agent
	// NOUS skills (5): project-map, architecture-review, okf-knowledge,
	//   skill-creator, nous-pipeline
	//   (nous-sync-check is dev-only — maintained locally, never distributed)
	engSkills := []string{
		"codebase-design", "domain-modeling", "grilling", "grill-with-docs",
		"improve-codebase-architecture", "openspec", "wayfinder", "nous-agent",
	}
	nousSkills := []string{
		"project-map", "architecture-review", "okf-knowledge", "skill-creator",
		"nous-pipeline",
	}

	srcBase := filepath.Join(nousDir, "hermes-skills")

	for _, name := range engSkills {
		src := filepath.Join(srcBase, name)
		dst := filepath.Join(engDir, name)
		if _, err := os.Stat(src); err == nil {
			if err := copyDir(src, dst); err != nil {
				fmt.Printf("[NOUS] Warning: failed to copy %s: %v\n", name, err)
			} else {
				fmt.Printf("[NOUS]   skill engineering/%s installed\n", name)
			}
		}
	}
	for _, name := range nousSkills {
		src := filepath.Join(srcBase, name)
		dst := filepath.Join(nousSkDir, name)
		if _, err := os.Stat(src); err == nil {
			if err := copyDir(src, dst); err != nil {
				fmt.Printf("[NOUS] Warning: failed to copy %s: %v\n", name, err)
			} else {
				fmt.Printf("[NOUS]   skill nous/%s installed\n", name)
			}
		}
	}

	// 3. Copy AGENTS-HERMES.md to ~/.hermes/
	agentsSrc := filepath.Join(nousDir, "hermes-skills", "AGENTS-HERMES.md")
	agentsDst := filepath.Join(hermesDir, "AGENTS-HERMES.md")
	if _, err := os.Stat(agentsSrc); err == nil {
		if err := copyFile(agentsSrc, agentsDst); err != nil {
			fmt.Printf("[NOUS] Warning: failed to copy AGENTS-HERMES.md: %v\n", err)
		} else {
			fmt.Printf("[NOUS]   AGENTS-HERMES.md → ~/.hermes/\n")
		}
	}

	// 4. Initialize OKF _system/ if not exists
	okfSystem := filepath.Join(hermesDir, "OKF", "_system")
	if err := os.MkdirAll(okfSystem, 0755); err != nil {
		return fmt.Errorf("failed to create OKF _system/: %w", err)
	}
	conventionsPath := filepath.Join(okfSystem, "conventions.md")
	if _, err := os.Stat(conventionsPath); os.IsNotExist(err) {
		conventions := `---
type: Conventions
title: Hermes + NOUS Conventions
description: Cross-project conventions injected by nous sync hermes.
tags: [conventions, system, hermes, nous]
---

# Conventions

## Paths
- **tmp-repos**: ` + "`" + `~/Downloads/tmp-repos/` + "`" + ` — canonical folder for external repo clones.
- **OKF global**: ` + "`" + `~/.hermes/OKF/` + "`" + ` — cross-project knowledge library.
- **OKF local**: ` + "`" + `.agents/OKF/` + "`" + ` — per-project reference.

## Language
- Conversation: Spanish (es).
- Code, commits, docs, UI strings: English (en).

## Skills
- Bundle ` + "`" + `/nous` + "`" + ` loads all engineering + NOUS skills.
- **Engineering**: grilling, domain-modeling, codebase-design, grill-with-docs,
  wayfinder, improve-codebase-architecture, openspec, nous-agent.
- **NOUS**: project-map, architecture-review, okf-knowledge, skill-creator.
- **nous-agent**: Loaded automatically — session start protocol, backups, git safety, OKF maintenance, Hermes tool usage.

## Workflow
1. /grill-with-docs → align + CONTEXT.md + ADRs
2. /openspec → proposal + tasks
3. /writing-plans → bite-sized plan
4. subagent-driven-dev → execute with TDD
5. /code-review → verify diff

## Safety
- Git: no commit/push without explicit user confirmation.
- Backups: pre-mutation copy to dev/backups/YYYYMMDD_HHMMSS_filename.ext.
- Credentials: .env only, never hardcoded.
`
		if err := os.WriteFile(conventionsPath, []byte(conventions), 0644); err != nil {
			fmt.Printf("[NOUS] Warning: failed to create conventions.md: %v\n", err)
		} else {
			fmt.Printf("[NOUS]   OKF _system/conventions.md created\n")
		}
	}

	// 5. Update OKF index.md to include _system
	okfIndex := filepath.Join(hermesDir, "OKF", "index.md")
	if _, err := os.Stat(okfIndex); os.IsNotExist(err) {
		indexContent := `---
type: Directory
title: Hermes OKF Knowledge Library
description: Master catalog of all known projects and their knowledge bundles.
tags: [okf, index, global-brain]
---

# Hermes OKF Knowledge Library

## System
* [_system/](_system/index.md) — Conventions, skills catalog, and workflow reference.

## Projects
<!-- Add project entries here -->
`
		if err := os.WriteFile(okfIndex, []byte(indexContent), 0644); err != nil {
			fmt.Printf("[NOUS] Warning: failed to create OKF index.md: %v\n", err)
		} else {
			fmt.Printf("[NOUS]   OKF index.md initialized\n")
		}
	}

	// Create _system/index.md
	sysIndex := filepath.Join(okfSystem, "index.md")
	if _, err := os.Stat(sysIndex); os.IsNotExist(err) {
		sysIndexContent := `# System Knowledge

* [Conventions](conventions.md) — Cross-project conventions, paths, and standards.
* [Skills Catalog](skills-catalog.md) — Available skills and bundles.
`
		if err := os.WriteFile(sysIndex, []byte(sysIndexContent), 0644); err != nil {
			fmt.Printf("[NOUS] Warning: failed to create _system/index.md: %v\n", err)
		}
	}

	// 6. Create/update SOUL.md with NOUS operating protocols (identity, always loaded)
	soulPath := filepath.Join(hermesDir, "SOUL.md")
	soulContent := `You are Hermes Agent, an intelligent AI assistant created by Nous Research. You are helpful, knowledgeable, and direct. You assist users with a wide range of tasks including answering questions, writing and editing code, analyzing information, creative work, and executing actions via your tools. You communicate clearly, admit uncertainty when appropriate, and prioritize being genuinely useful over being verbose unless otherwise directed below. Be targeted and efficient in your exploration and investigations.

# NOUS Operating Protocols (global)

These are the user's standing conventions — apply them in EVERY session, regardless of working directory.

## Language
- Conversation with the user: Spanish (es).
- Code, commits, docs, UI strings, and technical artifacts: English (en).
- Never mix languages in code or technical artifacts.

## Session Start Protocol (every session)
1. If the working directory has ` + "`.agents/MEMORY.md`" + `, read it first — active context, blockers, next action.
2. Read ` + "`~/.hermes/OKF/index.md`" + ` (global catalog) — follow only links relevant to the current task (progressive disclosure, don't load everything).
3. Use the ` + "`memory`" + ` tool for frequently-needed facts (auto-injected every turn).
4. Use ` + "`session_search(\"query\")`" + ` for historical conversation context.

## OKF Knowledge Maintenance
- After meaningful work (decisions, verified runbooks, diagnosed failures, architecture insights), update the relevant OKF concept at ` + "`~/.hermes/OKF/<project>/`" + `.
- Add milestones to ` + "`~/.hermes/OKF/log.md`" + ` (YYYY-MM-DD, newest first).
- NEVER duplicate durable knowledge across MEMORY.md and OKF concepts.
- NEVER migrate unverified or low-value chatter into OKF.

## Git Safety
- No ` + "`git commit`" + ` or ` + "`git push`" + ` without explicit user confirmation after showing ` + "`git diff`" + `.
- Actions on APIs, Cloud, or CI/CD require a detailed plan and prior human approval.
- Use ` + "`--no-verify`" + ` or ` + "`HUSKY=0`" + ` when husky hooks block legitimate operations.

## Backups
- Before editing any file outside ` + "`dev/sandbox/`" + `, create a copy in ` + "`dev/backups/`" + ` with format ` + "`YYYYMMDD_HHMMSS_filename.ext`" + `.
- Never execute rollback without explicit user confirmation and a proposed diff.

## Workflow Preference
- Prefer action over planning: "continua" means execute the next step immediately without asking for clarification.
- Deliver end-to-end working systems (skill + scripts + cron + delivery), not theoretical plans.
`
	if _, err := os.Stat(soulPath); os.IsNotExist(err) {
		if err := os.WriteFile(soulPath, []byte(soulContent), 0644); err != nil {
			fmt.Printf("[NOUS] Warning: failed to create SOUL.md: %v\n", err)
		} else {
			fmt.Printf("[NOUS]   SOUL.md created with NOUS operating protocols\n")
		}
	}

	// 7. Create/update skill-bundles/nous.yaml — /nous loads all NOUS skills
	bundlesDir := filepath.Join(hermesDir, "skill-bundles")
	if err := os.MkdirAll(bundlesDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill-bundles/: %w", err)
	}
	bundlePath := filepath.Join(bundlesDir, "nous.yaml")
	bundleContent := `name: nous
description: NOUS engineering + knowledge bundle — load all NOUS skills together.
skills:
  - codebase-design
  - domain-modeling
  - grilling
  - grill-with-docs
  - improve-codebase-architecture
  - openspec
  - wayfinder
  - nous-agent
  - nous-pipeline
  - project-map
  - architecture-review
  - okf-knowledge
  - skill-creator
instruction: |
  NOUS engineering and knowledge skills loaded together. Use them for the
  full pipeline: grill → openspec → plan → execute (TDD) → review. Follow
  the workflow in nous-agent skill and the conventions in
  ~/.hermes/OKF/_system/conventions.md. Use nous-pipeline to orchestrate the
  full SDD flow end-to-end.
`
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		if err := os.WriteFile(bundlePath, []byte(bundleContent), 0644); err != nil {
			fmt.Printf("[NOUS] Warning: failed to create skill-bundles/nous.yaml: %v\n", err)
		} else {
			fmt.Printf("[NOUS]   skill-bundles/nous.yaml created (bundle /nous)\n")
		}
	}

	fmt.Println("[NOUS] Hermes enhanced. Run 'nous sync' in your project to install AGENTS-HERMES.md.")
	return nil
}

// copyDir copies a directory recursively from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func writeConfigFile(path string, config map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	wrapper := map[string]interface{}{"nous": config}
	jsonBytes, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, jsonBytes, 0644)
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetAllAdapters returns all supported agent adapters.
func GetAllAdapters() []AgentAdapter {
	return []AgentAdapter{
		NewOpenCodeAdapter(),
		NewClaudeAdapter(),
		NewCursorAdapter(),
		NewKiroAdapter(),
		NewRooAdapter(),
		NewHermesAdapter(),
	}
}

// DetectAvailableAgents returns names of agents installed on this machine.
func DetectAvailableAgents() []string {
	var detected []string
	for _, adapter := range GetAllAdapters() {
		if adapter.Detect() {
			detected = append(detected, adapter.AgentName())
		}
	}
	return detected
}
