package memory

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type MCPServer struct {
	venvPath string
	port     int
	proc     *exec.Cmd
}

func NewMCPServer(venvPath string, port int) *MCPServer {
	return &MCPServer{
		venvPath: venvPath,
		port:     port,
	}
}

// Start launches the MCP server in background
func (m *MCPServer) Start() error {
	python := filepath.Join(m.venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		python = filepath.Join(m.venvPath, "Scripts", "python.exe")
	}

	mcpScript := filepath.Join(m.venvPath, "..", "mempalace", "mcp_server.py")

	// Check if mcp_server.py exists
	if _, err := os.Stat(mcpScript); os.IsNotExist(err) {
		// Create a minimal MCP server for integration
		return m.createMinimalMCPServer(mcpScript)
	}

	args := []string{mcpScript, "--port", strconv.Itoa(m.port)}
	cmd := exec.Command(python, args...)
	cmd.Start()

	m.proc = cmd

	// Wait briefly for server to start
	time.Sleep(500 * time.Millisecond)

	return nil
}

// Stop terminates the MCP server
func (m *MCPServer) Stop() error {
	if m.proc != nil && m.proc.Process != nil {
		return m.proc.Process.Kill()
	}
	return nil
}

// IsRunning checks if the MCP server is alive
func (m *MCPServer) IsRunning() bool {
	if m.proc == nil || m.proc.Process == nil {
		return false
	}

	err := m.proc.Process.Signal(os.Signal(nil))
	return err == nil
}

func (m *MCPServer) createMinimalMCPServer(path string) error {
	content := `# Minimal MCP server for NOUS integration
import sys
import json

def handle_request(req):
    method = req.get("method", "")
    if method == "tools/list":
        return {
            "result": {
                "tools": [
                    {"name": "mempalace_status", "description": "Get palace status"},
                    {"name": "mempalace_search", "description": "Search memories"},
                    {"name": "mempalace_add_drawer", "description": "Add verbatim content"},
                ]
            }
        }
    return {"error": "Method not found"}

if __name__ == "__main__":
    print("[NOUS] MCP server initialized", file=sys.stderr)
`
	return os.WriteFile(path, []byte(content), 0644)
}