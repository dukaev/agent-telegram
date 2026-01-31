// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	clearMessagesJSON bool
)

// clearMessagesCmd represents the clear-messages command.
var clearMessagesCmd = &cobra.Command{
	Use:   "clear-messages <id1,id2,...>",
	Short: "Clear specific Telegram messages",
	Long: `Delete specific messages by their IDs.

Message IDs can be obtained using the 'open' command.

Example: agent-telegram clear-messages 12345,12346,12347`,
	Args: cobra.ExactArgs(1),
	Run:  runClearMessages,
}

func init() {
	rootCmd.AddCommand(clearMessagesCmd)

	clearMessagesCmd.Flags().BoolVarP(&clearMessagesJSON, "json", "j", false, "Output as JSON")
}

func runClearMessages(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(clearMessagesJSON)

	idStrs := strings.Split(args[0], ",")
	messageIDs := make([]int64, 0, len(idStrs))

	for _, idStr := range idStrs {
		id := runner.MustParseInt64(strings.TrimSpace(idStr))
		messageIDs = append(messageIDs, id)
	}

	result := runner.CallWithParams("clear_messages", map[string]any{
		"messageIds": messageIDs,
	})
	runner.PrintResult(result, func(result any) {
		r, ok := result.(map[string]any)
		if !ok {
			fmt.Printf("Messages cleared successfully!\n")
			return
		}
		cleared := ExtractFloat64(r, "cleared")
		fmt.Printf("Messages cleared successfully!\n")
		fmt.Printf("  Cleared: %d messages\n", int(cleared))
	})
}
