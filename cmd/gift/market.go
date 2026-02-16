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
	Use:   "market <gift_name_or_id>",
	Short: "List gifts for resale (marketplace)",
	Long: `Browse star gifts listed for resale by gift name or type ID.
Use "gift list" to find gift names from the catalog.`,
	Example: `  agent-telegram gift market Heart
  agent-telegram gift market 5170145012310081536
  agent-telegram gift market Heart --sort-price --limit 10
  agent-telegram gift market Heart --model "Homunculus" --backdrop "Turquoise"`,
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
		runner.SetIDKey("slug")
		params := map[string]any{
			"limit": marketLimit,
		}
		// Try parsing as numeric ID, otherwise treat as name
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			params["giftId"] = id
		} else {
			params["name"] = args[0]
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

func printResaleGifts(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get resale gifts")
		return
	}

	count := cliutil.ExtractInt64(r, "count")
	fmt.Fprintf(os.Stderr, "Resale Listings (%d):\n", count)

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
	num := cliutil.ExtractInt64(gift, "num")
	resell := cliutil.ExtractInt64(gift, "resellStars")
	owner := cliutil.ExtractStringValue(gift, "ownerName")

	line := fmt.Sprintf("  - %s #%d", title, num)
	if resell > 0 {
		line += fmt.Sprintf(" — %d stars", resell)
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
	fmt.Fprintln(os.Stderr, line)
}
