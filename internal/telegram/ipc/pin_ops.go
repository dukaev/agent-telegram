// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// PinMessageParams represents parameters for pin_message request.
type PinMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// PinMessageHandler returns a handler for pin_message requests.
func PinMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p PinMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.MessageID == 0 {
			return nil, fmt.Errorf("messageId is required")
		}

		result, err := client.PinMessage(context.Background(), telegram.PinMessageParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to pin message: %w", err)
		}

		return result, nil
	}
}

// UnpinMessageParams represents parameters for unpin_message request.
type UnpinMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// UnpinMessageHandler returns a handler for unpin_message requests.
func UnpinMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p UnpinMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.MessageID == 0 {
			return nil, fmt.Errorf("messageId is required")
		}

		result, err := client.UnpinMessage(context.Background(), telegram.UnpinMessageParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to unpin message: %w", err)
		}

		return result, nil
	}
}
