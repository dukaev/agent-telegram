package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var voiceFlags SendFlags

// VoiceCmd represents the send voice subcommand.
var VoiceCmd = &cobra.Command{
	Use:   "voice <peer> <file>",
	Short: "Send a voice message",
	Long: `Send a voice message (OGG/OPUS) to a Telegram user or chat.

Examples:
  agent-telegram send voice @user audio.ogg
  agent-telegram send voice --to @user audio.ogg`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		file := resolvePeerAndFile(&voiceFlags, args)
		voiceFlags.RequirePeer()

		runner := voiceFlags.NewRunner()
		params := map[string]any{"file": file}
		voiceFlags.To.AddToParams(params)

		result := runner.CallWithParams("send_voice", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_voice")
		})
	},
}

func addVoiceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(VoiceCmd)
	voiceFlags.RegisterOptionalTo(VoiceCmd)
}
