package types

import "testing"

func TestMessageParamValidation(t *testing.T) {
	validPeerMsg := []interface{ Validate() error }{
		GetMessageParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		SendReplyParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}, Text: "hi"},
		UpdateMessageParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}, Text: "hi"},
		DeleteMessageParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		PinMessageParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		UnpinMessageParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		InspectInlineButtonsParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		GetRepliesParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		ReplyToCommentParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}, CommentID: 2, Text: "hi"},
	}
	for _, params := range validPeerMsg {
		if err := params.Validate(); err != nil {
			t.Fatalf("%T valid params failed: %v", params, err)
		}
	}

	invalid := []interface{ Validate() error }{
		GetMessageParams{MsgID: MsgID{MessageID: 1}},
		SendReplyParams{PeerInfo: PeerInfo{Peer: "@p"}},
		UpdateMessageParams{PeerInfo: PeerInfo{Peer: "@p"}},
		DeleteMessageParams{PeerInfo: PeerInfo{Peer: "@p"}},
		PinMessageParams{PeerInfo: PeerInfo{Peer: "@p"}},
		UnpinMessageParams{PeerInfo: PeerInfo{Peer: "@p"}},
		InspectInlineButtonsParams{PeerInfo: PeerInfo{Peer: "@p"}},
		GetRepliesParams{PeerInfo: PeerInfo{Peer: "@p"}},
		ReplyToCommentParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
	}
	for _, params := range invalid {
		if err := params.Validate(); err == nil {
			t.Fatalf("%T invalid params passed", params)
		}
	}

	if err := (PressInlineButtonParams{ButtonIndex: -1}).Validate(); err == nil {
		t.Fatal("negative button index should be invalid")
	}
	if err := (PressInlineButtonParams{ButtonIndex: 0}).Validate(); err != nil {
		t.Fatalf("button index zero should be valid: %v", err)
	}
}

func TestSendParamValidation(t *testing.T) {
	if err := (SendLocationParams{Latitude: 91, Longitude: 0}).Validate(); err == nil {
		t.Fatal("invalid latitude should fail")
	}
	if err := (SendLocationParams{Latitude: 45, Longitude: 181}).Validate(); err == nil {
		t.Fatal("invalid longitude should fail")
	}
	if err := (SendLocationParams{Latitude: 45, Longitude: 90}).Validate(); err != nil {
		t.Fatalf("valid location = %v", err)
	}

	if err := (SendPollParams{Options: []PollOption{{Text: "A"}}}).Validate(); err == nil {
		t.Fatal("poll with one option should fail")
	}
	if err := ValidateStruct(&SendPollParams{Question: "Q", Options: []PollOption{{Text: ""}, {Text: "B"}}}); err == nil {
		t.Fatal("poll option with empty text should fail struct validation")
	}
	if err := (SendPollParams{Options: make([]PollOption, 11)}).Validate(); err == nil {
		t.Fatal("poll with eleven options should fail")
	}
	poll := SendPollParams{Options: []PollOption{{Text: "A"}, {Text: "B"}}, CorrectIdx: 1}
	if err := poll.Validate(); err != nil {
		t.Fatalf("valid poll = %v", err)
	}
	if err := poll.ValidateForQuiz(); err != nil {
		t.Fatalf("valid quiz = %v", err)
	}
	poll.CorrectIdx = 3
	if err := poll.ValidateForQuiz(); err == nil {
		t.Fatal("quiz with out-of-range correct index should fail")
	}

	if err := (SendStickerParams{Peer: "@p"}).Validate(); err == nil {
		t.Fatal("sticker without id or file should fail")
	}
	if err := (SendStickerParams{Peer: "@p", StickerID: "id"}).Validate(); err != nil {
		t.Fatalf("sticker id valid = %v", err)
	}
	if err := (SendStickerParams{Peer: "@p", File: "s.webp"}).Validate(); err != nil {
		t.Fatalf("sticker file valid = %v", err)
	}
}

func TestOtherParamValidation(t *testing.T) {
	validators := []interface{ Validate() error }{
		CreateGroupParams{Title: "Group", Members: []string{"@a"}},
		InviteParams{Peer: "@p", Members: []string{"@a"}},
		DeleteContactParams{Username: "@a"},
		SendStarGiftParams{Peer: "@p", GiftID: 1},
		SendStarGiftParams{Peer: "@p", Name: "Heart"},
		TransferStarGiftParams{Peer: "@p", MsgID: 1},
		TransferStarGiftParams{Peer: "@p", Slug: "Gift-1"},
		ConvertStarGiftParams{MsgID: 1},
		ConvertStarGiftParams{Slug: "Gift-1"},
		UpdateGiftPriceParams{MsgID: 1, Price: 10},
		UpdateGiftPriceParams{Slug: "Gift-1", Price: 10},
		GetResaleGiftsParams{GiftID: 1},
		GetResaleGiftsParams{Name: "Heart"},
		GetGiftAttrsParams{GiftID: 1},
		GetGiftAttrsParams{Name: "Heart"},
		AddReactionParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}, Emoji: "ok"},
		RemoveReactionParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		ListReactionsParams{PeerInfo: PeerInfo{Peer: "@p"}, MsgID: MsgID{MessageID: 1}},
		GetUserInfoParams{Username: "@p"},
	}
	for _, params := range validators {
		if err := params.Validate(); err != nil {
			t.Fatalf("%T valid params failed: %v", params, err)
		}
	}

	invalid := []interface{ Validate() error }{
		CreateGroupParams{},
		InviteParams{Peer: "@p"},
		DeleteContactParams{},
		SendStarGiftParams{Peer: "@p"},
		TransferStarGiftParams{Peer: "@p"},
		ConvertStarGiftParams{},
		UpdateGiftPriceParams{MsgID: 1},
		GetResaleGiftsParams{},
		GetGiftAttrsParams{},
		ListReactionsParams{PeerInfo: PeerInfo{Peer: "@p"}},
		GetUserInfoParams{},
	}
	for _, params := range invalid {
		if err := params.Validate(); err == nil {
			t.Fatalf("%T invalid params passed", params)
		}
	}
}
