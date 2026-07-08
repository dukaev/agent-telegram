package message

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
	check := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}

	_, err := c.GetMessage(ctx, types.GetMessageParams{})
	check("GetMessage", err)
	_, err = c.GetMessages(ctx, types.GetMessagesParams{})
	check("GetMessages", err)
	_, err = c.InspectInlineButtons(ctx, types.InspectInlineButtonsParams{})
	check("InspectInlineButtons", err)
	_, err = c.PressInlineButton(ctx, types.PressInlineButtonParams{})
	check("PressInlineButton", err)
	_, err = c.InspectReplyKeyboard(ctx, types.PeerInfo{})
	check("InspectReplyKeyboard", err)
	_, err = c.UpdateMessage(ctx, types.UpdateMessageParams{})
	check("UpdateMessage", err)
	_, err = c.DeleteMessage(ctx, types.DeleteMessageParams{})
	check("DeleteMessage", err)
	_, err = c.ForwardMessage(ctx, types.ForwardMessageParams{})
	check("ForwardMessage", err)
	_, err = c.ReadMessages(ctx, types.ReadMessagesParams{})
	check("ReadMessages", err)
	_, err = c.SetTyping(ctx, types.SetTypingParams{})
	check("SetTyping", err)
	_, err = c.GetReplies(ctx, types.GetRepliesParams{})
	check("GetReplies", err)
	_, err = c.ReplyToComment(ctx, types.ReplyToCommentParams{})
	check("ReplyToComment", err)
	_, err = c.GetScheduledMessages(ctx, types.GetScheduledMessagesParams{})
	check("GetScheduledMessages", err)
	_, err = c.SendMessage(ctx, types.SendMessageParams{})
	check("SendMessage", err)
	_, err = c.SendReply(ctx, types.SendReplyParams{})
	check("SendReply", err)
}

func TestMessageConversionHelpers(t *testing.T) {
	if got := buildUserDisplayName(&tg.User{FirstName: "Ada", LastName: "Lovelace"}); got != "Ada Lovelace" {
		t.Fatalf("display name = %q", got)
	}
	if got := buildUserDisplayName(&tg.User{Username: "ada"}); got != "ada" {
		t.Fatalf("username display = %q", got)
	}
	if got := extractMessageID(&tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateMessageID{ID: 9}}}); got != 9 {
		t.Fatalf("extractMessageID updates = %d", got)
	}
	if got := extractMessageID(&tg.UpdateShortSentMessage{ID: 10}); got != 10 {
		t.Fatalf("extractMessageID short = %d", got)
	}
	if got := extractMessageID(&tg.Updates{}); got != 0 {
		t.Fatalf("extractMessageID empty = %d", got)
	}

	messages, users := extractMessagesData(&tg.MessagesMessages{
		Messages: []tg.MessageClass{&tg.Message{ID: 1}},
		Users:    []tg.UserClass{&tg.User{ID: 42}},
	})
	if len(messages) != 1 || len(users) != 1 {
		t.Fatalf("messages/users = %d/%d", len(messages), len(users))
	}
	messages, _ = extractMessagesData(&tg.MessagesMessagesSlice{Messages: []tg.MessageClass{&tg.Message{ID: 2}}})
	if len(messages) != 1 {
		t.Fatal("slice messages not extracted")
	}
	messages, _ = extractMessagesData(&tg.MessagesChannelMessages{Messages: []tg.MessageClass{&tg.Message{ID: 3}}})
	if len(messages) != 1 {
		t.Fatal("channel messages not extracted")
	}
	messages, users = extractMessagesData(nil)
	if messages != nil || users != nil {
		t.Fatalf("unknown messages = %#v/%#v", messages, users)
	}

	msg := &tg.Message{
		ID:         7,
		Date:       100,
		Message:    "hello",
		Out:        true,
		Pinned:     true,
		Mentioned:  true,
		Silent:     true,
		Post:       true,
		FromID:     &tg.PeerUser{UserID: 42},
		PeerID:     &tg.PeerChat{ChatID: 5},
		EditDate:   101,
		Views:      20,
		Forwards:   2,
		ViaBotID:   99,
		PostAuthor: "author",
		GroupedID:  88,
		TTLPeriod:  60,
		Media:      &tg.MessageMediaDice{Value: 6, Emoticon: "dice"},
		ReplyTo:    &tg.MessageReplyHeader{ReplyToMsgID: 1, ForumTopic: true},
		Entities: []tg.MessageEntityClass{
			&tg.MessageEntityBold{Offset: 0, Length: 5},
			&tg.MessageEntityTextURL{Offset: 0, Length: 5, URL: "https://example.com"},
			&tg.MessageEntityMentionName{Offset: 0, Length: 5, UserID: 42},
			&tg.MessageEntityPre{Offset: 0, Length: 5, Language: "go"},
			&tg.MessageEntityCustomEmoji{Offset: 0, Length: 1, DocumentID: 100},
		},
		ReplyMarkup: &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{Text: "URL", URL: "https://example.com"},
				&tg.KeyboardButtonCallback{Text: "Callback", Data: []byte("data")},
			},
		}}},
	}
	result := convertMessage(msg, map[int64]tg.UserClass{42: &tg.User{ID: 42, FirstName: "Ada"}})
	if result.ID != 7 || result.FromName != "Ada" || result.Media["type"] != "dice" || len(result.Buttons) != 2 {
		t.Fatalf("converted message = %+v", result)
	}
	if converted := convertMessagesToResult([]tg.MessageClass{msg, &tg.MessageService{}}, nil); len(converted) != 1 {
		t.Fatalf("converted messages len = %d", len(converted))
	}
}

