package install

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type System struct {
	OS      string
	Arch    string
	HasGo   bool
	HasPython bool
	HasNode  bool
	GoVersion string
	PythonVersion string
	NodeVersion string
}

func Detect() (System, error) {
	sys := System{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Detect Go
	if goVer, err := runCommand("go", "version"); err == nil && goVer != "" {
		sys.HasGo = true
		sys.GoVersion = goVer
	}

	// Detect Python
	if pyVer, err := runCommand("python3", "--version"); err == nil && pyVer != "" {
		sys.HasPython = true
		sys.PythonVersion = pyVer
	} else if pyVer, err := runCommand("python", "--version"); err == nil && pyVer != "" {
		sys.HasPython = true
		sys.PythonVersion = pyVer
	}

	// Detect Node
	if nodeVer, err := runCommand("node", "--version"); err == nil && nodeVer != "" {
		sys.HasNode = true
		sys.NodeVersion = nodeVer
	}

	return sys, nil
}

func (s System) String() string {
	return fmt.Sprintf("OS: %s/%s | Go: %v (%s) | Python: %v (%s) | Node: %v (%s)",
		s.OS, s.Arch, s.HasGo, s.GoVersion, s.HasPython, s.PythonVersion, s.HasNode, s.NodeVersion)
}

func (s System) IsSupported() bool {
	return s.HasGo
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}