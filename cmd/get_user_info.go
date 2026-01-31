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
	getUserInfoJSON bool
)

// getUserInfoCmd represents the get-user-info command.
var getUserInfoCmd = &cobra.Command{
	Use:   "get-user-info @username",
	Short: "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.

This returns user ID, username, name, bio, verification status, etc.

Example: agent-telegram get-user-info @username`,
	Args: cobra.ExactArgs(1),
	Run:  runGetUserInfo,
}

func init() {
	getUserInfoCmd.Flags().BoolVarP(&getUserInfoJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(getUserInfoCmd)
}

func runGetUserInfo(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	username := args[0]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("get_user_info", map[string]any{
		"username": username,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if getUserInfoJSON {
		printGetUserInfoJSON(result)
	} else {
		printGetUserInfoResult(result)
	}
}

// printGetUserInfoJSON prints the result as JSON.
func printGetUserInfoJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printGetUserInfoResult prints the result in a human-readable format.
func printGetUserInfoResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	username, _ := r["username"].(string)
	firstName, _ := r["firstName"].(string)
	lastName, _ := r["lastName"].(string)
	phone, _ := r["phone"].(string)
	bio, _ := r["bio"].(string)
	verified, _ := r["verified"].(bool)
	bot, _ := r["bot"].(bool)

	fmt.Printf("User Information:\n")
	fmt.Printf("  ID: %d\n", int64(id))
	fmt.Printf("  Username: @%s\n", username)
	fmt.Printf("  Name: %s %s\n", firstName, lastName)
	if phone != "" {
		fmt.Printf("  Phone: %s\n", phone)
	}
	if bio != "" {
		fmt.Printf("  Bio: %s\n", bio)
	}
	fmt.Printf("  Verified: %v\n", verified)
	fmt.Printf("  Bot: %v\n", bot)
}
