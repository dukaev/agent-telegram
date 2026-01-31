// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "agent-telegram",
	Short: "Telegram IPC agent CLI",
	Long: `agent-telegram is a CLI tool for interacting with Telegram via IPC server.

It provides commands to:
  - Start an IPC server with Telegram client
  - Interactively login to Telegram
  - Query chats, messages, and user info
  - Send and receive Telegram messages`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("socket", "s", "/tmp/agent-telegram.sock", "Path to Unix socket")
}
