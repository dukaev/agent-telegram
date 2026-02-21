// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// InfoCmd represents the gift info command.
var InfoCmd = &cobra.Command{
	Use:   "info <slug or URL>",
	Short: "Show detailed info about a unique gift",
	Long: `Show detailed information about a unique star gift by its slug or URL,
including owner, model, pattern, backdrop, and rarity.`,
	Example: `  agent-telegram gift info SwissWatch-718
  agent-telegram gift info https://t.me/nft/RestlessJar-55271`,
	Args: cobra.ExactArgs(1),
}

// AddInfoCommand adds the info command to the parent command.
func AddInfoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(InfoCmd)

	InfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, false)
		params := map[string]any{
			"slug": cliutil.ParseGiftSlug(args[0]),
		}
		result := runner.CallWithParams("get_gift_info", params)
		runner.PrintResult(result, PrintGiftInfo)
	}
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
