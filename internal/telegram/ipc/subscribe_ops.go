// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// SubscribeChannelHandler returns a handler for subscribe_channel requests.
func SubscribeChannelHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SubscribeChannel, "subscribe channel")
}
