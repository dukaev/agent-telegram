// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	clearHistoryJSON bool
	clearHistoryRevoke bool
)

// clearHistoryCmd represents the clear-history command.
var clearHistoryCmd = &cobra.Command{
	Use:   "clear-history @peer",
	Short: "Clear all chat history with a peer",
	Long: `Delete all message history with a Telegram user or chat.

Use --revoke to also delete messages for the other participant (they won't be able to recover them).

Example: agent-telegram clear-history @user --revoke`,
	Args: cobra.ExactArgs(1),
	Run:  runClearHistory,
}

func init() {
	rootCmd.AddCommand(clearHistoryCmd)

	clearHistoryCmd.Flags().BoolVarP(&clearHistoryJSON, "json", "j", false, "Output as JSON")
	clearHistoryCmd.Flags().BoolVar(&clearHistoryRevoke, "revoke", false, "Delete for both parties")
}

func runClearHistory(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("clear_history", map[string]any{
		"peer":   peer,
		"revoke": clearHistoryRevoke,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if clearHistoryJSON {
		printClearHistoryJSON(result)
	} else {
		printClearHistoryResult(result)
	}
}

// printClearHistoryJSON prints the result as JSON.
func printClearHistoryJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printClearHistoryResult prints the result in a human-readable format.
func printClearHistoryResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	peer, _ := r["peer"].(string)
	revoke, _ := r["revoke"].(bool)

	fmt.Printf("History cleared successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	if revoke {
		fmt.Printf("  Revoke: true (deleted for both parties)\n")
	}
}
