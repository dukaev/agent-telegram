// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/paths"
	"agent-telegram/internal/skills"
)

var (
	version = "dev"

	// GroupIDAuth is the command group ID for authentication commands.
	GroupIDAuth = "auth"
	// GroupIDMessage is the command group ID for message management commands.
	GroupIDMessage = "message"
	// GroupIDChat is the command group ID for chat commands.
	GroupIDChat = "chat"
	// GroupIDServer is the command group ID for server commands.
	GroupIDServer = "server"
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "agent-telegram",
	Short: "Telegram IPC agent CLI",
	Long: `agent-telegram is a CLI tool for interacting with Telegram via IPC server.

It provides commands to:
  - Start an IPC server with Telegram client
  - Authenticate to Telegram through a local browser page
  - Query chats, messages, and user info
  - Send and receive Telegram messages`,
	Version: version,
	Args:    cobra.NoArgs,
	RunE:    runRoot,
}

const rootWelcome = `Connect Telegram

  agent-telegram auth
      Sign in with a QR code in your browser

After sign-in

  agent-telegram server ensure    Start the local Telegram server
  agent-telegram chats            List your chats
  agent-telegram send @user hi    Send a message

Run agent-telegram --help to see every command.
`

const bundledSkillName = "agent-telegram"

type onboardingServices struct {
	isTerminal      func(any) bool
	getwd           func() (string, error)
	resolve         func(string, string) (skills.OnboardingDecision, error)
	install         func(string, string, bool) (string, error)
	dismissed       func(string) (bool, error)
	recordDismissal func(string) error
	ci              string
}

func defaultOnboardingServices() onboardingServices {
	return onboardingServices{
		isTerminal:      isTerminal,
		getwd:           os.Getwd,
		resolve:         skills.ResolveOnboarding,
		install:         skills.Install,
		dismissed:       skills.GlobalPromptDismissed,
		recordDismissal: skills.RecordGlobalPromptDismissal,
		ci:              os.Getenv("CI"),
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	return runRootWithServices(cmd, args, defaultOnboardingServices())
}

func runRootWithServices(cmd *cobra.Command, _ []string, services onboardingServices) error {
	if _, err := fmt.Fprint(cmd.OutOrStdout(), rootWelcome); err != nil {
		return err
	}
	maybeOnboardSkill(cmd, services)
	return nil
}

func maybeOnboardSkill(cmd *cobra.Command, services onboardingServices) {
	if services.ci != "" || cmd.Flags().NFlag() != 0 ||
		!services.isTerminal(cmd.InOrStdin()) || !services.isTerminal(cmd.OutOrStdout()) {
		return
	}
	cwd, err := services.getwd()
	if err != nil {
		return
	}
	decision, err := services.resolve(bundledSkillName, cwd)
	if err != nil {
		writeOnboardingMessage(cmd, "Warning: could not inspect Codex skill locations: %v\n", err)
		return
	}
	switch decision.Action {
	case skills.OnboardingNone:
		return
	case skills.OnboardingInstallProject:
		installed, err := services.install(bundledSkillName, decision.InstallDir, false)
		if err != nil {
			writeOnboardingMessage(cmd, "Warning: could not install project Codex skill: %v\n", err)
			return
		}
		writeOnboardingMessage(cmd, "Installed project Codex skill at %s\n", installed)
	case skills.OnboardingPromptGlobal:
		promptGlobalSkill(cmd, services, decision)
	}
}

func promptGlobalSkill(cmd *cobra.Command, services onboardingServices, decision skills.OnboardingDecision) {
	dismissed, err := services.dismissed(decision.Target)
	if err != nil {
		writeOnboardingMessage(cmd, "Warning: could not read global skill preference: %v\n", err)
		return
	}
	if dismissed {
		return
	}
	writeOnboardingMessage(
		cmd,
		"\nInstall the Agent Telegram skill globally for Codex?\nTarget: %s\nThis makes the skill available across projects. [y/N] ",
		decision.Target,
	)
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			writeOnboardingMessage(cmd, "\nWarning: could not read skill prompt response: %v\n", err)
		}
		return
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		installed, err := services.install(bundledSkillName, decision.InstallDir, false)
		if err != nil {
			writeOnboardingMessage(
				cmd,
				"Warning: could not install global Codex skill: %v\nInstall it later with: agent-telegram skills install agent-telegram\n",
				err,
			)
			return
		}
		writeOnboardingMessage(cmd, "Installed global Codex skill at %s\n", installed)
	default:
		if err := services.recordDismissal(decision.Target); err != nil {
			writeOnboardingMessage(cmd, "Warning: could not save global skill preference: %v\n", err)
		}
	}
}

func isTerminal(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func writeOnboardingMessage(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), format, args...)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	// Handle --schema before cobra's Execute to bypass required flag/arg validation.
	if hasFlag(os.Args[1:], "--schema") {
		cmd, _, _ := RootCmd.Find(os.Args[1:])
		if cmd != nil && cmd != RootCmd {
			cliutil.PrintCommandSchema(cmd)
		}
	}

	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// hasFlag checks if a flag is present in args.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if a == flag {
			return true
		}
	}
	return false
}

func init() {
	// Add command groups
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDAuth, Title: "Authentication"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessage, Title: "Manage Messages"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDChat, Title: "Chat"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDServer, Title: "Server"})

	// Global flags
	RootCmd.PersistentFlags().StringP("socket", "s", paths.DefaultSocketPath, "Path to Unix socket")
	RootCmd.PersistentFlags().String("session-provider", os.Getenv("AGENT_TELEGRAM_SESSION_PROVIDER"), "Session provider (default: native platform provider)")
	RootCmd.PersistentFlags().String("profile", os.Getenv("AGENT_TELEGRAM_PROFILE"), "Session profile (default: default)")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress status messages (data still goes to stdout)")
	RootCmd.PersistentFlags().String("output", "", "Output format: json or ids")
	RootCmd.PersistentFlags().StringSlice("filter", nil, "Filter results (e.g., 'stars>1000', 'type=channel')")
	RootCmd.PersistentFlags().String("verbosity", "full", "Output detail: minimal, compact, full, raw")
	RootCmd.PersistentFlags().Int("max-items", 0, "Maximum array items in JSON output (0 uses verbosity default)")
	RootCmd.PersistentFlags().Int("max-text-chars", 0, "Maximum text field characters in JSON output (0 uses verbosity default)")
	RootCmd.PersistentFlags().StringSlice("include", nil, "Include output fields (comma-separated, supports dot paths)")
	RootCmd.PersistentFlags().StringSlice("omit", nil, "Omit output fields (comma-separated, supports dot paths)")
	RootCmd.PersistentFlags().Bool("summary", false, "Output a compact result summary")
	RootCmd.PersistentFlags().Bool("receipt", false, "Wrap JSON output with trace/action receipt metadata")
	RootCmd.PersistentFlags().Bool("dry-run", false, "Preview action without executing")
	RootCmd.PersistentFlags().Bool("confirm", false, "Confirm a destructive or paid operation")
	RootCmd.PersistentFlags().Bool("schema", false, "Output operation schema without executing")
	RootCmd.PersistentFlags().Bool("agent", false, "Enable agent-friendly compact JSON, receipts, run IDs, and structured errors")
	RootCmd.PersistentFlags().String("run-id", "", "Agent run ID for correlating multiple commands")
}
