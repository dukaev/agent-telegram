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
	sendContactJSON bool
)

// sendContactCmd represents the send-contact command.
var sendContactCmd = &cobra.Command{
	Use:   "send-contact @peer <phone> <firstName>",
	Short: "Send a contact to a Telegram peer",
	Long: `Send a contact to a Telegram user or chat.

The contact will be shared with the peer, who can then add it to their contacts.

Example: agent-telegram send-contact @user +1234567890 "John"`,
	Args: cobra.RangeArgs(2, 3),
	Run:  runSendContact,
}

func init() {
	rootCmd.AddCommand(sendContactCmd)

	sendContactCmd.Flags().BoolVarP(&sendContactJSON, "json", "j", false, "Output as JSON")
}

func runSendContact(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	phone := args[1]
	firstName := args[2]

	lastName := ""
	if len(args) > 3 {
		lastName = args[3]
	}

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_contact", map[string]any{
		"peer":      peer,
		"phone":     phone,
		"firstName": firstName,
		"lastName":  lastName,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendContactJSON {
		printSendContactJSON(result)
	} else {
		printSendContactResult(result)
	}
}

// printSendContactJSON prints the result as JSON.
func printSendContactJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendContactResult prints the result in a human-readable format.
func printSendContactResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)
	phone, _ := r["phone"].(string)

	fmt.Printf("Contact sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  Phone: %s\n", phone)
	fmt.Printf("  ID: %d\n", int64(id))
}
