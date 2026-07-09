package session

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/config"
	"agent-telegram/internal/sessionstore"
)

func AddUseCommand(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "use <provider> [profile]",
		Short: "Select the default session provider and profile",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := sessionstore.DefaultProfile
			if len(args) == 2 {
				profile = args[1]
			}
			storage, err := sessionstore.Open(args[0], profile)
			if err != nil {
				return err
			}
			if err := config.UpdateSessionSelection(storage.Provider(), storage.Profile()); err != nil {
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
