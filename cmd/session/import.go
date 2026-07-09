package session

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func AddImportCommand(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "import",
		Short: "Import a base64 Telegram session from stdin",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirmed(cmd) {
				return printSessionError("--confirm is required to replace a Telegram session")
			}
			if sessionInputIsTerminal(cmd) {
				return fmt.Errorf("session must be piped on stdin")
			}
			raw, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
			if err != nil {
				return fmt.Errorf("session must be base64: %w", err)
			}
			if len(data) == 0 {
				return fmt.Errorf("session is empty")
			}
			storage, err := openSelectedStorage(cmd)
			if err != nil {
				return err
			}
			if err := storage.StoreSession(cmd.Context(), data); err != nil {
				return err
			}
			return writeSessionJSON(cmd, map[string]any{
				"ok":         true,
				"provider":   storage.Provider(),
				"profile":    storage.Profile(),
				"persistent": storage.Persistent(),
			})
		},
	})
}
