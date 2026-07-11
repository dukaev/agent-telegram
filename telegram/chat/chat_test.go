package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"

	"agent-telegram/internal/strictjson"
	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

type fakeParent struct {
	peers map[string]tg.InputPeerClass
}

func (f fakeParent) ResolvePeer(_ context.Context, peer string) (tg.InputPeerClass, error) {
	return f.peers[peer], nil
}

func (fakeParent) CachePeer(string, tg.InputPeerClass) {}

func TestClientMethodsRequireInitialization(t *testing.T) {
	c := NewClient(nil)
	ctx := context.Background()
	check := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}

	_, err := c.GetChats(ctx, &types.GetChatsParams{})
	check("GetChats", err)
	_, err = c.CreateGroup(ctx, types.CreateGroupParams{})
	check("CreateGroup", err)
	_, err = c.CreateChannel(ctx, types.CreateChannelParams{})
	check("CreateChannel", err)
	_, err = c.GetFolders(ctx, types.GetFoldersParams{})
	check("GetFolders", err)
	_, err = c.CreateFolder(ctx, types.CreateFolderParams{})
	check("CreateFolder", err)
	_, err = c.DeleteFolder(ctx, types.DeleteFolderParams{})
	check("DeleteFolder", err)
	_, err = c.Join(ctx, "https://t.me/+hash")
	check("Join", err)
	_, err = c.JoinChat(ctx, types.JoinChatParams{InviteLink: "https://t.me/+hash"})
	check("JoinChat", err)
	_, err = c.Subscribe(ctx, "@channel")
	check("Subscribe", err)
	_, err = c.SubscribeChannel(ctx, types.SubscribeChannelParams{Channel: "@channel"})
	check("SubscribeChannel", err)
	_, err = c.Leave(ctx, types.LeaveParams{Peer: "@chat"})
	check("Leave", err)
	_, err = c.Invite(ctx, types.InviteParams{Peer: "@chat", Members: []string{"@user"}})
	check("Invite", err)
	_, err = c.PromoteAdmin(ctx, types.PromoteAdminParams{})
	check("PromoteAdmin", err)
	_, err = c.DemoteAdmin(ctx, types.DemoteAdminParams{})
	check("DemoteAdmin", err)
	_, err = c.GetParticipants(ctx, types.GetParticipantsParams{})
	check("GetParticipants", err)
	_, err = c.GetAdmins(ctx, types.GetAdminsParams{})
	check("GetAdmins", err)
	_, err = c.GetBanned(ctx, types.GetBannedParams{})
	check("GetBanned", err)
	_, err = c.Archive(ctx, types.ArchiveParams{})
	check("Archive", err)
	_, err = c.Unarchive(ctx, types.UnarchiveParams{})
	check("Unarchive", err)
	_, err = c.Mute(ctx, types.MuteParams{})
	check("Mute", err)
	_, err = c.Unmute(ctx, types.UnmuteParams{})
	check("Unmute", err)
	_, err = c.EditTitle(ctx, types.EditTitleParams{})
	check("EditTitle", err)
	_, err = c.SetPhoto(ctx, types.SetPhotoParams{})
	check("SetPhoto", err)
	_, err = c.DeletePhoto(ctx, types.DeletePhotoParams{})
	check("DeletePhoto", err)
	_, err = c.SetSlowMode(ctx, types.SetSlowModeParams{})
	check("SetSlowMode", err)
	_, err = c.SetChatPermissions(ctx, types.SetChatPermissionsParams{})
	check("SetChatPermissions", err)
	_, err = c.ClearMessages(ctx, types.ClearMessagesParams{})
	check("ClearMessages", err)
	_, err = c.ClearHistory(ctx, types.ClearHistoryParams{})
	check("ClearHistory", err)
	_, err = c.PinChat(ctx, types.PinChatParams{})
	check("PinChat", err)
	_, err = c.GetTopics(ctx, types.GetTopicsParams{})
	check("GetTopics", err)
	_, err = c.GetInviteLink(ctx, types.GetInviteLinkParams{})
	check("GetInviteLink", err)
}

