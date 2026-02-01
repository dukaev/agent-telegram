// Package ipc provides Telegram IPC handlers for chat admin operations.
package ipc

// GetAdminsHandler returns a handler for get_admins requests.
func GetAdminsHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetAdmins, "get admins")
}

// GetBannedHandler returns a handler for get_banned requests.
func GetBannedHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetBanned, "get banned users")
}

// PromoteAdminHandler returns a handler for promote_admin requests.
func PromoteAdminHandler(client Client) HandlerFunc {
	return Handler(client.Chat().PromoteAdmin, "promote admin")
}

// DemoteAdminHandler returns a handler for demote_admin requests.
func DemoteAdminHandler(client Client) HandlerFunc {
	return Handler(client.Chat().DemoteAdmin, "demote admin")
}

// GetInviteLinkHandler returns a handler for get_invite_link requests.
func GetInviteLinkHandler(client Client) HandlerFunc {
	return Handler(client.Chat().GetInviteLink, "get invite link")
}
