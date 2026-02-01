// Package ipc provides Telegram IPC handlers.
package ipc

// GetMessagesHandler returns a handler for get_messages requests.
func GetMessagesHandler(client Client) HandlerFunc {
	return Handler(client.Message().GetMessages, "get messages")
}
