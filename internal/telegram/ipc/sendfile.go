// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Similar to send-video but for different command
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/telegram"
)

// SendFileParams represents parameters for send_file request.
type SendFileParams struct {
	Peer    string `json:"peer"`
	File    string `json:"file"`
	Caption string `json:"caption,omitempty"`
}

// SendFileHandler returns a handler for send_file requests.
func SendFileHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendFileParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

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

		result, err := client.SendFile(context.Background(), telegram.SendFileParams{
			Peer:    p.Peer,
			File:    p.File,
			Caption: p.Caption,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send file: %w", err)
		}

		return result, nil
	}
}
