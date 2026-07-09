package session

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// AddExportCommand adds the export subcommand to the session parent command.
func AddExportCommand(parentCmd *cobra.Command) {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export session as base64 for TELEGRAM_SESSION env",
		Long: `Export the selected Telegram session as base64.

The command reads from TELEGRAM_SESSION when it is set, otherwise from the
selected session provider/profile. Exporting grants full account access and
therefore requires --confirm.

Example:
  agent-telegram session export --confirm`,
		RunE: runExport,
	}

	parentCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, _ []string) error {
	if !confirmed(cmd) {
		return printSessionError("--confirm is required to export a Telegram session")
	}
	if sessionValue := os.Getenv("TELEGRAM_SESSION"); sessionValue != "" {
		_, err := fmt.Fprint(cmd.OutOrStdout(), sessionValue)
		return err
	}
	storage, err := openSelectedStorage(cmd)
	if err != nil {
		return err
	}
	data, err := storage.LoadSession(cmd.Context())
	if err != nil {
		return fmt.Errorf("load session from %s/%s: %w", storage.Provider(), storage.Profile(), err)
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), base64.StdEncoding.EncodeToString(data))
	return err
}
