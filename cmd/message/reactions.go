// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	addReactionBig        bool
	addReactionPeer       string
	addReactionUsername   string
)

// AddReactionCmd represents the add-reaction command.
var AddReactionCmd = &cobra.Command{
	Use:   "add-reaction <message_id> <emoji>",
	Short: "Add a reaction to a message",
	Long: `Add an emoji reaction to a message.

Use --big to send a big reaction.
Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddAddReactionCommand adds the add-reaction command to the root command.
func AddAddReactionCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AddReactionCmd)

	AddReactionCmd.Flags().BoolVar(&addReactionBig, "big", false, "Send as big reaction")
	AddReactionCmd.Flags().StringVarP(&addReactionPeer, "peer", "p", "", "Peer (e.g., @username)")
	AddReactionCmd.Flags().StringVarP(&addReactionUsername, "username", "u", "", "Username (without @)")
	AddReactionCmd.MarkFlagsOneRequired("peer", "username")
	AddReactionCmd.MarkFlagsMutuallyExclusive("peer", "username")

	AddReactionCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(AddReactionCmd, false)
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
}

var (
	removeReactionPeer     string
	removeReactionUsername string
)

// RemoveReactionCmd represents the remove-reaction command.
var RemoveReactionCmd = &cobra.Command{
	Use:   "remove-reaction <message_id>",
	Short: "Remove reactions from a message",
	Long: `Remove all reactions from a message (or your reaction).

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddRemoveReactionCommand adds the remove-reaction command to the root command.
func AddRemoveReactionCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(RemoveReactionCmd)

	RemoveReactionCmd.Flags().StringVarP(&removeReactionPeer, "peer", "p", "", "Peer (e.g., @username)")
	RemoveReactionCmd.Flags().StringVarP(&removeReactionUsername, "username", "u", "", "Username (without @)")
	RemoveReactionCmd.MarkFlagsOneRequired("peer", "username")
	RemoveReactionCmd.MarkFlagsMutuallyExclusive("peer", "username")

	RemoveReactionCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(RemoveReactionCmd, false)
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
}

var (
	listReactionsPeer     string
	listReactionsUsername string
)

// ListReactionsCmd represents the list-reactions command.
var ListReactionsCmd = &cobra.Command{
	Use:   "list-reactions <message_id>",
	Short: "List reactions on a message",
	Long: `List all reactions on a specific message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddListReactionsCommand adds the list-reactions command to the root command.
func AddListReactionsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ListReactionsCmd)

	ListReactionsCmd.Flags().StringVarP(&listReactionsPeer, "peer", "p", "", "Peer (e.g., @username)")
	ListReactionsCmd.Flags().StringVarP(&listReactionsUsername, "username", "u", "", "Username (without @)")
	ListReactionsCmd.MarkFlagsOneRequired("peer", "username")
	ListReactionsCmd.MarkFlagsMutuallyExclusive("peer", "username")

	ListReactionsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ListReactionsCmd, false)
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
}
