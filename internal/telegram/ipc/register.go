// Package ipc provides Telegram IPC handlers registration.
package ipc

import (
	"encoding/json"
	"errors"

	"agent-telegram/internal/ipc"
	"agent-telegram/telegram/client"
)

// methodHandlers maps method names to handler factory functions.
var methodHandlers = map[string]func(Client) HandlerFunc{
	// Basic
	"get_me":       GetMeHandler,
	"get_chats":    getChatsHandler,
	"get_updates":  GetUpdatesHandler,
	"get_messages": getMessagesHandler,
	"get_user_info": getUserInfoHandler,

	// Messages
	"send_message":    sendMessageHandler,
	"send_reply":      sendReplyHandler,
	"send_location":   sendLocationHandler,
	"send_photo":      sendPhotoHandler,
	"send_contact":    sendContactHandler,
	"send_file":       sendFileHandler,
	"send_document":   sendFileHandler, // alias for send_file
	"send_audio":      sendFileHandler, // alias for send_file
	"send_poll":       SendPollHandler,
	"send_checklist":  SendChecklistHandler,
	"send_video":      sendVideoHandler,
	"update_message":  updateMessageHandler,
	"delete_message":  deleteMessageHandler,
	"forward_message": forwardMessageHandler,
	"clear_messages":  clearMessagesHandler,
	"clear_history":   clearHistoryHandler,

	// Inline/keyboard
	"inspect_inline_buttons":  inspectInlineButtonsHandler,
	"press_inline_button":     pressInlineButtonHandler,
	"inspect_reply_keyboard":  inspectReplyKeyboardHandler,
	"pin_message":             pinMessageHandler,
	"unpin_message":           unpinMessageHandler,

	// Reactions
	"add_reaction":    addReactionHandler,
	"remove_reaction": removeReactionHandler,
	"list_reactions":  listReactionsHandler,

	// Chat operations
	"pin_chat":           func(c Client) HandlerFunc { return Handler(c.Chat().PinChat, "pin chat") },
	"archive":            func(c Client) HandlerFunc { return Handler(c.Chat().Archive, "archive chat") },
	"unarchive":          func(c Client) HandlerFunc { return Handler(c.Chat().Unarchive, "unarchive chat") },
	"mute":               func(c Client) HandlerFunc { return Handler(c.Chat().Mute, "mute chat") },
	"unmute":             func(c Client) HandlerFunc { return Handler(c.Chat().Unmute, "unmute chat") },
	"join_chat":          joinChatHandler,
	"subscribe_channel":  subscribeChannelHandler,
	"leave":              leaveHandler,
	"invite":             inviteHandler,
	"get_topics":         getTopicsHandler,
	"create_group":       createGroupHandler,
	"create_channel":     createChannelHandler,
	"edit_title":         editTitleHandler,
	"set_photo":          setPhotoHandler,
	"delete_photo":       deletePhotoHandler,
	"get_participants":   getParticipantsHandler,
	"get_admins":         getAdminsHandler,
	"get_banned":         getBannedHandler,
	"promote_admin":      promoteAdminHandler,
	"demote_admin":       demoteAdminHandler,
	"get_invite_link":    getInviteLinkHandler,

	// User operations
	"update_profile": updateProfileHandler,
	"update_avatar":  updateAvatarHandler,
	"block":          blockPeerHandler,
	"unblock":        unblockPeerHandler,
	"search_global":  searchGlobalHandler,
	"search_in_chat": searchInChatHandler,
	"get_contacts":   getContactsHandler,
	"add_contact":    addContactHandler,
	"delete_contact": deleteContactHandler,

	// New features
	"read_messages":          readMessagesHandler,
	"set_typing":             setTypingHandler,
	"get_scheduled_messages": getScheduledMessagesHandler,
	"send_voice":             sendVoiceHandler,
	"send_video_note":        sendVideoNoteHandler,
	"send_sticker":           sendStickerHandler,
	"get_sticker_packs":      getStickerPacksHandler,
	"send_gif":               sendGIFHandler,
	"set_slow_mode":          setSlowModeHandler,
	"set_chat_permissions":   setChatPermissionsHandler,
	"get_folders":            getFoldersHandler,
	"create_folder":          createFolderHandler,
	"delete_folder":          deleteFolderHandler,
	"get_privacy":            getPrivacyHandler,
	"set_privacy":            setPrivacyHandler,
}

// RegisterHandlers registers all Telegram IPC handlers.
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
	for method, factory := range methodHandlers {
		registerHandler(srv, method, factory(client))
	}
}

// registerHandler registers a single handler with error wrapping and panic recovery.
func registerHandler(srv ipc.MethodRegistrar, method string, handler HandlerFunc) {
	srv.Register(method, func(params json.RawMessage) (result interface{}, rpcErr *ipc.ErrorObject) {
		// Recover from panics (e.g., nil pointer when client not initialized)
		defer func() {
			if r := recover(); r != nil {
				// If panic is due to nil pointer, it's likely not initialized
				rpcErr = ipc.ErrNotInitialized
			}
		}()

		res, err := handler(params)
		if err != nil {
			// Check for specific error types
			if errors.Is(err, client.ErrNotInitialized) {
				return nil, ipc.ErrNotInitialized
			}
			return nil, &ipc.ErrorObject{Code: -32000, Message: err.Error()}
		}
		return res, nil
	})
}
