// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	bannedPeer string
	bannedLimit int
)

// BannedCmd represents the banned command.
var BannedCmd = &cobra.Command{
	Use:     "banned",
	Short:   "List banned users in a channel",
	Long: `List all banned users in a Telegram channel.

Use --peer @username or --peer username to specify the channel.
Use --limit to set the maximum number of banned users to return (max 200).

Example:
  agent-telegram banned --peer @mychannel --limit 20`,
	Args: cobra.NoArgs,
}

// AddBannedCommand adds the banned command to the root command.
func AddBannedCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BannedCmd)

	BannedCmd.Flags().StringVarP(&bannedPeer, "peer", "p", "", "Channel username (@username or username)")
	BannedCmd.Flags().IntVarP(&bannedLimit, "limit", "l", 100, "Maximum number of banned users (max 200)")
	_ = BannedCmd.MarkFlagRequired("peer")

	BannedCmd.Run = func(_ *cobra.Command, _ []string) {
		// Validate and sanitize limit
		if bannedLimit < 1 {
			bannedLimit = 1
		}
		if bannedLimit > 200 {
			bannedLimit = 200
		}

		runner := cliutil.NewRunnerFromCmd(BannedCmd, true)
		params := map[string]any{
			"peer":  bannedPeer,
			"limit": bannedLimit,
		}

		result := runner.CallWithParams("get_banned", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Also print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if count, ok := r["count"].(float64); ok {
				fmt.Fprintf(os.Stderr, "Found %d banned user(s)\n", int(count))
			}
			if banned, ok := r["banned"].([]any); ok {
				for _, b := range banned {
					if user, ok := b.(map[string]any); ok {
						peer := "N/A"
						if pr, ok := user["peer"].(string); ok {
							peer = pr
						}
						name := "Unknown"
						if fn, ok := user["firstName"].(string); ok {
							name = fn
							if ln, ok := user["lastName"].(string); ok && ln != "" {
								name += " " + ln
							}
						}
						fmt.Fprintf(os.Stderr, "  - %s (%s)\n", name, peer)
					}
				}
			}
		}
	}
}
