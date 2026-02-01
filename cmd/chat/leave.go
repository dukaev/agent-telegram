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
	leavePeer  string
)

// LeaveCmd represents the leave command.
var LeaveCmd = &cobra.Command{
	Use:     "leave",
	Short:   "Leave a chat or channel",
	Long: `Leave a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.

Example:
  agent-telegram leave --peer @mychannel`,
	Args: cobra.NoArgs,
}

// AddLeaveCommand adds the leave command to the root command.
func AddLeaveCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LeaveCmd)

	LeaveCmd.Flags().StringVarP(&leavePeer, "peer", "p", "", "Chat/channel username (@username or username)")
	_ = LeaveCmd.MarkFlagRequired("peer")

	LeaveCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(LeaveCmd, true)
		params := map[string]any{
			"peer":  leavePeer,
		}

		result := runner.CallWithParams("leave", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Left chat successfully\n")
			}
		}
	}
}
