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
	invitePeer  string
	inviteMembers []string
)

// InviteCmd represents the invite command.
var InviteCmd = &cobra.Command{
	Use:     "invite",
	Short:   "Invite users to a chat or channel",
	Long: `Invite users to a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --members to specify users to invite (can be specified multiple times).

Example:
  agent-telegram invite --peer @mychannel --members @user1 --members @user2`,
	Args: cobra.NoArgs,
}

// AddInviteCommand adds the invite command to the root command.
func AddInviteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InviteCmd)

	InviteCmd.Flags().StringVarP(&invitePeer, "peer", "p", "", "Chat/channel username (@username or username)")
	InviteCmd.Flags().StringSliceVarP(&inviteMembers, "members", "m", []string{}, "Users to invite (can be specified multiple times)")
	_ = InviteCmd.MarkFlagRequired("peer")
	_ = InviteCmd.MarkFlagRequired("members")

	InviteCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(InviteCmd, true)
		params := map[string]any{
			"peer":  invitePeer,
			"members": inviteMembers,
		}

		result := runner.CallWithParams("invite", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Members invited successfully\n")
			}
		}
	}
}
