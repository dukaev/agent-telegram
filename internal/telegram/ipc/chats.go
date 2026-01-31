// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

		"agent-telegram/telegram/types"
)

// GetChatsHandler returns a handler for get_chats requests.
func GetChatsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetChatsParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Set defaults
		if p.Limit <= 0 {
			p.Limit = 10
		}
		if p.Limit > 100 {
			p.Limit = 100
		}

		chats, err := client.GetChats(context.Background(), p.Limit, p.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get chats: %w", err)
		}

		return types.GetChatsResult{
			Chats:  chats,
			Limit:  p.Limit,
			Offset: p.Offset,
			Count:  len(chats),
		}, nil
	}
}
