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
	blockJSON bool
)

// blockCmd represents the block command.
var blockCmd = &cobra.Command{
	Use:   "block @peer",
	Short: "Block a Telegram peer",
	Long: `Block a Telegram user or chat.

Blocked peers will not be able to send you messages or see your online status.

Example: agent-telegram block @user`,
	Args: cobra.ExactArgs(1),
	Run:  runBlock,
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().BoolVarP(&blockJSON, "json", "j", false, "Output as JSON")
}

func runBlock(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("block", map[string]any{
		"peer": peer,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if blockJSON {
		printBlockJSON(result)
	} else {
		printBlockResult(result)
	}
}

// printBlockJSON prints the result as JSON.
func printBlockJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printBlockResult prints the result in a human-readable format.
func printBlockResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	peer, _ := r["peer"].(string)

	fmt.Printf("Peer blocked successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
}
