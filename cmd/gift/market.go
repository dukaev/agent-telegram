// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	marketLimit       int
	marketOffset      string
	marketSortByPrice bool
	marketSortByNum   bool
	marketModel       string
	marketPattern     string
	marketBackdrop    string
)

// MarketCmd represents the gift market command.
var MarketCmd = &cobra.Command{
	Use:   "market <gift_id>",
	Short: "List gifts for resale (marketplace)",
	Long: `Browse star gifts listed for resale by gift type ID.
Use "gift list" to find gift IDs from the catalog.

Example:
  agent-telegram gift market 5170145012310081536
  agent-telegram gift market 5170145012310081536 --sort-price --limit 10
  agent-telegram gift market 5170145012310081536 --model "Homunculus" --backdrop "Turquoise"`,
	Args: cobra.ExactArgs(1),
}

// AddMarketCommand adds the market command to the parent command.
func AddMarketCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(MarketCmd)

	MarketCmd.Flags().IntVarP(&marketLimit, "limit", "l", cliutil.DefaultLimitMedium, "Max gifts to show")
	MarketCmd.Flags().StringVarP(&marketOffset, "offset", "o", "", "Offset for pagination")
	MarketCmd.Flags().BoolVar(&marketSortByPrice, "sort-price", false, "Sort by price (lowest first)")
	MarketCmd.Flags().BoolVar(&marketSortByNum, "sort-num", false, "Sort by serial number")
	MarketCmd.Flags().StringVar(&marketModel, "model", "", "Filter by model name")
	MarketCmd.Flags().StringVar(&marketPattern, "pattern", "", "Filter by pattern name")
	MarketCmd.Flags().StringVar(&marketBackdrop, "backdrop", "", "Filter by backdrop name")

	MarketCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(MarketCmd, false)
		params := map[string]any{
			"giftId": parseGiftID(args[0]),
			"limit":  marketLimit,
		}
		if marketOffset != "" {
			params["offset"] = marketOffset
		}
		if marketSortByPrice {
			params["sortByPrice"] = true
		}
		if marketSortByNum {
			params["sortByNum"] = true
		}
		if marketModel != "" {
			params["model"] = marketModel
		}
		if marketPattern != "" {
			params["pattern"] = marketPattern
		}
		if marketBackdrop != "" {
			params["backdrop"] = marketBackdrop
		}
		result := runner.CallWithParams("get_resale_gifts", params)
		runner.PrintResult(result, printResaleGifts)
	}
}

func parseGiftID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}

func printResaleGifts(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get resale gifts")
		return
	}

	count, _ := r["count"].(float64)
	fmt.Fprintf(os.Stderr, "Resale Listings (%d):\n", int(count))

	gifts, _ := r["gifts"].([]any)
	for _, g := range gifts {
		gift, ok := g.(map[string]any)
		if !ok {
			continue
		}
		printResaleItem(gift)
	}

	if nextOffset := cliutil.ExtractStringValue(r, "nextOffset"); nextOffset != "" {
		fmt.Fprintf(os.Stderr, "\nNext offset: %s\n", nextOffset)
	}
}

func printResaleItem(gift map[string]any) {
	title := cliutil.ExtractStringValue(gift, "title")
	slug := cliutil.ExtractStringValue(gift, "slug")
	num, _ := gift["num"].(float64)
	resell, _ := gift["resellStars"].(float64)
	owner := cliutil.ExtractStringValue(gift, "ownerName")

	line := fmt.Sprintf("  - %s #%d", title, int(num))
	if resell > 0 {
		line += fmt.Sprintf(" — %d stars", int64(resell))
	}
	if owner != "" {
		line += fmt.Sprintf(" (owner: %s)", owner)
	}

	// Show attributes inline
	attrs, _ := gift["attributes"].([]any)
	if len(attrs) > 0 {
		var parts []string
		for _, a := range attrs {
			attr, ok := a.(map[string]any)
			if !ok {
				continue
			}
			name := cliutil.ExtractStringValue(attr, "name")
			if name != "" {
				parts = append(parts, name)
			}
		}
		if len(parts) > 0 {
			line += fmt.Sprintf(" [%s]", strings.Join(parts, ", "))
		}
	}

	line += fmt.Sprintf(" [slug:%s]", slug)
	fmt.Println(line)
}
