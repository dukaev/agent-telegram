// Package ipc provides Telegram IPC handlers registration.
package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"agent-telegram/internal/ipc"
	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

// methodHandlers maps method names to handler factory functions.
//
//nolint:funlen // Declarative handler registry — all entries are one-liners.
var methodHandlers = map[string]func(Client) HandlerFunc{
	// Basic
	"get_me":        GetMeHandler,
	"get_updates":   GetUpdatesHandler,
	"get_chats":     func(c Client) HandlerFunc { return Handler(c.Chat().GetChats, "get chats") },
	"get_message":   func(c Client) HandlerFunc { return Handler(c.Message().GetMessage, "get message") },
	"get_messages":  func(c Client) HandlerFunc { return Handler(c.Message().GetMessages, "get messages") },
	"get_user_info": func(c Client) HandlerFunc { return Handler(c.User().GetUserInfo, "get user info") },

	// Messages
	"send_message":    func(c Client) HandlerFunc { return Handler(c.Message().SendMessage, "send message") },
	"send_reply":      func(c Client) HandlerFunc { return Handler(c.Message().SendReply, "send reply") },
	"update_message":  func(c Client) HandlerFunc { return Handler(c.Message().UpdateMessage, "update message") },
	"delete_message":  func(c Client) HandlerFunc { return Handler(c.Message().DeleteMessage, "delete message") },
	"forward_message": func(c Client) HandlerFunc { return Handler(c.Message().ForwardMessage, "forward message") },
	"clear_messages":  func(c Client) HandlerFunc { return Handler(c.Chat().ClearMessages, "clear messages") },
	"clear_history":   func(c Client) HandlerFunc { return Handler(c.Chat().ClearHistory, "clear history") },

	// Media
	"send_photo": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendPhotoParams) string { return p.File }, c.Media().SendPhoto, "send photo")
	},
	"send_video": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendVideoParams) string { return p.File }, c.Media().SendVideo, "send video")
	},
	"send_file": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendFileParams) string { return p.File }, c.Media().SendFile, "send file")
	},
	"send_document": func(c Client) HandlerFunc { // alias for send_file
		return FileHandler(func(p types.SendFileParams) string { return p.File }, c.Media().SendFile, "send file")
	},
	"send_audio": func(c Client) HandlerFunc { // alias for send_file
		return FileHandler(func(p types.SendFileParams) string { return p.File }, c.Media().SendFile, "send file")
	},
	"send_location":  func(c Client) HandlerFunc { return Handler(c.Media().SendLocation, "send location") },
	"send_contact":   func(c Client) HandlerFunc { return Handler(c.Media().SendContact, "send contact") },
	"send_poll":      SendPollHandler,
	"send_checklist": SendChecklistHandler,
	"send_voice": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendVoiceParams) string { return p.File }, c.Media().SendVoice, "send voice")
	},
	"send_video_note": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendVideoNoteParams) string { return p.File }, c.Media().SendVideoNote, "send video note")
	},
	"send_sticker":      func(c Client) HandlerFunc { return Handler(c.Media().SendSticker, "send sticker") },
	"get_sticker_packs": func(c Client) HandlerFunc { return Handler(c.Media().GetStickerPacks, "get sticker packs") },
	"send_gif": func(c Client) HandlerFunc {
		return FileHandler(func(p types.SendGIFParams) string { return p.File }, c.Media().SendGIF, "send gif")
	},
	"send_dice": func(c Client) HandlerFunc { return Handler(c.Media().SendDice, "send dice") },

	// Inline/keyboard
	"inspect_inline_buttons": func(c Client) HandlerFunc {
		return Handler(c.Message().InspectInlineButtons, "inspect inline buttons")
	},
	"press_inline_button": func(c Client) HandlerFunc {
		return Handler(c.Message().PressInlineButton, "press inline button")
	},
	"inspect_reply_keyboard": func(c Client) HandlerFunc {
		return Handler(c.Message().InspectReplyKeyboard, "inspect reply keyboard")
	},
	"pin_message":   func(c Client) HandlerFunc { return Handler(c.Pin().PinMessage, "pin message") },
	"unpin_message": func(c Client) HandlerFunc { return Handler(c.Pin().UnpinMessage, "unpin message") },

	// Reactions
	"add_reaction":    func(c Client) HandlerFunc { return Handler(c.Reaction().AddReaction, "add reaction") },
	"remove_reaction": func(c Client) HandlerFunc { return Handler(c.Reaction().RemoveReaction, "remove reaction") },
	"list_reactions":  func(c Client) HandlerFunc { return Handler(c.Reaction().ListReactions, "list reactions") },

	// Chat operations
	"pin_chat":          func(c Client) HandlerFunc { return Handler(c.Chat().PinChat, "pin chat") },
	"archive":           func(c Client) HandlerFunc { return Handler(c.Chat().Archive, "archive chat") },
	"unarchive":         func(c Client) HandlerFunc { return Handler(c.Chat().Unarchive, "unarchive chat") },
	"mute":              func(c Client) HandlerFunc { return Handler(c.Chat().Mute, "mute chat") },
	"unmute":            func(c Client) HandlerFunc { return Handler(c.Chat().Unmute, "unmute chat") },
	"join_chat":         func(c Client) HandlerFunc { return Handler(c.Chat().JoinChat, "join chat") },
	"subscribe_channel": func(c Client) HandlerFunc { return Handler(c.Chat().SubscribeChannel, "subscribe channel") },
	"leave":             func(c Client) HandlerFunc { return Handler(c.Chat().Leave, "leave chat") },
	"invite":            func(c Client) HandlerFunc { return Handler(c.Chat().Invite, "invite users") },
	"get_topics":        func(c Client) HandlerFunc { return Handler(c.Chat().GetTopics, "get topics") },
	"create_group":      func(c Client) HandlerFunc { return Handler(c.Chat().CreateGroup, "create group") },
	"create_channel":    func(c Client) HandlerFunc { return Handler(c.Chat().CreateChannel, "create channel") },
	"edit_title":        func(c Client) HandlerFunc { return Handler(c.Chat().EditTitle, "edit title") },
	"set_photo":         func(c Client) HandlerFunc { return Handler(c.Chat().SetPhoto, "set photo") },
	"delete_photo":      func(c Client) HandlerFunc { return Handler(c.Chat().DeletePhoto, "delete photo") },
	"get_participants":  func(c Client) HandlerFunc { return Handler(c.Chat().GetParticipants, "get participants") },
	"get_admins":        func(c Client) HandlerFunc { return Handler(c.Chat().GetAdmins, "get admins") },
	"get_banned":        func(c Client) HandlerFunc { return Handler(c.Chat().GetBanned, "get banned users") },
	"promote_admin":     func(c Client) HandlerFunc { return Handler(c.Chat().PromoteAdmin, "promote admin") },
	"demote_admin":      func(c Client) HandlerFunc { return Handler(c.Chat().DemoteAdmin, "demote admin") },
	"get_invite_link":   func(c Client) HandlerFunc { return Handler(c.Chat().GetInviteLink, "get invite link") },
	"set_slow_mode":     func(c Client) HandlerFunc { return Handler(c.Chat().SetSlowMode, "set slow mode") },
	"set_chat_permissions": func(c Client) HandlerFunc {
		return Handler(c.Chat().SetChatPermissions, "set chat permissions")
	},
	"get_folders":   func(c Client) HandlerFunc { return Handler(c.Chat().GetFolders, "get folders") },
	"create_folder": func(c Client) HandlerFunc { return Handler(c.Chat().CreateFolder, "create folder") },
	"delete_folder": func(c Client) HandlerFunc { return Handler(c.Chat().DeleteFolder, "delete folder") },

	// User operations
	"update_profile": func(c Client) HandlerFunc { return Handler(c.User().UpdateProfile, "update profile") },
	"update_avatar": func(c Client) HandlerFunc {
		return FileHandler(func(p types.UpdateAvatarParams) string { return p.File }, c.User().UpdateAvatar, "update avatar")
	},
	"block":          func(c Client) HandlerFunc { return Handler(c.User().BlockPeer, "block peer") },
	"unblock":        func(c Client) HandlerFunc { return Handler(c.User().UnblockPeer, "unblock peer") },
	"get_contacts":   func(c Client) HandlerFunc { return Handler(c.User().GetContacts, "get contacts") },
	"add_contact":    func(c Client) HandlerFunc { return Handler(c.User().AddContact, "add contact") },
	"delete_contact": func(c Client) HandlerFunc { return Handler(c.User().DeleteContact, "delete contact") },
	"get_privacy":    func(c Client) HandlerFunc { return Handler(c.User().GetPrivacy, "get privacy") },
	"set_privacy":    func(c Client) HandlerFunc { return Handler(c.User().SetPrivacy, "set privacy") },

	// Search
	"search_global":  func(c Client) HandlerFunc { return Handler(c.Search().SearchGlobal, "search global") },
	"search_in_chat": func(c Client) HandlerFunc { return Handler(c.Search().SearchInChat, "search in chat") },

	// Message features
	"read_messages": func(c Client) HandlerFunc { return Handler(c.Message().ReadMessages, "read messages") },
	"set_typing":    func(c Client) HandlerFunc { return Handler(c.Message().SetTyping, "set typing") },
	"get_scheduled_messages": func(c Client) HandlerFunc {
		return Handler(c.Message().GetScheduledMessages, "get scheduled messages")
	},
	"get_replies":      func(c Client) HandlerFunc { return Handler(c.Message().GetReplies, "get replies") },
	"reply_to_comment": func(c Client) HandlerFunc { return Handler(c.Message().ReplyToComment, "reply to comment") },

	// Gift operations
	"get_star_gifts":     func(c Client) HandlerFunc { return Handler(c.Gift().GetStarGifts, "get star gifts") },
	"send_star_gift":     func(c Client) HandlerFunc { return Handler(c.Gift().SendStarGift, "send star gift") },
	"get_saved_gifts":    func(c Client) HandlerFunc { return Handler(c.Gift().GetSavedGifts, "get saved gifts") },
	"transfer_star_gift": func(c Client) HandlerFunc { return Handler(c.Gift().TransferStarGift, "transfer star gift") },
	"convert_star_gift":  func(c Client) HandlerFunc { return Handler(c.Gift().ConvertStarGift, "convert star gift") },
	"update_gift_price":  func(c Client) HandlerFunc { return Handler(c.Gift().UpdateGiftPrice, "update gift price") },
	"get_balance":        func(c Client) HandlerFunc { return Handler(c.Gift().GetBalance, "get balance") },
	"offer_gift":         func(c Client) HandlerFunc { return Handler(c.Gift().OfferGift, "offer gift") },
	"get_gift_info":      func(c Client) HandlerFunc { return Handler(c.Gift().GetGiftInfo, "get gift info") },
	"get_gift_value":     func(c Client) HandlerFunc { return Handler(c.Gift().GetGiftValue, "get gift value") },
	"get_resale_gifts":   func(c Client) HandlerFunc { return Handler(c.Gift().GetResaleGifts, "get resale gifts") },
	"buy_resale_gift":    func(c Client) HandlerFunc { return Handler(c.Gift().BuyResaleGift, "buy resale gift") },
	"get_gift_attrs":     func(c Client) HandlerFunc { return Handler(c.Gift().GetGiftAttrs, "get gift attrs") },
	"accept_gift_offer":  func(c Client) HandlerFunc { return Handler(c.Gift().AcceptGiftOffer, "accept gift offer") },
	"decline_gift_offer": func(c Client) HandlerFunc { return Handler(c.Gift().DeclineGiftOffer, "decline gift offer") },
}

