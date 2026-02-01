// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	adminsPeer string
	adminsLimit int
)

// AdminsCmd represents the admins command.
var AdminsCmd = &cobra.Command{
	Use:     "admins",
	Short:   "List admins in a chat or channel",
	Long: `List all administrators in a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --limit to set the maximum number of admins to return (max 200).

Example:
  agent-telegram admins --peer @mychannel --limit 20`,
	Args: cobra.NoArgs,
}

// AddAdminsCommand adds the admins command to the root command.
func AddAdminsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AdminsCmd)

	AdminsCmd.Flags().StringVarP(&adminsPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	AdminsCmd.Flags().IntVarP(&adminsLimit, "limit", "l", 100, "Maximum number of admins (max 200)")
	_ = AdminsCmd.MarkFlagRequired("peer")

	AdminsCmd.Run = func(_ *cobra.Command, _ []string) {
		// Validate and sanitize limit
		if adminsLimit < 1 {
			adminsLimit = 1
		}
		if adminsLimit > 200 {
			adminsLimit = 200
		}

		runner := cliutil.NewRunnerFromCmd(AdminsCmd, true)
		params := map[string]any{
			"peer":  adminsPeer,
			"limit": adminsLimit,
		}

		result := runner.CallWithParams("get_admins", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		cliutil.PrintAdmins(result, unknownName, naValue)
	}
}
