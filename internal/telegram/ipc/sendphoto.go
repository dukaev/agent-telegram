// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

		"agent-telegram/telegram/types"
)

// SendPhotoHandler returns a handler for send_photo requests.
func SendPhotoHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(func(ctx context.Context, p types.SendPhotoParams) (*types.SendPhotoResult, error) {
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}
		return client.SendPhoto(ctx, p)
	}, "send photo")
}
