// Package ipc provides Telegram IPC handlers registration.
package ipc

import (
	"encoding/json"

	"agent-telegram/internal/ipc"
)

// RegisterHandlers registers all Telegram IPC handlers.
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "get_me", GetMeHandler(client))
	registerHandler(srv, "get_chats", GetChatsHandler(client))
	registerHandler(srv, "get_updates", GetUpdatesHandler(client))
	registerHandler(srv, "get_messages", GetMessagesHandler(client))
	registerHandler(srv, "get_user_info", GetUserInfoHandler(client))
	registerHandler(srv, "send_message", SendMessageHandler(client))
	registerHandler(srv, "send_location", SendLocationHandler(client))
	registerHandler(srv, "send_photo", SendPhotoHandler(client))
	registerHandler(srv, "send_contact", SendContactHandler(client))
	registerHandler(srv, "send_file", SendFileHandler(client))
	registerHandler(srv, "send_poll", SendPollHandler(client))
	registerHandler(srv, "send_checklist", SendChecklistHandler(client))
	registerHandler(srv, "send_video", SendVideoHandler(client))
	registerHandler(srv, "send_reply", SendReplyHandler(client))
	registerHandler(srv, "update_message", UpdateMessageHandler(client))
	registerHandler(srv, "delete_message", DeleteMessageHandler(client))
	registerHandler(srv, "update_profile", UpdateProfileHandler(client))
	registerHandler(srv, "update_avatar", UpdateAvatarHandler(client))
	registerHandler(srv, "pin_message", PinMessageHandler(client))
	registerHandler(srv, "unpin_message", UnpinMessageHandler(client))
	registerHandler(srv, "inspect_inline_buttons", InspectInlineButtonsHandler(client))
	registerHandler(srv, "press_inline_button", PressInlineButtonHandler(client))
	registerHandler(srv, "add_reaction", AddReactionHandler(client))
	registerHandler(srv, "remove_reaction", RemoveReactionHandler(client))
	registerHandler(srv, "list_reactions", ListReactionsHandler(client))
	registerHandler(srv, "clear_messages", ClearMessagesHandler(client))
	registerHandler(srv, "clear_history", ClearHistoryHandler(client))
	registerHandler(srv, "block", BlockPeerHandler(client))
	registerHandler(srv, "unblock", UnblockPeerHandler(client))
}

// registerHandler registers a single handler with error wrapping.
func registerHandler(srv ipc.MethodRegistrar, method string, handler func(json.RawMessage) (interface{}, error)) {
	srv.Register(method, func(params json.RawMessage) (interface{}, *ipc.ErrorObject) {
		result, err := handler(params)
		if err != nil {
			return nil, &ipc.ErrorObject{
				Code:    -32000,
				Message: err.Error(),
			}
		}
		return result, nil
	})
}
