// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	addReactionBig     bool
	addReactionPeer    string
	addReactionUsername string
)

// addReactionCmd represents the add-reaction command.
var addReactionCmd = &cobra.Command{
	Use:   "add-reaction <message_id> <emoji>",
	Short: "Add a reaction to a message",
	Long: `Add an emoji reaction to a message.

Use --big to send a big reaction.
Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
	Run:  runAddReaction,
}

func init() {
	addReactionCmd.Flags().BoolVar(&addReactionBig, "big", false, "Send as big reaction")
	addReactionCmd.Flags().StringVarP(&addReactionPeer, "peer", "p", "", "Peer (e.g., @username)")
	addReactionCmd.Flags().StringVarP(&addReactionUsername, "username", "u", "", "Username (without @)")
	addReactionCmd.MarkFlagsOneRequired("peer", "username")
	addReactionCmd.MarkFlagsMutuallyExclusive("peer", "username")
	rootCmd.AddCommand(addReactionCmd)
}

func runAddReaction(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
		"emoji":     args[1],
		"big":       addReactionBig,
	}
	if addReactionPeer != "" {
		params["peer"] = addReactionPeer
	} else {
		params["username"] = addReactionUsername
	}
	result := runner.CallWithParams("add_reaction", params)
	runner.PrintResult(result, nil)
}

var (
	removeReactionPeer     string
	removeReactionUsername string
)

// removeReactionCmd represents the remove-reaction command.
var removeReactionCmd = &cobra.Command{
	Use:   "remove-reaction <message_id>",
	Short: "Remove reactions from a message",
	Long: `Remove all reactions from a message (or your reaction).

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runRemoveReaction,
}

func init() {
	removeReactionCmd.Flags().StringVarP(&removeReactionPeer, "peer", "p", "", "Peer (e.g., @username)")
	removeReactionCmd.Flags().StringVarP(&removeReactionUsername, "username", "u", "", "Username (without @)")
	removeReactionCmd.MarkFlagsOneRequired("peer", "username")
	removeReactionCmd.MarkFlagsMutuallyExclusive("peer", "username")
	rootCmd.AddCommand(removeReactionCmd)
}

func runRemoveReaction(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
	}
	if removeReactionPeer != "" {
		params["peer"] = removeReactionPeer
	} else {
		params["username"] = removeReactionUsername
	}
	result := runner.CallWithParams("remove_reaction", params)
	runner.PrintResult(result, nil)
}

var (
	listReactionsPeer     string
	listReactionsUsername string
)

// listReactionsCmd represents the list-reactions command.
var listReactionsCmd = &cobra.Command{
	Use:   "list-reactions <message_id>",
	Short: "List reactions on a message",
	Long: `List all reactions on a specific message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runListReactions,
}

func init() {
	listReactionsCmd.Flags().StringVarP(&listReactionsPeer, "peer", "p", "", "Peer (e.g., @username)")
	listReactionsCmd.Flags().StringVarP(&listReactionsUsername, "username", "u", "", "Username (without @)")
	listReactionsCmd.MarkFlagsOneRequired("peer", "username")
	listReactionsCmd.MarkFlagsMutuallyExclusive("peer", "username")
	rootCmd.AddCommand(listReactionsCmd)
}

func runListReactions(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
	}
	if listReactionsPeer != "" {
		params["peer"] = listReactionsPeer
	} else {
		params["username"] = listReactionsUsername
	}
	result := runner.CallWithParams("list_reactions", params)

	r, ok := result.(map[string]any)
	if !ok {
		runner.PrintResult(result, nil)
		return
	}

	reactions, _ := r["reactions"].([]any)
	messageID := runner.MustParseInt64(args[0])

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
