// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"
)

// GetMeParams represents parameters for get_me request.
type GetMeParams struct{}

// GetMeResult represents the result of get_me request.
type GetMeResult struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Verified   bool   `json:"verified"`
	Bot        bool   `json:"bot"`
}

// GetMeHandler returns a handler for get_me requests.
func GetMeHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(_ json.RawMessage) (interface{}, error) {
		user, err := client.GetMe(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		return GetMeResult{
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
}
