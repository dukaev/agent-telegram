package cliutil

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFilterItemsAndMapResponse(t *testing.T) {
	items := []map[string]any{
		{"type": "user", "bot": true, "username": "helper_bot", "peer": "@helper"},
		{"type": "channel", "title": "News", "peer": "@news"},
		{"type": "user", "first_name": "Ada", "last_name": "Lovelace"},
	}

	bots := FilterItems(items, FilterOptions{Type: "bot"})
	if bots.Count != 1 || bots.Items[0]["username"] != "helper_bot" || bots.Total != 3 {
		t.Fatalf("bot filter = %+v", bots)
	}
	search := FilterItems(items, FilterOptions{Search: "lovelace"})
	if search.Count != 1 || search.Items[0]["first_name"] != "Ada" {
		t.Fatalf("search filter = %+v", search)
	}
	none := FilterItems(items, FilterOptions{Type: "bot", Search: "news"})
	if none.Count != 0 {
		t.Fatalf("combined filter = %+v, want empty", none)
	}

	mapped := MapResponse(map[string]any{"limit": 10, "offset": 2, "count": 3}, "items", bots.Items)
	if mapped["count"] != 1 || mapped["total"] != 3 || mapped["limit"] != 10 || mapped["offset"] != 2 {
		t.Fatalf("mapped response = %#v", mapped)
	}
	if !ContainsAny("ada", "", "Ada Lovelace") || ContainsAny("missing", "Ada") {
		t.Fatal("ContainsAny returned unexpected result")
	}
}

func TestFilterExpressions(t *testing.T) {
	exprs, err := ParseFilterExpressions([]string{"stars>=100", "type!=channel", "title=Heart"})
	if err != nil {
		t.Fatal(err)
	}
	item := map[string]any{"stars": int64(150), "type": "gift", "title": "heart"}
	if !exprs.matchesAll(item) {
		t.Fatalf("item should match filters")
	}
	item["stars"] = 50
	if exprs.matchesAll(item) {
		t.Fatalf("item should not match filters")
	}
	if _, err := ParseFilterExpressions([]string{"broken"}); err == nil {
		t.Fatal("invalid filter should fail")
	}
	if exprs, err := ParseFilterExpressions(nil); err != nil || exprs != nil {
		t.Fatalf("empty filters = %#v, %v", exprs, err)
	}

	wrapped := map[string]any{
		"gifts": []any{
			map[string]any{"stars": float64(200), "title": "A"},
			map[string]any{"stars": float64(50), "title": "B"},
			"ignored",
		},
		"count": 2,
	}
	onlyExpensive := FilterExpressions{&FilterExpression{Key: "stars", Operator: ">", Value: "100"}}.
		Apply(wrapped).(map[string]any)
	if onlyExpensive["count"] != 1 || len(onlyExpensive["gifts"].([]map[string]any)) != 1 {
		t.Fatalf("filtered wrapper = %#v", onlyExpensive)
	}
	if got := (FilterExpressions{&FilterExpression{Key: "title", Operator: "=", Value: "a"}}).Apply(wrapped); got == nil {
		t.Fatal("string filter should match case-insensitively")
	}
	if got := (FilterExpressions{&FilterExpression{Key: "missing", Operator: "=", Value: "x"}}).Apply(map[string]any{}); got != nil {
		t.Fatalf("missing map filter = %#v, want nil", got)
	}
	if got := (FilterExpressions{}).Apply("unchanged"); got != "unchanged" {
		t.Fatalf("empty expressions changed result: %#v", got)
	}

	for _, op := range []string{"=", "!=", ">", "<", ">=", "<=", "?"} {
		_ = compareNumbers(1, op, 1)
		_ = compareStrings("a", op, "b")
	}
	for _, value := range []any{float64(1), int64(1), int(1), "x"} {
		_, _ = toFloat64(value)
	}
}

func TestFieldSelector(t *testing.T) {
	if NewFieldSelector(nil) != nil {
		t.Fatal("empty selector should be nil")
	}
	selector := NewFieldSelector([]string{"id", "from.id", "missing.value"})
	input := map[string]any{
		"messages": []any{
			map[string]any{"id": 1, "text": "hello", "from": map[string]any{"id": 42, "name": "Ada"}},
			"keep",
		},
		"count": 1,
	}
	out := selector.Apply(input).(map[string]any)
	items := out["messages"].([]any)
	first := items[0].(map[string]any)
	if first["id"] != 1 || first["from.id"] != 42 || first["text"] != nil || items[1] != "keep" {
		t.Fatalf("field-selected wrapper = %#v", out)
	}
	single := selector.Apply(map[string]any{"id": 7, "from": map[string]any{"id": 8}}).(map[string]any)
	if single["id"] != 7 || single["from.id"] != 8 {
		t.Fatalf("single field selection = %#v", single)
	}
	if got := selector.Apply("unchanged"); got != "unchanged" {
		t.Fatalf("non-map selection = %#v", got)
	}
}

