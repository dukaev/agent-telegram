package message

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddMsgCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddMsgCommand(root)

	msgCmd := childCommand(root, "msg")
	if msgCmd == nil {
		t.Fatal("msg command was not registered")
	}
	for _, name := range []string{
		"get", "list", "delete", "forward", "pin", "inspect-buttons",
		"press-button", "reaction", "inspect-keyboard", "press-keyboard",
		"wait", "read", "typing", "scheduled", "clear", "replies", "reply-comment",
	} {
		if childCommand(msgCmd, name) == nil {
			t.Fatalf("msg subcommand %q was not registered", name)
		}
	}
	if childCommand(root, "send") == nil {
		t.Fatal("send command should be registered as a top-level command")
	}
	if PressButtonCmd.Flags().Lookup("wait-reply") == nil || PressKeyboardCmd.Flags().Lookup("timeout") == nil {
		t.Fatal("expected button flags")
	}
}

func TestFindButtonText(t *testing.T) {
	keyboard := map[string]any{
		"found": true,
		"keyboard": map[string]any{
			"rows": []any{
				[]any{
					map[string]any{"text": "Start"},
					map[string]any{"text": "Settings"},
				},
				[]any{
					map[string]any{"text": "Help Center"},
					"ignored",
				},
			},
		},
	}

	if got := findButtonText(keyboard, "1"); got != "Settings" {
		t.Fatalf("index match = %q, want Settings", got)
	}
	if got := findButtonText(keyboard, "Start"); got != "Start" {
		t.Fatalf("exact match = %q, want Start", got)
	}
	if got := findButtonText(keyboard, "help"); got != "Help Center" {
		t.Fatalf("substring match = %q, want Help Center", got)
	}
}

func childCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
