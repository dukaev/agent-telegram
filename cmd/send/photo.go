package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var photoFlags SendFlags

// PhotoCmd represents the send photo subcommand.
var PhotoCmd = &cobra.Command{
	Use:   "photo <peer> <file>",
	Short: "Send a photo",
	Long: `Send a photo to a Telegram user or chat.

Examples:
  agent-telegram send photo @user image.png
  agent-telegram send photo --to @user image.png --caption "Nice photo"`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		file := resolvePeerAndFile(&photoFlags, args)
		photoFlags.RequirePeer()

		runner := photoFlags.NewRunner()
		params := map[string]any{"file": file}
		photoFlags.AddToParams(params)

		result := runner.CallWithParams("send_photo", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_photo")
		})
	},
}

func addPhotoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(PhotoCmd)
	photoFlags.RegisterOptionalTo(PhotoCmd)
}
