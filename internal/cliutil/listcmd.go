package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ListPrintFunc is the signature for list print functions.
type ListPrintFunc func(result any, unknownName, naValue string)

// ListCommandConfig defines configuration for list commands.
type ListCommandConfig struct {
	Use       string
	Short     string
	Long      string
	Method    string
	PrintFunc ListPrintFunc
	MaxLimit  int
	HasOffset bool // Whether to include --offset flag
}

// NewListCommand creates a cobra command for listing items.
func NewListCommand(cfg ListCommandConfig) *cobra.Command {
	var peer string
	var limit int
	var offset int

	maxLimit := cfg.MaxLimit
	if maxLimit == 0 {
		maxLimit = MaxLimitParticipants
	}

	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			pag := NewPagination(limit, offset, PaginationConfig{
				MaxLimit: maxLimit,
			})

			runner := NewRunnerFromCmd(cmd, true)
			params := map[string]any{
				"peer": peer,
			}
			pag.ToParams(params, cfg.HasOffset)

			result := runner.CallWithParams(cfg.Method, params)
			//nolint:errchkjson // Output to stdout, error handling not required
			_ = json.NewEncoder(os.Stdout).Encode(result)

			// Print human-readable summary
			if cfg.PrintFunc != nil {
				cfg.PrintFunc(result, "Unknown", "N/A")
			}
		},
	}

	cmd.Flags().StringVarP(&peer, "peer", "p", "", "Chat/channel username (@username or username)")
	cmd.Flags().IntVarP(&limit, "limit", "l", DefaultLimitMax, fmt.Sprintf("Maximum number of items (max %d)", maxLimit))
	if cfg.HasOffset {
		cmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	}
	_ = cmd.MarkFlagRequired("peer")

	return cmd
}
