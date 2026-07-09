package session

import (
	"errors"

	"github.com/gotd/td/session"
	"github.com/spf13/cobra"
)

func AddStatusCommand(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show the selected session storage status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			storage, err := openSelectedStorage(cmd)
			if err != nil {
				return err
			}
			stored := true
			if _, err := storage.LoadSession(cmd.Context()); err != nil {
				if !errors.Is(err, session.ErrNotFound) {
					return err
				}
				stored = false
			}
			return writeSessionJSON(cmd, map[string]any{
				"ok":         true,
				"provider":   storage.Provider(),
				"profile":    storage.Profile(),
				"persistent": storage.Persistent(),
				"stored":     stored,
			})
		},
	})
}
