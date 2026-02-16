// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// AcceptCmd represents the gift accept command.
var AcceptCmd = &cobra.Command{
	Use:   "accept <msgId>",
	Short: "Accept an incoming gift offer",
	Long:  `Accept a gift offer by specifying the message ID of the offer.`,
	Example: `  agent-telegram gift accept 12345`,
	Args: cobra.ExactArgs(1),
}

// AddAcceptCommand adds the accept command to the parent command.
func AddAcceptCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(AcceptCmd)

	AcceptCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(AcceptCmd, false)
		msgID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid message ID: %s\n", args[0])
			os.Exit(1)
		}
		params := map[string]any{
			"offerMsgId": msgID,
		}
		result := runner.CallWithParams("accept_gift_offer", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Gift offer accepted successfully!", runner.LastDuration())
		})
	}
}
