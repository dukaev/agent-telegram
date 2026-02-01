// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// JoinChatHandler returns a handler for join_chat requests.
func JoinChatHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.JoinChat, "join chat")
}
