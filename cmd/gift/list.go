// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var listLimit int

// ListCmd represents the gift list command.
var ListCmd = &cobra.Command{
	Use:   "list [peer]",
	Short: "List star gifts catalog or a user's saved gifts",
	Long: `List available star gifts from the Telegram catalog,
or list a specific user's saved gifts when a peer is provided.

Example:
  agent-telegram gift list
  agent-telegram gift list @username
  agent-telegram gift list --limit 20`,
	Args: cobra.MaximumNArgs(1),
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntVarP(&listLimit, "limit", "l", cliutil.DefaultLimitLarge, "Max gifts to show")

	ListCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ListCmd, false)

		if len(args) > 0 {
			var to cliutil.Recipient
			_ = to.Set(args[0])
			params := map[string]any{
				"limit": listLimit,
				"peer":  to.Peer(),
			}
			result := runner.CallWithParams("get_saved_gifts", params)
			runner.PrintResult(result, printSavedGifts)
		} else {
			params := map[string]any{
				"limit": listLimit,
			}
			result := runner.CallWithParams("get_star_gifts", params)
			runner.PrintResult(result, printGiftList)
		}
	}
}

func printGiftList(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get gifts")
		return
	}

	count, _ := r["count"].(float64)
	fmt.Fprintf(os.Stderr, "Star Gifts (%d):\n", int(count))

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
	id, _ := gift["id"].(float64)
	stars, _ := gift["stars"].(float64)
	title := cliutil.ExtractStringValue(gift, "title")
	slug := cliutil.ExtractStringValue(gift, "slug")
	limited, _ := gift["limited"].(bool)
	soldOut, _ := gift["soldOut"].(bool)
	birthday, _ := gift["birthday"].(bool)
	remains, _ := gift["availabilityRemains"].(float64)
	total, _ := gift["availabilityTotal"].(float64)

	line := fmt.Sprintf("  - #%d", int64(id))
	if title != "" {
		line += fmt.Sprintf(" %s", title)
	}
	if stars > 0 {
		line += fmt.Sprintf(" (%d stars)", int64(stars))
	}
	if limited && total > 0 {
		line += fmt.Sprintf(" [%d/%d left]", int(remains), int(total))
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
	fmt.Println(line)
}
