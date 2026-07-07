package callback

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackServerAcceptsSnakeCaseSubscriptionPayloads(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store.data.CallbackURL = "https://example.test/hook"
	store.data.Verified = true

	manager := NewManager(store)
	manager.WithPeerResolver(func(_ context.Context, peer string) (string, error) {
		if peer != "@channel" {
			t.Fatalf("peer = %q, want @channel", peer)
		}
		return "channel:123", nil
	})
	srv := NewServer(manager, 0, "secret")

	req := httptest.NewRequest(
		http.MethodPost,
		"/callback/subscribe-channel",
		strings.NewReader(`{"channel_id":"@channel","event_types":"new_post,edit_post"}`),
	)
	req.Header.Set("X-Secret", "secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status body = %q, want ok; body=%s", body.Status, rec.Body.String())
	}

	state := store.Get()
	if len(state.Subscriptions) != 1 {
		t.Fatalf("subscriptions = %d, want 1", len(state.Subscriptions))
	}
	if state.Subscriptions[0].ChannelID != "channel:123" {
		t.Fatalf("channel id = %q, want channel:123", state.Subscriptions[0].ChannelID)
	}
}

func TestCallbackServerAcceptsSnakeCaseUnsubscribePayload(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, err := store.AddSubscription(Subscription{
		Type:       "channel",
		ChannelID:  "channel:123",
		EventTypes: []string{"new_post"},
	})
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManager(store)
	srv := NewServer(manager, 0, "secret")

	req := httptest.NewRequest(
		http.MethodPost,
		"/callback/unsubscribe",
		strings.NewReader(`{"subscription_id":`+jsonNumber(id)+`}`),
	)
	req.Header.Set("X-Secret", "secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := len(store.Get().Subscriptions); got != 0 {
		t.Fatalf("subscriptions = %d, want 0", got)
	}
}

func TestCallbackServerRequiresSecret(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServer(NewManager(store), 0, "secret")

	req := httptest.NewRequest(http.MethodGet, "/callback/get-callback-info", nil)
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	var body struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "error" || body.Error != "unauthorized" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func jsonNumber(n int64) string {
	data, _ := json.Marshal(n)
	return string(data)
}