func TestOutputFormattingAndIDs(t *testing.T) {
	if ParseOutputFormat("ids") != OutputIDs || ParseOutputFormat("json") != OutputJSON || ParseOutputFormat("") != OutputJSON {
		t.Fatal("unexpected output format parsing")
	}
	warning := captureCLIStderr(t, func() { _ = ParseOutputFormat("weird") })
	if !strings.Contains(warning, "unknown output format") {
		t.Fatalf("warning = %q", warning)
	}

	output := captureStdout(t, func() {
		printIDs(map[string]any{
			"items": []any{
				map[string]any{"slug": "gift-1"},
				map[string]any{"id": float64(2)},
				map[string]any{"slug": float64(3.5)},
				"ignored",
			},
		}, "slug")
		printIDs(map[string]any{"id": json.Number("42")}, "id")
		printIDs("single", "")
	})
	for _, want := range []string{"gift-1", "2", "3.5", "42", "single"} {
		if !strings.Contains(output, want) {
			t.Fatalf("ids output missing %q:\n%s", want, output)
		}
	}
}

func TestPrintHelpers(t *testing.T) {
	output := captureCLIStderr(t, func() {
		PrintSuccessSummary(map[string]any{"success": true}, "done")
		PrintSuccessSummary("bad", "ignored")
		PrintSuccessWithDuration(map[string]any{"success": true}, "fast", 250*time.Millisecond)
		PrintSuccessWithDuration(map[string]any{"success": true}, "slow", 1500*time.Millisecond)
		PrintResultField(map[string]any{"name": "Ada"}, "name", "Name: %s\n")
		PrintResultField(map[string]any{"id": float64(7)}, "id", "ID: %d\n")
		PrintResultField(map[string]any{"id": int64(8)}, "id", "ID: %d\n")
		PrintInviteLinkSummary(map[string]any{"link": "https://t.me/+x", "usage": float64(2), "usageLimit": int64(5)})
		PrintResultCount(map[string]any{"count": float64(3)}, "count", "Count")
		PrintParticipants(map[string]any{"count": float64(1), "participants": []any{map[string]any{"firstName": "Ada", "peer": "@ada"}}}, "Unknown", "N/A")
		PrintBanned(map[string]any{"count": float64(1), "banned": []any{map[string]any{"lastName": "NoName"}}}, "Unknown", "N/A")
		PrintAdmins(map[string]any{"count": float64(1), "admins": []any{map[string]any{"firstName": "Root", "creator": true}}}, "Unknown", "N/A")
		PrintTopics(map[string]any{"count": float64(1), "topics": []any{map[string]any{"id": float64(9), "title": "General"}}}, "N/A")
		FormatSuccess(map[string]any{"id": float64(10), "peer": "@p"}, "send_message")
		FormatSuccess("ok", "custom")
	})
	for _, want := range []string{
		"done", "fast (250ms)", "slow (1.5s)", "Name: Ada", "Invite link",
		"Count: 3", "Ada (@ada)", "NoName", "Root", "(Creator)", "[9] General",
		"send_message sent successfully", "custom succeeded",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("print output missing %q:\n%s", want, output)
		}
	}
}

func TestPaginationSlugChatsAndExtractors(t *testing.T) {
	pag := NewPagination(-1, -5, PaginationConfig{MaxLimit: 50})
	if pag.Limit != 1 || pag.Offset != 0 {
		t.Fatalf("normalized pagination = %+v", pag)
	}
	pag = NewPagination(1000, 5, PaginationConfig{MaxLimit: 50})
	params := map[string]any{}
	pag.ToParams(params, true)
	if params["limit"] != 50 || params["offset"] != 5 {
		t.Fatalf("pagination params = %#v", params)
	}
	if got := ParseGiftSlug(" https://t.me/nft/Watch-7 "); got != "Watch-7" {
		t.Fatalf("slug = %q", got)
	}

	chats := filterChatsResult(map[string]any{
		"limit":  float64(10),
		"offset": float64(0),
		"chats": []any{
			map[string]any{"type": "channel", "channel_id": int64(1), "title": "News", "peer": "@news"},
			map[string]any{"type": "user", "user_id": float64(2), "username": "ada", "peer": "@ada", "bot": true},
			"ignored",
		},
	}, "ada", "bot").(map[string]any)
	if chats["count"] != 1 {
		t.Fatalf("filtered chats = %#v", chats)
	}
	if got := filterChatsResult("bad", "", ""); got != "bad" {
		t.Fatalf("non-map chats result = %#v", got)
	}

	if ExtractString(map[string]any{"s": "x"}, "s") != "x" {
		t.Fatal("ExtractString failed")
	}
	if ExtractFloat64(map[string]any{"n": json.Number("1.5")}, "n") != 1.5 {
		t.Fatal("ExtractFloat64 failed")
	}
	if ExtractInt64(map[string]any{"n": json.Number("42")}, "n") != 42 {
		t.Fatal("ExtractInt64 failed")
	}
	if m, ok := ToMap(map[string]any{"ok": true}); !ok || m["ok"] != true {
		t.Fatal("ToMap failed")
	}
}

func captureCLIStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return <-done
}
