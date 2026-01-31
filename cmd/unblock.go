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
	unblockJSON bool
)

// unblockCmd represents the unblock command.
var unblockCmd = &cobra.Command{
	Use:   "unblock @peer",
	Short: "Unblock a Telegram peer",
	Long: `Unblock a Telegram user or chat.

Unblocked peers will be able to send you messages again.

Example: agent-telegram unblock @user`,
	Args: cobra.ExactArgs(1),
	Run:  runUnblock,
}

func init() {
	unblockCmd.Flags().BoolVarP(&unblockJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(unblockCmd)
}

func runUnblock(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("unblock", map[string]any{
		"peer": peer,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if unblockJSON {
		printUnblockJSON(result)
	} else {
		printUnblockResult(result)
	}
}

// printUnblockJSON prints the result as JSON.
func printUnblockJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printUnblockResult prints the result in a human-readable format.
func printUnblockResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	peer, _ := r["peer"].(string)

	fmt.Printf("Peer unblocked successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
}
