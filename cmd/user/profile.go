// Package user provides commands for managing users.
package user

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	updateProfileJSON bool
)

// ProfileCmd represents the update-profile command.
var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Update your Telegram profile",
	Long: `Update your profile information on Telegram.

You can update your first name, last name, and bio.

Example: agent-telegram user profile --first "John" --last "Doe" --bio "Developer"`,
	Args: cobra.NoArgs,
}

// AddProfileCommand adds the profile command to the root command.
func AddProfileCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ProfileCmd)

	ProfileCmd.Flags().BoolVarP(&updateProfileJSON, "json", "j", false, "Output as JSON")
	ProfileCmd.Flags().String("first", "", "First name (required)")
	ProfileCmd.Flags().String("last", "", "Last name")
	ProfileCmd.Flags().String("bio", "", "Bio/about text")

	ProfileCmd.Run = func(command *cobra.Command, _ []string) {
		firstName, _ := command.Flags().GetString("first")
		lastName, _ := command.Flags().GetString("last")
		bio, _ := command.Flags().GetString("bio")

		if firstName == "" {
			fmt.Fprintf(os.Stderr, "Error: --first is required\n")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(ProfileCmd, updateProfileJSON)
		result := runner.CallWithParams("update_profile", map[string]any{
			"firstName": firstName,
			"lastName":  lastName,
			"bio":       bio,
		})
		runner.PrintResult(result, func(any) {
			fmt.Printf("Profile updated successfully!\n")
		})
	}
}
