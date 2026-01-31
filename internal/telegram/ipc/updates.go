// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// GetUpdatesParams represents parameters for get_updates request.
type GetUpdatesParams struct {
	Limit int `json:"limit"`
}

// GetUpdatesResult represents the result of get_updates request.
type GetUpdatesResult struct {
	Updates []telegram.StoredUpdate `json:"updates"`
	Count   int                     `json:"count"`
}

// GetUpdatesHandler returns a handler for get_updates requests.
func GetUpdatesHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p GetUpdatesParams
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

		updates := client.GetUpdates(p.Limit)

		return GetUpdatesResult{
			Updates: updates,
			Count:   len(updates),
		}, nil
	}
}
