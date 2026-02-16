package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var videoFlags SendFlags

// VideoCmd represents the send video subcommand.
var VideoCmd = &cobra.Command{
	Use:   "video <peer> <file>",
	Short: "Send a video",
	Long: `Send a video to a Telegram user or chat.

Examples:
  agent-telegram send video @user video.mp4
  agent-telegram send video --to @user video.mp4 --caption "Check this out"`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		file := resolvePeerAndFile(&videoFlags, args)
		videoFlags.RequirePeer()

		runner := videoFlags.NewRunner()
		params := map[string]any{"file": file}
		videoFlags.AddToParams(params)

		result := runner.CallWithParams("send_video", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_video")
		})
	},
}

func addVideoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(VideoCmd)
	videoFlags.RegisterOptionalTo(VideoCmd)
}
