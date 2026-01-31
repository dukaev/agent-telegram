// Package user provides commands for managing users.
package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	updateAvatarJSON bool
)

// AvatarCmd represents the update-avatar command.
var AvatarCmd = &cobra.Command{
	Use:   "avatar <file>",
	Short: "Update your Telegram avatar/profile photo",
	Long: `Update your profile photo on Telegram.

Supported formats: jpg, jpeg, png, gif, webp

Example: agent-telegram user avatar /path/to/photo.jpg`,
	Args: cobra.ExactArgs(1),
}

// AddAvatarCommand adds the avatar command to the root command.
func AddAvatarCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AvatarCmd)

	AvatarCmd.Flags().BoolVarP(&updateAvatarJSON, "json", "j", false, "Output as JSON")

	AvatarCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(AvatarCmd, updateAvatarJSON)
		result := runner.CallWithParams("update_avatar", map[string]any{
			"file": args[0],
		})
		runner.PrintResult(result, func(any) {
			fmt.Printf("Avatar updated successfully!\n")
		})
	}
}
