// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// PinMessageHandler returns a handler for pin_message requests.
func PinMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.PinMessage, "pin message")
}

// UnpinMessageHandler returns a handler for unpin_message requests.
func UnpinMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.UnpinMessage, "unpin message")
}
