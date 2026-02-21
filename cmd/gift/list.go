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
	listLimit       int
	listOffset      string
	listSortByPrice bool
	listSortByNum   bool
	listModel       string
	listPattern     string
	listBackdrop    string
)

// ListCmd represents the gift list command.
var ListCmd = &cobra.Command{
	Use:     "list [peer_or_gift_name]",
	Aliases: []string{"catalog"},
	Short:   "List gift catalog, user's saved gifts, or marketplace",
	Long: `List available star gifts from the Telegram catalog,
list a specific user's saved gifts when a peer is provided,
or browse the resale marketplace by gift name or type ID.`,
	Example: `  agent-telegram gift list
  agent-telegram gift list @username
  agent-telegram gift list Heart
  agent-telegram gift list 5170145012310081536
  agent-telegram gift list Heart --sort-price --model Homunculus`,
	Args: cobra.MaximumNArgs(1),
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntVarP(&listLimit, "limit", "l", cliutil.DefaultLimitLarge, "Max gifts to show")
	ListCmd.Flags().StringVarP(&listOffset, "offset", "o", "", "Offset for pagination")
	ListCmd.Flags().BoolVar(&listSortByPrice, "sort-price", false, "Sort by price (lowest first)")
	ListCmd.Flags().BoolVar(&listSortByNum, "sort-num", false, "Sort by serial number")
	ListCmd.Flags().StringVar(&listModel, "model", "", "Filter by model name")
	ListCmd.Flags().StringVar(&listPattern, "pattern", "", "Filter by pattern name")
	ListCmd.Flags().StringVar(&listBackdrop, "backdrop", "", "Filter by backdrop name")

	ListCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ListCmd, false)

		if len(args) == 0 {
			// No arg → catalog
			runner.SetIDKey("id")
			params := map[string]any{
				"limit": listLimit,
			}
			result := runner.CallWithParams("get_star_gifts", params)
			runner.PrintResult(result, printGiftList)
			return
		}

		arg := args[0]

		if strings.HasPrefix(arg, "@") {
			// Peer → saved gifts
			runner.SetIDKey("id")
			var to cliutil.Recipient
			_ = to.Set(arg)
			params := map[string]any{
				"limit": listLimit,
				"peer":  to.Peer(),
			}
			result := runner.CallWithParams("get_saved_gifts", params)
			runner.PrintResult(result, PrintSavedGifts)
			return
		}

		// Otherwise → marketplace (resale)
		runner.SetIDKey("slug")
		params := map[string]any{
			"limit": listLimit,
		}
		if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
			params["giftId"] = id
		} else {
			params["name"] = arg
		}
		if listOffset != "" {
			params["offset"] = listOffset
		}
		if listSortByPrice {
			params["sortByPrice"] = true
		}
		if listSortByNum {
			params["sortByNum"] = true
		}
		if listModel != "" {
			params["model"] = listModel
		}
		if listPattern != "" {
			params["pattern"] = listPattern
		}
		if listBackdrop != "" {
			params["backdrop"] = listBackdrop
		}
		result := runner.CallWithParams("get_resale_gifts", params)
		runner.PrintResult(result, printResaleGifts)
	}
}

func printGiftList(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get gifts")
		return
	}

	count := cliutil.ExtractInt64(r, "count")
	fmt.Fprintf(os.Stderr, "Star Gifts (%d):\n", count)

	gifts, _ := r["gifts"].([]any)
	for _, g := range gifts {
		gift, ok := g.(map[string]any)
		if !ok {
			continue
		}
		printGiftItem(gift)
	}
}

func printGiftItem(gift map[string]any) {
	id := cliutil.ExtractInt64(gift, "id")
	stars := cliutil.ExtractInt64(gift, "stars")
	title := cliutil.ExtractStringValue(gift, "title")
	slug := cliutil.ExtractStringValue(gift, "slug")
	limited, _ := gift["limited"].(bool)
	soldOut, _ := gift["soldOut"].(bool)
	birthday, _ := gift["birthday"].(bool)
	remains := cliutil.ExtractInt64(gift, "availabilityRemains")
	total := cliutil.ExtractInt64(gift, "availabilityTotal")

	line := fmt.Sprintf("  - #%d", id)
	if title != "" {
		line += fmt.Sprintf(" %s", title)
	}
	if stars > 0 {
		line += fmt.Sprintf(" (%d stars)", stars)
	}
	if limited && total > 0 {
		line += fmt.Sprintf(" [%d/%d left]", remains, total)
	}
	if soldOut {
		line += " [sold out]"
	}
	if birthday {
		line += " [birthday]"
	}
	if slug != "" {
		line += fmt.Sprintf(" [slug:%s]", slug)
	}
	fmt.Fprintln(os.Stderr, line)
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
