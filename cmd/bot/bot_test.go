package bot

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestAddBotCommandScopesStepTextAlias(t *testing.T) {
	resetBotCommandTreeForTest(BotCmd)
	root := &cobra.Command{Use: "agent-telegram"}
	AddBotCommand(root)

	if StepCmd.Flags().Lookup("send") == nil || StepCmd.Flags().Lookup("text") == nil {
		t.Fatal("bot step should expose --send and --text")
	}
	if PressCmd.Flags().Lookup("text") == nil {
		t.Fatal("bot press should expose --text")
	}
	if StepCmd.Flags().Lookup("thread-id") == nil {
		t.Fatal("bot step should expose --thread-id")
	}
	if !strings.Contains(StepCmd.Example, "--send") || !strings.Contains(StepCmd.Example, "-5424738551") || strings.Contains(StepCmd.Example, "--text") {
		t.Fatalf("bot step example = %q, want canonical --send and negative peer guidance", StepCmd.Example)
	}
}

func resetBotCommandTreeForTest(cmd *cobra.Command) {
	if parent := cmd.Parent(); parent != nil {
		parent.RemoveCommand(cmd)
	}
	for _, child := range cmd.Commands() {
		resetBotCommandTreeForTest(child)
		cmd.RemoveCommand(child)
	}
	cmd.ResetFlags()
}

func TestResolveStepText(t *testing.T) {
	tests := []struct {
		name, send, text, want, wantErr string
		sendChanged, textChanged        bool
	}{
		{name: "send", send: "hello", want: "hello", sendChanged: true},
		{name: "text", text: "hello", want: "hello", textChanged: true},
		{name: "same", send: "hello", text: "hello", want: "hello", sendChanged: true, textChanged: true},
		{name: "conflict", send: "a", text: "b", sendChanged: true, textChanged: true, wantErr: "use only --send or --text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "step"}
			cmd.Flags().String("send", "", "")
			cmd.Flags().String("text", "", "")
			if tt.sendChanged {
				_ = cmd.Flags().Set("send", tt.send)
			}
			if tt.textChanged {
				_ = cmd.Flags().Set("text", tt.text)
			}
			got, err := resolveStepText(cmd, tt.send, tt.text)
			if got != tt.want || (tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr))) {
				t.Fatalf("got %q, error %v", got, err)
			}
		})
	}
}

func TestNextActionsAreStructured(t *testing.T) {
	state := map[string]any{
		"inlineButtons": []any{map[string]any{"text": "Start"}},
		"replyKeyboard": map[string]any{
			"rows": []any{[]any{"Menu"}},
		},
	}

	actions := nextActions("@testbot", 77, 123, state)
	if len(actions) != 3 {
		t.Fatalf("actions len = %d, want 3: %#v", len(actions), actions)
	}
	if actions[0]["kind"] != "press_inline_button" || actions[0]["safety"] != "write" {
		t.Fatalf("first action = %#v", actions[0])
	}
	if actions[0]["command"] == "" {
		t.Fatalf("first action missing command: %#v", actions[0])
	}
	if command, _ := actions[0]["command"].(string); !strings.Contains(command, "--text Start") {
		t.Fatalf("press action command = %q", command)
	}
	if command, _ := actions[2]["command"].(string); !strings.Contains(command, "--send") || strings.Contains(command, "--text") {
		t.Fatalf("send action command = %q, want canonical --send only", command)
	}
	if command, _ := actions[2]["command"].(string); !strings.Contains(command, "--thread-id 77") {
		t.Fatalf("send action lost thread: %q", command)
	}

	names := actionNames(actions)
	if len(names) != 3 || names[0] != "press_inline_button" || names[2] != "send_text" {
		t.Fatalf("names = %#v", names)
	}
}

func TestResolvePressArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		peerProvided bool
		text         string
		wantPeer     string
		wantMessage  string
		wantIndex    string
		wantErr      bool
	}{
		{name: "positional index", args: []string{"@bot", "123", "2"}, wantPeer: "@bot", wantMessage: "123", wantIndex: "2"},
		{name: "positional text", args: []string{"@bot", "123"}, text: "Edit", wantPeer: "@bot", wantMessage: "123"},
		{name: "to text", args: []string{"123"}, peerProvided: true, text: "Edit", wantMessage: "123"},
		{name: "to index", args: []string{"123", "2"}, peerProvided: true, wantMessage: "123", wantIndex: "2"},
		{name: "missing selector", args: []string{"@bot", "123"}, wantErr: true},
		{name: "conflicting selector", args: []string{"@bot", "123", "2"}, text: "Edit", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer, message, index, err := resolvePressArgs(tt.args, tt.peerProvided, tt.text)
			if (err != nil) != tt.wantErr || peer != tt.wantPeer || message != tt.wantMessage || index != tt.wantIndex {
				t.Fatalf("got peer=%q message=%q index=%q err=%v", peer, message, index, err)
			}
		})
	}
}

func TestSnapshotChanged(t *testing.T) {
	base := map[string]any{
		"id": float64(123), "threadId": float64(77), "editDate": float64(10), "text": "before",
		"buttons": []any{map[string]any{"text": "Edit", "index": float64(0)}},
	}
	snapshot := newMessageSnapshot(base)
	if snapshotChanged(snapshot, base) {
		t.Fatal("unchanged message reported as edited")
	}
	for _, change := range []map[string]any{
		{"id": float64(123), "threadId": float64(77), "editDate": float64(11), "text": "before", "buttons": base["buttons"]},
		{"id": float64(123), "threadId": float64(77), "editDate": float64(10), "text": "after", "buttons": base["buttons"]},
		{"id": float64(123), "threadId": float64(77), "editDate": float64(10), "text": "before", "buttons": []any{map[string]any{"text": "Demo"}}},
	} {
		if !snapshotChanged(snapshot, change) {
			t.Fatalf("change not detected: %#v", change)
		}
	}
}

type editedMessagePoller struct {
	message map[string]any
}

type newMessagePoller struct{}

func (newMessagePoller) CallInternal(method string, params any) any {
	if method != "get_messages" {
		return map[string]any{}
	}
	request := params.(map[string]any)
	if request["threadId"] != int64(77) {
		panic("threadId missing from bot poll")
	}
	return map[string]any{"messages": []any{
		map[string]any{"id": float64(124), "out": false, "threadId": float64(77), "text": "new"},
	}}
}

func (p editedMessagePoller) CallInternal(method string, _ any) any {
	if method == "get_messages" {
		return map[string]any{"messages": []any{}}
	}
	return map[string]any{"message": p.message}
}

func TestWaitForBotEventCompletesOnEdit(t *testing.T) {
	oldNow, oldSleep := botWaitNow, botWaitSleep
	t.Cleanup(func() { botWaitNow, botWaitSleep = oldNow, oldSleep })
	now := time.Unix(1, 0)
	botWaitNow = func() time.Time { return now }
	botWaitSleep = func(d time.Duration) { now = now.Add(d) }

	before := newMessageSnapshot(map[string]any{"id": float64(123), "threadId": float64(77), "text": "before"})
	edited := map[string]any{"id": float64(123), "threadId": float64(77), "editDate": float64(20), "text": "after"}
	outcome := waitForBotEvent(editedMessagePoller{message: edited}, "@bot", 77, 123, before, time.Second)
	if !outcome.Completed || outcome.Event != "message_edited" || outcome.Message["text"] != "after" {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestWaitForBotEventCompletesOnNewMessage(t *testing.T) {
	oldNow, oldSleep := botWaitNow, botWaitSleep
	t.Cleanup(func() { botWaitNow, botWaitSleep = oldNow, oldSleep })
	now := time.Unix(1, 0)
	botWaitNow = func() time.Time { return now }
	botWaitSleep = func(d time.Duration) { now = now.Add(d) }

	outcome := waitForBotEvent(newMessagePoller{}, "@bot", 77, 123, messageSnapshot{ID: 123}, time.Second)
	if !outcome.Completed || outcome.Event != "new_message" || outcome.Message["text"] != "new" {
		t.Fatalf("outcome = %+v", outcome)
	}
}
