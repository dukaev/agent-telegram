package bot

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddBotCommandScopesStepTextAlias(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	AddBotCommand(root)

	if StepCmd.Flags().Lookup("send") == nil || StepCmd.Flags().Lookup("text") == nil {
		t.Fatal("bot step should expose --send and --text")
	}
	if PressCmd.Flags().Lookup("text") != nil {
		t.Fatal("bot press should not expose --text")
	}
	if !strings.Contains(StepCmd.Example, "--send") || strings.Contains(StepCmd.Example, "--text") {
		t.Fatalf("bot step example = %q, want canonical --send only", StepCmd.Example)
	}
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

	actions := nextActions("@testbot", 123, state)
	if len(actions) != 3 {
		t.Fatalf("actions len = %d, want 3: %#v", len(actions), actions)
	}
	if actions[0]["kind"] != "press_inline_button" || actions[0]["safety"] != "write" {
		t.Fatalf("first action = %#v", actions[0])
	}
	if actions[0]["command"] == "" {
		t.Fatalf("first action missing command: %#v", actions[0])
	}
	if command, _ := actions[2]["command"].(string); !strings.Contains(command, "--send") || strings.Contains(command, "--text") {
		t.Fatalf("send action command = %q, want canonical --send only", command)
	}

	names := actionNames(actions)
	if len(names) != 3 || names[0] != "press_inline_button" || names[2] != "send_text" {
		t.Fatalf("names = %#v", names)
	}
}
