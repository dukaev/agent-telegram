// Package ipc provides Telegram IPC handlers.
package ipc

// SubscribeChannelHandler returns a handler for subscribe_channel requests.
func SubscribeChannelHandler(client Client) HandlerFunc {
	return Handler(client.Chat().SubscribeChannel, "subscribe channel")
}
