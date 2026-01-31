// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)
	emoji := args[2]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("add_reaction", map[string]any{
		"peer":      peer,
		"messageId": messageID,
		"emoji":     emoji,
		"big":       addReactionBig,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("remove_reaction", map[string]any{
		"peer":      peer,
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("list_reactions", map[string]any{
		"peer":      peer,
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	reactions, _ := r["reactions"].([]any)

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
