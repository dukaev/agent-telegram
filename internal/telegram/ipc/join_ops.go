// Package ipc provides Telegram IPC handlers.
package ipc

// JoinChatHandler returns a handler for join_chat requests.
func JoinChatHandler(client Client) HandlerFunc {
	return Handler(client.Chat().JoinChat, "join chat")
}
