package search

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"

	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

type fakeParent struct {
	peer tg.InputPeerClass
}

func (f fakeParent) ResolvePeer(context.Context, string) (tg.InputPeerClass, error) {
	return f.peer, nil
}

func (f fakeParent) CachePeer(string, tg.InputPeerClass) {}

func TestClientMethodsRequireInitialization(t *testing.T) {
	c := NewClient(nil)
	ctx := context.Background()

	_, err := c.SearchGlobal(ctx, types.SearchGlobalParams{})
	if !errors.Is(err, client.ErrNotInitialized) {
		t.Fatalf("SearchGlobal err = %v", err)
	}
	_, err = c.SearchInChat(ctx, types.SearchInChatParams{})
	if !errors.Is(err, client.ErrNotInitialized) {
		t.Fatalf("SearchInChat err = %v", err)
	}
}

func TestSearchHelpers(t *testing.T) {
	userMap := map[int64]*tg.User{1: &tg.User{ID: 1, FirstName: "Ada", LastName: "Lovelace"}}
	if got := getFromName(&tg.PeerUser{UserID: 1}, userMap); got != "Ada Lovelace" {
		t.Fatalf("from name = %q", got)
	}
	if got := getFromName(&tg.PeerUser{UserID: 2}, userMap); got != "" {
		t.Fatalf("missing from name = %q", got)
	}
}

func TestSearchWithFakeAPI(t *testing.T) {
	c := NewClient(fakeParent{peer: &tg.InputPeerSelf{}})
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.ContactsSearchRequest:
			return &tg.ContactsFound{
				MyResults: []tg.PeerClass{&tg.PeerUser{UserID: 1}},
				Results:   []tg.PeerClass{&tg.PeerChannel{ChannelID: 2}},
				Users:     []tg.UserClass{&tg.User{ID: 1, FirstName: "Ada", Username: "ada", Bot: true}},
				Chats:     []tg.ChatClass{&tg.Channel{ID: 2, Title: "News", Photo: &tg.ChatPhotoEmpty{}}},
			}, nil
		case *tg.MessagesSearchRequest:
			return &tg.MessagesMessages{
				Messages: []tg.MessageClass{&tg.Message{
					ID:      5,
					Date:    50,
					Message: "found",
					PeerID:  &tg.PeerUser{UserID: 1},
					FromID:  &tg.PeerUser{UserID: 1},
					Media:   &tg.MessageMediaPhoto{},
				}},
				Users: []tg.UserClass{&tg.User{ID: 1, FirstName: "Ada"}},
			}, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))

	global, err := c.SearchGlobal(context.Background(), types.SearchGlobalParams{Query: "a", Limit: 0})
	if err != nil {
		t.Fatal(err)
	}
	if global.Count != 2 || global.Results[0].Peer != "bot:ada" || global.Results[1].FromName != "News" {
		t.Fatalf("global = %+v", global)
	}
	inChat, err := c.SearchInChat(context.Background(), types.SearchInChatParams{Peer: "me", Query: "found", Limit: 0, Offset: -1})
	if err != nil {
		t.Fatal(err)
	}
	if inChat.Count != 1 || inChat.Messages[0].Text != "found" || inChat.Messages[0].Media["type"] != "photo" {
		t.Fatalf("inChat = %+v", inChat)
	}

	for _, media := range []tg.MessageMediaClass{
		&tg.MessageMediaDocument{}, &tg.MessageMediaWebPage{}, &tg.MessageMediaGeo{},
		&tg.MessageMediaContact{}, &tg.MessageMediaPoll{},
	} {
		if got := extractMediaInfo(media); got["type"] == nil {
			t.Fatalf("media info = %#v", got)
		}
	}
}
