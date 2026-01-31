// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

		"agent-telegram/telegram/types"
)

// GetUserInfoParams represents parameters for get_user_info request.
type GetUserInfoParams struct {
	Username string `json:"username"`
}

// GetUserInfoHandler returns a handler for get_user_info requests.
func GetUserInfoHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p GetUserInfoParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Username == "" {
			return nil, fmt.Errorf("username is required")
		}

		result, err := client.GetUserInfo(context.Background(), types.GetUserInfoParams{
			Username: p.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}

		return result, nil
	}
}
