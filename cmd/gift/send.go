// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendTo       cliutil.Recipient
	sendHideName bool
	sendMessage  string
	sendMsgID    int
	sendBuy      bool
)

// SendCmd represents the gift send command.
var SendCmd = &cobra.Command{
	Use:   "send <gift_name_or_slug>",
	Short: "Send, transfer, or buy a gift",
	Long: `Send a star gift from the catalog, transfer a saved gift, or buy from the marketplace.

  gift send Heart --to @user           Send a new gift from catalog
  gift send SantaHat-55373 --to @user  Transfer your saved gift
  gift send SantaHat-55373             Buy from marketplace (for yourself)
  gift send SantaHat-55373 --to @user --buy  Buy for another user
  gift send --to @user --msg-id 123    Transfer by message ID`,
	Example: `  agent-telegram gift send Heart --to @username
  agent-telegram gift send Heart --to @username -m "Happy birthday!"
  agent-telegram gift send SantaHat-55373 --to @username
  agent-telegram gift send SantaHat-55373
  agent-telegram gift send https://t.me/nft/SwissWatch-718 --buy
  agent-telegram gift send --to @username --msg-id 123`,
	Args: cobra.MaximumNArgs(1),
}

// AddSendCommand adds the send command to the parent command.
func AddSendCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(SendCmd)

	SendCmd.Flags().VarP(&sendTo, "to", "t", "Recipient (@username or ID)")
	SendCmd.Flags().StringVarP(&sendMessage, "msg", "m", "", "Message to attach with the gift")
	SendCmd.Flags().BoolVar(&sendHideName, "hide-name", false, "Hide your name from the recipient's profile")
	SendCmd.Flags().IntVar(&sendMsgID, "msg-id", 0, "Message ID of a saved gift to transfer")
	SendCmd.Flags().BoolVar(&sendBuy, "buy", false, "Buy from marketplace instead of transferring")

	SendCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(SendCmd, false)
		hasSlug := len(args) > 0 && isGiftSlug(args[0])
		hasTo := sendTo.Peer() != ""

		// Buy from marketplace: slug without --to, or --buy flag
		if sendBuy || (hasSlug && !hasTo && sendMsgID == 0) {
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: requires a gift slug or URL to buy")
				os.Exit(1)
			}
			params := map[string]any{
				"slug": cliutil.ParseGiftSlug(args[0]),
			}
			if hasTo {
				params["peer"] = sendTo.Peer()
			}
			result := runner.CallWithParams("buy_resale_gift", params)
			runner.PrintResult(result, func(result any) {
				cliutil.PrintSuccessWithDuration(result, "Gift purchased successfully!", runner.LastDuration())
			})
			return
		}

		// Transfer: slug/URL with --to, or --msg-id
		if sendMsgID != 0 || hasSlug {
			if !hasTo {
				fmt.Fprintln(os.Stderr, "Error: --to is required for transfers")
				os.Exit(1)
			}
			params := map[string]any{}
			sendTo.AddToParams(params)
			if len(args) > 0 {
				params["slug"] = cliutil.ParseGiftSlug(args[0])
			}
			if sendMsgID != 0 {
				params["msgId"] = sendMsgID
			}
			result := runner.CallWithParams("transfer_star_gift", params)
			runner.PrintResult(result, func(result any) {
				cliutil.PrintSuccessWithDuration(result, "Star gift transferred successfully!", runner.LastDuration())
			})
			return
		}

		// Send new gift from catalog
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: requires a gift name, catalog ID, slug, or --msg-id")
			os.Exit(1)
		}
		if !hasTo {
			fmt.Fprintln(os.Stderr, "Error: --to is required when sending a gift")
			os.Exit(1)
		}

		params := map[string]any{}
		sendTo.AddToParams(params)
		if sendMessage != "" {
			params["message"] = sendMessage
		}
		if sendHideName {
			params["hideName"] = true
		}
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			params["giftId"] = id
		} else {
			params["name"] = args[0]
		}
		result := runner.CallWithParams("send_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Star gift sent successfully!", runner.LastDuration())
		})
	}
}

// isGiftSlug checks if the argument looks like a gift slug (Name-Number) or URL.
func isGiftSlug(arg string) bool {
	// URLs are always slugs
	if strings.HasPrefix(arg, "http") || strings.HasPrefix(arg, "t.me/") {
		return true
	}
	// Slugs contain a hyphen (e.g. SantaHat-55373), gift names don't
	return strings.Contains(arg, "-")
}
