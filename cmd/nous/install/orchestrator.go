package install

import (
	"encoding/json"
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
	fmt.Printf("[NOUS] Starting installation...\n")
	fmt.Printf("[NOUS] System: %s\n", o.system.String())

	if !o.system.IsSupported() {
		return fmt.Errorf("unsupported system — Go 1.21+ required")
	}

	// Phase 1: create ~/.nous/config/
	configDir := filepath.Join(o.nousDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Phase 2: ensure ~/.nous/skills/ has AGENTS.md
	// On first install the installer placed it there. If missing, try to restore from
	// platform-specific skills directory.
	skillsDstDir := filepath.Join(o.nousDir, "skills")
	if err := os.MkdirAll(skillsDstDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}
	dstAgents := filepath.Join(skillsDstDir, "AGENTS.md")
	if _, err := os.Stat(dstAgents); os.IsNotExist(err) {
		// Try to restore from platform-specific location
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

	// Phase 3: detect agents and inject configs
	if err := o.detectAndInjectConfigs(); err != nil {
		fmt.Printf("[NOUS] Warning: config injection: %v\n", err)
	}

	fmt.Printf("[NOUS] Installation complete!\n")
	fmt.Printf("[NOUS] Config installed at: %s\n", o.nousDir)
	fmt.Printf("[NOUS] Run 'nous sdd-init' inside a project to activate SDD workflow.\n")
	return nil
}

// SetupProject creates the dev/ structure and copies AGENTS.md into the project.
func (o *Orchestrator) SetupProject(projectDir string) error {
	fmt.Printf("[NOUS] Setting up project structure...\n")

	// 1. Create dev/ with all subdirectories
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

	// 2. Add dev/ to .gitignore if not already present
	gitignore := filepath.Join(projectDir, ".gitignore")
	if err := addGitignoreEntry(gitignore, "dev/"); err != nil {
		fmt.Printf("[NOUS] Warning: could not update .gitignore: %v\n", err)
	} else {
		fmt.Printf("[NOUS] dev/ added to .gitignore\n")
	}

	// 3. Backup existing AGENTS.md
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

	// 4. Copy AGENTS.md to project
	agentsSrc := filepath.Join(o.nousDir, "skills", "AGENTS.md")
	if _, err := os.Stat(agentsSrc); err != nil {
		return fmt.Errorf("AGENTS.md not found in ~/.nous/skills/: run 'nous install' first")
	}
	if err := copyFile(agentsSrc, agentsDst); err != nil {
		return fmt.Errorf("failed to copy AGENTS.md to project: %w", err)
	}
	fmt.Printf("[NOUS] AGENTS.md installed in project\n")

	return nil
}

// detectAndInjectConfigs finds all installed agents and injects their configs.
func (o *Orchestrator) detectAndInjectConfigs() error {
	fmt.Printf("[NOUS] Detecting and configuring agents...\n")
	agents := detectAgents()
	if len(agents) == 0 {
		fmt.Printf("[NOUS] No agents detected — skipping config injection\n")
		return nil
	}
	for _, agent := range agents {
		fmt.Printf("[NOUS] Configuring %s...\n", agent)
		if err := injectConfig(agent, o.nousDir); err != nil {
			fmt.Printf("[NOUS] Warning: failed to configure %s: %v\n", agent, err)
		}
	}
	return nil
}

// SyncAgents re-injects configs for all detected agents.
func (o *Orchestrator) SyncAgents() error {
	fmt.Printf("[NOUS] Syncing agent configurations...\n")
	return o.detectAndInjectConfigs()
}

// Status returns a human-readable summary of the current install state.
func (o *Orchestrator) Status() string {
	agents := detectAgents()
	agentStr := "none detected"
	if len(agents) > 0 {
		agentStr = strings.Join(agents, ", ")
	}
	return fmt.Sprintf("  nousDir:  %s\n  agents:   %s\n", o.nousDir, agentStr)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func detectAgents() []string {
	var agents []string
	home, _ := os.UserHomeDir()
	paths := map[string]string{
		"opencode": filepath.Join(home, ".opencode"),
		"claude":   filepath.Join(home, ".claude"),
		"cursor":   filepath.Join(home, ".cursor"),
		"kiro":     filepath.Join(home, ".kiro"),
		"roo":      filepath.Join(home, ".roo"),
	}
	for name, path := range paths {
		if _, err := os.Stat(path); err == nil {
			agents = append(agents, name)
		}
	}
	return agents
}

func injectConfig(agent, nousDir string) error {
	configDir := filepath.Join(nousDir, "config", agent)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	config := map[string]interface{}{
		"openspec": map[string]interface{}{
			"enabled": true,
		},
	}

	filename := "config.json"
	if agent == "opencode" || agent == "cursor" {
		filename = "settings.json"
	}

	content := formatConfig(config)
	return os.WriteFile(filepath.Join(configDir, filename), []byte(content), 0644)
}

func formatConfig(config map[string]interface{}) string {
	wrapper := map[string]interface{}{"nous": config}
	jsonBytes, _ := json.MarshalIndent(wrapper, "", "  ")
	return string(jsonBytes) + "\n"
}

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

	// Preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// addGitignoreEntry appends an entry to .gitignore if not already present.
func addGitignoreEntry(gitignorePath, entry string) error {
	// Check if .gitignore exists
	var existing []byte
	if _, err := os.Stat(gitignorePath); err == nil {
		existing, err = os.ReadFile(gitignorePath)
		if err != nil {
			return err
		}
		// Check if entry already exists
		if strings.Contains(string(existing), entry) {
			return nil
		}
	}

	// Append entry
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		_, err = f.WriteString("\n")
		if err != nil {
			return err
		}
	}
	_, err = f.WriteString(entry + "\n")
	return err
}