// RegisterHandlers registers all Telegram IPC handlers.
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
	for method, factory := range methodHandlers {
		registerHandler(srv, method, factory(client))
	}
}

// RegisteredMethods returns the Telegram IPC method names exposed by this package.
func RegisteredMethods() []string {
	methods := make([]string, 0, len(methodHandlers))
	for method := range methodHandlers {
		methods = append(methods, method)
	}
	return methods
}

// registerHandler registers a single handler with error wrapping and request timeout.
func registerHandler(srv ipc.MethodRegistrar, method string, handler HandlerFunc) {
	srv.Register(method, func(params json.RawMessage) (result interface{}, rpcErr *ipc.ErrorObject) {
		ctx, cancel := context.WithTimeout(context.Background(), ipc.RequestTimeout())
		defer cancel()

		res, err := handler(ctx, params)
		if err != nil {
			return nil, classifyRPCError(err)
		}
		return res, nil
	})
}

func classifyRPCError(err error) *ipc.ErrorObject {
	if errors.Is(err, client.ErrNotInitialized) {
		return ipc.ErrNotInitialized
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ipc.NewTypedError(ipc.ErrCodeTimeout, ipc.ErrorTypeTimeout, "request timed out", nil)
	}

	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "flood_wait") || strings.Contains(lower, "flood wait"):
		data := map[string]any{}
		if retryAfter := parseRetryAfter(lower); retryAfter > 0 {
			data["retryAfter"] = retryAfter
		}
		return ipc.NewTypedError(ipc.ErrCodeFloodWait, ipc.ErrorTypeFloodWait, msg, data)
	case strings.Contains(lower, "peer") &&
		(strings.Contains(lower, "not found") || strings.Contains(lower, "invalid")):
		return ipc.NewTypedError(ipc.ErrCodePeerNotFound, ipc.ErrorTypePeerNotFound, msg, nil)
	case strings.Contains(lower, "username_invalid") ||
		strings.Contains(lower, "channel_invalid") ||
		strings.Contains(lower, "user_id_invalid"):
		return ipc.NewTypedError(ipc.ErrCodePeerNotFound, ipc.ErrorTypePeerNotFound, msg, nil)
	case strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "chat_admin_required") ||
		strings.Contains(lower, "not enough rights"):
		return ipc.NewTypedError(ipc.ErrCodeForbidden, ipc.ErrorTypeForbidden, msg, nil)
	case strings.Contains(lower, "invalid params") ||
		strings.Contains(lower, "validation") ||
		strings.Contains(lower, "required"):
		return ipc.NewTypedError(ipc.ErrCodeInvalidParams, ipc.ErrorTypeValidation, msg, nil)
	default:
		return ipc.NewTypedError(-32000, ipc.ErrorTypeInternal, msg, nil)
	}
}

func parseRetryAfter(msg string) int {
	for _, marker := range []string{"flood_wait_", "flood wait "} {
		idx := strings.Index(msg, marker)
		if idx == -1 {
			continue
		}
		rest := msg[idx+len(marker):]
		digits := strings.Builder{}
		for _, r := range rest {
			if r < '0' || r > '9' {
				break
			}
			digits.WriteRune(r)
		}
		if digits.Len() == 0 {
			continue
		}
		value, err := strconv.Atoi(digits.String())
		if err == nil {
			return value
		}
	}
	return 0
}
