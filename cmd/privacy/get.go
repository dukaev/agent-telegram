// Package privacy provides commands for managing privacy settings.
package privacy

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var getKey string

// GetCmd represents the privacy get command.
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get privacy setting",
	Long: `Get privacy setting for a specific key.

Keys: status_timestamp, phone_number, profile_photo, forwards,
      phone_call, phone_p2p, voice_messages, about

Example:
  agent-telegram privacy get --key phone_number
  agent-telegram privacy get --key status_timestamp`,
	Args: cobra.NoArgs,
}

// AddGetCommand adds the get command to the parent command.
func AddGetCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(GetCmd)

	GetCmd.Flags().StringVarP(&getKey, "key", "k", "", "Privacy key")
	_ = GetCmd.MarkFlagRequired("key")

	GetCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(GetCmd, true)
		params := map[string]any{
			"key": getKey,
		}

		result := runner.CallWithParams("get_privacy", params)
		runner.PrintResult(result, func(r any) {
			if m, ok := r.(map[string]any); ok {
				if rules, ok := m["rules"].([]any); ok {
					fmt.Fprintf(os.Stderr, "Privacy rules for '%s':\n", getKey)
					for _, rule := range rules {
						if ruleMap, ok := rule.(map[string]any); ok {
							ruleType := cliutil.ExtractString(ruleMap, "type")
							fmt.Fprintf(os.Stderr, "  - %s\n", ruleType)
						}
					}
				}
			}
		})
	}
}
