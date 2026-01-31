// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	updateProfileJSON bool
)

// updateProfileCmd represents the update-profile command.
var updateProfileCmd = &cobra.Command{
	Use:   "update-profile --first <name> [--last <name>] [--bio <text>]",
	Short: "Update your Telegram profile",
	Long: `Update your profile information on Telegram.

You can update your first name, last name, and bio.

Example: agent-telegram update-profile --first "John" --last "Doe" --bio "Developer"`,
	Args: cobra.NoArgs,
	Run:  runUpdateProfile,
}

func init() {
	updateProfileCmd.Flags().BoolVarP(&updateProfileJSON, "json", "j", false, "Output as JSON")
	updateProfileCmd.Flags().String("first", "", "First name (required)")
	updateProfileCmd.Flags().String("last", "", "Last name")
	updateProfileCmd.Flags().String("bio", "", "Bio/about text")
	rootCmd.AddCommand(updateProfileCmd)
}

func runUpdateProfile(cmd *cobra.Command, _ []string) {
	firstName, _ := cmd.Flags().GetString("first")
	lastName, _ := cmd.Flags().GetString("last")
	bio, _ := cmd.Flags().GetString("bio")

	if firstName == "" {
		fmt.Fprintf(os.Stderr, "Error: --first is required\n")
		os.Exit(1)
	}

	runner := NewRunnerFromRoot(updateProfileJSON)
	result := runner.CallWithParams("update_profile", map[string]any{
		"firstName": firstName,
		"lastName":  lastName,
		"bio":       bio,
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Profile updated successfully!\n")
	})
}
