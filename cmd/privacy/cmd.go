// Package privacy provides commands for managing privacy settings.
package privacy

import (
	"github.com/spf13/cobra"
)

// PrivacyCmd represents the privacy command group.
var PrivacyCmd = &cobra.Command{
	Use:   "privacy",
	Short: "Manage privacy settings",
	Long: `Commands for managing Telegram privacy settings.

Available privacy keys:
  status_timestamp - Who can see your last seen
  phone_number     - Who can see your phone number
  profile_photo    - Who can see your profile photo
  forwards         - Who can forward your messages
  phone_call       - Who can call you
  phone_p2p        - Peer-to-peer in calls
  voice_messages   - Who can send you voice messages
  about            - Who can see your bio`,
}

// AddPrivacyCommand adds the privacy command group to the root command.
func AddPrivacyCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PrivacyCmd)

	AddGetCommand(PrivacyCmd)
	AddSetCommand(PrivacyCmd)
}
