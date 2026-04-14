package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// hookPath returns the correct hook file for the current OS
func hookPath(nousDir, hookName string) string {
	hooksDir := filepath.Join(nousDir, "hooks")
	if runtime.GOOS == "windows" {
		return filepath.Join(hooksDir, hookName+".ps1")
	}
	return filepath.Join(hooksDir, hookName+".sh")
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
		"mempalace": map[string]interface{}{
			"enabled": true,
			"venv":    filepath.Join(nousDir, "venv"),
		},
		"hooks": map[string]interface{}{
			"Stop":       hookPath(nousDir, "mempal_save_hook"),
			"PreCompact": hookPath(nousDir, "mempal_precompact_hook"),
		},
		"openspec": map[string]interface{}{
			"enabled": true,
		},
		"mcp": map[string]interface{}{
			"enabled": true,
			"server":  "mempalace",
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
		"mempalace": map[string]interface{}{
			"enabled": true,
			"venv":    filepath.Join(nousDir, "venv"),
		},
		"hooks": map[string]interface{}{
			"Stop":       hookPath(nousDir, "mempal_save_hook"),
			"PreCompact": hookPath(nousDir, "mempal_precompact_hook"),
		},
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
		"mempalace": map[string]interface{}{
			"enabled": true,
			"venv":    filepath.Join(nousDir, "venv"),
		},
		"hooks": map[string]interface{}{
			"Stop":       hookPath(nousDir, "mempal_save_hook"),
			"PreCompact": hookPath(nousDir, "mempal_precompact_hook"),
		},
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
		"mempalace": map[string]interface{}{
			"enabled": true,
			"venv":    filepath.Join(nousDir, "venv"),
		},
		"hooks": map[string]interface{}{
			"Stop":       hookPath(nousDir, "mempal_save_hook"),
			"PreCompact": hookPath(nousDir, "mempal_precompact_hook"),
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
		"mempalace": map[string]interface{}{
			"enabled": true,
			"venv":    filepath.Join(nousDir, "venv"),
		},
		"hooks": map[string]interface{}{
			"Stop":       hookPath(nousDir, "mempal_save_hook"),
			"PreCompact": hookPath(nousDir, "mempal_precompact_hook"),
		},
		"subagents": map[string]interface{}{
			"enabled": true,
		},
	}
	return writeConfigFile(configPath, config)
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

// GetAllAdapters returns all supported agent adapters.
func GetAllAdapters() []AgentAdapter {
	return []AgentAdapter{
		NewOpenCodeAdapter(),
		NewClaudeAdapter(),
		NewCursorAdapter(),
		NewKiroAdapter(),
		NewRooAdapter(),
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
