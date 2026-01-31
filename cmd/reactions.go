// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addReactionCmd represents the add-reaction command.
var addReactionCmd = &cobra.Command{
	Use:   "add-reaction @peer <message_id> <emoji>",
	Short: "Add a reaction to a message",
	Long: `Add an emoji reaction to a message.

Use --big to send a big reaction.

Example: agent-telegram add-reaction @user 123456 👍`,
	Args: cobra.ExactArgs(3),
	Run:  runAddReaction,
}

func init() {
	addReactionCmd.Flags().BoolVar(&addReactionBig, "big", false, "Send as big reaction")
	rootCmd.AddCommand(addReactionCmd)
}

var addReactionBig bool

func runAddReaction(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("add_reaction", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
		"emoji":     args[2],
		"big":       addReactionBig,
	})
	runner.PrintResult(result, nil)
}

// removeReactionCmd represents the remove-reaction command.
var removeReactionCmd = &cobra.Command{
	Use:   "remove-reaction @peer <message_id>",
	Short: "Remove reactions from a message",
	Long: `Remove all reactions from a message (or your reaction).

Example: agent-telegram remove-reaction @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runRemoveReaction,
}

func init() {
	rootCmd.AddCommand(removeReactionCmd)
}

func runRemoveReaction(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("remove_reaction", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
	})
	runner.PrintResult(result, nil)
}

// listReactionsCmd represents the list-reactions command.
var listReactionsCmd = &cobra.Command{
	Use:   "list-reactions @peer <message_id>",
	Short: "List reactions on a message",
	Long: `List all reactions on a specific message.

Example: agent-telegram list-reactions @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runListReactions,
}

func init() {
	rootCmd.AddCommand(listReactionsCmd)
}

func runListReactions(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("list_reactions", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
	})

	r, ok := result.(map[string]any)
	if !ok {
		runner.PrintResult(result, nil)
		return
	}

	reactions, _ := r["reactions"].([]any)
	messageID := runner.MustParseInt64(args[1])

	fmt.Printf("Reactions on message %d:\n", messageID)
	for _, rx := range reactions {
		r, _ := rx.(map[string]any)
		emoji, _ := r["emoji"].(string)
		count, _ := r["count"].(float64)
		fromMe, _ := r["fromMe"].(bool)
		mark := ""
		if fromMe {
			mark = " (you)"
		}
		fmt.Printf("  %s %d%s\n", emoji, int(count), mark)
	}
}
