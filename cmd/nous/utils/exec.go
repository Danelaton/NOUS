package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct {
	dryRun bool
}

func NewRunner(dryRun bool) *Runner {
	return &Runner{dryRun: dryRun}
}

func (r *Runner) Run(name string, args ...string) error {
	if r.dryRun {
		fmt.Printf("[DRY RUN] Would execute: %s %s\n", name, strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\noutput: %s", err, string(output))
	}
	return nil
}

func RunShellScript(script string) error {
	cmd := exec.Command("sh", "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shell script failed: %w\n%s", err, string(output))
	}
	return nil
}

func RunPowerShell(script string) error {
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell script failed: %w\n%s", err, string(output))
	}
	return nil
}