func TestInviteHashAndDialogMapping(t *testing.T) {
	for _, link := range []string{
		" https://t.me/+abc ",
		"https://t.me/joinchat/abc",
		"tg://join?invite=abc",
		"+abc",
		"abc",
	} {
		hash, err := extractInviteHash(link)
		if err != nil || hash != "abc" {
			t.Fatalf("extractInviteHash(%q) = %q, %v", link, hash, err)
		}
	}
	if _, err := extractInviteHash(" \n\t"); err == nil {
		t.Fatal("empty invite link should fail")
	}

	data, err := extractDialogData(&tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerUser{UserID: 1}, TopMessage: 10, UnreadCount: 2}},
		Users:   []tg.UserClass{&tg.User{ID: 1, FirstName: "Ada", Username: "ada", Bot: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	userMap := buildUserMap(data.users)
	chats := convertDialogsToResult(data.dialogs, buildChatMap(data.chats), userMap)
	if len(chats) != 1 || chats[0]["type"] != "user" || chats[0]["username"] != "ada" {
		t.Fatalf("user dialog = %#v", chats)
	}

	channelChats := buildChatMap([]tg.ChatClass{&tg.Channel{ID: 2, Title: "News", Username: "news", Photo: &tg.ChatPhotoEmpty{}}})
	channelDialogs := []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}}}
	chats = convertDialogsToResult(channelDialogs, channelChats, nil)
	if len(chats) != 1 || chats[0]["type"] != "channel" || chats[0]["title"] != "News" {
		t.Fatalf("channel dialog = %#v", chats)
	}

	if _, err := extractDialogData(&tg.MessagesDialogsNotModified{}); err == nil {
		t.Fatal("not modified dialogs should fail")
	}
	if _, err := extractDialogData(nil); err == nil {
		t.Fatal("unknown dialogs type should fail")
	}
}

func TestGetChatsWithFakeAPI(t *testing.T) {
	c := NewClient(nil)
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.MessagesGetDialogsRequest:
			return &tg.MessagesDialogs{
				Dialogs: []tg.DialogClass{
					&tg.Dialog{Peer: &tg.PeerUser{UserID: 1}, TopMessage: 10, UnreadCount: 2},
					&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: 2}, TopMessage: 20},
				},
				Users: []tg.UserClass{
					&tg.User{ID: 1, FirstName: "Ada", Username: "ada", Bot: true},
				},
				Chats: []tg.ChatClass{
					&tg.Channel{ID: 2, Title: "News", Username: "news", Photo: &tg.ChatPhotoEmpty{}},
				},
			}, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))

	result, err := c.GetChats(context.Background(), &types.GetChatsParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 2 || result.Chats[0]["type"] != "user" || result.Chats[1]["type"] != "channel" {
		t.Fatalf("result = %+v", result)
	}
}

func TestParticipantParamsAcceptOffset(t *testing.T) {
	for _, target := range []any{
		&types.GetParticipantsParams{},
		&types.GetAdminsParams{},
		&types.GetBannedParams{},
	} {
		if err := strictjson.Decode([]byte(`{"peer":"@chat","offset":3}`), target); err != nil {
			t.Fatalf("decode %T: %v", target, err)
		}
	}
}

