// Package get provides commands for retrieving information.
package get

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetUserInfoJSON enables JSON output.
	GetUserInfoJSON bool
)

// UserInfoCmd represents the info command.
var UserInfoCmd = &cobra.Command{
	GroupID: "get",
	Use:   "info [@username]",
	Short: "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.
If no username is provided, returns info about the current user (me).

This returns user ID, username, name, bio, verification status, etc.

Examples:
  agent-telegram info          # Get current user info
  agent-telegram info @username  # Get info about a specific user`,
	Args: cobra.MaximumNArgs(1),
}

// AddUserInfoCommand adds the user-info command to the root command.
func AddUserInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UserInfoCmd)

	UserInfoCmd.Flags().BoolVarP(&GetUserInfoJSON, "json", "j", false, "Output as JSON")
	UserInfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UserInfoCmd, true) // Always JSON

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
