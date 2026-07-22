package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitializeKnowledgeCreatesOKFBundle(t *testing.T) {
	agentDir := filepath.Join(t.TempDir(), ".agents")

	if err := initializeKnowledge(agentDir); err != nil {
		t.Fatalf("initializeKnowledge returned an error: %v", err)
	}

	requiredFiles := []string{
		"MEMORY.md",
		filepath.Join("OKF", "index.md"),
		filepath.Join("OKF", "log.md"),
		filepath.Join("OKF", "architecture.md"),
		filepath.Join("OKF", "workflows", "index.md"),
		filepath.Join("OKF", "workflows", "runbook.md"),
		filepath.Join("OKF", "decisions", "index.md"),
		filepath.Join("OKF", "troubleshooting", "index.md"),
		filepath.Join("OKF", "references", "index.md"),
	}

	for _, relativePath := range requiredFiles {
		if _, err := os.Stat(filepath.Join(agentDir, relativePath)); err != nil {
			t.Errorf("expected %s to exist: %v", relativePath, err)
		}
	}

	if _, err := os.Stat(filepath.Join(agentDir, "docs_index.md")); !os.IsNotExist(err) {
		t.Errorf("legacy docs_index.md should not be created")
	}
}

func TestInitializeKnowledgePreservesExistingContent(t *testing.T) {
	agentDir := filepath.Join(t.TempDir(), ".agents")

	if err := initializeKnowledge(agentDir); err != nil {
		t.Fatalf("first initializeKnowledge returned an error: %v", err)
	}

	memoryPath := filepath.Join(agentDir, "MEMORY.md")
	architecturePath := filepath.Join(agentDir, "OKF", "architecture.md")
	if err := os.WriteFile(memoryPath, []byte("custom memory"), 0644); err != nil {
		t.Fatalf("failed to write custom memory: %v", err)
	}
	if err := os.WriteFile(architecturePath, []byte("custom architecture"), 0644); err != nil {
		t.Fatalf("failed to write custom architecture: %v", err)
	}

	if err := initializeKnowledge(agentDir); err != nil {
		t.Fatalf("second initializeKnowledge returned an error: %v", err)
	}

	assertFileContent(t, memoryPath, "custom memory")
	assertFileContent(t, architecturePath, "custom architecture")
}

func TestOKFConceptTemplatesHaveTypeFrontmatter(t *testing.T) {
	agentDir := filepath.Join(t.TempDir(), ".agents")

	if err := initializeKnowledge(agentDir); err != nil {
		t.Fatalf("initializeKnowledge returned an error: %v", err)
	}

	conceptPaths := []string{
		filepath.Join(agentDir, "OKF", "architecture.md"),
		filepath.Join(agentDir, "OKF", "workflows", "runbook.md"),
	}

	for _, path := range conceptPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		text := string(content)
		if !strings.HasPrefix(text, "---\n") || !strings.Contains(text, "\ntype: ") {
			t.Errorf("%s does not contain OKF type frontmatter", path)
		}
	}
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(content) != expected {
		t.Errorf("%s was overwritten: got %q, want %q", path, content, expected)
	}
}
