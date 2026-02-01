// Package ipc provides Telegram IPC handlers.
package ipc

// GetUserInfoHandler returns a handler for get_user_info requests.
func GetUserInfoHandler(client Client) HandlerFunc {
	return Handler(client.User().GetUserInfo, "get user info")
}
