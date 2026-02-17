package session

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var exportSessionPath string

// AddExportCommand adds the export subcommand to the session parent command.
func AddExportCommand(parentCmd *cobra.Command) {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export session as base64 for TELEGRAM_SESSION env",
		Long: `Read the Telegram session file and output it as a base64-encoded string.

Use this to set the TELEGRAM_SESSION environment variable for Docker/Coolify/Heroku deployments
where persistent filesystem is not available.

Example:
  # Export and set in one step
  export TELEGRAM_SESSION=$(agent-telegram session export)

  # Or save to a file
  agent-telegram session export > session.b64`,
		Run: runExport,
	}

	exportCmd.Flags().StringVar(&exportSessionPath, "session", "",
		"Path to session file (default: ~/.agent-telegram/session.json)")

	parentCmd.AddCommand(exportCmd)
}

func runExport(_ *cobra.Command, _ []string) {
	sessionPath := exportSessionPath

	// Check env override
	if sessionPath == "" {
		sessionPath = os.Getenv("AGENT_TELEGRAM_SESSION_PATH")
	}

	// Default path
	if sessionPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
			os.Exit(1)
		}
		sessionPath = filepath.Join(home, ".agent-telegram", "session.json")
	}

	//nolint:gosec // sessionPath is from trusted flag/env, not arbitrary user input
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read session file %s: %v\n", sessionPath, err)
		os.Exit(1)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	fmt.Print(encoded)
}
