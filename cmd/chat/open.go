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
	// OpenLimit is the number of messages to return.
	openLimit  int
	// OpenOffset is the offset for pagination.
	openOffset int
	// openJSON enables JSON output.
	openJSON bool
)

// OpenCmd represents the open command.
var OpenCmd = &cobra.Command{
	Use:     "open @username",
	Short:   "Open and view messages from a Telegram user/chat",
	Long: `Open and view messages from a Telegram user or chat by username.

Supports pagination with --limit and --offset flags.
Use --json flag for machine-readable output.`,
	Args: cobra.ExactArgs(1),
}

// AddOpenCommand adds the open command to the parent command.
func AddOpenCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(OpenCmd)

	SetupOpenFlags(OpenCmd)

	OpenCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(OpenCmd, true) // Always JSON

		username := args[0]

		// Validate and sanitize limit/offset
		limit := openLimit
		if limit < 1 {
			limit = 1
		}
		if limit > 100 {
			limit = 100
		}
		offset := openOffset
		if offset < 0 {
			offset = 0
		}

		result := runner.CallWithParams("get_messages", map[string]any{
			"username": username,
			"limit":    limit,
			"offset":   offset,
		})

		// Output as JSON
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	}
}

// SetupOpenFlags configures the flags for an open command.
func SetupOpenFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&openLimit, "limit", "l", 10, "Number of messages to return (max 100)")
	cmd.Flags().IntVarP(&openOffset, "offset", "o", 0, "Offset for pagination (message ID)")
	cmd.Flags().BoolVarP(&openJSON, "json", "j", false, "Output as JSON")
}

