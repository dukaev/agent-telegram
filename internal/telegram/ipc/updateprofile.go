// Package ipc provides Telegram IPC handlers.
package ipc

import "agent-telegram/telegram/types"

// UpdateProfileHandler returns a handler for update_profile requests.
func UpdateProfileHandler(client Client) HandlerFunc {
	return Handler(client.User().UpdateProfile, "update profile")
}

// UpdateAvatarHandler returns a handler for update_avatar requests.
func UpdateAvatarHandler(client Client) HandlerFunc {
	return FileHandler(
		func(p types.UpdateAvatarParams) string { return p.File },
		client.User().UpdateAvatar,
		"update avatar",
	)
}
