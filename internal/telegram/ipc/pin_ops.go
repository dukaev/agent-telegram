// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// PinMessageHandler returns a handler for pin_message requests.
func PinMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.PinMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.PinMessage(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to pin message: %w", err)
		}

		return result, nil
	}
}

// UnpinMessageHandler returns a handler for unpin_message requests.
func UnpinMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.UnpinMessageParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.UnpinMessage(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to unpin message: %w", err)
		}

		return result, nil
	}
}
