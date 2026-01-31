// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	runner := NewRunnerFromRoot(getUserInfoJSON)
	username := args[0]

	result := runner.CallWithParams("get_user_info", map[string]any{
		"username": username,
	})

	runner.PrintResult(result, printGetUserInfoResult)
}

// printGetUserInfoResult prints the result in a human-readable format.
func printGetUserInfoResult(result any) {
	r, ok := ToMap(result)
	if !ok {
		return
	}

	id := ExtractInt64(r, "id")
	username := ExtractString(r, "username")
	firstName := ExtractString(r, "firstName")
	lastName := ExtractString(r, "lastName")
	phone := ExtractString(r, "phone")
	bio := ExtractString(r, "bio")
	verified, _ := r["verified"].(bool)
	bot, _ := r["bot"].(bool)

	fmt.Printf("User Information:\n")
	fmt.Printf("  ID: %d\n", id)
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
