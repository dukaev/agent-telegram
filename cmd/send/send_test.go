package send

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddSendCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddSendCommand(root)

	sendCmd := childCommand(root, "send")
	if sendCmd == nil {
		t.Fatal("send command was not registered")
	}
	for _, name := range []string{"text", "photo", "video", "voice", "sticker", "dice", "poll", "location", "contact"} {
		if childCommand(sendCmd, name) == nil {
			t.Fatalf("send subcommand %q was not registered", name)
		}
	}
	for _, flag := range []string{"to", "caption", "wait-reply", "file", "photo", "video", "poll", "latitude", "contact", "dice"} {
		if SendCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("send flag --%s was not registered", flag)
		}
	}
}

func TestBuildSendParamsBranches(t *testing.T) {
	tests := []struct {
		name   string
		setup  func()
		args   []string
		method string
		key    string
		value  any
	}{
		{
			name: "text",
			setup: func() {
				sendFlags.Caption = "caption"
				replyToMessageID = 42
			},
			args:   []string{"hello"},
			method: methodSendReply,
			key:    "messageId",
			value:  int64(42),
		},
		{
			name: "dice",
			setup: func() {
				sendDice = true
				diceEmoticon = "dart"
				replyToMessageID = 7
			},
			method: "send_dice",
			key:    "emoticon",
			value:  "dart",
		},
		{
			name: "contact",
			setup: func() {
				sendContact = "+123"
				sendFirstName = "Ada"
				sendLastName = "Lovelace"
			},
			method: "send_contact",
			key:    "firstName",
			value:  "Ada",
		},
		{
			name: "poll",
			setup: func() {
				pollQuestion = "Choose"
				pollOptions = []string{"A", "B"}
			},
			method: "send_poll",
			key:    "question",
			value:  "Choose",
		},
		{
			name: "location",
			setup: func() {
				latitude = 41.1
				longitude = 44.8
			},
			method: "send_location",
			key:    "latitude",
			value:  41.1,
		},
		{name: "sticker", setup: func() { sendSticker = "s.webp" }, method: "send_sticker", key: "file", value: "s.webp"},
		{name: "voice", setup: func() { sendVoice = "v.ogg" }, method: "send_voice", key: "file", value: "v.ogg"},
		{name: "video note", setup: func() { sendVideoNote = "n.mp4" }, method: "send_video_note", key: "file", value: "n.mp4"},
		{name: "gif", setup: func() { sendGIF = "a.gif" }, method: "send_gif", key: "file", value: "a.gif"},
		{name: "photo", setup: func() { sendPhoto = "p.jpg" }, method: "send_photo", key: "file", value: "p.jpg"},
		{name: "video", setup: func() { sendVideo = "v.mp4" }, method: "send_video", key: "file", value: "v.mp4"},
		{name: "audio", setup: func() { sendAudio = "a.mp3" }, method: "send_audio", key: "file", value: "a.mp3"},
		{name: "document", setup: func() { sendDocument = "d.pdf" }, method: "send_document", key: "file", value: "d.pdf"},
		{name: "file", setup: func() { sendFile = "f.bin" }, method: "send_file", key: "file", value: "f.bin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSendGlobals(t)
			_ = sendFlags.To.Set("@peer")
			tt.setup()

			method, params := buildSendParams(tt.args)
			if method != tt.method {
				t.Fatalf("method = %q, want %q", method, tt.method)
			}
			if params["peer"] != "@peer" {
				t.Fatalf("peer = %v, want @peer", params["peer"])
			}
			if params[tt.key] != tt.value {
				t.Fatalf("%s = %v, want %v; params=%#v", tt.key, params[tt.key], tt.value, params)
			}
		})
	}
}

func TestBuildSendParamsDefaultMessage(t *testing.T) {
	resetSendGlobals(t)
	_ = sendFlags.To.Set("123")

	method, params := buildSendParams([]string{"hello"})
	if method != methodSendMessage {
		t.Fatalf("method = %q, want send_message", method)
	}
	if params["message"] != "hello" || params["peer"] != "123" {
		t.Fatalf("params = %#v", params)
	}
}

func resetSendGlobals(t *testing.T) {
	t.Helper()
	sendFlags = SendFlags{}
	sendFile = ""
	sendPhoto = ""
	sendVideo = ""
	sendAudio = ""
	sendDocument = ""
	sendVoice = ""
	sendVideoNote = ""
	sendSticker = ""
	sendGIF = ""
	replyToMessageID = 0
	pollQuestion = ""
	pollOptions = nil
	latitude = 0
	longitude = 0
	sendContact = ""
	sendFirstName = ""
	sendLastName = ""
	sendDice = false
	diceEmoticon = ""
}

func childCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
