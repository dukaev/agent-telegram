// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inviteLinkTo        cliutil.Recipient
	inviteLinkCreateNew bool
)

// InviteLinkCmd represents the invite-link command.
var InviteLinkCmd = &cobra.Command{
	Use:     "invite-link",
	Short:   "Get or create an invite link for a chat or channel",
	Long: `Get an existing invite link or create a new one for a Telegram chat or channel.

Use --to @username or --to username to specify the chat/channel.
Use --create-new to create a new invite link instead of getting an existing one.

Example:
  agent-telegram chat invite-link --to @mychannel
  agent-telegram chat invite-link --to @mychannel --create-new`,
	Args: cobra.NoArgs,
}

// AddInviteLinkCommand adds the invite-link command to the root command.
func AddInviteLinkCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InviteLinkCmd)

	InviteLinkCmd.Flags().VarP(&inviteLinkTo, "to", "t", "Chat/channel (@username or username)")
	InviteLinkCmd.Flags().BoolVarP(&inviteLinkCreateNew, "create-new", "n", false, "Create a new invite link")
	_ = InviteLinkCmd.MarkFlagRequired("to")

	InviteLinkCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(InviteLinkCmd, true)
		params := map[string]any{
			"createNew": inviteLinkCreateNew,
		}
		inviteLinkTo.AddToParams(params)

		result := runner.CallWithParams("get_invite_link", params)
		runner.PrintResult(result, func(any) {
			cliutil.PrintInviteLinkSummary(result)
		})
	}
}
