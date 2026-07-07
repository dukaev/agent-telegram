package cliutil

import (
	"fmt"
	"testing"
)

func TestApplyOutputBudgetMinimal(t *testing.T) {
	result := map[string]any{
		"messages": []any{
			map[string]any{"messageId": 1, "text": "hello world", "entities": []any{"bold"}},
			map[string]any{"messageId": 2, "text": "second", "entities": []any{"italic"}},
		},
		"count": 2,
	}

	out := ApplyOutputBudget(result, OutputBudgetOptions{
		Verbosity: VerbosityMinimal,
		MaxItems:  1,
	}).(map[string]any)

	messages := out["messages"].([]any)
	if len(messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(messages))
	}
	first := messages[0].(map[string]any)
	if first["messageId"] == nil {
		t.Fatal("minimal output should keep messageId")
	}
	if first["textPreview"] == nil {
		t.Fatal("minimal output should include text preview")
	}
	if first["entities"] != nil {
		t.Fatal("minimal output should omit noisy fields")
	}
}

func TestApplyOutputBudgetCompactTruncatesAndOmits(t *testing.T) {
	result := map[string]any{
		"messages": []any{
			map[string]any{"messageId": 1, "text": "abcdefghijklmnopqrstuvwxyz", "entities": []any{"bold"}},
		},
	}

	out := ApplyOutputBudget(result, OutputBudgetOptions{
		Verbosity:    VerbosityCompact,
		MaxTextChars: 5,
		Omit:         []string{"entities"},
	}).(map[string]any)

	first := out["messages"].([]any)[0].(map[string]any)
	text := first["text"].(string)
	if text == "abcdefghijklmnopqrstuvwxyz" {
		t.Fatal("compact output should truncate text")
	}
	if first["entities"] != nil {
		t.Fatal("omit should remove entities")
	}
}

func TestApplyOutputBudgetSummary(t *testing.T) {
	out := ApplyOutputBudget(map[string]any{"messages": []any{map[string]any{"id": 1}}}, OutputBudgetOptions{
		Summary: true,
	}).(map[string]any)
	if fmt.Sprint(out["messagesCount"]) != "1" {
		t.Fatalf("summary = %+v, want messagesCount", out)
	}
}

func TestApplyOutputBudgetSummaryCountsGenericArrays(t *testing.T) {
	out := ApplyOutputBudget(map[string]any{
		"operations": []any{map[string]any{"method": "send_message"}},
		"errorTypes": []any{
			map[string]any{"type": "VALIDATION"},
			map[string]any{"type": "SERVER_NOT_RUNNING"},
		},
	}, OutputBudgetOptions{Summary: true}).(map[string]any)

	if fmt.Sprint(out["operationsCount"]) != "1" {
		t.Fatalf("summary = %+v, want operationsCount", out)
	}
	if fmt.Sprint(out["errorTypesCount"]) != "2" {
		t.Fatalf("summary = %+v, want errorTypesCount", out)
	}
}

func TestApplyOutputBudgetTailsAuditAndLogArrays(t *testing.T) {
	out := ApplyOutputBudget(map[string]any{
		"events": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2},
			map[string]any{"id": 3},
		},
		"messages": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2},
			map[string]any{"id": 3},
		},
	}, OutputBudgetOptions{
		Verbosity: VerbosityCompact,
		MaxItems:  2,
	}).(map[string]any)

	events := out["events"].([]any)
	if fmt.Sprint(events[0].(map[string]any)["id"]) != "2" {
		t.Fatalf("events = %+v, want tail", events)
	}
	messages := out["messages"].([]any)
	if fmt.Sprint(messages[0].(map[string]any)["id"]) != "1" {
		t.Fatalf("messages = %+v, want head", messages)
	}
}
