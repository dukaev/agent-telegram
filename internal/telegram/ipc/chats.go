// Package ipc provides Telegram IPC handlers.
package ipc

// GetChatsHandler returns a handler for get_chats requests.
func GetChatsHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetChats, "get chats")
}
