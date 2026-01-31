// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// BlockPeerParams represents parameters for block request.
type BlockPeerParams struct {
	Peer     string `json:"peer,omitempty"`
	Username string `json:"username,omitempty"`
}

// BlockPeerHandler returns a handler for block requests.
func BlockPeerHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p BlockPeerParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}

		result, err := client.BlockPeer(context.Background(), telegram.BlockPeerParams{
			Peer:     p.Peer,
			Username: p.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to block peer: %w", err)
		}

		return result, nil
	}
}
