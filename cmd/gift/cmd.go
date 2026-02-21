// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"
)

// GiftCmd represents the parent gift command.
var GiftCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "gift",
	Short:   "Manage Telegram star gifts",
	Long:    `Commands for managing Telegram star gifts - list catalog, send, view saved, transfer, and convert.`,
}

// AddGiftCommand adds the parent gift command and all its subcommands to the root command.
func AddGiftCommand(rootCmd *cobra.Command) {
	AddListCommand(GiftCmd)
	AddSendCommand(GiftCmd)
	AddMyCommand(GiftCmd)
	AddConvertCommand(GiftCmd)
	AddPriceCommand(GiftCmd)
	AddOfferCommand(GiftCmd)
	AddInfoCommand(GiftCmd)
	AddBuyCommand(GiftCmd)
	AddAttrsCommand(GiftCmd)
	AddAcceptCommand(GiftCmd)
	AddDeclineCommand(GiftCmd)

	rootCmd.AddCommand(GiftCmd)
}
