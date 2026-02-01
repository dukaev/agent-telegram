// Package ipc provides Telegram IPC handlers.
package ipc

import "agent-telegram/telegram/types"

// Message handlers.
func sendMessageHandler(c Client) HandlerFunc { return Handler(c.Message().SendMessage, "send message") }
func sendReplyHandler(c Client) HandlerFunc   { return Handler(c.Message().SendReply, "send reply") }
func updateMessageHandler(c Client) HandlerFunc {
	return Handler(c.Message().UpdateMessage, "update message")
}
func deleteMessageHandler(c Client) HandlerFunc {
	return Handler(c.Message().DeleteMessage, "delete message")
}
func forwardMessageHandler(c Client) HandlerFunc {
	return Handler(c.Message().ForwardMessage, "forward message")
}
func getMessagesHandler(c Client) HandlerFunc { return Handler(c.Message().GetMessages, "get messages") }
func clearMessagesHandler(c Client) HandlerFunc {
	return Handler(c.Chat().ClearMessages, "clear messages")
}
func clearHistoryHandler(c Client) HandlerFunc {
	return Handler(c.Chat().ClearHistory, "clear history")
}

// Media handlers.
func sendPhotoHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.SendPhotoParams) string { return p.File }, c.Media().SendPhoto, "send photo")
}
func sendVideoHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.SendVideoParams) string { return p.File }, c.Media().SendVideo, "send video")
}
func sendFileHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.SendFileParams) string { return p.File }, c.Media().SendFile, "send file")
}
func sendLocationHandler(c Client) HandlerFunc {
	return Handler(c.Media().SendLocation, "send location")
}
func sendContactHandler(c Client) HandlerFunc { return Handler(c.Media().SendContact, "send contact") }

// Inline/keyboard handlers.
func inspectInlineButtonsHandler(c Client) HandlerFunc {
	return Handler(c.Message().InspectInlineButtons, "inspect inline buttons")
}
func pressInlineButtonHandler(c Client) HandlerFunc {
	return Handler(c.Message().PressInlineButton, "press inline button")
}
func inspectReplyKeyboardHandler(c Client) HandlerFunc {
	return Handler(c.Message().InspectReplyKeyboard, "inspect reply keyboard")
}

// Pin handlers.
func pinMessageHandler(c Client) HandlerFunc   { return Handler(c.Pin().PinMessage, "pin message") }
func unpinMessageHandler(c Client) HandlerFunc { return Handler(c.Pin().UnpinMessage, "unpin message") }

// Reaction handlers.
func addReactionHandler(c Client) HandlerFunc {
	return Handler(c.Reaction().AddReaction, "add reaction")
}
func removeReactionHandler(c Client) HandlerFunc {
	return Handler(c.Reaction().RemoveReaction, "remove reaction")
}
func listReactionsHandler(c Client) HandlerFunc {
	return Handler(c.Reaction().ListReactions, "list reactions")
}

// Chat handlers.
func getChatsHandler(c Client) HandlerFunc       { return Handler(c.Chat().GetChats, "get chats") }
func getTopicsHandler(c Client) HandlerFunc      { return Handler(c.Chat().GetTopics, "get topics") }
func joinChatHandler(c Client) HandlerFunc       { return Handler(c.Chat().JoinChat, "join chat") }
func subscribeChannelHandler(c Client) HandlerFunc {
	return Handler(c.Chat().SubscribeChannel, "subscribe channel")
}
func leaveHandler(c Client) HandlerFunc       { return Handler(c.Chat().Leave, "leave chat") }
func inviteHandler(c Client) HandlerFunc      { return Handler(c.Chat().Invite, "invite users") }
func createGroupHandler(c Client) HandlerFunc { return Handler(c.Chat().CreateGroup, "create group") }
func createChannelHandler(c Client) HandlerFunc {
	return Handler(c.Chat().CreateChannel, "create channel")
}
func editTitleHandler(c Client) HandlerFunc   { return Handler(c.Chat().EditTitle, "edit title") }
func setPhotoHandler(c Client) HandlerFunc    { return Handler(c.Chat().SetPhoto, "set photo") }
func deletePhotoHandler(c Client) HandlerFunc { return Handler(c.Chat().DeletePhoto, "delete photo") }
func getParticipantsHandler(c Client) HandlerFunc {
	return Handler(c.Chat().GetParticipants, "get participants")
}
func getAdminsHandler(c Client) HandlerFunc { return Handler(c.Chat().GetAdmins, "get admins") }
func getBannedHandler(c Client) HandlerFunc { return Handler(c.Chat().GetBanned, "get banned users") }
func promoteAdminHandler(c Client) HandlerFunc {
	return Handler(c.Chat().PromoteAdmin, "promote admin")
}
func demoteAdminHandler(c Client) HandlerFunc { return Handler(c.Chat().DemoteAdmin, "demote admin") }
func getInviteLinkHandler(c Client) HandlerFunc {
	return Handler(c.Chat().GetInviteLink, "get invite link")
}

