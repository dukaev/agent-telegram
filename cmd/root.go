// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/paths"
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
  - Authenticate to Telegram with headless JSON commands
  - Query chats, messages, and user info
  - Send and receive Telegram messages`,
	Version: version,
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
