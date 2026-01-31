// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateAvatarJSON bool
)

// updateAvatarCmd represents the update-avatar command.
var updateAvatarCmd = &cobra.Command{
	Use:   "update-avatar <file>",
	Short: "Update your Telegram avatar/profile photo",
	Long: `Update your profile photo on Telegram.

Supported formats: jpg, jpeg, png, gif, webp

Example: agent-telegram update-avatar /path/to/photo.jpg`,
	Args: cobra.ExactArgs(1),
	Run:  runUpdateAvatar,
}

func init() {
	updateAvatarCmd.Flags().BoolVarP(&updateAvatarJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(updateAvatarCmd)
}

func runUpdateAvatar(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(updateAvatarJSON)
	result := runner.CallWithParams("update_avatar", map[string]any{
		"file": args[0],
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Avatar updated successfully!\n")
	})
}
