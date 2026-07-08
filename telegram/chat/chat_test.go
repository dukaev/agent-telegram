package chat

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
