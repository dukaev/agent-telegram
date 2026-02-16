// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

const (
	methodSendMessage = "send_message"
	methodSendReply   = "send_reply"
)

var (
	sendFlags SendFlags
	// File to send
	sendFile     string
	sendPhoto    string
	sendVideo    string
	sendAudio    string
	sendDocument string
	// New media types
	sendVoice     string
	sendVideoNote string
	sendSticker   string
	sendGIF       string
	// Reply options
	replyToMessageID int64
	// Poll options
	pollQuestion string
	pollOptions  []string
	// Location
	latitude  float64
	longitude float64
	// Contact
	sendContact   string
	sendFirstName string
	sendLastName  string
	// Dice
	sendDice bool
	diceEmoticon string
)

// SendCmd represents the unified send command.
var SendCmd = &cobra.Command{
	GroupID: "message",
	Use:     "send [peer] [message]",
	Short:   "Send a message or media to a Telegram peer",
	Long: `Send a message or media to a Telegram user or chat.
Peer can be positional or via --to flag.

By default, sends a text message. Use flags to send other types.
Use @username, username, or <chat_id> to specify the recipient.`,
	Example: `  agent-telegram send @user "Hello world"
  agent-telegram send --to @user "Hello world"
  agent-telegram send @user --photo image.png
  agent-telegram send @user --poll "Question?" --option "Yes" --option "No"`,
	Args: cobra.MaximumNArgs(2),
}

// AddSendCommand adds the unified send command to the root command.
//
//nolint:funlen // Function registers many flags and handles peer resolution
func AddSendCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SendCmd)

	// Register subcommands
	addTextCommand(SendCmd)
	addPhotoCommand(SendCmd)
	addVideoCommand(SendCmd)
	addVoiceCommand(SendCmd)
	addDiceCommand(SendCmd)
	addPollCommand(SendCmd)
	addLocationCommand(SendCmd)
	addContactCommand(SendCmd)
	addStickerCommand(SendCmd)

	// Register common flags with optional --to (positional peer supported)
	sendFlags.RegisterOptionalTo(SendCmd)

	// Content type flags (mutually exclusive)
	SendCmd.Flags().StringVar(&sendFile, "file", "", "Send file (auto-detect type)")
	SendCmd.Flags().StringVar(&sendPhoto, "photo", "", "Send photo")
	SendCmd.Flags().StringVar(&sendVideo, "video", "", "Send video")
	SendCmd.Flags().StringVar(&sendAudio, "audio", "", "Send audio")
	SendCmd.Flags().StringVar(&sendDocument, "document", "", "Send document")
	SendCmd.Flags().StringVar(&sendVoice, "voice", "", "Send voice message (OGG/OPUS)")
	SendCmd.Flags().StringVar(&sendVideoNote, "video-note", "", "Send video note (circle)")
	SendCmd.Flags().StringVar(&sendSticker, "sticker", "", "Send sticker (WEBP)")
	SendCmd.Flags().StringVar(&sendGIF, "gif", "", "Send GIF/animation")

	// Reply flag
	SendCmd.Flags().Int64Var(&replyToMessageID, "reply-to", 0, "Reply to message ID")

	// Poll flags
	SendCmd.Flags().StringVar(&pollQuestion, "poll", "", "Create poll with question")
	SendCmd.Flags().StringSliceVar(&pollOptions, "option", nil, "Poll options (can be used multiple times)")

	// Location flags
	SendCmd.Flags().Float64Var(&latitude, "latitude", 0, "Send location latitude")
	SendCmd.Flags().Float64Var(&longitude, "longitude", 0, "Send location longitude")

	// Contact flag
	SendCmd.Flags().StringVar(&sendContact, "contact", "", "Send contact (phone number)")
	SendCmd.Flags().StringVar(&sendFirstName, "first-name", "", "Contact first name")
	SendCmd.Flags().StringVar(&sendLastName, "last-name", "", "Contact last name")

	// Dice flag
	SendCmd.Flags().BoolVar(&sendDice, "dice", false, "Send dice (random value)")
	SendCmd.Flags().StringVar(&diceEmoticon, "emoticon", "", "Dice emoticon (default: 🎲, also: 🎯, 🏀, ⚽, 🎳, 🎰)")

	SendCmd.Run = func(_ *cobra.Command, args []string) {
		// Resolve peer: positional arg or --to flag
		// 2 args: args[0]=peer, args[1]=message
		// 1 arg + --to set: args[0]=message
		// 1 arg + --to NOT set: args[0]=peer (no message)
		// 0 args: --to must be set
		var messageText string
		stdinText := cliutil.ReadStdinIfPiped()

		switch len(args) {
		case 2:
			_ = sendFlags.To.Set(args[0])
			messageText = args[1] // positional arg wins over stdin
		case 1:
			if sendFlags.To.Peer() != "" {
				messageText = args[0] // positional arg wins over stdin
			} else {
				_ = sendFlags.To.Set(args[0])
				messageText = stdinText // no positional message, use stdin
			}
		default:
			messageText = stdinText // no positional args, use stdin
		}

		if sendFlags.To.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := sendFlags.NewRunner()

		// Build messageArgs for buildSendParams
		var messageArgs []string
		if messageText != "" {
			messageArgs = []string{messageText}
		}

		// Determine what type of content to send
		method, params := buildSendParams(messageArgs)

		result := runner.CallWithParams(method, params)

		// For send_message, extract and output just the message ID
		if method == methodSendMessage || method == methodSendReply {
			if r, ok := result.(map[string]any); ok {
				if id, ok := r["id"].(float64); ok {
					runner.PrintResult(map[string]any{"id": int64(id)}, func(r any) {
						cliutil.FormatSuccess(r, method)
					})
					return
				}
			}
		}

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, method)
		})
	}
}

