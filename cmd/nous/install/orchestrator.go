package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Orchestrator struct {
	nousDir string // always ~/.nous — global, never inside a project
	system  System
}

func NewOrchestrator() (*Orchestrator, error) {
	sys, err := Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect system: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &Orchestrator{
		nousDir: filepath.Join(home, ".nous"),
		system:  sys,
	}, nil
}

func (o *Orchestrator) Run() error {
	fmt.Printf("[NOUS] Setting up ~/.nous/ ...\n")

	if !o.system.IsSupported() {
		return fmt.Errorf("unsupported system — Go 1.21+ required")
	}

	// Phase 1: create ~/.nous/config/
	configDir := filepath.Join(o.nousDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Phase 2: ensure ~/.nous/skills/ exists
	skillsDstDir := filepath.Join(o.nousDir, "skills")
	if err := os.MkdirAll(skillsDstDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}
	dstAgents := filepath.Join(skillsDstDir, "AGENTS.md")
	if _, err := os.Stat(dstAgents); os.IsNotExist(err) {
		var srcDir string
		if runtime.GOOS == "windows" {
			srcDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "nous", "skills")
		} else {
			srcDir = filepath.Join(o.nousDir, "skills") // fallback: same dir
		}
		srcAgents := filepath.Join(srcDir, "AGENTS.md")
		if _, err := os.Stat(srcAgents); err == nil {
			if err := copyFile(srcAgents, dstAgents); err != nil {
				fmt.Printf("[NOUS] Warning: failed to restore AGENTS.md: %v\n", err)
			}
		}
	}

	fmt.Printf("[NOUS] ~/.nous/ ready at: %s\n", o.nousDir)
	fmt.Printf("[NOUS] Run 'nous sync' in any project to install skills and project structure.\n")
	return nil
}

// SetupProject creates the project structure: dev/, .agents/, and copies AGENTS.md.
func (o *Orchestrator) SetupProject(projectDir string) error {
	fmt.Printf("[NOUS] Setting up project structure...\n")

	// ── 1. Create dev/ with all subdirectories ──────────────────────────────
	devDirs := []string{
		"sandbox",
		"tmp-repos",
		"docs",
		"scripts",
		"tests",
		"backups",
	}
	for _, d := range devDirs {
		path := filepath.Join(projectDir, "dev", d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create dev/%s: %w", d, err)
		}
	}
	fmt.Printf("[NOUS] dev/ structure created\n")

	// ── 2. Create .agents/ directory ─────────────────────────────────────────
	agentDir := filepath.Join(projectDir, ".agents")
	agentSkillsDir := filepath.Join(agentDir, "skills")
	if err := os.MkdirAll(agentSkillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/: %w", err)
	}
	fmt.Printf("[NOUS] .agents/ directory created\n")

	// ── 3. Initialize project memory and OKF knowledge bundle ────────────────
	if err := initializeKnowledge(agentDir); err != nil {
		return err
	}

	// ── 5. Add entries to .gitignore ──────────────────────────────────────────
	gitignore := filepath.Join(projectDir, ".gitignore")
	for _, entry := range []string{"dev/", ".agents/"} {
		if err := addGitignoreEntry(gitignore, entry); err != nil {
			fmt.Printf("[NOUS] Warning: could not update .gitignore: %v\n", err)
		}
	}
	fmt.Printf("[NOUS] dev/ and .agents/ added to .gitignore\n")

	// ── 6. Backup existing AGENTS.md ─────────────────────────────────────────
	agentsDst := filepath.Join(projectDir, "AGENTS.md")
	if _, err := os.Stat(agentsDst); err == nil {
		ts := time.Now().Format("20060102_150405")
		backup := filepath.Join(projectDir, "dev", "backups", "AGENTS.md."+ts)
		if err := copyFile(agentsDst, backup); err != nil {
			fmt.Printf("[NOUS] Warning: failed to backup existing AGENTS.md: %v\n", err)
		} else {
			fmt.Printf("[NOUS] Existing AGENTS.md backed up to dev/backups/\n")
		}
	}

	// ── 7. Copy AGENTS.md to project ─────────────────────────────────────────
	agentsSrc := filepath.Join(o.nousDir, "skills", "AGENTS.md")
	if _, err := os.Stat(agentsSrc); err != nil {
		fmt.Printf("[NOUS] Warning: AGENTS.md not found in ~/.nous/skills/ — skipping (run 'nous install' to download skills)\n")
	} else if err := copyFile(agentsSrc, agentsDst); err != nil {
		fmt.Printf("[NOUS] Warning: failed to copy AGENTS.md: %v\n", err)
	} else {
		fmt.Printf("[NOUS] AGENTS.md installed in project\n")
	}

	// ── 8. Copy skill folders to .agents/skills/ ──────────────────────────────
	skillsSrcDir := filepath.Join(o.nousDir, "skills")
	skillsDstDir := filepath.Join(projectDir, ".agents", "skills")
	if err := os.MkdirAll(skillsDstDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/skills/: %w", err)
	}
	entries, err := os.ReadDir(skillsSrcDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "config" {
				continue // skip files and config dir
			}
			srcPath := filepath.Join(skillsSrcDir, entry.Name())
			dstPath := filepath.Join(skillsDstDir, entry.Name())
			// Merge: overwrite files that exist in src, keep files only in dst
			if err := mergeDir(srcPath, dstPath); err != nil {
				fmt.Printf("[NOUS] Warning: failed to merge skill %s: %v\n", entry.Name(), err)
			} else {
				fmt.Printf("[NOUS] skill %s merged into project\n", entry.Name())
			}
		}
	}

	return nil
}

