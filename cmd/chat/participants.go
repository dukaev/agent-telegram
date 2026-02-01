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
	participantsPeer string
	participantsLimit int
)

// ParticipantsCmd represents the participants command.
var ParticipantsCmd = &cobra.Command{
	Use:     "participants",
	Short:   "List participants in a chat or channel",
	Long: `List all participants in a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --limit to set the maximum number of participants to return (max 200).

Example:
  agent-telegram participants --peer @mychannel --limit 50`,
	Args: cobra.NoArgs,
}

// AddParticipantsCommand adds the participants command to the root command.
func AddParticipantsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ParticipantsCmd)

	ParticipantsCmd.Flags().StringVarP(&participantsPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	ParticipantsCmd.Flags().IntVarP(&participantsLimit, "limit", "l", 100, "Maximum number of participants (max 200)")
	_ = ParticipantsCmd.MarkFlagRequired("peer")

	ParticipantsCmd.Run = func(_ *cobra.Command, _ []string) {
		// Validate and sanitize limit
		if participantsLimit < 1 {
			participantsLimit = 1
		}
		if participantsLimit > 200 {
			participantsLimit = 200
		}

		runner := cliutil.NewRunnerFromCmd(ParticipantsCmd, true)
		params := map[string]any{
			"peer":  participantsPeer,
			"limit": participantsLimit,
		}

		result := runner.CallWithParams("get_participants", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Also print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if count, ok := r["count"].(float64); ok {
				fmt.Fprintf(os.Stderr, "Found %d participant(s)\n", int(count))
			}
			if participants, ok := r["participants"].([]any); ok {
				for _, p := range participants {
					if participant, ok := p.(map[string]any); ok {
						name := "Unknown"
						if fn, ok := participant["firstName"].(string); ok {
							name = fn
							if ln, ok := participant["lastName"].(string); ok && ln != "" {
								name += " " + ln
							}
						}
						peer := "N/A"
						if pr, ok := participant["peer"].(string); ok {
							peer = pr
						}
						fmt.Fprintf(os.Stderr, "  - %s (%s)\n", name, peer)
					}
				}
			}
		}
	}
}
