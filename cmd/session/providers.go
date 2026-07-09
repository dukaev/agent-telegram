package session

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/sessionstore"
)

func AddProvidersCommand(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "providers",
		Short: "List available session storage providers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeSessionJSON(cmd, map[string]any{
				"default":   sessionstore.DefaultProvider(),
				"providers": sessionstore.Providers(),
			})
		},
	})
}
