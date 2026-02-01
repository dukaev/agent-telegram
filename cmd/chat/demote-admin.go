//nolint:dupl // Similar command structure is intentional
package chat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	demoteAdminPeer string
	demoteAdminUser string
)

// DemoteAdminCmd represents the demote-admin command.
var DemoteAdminCmd = &cobra.Command{
	Use:     "demote-admin",
	Short:   "Demote an admin to regular user",
	Long: `Demote an administrator to a regular user in a Telegram channel or supergroup.

Use --peer @username or --peer username to specify the channel.
Use --user @username or --user username to specify the admin to demote.

Example:
  agent-telegram demote-admin --peer @mychannel --user @username`,
	Args: cobra.NoArgs,
}

// AddDemoteAdminCommand adds the demote-admin command to the root command.
func AddDemoteAdminCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DemoteAdminCmd)

	DemoteAdminCmd.Flags().StringVarP(&demoteAdminPeer, "peer", "p", "", "Channel username (@username or username)")
	DemoteAdminCmd.Flags().StringVarP(&demoteAdminUser, "user", "u", "", "Admin to demote (@username or username)")
	_ = DemoteAdminCmd.MarkFlagRequired("peer")
	_ = DemoteAdminCmd.MarkFlagRequired("user")

	DemoteAdminCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(DemoteAdminCmd, true)
		params := map[string]any{
			"peer":  demoteAdminPeer,
			"user":  demoteAdminUser,
		}

		result := runner.CallWithParams("demote_admin", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Admin demoted successfully\n")
			}
		}
	}
}
