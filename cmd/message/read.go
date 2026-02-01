// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	readPeer  string
	readMaxID int64
)

// ReadCmd represents the read command.
var ReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Mark messages as read",
	Long: `Mark messages as read in a chat.

Example:
  agent-telegram msg read --peer @channel
  agent-telegram msg read --peer @channel --max-id 12345`,
	Args: cobra.NoArgs,
}

// AddReadCommand adds the read command to the parent command.
func AddReadCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ReadCmd)

	ReadCmd.Flags().StringVarP(&readPeer, "peer", "p", "", "Chat/channel to mark as read")
	ReadCmd.Flags().Int64Var(&readMaxID, "max-id", 0, "Mark messages up to this ID as read")
	_ = ReadCmd.MarkFlagRequired("peer")

	ReadCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ReadCmd, true)
		params := map[string]any{
			"peer": readPeer,
		}
		if readMaxID > 0 {
			params["maxId"] = readMaxID
		}

		result := runner.CallWithParams("read_messages", params)
		//nolint:errchkjson // Output to stdout
		_ = json.NewEncoder(os.Stdout).Encode(result)
		cliutil.PrintSuccessSummary(result, "Messages marked as read")
	}
}
