// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"encoding/json"
)

// BlockPeerHandler returns a handler for block requests.
func BlockPeerHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.BlockPeer, "block peer")
}
