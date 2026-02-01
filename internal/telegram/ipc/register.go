// Package ipc provides Telegram IPC handlers registration.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/internal/ipc"
	"agent-telegram/telegram/types"
)

// RegisterHandlers registers all Telegram IPC handlers.
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
	registerBasicHandlers(srv, client)
	registerMessageHandlers(srv, client)
	registerMediaHandlers(srv, client)
	registerReactionHandlers(srv, client)
	registerChatHandlers(srv, client)
	registerUserHandlers(srv, client)
}

// registerBasicHandlers registers basic info handlers.
func registerBasicHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "get_me", GetMeHandler(client))
	registerHandler(srv, "get_chats", GetChatsHandler(client))
	registerHandler(srv, "get_updates", GetUpdatesHandler(client))
	registerHandler(srv, "get_messages", GetMessagesHandler(client))
	registerHandler(srv, "get_user_info", GetUserInfoHandler(client))
}

// registerMessageHandlers registers message operation handlers.
func registerMessageHandlers(srv ipc.MethodRegistrar, client Client) {
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
	registerHandler(srv, "forward_message", ForwardMessageHandler(client))
	registerHandler(srv, "clear_messages", ClearMessagesHandler(client))
	registerHandler(srv, "clear_history", ClearHistoryHandler(client))
}

// registerMediaHandlers registers media and keyboard handlers.
func registerMediaHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "inspect_inline_buttons", InspectInlineButtonsHandler(client))
	registerHandler(srv, "press_inline_button", PressInlineButtonHandler(client))
	registerHandler(srv, "inspect_reply_keyboard", InspectReplyKeyboardHandler(client))
	registerHandler(srv, "pin_message", PinMessageHandler(client))
	registerHandler(srv, "unpin_message", UnpinMessageHandler(client))
}

// registerReactionHandlers registers reaction handlers.
func registerReactionHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "add_reaction", AddReactionHandler(client))
	registerHandler(srv, "remove_reaction", RemoveReactionHandler(client))
	registerHandler(srv, "list_reactions", ListReactionsHandler(client))
}

// registerChatHandlers registers chat operation handlers.
func registerChatHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "pin_chat", Handler(client.PinChat, "pin chat"))
	registerHandler(srv, "join_chat", JoinChatHandler(client))
	registerHandler(srv, "subscribe_channel", SubscribeChannelHandler(client))
	registerHandler(srv, "leave", LeaveHandler(client))
	registerHandler(srv, "invite", InviteHandler(client))
	registerHandler(srv, "get_topics", GetTopicsHandler(client))
	registerHandler(srv, "create_group", CreateGroupHandler(client))
	registerHandler(srv, "create_channel", CreateChannelHandler(client))
	registerHandler(srv, "edit_title", EditTitleHandler(client))
	registerHandler(srv, "set_photo", SetPhotoHandler(client))
	registerHandler(srv, "delete_photo", DeletePhotoHandler(client))
	registerHandler(srv, "get_participants", GetParticipantsHandler(client))
	registerHandler(srv, "get_admins", GetAdminsHandler(client))
	registerHandler(srv, "get_banned", GetBannedHandler(client))
	registerHandler(srv, "promote_admin", PromoteAdminHandler(client))
	registerHandler(srv, "demote_admin", DemoteAdminHandler(client))
	registerHandler(srv, "get_invite_link", GetInviteLinkHandler(client))
}

// registerUserHandlers registers user and contact handlers.
func registerUserHandlers(srv ipc.MethodRegistrar, client Client) {
	registerHandler(srv, "update_profile", UpdateProfileHandler(client))
	registerHandler(srv, "update_avatar", UpdateAvatarHandler(client))
	registerHandler(srv, "block", BlockPeerHandler(client))
	registerHandler(srv, "unblock", UnblockPeerHandler(client))
	registerHandler(srv, "search_global", SearchGlobalHandler(client))
	registerHandler(srv, "search_in_chat", SearchInChatHandler(client))
	registerHandler(srv, "get_contacts", GetContactsHandler(client))
	registerHandler(srv, "add_contact", AddContactHandler(client))
	registerHandler(srv, "delete_contact", DeleteContactHandler(client))
}

// InspectReplyKeyboardHandler returns a handler for inspecting reply keyboard requests.
func InspectReplyKeyboardHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.PeerInfo
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.ValidatePeer(); err != nil {
			return nil, err
		}

		result, err := client.InspectReplyKeyboard(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect reply keyboard: %w", err)
		}

		return result, nil
	}
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
