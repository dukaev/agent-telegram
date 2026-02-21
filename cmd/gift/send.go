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
)

// SendCmd represents the gift send command.
var SendCmd = &cobra.Command{
	Use:   "send <gift_name_or_slug>",
	Short: "Send a gift or transfer an existing one",
	Long: `Send a star gift from the catalog, or transfer a saved gift to another user.

If the argument is a gift name or catalog ID, a new gift is purchased and sent.
If the argument is a slug (e.g. SantaHat-55373) or URL, an existing gift is transferred.
Use --msg-id to transfer by saved gift message ID.`,
	Example: `  agent-telegram gift send Heart --to @username
  agent-telegram gift send Heart --to @username -m "Happy birthday!"
  agent-telegram gift send 5170145012310081615 --to @username
  agent-telegram gift send SantaHat-55373 --to @username
  agent-telegram gift send https://t.me/nft/SantaHat-55373 --to @username
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
	_ = SendCmd.MarkFlagRequired("to")

	SendCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(SendCmd, false)

		// Determine if this is a transfer or a catalog send
		if sendMsgID != 0 || (len(args) > 0 && isGiftSlug(args[0])) {
			// Transfer an existing gift
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

		// Send a new gift from catalog
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: requires a gift name, catalog ID, slug, or --msg-id")
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
