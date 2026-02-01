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
	inviteLinkPeer string
	inviteLinkCreateNew bool
)

// InviteLinkCmd represents the invite-link command.
var InviteLinkCmd = &cobra.Command{
	Use:     "invite-link",
	Short:   "Get or create an invite link for a chat or channel",
	Long: `Get an existing invite link or create a new one for a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --create-new to create a new invite link instead of getting an existing one.

Example:
  agent-telegram invite-link --peer @mychannel
  agent-telegram invite-link --peer @mychannel --create-new`,
	Args: cobra.NoArgs,
}

// AddInviteLinkCommand adds the invite-link command to the root command.
func AddInviteLinkCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InviteLinkCmd)

	InviteLinkCmd.Flags().StringVarP(&inviteLinkPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	InviteLinkCmd.Flags().BoolVarP(&inviteLinkCreateNew, "create-new", "n", false, "Create a new invite link")
	_ = InviteLinkCmd.MarkFlagRequired("peer")

	InviteLinkCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(InviteLinkCmd, true)
		params := map[string]any{
			"peer":       inviteLinkPeer,
			"createNew":  inviteLinkCreateNew,
		}

		result := runner.CallWithParams("get_invite_link", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if link, ok := r["link"].(string); ok {
				fmt.Fprintf(os.Stderr, "Invite link: %s\n", link)
				if usage, ok := r["usage"].(float64); ok {
					fmt.Fprintf(os.Stderr, "Usage: %d\n", int(usage))
				}
				if limit, ok := r["usageLimit"].(float64); ok {
					fmt.Fprintf(os.Stderr, "Usage limit: %d\n", int(limit))
				}
			}
		}
	}
}
