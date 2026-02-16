package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var stickerFlags SendFlags

// StickerCmd represents the send sticker subcommand.
var StickerCmd = &cobra.Command{
	Use:   "sticker <peer> <file>",
	Short: "Send a sticker",
	Long: `Send a sticker (WEBP) to a Telegram user or chat.

Examples:
  agent-telegram send sticker @user sticker.webp
  agent-telegram send sticker --to @user sticker.webp`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		file := resolvePeerAndFile(&stickerFlags, args)
		stickerFlags.RequirePeer()

		runner := stickerFlags.NewRunner()
		params := map[string]any{"file": file}
		stickerFlags.To.AddToParams(params)

		result := runner.CallWithParams("send_sticker", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_sticker")
		})
	},
}

func addStickerCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(StickerCmd)
	stickerFlags.RegisterOptionalTo(StickerCmd)
}
