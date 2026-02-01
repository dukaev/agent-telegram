// Package ipc provides Telegram IPC handlers.
package ipc

// BlockPeerHandler returns a handler for block requests.
func BlockPeerHandler(client Client) HandlerFunc {
	return Handler(client.User().BlockPeer, "block peer")
}
