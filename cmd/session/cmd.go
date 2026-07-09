// Package session provides commands for managing Telegram sessions.
package session

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/sessionstore"
)

// SessionCmd represents the parent session command.
var SessionCmd = &cobra.Command{
	GroupID: "server",
	Use:     "session",
	Short:   "Manage Telegram session",
	Long:    `Inspect, import, export, select, and remove sessions through pluggable storage providers.`,
}

// AddSessionCommand adds the parent session command and all its subcommands to the root command.
func AddSessionCommand(rootCmd *cobra.Command) {
	AddExportCommand(SessionCmd)
	AddImportCommand(SessionCmd)
	AddStatusCommand(SessionCmd)
	AddProvidersCommand(SessionCmd)
	AddForgetCommand(SessionCmd)
	AddUseCommand(SessionCmd)
	rootCmd.AddCommand(SessionCmd)
}

func openSelectedStorage(cmd *cobra.Command) (*sessionstore.Storage, error) {
	provider := stringFlag(cmd, "session-provider")
	profile := stringFlag(cmd, "profile")
	if stored, err := config.LoadStoredConfig(); err == nil {
		if provider == "" {
			provider = stored.SessionProvider
		}
		if profile == "" {
			profile = stored.SessionProfile
		}
	}
	return sessionstore.Open(provider, profile)
}

func stringFlag(cmd *cobra.Command, name string) string {
	if cmd == nil {
		return ""
	}
	if value, err := cmd.Flags().GetString(name); err == nil {
		return value
	}
	if root := cmd.Root(); root != nil {
		value, _ := root.PersistentFlags().GetString(name)
		return value
	}
	return ""
}

func confirmed(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if value, err := cmd.Flags().GetBool("confirm"); err == nil {
		return value
	}
	if root := cmd.Root(); root != nil {
		value, _ := root.PersistentFlags().GetBool("confirm")
		return value
	}
	return false
}

func writeSessionJSON(cmd *cobra.Command, value any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func printSessionError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func sessionInputIsTerminal(cmd *cobra.Command) bool {
	file, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
