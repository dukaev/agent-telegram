// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// UnblockPeerParams represents parameters for unblock request.
type UnblockPeerParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// UnblockPeerHandler returns a handler for unblock requests.
func UnblockPeerHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p UnblockPeerParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}

		result, err := client.UnblockPeer(context.Background(), telegram.UnblockPeerParams{
			Peer:     p.Peer,
			Username: p.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to unblock peer: %w", err)
		}

		return result, nil
	}
}
