// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// GiftCmd represents the parent gift command.
var GiftCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "gift",
	Short:   "Manage Telegram star gifts",
	Long:    `Commands for managing Telegram star gifts - list catalog, send, view saved, transfer, and convert.`,
	RunE:    runDashboard,
}

func runDashboard(cmd *cobra.Command, _ []string) error {
	// If --help was explicitly passed, let cobra handle it
	if help, _ := cmd.Flags().GetBool("help"); help {
		return cmd.Help()
	}

	runner := cliutil.NewRunnerFromCmd(cmd, false)

	// Get balance
	balanceResult := runner.CallWithParams("get_balance", map[string]any{})
	PrintBalance(balanceResult)

	// Get recent saved gifts
	giftsResult := runner.CallWithParams("get_saved_gifts", map[string]any{
		"limit": 5,
	})
	if r, ok := giftsResult.(map[string]any); ok {
		count := cliutil.ExtractInt64(r, "count")
		fmt.Fprintf(os.Stderr, "\nRecent gifts (showing up to 5 of %d):\n", count)
		gifts, _ := r["gifts"].([]any)
		for _, g := range gifts {
			gift, ok := g.(map[string]any)
			if !ok {
				continue
			}
			PrintSavedGiftItem(gift)
		}
	}

	fmt.Fprintln(os.Stderr, "\nRun 'gift my' for full list, 'gift list' for catalog.")
	return nil
}

// AddGiftCommand adds the parent gift command and all its subcommands to the root command.
func AddGiftCommand(rootCmd *cobra.Command) {
	AddListCommand(GiftCmd)
	AddSendCommand(GiftCmd)
	AddMyCommand(GiftCmd)
	AddTransferCommand(GiftCmd)
	AddConvertCommand(GiftCmd)
	AddPriceCommand(GiftCmd)
	AddBalanceCommand(GiftCmd)
	AddOfferCommand(GiftCmd)
	AddInfoCommand(GiftCmd)
	AddValueCommand(GiftCmd)
	AddMarketCommand(GiftCmd)
	AddBuyCommand(GiftCmd)
	AddAttrsCommand(GiftCmd)
	AddAcceptCommand(GiftCmd)
	AddDeclineCommand(GiftCmd)

	rootCmd.AddCommand(GiftCmd)
}
