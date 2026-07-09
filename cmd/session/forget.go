package session

import (
	"github.com/spf13/cobra"
)

func AddForgetCommand(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "forget",
		Short: "Delete the selected local session without Telegram logout",
		Long:  "Delete the selected local session. Stop the daemon first; this does not revoke the session in Telegram.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirmed(cmd) {
				return printSessionError("--confirm is required to forget a Telegram session")
			}
			storage, err := openSelectedStorage(cmd)
			if err != nil {
				return err
			}
			if err := storage.Delete(cmd.Context()); err != nil {
				return err
			}
			return writeSessionJSON(cmd, map[string]any{
				"ok":       true,
				"provider": storage.Provider(),
				"profile":  storage.Profile(),
				"warning":  "Local credentials were deleted without revoking the Telegram session.",
			})
		},
	})
}
