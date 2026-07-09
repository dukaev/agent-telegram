// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram"
	"agent-telegram/telegram/types"
)

// GetMeHandler returns a handler for get_me requests.
func GetMeHandler(client Client) HandlerFunc {
	return func(ctx context.Context, _ json.RawMessage) (interface{}, error) {
		user, err := client.GetMe(ctx)
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

// Client is the main interface for Telegram operations.
type Client interface {
	GetMe(ctx context.Context) (*tg.User, error)
	GetUpdates(limit int, offset ...int64) []types.StoredUpdate
	GetUpdatePage(limit int, offset int64, epoch string) telegram.UpdatePage
	Chat() telegram.ChatClient
	Message() telegram.MessageClient
	Media() telegram.MediaClient
	User() telegram.UserClient
	Pin() telegram.PinClient
	Reaction() telegram.ReactionClient
	Search() telegram.SearchClient
	Gift() telegram.GiftClient
}
