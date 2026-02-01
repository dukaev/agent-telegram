// Package ipc provides Telegram IPC handlers.
package ipc

import "agent-telegram/telegram/types"

// SendPhotoHandler returns a handler for send_photo requests.
func SendPhotoHandler(client Client) HandlerFunc {
	return FileHandler(
		func(p types.SendPhotoParams) string { return p.File },
		client.Media().SendPhoto,
		"send photo",
	)
}
