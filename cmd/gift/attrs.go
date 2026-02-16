// Package gift provides commands for managing star gifts.
package gift

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var attrsType string

// AttrsCmd represents the gift attrs command.
var AttrsCmd = &cobra.Command{
	Use:   "attrs <gift_name_or_id>",
	Short: "List available attributes for a gift type",
	Long: `List all available models, patterns, and backdrops for a gift type.
Use --type to show only one attribute category.`,
	Example: `  agent-telegram gift attrs Heart
  agent-telegram gift attrs 5170145012310081536
  agent-telegram gift attrs Heart --type backdrop
  agent-telegram gift attrs Heart --type model`,
	Args: cobra.ExactArgs(1),
}

// AddAttrsCommand adds the attrs command to the parent command.
func AddAttrsCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(AttrsCmd)

	AttrsCmd.Flags().StringVar(&attrsType, "type", "", "Show only one type: model, pattern, backdrop")

	AttrsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(AttrsCmd, false)
		runner.SetIDKey("name")
		params := map[string]any{}
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			params["giftId"] = id
		} else {
			params["name"] = args[0]
		}
		result := runner.CallWithParams("get_gift_attrs", params)
		runner.PrintResult(result, func(r any) {
			printGiftAttrs(r, attrsType)
		})
	}
}

func printGiftAttrs(result any, filterType string) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to get gift attributes")
		return
	}

	if filterType == "" || filterType == "model" {
		printAttrGroup(r, "models", "Models")
	}
	if filterType == "" || filterType == "pattern" {
		printAttrGroup(r, "patterns", "Patterns")
	}
	if filterType == "" || filterType == "backdrop" {
		printAttrGroup(r, "backdrops", "Backdrops")
	}
}

func printAttrGroup(r map[string]any, key, title string) {
	items, _ := r[key].([]any)
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "%s (%d):\n", title, len(items))
	for _, item := range items {
		attr, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := cliutil.ExtractStringValue(attr, "name")
		rarity := cliutil.ExtractFloat64(attr, "rarityPermille")

		line := fmt.Sprintf("  - %s", name)
		if rarity > 0 {
			line += fmt.Sprintf(" (%.1f%%)", rarity/10)
		}
		fmt.Fprintln(os.Stderr, line)
	}
}