func TestMediaButtonKeyboardAndEntityHelpers(t *testing.T) {
	for _, media := range []tg.MessageMediaClass{
		&tg.MessageMediaPhoto{Photo: &tg.Photo{ID: 1}},
		&tg.MessageMediaDocument{Document: &tg.Document{ID: 2}},
		&tg.MessageMediaWebPage{Webpage: &tg.WebPage{URL: "https://e", DisplayURL: "e"}},
		&tg.MessageMediaGeo{},
		&tg.MessageMediaContact{},
		&tg.MessageMediaPoll{},
		&tg.MessageMediaDice{Value: 3, Emoticon: "dice"},
		nil,
	} {
		if got := convertMedia(media); got["type"] == nil {
			t.Fatalf("media %T produced no type: %#v", media, got)
		}
	}
	reply := convertReplyHeader(&tg.MessageReplyHeader{ReplyToMsgID: 10, ForumTopic: true})
	if reply["reply_to_msg_id"] != 10 || reply["forum_topic"] != true {
		t.Fatalf("reply = %#v", reply)
	}
	fwd := convertFwdHeader(tg.MessageFwdHeader{FromID: &tg.PeerUser{UserID: 1}, Date: 2, FromName: "Ada"})
	if fwd["from_id"] != "user1" || fwd["date"] != 2 || fwd["from_name"] != "Ada" {
		t.Fatalf("fwd = %#v", fwd)
	}

	buttons := extractButtons(&tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{Text: "URL", URL: "https://e"},
			&tg.KeyboardButtonCallback{Text: "Callback", Data: []byte("cb")},
			&tg.KeyboardButtonSwitchInline{Text: "Switch", Query: "q"},
			&tg.KeyboardButtonGame{Text: "Game"},
			&tg.KeyboardButtonBuy{Text: "Buy"},
			&tg.KeyboardButtonURLAuth{Text: "Auth", URL: "https://auth"},
		},
	}}})
	if len(buttons) != 6 || buttons[1].Data != "cb" || buttons[5].Text != "Auth" {
		t.Fatalf("buttons = %#v", buttons)
	}
	if got := extractButtons(&tg.ReplyKeyboardMarkup{}); got != nil {
		t.Fatalf("non-inline buttons = %#v", got)
	}

	keyboard := convertReplyKeyboardMarkup(&tg.ReplyKeyboardMarkup{
		Resize: true,
		Rows: []tg.KeyboardButtonRow{{Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButton{Text: "Text"},
			&tg.KeyboardButtonURL{Text: "URL", URL: "https://e"},
			&tg.KeyboardButtonRequestPhone{Text: "Phone"},
			&tg.KeyboardButtonRequestGeoLocation{Text: "Geo"},
			&tg.KeyboardButtonRequestPoll{Text: "Poll", Quiz: true},
			&tg.KeyboardButtonWebView{Text: "Web", URL: "https://w"},
			&tg.KeyboardButtonRequestPeer{Text: "Peer", ButtonID: 1, MaxQuantity: 2},
			&tg.KeyboardButtonUserProfile{Text: "User", UserID: 42},
			nil,
		}}},
	})
	if len(keyboard.Rows) != 1 || len(keyboard.Rows[0]) != 9 || keyboard.Rows[0][4].PollType != "quiz" {
		t.Fatalf("keyboard = %#v", keyboard)
	}

	reactions := convertReactions(tg.MessageReactions{Results: []tg.ReactionCount{
		{Reaction: &tg.ReactionEmoji{Emoticon: "ok"}, Count: 2, ChosenOrder: 1},
		{Reaction: &tg.ReactionCustomEmoji{DocumentID: 3}, Count: 1},
	}})
	if len(reactions) != 2 || reactions[0]["emoticon"] != "ok" || reactions[1]["document_id"] != int64(3) {
		t.Fatalf("reactions = %#v", reactions)
	}
	entities := convertEntities([]tg.MessageEntityClass{
		&tg.MessageEntityURL{Offset: 0, Length: 1},
		&tg.MessageEntityEmail{Offset: 0, Length: 1},
		&tg.MessageEntityHashtag{Offset: 0, Length: 1},
		&tg.MessageEntityCashtag{Offset: 0, Length: 1},
		&tg.MessageEntityMention{Offset: 0, Length: 1},
		&tg.MessageEntityBotCommand{Offset: 0, Length: 1},
		&tg.MessageEntityItalic{Offset: 0, Length: 1},
		&tg.MessageEntityUnderline{Offset: 0, Length: 1},
		&tg.MessageEntityStrike{Offset: 0, Length: 1},
		&tg.MessageEntityCode{Offset: 0, Length: 1},
		&tg.MessageEntityBlockquote{Offset: 0, Length: 1},
	})
	if len(entities) != 11 || entities[0]["type"] != "url" {
		t.Fatalf("entities = %#v", entities)
	}
}

