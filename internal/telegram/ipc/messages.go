// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// GetMessagesParams represents parameters for get_messages request.
type GetMessagesParams struct {
	Username string `json:"username"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// GetMessagesHandler returns a handler for get_messages requests.
func GetMessagesHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p GetMessagesParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Validate username
		if p.Username == "" {
			return nil, fmt.Errorf("username is required")
		}

		// Set defaults
		if p.Limit <= 0 {
			p.Limit = 10
		}
		if p.Limit > 100 {
			p.Limit = 100
		}
		if p.Offset < 0 {
			p.Offset = 0
		}

		result, err := client.GetMessages(context.Background(), telegram.GetMessagesParams{
			Username: p.Username,
			Limit:    p.Limit,
			Offset:   p.Offset,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get messages: %w", err)
		}

		return result, nil
	}
}
