// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// ValueCmd represents the gift value command.
var ValueCmd = &cobra.Command{
	Use:   "value <slug>",
	Short: "Show value analytics for a unique gift",
	Long: `Show pricing and value information for a unique star gift,
including floor price, average price, last sale, and Fragment listings.

Example:
  agent-telegram gift value SwissWatch-718
  agent-telegram gift value RestlessJar-55271`,
	Args: cobra.ExactArgs(1),
}

// AddValueCommand adds the value command to the parent command.
func AddValueCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ValueCmd)

	ValueCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ValueCmd, false)
		params := map[string]any{
			"slug": args[0],
		}
		result := runner.CallWithParams("get_gift_value", params)
		runner.PrintResult(result, printGiftValue)
	}
}

func printGiftValue(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get gift value")
		return
	}

	currency := cliutil.ExtractStringValue(r, "currency")
	value := cliutil.ExtractInt64(r, "value")
	initialStars := cliutil.ExtractInt64(r, "initialSaleStars")
	initialPrice := cliutil.ExtractInt64(r, "initialSalePrice")
	initialDate := cliutil.ExtractInt64(r, "initialSaleDate")
	floorPrice := cliutil.ExtractInt64(r, "floorPrice")
	avgPrice := cliutil.ExtractInt64(r, "averagePrice")
	lastPrice := cliutil.ExtractInt64(r, "lastSalePrice")
	lastDate := cliutil.ExtractInt64(r, "lastSaleDate")
	lastOnFragment, _ := r["lastSaleOnFragment"].(bool)
	valueIsAvg, _ := r["valueIsAverage"].(bool)
	listedCount := cliutil.ExtractInt64(r, "listedCount")
	fragmentCount := cliutil.ExtractInt64(r, "fragmentListedCount")
	fragmentURL := cliutil.ExtractStringValue(r, "fragmentListedUrl")

	// Estimated value
	valueLabel := "Estimated value"
	if valueIsAvg {
		valueLabel = "Estimated value (avg)"
	}
	if value > 0 {
		fmt.Fprintf(os.Stderr, "%s: %d %s\n", valueLabel, value, currency)
	}

	// Initial sale
	if initialStars > 0 {
		line := fmt.Sprintf("Initial sale: %d stars", initialStars)
		if initialPrice > 0 {
			line += fmt.Sprintf(" (%d %s)", initialPrice, currency)
		}
		if initialDate > 0 {
			t := time.Unix(initialDate, 0)
			line += fmt.Sprintf(" on %s", t.Format("2006-01-02"))
		}
		fmt.Fprintln(os.Stderr, line)
	}

	// Floor / Average
	if floorPrice > 0 {
		fmt.Fprintf(os.Stderr, "Floor price: %d stars\n", floorPrice)
	}
	if avgPrice > 0 {
		fmt.Fprintf(os.Stderr, "Average price: %d stars\n", avgPrice)
	}

	// Last sale
	if lastPrice > 0 {
		line := fmt.Sprintf("Last sale: %d stars", lastPrice)
		if lastDate > 0 {
			t := time.Unix(lastDate, 0)
			line += fmt.Sprintf(" on %s", t.Format("2006-01-02"))
		}
		if lastOnFragment {
			line += " (on Fragment)"
		}
		fmt.Fprintln(os.Stderr, line)
	}

	// Listings
	if listedCount > 0 {
		fmt.Fprintf(os.Stderr, "Listed for sale: %d\n", listedCount)
	}
	if fragmentCount > 0 {
		fmt.Fprintf(os.Stderr, "Listed on Fragment: %d\n", fragmentCount)
	}
	if fragmentURL != "" {
		fmt.Fprintf(os.Stderr, "Fragment URL: %s\n", fragmentURL)
	}
}
