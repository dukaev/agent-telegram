// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/telegram"
)

// SendVideoHandler returns a handler for send_video requests.
func SendVideoHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.SendVideoParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}

		result, err := client.SendVideo(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to send video: %w", err)
		}

		return result, nil
	}
}
