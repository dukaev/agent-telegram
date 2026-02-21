// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// BalanceCmd represents the top-level balance command.
var BalanceCmd = &cobra.Command{
	GroupID: GroupIDChat,
	Use:     "balance",
	Short:   "Show stars and TON balance",
	Long:    `Show your current Telegram Stars and TON balance.`,
	Example: `  agent-telegram balance`,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, false)
		result := runner.CallWithParams("get_balance", map[string]any{})
		runner.PrintResult(result, PrintBalance)
	},
}

func init() {
	RootCmd.AddCommand(BalanceCmd)
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
