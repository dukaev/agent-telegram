// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	typingTo     cliutil.Recipient
	typingAction string
)

// TypingCmd represents the typing command.
var TypingCmd = &cobra.Command{
	Use:   "typing",
	Short: "Send typing indicator",
	Long: `Send a typing indicator to a chat.

Actions: typing (default), upload_photo, record_video, record_audio,
         upload_document, choose_sticker, record_round, cancel

Example:
  agent-telegram msg typing --to @user
  agent-telegram msg typing --to @user --action record_audio
  agent-telegram msg typing --to @user --action cancel`,
	Args: cobra.NoArgs,
}

// AddTypingCommand adds the typing command to the parent command.
func AddTypingCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(TypingCmd)

	TypingCmd.Flags().VarP(&typingTo, "to", "t", "Chat/user to send typing to")
	TypingCmd.Flags().StringVarP(&typingAction, "action", "a", "typing", "Typing action")
	_ = TypingCmd.MarkFlagRequired("to")

	TypingCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(TypingCmd, true)
		params := map[string]any{
			"action": typingAction,
		}
		typingTo.AddToParams(params)

		result := runner.CallWithParams("set_typing", params)
		//nolint:errchkjson // Output to stdout
		_ = json.NewEncoder(os.Stdout).Encode(result)
		if !runner.IsQuiet() {
			cliutil.PrintSuccessSummary(result, "Typing indicator sent")
		}
	}
}
