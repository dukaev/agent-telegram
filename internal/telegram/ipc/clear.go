// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// ClearMessagesHandler returns a handler for clear_messages requests.
func ClearMessagesHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.ClearMessages, "clear messages")
}

// ClearHistoryHandler returns a handler for clear_history requests.
func ClearHistoryHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.ClearHistory, "clear history")
}
