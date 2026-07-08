package sys

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/skills"
)

var (
	skillInstallTarget string
	skillInstallForce  bool
)

// SkillsCmd groups bundled agent skill helpers.
var SkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List and install bundled agent skills",
	Long: `List and install agent skills bundled with agent-telegram.

Skills help AI agents discover best practices for using the CLI. Install them
into CODEX_HOME/skills or ~/.codex/skills so Codex can load them directly.`,
}

// SkillsListCmd lists bundled skills.
var SkillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bundled agent skills",
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, true)
		runner.SetAction("skills_list")
		runner.PrintJSON(map[string]any{
			"ok":                true,
			"defaultInstallDir": skills.DefaultInstallDir(),
			"skills":            skills.Manifest(),
		})
	},
}

// SkillsPathCmd prints the default skill install directory.
var SkillsPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show the default agent skill install directory",
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, true)
		runner.SetAction("skills_path")
		runner.PrintJSON(map[string]any{
			"ok":                true,
			"defaultInstallDir": skills.DefaultInstallDir(),
		})
	},
}

// SkillsInstallCmd installs a bundled skill into the Codex skill directory.
var SkillsInstallCmd = &cobra.Command{
	Use:   "install <skill>",
	Short: "Install a bundled agent skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, true)
		runner.SetAction("skills_install")
		installedPath, err := skills.Install(args[0], skillInstallTarget, skillInstallForce)
		if err != nil {
			runner.Fatal(err.Error())
		}
		runner.PrintJSON(map[string]any{
			"ok":     true,
			"name":   args[0],
			"path":   installedPath,
			"forced": skillInstallForce,
		})
	},
}

// AddSkillsCommand adds bundled skill commands.
func AddSkillsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SkillsCmd)
	SkillsCmd.AddCommand(SkillsListCmd, SkillsPathCmd, SkillsInstallCmd)
	SkillsInstallCmd.Flags().StringVar(&skillInstallTarget, "target", "", "Skill install directory (default: CODEX_HOME/skills or ~/.codex/skills)")
	SkillsInstallCmd.Flags().BoolVar(&skillInstallForce, "force", false, "Overwrite an existing installed skill")
}
