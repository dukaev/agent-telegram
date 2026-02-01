// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	readTo    cliutil.Recipient
	readMaxID int64
)

// ReadCmd represents the read command.
var ReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Mark messages as read",
	Long: `Mark messages as read in a chat.

Example:
  agent-telegram msg read --to @channel
  agent-telegram msg read --to @channel --max-id 12345`,
	Args: cobra.NoArgs,
}

// AddReadCommand adds the read command to the parent command.
func AddReadCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ReadCmd)

	ReadCmd.Flags().VarP(&readTo, "to", "t", "Chat/channel to mark as read")
	ReadCmd.Flags().Int64Var(&readMaxID, "max-id", 0, "Mark messages up to this ID as read")
	_ = ReadCmd.MarkFlagRequired("to")

	ReadCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ReadCmd, true)
		params := map[string]any{}
		readTo.AddToParams(params)
		if readMaxID > 0 {
			params["maxId"] = readMaxID
		}

		result := runner.CallWithParams("read_messages", params)
		//nolint:errchkjson // Output to stdout
		_ = json.NewEncoder(os.Stdout).Encode(result)
		if !runner.IsQuiet() {
			cliutil.PrintSuccessSummary(result, "Messages marked as read")
		}
	}
}
