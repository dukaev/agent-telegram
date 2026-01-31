// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendReplyHandler returns a handler for send_reply requests.
func SendReplyHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.SendReplyParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.SendReply(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to send reply: %w", err)
		}

		return result, nil
	}
}

// UpdateMessageHandler returns a handler for update_message requests.
func UpdateMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.UpdateMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.UpdateMessage(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to update message: %w", err)
		}

		return result, nil
	}
}

// DeleteMessageHandler returns a handler for delete_message requests.
func DeleteMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.DeleteMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.DeleteMessage(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to delete message: %w", err)
		}

		return result, nil
	}
}
