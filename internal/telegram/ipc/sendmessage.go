// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// SendMessageHandler returns a handler for send_message requests.
func SendMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SendMessage, "send message")
}
