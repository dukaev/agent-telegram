// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	permPeer           string
	permSendMessages   bool
	permSendMedia      bool
	permSendStickers   bool
	permSendGifs       bool
	permSendPolls      bool
	permEmbedLinks     bool
	permChangeInfo     bool
	permInviteUsers    bool
	permPinMessages    bool
	permSendPhotos     bool
	permSendVideos     bool
	permSendAudios     bool
	permSendVoices     bool
	permSendDocs       bool
)

// PermissionsCmd represents the permissions command.
var PermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Set default chat permissions",
	Long: `Set default permissions for all members in a chat or channel.

All flags default to false (restricted). Use flags to allow specific actions.

Example:
  agent-telegram chat permissions --peer @mygroup --send-messages --send-media
  agent-telegram chat permissions --peer @mygroup --send-messages --send-photos --send-videos`,
	Args: cobra.NoArgs,
}

// AddPermissionsCommand adds the permissions command to the root command.
func AddPermissionsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PermissionsCmd)

	PermissionsCmd.Flags().StringVarP(&permPeer, "peer", "p", "", "Chat/channel")
	PermissionsCmd.Flags().BoolVar(&permSendMessages, "send-messages", false, "Allow sending messages")
	PermissionsCmd.Flags().BoolVar(&permSendMedia, "send-media", false, "Allow sending media")
	PermissionsCmd.Flags().BoolVar(&permSendStickers, "send-stickers", false, "Allow sending stickers")
	PermissionsCmd.Flags().BoolVar(&permSendGifs, "send-gifs", false, "Allow sending GIFs")
	PermissionsCmd.Flags().BoolVar(&permSendPolls, "send-polls", false, "Allow sending polls")
	PermissionsCmd.Flags().BoolVar(&permEmbedLinks, "embed-links", false, "Allow embedding links")
	PermissionsCmd.Flags().BoolVar(&permChangeInfo, "change-info", false, "Allow changing info")
	PermissionsCmd.Flags().BoolVar(&permInviteUsers, "invite-users", false, "Allow inviting users")
	PermissionsCmd.Flags().BoolVar(&permPinMessages, "pin-messages", false, "Allow pinning messages")
	PermissionsCmd.Flags().BoolVar(&permSendPhotos, "send-photos", false, "Allow sending photos")
	PermissionsCmd.Flags().BoolVar(&permSendVideos, "send-videos", false, "Allow sending videos")
	PermissionsCmd.Flags().BoolVar(&permSendAudios, "send-audios", false, "Allow sending audios")
	PermissionsCmd.Flags().BoolVar(&permSendVoices, "send-voices", false, "Allow sending voice messages")
	PermissionsCmd.Flags().BoolVar(&permSendDocs, "send-docs", false, "Allow sending documents")
	_ = PermissionsCmd.MarkFlagRequired("peer")

	PermissionsCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(PermissionsCmd, true)
		params := map[string]any{
			"peer":         permPeer,
			"sendMessages": permSendMessages,
			"sendMedia":    permSendMedia,
			"sendStickers": permSendStickers,
			"sendGifs":     permSendGifs,
			"sendPolls":    permSendPolls,
			"embedLinks":   permEmbedLinks,
			"changeInfo":   permChangeInfo,
			"inviteUsers":  permInviteUsers,
			"pinMessages":  permPinMessages,
			"sendPhotos":   permSendPhotos,
			"sendVideos":   permSendVideos,
			"sendAudios":   permSendAudios,
			"sendVoices":   permSendVoices,
			"sendDocs":     permSendDocs,
		}

		result := runner.CallWithParams("set_chat_permissions", params)
		//nolint:errchkjson // Output to stdout
		_ = json.NewEncoder(os.Stdout).Encode(result)
		cliutil.PrintSuccessSummary(result, "Permissions updated")
	}
}
