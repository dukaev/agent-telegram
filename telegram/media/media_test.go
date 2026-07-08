package media

import (
	"context"
	"errors"
	"path/filepath"
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
	check := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}

	_, err := c.SendContact(ctx, types.SendContactParams{})
	check("SendContact", err)
	_, err = c.SendDice(ctx, types.SendDiceParams{})
	check("SendDice", err)
	_, err = c.SendFile(ctx, types.SendFileParams{})
	check("SendFile", err)
	_, err = c.SendDocument(ctx, "@p", "file.txt", "text/plain", "")
	check("SendDocument", err)
	_, err = c.SendVideo(ctx, types.SendVideoParams{})
	check("SendVideo", err)
	_, err = c.SendLocation(ctx, types.SendLocationParams{})
	check("SendLocation", err)
	_, err = c.SendPhoto(ctx, types.SendPhotoParams{})
	check("SendPhoto", err)
	_, err = c.SendPoll(ctx, types.SendPollParams{})
	check("SendPoll", err)
	_, err = c.SendSticker(ctx, types.SendStickerParams{})
	check("SendSticker", err)
	_, err = c.GetStickerPacks(ctx, types.GetStickerPacksParams{})
	check("GetStickerPacks", err)
	_, err = c.SendGIF(ctx, types.SendGIFParams{})
	check("SendGIF", err)
	_, err = c.SendVoice(ctx, types.SendVoiceParams{})
	check("SendVoice", err)
	_, err = c.SendVideoNote(ctx, types.SendVideoNoteParams{})
	check("SendVideoNote", err)
}

func TestMediaExtractors(t *testing.T) {
	if got := extractMessageID(&tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateMessageID{ID: 9}}}); got != 9 {
		t.Fatalf("extractMessageID updates = %d", got)
	}
	if got := extractMessageID(&tg.UpdateShortSentMessage{ID: 10}); got != 10 {
		t.Fatalf("extractMessageID short = %d", got)
	}
	if got := extractMessageID(&tg.Updates{}); got != 0 {
		t.Fatalf("extractMessageID empty = %d", got)
	}

	msgs := &tg.MessagesMessages{Messages: []tg.MessageClass{
		&tg.Message{Media: &tg.MessageMediaDice{Value: 6}},
	}}
	if got := extractDiceFromMessages(msgs); got != 6 {
		t.Fatalf("dice from messages = %d", got)
	}
	channelMsgs := &tg.MessagesChannelMessages{Messages: []tg.MessageClass{
		&tg.Message{Media: &tg.MessageMediaDice{Value: 5}},
	}}
	if got := extractDiceFromMessages(channelMsgs); got != 5 {
		t.Fatalf("dice from channel messages = %d", got)
	}
	if got := extractDiceFromMessages(&tg.MessagesMessages{Messages: []tg.MessageClass{&tg.MessageService{}}}); got != 0 {
		t.Fatalf("dice from service = %d", got)
	}

	short := &tg.UpdateShortSentMessage{ID: 1}
	short.SetMedia(&tg.MessageMediaDice{Value: 4})
	if got := extractDiceValue(short); got != 4 {
		t.Fatalf("short dice = %d", got)
	}
	updates := &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.Message{Media: &tg.MessageMediaDice{Value: 3}}},
	}}
	if got := extractDiceValue(updates); got != 3 {
		t.Fatalf("updates dice = %d", got)
	}
	channelUpdates := &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateNewChannelMessage{Message: &tg.Message{Media: &tg.MessageMediaDice{Value: 2}}},
	}}
	if got := extractDiceValue(channelUpdates); got != 2 {
		t.Fatalf("channel updates dice = %d", got)
	}
	if got := extractDiceValue(&tg.Updates{}); got != 0 {
		t.Fatalf("empty dice = %d", got)
	}
}

func TestUploadFileOpenError(t *testing.T) {
	_, err := uploadFile(context.Background(), nil, filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("missing file should fail")
	}
}

func TestSimpleSendMediaMethodsWithFakeAPI(t *testing.T) {
	c := NewClient(fakeParent{peer: &tg.InputPeerSelf{}})
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.MessagesSendMediaRequest:
			out := &tg.UpdateShortSentMessage{ID: 44}
			out.SetMedia(&tg.MessageMediaDice{Value: 6, Emoticon: "dice"})
			return out, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))
	ctx := context.Background()

	contact, err := c.SendContact(ctx, types.SendContactParams{
		PeerInfo:  types.PeerInfo{Peer: "me"},
		Phone:     "+1",
		FirstName: "Ada",
	})
	if err != nil {
		t.Fatal(err)
	}
	if contact.ID != 44 || contact.Phone != "+1" {
		t.Fatalf("contact = %+v", contact)
	}
	location, err := c.SendLocation(ctx, types.SendLocationParams{
		PeerInfo:  types.PeerInfo{Peer: "me"},
		Latitude:  1.2,
		Longitude: 3.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if location.ID != 44 || location.Latitude != 1.2 || location.Longitude != 3.4 {
		t.Fatalf("location = %+v", location)
	}
	poll, err := c.SendPoll(ctx, types.SendPollParams{
		PeerInfo: types.PeerInfo{Peer: "me"},
		Question: "Q",
		Options:  []types.PollOption{{Text: "A"}, {Text: "B"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if poll.ID != 44 || poll.Question != "Q" {
		t.Fatalf("poll = %+v", poll)
	}
	dice, err := c.SendDice(ctx, types.SendDiceParams{PeerInfo: types.PeerInfo{Peer: "me"}})
	if err != nil {
		t.Fatal(err)
	}
	if dice.ID != 44 || dice.Value != 6 {
		t.Fatalf("dice = %+v", dice)
	}
}
