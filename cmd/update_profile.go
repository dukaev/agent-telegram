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
	socketPath, _ := rootCmd.Flags().GetString("socket")

	firstName, _ := cmd.Flags().GetString("first")
	lastName, _ := cmd.Flags().GetString("last")
	bio, _ := cmd.Flags().GetString("bio")

	if firstName == "" {
		fmt.Fprintf(os.Stderr, "Error: --first is required\n")
		os.Exit(1)
	}

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("update_profile", map[string]any{
		"firstName": firstName,
		"lastName":  lastName,
		"bio":       bio,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if updateProfileJSON {
		printUpdateProfileJSON(result)
	} else {
		printUpdateProfileResult(result)
	}
}

// printUpdateProfileJSON prints the result as JSON.
func printUpdateProfileJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printUpdateProfileResult prints the result in a human-readable format.
func printUpdateProfileResult(_ any) {
	fmt.Printf("Profile updated successfully!\n")
}
