// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

		"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// GetMeHandler returns a handler for get_me requests.
func GetMeHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(_ json.RawMessage) (interface{}, error) {
		user, err := client.GetMe(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		return types.GetMeResult{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Verified:  user.Verified,
			Bot:       user.Bot,
		}, nil
	}
}

// Client is an interface for Telegram operations.
type Client interface {
	GetMe(ctx context.Context) (*tg.User, error)
	GetChats(ctx context.Context, limit, offset int) ([]map[string]interface{}, error)
	GetUpdates(limit int) []types.StoredUpdate
	GetMessages(ctx context.Context, params types.GetMessagesParams) (*types.GetMessagesResult, error)
	GetUserInfo(ctx context.Context, params types.GetUserInfoParams) (*types.GetUserInfoResult, error)
	SendMessage(ctx context.Context, params types.SendMessageParams) (*types.SendMessageResult, error)
	SendLocation(ctx context.Context, params types.SendLocationParams) (*types.SendLocationResult, error)
	SendPhoto(ctx context.Context, params types.SendPhotoParams) (*types.SendPhotoResult, error)
	SendContact(ctx context.Context, params types.SendContactParams) (*types.SendContactResult, error)
	SendFile(ctx context.Context, params types.SendFileParams) (*types.SendFileResult, error)
	SendPoll(ctx context.Context, params types.SendPollParams) (*types.SendPollResult, error)
	SendVideo(ctx context.Context, params types.SendVideoParams) (*types.SendVideoResult, error)
	SendReply(ctx context.Context, params types.SendReplyParams) (*types.SendReplyResult, error)
	UpdateMessage(
		ctx context.Context, params types.UpdateMessageParams,
	) (*types.UpdateMessageResult, error)
	DeleteMessage(ctx context.Context, params types.DeleteMessageParams) (*types.DeleteMessageResult, error)
	ForwardMessage(ctx context.Context, params types.ForwardMessageParams) (*types.ForwardMessageResult, error)
	PinMessage(ctx context.Context, params types.PinMessageParams) (*types.PinMessageResult, error)
	UnpinMessage(ctx context.Context, params types.UnpinMessageParams) (*types.UnpinMessageResult, error)
	InspectInlineButtons(
		ctx context.Context, params types.InspectInlineButtonsParams,
	) (*types.InspectInlineButtonsResult, error)
	PressInlineButton(
		ctx context.Context, params types.PressInlineButtonParams,
	) (*types.PressInlineButtonResult, error)
	AddReaction(ctx context.Context, params types.AddReactionParams) (*types.AddReactionResult, error)
	RemoveReaction(ctx context.Context, params types.RemoveReactionParams) (*types.RemoveReactionResult, error)
	ListReactions(ctx context.Context, params types.ListReactionsParams) (*types.ListReactionsResult, error)
	UpdateProfile(ctx context.Context, params types.UpdateProfileParams) (*types.UpdateProfileResult, error)
	UpdateAvatar(ctx context.Context, params types.UpdateAvatarParams) (*types.UpdateAvatarResult, error)
	ClearMessages(ctx context.Context, params types.ClearMessagesParams) (*types.ClearMessagesResult, error)
	ClearHistory(ctx context.Context, params types.ClearHistoryParams) (*types.ClearHistoryResult, error)
	PinChat(ctx context.Context, params types.PinChatParams) (*types.PinChatResult, error)
	JoinChat(ctx context.Context, params types.JoinChatParams) (*types.JoinChatResult, error)
	SubscribeChannel(ctx context.Context, params types.SubscribeChannelParams) (*types.SubscribeChannelResult, error)
	BlockPeer(ctx context.Context, params types.BlockPeerParams) (*types.BlockPeerResult, error)
	UnblockPeer(ctx context.Context, params types.UnblockPeerParams) (*types.UnblockPeerResult, error)
	InspectReplyKeyboard(ctx context.Context, params types.PeerInfo) (*types.ReplyKeyboardResult, error)
	SearchGlobal(ctx context.Context, params types.SearchGlobalParams) (*types.SearchGlobalResult, error)
	SearchInChat(ctx context.Context, params types.SearchInChatParams) (*types.SearchInChatResult, error)
	GetContacts(ctx context.Context, params types.GetContactsParams) (*types.GetContactsResult, error)
	AddContact(ctx context.Context, params types.AddContactParams) (*types.AddContactResult, error)
	DeleteContact(ctx context.Context, params types.DeleteContactParams) (*types.DeleteContactResult, error)
	GetTopics(ctx context.Context, params types.GetTopicsParams) (*types.GetTopicsResult, error)
	CreateGroup(ctx context.Context, params types.CreateGroupParams) (*types.CreateGroupResult, error)
	CreateChannel(ctx context.Context, params types.CreateChannelParams) (*types.CreateChannelResult, error)
	EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error)
	SetPhoto(ctx context.Context, params types.SetPhotoParams) (*types.SetPhotoResult, error)
	DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error)
	Leave(ctx context.Context, params types.LeaveParams) (*types.LeaveResult, error)
	Invite(ctx context.Context, params types.InviteParams) (*types.InviteResult, error)
	GetParticipants(ctx context.Context, params types.GetParticipantsParams) (*types.GetParticipantsResult, error)
	GetAdmins(ctx context.Context, params types.GetAdminsParams) (*types.GetAdminsResult, error)
	GetBanned(ctx context.Context, params types.GetBannedParams) (*types.GetBannedResult, error)
	PromoteAdmin(ctx context.Context, params types.PromoteAdminParams) (*types.PromoteAdminResult, error)
	DemoteAdmin(ctx context.Context, params types.DemoteAdminParams) (*types.DemoteAdminResult, error)
	GetInviteLink(ctx context.Context, params types.GetInviteLinkParams) (*types.GetInviteLinkResult, error)
}
