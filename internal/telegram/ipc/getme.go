// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
	"github.com/gotd/td/tg"
)

// GetMeHandler returns a handler for get_me requests.
func GetMeHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(_ json.RawMessage) (interface{}, error) {
		user, err := client.GetMe(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		return telegram.GetMeResult{
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
	GetUpdates(limit int) []telegram.StoredUpdate
	GetMessages(ctx context.Context, params telegram.GetMessagesParams) (*telegram.GetMessagesResult, error)
	GetUserInfo(ctx context.Context, params telegram.GetUserInfoParams) (*telegram.GetUserInfoResult, error)
	SendMessage(ctx context.Context, params telegram.SendMessageParams) (*telegram.SendMessageResult, error)
	SendLocation(ctx context.Context, params telegram.SendLocationParams) (*telegram.SendLocationResult, error)
	SendPhoto(ctx context.Context, params telegram.SendPhotoParams) (*telegram.SendPhotoResult, error)
	SendContact(ctx context.Context, params telegram.SendContactParams) (*telegram.SendContactResult, error)
	SendFile(ctx context.Context, params telegram.SendFileParams) (*telegram.SendFileResult, error)
	SendPoll(ctx context.Context, params telegram.SendPollParams) (*telegram.SendPollResult, error)
	SendVideo(ctx context.Context, params telegram.SendVideoParams) (*telegram.SendVideoResult, error)
	SendReply(ctx context.Context, params telegram.SendReplyParams) (*telegram.SendReplyResult, error)
	UpdateMessage(
		ctx context.Context, params telegram.UpdateMessageParams,
	) (*telegram.UpdateMessageResult, error)
	DeleteMessage(ctx context.Context, params telegram.DeleteMessageParams) (*telegram.DeleteMessageResult, error)
	PinMessage(ctx context.Context, params telegram.PinMessageParams) (*telegram.PinMessageResult, error)
	UnpinMessage(ctx context.Context, params telegram.UnpinMessageParams) (*telegram.UnpinMessageResult, error)
	InspectInlineButtons(
		ctx context.Context, params telegram.InspectInlineButtonsParams,
	) (*telegram.InspectInlineButtonsResult, error)
	PressInlineButton(
		ctx context.Context, params telegram.PressInlineButtonParams,
	) (*telegram.PressInlineButtonResult, error)
	AddReaction(ctx context.Context, params telegram.AddReactionParams) (*telegram.AddReactionResult, error)
	RemoveReaction(ctx context.Context, params telegram.RemoveReactionParams) (*telegram.RemoveReactionResult, error)
	ListReactions(ctx context.Context, params telegram.ListReactionsParams) (*telegram.ListReactionsResult, error)
	UpdateProfile(ctx context.Context, params telegram.UpdateProfileParams) (*telegram.UpdateProfileResult, error)
	UpdateAvatar(ctx context.Context, params telegram.UpdateAvatarParams) (*telegram.UpdateAvatarResult, error)
	ClearMessages(ctx context.Context, params telegram.ClearMessagesParams) (*telegram.ClearMessagesResult, error)
	ClearHistory(ctx context.Context, params telegram.ClearHistoryParams) (*telegram.ClearHistoryResult, error)
	BlockPeer(ctx context.Context, params telegram.BlockPeerParams) (*telegram.BlockPeerResult, error)
	UnblockPeer(ctx context.Context, params telegram.UnblockPeerParams) (*telegram.UnblockPeerResult, error)
}
