// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getMeJSON bool
)

// getMeCmd represents the get-me command.
var getMeCmd = &cobra.Command{
	Use:   "get-me",
	Short: "Get current Telegram user info",
	Long:  `Get information about the currently authenticated Telegram user.`,
	Run: runGetMe,
}

func init() {
	rootCmd.AddCommand(getMeCmd)

	getMeCmd.Flags().BoolVarP(&getMeJSON, "json", "j", false, "Output as JSON")
}

func runGetMe(_ *cobra.Command, _ []string) {
	runner := NewRunnerFromRoot(getMeJSON)
	result := runner.Call("get_me", nil)

	runner.PrintResult(result, printUserInfo)
}

// printUserInfo prints user information in a human-readable format.
func printUserInfo(result any) {
	user, ok := ToMap(result)
	if !ok {
		return
	}

	fmt.Printf("User Info:\n")
	printName(user)
	printField(user, "username", "Username: @%s\n")
	printField(user, "phone", "Phone: %s\n")
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

// printField prints a field if it exists and is a non-empty string.
func printField(user map[string]any, key, format string) {
	value, ok := user[key].(string)
	if !ok || value == "" {
		return
	}
	fmt.Printf(format, value)
}

// printID prints the user's ID.
func printID(user map[string]any) {
	if id, ok := user["id"].(float64); ok {
		fmt.Printf("  ID: %d\n", int64(id))
	}
}
