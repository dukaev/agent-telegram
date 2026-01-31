// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// ClearMessagesHandler returns a handler for clear_messages requests.
func ClearMessagesHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.ClearMessagesParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.ClearMessages(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to clear messages: %w", err)
		}

		return result, nil
	}
}

// ClearHistoryHandler returns a handler for clear_history requests.
func ClearHistoryHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.ClearHistoryParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.ClearHistory(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to clear history: %w", err)
		}

		return result, nil
	}
}
