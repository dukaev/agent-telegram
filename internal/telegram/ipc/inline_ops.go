// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// InspectInlineButtonsHandler returns a handler for inspect_inline_buttons requests.
func InspectInlineButtonsHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.InspectInlineButtonsParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.InspectInlineButtons(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect inline buttons: %w", err)
		}

		return result, nil
	}
}

// PressInlineButtonHandler returns a handler for press_inline_button requests.
func PressInlineButtonHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.PressInlineButtonParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.PressInlineButton(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to press inline button: %w", err)
		}

		return result, nil
	}
}
