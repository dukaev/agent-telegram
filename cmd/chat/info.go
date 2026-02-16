// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var chatInfoTo cliutil.Recipient

// InfoCmd represents the chat info command.
var InfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Get information about a chat or channel",
	Long: `Get detailed information about a Telegram chat or channel.

This returns chat ID, title, username, member count, type, etc.

Use --to @username or --to username to specify the chat/channel.

Example:
  agent-telegram chat info --to @mychannel`,
	Args: cobra.NoArgs,
}

// AddInfoCommand adds the info command to the parent command.
func AddInfoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(InfoCmd)

	InfoCmd.Flags().VarP(&chatInfoTo, "to", "t", "Chat/channel (@username or username)")
	_ = InfoCmd.MarkFlagRequired("to")

	InfoCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, true) // Always JSON

		result := runner.CallWithParams("get_chats", map[string]any{
			"limit":  100,
			"offset": 0,
		})

		// Filter the result to find the matching chat
		filteredResult := filterChatInfo(result, chatInfoTo.Peer())
		runner.PrintResult(filteredResult, nil)
	}
}

// filterChatInfo filters the chats result to find info about a specific peer.
func filterChatInfo(result any, peer string) any {
	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	chats, _ := rMap["chats"].([]any)
	peer = trimPeer(peer)

	for _, chat := range chats {
		chatInfo, ok := chat.(map[string]any)
		if !ok {
			continue
		}

		chatPeer := cliutil.ExtractString(chatInfo, "peer")
		chatPeer = trimPeer(chatPeer)

		if chatPeer == peer {
			// Found the matching chat
			return map[string]any{
				"chat": chatInfo,
			}
		}
	}

	// Not found - return empty result
	return map[string]any{
		"error": fmt.Sprintf("chat not found: %s", peer),
	}
}

// trimPeer removes the @ prefix from a peer string.
func trimPeer(peer string) string {
	if len(peer) > 0 && peer[0] == '@' {
		return peer[1:]
	}
	return peer
}
