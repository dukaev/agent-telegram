// Package ipc provides Telegram IPC handlers.
package ipc

// SearchGlobalHandler returns a handler for search_global requests.
func SearchGlobalHandler(client Client) HandlerFunc {
	return Handler(client.Search().SearchGlobal, "search globally")
}

// SearchInChatHandler returns a handler for search_in_chat requests.
func SearchInChatHandler(client Client) HandlerFunc {
	return Handler(client.Search().SearchInChat, "search in chat")
}
