package telegram

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"

	"agent-telegram/telegram/types"
)

func TestUpdateStoreAddGetLimitOffsetAndCallback(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := NewUpdateStore(2)
		var seenMu sync.Mutex
		var seen []types.StoredUpdate
		store.SetOnUpdate(func(update types.StoredUpdate) {
			seenMu.Lock()
			defer seenMu.Unlock()
			seen = append(seen, update)
		})

		store.Add(NewStoredUpdate(types.UpdateTypeNewMessage, map[string]any{"text": "one"}))
		store.Add(NewStoredUpdate(types.UpdateTypeEditMessage, map[string]any{"text": "two"}))
		store.Add(NewStoredUpdate(types.UpdateTypeDelete, map[string]any{"text": "three"}))
		synctest.Wait()
		store.Wait()

		seenMu.Lock()
		defer seenMu.Unlock()
		if len(seen) != 3 {
			t.Fatalf("callbacks = %+v, want 3", seen)
		}
		if first := seen[0]; first.ID == 0 || first.Timestamp.IsZero() {
			t.Fatalf("callback update should have assigned metadata: %+v", first)
		}

		updates := store.Get(10)
		if len(updates) != 2 || updates[0].ID != 3 || updates[1].ID != 2 {
			t.Fatalf("updates = %+v, want newest two", updates)
		}
		if after := store.Get(10, 2); len(after) != 1 || after[0].ID != 3 {
			t.Fatalf("offset updates = %+v, want only id 3", after)
		}
		if empty := store.Get(10, 3); len(empty) != 0 {
			t.Fatalf("offset empty = %+v", empty)
		}
		if empty := NewUpdateStore(0).Get(10); len(empty) != 0 {
			t.Fatalf("empty store = %+v", empty)
		}
	})
}

func TestMessageDataAndHelpers(t *testing.T) {
	msg := &tg.Message{
		ID:      7,
		Message: "hello",
		Date:    123,
		Out:     true,
		FromID:  &tg.PeerUser{UserID: 42},
		PeerID:  &tg.PeerChannel{ChannelID: 99},
		Media:   &tg.MessageMediaDice{Value: 6, Emoticon: "dice"},
		ReplyMarkup: &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{Text: "Site", URL: "https://example.com"},
				&tg.KeyboardButtonCallback{Text: "Go", Data: []byte("cb")},
				&tg.KeyboardButtonSwitchInline{Text: "Switch", Query: "q"},
				&tg.KeyboardButtonGame{Text: "Game"},
				&tg.KeyboardButtonBuy{Text: "Buy"},
				&tg.KeyboardButtonURLAuth{Text: "Auth", URL: "https://auth.example"},
			},
		}}},
		Views:      10,
		Forwards:   2,
		GroupedID:  55,
		PostAuthor: "author",
		Post:       true,
		FwdFrom:    tg.MessageFwdHeader{FromName: "forwarded"},
	}
	entities := tg.Entities{Users: map[int64]*tg.User{
		42: {ID: 42, FirstName: "Bot", Username: "bot", Bot: true},
	}}

	data := MessageData(msg, entities)
	if data["id"] != 7 || data["text"] != "hello" || data["from_name"] != "Bot (bot)" {
		t.Fatalf("message data = %#v", data)
	}
	if media := data["media"].(map[string]any); media["type"] != "dice" || media["value"] != 6 {
		t.Fatalf("media = %#v", media)
	}
	buttons := data["buttons"].([]map[string]interface{})
	if len(buttons) != 6 || buttons[1]["data"] != "cb" || buttons[5]["type"] != "url_auth" {
		t.Fatalf("buttons = %#v", buttons)
	}

	if got := MessageData(&tg.MessageService{}, entities); len(got) != 0 {
		t.Fatalf("service message data = %#v", got)
	}
}

