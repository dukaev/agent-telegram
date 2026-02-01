// Package ipc provides Telegram IPC handlers.
package ipc

// ClearMessagesHandler returns a handler for clear_messages requests.
func ClearMessagesHandler(client Client) HandlerFunc {
	return Handler(client.Chat().ClearMessages, "clear messages")
}

// ClearHistoryHandler returns a handler for clear_history requests.
func ClearHistoryHandler(client Client) HandlerFunc {
	return Handler(client.Chat().ClearHistory, "clear history")
}
