// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram/types"
)

// GetTopicsHandler returns a handler for get_topics requests.
func GetTopicsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetTopicsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.GetTopics(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get topics: %w", err)
		}

		return result, nil
	}
}