// Status returns a human-readable summary of the current install state.
func (o *Orchestrator) Status() string {
	skillsDir := filepath.Join(o.nousDir, "skills")
	skillsInfo := "not found"
	if entries, err := os.ReadDir(skillsDir); err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				count++
			}
		}
		skillsInfo = fmt.Sprintf("%s (%d skill folders)", skillsDir, count)
	}
	return fmt.Sprintf("  nousDir:  %s\n  skills:   %s\n", o.nousDir, skillsInfo)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func initializeKnowledge(agentDir string) error {
	files := map[string]string{
		"MEMORY.md": `# NOUS Memory Router

## Active Context

- Current work:
- Blockers:
- Next action:

## Knowledge Router

- Start with [OKF/index.md](OKF/index.md).
- Follow only links relevant to the current task.
- Store durable architecture, decisions, runbooks, and solved problems in the OKF bundle.
- Keep this file concise; it is a router and working-state summary, not the knowledge archive.

## Legacy Context

- If ` + "`docs_index.md`" + ` exists, treat it as a legacy index and preserve it.
- If ` + "`../dev/docs/`" + ` contains useful history, migrate durable knowledge into OKF incrementally.
`,
		filepath.Join("OKF", "index.md"): `---
okf_version: "0.1"
---

# Project Knowledge

Read this catalog first, then open only the concepts relevant to the task.

## Core

* [Architecture](architecture.md) - High-level system design and boundaries.
* [Workflows](workflows/) - Operational procedures and runbooks.
* [Decisions](decisions/) - Durable technical and product decisions.
* [Troubleshooting](troubleshooting/) - Diagnosed failures and verified solutions.
* [References](references/) - Curated project references and external sources.
* [Update log](log.md) - Major knowledge milestones, newest first.
`,
		filepath.Join("OKF", "log.md"): `# Project Knowledge Update Log

`,
		filepath.Join("OKF", "architecture.md"): `---
type: Architecture
title: Project Architecture
description: High-level system structure, boundaries, dependencies, and constraints.
tags: [architecture]
---

# Project Architecture

Record only verified architectural knowledge. Link to supporting concepts and sources.
`,
		filepath.Join("OKF", "workflows", "index.md"): `# Workflows

* [Project runbook](runbook.md) - Verified commands for setup, development, testing, deployment, and rollback.
`,
		filepath.Join("OKF", "workflows", "runbook.md"): `---
type: Workflow
title: Project Runbook
description: Verified operational commands and recovery procedures for this project.
tags: [workflow, runbook]
---

# Project Runbook

Document commands only after they have been verified.
`,
		filepath.Join("OKF", "decisions", "index.md"): `# Decisions

Add one concept document per durable decision and link it here.
`,
		filepath.Join("OKF", "troubleshooting", "index.md"): `# Troubleshooting

Add diagnosed failures and verified solutions as concept documents, then link them here.
`,
		filepath.Join("OKF", "references", "index.md"): `# References

Add curated project references as concept documents, then link them here.
`,
	}

	for relativePath, content := range files {
		path := filepath.Join(agentDir, relativePath)
		created, err := writeFileIfNotExists(path, content)
		if err != nil {
			return fmt.Errorf("failed to initialize %s: %w", relativePath, err)
		}
		if created {
			fmt.Printf("[NOUS] .agents/%s created\n", filepath.ToSlash(relativePath))
		}
	}

	fmt.Printf("[NOUS] Project memory and OKF knowledge bundle ready\n")
	return nil
}

func writeFileIfNotExists(path, content string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return false, err
	}
	return true, nil
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

// SyncSkills copies skills from ~/.nous/skills/ to .agents/skills/ using merge (no delete).
func (o *Orchestrator) SyncSkills(projectDir string) error {
	fmt.Printf("[NOUS] Syncing skills (merge)...\n")
	skillsSrcDir := filepath.Join(o.nousDir, "skills")
	skillsDstDir := filepath.Join(projectDir, ".agents", "skills")
	if err := os.MkdirAll(skillsDstDir, 0755); err != nil {
		return fmt.Errorf("failed to create .agents/skills/: %w", err)
	}
	entries, err := os.ReadDir(skillsSrcDir)
	if err != nil {
		return fmt.Errorf("failed to read ~/.nous/skills/: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "config" {
			continue
		}
		srcPath := filepath.Join(skillsSrcDir, entry.Name())
		dstPath := filepath.Join(skillsDstDir, entry.Name())
		if err := mergeDir(srcPath, dstPath); err != nil {
			fmt.Printf("[NOUS] Warning: failed to merge skill %s: %v\n", entry.Name(), err)
		} else {
			fmt.Printf("[NOUS] skill %s merged into project\n", entry.Name())
		}
	}
	fmt.Printf("[NOUS] Skills sync complete\n")
	return nil
}

// mergeDir copies files from src to dst without removing existing content.
// Existing files in dst are overwritten only if they exist in src.
func mergeDir(src, dst string) error {
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

// addGitignoreEntry appends an entry to .gitignore if not already present.
func addGitignoreEntry(gitignorePath, entry string) error {
	var existing []byte
	if _, err := os.Stat(gitignorePath); err == nil {
		existing, err = os.ReadFile(gitignorePath)
		if err != nil {
			return err
		}
		if strings.Contains(string(existing), entry) {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = f.WriteString(entry + "\n")
	return err
}
