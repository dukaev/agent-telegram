// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// BalanceCmd represents the gift balance command.
var BalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show stars and TON balance",
	Long:    `Show your current Telegram Stars and TON balance.`,
	Example: `  agent-telegram gift balance`,
	Args: cobra.NoArgs,
}

// AddBalanceCommand adds the balance command to the parent command.
func AddBalanceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(BalanceCmd)

	BalanceCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(BalanceCmd, false)
		result := runner.CallWithParams("get_balance", map[string]any{})
		runner.PrintResult(result, PrintBalance)
	}
}

// PrintBalance prints the stars and TON balance from a result map.
func PrintBalance(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get balance")
		return
	}

	stars, _ := r["stars"].(float64)
	nanos, _ := r["nanos"].(float64)
	ton, _ := r["ton"].(float64)

	if nanos != 0 {
		// Show fractional stars (nanos / 1e9)
		frac := float64(int64(stars)) + float64(int(nanos))/1e9
		fmt.Fprintf(os.Stderr, "Stars: %.2f\n", frac)
	} else {
		fmt.Fprintf(os.Stderr, "Stars: %d\n", int64(stars))
	}
	if ton != 0 {
		fmt.Fprintf(os.Stderr, "TON:   %d\n", int64(ton))
	}
}
