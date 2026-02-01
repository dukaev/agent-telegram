//nolint:dupl // Similar command structure is intentional
package chat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	setPhotoPeer  string
	setPhotoFile  string
)

// SetPhotoCmd represents the set-photo command.
var SetPhotoCmd = &cobra.Command{
	Use:     "set-photo",
	Short:   "Set the photo for a chat or channel",
	Long: `Set the profile photo for a Telegram chat or channel.

Use --peer @username or --peer username to specify the chat/channel.
Use --file to specify the photo file path.

Example:
  agent-telegram set-photo --peer @mychannel --file photo.jpg`,
	Args: cobra.NoArgs,
}

// AddSetPhotoCommand adds the set-photo command to the root command.
func AddSetPhotoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SetPhotoCmd)

	SetPhotoCmd.Flags().StringVarP(&setPhotoPeer, "peer", "p", "", "Chat/channel username (@username or username)")
	SetPhotoCmd.Flags().StringVarP(&setPhotoFile, "file", "f", "", "Photo file path")
	_ = SetPhotoCmd.MarkFlagRequired("peer")
	_ = SetPhotoCmd.MarkFlagRequired("file")

	SetPhotoCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(SetPhotoCmd, true)
		params := map[string]any{
			"peer":  setPhotoPeer,
			"file":  setPhotoFile,
		}

		result := runner.CallWithParams("set_photo", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "Photo set successfully\n")
			}
		}
	}
}
