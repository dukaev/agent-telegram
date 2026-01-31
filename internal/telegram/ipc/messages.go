// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// GetMessagesHandler returns a handler for get_messages requests.
func GetMessagesHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p telegram.GetMessagesParams
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

		result, err := client.GetMessages(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages: %w", err)
		}

		return result, nil
	}
}
