// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendMessageParams represents parameters for send_message request.
type SendMessageParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
	Message  string `json:"message"`
}

// SendMessageHandler returns a handler for send_message requests.
func SendMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Validate
		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.Message == "" {
			return nil, fmt.Errorf("message is required")
		}

		result, err := client.SendMessage(context.Background(), telegram.SendMessageParams{
			Peer:     p.Peer,
			Username: p.Username,
			Message:  p.Message,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}

		return result, nil
	}
}
