package bot

import "testing"

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

	names := actionNames(actions)
	if len(names) != 3 || names[0] != "press_inline_button" || names[2] != "send_text" {
		t.Fatalf("names = %#v", names)
	}
}
