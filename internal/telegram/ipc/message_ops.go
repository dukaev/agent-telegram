// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendReplyParams represents parameters for send_reply request.
type SendReplyParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
}

// SendReplyHandler returns a handler for send_reply requests.
func SendReplyHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendReplyParams
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
		if p.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		result, err := client.SendReply(context.Background(), telegram.SendReplyParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
			Text:      p.Text,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send reply: %w", err)
		}

		return result, nil
	}
}

// UpdateMessageParams represents parameters for update_message request.
type UpdateMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Text      string `json:"text"`
}

// UpdateMessageHandler returns a handler for update_message requests.
func UpdateMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p UpdateMessageParams
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
		if p.Text == "" {
			return nil, fmt.Errorf("text is required")
		}

		result, err := client.UpdateMessage(context.Background(), telegram.UpdateMessageParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
			Text:      p.Text,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update message: %w", err)
		}

		return result, nil
	}
}

// DeleteMessageParams represents parameters for delete_message request.
type DeleteMessageParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
}

// DeleteMessageHandler returns a handler for delete_message requests.
func DeleteMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p DeleteMessageParams
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

		result, err := client.DeleteMessage(context.Background(), telegram.DeleteMessageParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete message: %w", err)
		}

		return result, nil
	}
}
