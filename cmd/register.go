// Package cmd provides command registration.
package cmd

import (
	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/contact"
	"agent-telegram/cmd/folders"
	"agent-telegram/cmd/game"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/gift"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/open"
	"agent-telegram/cmd/privacy"
	"agent-telegram/cmd/search"
	"agent-telegram/cmd/send"
	"agent-telegram/cmd/session"
	"agent-telegram/cmd/sys"
	"agent-telegram/cmd/user"
	"agent-telegram/internal/cliutil"
)

func init() {
	// Auth commands
	auth.AddLoginCommand(RootCmd)
	auth.AddLogoutCommand(RootCmd)
	get.AddMyInfoCommand(RootCmd)

	// Get commands
	get.AddUpdatesCommand(RootCmd)

	// Search commands
	search.AddSearchCommand(RootCmd)

	// Open command
	open.AddOpenCommand(RootCmd)

	// Message commands
	message.AddMsgCommand(RootCmd)

	// User commands
	user.AddUserCommand(RootCmd)

	// Contact commands
	contact.AddContactCommand(RootCmd)

	// Chat commands
	chat.AddChatCommand(RootCmd)

	// Folders commands
	folders.AddFoldersCommand(RootCmd)

	// Game commands
	game.AddGameCommand(RootCmd)

	// Gift commands
	gift.AddGiftCommand(RootCmd)

	// Privacy commands
	privacy.AddPrivacyCommand(RootCmd)

	// Session commands
	session.AddSessionCommand(RootCmd)

	// System commands
	sys.AddStatusCommand(RootCmd)
	sys.AddLLMsTxtCommand(RootCmd)

	// Register schema methods for commands not using helper constructors.
	// Commands using NewSimpleCommand/NewToggleCommand/NewListCommand auto-register.
	registerSchemaMethods()
}

func registerSchemaMethods() {
	r := cliutil.RegisterMethod

	// Root-level
	r(BalanceCmd, "get_balance")

	// Get
	r(get.MyInfoCmd, "get_me")
	r(get.UpdatesCmd, "get_updates")

	// Message
	r(message.ListCmd, "get_messages")
	r(message.DeleteCmd, "delete_message")
	r(message.ForwardCmd, "forward_message")
	r(message.PinMessageCmd, "pin_message")
	r(message.InspectButtonsCmd, "inspect_inline_buttons")
	r(message.PressButtonCmd, "press_inline_button")
	r(message.ReactionCmd, "add_reaction")
	r(message.InspectKeyboardCmd, "inspect_reply_keyboard")
	r(message.ReadCmd, "read_messages")
	r(message.TypingCmd, "set_typing")
	r(message.ScheduledCmd, "get_scheduled_messages")
	r(message.ClearCmd, "clear_messages")
	r(message.RepliesCmd, "get_replies")
	r(message.ReplyCommentCmd, "reply_to_comment")

	// Send
	r(send.SendCmd, "send_message")
	r(send.TextCmd, "send_message")
	r(send.PhotoCmd, "send_photo")
	r(send.VideoCmd, "send_video")
	r(send.VoiceCmd, "send_voice")
	r(send.StickerCmd, "send_sticker")
	r(send.ContactCmd, "send_contact")
	r(send.LocationCmd, "send_location")
	r(send.PollCmd, "send_poll")
	r(send.DiceCmd, "send_dice")
	r(send.UpdateCmd, "update_message")

	// Chat (non-helper commands)
	r(chat.ListCmd, "get_chats")
	r(chat.InfoCmd, "get_chats")
	r(chat.TopicsCmd, "get_topics")
	r(chat.KeyboardCmd, "inspect_reply_keyboard")
	r(chat.InviteLinkCmd, "get_invite_link")
	r(chat.PermissionsCmd, "set_chat_permissions")
	r(chat.OpenCmd, "get_messages")

	// User
	r(user.BlockCmd, "block")
	r(user.InfoCmd, "get_user_info")

	// Search
	r(search.SearchGlobalCmd, "search_global")
	r(search.SearchInChatCmd, "search_in_chat")

	// Contact
	r(contact.ListContactsCmd, "get_contacts")
	r(contact.AddContactCmd, "add_contact")
	r(contact.DeleteContactCmd, "delete_contact")

	// Folders (non-helper commands)
	r(folders.ListCmd, "get_folders")
	r(folders.CreateCmd, "create_folder")

	// Privacy
	r(privacy.GetCmd, "get_privacy")
	r(privacy.SetCmd, "set_privacy")

	// Gift (non-helper commands)
	r(gift.ListCmd, "get_saved_gifts")
	r(gift.InfoCmd, "get_gift_info")
	r(gift.SendCmd, "send_star_gift")
	r(gift.ConvertCmd, "convert_star_gift")
	r(gift.PriceCmd, "update_gift_price")
	r(gift.OfferCmd, "offer_gift")
	r(gift.AcceptCmd, "accept_gift_offer")
	r(gift.DeclineCmd, "decline_gift_offer")
	r(gift.AttrsCmd, "get_gift_attrs")
}
