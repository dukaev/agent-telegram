package chat

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddChatCommandRegistersExpectedSurface(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	AddChatCommand(root)

	chatCmd := childCommand(root, "chat")
	if chatCmd == nil {
		t.Fatal("chat command was not registered")
	}
	for _, name := range []string{
		"pin", "join", "subscribe", "topics", "mute", "archive",
		"create-group", "create-channel", "edit-title", "set-photo",
		"delete-photo", "leave", "invite", "participants", "admins",
		"banned", "promote-admin", "demote-admin", "invite-link",
		"list", "open", "info", "slow-mode", "permissions", "keyboard",
	} {
		if childCommand(chatCmd, name) == nil {
			t.Fatalf("chat subcommand %q was not registered", name)
		}
	}
	if ListCmd.Flags().Lookup("limit") == nil || OpenCmd.Flags().Lookup("offset") == nil {
		t.Fatal("expected list/open pagination flags")
	}
	if PermissionsCmd.Flags().Lookup("send-messages") == nil {
		t.Fatal("expected permissions flags")
	}
}

func TestFilterChatInfo(t *testing.T) {
	result := map[string]any{
		"chats": []any{
			map[string]any{"peer": "@first", "title": "First"},
			map[string]any{"peer": "second", "title": "Second"},
			"ignored",
		},
	}

	found := filterChatInfo(result, "@second").(map[string]any)
	chat := found["chat"].(map[string]any)
	if chat["title"] != "Second" {
		t.Fatalf("chat = %#v, want Second", chat)
	}

	missing := filterChatInfo(result, "@missing").(map[string]any)
	if missing["error"] == "" {
		t.Fatalf("missing result = %#v, want error", missing)
	}

	unchanged := filterChatInfo("not a result", "@second")
	if unchanged != "not a result" {
		t.Fatalf("unexpected non-map handling: %#v", unchanged)
	}
}

func TestTrimPeer(t *testing.T) {
	if got := trimPeer("@channel"); got != "channel" {
		t.Fatalf("trimPeer = %q, want channel", got)
	}
	if got := trimPeer("channel"); got != "channel" {
		t.Fatalf("trimPeer = %q, want channel", got)
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
