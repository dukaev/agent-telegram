// Package get provides commands for retrieving information.
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetUpdatesLimit is the number of updates to get.
	GetUpdatesLimit int
	// GetUpdatesJSON enables JSON output.
	GetUpdatesJSON bool
	// GetUpdatesTo filters updates by recipient.
	GetUpdatesTo cliutil.Recipient
	// GetUpdatesFollow enables continuous polling mode.
	GetUpdatesFollow bool
	// GetUpdatesType filters updates by type.
	GetUpdatesType string
	// GetUpdatesInterval is the polling interval in seconds.
	GetUpdatesInterval int
)

// UpdatesCmd represents the get-updates command.
var UpdatesCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "updates",
	Short:   "Get Telegram updates (pops from store)",
	Long: `Retrieve Telegram updates from the update store. This removes them from the store.

Use --follow to continuously poll for updates (JSON Lines output).
Use --type to filter by update type (e.g., new_message, edit_message).`,
}

// AddUpdatesCommand adds the updates command to the root command.
func AddUpdatesCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UpdatesCmd)

	UpdatesCmd.Flags().IntVarP(&GetUpdatesLimit, "limit", "l", cliutil.DefaultLimitSmall, "Number of updates (max 100)")
	UpdatesCmd.Flags().BoolVarP(&GetUpdatesJSON, "json", "j", false, "Output as JSON")
	UpdatesCmd.Flags().VarP(&GetUpdatesTo, "to", "t", "Recipient (@username, username, or chat ID) to filter updates")
	UpdatesCmd.Flags().BoolVarP(&GetUpdatesFollow, "follow", "f", false, "Continuously poll for updates (JSON Lines)")
	UpdatesCmd.Flags().StringVar(&GetUpdatesType, "type", "", "Filter by update type (e.g., new_message, edit_message)")
	UpdatesCmd.Flags().IntVar(&GetUpdatesInterval, "interval", 2, "Polling interval in seconds (with --follow)")
	UpdatesCmd.Run = func(*cobra.Command, []string) {
		pag := cliutil.NewPagination(GetUpdatesLimit, 0, cliutil.PaginationConfig{
			MaxLimit: cliutil.MaxLimitStandard,
		})

		runner := cliutil.NewRunnerFromCmd(UpdatesCmd, true) // Always JSON
		params := map[string]any{}
		pag.ToParams(params, false) // updates API doesn't support offset
		GetUpdatesTo.AddToParams(params)

		if GetUpdatesFollow {
			runFollowMode(runner, params)
			return
		}

		result := runner.CallWithParams("get_updates", params)
		runner.PrintResult(result, nil)
	}
}

// runFollowMode continuously polls for updates and outputs JSON Lines to stdout.
func runFollowMode(runner *cliutil.Runner, params map[string]any) {
	interval := time.Duration(GetUpdatesInterval) * time.Second
	if interval < time.Second {
		interval = time.Second
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	encoder := json.NewEncoder(os.Stdout)
	for {
		select {
		case <-sigCh:
			return
		default:
		}

		result := runner.CallWithParams("get_updates", params)

		// Extract updates from result
		updates := extractUpdates(result)
		for _, update := range updates {
			// Apply type filter
			if GetUpdatesType != "" {
				if t, ok := update["type"].(string); !ok || t != GetUpdatesType {
					continue
				}
			}
			//nolint:errchkjson // JSON Lines output to stdout
			_ = encoder.Encode(update)
		}

		time.Sleep(interval)
	}
}

// extractUpdates extracts update items from the result.
func extractUpdates(result any) []map[string]any {
	// Result might be a slice or a map with "updates" key
	switch r := result.(type) {
	case []any:
		updates := make([]map[string]any, 0, len(r))
		for _, item := range r {
			if m, ok := item.(map[string]any); ok {
				updates = append(updates, m)
			}
		}
		return updates
	case map[string]any:
		if items, ok := r["updates"].([]any); ok {
			updates := make([]map[string]any, 0, len(items))
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					updates = append(updates, m)
				}
			}
			return updates
		}
	}
	return nil
}

// FormatUpdateType returns a human-readable string for an update type.
func FormatUpdateType(updateType string) string {
	switch updateType {
	case "new_message":
		return "New Message"
	case "edit_message":
		return "Edited Message"
	case "star_gift":
		return "Star Gift"
	default:
		return fmt.Sprintf("Update (%s)", updateType)
	}
}
