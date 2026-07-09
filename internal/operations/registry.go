package operations

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"agent-telegram/internal/strictjson"
	"agent-telegram/telegram/types"
)

const (
	SafetyRead        = "read"
	SafetyWrite       = "write"
	SafetyDestructive = "destructive"
	SafetyPaid        = "paid"
)

// NoParams is used for methods that accept an empty params object.
type NoParams struct{}

// Validate implements the IPC params contract.
func (NoParams) Validate() error { return nil }

// Control operation contracts are shared by the local IPC and HTTP surfaces.
type PingParams struct {
	Message string `json:"message,omitempty"`
}

type PingResult struct {
	Message string `json:"message"`
	Pong    bool   `json:"pong"`
}

type ReloadSessionParams struct {
	Session string `json:"session" validate:"required"`
}

type ControlResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Operation describes one IPC/HTTP operation for agents and documentation.
type Operation struct {
	Method               string
	Summary              string
	Category             string
	Safety               string
	Idempotent           bool
	Retryable            bool
	RequiresConfirmation bool
	ParamsType           reflect.Type
	ResultType           reflect.Type
	Examples             []map[string]any
}

// ManifestOperation is the JSON-safe representation returned to agents.
type ManifestOperation struct {
	Method               string           `json:"method"`
	Summary              string           `json:"summary"`
	Category             string           `json:"category"`
	Safety               string           `json:"safety"`
	Idempotent           bool             `json:"idempotent"`
	Retryable            bool             `json:"retryable"`
	RequiresConfirmation bool             `json:"requiresConfirmation"`
	InputSchema          JSONSchema       `json:"inputSchema"`
	OutputSchema         JSONSchema       `json:"outputSchema"`
	Examples             []map[string]any `json:"examples,omitempty"`
}

var registry = map[string]Operation{}

func init() {
	registerCore()
	registerMessages()
	registerMedia()
	registerChats()
	registerUsers()
	registerSearch()
	registerGifts()
}

// Register adds an operation to the registry.
func Register(op Operation) {
	if op.Method == "" {
		panic("operation method is required")
	}
	if op.Safety == "" {
		op.Safety = SafetyRead
	}
	if op.ParamsType == nil {
		op.ParamsType = reflect.TypeOf(NoParams{})
	}
	if op.ResultType == nil {
		op.ResultType = reflect.TypeOf(map[string]any{})
	}
	registry[op.Method] = op
}

// Get returns an operation by method.
func Get(method string) (Operation, bool) {
	op, ok := registry[method]
	return op, ok
}

// Methods returns all registered operation methods sorted alphabetically.
func Methods() []string {
	return slices.Sorted(maps.Keys(registry))
}

// Manifest returns all operations in a machine-readable format.
func Manifest() []ManifestOperation {
	methods := Methods()
	out := make([]ManifestOperation, 0, len(methods))
	for _, method := range methods {
		op := registry[method]
		out = append(out, ManifestFor(op))
	}
	return out
}

// ManifestFor returns the JSON-safe representation of one operation.
func ManifestFor(op Operation) ManifestOperation {
	return ManifestOperation{
		Method:               op.Method,
		Summary:              op.Summary,
		Category:             op.Category,
		Safety:               op.Safety,
		Idempotent:           op.Idempotent,
		Retryable:            op.Retryable,
		RequiresConfirmation: op.RequiresConfirmation,
		InputSchema:          InputSchemaFor(op.ParamsType),
		OutputSchema:         SchemaFor(op.ResultType),
		Examples:             op.Examples,
	}
}

// HasSchema reports whether a method is present in the operation registry.
func HasSchema(method string) bool {
	_, ok := registry[method]
	return ok
}

