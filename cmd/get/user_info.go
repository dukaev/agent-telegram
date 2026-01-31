// Package get provides commands for retrieving information.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetUserInfoJSON enables JSON output.
	GetUserInfoJSON bool
)

// UserInfoCmd represents the get-user-info command.
var UserInfoCmd = &cobra.Command{
	Use:   "user-info @username",
	Short: "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.

This returns user ID, username, name, bio, verification status, etc.

Example: agent-telegram get-user-info @username`,
	Args: cobra.ExactArgs(1),
}

// AddUserInfoCommand adds the user-info command to the root command.
func AddUserInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UserInfoCmd)

	UserInfoCmd.Flags().BoolVarP(&GetUserInfoJSON, "json", "j", false, "Output as JSON")
	UserInfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UserInfoCmd, GetUserInfoJSON)
		username := args[0]

		result := runner.CallWithParams("get_user_info", map[string]any{
			"username": username,
		})

		runner.PrintResult(result, printGetUserInfoResult)
	}
}

// printGetUserInfoResult prints the result in a human-readable format.
func printGetUserInfoResult(result any) {
	r, ok := cliutil.ToMap(result)
	if !ok {
		return
	}

	id := cliutil.ExtractInt64(r, "id")
	username := cliutil.ExtractString(r, "username")
	firstName := cliutil.ExtractString(r, "firstName")
	lastName := cliutil.ExtractString(r, "lastName")
	phone := cliutil.ExtractString(r, "phone")
	bio := cliutil.ExtractString(r, "bio")
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
