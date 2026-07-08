package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"agent-telegram/telegram/types"
)

type testParams struct {
	types.NoValidation
	Name string `json:"name" validate:"required"`
}

func TestHandlerValidationAndCall(t *testing.T) {
	var called testParams
	handler := Handler(func(_ context.Context, p testParams) (map[string]any, error) {
		called = p
		return map[string]any{"name": p.Name}, nil
	}, "test")

	result, err := handler(context.Background(), json.RawMessage(`{"name":"ada"}`))
	if err != nil {
		t.Fatal(err)
	}
	if called.Name != "ada" || result.(map[string]any)["name"] != "ada" {
		t.Fatalf("result=%#v called=%+v", result, called)
	}
	if _, err := handler(context.Background(), json.RawMessage(`{"name":`)); err == nil {
		t.Fatal("invalid JSON should fail")
	}
	if _, err := handler(context.Background(), json.RawMessage(`{}`)); err == nil {
		t.Fatal("missing required field should fail")
	}

	want := errors.New("boom")
	handler = Handler(func(context.Context, testParams) (map[string]any, error) {
		return nil, want
	}, "explode")
	if _, err := handler(context.Background(), json.RawMessage(`{"name":"ada"}`)); !errors.Is(err, want) {
		t.Fatalf("call error = %v", err)
	}
}

func TestFilterUpdatesByPeer(t *testing.T) {
	updates := []types.StoredUpdate{
		{Data: map[string]any{"message": map[string]any{"peer": "user:42", "from_name": "Ada"}}},
		{Data: map[string]any{"message": map[string]any{"peer": "channel:7", "from_name": "News"}}},
		{Data: map[string]any{"other": true}},
	}
	if got := filterByPeer(updates, "42", ""); len(got) != 1 || got[0].Data["message"].(map[string]any)["peer"] != "user:42" {
		t.Fatalf("numeric filter = %+v", got)
	}
	if got := filterByPeer(updates, "", "@news"); len(got) != 1 {
		t.Fatalf("username filter = %+v", got)
	}
	if got := filterByPeer(updates, "", ""); len(got) != len(updates) {
		t.Fatalf("empty filter = %+v", got)
	}
	if peerMatches(map[string]any{"message": "bad"}, "42") {
		t.Fatal("bad message should not match")
	}
	if !isNumeric("42") || isNumeric("abc") {
		t.Fatal("isNumeric mismatch")
	}
}
