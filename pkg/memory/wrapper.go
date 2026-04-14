package memory

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

type MemPalaceWrapper struct {
	basePath string
	venvPath string
}

func NewMemPalaceWrapper(basePath, venvPath string) *MemPalaceWrapper {
	return &MemPalaceWrapper{
		basePath: basePath,
		venvPath: venvPath,
	}
}

// Status returns mempalace status via Python CLI
func (m *MemPalaceWrapper) Status() (string, error) {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mempalace status failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}

// Search performs a semantic search in MemPalace
func (m *MemPalaceWrapper) Search(query string) (string, error) {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "search", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mempalace search failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}

// Initialize sets up MemPalace for a project
func (m *MemPalaceWrapper) Init(projectDir string) error {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "init", projectDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mempalace init failed: %w\n%s", err, string(output))
	}
	_ = output // suppress unused
	return nil
}

// Mine project files
func (m *MemPalaceWrapper) Mine(dir string) error {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "mine", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mempalace mine failed: %w\n%s", err, string(output))
	}
	_ = output
	return nil
}

// WakeUp generates wake-up context
func (m *MemPalaceWrapper) WakeUp(wing string) (string, error) {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	args := []string{"-m", "mempalace", "wake-up"}
	if wing != "" {
		args = append(args, "--wing", wing)
	}

	cmd := exec.Command(python, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mempalace wake-up failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}

// AddDrawer adds verbatim content to storage
func (m *MemPalaceWrapper) AddDrawer(wing, room, content string) error {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "add-drawer",
		"--wing", wing, "--room", room, "--content", content)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mempalace add-drawer failed: %w\n%s", err, string(output))
	}
	_ = output
	return nil
}

// ListWings returns all wings in the palace
func (m *MemPalaceWrapper) ListWings() (string, error) {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	cmd := exec.Command(python, "-m", "mempalace", "list-wings")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mempalace list-wings failed: %w\n%s", err, string(output))
	}
	return string(output), nil
}