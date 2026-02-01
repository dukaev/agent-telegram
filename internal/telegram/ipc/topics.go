// Package ipc provides Telegram IPC handlers.
package ipc

// GetTopicsHandler returns a handler for get_topics requests.
func GetTopicsHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetTopics, "get topics")
}
