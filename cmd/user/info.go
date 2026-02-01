// Package user provides commands for managing user-related operations.
package user

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// InfoCmd represents the user info command.
var InfoCmd = &cobra.Command{
	Use:     "info [@username]",
	Short:   "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.
If no username is provided, returns info about the current user (me).

This returns user ID, username, name, bio, verification status, etc.

Examples:
  agent-telegram user info          # Get current user info
  agent-telegram user info @username  # Get info about a specific user`,
	Args: cobra.MaximumNArgs(1),
}

// AddInfoCommand adds the info command to the parent command.
func AddInfoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(InfoCmd)

	InfoCmd.Flags().BoolVarP(&infoJSON, "json", "j", false, "Output as JSON")

	InfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, true) // Always JSON

		var result any
		if len(args) == 0 {
			// No username provided - get current user info
			result = runner.Call("get_me", nil)
		} else {
			// Username provided - get specific user info
			username := args[0]
			result = runner.CallWithParams("get_user_info", map[string]any{
				"username": username,
			})
		}

		// Output as JSON
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)
	}
}

var (
	infoJSON bool
)
