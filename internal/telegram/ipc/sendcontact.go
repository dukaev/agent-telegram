// Package ipc provides Telegram IPC handlers.
package ipc

// SendContactHandler returns a handler for send_contact requests.
func SendContactHandler(client Client) HandlerFunc {
	return Handler(client.Media().SendContact, "send contact")
}
