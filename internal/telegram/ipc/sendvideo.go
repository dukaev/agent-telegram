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
	return Handler(func(ctx context.Context, p telegram.SendVideoParams) (*telegram.SendVideoResult, error) {
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}
		return client.SendVideo(ctx, p)
	}, "send video")
}
