package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// AddExportCommand adds the export subcommand to the session parent command.
func AddExportCommand(parentCmd *cobra.Command) {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export session as base64 for TELEGRAM_SESSION env",
		Long: `Print the TELEGRAM_SESSION environment variable when it is already set.

agent-telegram now keeps sessions in memory and no longer writes session.json.
Use auth web while the server is running, or provide TELEGRAM_SESSION explicitly
for deployments where the session must be injected from an external secret store.

Example:
  agent-telegram session export`,
		Run: runExport,
	}

	parentCmd.AddCommand(exportCmd)
}

func runExport(_ *cobra.Command, _ []string) {
	if sessionValue := os.Getenv("TELEGRAM_SESSION"); sessionValue != "" {
		fmt.Print(sessionValue)
		return
	}
	fmt.Fprintln(os.Stderr, "Error: session.json is no longer used; TELEGRAM_SESSION is not set")
	os.Exit(1)
}