func TestReadAndSendMethodsWithFakeAPI(t *testing.T) {
	c := NewClient(fakeParent{peer: &tg.InputPeerSelf{}})
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.MessagesGetHistoryRequest:
			return &tg.MessagesMessages{
				Messages: []tg.MessageClass{&tg.Message{ID: 1, Date: 10, Message: "history", PeerID: &tg.PeerUser{UserID: 1}}},
				Users:    []tg.UserClass{},
			}, nil
		case *tg.MessagesGetMessagesRequest:
			return &tg.MessagesMessages{
				Messages: []tg.MessageClass{&tg.Message{
					ID:      2,
					Date:    20,
					Message: "one",
					PeerID:  &tg.PeerUser{UserID: 1},
					ReplyMarkup: &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{{
						Buttons: []tg.KeyboardButtonClass{&tg.KeyboardButtonCallback{Text: "Go", Data: []byte("cb")}},
					}}},
				}},
				Users: []tg.UserClass{},
			}, nil
		case *tg.MessagesSendMessageRequest:
			return &tg.UpdateShortSentMessage{ID: 33}, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))

	messages, err := c.GetMessages(context.Background(), types.GetMessagesParams{Username: "me", Limit: 0, Offset: -1})
	if err != nil {
		t.Fatal(err)
	}
	if messages.Count != 1 || messages.Messages[0].Text != "history" || messages.Limit != 10 {
		t.Fatalf("messages = %+v", messages)
	}
	one, err := c.GetMessage(context.Background(), types.GetMessageParams{
		PeerInfo: types.PeerInfo{Peer: "me"},
		MsgID:    types.MsgID{MessageID: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if one.Message.ID != 2 || len(one.Message.Buttons) != 1 {
		t.Fatalf("message = %+v", one)
	}
	buttons, err := c.InspectInlineButtons(context.Background(), types.InspectInlineButtonsParams{
		PeerInfo: types.PeerInfo{Peer: "me"},
		MsgID:    types.MsgID{MessageID: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(buttons.Buttons) != 1 || buttons.Buttons[0].Data != "cb" {
		t.Fatalf("buttons = %+v", buttons)
	}
	sent, err := c.SendMessage(context.Background(), types.SendMessageParams{
		PeerInfo: types.PeerInfo{Peer: "me"},
		Message:  "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sent.ID != 33 || sent.Message != "hello" {
		t.Fatalf("sent = %+v", sent)
	}
	reply, err := c.SendReply(context.Background(), types.SendReplyParams{
		PeerInfo: types.PeerInfo{Peer: "me"},
		MsgID:    types.MsgID{MessageID: 2},
		Text:     "reply",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply.ID != 33 || reply.ReplyTo != 2 {
		t.Fatalf("reply = %+v", reply)
	}
}
