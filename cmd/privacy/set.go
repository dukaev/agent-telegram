// Package privacy provides commands for managing privacy settings.
package privacy

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	setKey  string
	setRule string
)

// SetCmd represents the privacy set command.
var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set privacy setting",
	Long: `Set privacy setting for a specific key.

Keys: status_timestamp, phone_number, profile_photo, forwards,
      phone_call, phone_p2p, voice_messages, about

Rules: allow_all, allow_contacts, disallow_all, allow_close_friends

Example:
  agent-telegram privacy set --key phone_number --rule allow_contacts
  agent-telegram privacy set --key status_timestamp --rule disallow_all`,
	Args: cobra.NoArgs,
}

// AddSetCommand adds the set command to the parent command.
func AddSetCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(SetCmd)

	SetCmd.Flags().StringVarP(&setKey, "key", "k", "", "Privacy key")
	SetCmd.Flags().StringVarP(&setRule, "rule", "r", "", "Privacy rule")
	_ = SetCmd.MarkFlagRequired("key")
	_ = SetCmd.MarkFlagRequired("rule")

	SetCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(SetCmd, true)
		params := map[string]any{
			"key": setKey,
			"rules": []map[string]any{
				{"type": setRule},
			},
		}

		result := runner.CallWithParams("set_privacy", params)
		runner.PrintResult(result, func(any) {
			cliutil.PrintSuccessSummary(result, "Privacy setting updated")
		})
	}
}