// ValidateParams validates raw params using the registered input schema type.
func ValidateParams(method string, raw json.RawMessage) error {
	op, ok := registry[method]
	if !ok {
		return fmt.Errorf("unknown method %q", method)
	}
	params := reflect.New(op.ParamsType).Interface()
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if err := strictjson.Decode(raw, params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	if err := types.ValidateStruct(params); err != nil {
		return err
	}
	if err := validatePeerInfo(params); err != nil {
		return err
	}
	if validator, ok := params.(interface{ Validate() error }); ok {
		return validator.Validate()
	}
	return nil
}

func validatePeerInfo(params any) error {
	if optional, ok := params.(interface{ AllowEmptyPeer() bool }); ok && optional.AllowEmptyPeer() {
		return nil
	}
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	peer := visibleStringField(v, "Peer")
	username := visibleStringField(v, "Username")
	if peer == nil || username == nil {
		return nil
	}
	if *peer == "" && *username == "" {
		return fmt.Errorf("peer or username is required")
	}
	return nil
}

func visibleStringField(v reflect.Value, name string) *string {
	field := v.FieldByName(name)
	if !field.IsValid() || field.Kind() != reflect.String {
		return nil
	}
	value := field.String()
	return &value
}

func op(
	method, summary, category, safety string,
	params any,
	result any,
	idempotent, retryable, confirm bool,
	examples ...map[string]any,
) Operation {
	paramsType := reflect.TypeOf(params)
	if paramsType.Kind() == reflect.Ptr {
		paramsType = paramsType.Elem()
	}
	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr {
		resultType = resultType.Elem()
	}
	return Operation{
		Method:               method,
		Summary:              summary,
		Category:             category,
		Safety:               safety,
		ParamsType:           paramsType,
		ResultType:           resultType,
		Idempotent:           idempotent,
		Retryable:            retryable,
		RequiresConfirmation: confirm,
		Examples:             examples,
	}
}

func read(method, summary, category string, params any, result any, examples ...map[string]any) {
	Register(op(method, summary, category, SafetyRead, params, result, true, true, false, examples...))
}

func write(method, summary, category string, params any, result any, examples ...map[string]any) {
	Register(op(method, summary, category, SafetyWrite, params, result, false, false, false, examples...))
}

func confirmedWrite(method, summary, category string, params any, result any, examples ...map[string]any) {
	Register(op(method, summary, category, SafetyWrite, params, result, false, false, true, examples...))
}

func destructive(method, summary, category string, params any, result any, examples ...map[string]any) {
	Register(op(method, summary, category, SafetyDestructive, params, result, false, false, true, examples...))
}

func paid(method, summary, category string, params any, result any, examples ...map[string]any) {
	Register(op(method, summary, category, SafetyPaid, params, result, false, false, true, examples...))
}

func registerCore() {
	read("ping", "Check local RPC availability", "system", PingParams{}, PingResult{})
	read("echo", "Echo a local RPC message", "system", PingParams{}, map[string]any{})
	read("status", "Get local daemon and Telegram status", "system", NoParams{}, map[string]any{})
	write("shutdown", "Stop the local daemon", "system", NoParams{}, ControlResult{})
	confirmedWrite("logout", "Invalidate the Telegram session and stop the daemon", "system", NoParams{}, ControlResult{})
	write("reload_session", "Replace the in-memory Telegram session", "system", ReloadSessionParams{}, ControlResult{})
	read("get_me", "Get the authorized account profile", "account", NoParams{}, types.GetMeResult{})
	read("get_updates", "Get pending Telegram updates", "updates", types.GetUpdatesParams{}, types.GetUpdatesResult{})
	read("get_balance", "Get Stars and TON balance", "gifts", types.GetBalanceParams{}, types.GetBalanceResult{})
}

func registerMessages() {
	read("get_message", "Get a single message by ID", "messages", types.GetMessageParams{}, types.GetMessageResult{})
	read("get_messages", "List messages from a chat", "messages", types.GetMessagesParams{}, types.GetMessagesResult{})
	write("send_message", "Send a text message", "messages", types.SendMessageParams{}, types.SendMessageResult{}, map[string]any{"peer": "@username", "message": "Hello"})
	write("send_reply", "Send a reply to a message", "messages", types.SendReplyParams{}, types.SendReplyResult{})
	write("update_message", "Edit a previously sent message", "messages", types.UpdateMessageParams{}, types.UpdateMessageResult{})
	destructive("delete_message", "Delete one or more messages", "messages", types.DeleteMessageParams{}, types.DeleteMessageResult{})
	write("forward_message", "Forward a message to another peer", "messages", types.ForwardMessageParams{}, types.ForwardMessageResult{})
	destructive("clear_messages", "Clear selected messages", "messages", types.ClearMessagesParams{}, types.ClearMessagesResult{})
	destructive("clear_history", "Clear chat history", "messages", types.ClearHistoryParams{}, types.ClearHistoryResult{})
	read("inspect_inline_buttons", "Inspect inline buttons on a message", "buttons", types.InspectInlineButtonsParams{}, types.InspectInlineButtonsResult{})
	write("press_inline_button", "Press an inline button", "buttons", types.PressInlineButtonParams{}, types.PressInlineButtonResult{})
	read("inspect_reply_keyboard", "Inspect reply keyboard buttons", "buttons", types.PeerInfo{}, types.ReplyKeyboardResult{})
	write("pin_message", "Pin a message", "messages", types.PinMessageParams{}, types.PinMessageResult{})
	write("unpin_message", "Unpin a message", "messages", types.UnpinMessageParams{}, types.UnpinMessageResult{})
	write("add_reaction", "Add a reaction to a message", "reactions", types.AddReactionParams{}, types.AddReactionResult{})
	write("remove_reaction", "Remove a reaction from a message", "reactions", types.RemoveReactionParams{}, types.RemoveReactionResult{})
	read("list_reactions", "List message reactions", "reactions", types.ListReactionsParams{}, types.ListReactionsResult{})
	write("read_messages", "Mark messages as read", "messages", types.ReadMessagesParams{}, types.ReadMessagesResult{})
	write("set_typing", "Send typing indicator", "messages", types.SetTypingParams{}, types.SetTypingResult{})
	read("get_scheduled_messages", "List scheduled messages", "messages", types.GetScheduledMessagesParams{}, types.GetScheduledMessagesResult{})
	read("get_replies", "Get replies/comments for a channel post", "messages", types.GetRepliesParams{}, types.GetRepliesResult{})
	write("reply_to_comment", "Reply to a channel post comment", "messages", types.ReplyToCommentParams{}, types.ReplyToCommentResult{})
}

func registerMedia() {
	write("send_photo", "Send a photo", "media", types.SendPhotoParams{}, types.SendPhotoResult{})
	write("send_video", "Send a video", "media", types.SendVideoParams{}, types.SendVideoResult{})
	write("send_file", "Send a generic file", "media", types.SendFileParams{}, types.SendFileResult{})
	write("send_document", "Send a document", "media", types.SendFileParams{}, types.SendFileResult{})
	write("send_audio", "Send an audio file", "media", types.SendFileParams{}, types.SendFileResult{})
	write("send_location", "Send a location", "media", types.SendLocationParams{}, types.SendLocationResult{})
	write("send_contact", "Send a contact", "media", types.SendContactParams{}, types.SendContactResult{})
	write("send_poll", "Send a poll", "media", types.SendPollParams{}, types.SendPollResult{})
	write("send_checklist", "Send a checklist/poll-style item", "media", types.SendPollParams{}, types.SendPollResult{})
	write("send_voice", "Send a voice message", "media", types.SendVoiceParams{}, types.SendVoiceResult{})
	write("send_video_note", "Send a round video note", "media", types.SendVideoNoteParams{}, types.SendVideoNoteResult{})
	write("send_sticker", "Send a sticker", "media", types.SendStickerParams{}, types.SendStickerResult{})
	read("get_sticker_packs", "List sticker packs", "media", types.GetStickerPacksParams{}, types.GetStickerPacksResult{})
	write("send_gif", "Send a GIF or animation", "media", types.SendGIFParams{}, types.SendGIFResult{})
	write("send_dice", "Send a Telegram dice/game roll", "media", types.SendDiceParams{}, types.SendDiceResult{})
}

func registerChats() {
	read("get_chats", "List chats and dialogs", "chats", &types.GetChatsParams{}, types.GetChatsResult{})
	write("pin_chat", "Pin a chat in the dialog list", "chats", types.PinChatParams{}, types.PinChatResult{})
	write("archive", "Archive a chat", "chats", types.ArchiveParams{}, types.ArchiveResult{})
	write("unarchive", "Unarchive a chat", "chats", types.UnarchiveParams{}, types.UnarchiveResult{})
	write("mute", "Mute chat notifications", "chats", types.MuteParams{}, types.MuteResult{})
	write("unmute", "Unmute chat notifications", "chats", types.UnmuteParams{}, types.UnmuteResult{})
	write("join_chat", "Join a chat via invite link", "chats", types.JoinChatParams{}, types.JoinChatResult{})
	write("subscribe_channel", "Subscribe to a public channel", "chats", types.SubscribeChannelParams{}, types.SubscribeChannelResult{})
	destructive("leave", "Leave a chat or channel", "chats", types.LeaveParams{}, types.LeaveResult{})
	write("invite", "Invite users to a chat", "chats", types.InviteParams{}, types.InviteResult{})
	read("get_topics", "List forum topics", "chats", types.GetTopicsParams{}, types.GetTopicsResult{})
	write("create_group", "Create a group chat", "chats", types.CreateGroupParams{}, types.CreateGroupResult{})
	write("create_channel", "Create a channel or supergroup", "chats", types.CreateChannelParams{}, types.CreateChannelResult{})
	write("edit_title", "Edit chat title", "chats", types.EditTitleParams{}, types.EditTitleResult{})
	write("set_photo", "Set chat photo", "chats", types.SetPhotoParams{}, types.SetPhotoResult{})
	destructive("delete_photo", "Delete chat photo", "chats", types.DeletePhotoParams{}, types.DeletePhotoResult{})
	read("get_participants", "List chat participants", "chats", types.GetParticipantsParams{}, types.GetParticipantsResult{})
	read("get_admins", "List chat admins", "chats", types.GetAdminsParams{}, types.GetAdminsResult{})
	read("get_banned", "List banned users", "chats", types.GetBannedParams{}, types.GetBannedResult{})
	write("promote_admin", "Promote a chat admin", "chats", types.PromoteAdminParams{}, types.PromoteAdminResult{})
	write("demote_admin", "Demote a chat admin", "chats", types.DemoteAdminParams{}, types.DemoteAdminResult{})
	write("get_invite_link", "Get or create an invite link", "chats", types.GetInviteLinkParams{}, types.GetInviteLinkResult{})
	write("set_slow_mode", "Set chat slow mode", "chats", types.SetSlowModeParams{}, types.SetSlowModeResult{})
	write("set_chat_permissions", "Set default chat permissions", "chats", types.SetChatPermissionsParams{}, types.SetChatPermissionsResult{})
	read("get_folders", "List chat folders", "folders", types.GetFoldersParams{}, types.GetFoldersResult{})
	write("create_folder", "Create a chat folder", "folders", types.CreateFolderParams{}, types.CreateFolderResult{})
	destructive("delete_folder", "Delete a chat folder", "folders", types.DeleteFolderParams{}, types.DeleteFolderResult{})
}

func registerUsers() {
	write("update_profile", "Update account profile", "users", types.UpdateProfileParams{}, types.UpdateProfileResult{})
	write("update_avatar", "Update account avatar", "users", types.UpdateAvatarParams{}, types.UpdateAvatarResult{})
	destructive("block", "Block a peer", "users", types.BlockPeerParams{}, types.BlockPeerResult{})
	write("unblock", "Unblock a peer", "users", types.UnblockPeerParams{}, types.UnblockPeerResult{})
	read("get_user_info", "Get user info", "users", types.GetUserInfoParams{}, types.GetUserInfoResult{})
	read("get_contacts", "List contacts", "contacts", types.GetContactsParams{}, types.GetContactsResult{})
	write("add_contact", "Add a contact", "contacts", types.AddContactParams{}, types.AddContactResult{})
	destructive("delete_contact", "Delete a contact", "contacts", types.DeleteContactParams{}, types.DeleteContactResult{})
	read("get_privacy", "Get privacy settings", "privacy", types.GetPrivacyParams{}, types.GetPrivacyResult{})
	write("set_privacy", "Set privacy settings", "privacy", types.SetPrivacyParams{}, types.SetPrivacyResult{})
}

func registerSearch() {
	read("search_global", "Search public Telegram content", "search", types.SearchGlobalParams{}, types.SearchGlobalResult{})
	read("search_in_chat", "Search messages within a chat", "search", types.SearchInChatParams{}, types.SearchInChatResult{})
}

func registerGifts() {
	read("get_star_gifts", "List star gift catalog", "gifts", types.GetStarGiftsParams{}, types.GetStarGiftsResult{})
	paid("send_star_gift", "Send or buy a star gift", "gifts", types.SendStarGiftParams{}, types.SendStarGiftResult{})
	read("get_saved_gifts", "List saved gifts", "gifts", types.GetSavedGiftsParams{}, types.GetSavedGiftsResult{})
	paid("transfer_star_gift", "Transfer a star gift", "gifts", types.TransferStarGiftParams{}, types.TransferStarGiftResult{})
	destructive("convert_star_gift", "Convert a star gift to stars", "gifts", types.ConvertStarGiftParams{}, types.ConvertStarGiftResult{})
	paid("update_gift_price", "Update star gift resale price", "gifts", types.UpdateGiftPriceParams{}, types.UpdateGiftPriceResult{})
	paid("offer_gift", "Make an offer to buy a gift", "gifts", types.OfferGiftParams{}, types.OfferGiftResult{})
	read("get_gift_info", "Get gift info and ownership details", "gifts", types.GetGiftInfoParams{}, types.GetGiftInfoResult{})
	read("get_gift_value", "Get gift value analytics", "gifts", types.GetGiftValueParams{}, types.GetGiftValueResult{})
	read("get_resale_gifts", "List resale gifts", "gifts", types.GetResaleGiftsParams{}, types.GetResaleGiftsResult{})
	paid("buy_resale_gift", "Buy a resale gift", "gifts", types.BuyResaleGiftParams{}, types.BuyResaleGiftResult{})
	read("get_gift_attrs", "List gift attributes", "gifts", types.GetGiftAttrsParams{}, types.GetGiftAttrsResult{})
	paid("accept_gift_offer", "Accept an incoming gift offer", "gifts", types.AcceptGiftOfferParams{}, types.AcceptGiftOfferResult{})
	destructive("decline_gift_offer", "Decline an incoming gift offer", "gifts", types.DeclineGiftOfferParams{}, types.DeclineGiftOfferResult{})
}
