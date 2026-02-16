package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	diceSubFlags      SendFlags
	diceSubEmoticon   string
	diceSubReplyToID  int64
)

// DiceCmd represents the send dice subcommand.
var DiceCmd = &cobra.Command{
	Use:   "dice <peer>",
	Short: "Send a dice (random value)",
	Long: `Send a dice with a random value to a Telegram user or chat.

Emoticons: 🎲 (default), 🎯, 🏀, ⚽, 🎳, 🎰

Examples:
  agent-telegram send dice @user
  agent-telegram send dice --to @user --emoticon 🎯`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = diceSubFlags.To.Set(args[0])
		}
		diceSubFlags.RequirePeer()

		runner := diceSubFlags.NewRunner()
		params := map[string]any{}
		diceSubFlags.To.AddToParams(params)
		if diceSubEmoticon != "" {
			params["emoticon"] = diceSubEmoticon
		}
		if diceSubReplyToID != 0 {
			params["replyTo"] = diceSubReplyToID
		}

		result := runner.CallWithParams("send_dice", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_dice")
		})
	},
}

func addDiceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(DiceCmd)
	diceSubFlags.RegisterOptionalTo(DiceCmd)
	DiceCmd.Flags().StringVar(&diceSubEmoticon, "emoticon", "", "Dice emoticon (default: 🎲, also: 🎯, 🏀, ⚽, 🎳, 🎰)")
	DiceCmd.Flags().Int64Var(&diceSubReplyToID, "reply-to", 0, "Reply to message ID")
}
