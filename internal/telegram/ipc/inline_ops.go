// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// InspectInlineButtonsParams represents parameters for inspect_inline_buttons request.
type InspectInlineButtonsParams struct {
	Peer      string `json:"peer,omitempty"`
	Username  string `json:"username,omitempty"`
	MessageID int64  `json:"messageId"`
	Limit     int    `json:"limit,omitempty"`
}

// InspectInlineButtonsHandler returns a handler for inspect_inline_buttons requests.
func InspectInlineButtonsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p InspectInlineButtonsParams
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

		result, err := client.InspectInlineButtons(context.Background(), telegram.InspectInlineButtonsParams{
			Peer:      p.Peer,
			Username:  p.Username,
			MessageID: p.MessageID,
			Limit:     p.Limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to inspect inline buttons: %w", err)
		}

		return result, nil
	}
}

// PressInlineButtonParams represents parameters for press_inline_button request.
type PressInlineButtonParams struct {
	Peer        string `json:"peer,omitempty"`
	Username    string `json:"username,omitempty"`
	MessageID   int64  `json:"messageId"`
	ButtonText  string `json:"buttonText,omitempty"`
	ButtonIndex int    `json:"buttonIndex"`
}

// PressInlineButtonHandler returns a handler for press_inline_button requests.
func PressInlineButtonHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p PressInlineButtonParams
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
		if p.ButtonIndex < 0 {
			return nil, fmt.Errorf("buttonIndex must be >= 0")
		}

		result, err := client.PressInlineButton(context.Background(), telegram.PressInlineButtonParams{
			Peer:        p.Peer,
			Username:    p.Username,
			MessageID:   p.MessageID,
			ButtonText:  p.ButtonText,
			ButtonIndex: p.ButtonIndex,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to press inline button: %w", err)
		}

		return result, nil
	}
}
