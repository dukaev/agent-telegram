// Package send provides commands for sending messages and media.
package send

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

const (
	methodSendMessage = "send_message"
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
)

// SendCmd represents the unified send command.
var SendCmd = &cobra.Command{
	GroupID: "message",
	Use:     "send [message]",
	Short: "Send a message or media to a Telegram peer",
	Long: `Send a message or media to a Telegram user or chat.

By default, sends a text message. Use flags to send other types:

  send --to @user "Hello world"
  send --to @user --photo image.png
  send --to @user --video video.mp4
  send --to @user --voice voice.ogg
  send --to @user --video-note circle.mp4
  send --to @user --sticker sticker.webp
  send --to @user --gif animation.mp4
  send --to @user --document file.pdf
  send --to @user --audio music.mp3
  send --to @user --contact "+1234567890" --first-name "John"
  send --to @user --reply-to 123 "Reply text"
  send --to @user --poll "Question?" --option "Yes" --option "No"
  send --to @user --latitude 55.7558 --longitude 37.6173

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.MaximumNArgs(1),
}

// AddSendCommand adds the unified send command to the root command.
func AddSendCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SendCmd)

	// Register common flags (sets sendFlags.cmd)
	sendFlags.Register(SendCmd)

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

	SendCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendFlags.NewRunner()

		// Determine what type of content to send
		method, params := buildSendParams(args)

		result := runner.CallWithParams(method, params)

		// For send_message, extract and output just the message ID
		if method == methodSendMessage || method == "send_reply" {
			if r, ok := result.(map[string]any); ok {
				if id, ok := r["id"].(float64); ok {
					//nolint:errchkjson // Output to stdout, error handling not required
					_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"id": int64(id)})
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

	// Priority: contact > poll > location > sticker > voice > video-note > gif >
	// photo > video > audio > document > file > message
	switch {
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
			params["message"] = ""
		} else {
			params["message"] = args[0]
		}
		params["reply_to"] = replyToMessageID
		return "send_reply", params

	case len(args) > 0:
		params["message"] = args[0]
		return methodSendMessage, params

	default:
		params["message"] = ""
		return methodSendMessage, params
	}
}
