// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetChatsParams represents parameters for get_chats request.
type GetChatsParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetChatsResult represents the result of get_chats request.
type GetChatsResult struct {
	Chats  []map[string]interface{} `json:"chats"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
	Count  int                     `json:"count"`
}

// GetChatsHandler returns a handler for get_chats requests.
func GetChatsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p GetChatsParams
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

		return GetChatsResult{
			Chats:  chats,
			Limit:  p.Limit,
			Offset: p.Offset,
			Count:  len(chats),
		}, nil
	}
}
