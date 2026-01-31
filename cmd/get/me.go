// Package get provides commands for retrieving information.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetMeJSON enables JSON output.
	GetMeJSON bool
)

// MeCmd represents the get-me command.
var MeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current Telegram user info",
	Long:  `Get information about the currently authenticated Telegram user.`,
}

// AddMeCommand adds the me command to the root command.
func AddMeCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MeCmd)

	MeCmd.Flags().BoolVarP(&GetMeJSON, "json", "j", false, "Output as JSON")
	MeCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(MeCmd, GetMeJSON)
		result := runner.Call("get_me", nil)
		runner.PrintResult(result, printUserInfo)
	}
}

// printUserInfo prints user information in a human-readable format.
func printUserInfo(result any) {
	user, ok := cliutil.ToMap(result)
	if !ok {
		return
	}

	fmt.Printf("User Info:\n")
	printName(user)
	if username := cliutil.ExtractString(user, "username"); username != "" {
		fmt.Printf("Username: @%s\n", username)
	}
	if phone := cliutil.ExtractString(user, "phone"); phone != "" {
		fmt.Printf("Phone: %s\n", phone)
	}
	printID(user)
}

// printName prints the user's first and last name.
func printName(user map[string]any) {
	firstName, firstOk := user["first_name"].(string)
	if !firstOk || firstName == "" {
		return
	}
	fmt.Printf("  Name: %s", firstName)
	if lastName, ok := user["last_name"].(string); ok && lastName != "" {
		fmt.Printf(" %s", lastName)
	}
	fmt.Println()
}

// printID prints the user's ID.
func printID(user map[string]any) {
	if id, ok := user["id"].(float64); ok {
		fmt.Printf("  ID: %d\n", int64(id))
	}
}
