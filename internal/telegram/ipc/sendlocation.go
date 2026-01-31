// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendLocationHandler returns a handler for send_location requests.
func SendLocationHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.SendLocationParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.SendLocation(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to send location: %w", err)
		}

		return result, nil
	}
}
