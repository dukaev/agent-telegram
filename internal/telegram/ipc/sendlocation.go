// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// SendLocationHandler returns a handler for send_location requests.
func SendLocationHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SendLocation, "send location")
}
