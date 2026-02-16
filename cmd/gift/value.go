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
	value, _ := r["value"].(float64)
	initialStars, _ := r["initialSaleStars"].(float64)
	initialPrice, _ := r["initialSalePrice"].(float64)
	initialDate, _ := r["initialSaleDate"].(float64)
	floorPrice, _ := r["floorPrice"].(float64)
	avgPrice, _ := r["averagePrice"].(float64)
	lastPrice, _ := r["lastSalePrice"].(float64)
	lastDate, _ := r["lastSaleDate"].(float64)
	lastOnFragment, _ := r["lastSaleOnFragment"].(bool)
	valueIsAvg, _ := r["valueIsAverage"].(bool)
	listedCount, _ := r["listedCount"].(float64)
	fragmentCount, _ := r["fragmentListedCount"].(float64)
	fragmentURL := cliutil.ExtractStringValue(r, "fragmentListedUrl")

	// Estimated value
	valueLabel := "Estimated value"
	if valueIsAvg {
		valueLabel = "Estimated value (avg)"
	}
	if value > 0 {
		fmt.Fprintf(os.Stderr, "%s: %d %s\n", valueLabel, int64(value), currency)
	}

	// Initial sale
	if initialStars > 0 {
		line := fmt.Sprintf("Initial sale: %d stars", int64(initialStars))
		if initialPrice > 0 {
			line += fmt.Sprintf(" (%d %s)", int64(initialPrice), currency)
		}
		if initialDate > 0 {
			t := time.Unix(int64(initialDate), 0)
			line += fmt.Sprintf(" on %s", t.Format("2006-01-02"))
		}
		fmt.Fprintln(os.Stderr, line)
	}

	// Floor / Average
	if floorPrice > 0 {
		fmt.Fprintf(os.Stderr, "Floor price: %d stars\n", int64(floorPrice))
	}
	if avgPrice > 0 {
		fmt.Fprintf(os.Stderr, "Average price: %d stars\n", int64(avgPrice))
	}

	// Last sale
	if lastPrice > 0 {
		line := fmt.Sprintf("Last sale: %d stars", int64(lastPrice))
		if lastDate > 0 {
			t := time.Unix(int64(lastDate), 0)
			line += fmt.Sprintf(" on %s", t.Format("2006-01-02"))
		}
		if lastOnFragment {
			line += " (on Fragment)"
		}
		fmt.Fprintln(os.Stderr, line)
	}

	// Listings
	if listedCount > 0 {
		fmt.Fprintf(os.Stderr, "Listed for sale: %d\n", int(listedCount))
	}
	if fragmentCount > 0 {
		fmt.Fprintf(os.Stderr, "Listed on Fragment: %d\n", int(fragmentCount))
	}
	if fragmentURL != "" {
		fmt.Fprintf(os.Stderr, "Fragment URL: %s\n", fragmentURL)
	}
}
