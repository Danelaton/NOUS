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
	Short: "NOUS — AI Skills Installer for coding agents",
	Long: `NOUS installs AI agent skills into your projects.

  nous sync     # setup project: dev/ + .agent/ + AGENTS.md + skills
  nous skills   # install/update skills from ~/.nous/skills/
  nous status   # show installed skills and runtime info

Skills install globally to ~/.nous/skills/ — never locked to a single project.`,
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
	Short: "Show NOUS status and installed skills",
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
and syncs skills from ~/.nous/skills/ into .agent/skills/.

Creates:
  dev/sandbox/  dev/tmp-repos/  dev/docs/
  dev/scripts/   dev/tests/      dev/backups/
  .agent/MEMORY.md  .agent/docs_index.md  .agent/skills/
  .gitignore (adds dev/ and .agent/ if missing)

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
		fmt.Printf("[NOUS] Project ready: %s\n", projectDir)
		return nil
	},
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Sync skills from ~/.nous/skills/ into .agent/skills/ (merge, no delete)",
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
		return orch.SyncSkills(projectDir)
	},
}

func init() {
	rootCmd.AddCommand(installCmd, statusCmd, syncCmd, skillsCmd)
	syncCmd.Flags().StringP("dir", "d", "", "Project directory (default: current directory)")
	skillsCmd.Flags().StringP("dir", "d", "", "Project directory (default: current directory)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}