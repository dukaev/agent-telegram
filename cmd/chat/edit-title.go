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
	editTitlePeer  string
	editTitleTitle string
)

// EditTitleCmd represents the edit-title command.
var EditTitleCmd = &cobra.Command{
	Use:     "edit-title",
	Short:   "Edit the title of a chat or channel",
	Long: `Edit the title of a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --title to set the new title.

Example:
  agent-telegram edit-title --peer @mychannel --title "New Title"`,
	Args: cobra.NoArgs,
}

// AddEditTitleCommand adds the edit-title command to the root command.
func AddEditTitleCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(EditTitleCmd)

	EditTitleCmd.Flags().StringVarP(&editTitlePeer, "peer", "p", "", "Chat/channel username (@username or username)")
	EditTitleCmd.Flags().StringVarP(&editTitleTitle, "title", "t", "", "New title")
	_ = EditTitleCmd.MarkFlagRequired("peer")
	_ = EditTitleCmd.MarkFlagRequired("title")

	EditTitleCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(EditTitleCmd, true)
		params := map[string]any{
			"peer":  editTitlePeer,
			"title": editTitleTitle,
		}

		result := runner.CallWithParams("edit_title", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Title updated successfully\n")
				if title, ok := r["title"].(string); ok {
					fmt.Fprintf(os.Stderr, "  New title: %s\n", title)
				}
			}
		}
	}
}
