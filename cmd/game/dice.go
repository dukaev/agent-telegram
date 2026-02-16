package game

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	diceTo       cliutil.Recipient
	diceEmoticon string
)

var diceCmd = &cobra.Command{
	Use:   "dice <peer>",
	Short: "Roll a dice in a Telegram chat",
	Long: `Send a dice to a Telegram chat and return the rolled value.

Examples:
  agent-telegram game dice @user
  agent-telegram game dice @channel --emoticon 🎯

  # Roll and send gift if 6:
  val=$(agent-telegram game dice @user | jq -r .value)
  [ "$val" = "6" ] && agent-telegram gift send Bear --to @user`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDiceGame(cmd, args)
	},
}

func addDiceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(diceCmd)
	diceCmd.Flags().VarP(&diceTo, "to", "t", "Recipient (@username or ID)")
	diceCmd.Flags().StringVar(&diceEmoticon, "emoticon", "", "Dice emoticon (default: 🎲, also: 🎯, 🏀, ⚽, 🎳, 🎰)")
}

func runDiceGame(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		_ = diceTo.Set(args[0])
	}
	if diceTo.Peer() == "" {
		fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
		os.Exit(1)
	}

	runner := cliutil.NewRunnerFromCmd(cmd, false)

	params := map[string]any{}
	diceTo.AddToParams(params)
	if diceEmoticon != "" {
		params["emoticon"] = diceEmoticon
	}

	result := runner.CallWithParams("send_dice", params)
	runner.PrintResult(result, func(r any) {
		cliutil.FormatSuccess(r, "send_dice")
	})
}
