// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// DeclineCmd represents the gift decline command.
var DeclineCmd = &cobra.Command{
	Use:   "decline <msgId>",
	Short: "Decline an incoming gift offer",
	Long:  `Decline a gift offer by specifying the message ID of the offer.`,
	Example: `  agent-telegram gift decline 12345`,
	Args: cobra.ExactArgs(1),
}

// AddDeclineCommand adds the decline command to the parent command.
func AddDeclineCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(DeclineCmd)

	DeclineCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(DeclineCmd, false)
		msgID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid message ID: %s\n", args[0])
			os.Exit(1)
		}
		params := map[string]any{
			"offerMsgId": msgID,
		}
		result := runner.CallWithParams("decline_gift_offer", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Gift offer declined successfully!", runner.LastDuration())
		})
	}
}
