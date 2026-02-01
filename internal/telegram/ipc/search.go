// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram/types"
)

// SearchGlobalHandler returns a handler for search_global requests.
func SearchGlobalHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.SearchGlobalParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if p.Query == "" {
			return nil, fmt.Errorf("query is required")
		}

		result, err := client.SearchGlobal(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to search globally: %w", err)
		}

		return result, nil
	}
}

// SearchInChatHandler returns a handler for search_in_chat requests.
func SearchInChatHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.SearchInChatParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.Query == "" {
			return nil, fmt.Errorf("query is required")
		}

		result, err := client.SearchInChat(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to search in chat: %w", err)
		}

		return result, nil
	}
}
