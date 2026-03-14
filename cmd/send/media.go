package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// mediaCommandDef defines a media send subcommand.
type mediaCommandDef struct {
	Name    string
	Short   string
	Long    string
	Method  string
}

var mediaCommands = []mediaCommandDef{
	{Name: "photo", Short: "Send a photo", Method: "send_photo",
		Long: "Send a photo to a Telegram user or chat.\n\nExamples:\n  agent-telegram send photo @user image.png\n  agent-telegram send photo --to @user image.png --caption \"Nice photo\""},
	{Name: "video", Short: "Send a video", Method: "send_video",
		Long: "Send a video to a Telegram user or chat.\n\nExamples:\n  agent-telegram send video @user video.mp4\n  agent-telegram send video --to @user video.mp4 --caption \"Check this out\""},
	{Name: "voice", Short: "Send a voice message", Method: "send_voice",
		Long: "Send a voice message (OGG/OPUS) to a Telegram user or chat.\n\nExamples:\n  agent-telegram send voice @user audio.ogg\n  agent-telegram send voice --to @user audio.ogg"},
	{Name: "sticker", Short: "Send a sticker", Method: "send_sticker",
		Long: "Send a sticker (WEBP) to a Telegram user or chat.\n\nExamples:\n  agent-telegram send sticker @user sticker.webp\n  agent-telegram send sticker --to @user sticker.webp"},
}

// Exported command vars for schema registration.
var (
	PhotoCmd   *cobra.Command
	VideoCmd   *cobra.Command
	VoiceCmd   *cobra.Command
	StickerCmd *cobra.Command
)

func addMediaCommands(parentCmd *cobra.Command) {
	cmdMap := map[string]**cobra.Command{
		"photo": &PhotoCmd, "video": &VideoCmd,
		"voice": &VoiceCmd, "sticker": &StickerCmd,
	}
	for _, def := range mediaCommands {
		cmd, flags := newMediaCommand(def)
		parentCmd.AddCommand(cmd)
		flags.RegisterOptionalTo(cmd)
		if ptr, ok := cmdMap[def.Name]; ok {
			*ptr = cmd
		}
	}
}

func newMediaCommand(def mediaCommandDef) (*cobra.Command, *SendFlags) {
	flags := &SendFlags{}
	cmd := &cobra.Command{
		Use:   def.Name + " <peer> <file>",
		Short: def.Short,
		Long:  def.Long,
		Args:  cobra.RangeArgs(1, 2),
		Run: func(_ *cobra.Command, args []string) {
			file := resolvePeerAndFile(flags, args)
			flags.RequirePeer()

			runner := flags.NewRunner()
			params := map[string]any{"file": file}
			flags.To.AddToParams(params)
			if flags.Caption != "" {
				params["caption"] = flags.Caption
			}

			result := runner.CallWithParams(def.Method, params)
			runner.PrintResult(result, func(r any) {
				cliutil.FormatSuccess(r, def.Method)
			})
		},
	}
	return cmd, flags
}
