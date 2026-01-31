// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// ClearMessagesParams represents parameters for clear_messages request.
type ClearMessagesParams struct {
	Peer      string   `json:"peer,omitempty"`
	Username  string   `json:"username,omitempty"`
	MessageIDs []int64 `json:"messageIds"`
}

// ClearMessagesHandler returns a handler for clear_messages requests.
func ClearMessagesHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p ClearMessagesParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if len(p.MessageIDs) == 0 {
			return nil, fmt.Errorf("messageIds is required")
		}

		result, err := client.ClearMessages(context.Background(), telegram.ClearMessagesParams{
			Peer:       p.Peer,
			Username:   p.Username,
			MessageIDs: p.MessageIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clear messages: %w", err)
		}

		return result, nil
	}
}

// ClearHistoryParams represents parameters for clear_history request.
type ClearHistoryParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
	Revoke   bool   `json:"revoke,omitempty"`
}

// ClearHistoryHandler returns a handler for clear_history requests.
func ClearHistoryHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p ClearHistoryParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}

		result, err := client.ClearHistory(context.Background(), telegram.ClearHistoryParams{
			Peer:     p.Peer,
			Username: p.Username,
			Revoke:   p.Revoke,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clear history: %w", err)
		}

		return result, nil
	}
}