// User handlers.
func getUserInfoHandler(c Client) HandlerFunc { return Handler(c.User().GetUserInfo, "get user info") }
func updateProfileHandler(c Client) HandlerFunc {
	return Handler(c.User().UpdateProfile, "update profile")
}
func updateAvatarHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.UpdateAvatarParams) string { return p.File }, c.User().UpdateAvatar, "update avatar")
}
func blockPeerHandler(c Client) HandlerFunc   { return Handler(c.User().BlockPeer, "block peer") }
func unblockPeerHandler(c Client) HandlerFunc { return Handler(c.User().UnblockPeer, "unblock peer") }

// Contact handlers.
func getContactsHandler(c Client) HandlerFunc { return Handler(c.User().GetContacts, "get contacts") }
func addContactHandler(c Client) HandlerFunc  { return Handler(c.User().AddContact, "add contact") }
func deleteContactHandler(c Client) HandlerFunc {
	return Handler(c.User().DeleteContact, "delete contact")
}

// Search handlers.
func searchGlobalHandler(c Client) HandlerFunc {
	return Handler(c.Search().SearchGlobal, "search global")
}
func searchInChatHandler(c Client) HandlerFunc {
	return Handler(c.Search().SearchInChat, "search in chat")
}

// New feature handlers.
func readMessagesHandler(c Client) HandlerFunc {
	return Handler(c.Message().ReadMessages, "read messages")
}
func setTypingHandler(c Client) HandlerFunc {
	return Handler(c.Message().SetTyping, "set typing")
}
func getScheduledMessagesHandler(c Client) HandlerFunc {
	return Handler(c.Message().GetScheduledMessages, "get scheduled messages")
}
func sendVoiceHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.SendVoiceParams) string { return p.File }, c.Media().SendVoice, "send voice")
}
func sendVideoNoteHandler(c Client) HandlerFunc {
	return FileHandler(
		func(p types.SendVideoNoteParams) string { return p.File }, c.Media().SendVideoNote, "send video note",
	)
}
func sendStickerHandler(c Client) HandlerFunc {
	return Handler(c.Media().SendSticker, "send sticker")
}
func getStickerPacksHandler(c Client) HandlerFunc {
	return Handler(c.Media().GetStickerPacks, "get sticker packs")
}
func sendGIFHandler(c Client) HandlerFunc {
	return FileHandler(func(p types.SendGIFParams) string { return p.File }, c.Media().SendGIF, "send gif")
}
func setSlowModeHandler(c Client) HandlerFunc {
	return Handler(c.Chat().SetSlowMode, "set slow mode")
}
func setChatPermissionsHandler(c Client) HandlerFunc {
	return Handler(c.Chat().SetChatPermissions, "set chat permissions")
}
func getFoldersHandler(c Client) HandlerFunc {
	return Handler(c.Chat().GetFolders, "get folders")
}
func createFolderHandler(c Client) HandlerFunc {
	return Handler(c.Chat().CreateFolder, "create folder")
}
func deleteFolderHandler(c Client) HandlerFunc {
	return Handler(c.Chat().DeleteFolder, "delete folder")
}
func getPrivacyHandler(c Client) HandlerFunc {
	return Handler(c.User().GetPrivacy, "get privacy")
}
func setPrivacyHandler(c Client) HandlerFunc {
	return Handler(c.User().SetPrivacy, "set privacy")
}