func TestConvertMediaForUpdate(t *testing.T) {
	tests := []struct {
		name string
		in   tg.MessageMediaClass
		typ  string
		key  string
		val  any
	}{
		{"photo", &tg.MessageMediaPhoto{Photo: &tg.Photo{ID: 1}}, "photo", "photo_id", int64(1)},
		{"document", &tg.MessageMediaDocument{Document: &tg.Document{ID: 2}}, "document", "document_id", int64(2)},
		{
			"webpage",
			&tg.MessageMediaWebPage{Webpage: &tg.WebPage{URL: "https://e", DisplayURL: "e"}},
			"webpage",
			"url",
			"https://e",
		},
		{"geo", &tg.MessageMediaGeo{Geo: &tg.GeoPoint{Lat: 1.2, Long: 3.4}}, "geo", "lat", 1.2},
		{"contact", &tg.MessageMediaContact{PhoneNumber: "+1", FirstName: "Ada", LastName: "L"}, "contact", "phone", "+1"},
		{"poll", &tg.MessageMediaPoll{}, "poll", "", nil},
		{"dice", &tg.MessageMediaDice{Value: 5, Emoticon: "dice"}, "dice", "value", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMediaForUpdate(tt.in)
			if got["type"] != tt.typ {
				t.Fatalf("type = %v, want %s", got["type"], tt.typ)
			}
			if tt.key != "" && got[tt.key] != tt.val {
				t.Fatalf("%s = %v, want %v", tt.key, got[tt.key], tt.val)
			}
		})
	}
	if got := convertMediaForUpdate(nil); got["type"] != "unknown" {
		t.Fatalf("nil media = %#v", got)
	}
	if got := extractButtonsData(&tg.ReplyKeyboardMarkup{}); got != nil {
		t.Fatalf("non-inline buttons = %#v", got)
	}
}

func TestEnvStorage(t *testing.T) {
	storage, err := NewEnvStorage(base64.StdEncoding.EncodeToString([]byte("session")))
	if err != nil {
		t.Fatal(err)
	}
	data, err := storage.LoadSession(context.Background())
	if err != nil || string(data) != "session" {
		t.Fatalf("LoadSession = %q, %v", data, err)
	}
	data[0] = 'X'
	again, _ := storage.LoadSession(context.Background())
	if string(again) != "session" {
		t.Fatalf("LoadSession should return a copy, got %q", again)
	}
	if err := storage.StoreSession(context.Background(), []byte("next")); err != nil {
		t.Fatal(err)
	}
	if data, _ := storage.LoadSession(context.Background()); string(data) != "next" {
		t.Fatalf("stored session = %q", data)
	}
	empty := &EnvStorage{}
	if _, err := empty.LoadSession(context.Background()); err != session.ErrNotFound {
		t.Fatalf("empty LoadSession err = %v", err)
	}
	if _, err := NewEnvStorage("not base64"); err == nil || !strings.Contains(err.Error(), "failed to decode") {
		t.Fatalf("invalid base64 err = %v", err)
	}
}

func TestClientAccessorsAndPeerCache(t *testing.T) {
	client := NewClient(1, "hash")
	store := NewUpdateStore(1)
	client.WithSessionPath("/tmp/session").WithSessionStorage(&EnvStorage{}).WithUpdateStore(store)
	if path, err := client.GetSessionPath(); err != nil || path != "/tmp/session" {
		t.Fatalf("session path = %q, %v", path, err)
	}
	if client.Client() != nil {
		t.Fatal("underlying Telegram client should be nil before Start")
	}
	if updates := client.GetUpdates(1); len(updates) != 0 {
		t.Fatalf("empty store updates = %+v", updates)
	}
	client.CachePeer("@cached", &tg.InputPeerSelf{})
	resolved, err := client.ResolvePeer(context.Background(), "@cached")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := resolved.(*tg.InputPeerSelf); !ok {
		t.Fatalf("resolved cached peer = %T", resolved)
	}
	self, err := client.ResolvePeer(context.Background(), "me")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := self.(*tg.InputPeerSelf); !ok {
		t.Fatalf("self peer = %T", self)
	}
	chat, err := client.ResolvePeer(context.Background(), "-123")
	if err != nil {
		t.Fatal(err)
	}
	if input, ok := chat.(*tg.InputPeerChat); !ok || input.ChatID != 123 {
		t.Fatalf("chat peer = %#v", chat)
	}
	if id, err := client.ResolvePeerID(context.Background(), "-123"); err != nil || id != "chat:123" {
		t.Fatalf("ResolvePeerID = %q, %v", id, err)
	}
	if _, err := client.ResolvePeer(context.Background(), "not numeric"); err == nil {
		t.Fatal("invalid numeric peer should fail")
	}
	if _, err := client.GetMe(context.Background()); err == nil {
		t.Fatal("GetMe without client should fail")
	}
	empty := NewClient(1, "hash")
	if updates := empty.GetUpdates(1); len(updates) != 0 {
		t.Fatalf("nil store updates = %+v", updates)
	}
	client.Reload()
	select {
	case <-client.ReloadCh():
	case <-time.After(time.Second):
		t.Fatal("reload signal was not sent")
	}
}
