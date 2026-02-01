// Package info provides commands for getting chat information.
package info

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	infoPeer string
)

// InfoCmd represents the info command.
var InfoCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "info --peer @username",
	Short:   "Get information about a chat or channel",
	Long: `Get detailed information about a Telegram chat or channel.

This returns chat ID, title, username, member count, type, etc.

Use --peer @username or --peer username to specify the chat/channel.

Example:
  agent-telegram info --peer @mychannel`,
	Args: cobra.NoArgs,
}

// AddInfoCommand adds the info command to the root command.
func AddInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InfoCmd)

	InfoCmd.Flags().StringVarP(&infoPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	_ = InfoCmd.MarkFlagRequired("peer")

	InfoCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, true)

		result := runner.CallWithParams("get_chats", map[string]any{
			"limit":  100,
			"offset": 0,
		})

		// Filter the result to find the matching chat
		filteredResult := filterChatInfo(result, infoPeer)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(filteredResult)
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
