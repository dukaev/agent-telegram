package helpers

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestFormatPeer(t *testing.T) {
	tests := []struct {
		peer   tg.PeerClass
		format PeerFormat
		want   string
	}{
		{&tg.PeerUser{UserID: 1}, PeerFormatCompact, "user1"},
		{&tg.PeerUser{UserID: 1}, PeerFormatTyped, "user:1"},
		{&tg.PeerChat{ChatID: 2}, PeerFormatCompact, "-2"},
		{&tg.PeerChat{ChatID: 2}, PeerFormatTyped, "chat:2"},
		{&tg.PeerChannel{ChannelID: 3}, PeerFormatCompact, "-1003"},
		{&tg.PeerChannel{ChannelID: 3}, PeerFormatTyped, "channel:3"},
		{nil, PeerFormatTyped, ""},
	}
	for _, tt := range tests {
		if got := FormatPeer(tt.peer, tt.format); got != tt.want {
			t.Fatalf("FormatPeer(%T) = %q, want %q", tt.peer, got, tt.want)
		}
	}
}

func TestGetAccessHash(t *testing.T) {
	peer := &tg.ContactsResolvedPeer{
		Chats: []tg.ChatClass{
			&tg.Channel{ID: 10, AccessHash: 100},
			&tg.Chat{ID: 11},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 12, AccessHash: 120},
		},
	}
	if got := GetAccessHash(peer, 10); got != 100 {
		t.Fatalf("channel hash = %d", got)
	}
	if got := GetAccessHash(peer, 11); got != 0 {
		t.Fatalf("chat hash = %d", got)
	}
	if got := GetAccessHash(peer, 12); got != 120 {
		t.Fatalf("user hash = %d", got)
	}
	if got := GetAccessHash(peer, 99); got != 0 {
		t.Fatalf("missing hash = %d", got)
	}
}

func TestParseCustomEmojis(t *testing.T) {
	text, entities := ParseCustomEmojis("hi <custom:123>!")
	if text == "hi <custom:123>!" || len(entities) != 1 {
		t.Fatalf("parsed = %q, entities=%#v", text, entities)
	}
	ent := entities[0].(*tg.MessageEntityCustomEmoji)
	if ent.DocumentID != 123 || ent.Offset != 3 || ent.Length != 1 {
		t.Fatalf("entity = %+v", ent)
	}
	plain, entities := ParseCustomEmojis("plain")
	if plain != "plain" || entities != nil {
		t.Fatalf("plain = %q, %#v", plain, entities)
	}
	invalid, entities := ParseCustomEmojis("<custom:not-a-number>")
	if invalid != "<custom:not-a-number>" || len(entities) != 0 {
		t.Fatalf("invalid = %q, %#v", invalid, entities)
	}
}
