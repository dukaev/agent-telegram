// Package message provides commands for managing message reactions.
package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	reactionTo      cliutil.Recipient
	reactionMessage int64
	reactionEmoji   string
	reactionBig     bool
)

// ReactionCmd represents the reaction command.
var ReactionCmd = &cobra.Command{
	GroupID: "message",
	Use:   "reaction <emoji>",
	Short: "Add a reaction to a Telegram message",
	Long: `Add a reaction emoji to a message in a chat.

Use --big to send a big reaction.
Use --to @username, --to username, or --to <chat_id> to specify the recipient.
Use --message to specify the message ID to react to.`,
	Args: cobra.ExactArgs(1),
}

// AddReactionCommand adds the reaction command to the root command.
func AddReactionCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ReactionCmd)

	ReactionCmd.Flags().VarP(&reactionTo, "to", "t", "Recipient (@username, username, or chat ID)")
	ReactionCmd.Flags().Int64VarP(&reactionMessage, "message", "m", 0, "Message ID to react to")
	ReactionCmd.Flags().BoolVar(&reactionBig, "big", false, "Send a big reaction")
	_ = ReactionCmd.MarkFlagRequired("to")
	_ = ReactionCmd.MarkFlagRequired("message")

	ReactionCmd.Run = func(_ *cobra.Command, args []string) {
		reactionEmoji = args[0]
		runner := cliutil.NewRunnerFromCmd(ReactionCmd, false)
		params := map[string]any{
			"messageId": reactionMessage,
			"emoji":     reactionEmoji,
		}
		if reactionBig {
			params["big"] = true
		}
		reactionTo.AddToParams(params)

		result := runner.CallWithParams("add_reaction", params)
		runner.PrintResult(result, func(any) {
			fmt.Fprintf(os.Stderr, "Reaction added successfully!\n")
		})
	}
}
