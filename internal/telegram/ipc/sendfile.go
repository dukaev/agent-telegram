// Package ipc provides Telegram IPC handlers.
package ipc

import "agent-telegram/telegram/types"

// SendFileHandler returns a handler for send_file requests.
func SendFileHandler(client Client) HandlerFunc {
	return FileHandler(
		func(p types.SendFileParams) string { return p.File },
		client.Media().SendFile,
		"send file",
	)
}
