// Package ipc provides Telegram IPC handlers.
package ipc

import "agent-telegram/telegram/types"

// SendVideoHandler returns a handler for send_video requests.
func SendVideoHandler(client Client) HandlerFunc {
	return FileHandler(
		func(p types.SendVideoParams) string { return p.File },
		client.Media().SendVideo,
		"send video",
	)
}
