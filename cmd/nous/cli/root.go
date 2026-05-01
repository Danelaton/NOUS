package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Danelaton/NOUS/cmd/nous/install"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nous",
	Short: "NOUS — AI Agent Ecosystem Configurator",
	Long: `NOUS configures and enhances AI coding agents with:
  - Project structure and skills (dev/, AGENTS.md)
  - Automatic agent configuration injection

Runtime installs globally to ~/.nous/ — never inside your projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install NOUS runtime globally to ~/.nous/",
	RunE: func(cmd *cobra.Command, args []string) error {
		orch, err := install.NewOrchestrator()
		if err != nil {
			return err
		}
		return orch.Run()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show NOUS status and detected agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		sys, err := install.Detect()
		if err != nil {
			return err
		}
		orch, err := install.NewOrchestrator()
		if err != nil {
			return err
		}
		home, _ := os.UserHomeDir()
		_, nousInstalled := os.Stat(filepath.Join(home, ".nous"))
		fmt.Println("=== NOUS Status ===")
		fmt.Printf("System:  %s\n", sys.String())
		fmt.Printf("Supported: %v\n", sys.IsSupported())
		fmt.Println()
		if os.IsNotExist(nousInstalled) {
			fmt.Println("Runtime: not installed — run 'nous install'")
		} else {
			fmt.Println("Config:")
			fmt.Print(orch.Status())
		}
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync skills and setup project structure",
	Long: `Creates dev/ directory structure, copies AGENTS.md into the project,
and re-injects agent configurations.

Creates:
  dev/sandbox/  dev/tmp-repos/  dev/docs/
  dev/scripts/   dev/tests/      dev/backups/
  .gitignore (adds dev/ if missing)

Backs up existing AGENTS.md to dev/backups/ if one already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDir, _ := cmd.Flags().GetString("dir")
		if projectDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectDir = cwd
		}

		orch, err := install.NewOrchestrator()
		if err != nil {
			return err
		}
		if err := orch.SetupProject(projectDir); err != nil {
			return err
		}
		if err := orch.SyncAgents(); err != nil {
			fmt.Printf("[NOUS] Warning: agent sync: %v\n", err)
		}
		fmt.Printf("[NOUS] Project ready: %s\n", projectDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd, statusCmd, syncCmd)
	syncCmd.Flags().StringP("dir", "d", "", "Project directory (default: current directory)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}