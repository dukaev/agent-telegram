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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	filePath := args[0]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("update_avatar", map[string]any{
		"file": filePath,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if updateAvatarJSON {
		printUpdateAvatarJSON(result)
	} else {
		printUpdateAvatarResult(result)
	}
}

// printUpdateAvatarJSON prints the result as JSON.
func printUpdateAvatarJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printUpdateAvatarResult prints the result in a human-readable format.
func printUpdateAvatarResult(_ any) {
	fmt.Printf("Avatar updated successfully!\n")
}
