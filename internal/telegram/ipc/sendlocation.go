// Package ipc provides Telegram IPC handlers.
package ipc

// SendLocationHandler returns a handler for send_location requests.
func SendLocationHandler(client Client) HandlerFunc {
	return Handler(client.Media().SendLocation, "send location")
}
