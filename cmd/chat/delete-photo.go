// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	deletePhotoPeer  string
)

// DeletePhotoCmd represents the delete-photo command.
var DeletePhotoCmd = &cobra.Command{
	Use:     "delete-photo",
	Short:   "Delete the photo from a chat or channel",
	Long: `Delete the profile photo from a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.

Example:
  agent-telegram delete-photo --peer @mychannel`,
	Args: cobra.NoArgs,
}

// AddDeletePhotoCommand adds the delete-photo command to the root command.
func AddDeletePhotoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeletePhotoCmd)

	DeletePhotoCmd.Flags().StringVarP(&deletePhotoPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	_ = DeletePhotoCmd.MarkFlagRequired("peer")

	DeletePhotoCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(DeletePhotoCmd, true)
		params := map[string]any{
			"peer":  deletePhotoPeer,
		}

		result := runner.CallWithParams("delete_photo", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Photo deleted successfully\n")
			}
		}
	}
}
