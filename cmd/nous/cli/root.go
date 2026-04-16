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
  - Spec-Driven Development workflow (OpenSpec)
  - Project structure and skills (dev/, AGENTS.md)
  - Automatic agent configuration injection

Runtime installs globally to ~/.nous/ — never inside your projects.
Run 'nous sdd-init' inside any project to activate the SDD workflow there.`,
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

var sddInitCmd = &cobra.Command{
	Use:   "sdd-init",
	Short: "Initialize SDD workflow in the current project (creates openspec/)",
	Long: `Creates openspec/specs/ and openspec/changes/ in the target directory.

By default uses the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDir, _ := cmd.Flags().GetString("dir")
		if projectDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectDir = cwd
		}
		gen := install.NewOpenSpecGenerator(projectDir)
		if err := gen.GenerateStructure(); err != nil {
			return err
		}
		fmt.Printf("[NOUS] SDD context initialized at %s\n", projectDir)
		fmt.Printf("[NOUS]   openspec/specs/   — write your spec here before coding\n")
		fmt.Printf("[NOUS]   openspec/changes/ — propose changes here\n")
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
			return err
		}
		fmt.Printf("[NOUS] Project ready: %s\n", projectDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd, statusCmd, sddInitCmd, syncCmd)
	sddInitCmd.Flags().StringP("dir", "d", "", "Project directory (default: current directory)")
	syncCmd.Flags().StringP("dir", "d", "", "Project directory (default: current directory)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
