// Package ipc provides Telegram IPC handlers.
package ipc

// UnblockPeerHandler returns a handler for unblock requests.
func UnblockPeerHandler(client Client) HandlerFunc {
	return Handler(client.User().UnblockPeer, "unblock peer")
}
