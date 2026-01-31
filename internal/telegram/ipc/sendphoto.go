// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/telegram"
)

// SendPhotoParams represents parameters for send_photo request.
type SendPhotoParams struct {
	Peer string `json:"peer"`
	File string `json:"file"`
}

// SendPhotoHandler returns a handler for send_photo requests.
func SendPhotoHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendPhotoParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Validate
		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.File == "" {
			return nil, fmt.Errorf("file is required")
		}

		// Check file exists
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}

		result, err := client.SendPhoto(context.Background(), telegram.SendPhotoParams{
			Peer: p.Peer,
			File: p.File,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send photo: %w", err)
		}

		return result, nil
	}
}