// buildSendParams determines the method and parameters based on flags and args.
//
//nolint:funlen // Function requires handling multiple content types
func buildSendParams(args []string) (string, map[string]any) {
	params := make(map[string]any)

	// Always add recipient
	sendFlags.To.AddToParams(params)

	// Add caption if present
	if sendFlags.Caption != "" {
		params["caption"] = sendFlags.Caption
	}

	// Priority: dice > contact > poll > location > sticker > voice > video-note > gif >
	// photo > video > audio > document > file > message
	switch {
	case sendDice:
		if diceEmoticon != "" {
			params["emoticon"] = diceEmoticon
		}
		if replyToMessageID != 0 {
			params["replyTo"] = replyToMessageID
		}
		return "send_dice", params

	case sendContact != "":
		params["phone"] = sendContact
		if sendFirstName != "" {
			params["firstName"] = sendFirstName
		}
		if sendLastName != "" {
			params["lastName"] = sendLastName
		}
		return "send_contact", params

	case pollQuestion != "":
		params["question"] = pollQuestion
		if len(pollOptions) > 0 {
			// Convert []string to []map[string]string for PollOption
			optionMaps := make([]map[string]string, len(pollOptions))
			for i, opt := range pollOptions {
				optionMaps[i] = map[string]string{"text": opt}
			}
			params["options"] = optionMaps
		}
		return "send_poll", params

	case latitude != 0 && longitude != 0:
		params["latitude"] = latitude
		params["longitude"] = longitude
		return "send_location", params

	case sendSticker != "":
		params["file"] = sendSticker
		return "send_sticker", params

	case sendVoice != "":
		params["file"] = sendVoice
		return "send_voice", params

	case sendVideoNote != "":
		params["file"] = sendVideoNote
		return "send_video_note", params

	case sendGIF != "":
		params["file"] = sendGIF
		return "send_gif", params

	case sendPhoto != "":
		params["file"] = sendPhoto
		return "send_photo", params

	case sendVideo != "":
		params["file"] = sendVideo
		return "send_video", params

	case sendAudio != "":
		params["file"] = sendAudio
		return "send_audio", params

	case sendDocument != "":
		params["file"] = sendDocument
		return "send_document", params

	case sendFile != "":
		params["file"] = sendFile
		return "send_file", params

	case replyToMessageID != 0:
		if len(args) == 0 {
			params["text"] = ""
		} else {
			params["text"] = args[0]
		}
		params["messageId"] = replyToMessageID
		return methodSendReply, params

	case len(args) > 0:
		params["message"] = args[0]
		return methodSendMessage, params

	default:
		params["message"] = ""
		return methodSendMessage, params
	}
}
