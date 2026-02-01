// Package ipc provides Telegram IPC handlers.
package ipc

// SendMessageHandler returns a handler for send_message requests.
func SendMessageHandler(client Client) HandlerFunc {
	return Handler(client.Message().SendMessage, "send message")
}
