// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

		"agent-telegram/telegram/types"
)

// SendFileHandler returns a handler for send_file requests.
func SendFileHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(func(ctx context.Context, p types.SendFileParams) (*types.SendFileResult, error) {
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}
		return client.SendFile(ctx, p)
	}, "send file")
}
