// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// BlockPeerHandler returns a handler for block requests.
func BlockPeerHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.BlockPeerParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.BlockPeer(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to block peer: %w", err)
		}

		return result, nil
	}
}
