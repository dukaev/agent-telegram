package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	textFlags         SendFlags
	textReplyToMsgID  int64
)

// TextCmd represents the send text subcommand.
var TextCmd = &cobra.Command{
	Use:   "text <peer> <message>",
	Short: "Send a text message",
	Long: `Send a text message to a Telegram user or chat.

Examples:
  agent-telegram send text @user "Hello world"
  agent-telegram send text --to @user "Hello world"
  echo "Hello" | agent-telegram send text @user`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		var messageText string
		stdinText := cliutil.ReadStdinIfPiped()

		switch len(args) {
		case 2:
			_ = textFlags.To.Set(args[0])
			messageText = args[1]
		case 1:
			if textFlags.To.Peer() != "" {
				messageText = args[0]
			} else {
				_ = textFlags.To.Set(args[0])
				messageText = stdinText
			}
		}

		textFlags.RequirePeer()
		runner := textFlags.NewRunner()

		params := map[string]any{
			"message": messageText,
		}
		textFlags.To.AddToParams(params)

		method := "send_message"
		if textReplyToMsgID != 0 {
			params["text"] = messageText
			delete(params, "message")
			params["messageId"] = textReplyToMsgID
			method = methodSendReply
		}

		result := runner.CallWithParams(method, params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, method)
		})
	},
}

func addTextCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(TextCmd)
	textFlags.RegisterOptionalTo(TextCmd)
	TextCmd.Flags().Int64Var(&textReplyToMsgID, "reply-to", 0, "Reply to message ID")
}
