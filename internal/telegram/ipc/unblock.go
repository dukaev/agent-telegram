// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"encoding/json"
)

// UnblockPeerHandler returns a handler for unblock requests.
func UnblockPeerHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.UnblockPeer, "unblock peer")
}
