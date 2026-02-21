// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// InfoCmd represents the gift info command.
var InfoCmd = &cobra.Command{
	Use:   "info <slug or URL>",
	Short: "Show detailed info and value analytics for a unique gift",
	Long: `Show detailed information about a unique star gift by its slug or URL,
including owner, model, pattern, backdrop, rarity, and value analytics.`,
	Example: `  agent-telegram gift info SwissWatch-718
  agent-telegram gift info https://t.me/nft/RestlessJar-55271`,
	Args: cobra.ExactArgs(1),
}

// AddInfoCommand adds the info command to the parent command.
func AddInfoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(InfoCmd)

	InfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, false)
		slug := cliutil.ParseGiftSlug(args[0])
		params := map[string]any{
			"slug": slug,
		}

		infoResult := runner.CallWithParams("get_gift_info", params)

		// Fetch value analytics (may not be available for all gifts)
		valueResult, valueErr := runner.Client().Call("get_gift_value", params)

		// Merge value into info result for JSON output
		if valueErr == nil && valueResult != nil {
			if infoMap, ok := infoResult.(map[string]any); ok {
				infoMap["value"] = valueResult
			}
		}

		runner.PrintResult(infoResult, func(r any) {
			PrintGiftInfo(r)
			if valueErr == nil && valueResult != nil {
				fmt.Fprintln(os.Stderr)
				printGiftValue(valueResult)
			}
			printGiftInfoHints(slug)
		})
	}
}

func printGiftInfoHints(slug string) {
	fmt.Fprintln(os.Stderr, "\nRelated commands:")
	fmt.Fprintf(os.Stderr, "  gift attrs %s     # attributes for this gift type\n", slug)
	fmt.Fprintf(os.Stderr, "  gift buy %s       # buy from marketplace\n", slug)
	fmt.Fprintf(os.Stderr, "  gift offer %s     # make an offer to owner\n", slug)
	fmt.Fprintf(os.Stderr, "  gift price %s     # set resale price\n", slug)
	fmt.Fprintf(os.Stderr, "  gift send %s --to @user  # transfer to another user\n", slug)
}

//nolint:funlen // Printing many fields requires many statements
func PrintGiftInfo(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get gift info")
		return
	}

	title := cliutil.ExtractStringValue(r, "title")
	slug := cliutil.ExtractStringValue(r, "slug")
	num := cliutil.ExtractInt64(r, "num")
	issued := cliutil.ExtractInt64(r, "availabilityIssued")
	total := cliutil.ExtractInt64(r, "availabilityTotal")
	ownerName := cliutil.ExtractStringValue(r, "ownerName")
	ownerAddr := cliutil.ExtractStringValue(r, "ownerAddress")
	giftAddr := cliutil.ExtractStringValue(r, "giftAddress")
	resell := cliutil.ExtractInt64(r, "resellStars")
	giftID := cliutil.ExtractInt64(r, "giftId")

	fmt.Fprintf(os.Stderr, "%s [%s] #%d of %d\n", title, slug, num, total)
	fmt.Fprintf(os.Stderr, "Gift ID: %d\n", giftID)
	fmt.Fprintf(os.Stderr, "Issued: %d / %d\n", issued, total)

	ownerID := cliutil.ExtractStringValue(r, "ownerId")
	if ownerName != "" {
		line := fmt.Sprintf("Owner: %s", ownerName)
		if ownerID != "" {
			line += fmt.Sprintf(" (ID: %s)", ownerID)
		}
		fmt.Fprintln(os.Stderr, line)
	} else if ownerID != "" {
		fmt.Fprintf(os.Stderr, "Owner ID: %s\n", ownerID)
	}
	if ownerAddr != "" {
		fmt.Fprintf(os.Stderr, "Owner address: %s\n", ownerAddr)
	}
	if giftAddr != "" {
		fmt.Fprintf(os.Stderr, "Gift address: %s\n", giftAddr)
	}
	if resell > 0 {
		fmt.Fprintf(os.Stderr, "Resale price: %d stars\n", resell)
	}

	attrs, _ := r["attributes"].([]any)
	if len(attrs) > 0 {
		fmt.Fprintln(os.Stderr, "Attributes:")
		for _, a := range attrs {
			attr, ok := a.(map[string]any)
			if !ok {
				continue
			}
			attrType := cliutil.ExtractStringValue(attr, "type")
			name := cliutil.ExtractStringValue(attr, "name")
			rarity := cliutil.ExtractFloat64(attr, "rarityPermille")

			line := fmt.Sprintf("  - %s: %s", attrType, name)
			if rarity > 0 {
				line += fmt.Sprintf(" (%.1f%%)", rarity/10)
			}
			fmt.Fprintln(os.Stderr, line)
		}
	}
}

//nolint:funlen // Printing many fields requires many statements
func printGiftValue(result any) {
	r, ok := result.(map[string]any)
	if !ok {
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
