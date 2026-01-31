// Package chat provides commands for managing chats.
package chat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	clearMessagesJSON bool
)

// ClearCmd represents the clear-messages command.
var ClearCmd = &cobra.Command{
	Use:   "clear <id1,id2,...>",
	Short: "Clear specific Telegram messages",
	Long: `Delete specific messages by their IDs.

Message IDs can be obtained using the 'open' command.

Example: agent-telegram chat clear 12345,12346,12347`,
	Args: cobra.ExactArgs(1),
}

// AddClearCommand adds the clear command to the root command.
func AddClearCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ClearCmd)

	ClearCmd.Flags().BoolVarP(&clearMessagesJSON, "json", "j", false, "Output as JSON")

	ClearCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ClearCmd, clearMessagesJSON)

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
			cleared := cliutil.ExtractFloat64(r, "cleared")
			fmt.Printf("Messages cleared successfully!\n")
			fmt.Printf("  Cleared: %d messages\n", int(cleared))
		})
	}
}
