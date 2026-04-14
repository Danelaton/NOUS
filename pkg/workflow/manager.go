package workflow

import (
	"fmt"
	"os"
	"path/filepath"
)

// Spec represents an OpenSpec specification
type Spec struct {
	Number   string
	Title    string
	Content  string
	Phase    string // "design", "implement", "verify", "document"
	Status   string // "draft", "proposed", "accepted", "implemented"
	Priority int    // 1-5
}

// Change represents an OpenSpec change proposal
type Change struct {
	Number      string
	Title       string
	Description string
	Files       []string
	Breaking    bool
	Status      string // "proposed", "accepted", "rejected", "implemented"
}

// Manager handles OpenSpec workflow operations
type Manager struct {
	baseDir string
	specsDir string
	changesDir string
}

// NewManager creates a new OpenSpec workflow manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir:   baseDir,
		specsDir:  filepath.Join(baseDir, "openspec", "specs"),
		changesDir: filepath.Join(baseDir, "openspec", "changes"),
	}
}

// EnsureStructure creates OpenSpec directory structure
func (m *Manager) EnsureStructure() error {
	if err := os.MkdirAll(m.specsDir, 0755); err != nil {
		return fmt.Errorf("failed to create specs dir: %w", err)
	}
	if err := os.MkdirAll(m.changesDir, 0755); err != nil {
		return fmt.Errorf("failed to create changes dir: %w", err)
	}
	return nil
}

// CreateSpec creates a new specification
func (m *Manager) CreateSpec(spec *Spec) error {
	if spec.Number == "" {
		return fmt.Errorf("spec number is required")
	}

	content := fmt.Sprintf(`# Spec: %s

## Number
%s

## Title
%s

## Phase
%s

## Status
%s

## Priority
%d

## Content
%s

---
Created by NOUS
`, spec.Title, spec.Number, spec.Title, spec.Phase, spec.Status, spec.Priority, spec.Content)

	filename := fmt.Sprintf("SPEC_%s.md", spec.Number)
	return os.WriteFile(filepath.Join(m.specsDir, filename), []byte(content), 0644)
}

// CreateChange creates a new change proposal
func (m *Manager) CreateChange(change *Change) error {
	if change.Number == "" {
		return fmt.Errorf("change number is required")
	}

	content := fmt.Sprintf(`# Change: %s

## Number
%s

## Title
%s

## Description
%s

## Files Affected
%s

## Breaking Changes
%v

## Status
%s

---
Created by NOUS
`, change.Title, change.Number, change.Title, change.Description,
	   joinStrings(change.Files), change.Breaking, change.Status)

	filename := fmt.Sprintf("CHG_%s_proposal.md", change.Number)
	return os.WriteFile(filepath.Join(m.changesDir, filename), []byte(content), 0644)
}

// ListSpecs returns all specifications
func (m *Manager) ListSpecs() ([]string, error) {
	entries, err := os.ReadDir(m.specsDir)
	if err != nil {
		return nil, err
	}

	var specs []string
	for _, e := range entries {
		if !e.IsDir() {
			specs = append(specs, e.Name())
		}
	}
	return specs, nil
}

// ListChanges returns all change proposals
func (m *Manager) ListChanges() ([]string, error) {
	entries, err := os.ReadDir(m.changesDir)
	if err != nil {
		return nil, err
	}

	var changes []string
	for _, e := range entries {
		if !e.IsDir() {
			changes = append(changes, e.Name())
		}
	}
	return changes, nil
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}