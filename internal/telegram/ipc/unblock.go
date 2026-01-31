// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// UnblockPeerHandler returns a handler for unblock requests.
func UnblockPeerHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.UnblockPeerParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.UnblockPeer(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to unblock peer: %w", err)
		}

		return result, nil
	}
}
