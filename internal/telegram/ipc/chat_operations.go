// Package ipc provides Telegram IPC handlers for chat operations.
package ipc

// CreateGroupHandler returns a handler for create_group requests.
func CreateGroupHandler(client Client) HandlerFunc {
	return Handler(client.Chat().CreateGroup, "create group")
}

// CreateChannelHandler returns a handler for create_channel requests.
func CreateChannelHandler(client Client) HandlerFunc {
	return Handler(client.Chat().CreateChannel, "create channel")
}

// EditTitleHandler returns a handler for edit_title requests.
func EditTitleHandler(client Client) HandlerFunc {
	return Handler(client.Chat().EditTitle, "edit title")
}

// SetPhotoHandler returns a handler for set_photo requests.
func SetPhotoHandler(client Client) HandlerFunc {
	return Handler(client.Chat().SetPhoto, "set photo")
}

// DeletePhotoHandler returns a handler for delete_photo requests.
func DeletePhotoHandler(client Client) HandlerFunc {
	return Handler(client.Chat().DeletePhoto, "delete photo")
}

// LeaveHandler returns a handler for leave requests.
func LeaveHandler(client Client) HandlerFunc {
	return Handler(client.Chat().Leave, "leave chat")
}

// InviteHandler returns a handler for invite requests.
func InviteHandler(client Client) HandlerFunc {
	return Handler(client.Chat().Invite, "invite users")
}

// GetParticipantsHandler returns a handler for get_participants requests.
func GetParticipantsHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetParticipants, "get participants")
}
