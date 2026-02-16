// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	myLimit  int
	myOffset string
	myTo     cliutil.Recipient
)

// MyCmd represents the gift my command.
var MyCmd = &cobra.Command{
	Use:   "my",
	Short: "List saved star gifts",
	Long: `List saved/received star gifts for yourself or another user.

Example:
  agent-telegram gift my
  agent-telegram gift my --limit 20
  agent-telegram gift my --to @username`,
	Args: cobra.NoArgs,
}

// AddMyCommand adds the my command to the parent command.
func AddMyCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(MyCmd)

	MyCmd.Flags().IntVarP(&myLimit, "limit", "l", cliutil.DefaultLimitLarge, "Max gifts to show")
	MyCmd.Flags().StringVarP(&myOffset, "offset", "o", "", "Offset for pagination")
	MyCmd.Flags().VarP(&myTo, "to", "t", "User whose gifts to view (@username or ID)")

	MyCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(MyCmd, false)
		params := map[string]any{
			"limit": myLimit,
		}
		if myOffset != "" {
			params["offset"] = myOffset
		}
		if myTo.Peer() != "" {
			params["peer"] = myTo.Peer()
		}
		result := runner.CallWithParams("get_saved_gifts", params)
		runner.PrintResult(result, printSavedGifts)
	}
}

func printSavedGifts(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get saved gifts")
		return
	}

	count := cliutil.ExtractInt64(r, "count")
	fmt.Fprintf(os.Stderr, "Saved Gifts (%d):\n", count)

	gifts, _ := r["gifts"].([]any)
	var totalStars, uniqueCount, listedCount int64
	for _, g := range gifts {
		gift, ok := g.(map[string]any)
		if !ok {
			continue
		}
		printSavedGiftItem(gift)

		totalStars += cliutil.ExtractInt64(gift, "stars")
		if cliutil.ExtractStringValue(gift, "slug") != "" {
			uniqueCount++
		}
		if cliutil.ExtractInt64(gift, "resellStars") > 0 {
			listedCount++
		}
	}

	if nextOffset := cliutil.ExtractStringValue(r, "nextOffset"); nextOffset != "" {
		fmt.Fprintf(os.Stderr, "\nNext offset: %s\n", nextOffset)
	}

	fmt.Fprintf(os.Stderr, "\nSummary: %d gifts, %d stars value, %d unique, %d listed for resale\n",
		count, totalStars, uniqueCount, listedCount)
}

func printSavedGiftItem(gift map[string]any) {
	giftID := cliutil.ExtractInt64(gift, "giftId")
	msgID := cliutil.ExtractInt64(gift, "msgId")
	stars := cliutil.ExtractInt64(gift, "stars")
	convertStars := cliutil.ExtractInt64(gift, "convertStars")
	resellStars := cliutil.ExtractInt64(gift, "resellStars")
	title := cliutil.ExtractStringValue(gift, "title")
	fromID := cliutil.ExtractStringValue(gift, "fromId")
	slug := cliutil.ExtractStringValue(gift, "slug")

	line := fmt.Sprintf("  - #%d", giftID)
	if title != "" {
		line += fmt.Sprintf(" %s", title)
	}
	if stars > 0 {
		line += fmt.Sprintf(" (%d stars)", stars)
	}
	if convertStars > 0 {
		line += fmt.Sprintf(" convert: %d", convertStars)
	}
	transferStars := cliutil.ExtractInt64(gift, "transferStars")
	if transferStars > 0 {
		line += fmt.Sprintf(" transfer: %d", transferStars)
	}
	if resellStars > 0 {
		line += fmt.Sprintf(" price: %d stars", resellStars)
	}
	if fromID != "" {
		line += fmt.Sprintf(" from %s", fromID)
	}
	if slug != "" {
		line += fmt.Sprintf(" [slug:%s]", slug)
	}
	if msgID > 0 {
		line += fmt.Sprintf(" [msg:%d]", msgID)
	}
	fmt.Fprintln(os.Stderr, line)
}
