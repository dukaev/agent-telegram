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

// Client is the main interface for Telegram operations.
type Client interface {
	GetMe(ctx context.Context) (*tg.User, error)
	GetUpdates(limit int) []types.StoredUpdate
	Chat() telegram.ChatClient
	Message() telegram.MessageClient
	Media() telegram.MediaClient
	User() telegram.UserClient
	Pin() telegram.PinClient
	Reaction() telegram.ReactionClient
	Search() telegram.SearchClient
}
