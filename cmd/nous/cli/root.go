package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Danelaton/NOUS/cmd/nous/install"
	"github.com/Danelaton/NOUS/pkg/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nous",
	Short: "NOUS — AI Skills Installer for coding agents",
	Long: `NOUS installs AI agent skills into your projects.

  nous sync     # setup project: dev/ + .agents/OKF/ + AGENTS.md + skills
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
	Use:   "sync [agent]",
	Short: "Sync skills and setup project structure",
	Long: `Creates dev/ directory structure, copies AGENTS.md into the project,
and syncs skills from ~/.nous/skills/ into .agents/skills/.

With an agent name (e.g. 'hermes'), injects configuration globally
into that agent instead of setting up a project.

Supported agents: hermes, claude, cursor, kiro, roo, opencode

Creates (project mode):
  dev/sandbox/  dev/tmp-repos/  dev/docs/
  dev/scripts/   dev/tests/      dev/backups/
  .agents/MEMORY.md  .agents/OKF/  .agents/skills/
  .gitignore (adds dev/ and .agents/ if missing)

Backs up existing AGENTS.md to dev/backups/ if one already exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If an agent name is provided, inject config for that agent
		if len(args) > 0 {
			agentName := args[0]
			home, _ := os.UserHomeDir()
			nousDir := filepath.Join(home, ".nous")

			for _, adapter := range config.GetAllAdapters() {
				if adapter.AgentName() == agentName {
					if !adapter.Detect() {
						return fmt.Errorf("%s not detected on this system", agentName)
					}
					fmt.Printf("[NOUS] Injecting %s configuration...\n", agentName)
					return adapter.Inject(nousDir)
				}
			}
			return fmt.Errorf("unknown agent: %s — supported: hermes, claude, cursor, kiro, roo, opencode", agentName)
		}

		// Default: project sync
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
	Short: "Sync skills from ~/.nous/skills/ into .agents/skills/ (merge, no delete)",
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