func TestGetParticipantsPagination(t *testing.T) {
	t.Run("channel forwards offset", func(t *testing.T) {
		c := NewClient(fakeParent{peers: map[string]tg.InputPeerClass{
			"@channel": &tg.InputPeerChannel{ChannelID: 7, AccessHash: 8},
		}})
		c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
			request, ok := input.(*tg.ChannelsGetParticipantsRequest)
			if !ok {
				t.Fatalf("unexpected request %T", input)
			}
			if request.Offset != 2 || request.Limit != 2 {
				t.Fatalf("pagination = offset %d, limit %d", request.Offset, request.Limit)
			}
			return &tg.ChannelsChannelParticipants{
				Count: 4,
				Participants: []tg.ChannelParticipantClass{
					&tg.ChannelParticipant{UserID: 3},
					&tg.ChannelParticipant{UserID: 4},
				},
				Users: []tg.UserClass{
					&tg.User{ID: 3, FirstName: "Three"},
					&tg.User{ID: 4, FirstName: "Four"},
				},
			}, nil
		})))

		result, err := c.GetParticipants(context.Background(), types.GetParticipantsParams{
			Peer: "@channel", Limit: 2, Offset: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count != 4 || len(result.Participants) != 2 || result.Participants[0].ID != 3 {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("basic group slices locally", func(t *testing.T) {
		c := NewClient(fakeParent{peers: map[string]tg.InputPeerClass{
			"@group": &tg.InputPeerChat{ChatID: 9},
		}})
		c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
			if _, ok := input.(*tg.MessagesGetFullChatRequest); !ok {
				t.Fatalf("unexpected request %T", input)
			}
			return &tg.MessagesChatFull{
				FullChat: &tg.ChatFull{Participants: &tg.ChatParticipants{
					Participants: []tg.ChatParticipantClass{
						&tg.ChatParticipant{UserID: 1},
						&tg.ChatParticipant{UserID: 2},
						&tg.ChatParticipant{UserID: 3},
					},
				}},
				Users: []tg.UserClass{
					&tg.User{ID: 1, FirstName: "One"},
					&tg.User{ID: 2, FirstName: "Two"},
					&tg.User{ID: 3, FirstName: "Three"},
				},
			}, nil
		})))

		result, err := c.GetParticipants(context.Background(), types.GetParticipantsParams{
			Peer: "@group", Limit: 1, Offset: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count != 3 || len(result.Participants) != 1 || result.Participants[0].ID != 2 {
			t.Fatalf("result = %+v", result)
		}
	})
}

func TestAlreadyParticipantMembershipIsIdempotent(t *testing.T) {
	alreadyParticipant := tgerr.New(400, tg.ErrUserAlreadyParticipant)

	t.Run("join", func(t *testing.T) {
		c := NewClient(nil)
		c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
			if _, ok := input.(*tg.MessagesImportChatInviteRequest); !ok {
				t.Fatalf("unexpected request %T", input)
			}
			return nil, alreadyParticipant
		})))

		result, err := c.JoinChat(context.Background(), types.JoinChatParams{InviteLink: "+hash"})
		if err != nil || result == nil || !result.Success {
			t.Fatalf("result = %+v, err = %v", result, err)
		}
	})

	t.Run("subscribe", func(t *testing.T) {
		c := NewClient(fakeParent{peers: map[string]tg.InputPeerClass{
			"@channel": &tg.InputPeerChannel{ChannelID: 7, AccessHash: 8},
		}})
		c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
			if _, ok := input.(*tg.ChannelsJoinChannelRequest); !ok {
				t.Fatalf("unexpected request %T", input)
			}
			return nil, alreadyParticipant
		})))

		result, err := c.SubscribeChannel(context.Background(), types.SubscribeChannelParams{Channel: "@channel"})
		if err != nil || result == nil || !result.Success {
			t.Fatalf("result = %+v, err = %v", result, err)
		}
	})

	t.Run("invite continues", func(t *testing.T) {
		c := NewClient(fakeParent{peers: map[string]tg.InputPeerClass{
			"@group":   &tg.InputPeerChat{ChatID: 9},
			"@already": &tg.InputPeerUser{UserID: 1, AccessHash: 11},
			"@new":     &tg.InputPeerUser{UserID: 2, AccessHash: 22},
		}})
		c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
			request, ok := input.(*tg.MessagesAddChatUserRequest)
			if !ok {
				t.Fatalf("unexpected request %T", input)
			}
			if request.UserID.(*tg.InputUser).UserID == 1 {
				return nil, alreadyParticipant
			}
			return &tg.MessagesInvitedUsers{Updates: &tg.Updates{}}, nil
		})))

		result, err := c.Invite(context.Background(), types.InviteParams{
			Peer: "@group", Members: []string{"@already", "@new"},
		})
		if err != nil || result == nil || !result.Success {
			t.Fatalf("result = %+v, err = %v", result, err)
		}
	})
}
