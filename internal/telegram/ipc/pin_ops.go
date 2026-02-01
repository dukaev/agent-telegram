// Package ipc provides Telegram IPC handlers.
package ipc

// PinMessageHandler returns a handler for pin_message requests.
func PinMessageHandler(client Client) HandlerFunc {
	return Handler(client.Pin().PinMessage, "pin message")
}

// UnpinMessageHandler returns a handler for unpin_message requests.
func UnpinMessageHandler(client Client) HandlerFunc {
	return Handler(client.Pin().UnpinMessage, "unpin message")
}
