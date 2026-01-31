// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// SendContactHandler returns a handler for send_contact requests.
func SendContactHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SendContact, "send contact")
}
