// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Similar to send-file but for different command
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/telegram"
)

// SendVideoParams represents parameters for send_video request.
type SendVideoParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
	File     string `json:"file"`
	Caption  string `json:"caption,omitempty"`
}

// SendVideoHandler returns a handler for send_video requests.
func SendVideoHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendVideoParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.File == "" {
			return nil, fmt.Errorf("file is required")
		}

		// Check file exists
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}

		result, err := client.SendVideo(context.Background(), telegram.SendVideoParams{
			Peer:     p.Peer,
			Username: p.Username,
			File:     p.File,
			Caption:  p.Caption,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send video: %w", err)
		}

		return result, nil
	}
}
