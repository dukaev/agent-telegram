// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/telegram/types"
)

// resultTypes maps IPC method names to their result types for schema generation.
var resultTypes = map[string]reflect.Type{
	// Basic
	"get_me":        reflect.TypeOf(types.GetMeResult{}),
	"get_chats":     reflect.TypeOf(types.GetChatsResult{}),
	"get_updates":   reflect.TypeOf(types.GetUpdatesResult{}),
	"get_message":   reflect.TypeOf(types.GetMessageResult{}),
	"get_messages":  reflect.TypeOf(types.GetMessagesResult{}),
	"get_user_info": reflect.TypeOf(types.GetUserInfoResult{}),

	// Messages
	"send_message":    reflect.TypeOf(types.SendMessageResult{}),
	"send_reply":      reflect.TypeOf(types.SendReplyResult{}),
	"send_location":   reflect.TypeOf(types.SendLocationResult{}),
	"send_photo":      reflect.TypeOf(types.SendPhotoResult{}),
	"send_contact":    reflect.TypeOf(types.SendContactResult{}),
	"send_file":       reflect.TypeOf(types.SendFileResult{}),
	"send_document":   reflect.TypeOf(types.SendFileResult{}),
	"send_audio":      reflect.TypeOf(types.SendFileResult{}),
	"send_poll":       reflect.TypeOf(types.SendPollResult{}),
	"send_checklist":  reflect.TypeOf(types.SendPollResult{}),
	"send_video":      reflect.TypeOf(types.SendVideoResult{}),
	"update_message":  reflect.TypeOf(types.UpdateMessageResult{}),
	"delete_message":  reflect.TypeOf(types.DeleteMessageResult{}),
	"forward_message": reflect.TypeOf(types.ForwardMessageResult{}),
	"clear_messages":  reflect.TypeOf(types.ClearMessagesResult{}),
	"clear_history":   reflect.TypeOf(types.ClearHistoryResult{}),

	// Inline/keyboard
	"inspect_inline_buttons": reflect.TypeOf(types.InspectInlineButtonsResult{}),
	"press_inline_button":    reflect.TypeOf(types.PressInlineButtonResult{}),
	"inspect_reply_keyboard": reflect.TypeOf(types.ReplyKeyboardResult{}),
	"pin_message":            reflect.TypeOf(types.PinMessageResult{}),
	"unpin_message":          reflect.TypeOf(types.UnpinMessageResult{}),

	// Reactions
	"add_reaction":    reflect.TypeOf(types.AddReactionResult{}),
	"remove_reaction": reflect.TypeOf(types.RemoveReactionResult{}),
	"list_reactions":  reflect.TypeOf(types.ListReactionsResult{}),

	// Chat operations
	"pin_chat":          reflect.TypeOf(types.PinChatResult{}),
	"archive":           reflect.TypeOf(types.ArchiveResult{}),
	"unarchive":         reflect.TypeOf(types.UnarchiveResult{}),
	"mute":              reflect.TypeOf(types.MuteResult{}),
	"unmute":            reflect.TypeOf(types.UnmuteResult{}),
	"join_chat":         reflect.TypeOf(types.JoinChatResult{}),
	"subscribe_channel": reflect.TypeOf(types.SubscribeChannelResult{}),
	"leave":             reflect.TypeOf(types.LeaveResult{}),
	"invite":            reflect.TypeOf(types.InviteResult{}),
	"get_topics":        reflect.TypeOf(types.GetTopicsResult{}),
	"create_group":      reflect.TypeOf(types.CreateGroupResult{}),
	"create_channel":    reflect.TypeOf(types.CreateChannelResult{}),
	"edit_title":        reflect.TypeOf(types.EditTitleResult{}),
	"set_photo":         reflect.TypeOf(types.SetPhotoResult{}),
	"delete_photo":      reflect.TypeOf(types.DeletePhotoResult{}),
	"get_participants":  reflect.TypeOf(types.GetParticipantsResult{}),
	"get_admins":        reflect.TypeOf(types.GetAdminsResult{}),
	"get_banned":        reflect.TypeOf(types.GetBannedResult{}),
	"promote_admin":     reflect.TypeOf(types.PromoteAdminResult{}),
	"demote_admin":      reflect.TypeOf(types.DemoteAdminResult{}),
	"get_invite_link":   reflect.TypeOf(types.GetInviteLinkResult{}),

	// User operations
	"update_profile": reflect.TypeOf(types.UpdateProfileResult{}),
	"update_avatar":  reflect.TypeOf(types.UpdateAvatarResult{}),
	"block":          reflect.TypeOf(types.BlockPeerResult{}),
	"unblock":        reflect.TypeOf(types.UnblockPeerResult{}),
	"search_global":  reflect.TypeOf(types.SearchGlobalResult{}),
	"search_in_chat": reflect.TypeOf(types.SearchInChatResult{}),
	"get_contacts":   reflect.TypeOf(types.GetContactsResult{}),
	"add_contact":    reflect.TypeOf(types.AddContactResult{}),
	"delete_contact": reflect.TypeOf(types.DeleteContactResult{}),

	// Message features
	"read_messages":          reflect.TypeOf(types.ReadMessagesResult{}),
	"set_typing":             reflect.TypeOf(types.SetTypingResult{}),
	"get_scheduled_messages": reflect.TypeOf(types.GetScheduledMessagesResult{}),
	"get_replies":            reflect.TypeOf(types.GetRepliesResult{}),
	"reply_to_comment":       reflect.TypeOf(types.ReplyToCommentResult{}),
	"send_voice":             reflect.TypeOf(types.SendVoiceResult{}),
	"send_video_note":        reflect.TypeOf(types.SendVideoNoteResult{}),
	"send_sticker":           reflect.TypeOf(types.SendStickerResult{}),
	"get_sticker_packs":      reflect.TypeOf(types.GetStickerPacksResult{}),
	"send_gif":               reflect.TypeOf(types.SendGIFResult{}),
	"send_dice":              reflect.TypeOf(types.SendDiceResult{}),

	// Chat settings
	"set_slow_mode":        reflect.TypeOf(types.SetSlowModeResult{}),
	"set_chat_permissions": reflect.TypeOf(types.SetChatPermissionsResult{}),
	"get_folders":          reflect.TypeOf(types.GetFoldersResult{}),
	"create_folder":        reflect.TypeOf(types.CreateFolderResult{}),
	"delete_folder":        reflect.TypeOf(types.DeleteFolderResult{}),
	"get_privacy":          reflect.TypeOf(types.GetPrivacyResult{}),
	"set_privacy":          reflect.TypeOf(types.SetPrivacyResult{}),

	// Gift operations
	"get_star_gifts":     reflect.TypeOf(types.GetStarGiftsResult{}),
	"send_star_gift":     reflect.TypeOf(types.SendStarGiftResult{}),
	"get_saved_gifts":    reflect.TypeOf(types.GetSavedGiftsResult{}),
	"transfer_star_gift": reflect.TypeOf(types.TransferStarGiftResult{}),
	"convert_star_gift":  reflect.TypeOf(types.ConvertStarGiftResult{}),
	"update_gift_price":  reflect.TypeOf(types.UpdateGiftPriceResult{}),
	"get_balance":        reflect.TypeOf(types.GetBalanceResult{}),
	"offer_gift":         reflect.TypeOf(types.OfferGiftResult{}),
	"get_gift_info":      reflect.TypeOf(types.GetGiftInfoResult{}),
	"get_gift_value":     reflect.TypeOf(types.GetGiftValueResult{}),
	"get_resale_gifts":   reflect.TypeOf(types.GetResaleGiftsResult{}),
	"buy_resale_gift":    reflect.TypeOf(types.BuyResaleGiftResult{}),
	"get_gift_attrs":     reflect.TypeOf(types.GetGiftAttrsResult{}),
	"accept_gift_offer":  reflect.TypeOf(types.AcceptGiftOfferResult{}),
	"decline_gift_offer": reflect.TypeOf(types.DeclineGiftOfferResult{}),
}

// commandMethods maps cobra commands to their IPC method names.
var commandMethods = map[*cobra.Command]string{}

// RegisterMethod associates a cobra command with its IPC method name.
// It also wraps the command's Args validator to skip validation when --schema is set.
func RegisterMethod(cmd *cobra.Command, method string) {
	commandMethods[cmd] = method
	if cmd.Args != nil {
		original := cmd.Args
		cmd.Args = func(c *cobra.Command, args []string) error {
			if s, _ := c.Flags().GetBool("schema"); s {
				return nil
			}
			return original(c, args)
		}
	}
}

// PrintCommandSchema outputs the result schema for a command and exits.
// Called from Execute() before cobra's validation to bypass required flag/arg checks.
func PrintCommandSchema(cmd *cobra.Command) {
	method, ok := commandMethods[cmd]
	if !ok {
		fmt.Fprintln(os.Stderr, "Error: no schema available for this command")
		os.Exit(1)
	}
	if !printSchema(method) {
		fmt.Fprintf(os.Stderr, "Error: no schema for method %q\n", method)
		os.Exit(1)
	}
	os.Exit(0)
}

var timeType = reflect.TypeOf(time.Time{})

// printSchema outputs the JSON schema for a method. Returns true if method was found.
func printSchema(method string) bool {
	t, ok := resultTypes[method]
	if !ok {
		return false
	}
	output := map[string]any{
		"method": method,
		"schema": buildSchema(t),
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
	return true
}

// buildSchema generates a schema map from a struct type via reflection.
func buildSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	schema := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Promote embedded struct fields
		if field.Anonymous {
			for k, v := range buildSchema(field.Type) {
				schema[k] = v
			}
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		optional := false
		for _, p := range parts[1:] {
			if p == "omitempty" {
				optional = true
			}
		}

		schema[name] = schemaType(field.Type, optional)
	}
	return schema
}

// schemaType returns the schema representation for a Go type.
func schemaType(t reflect.Type, optional bool) any {
	suffix := ""
	if optional {
		suffix = "?"
	}

	switch t.Kind() {
	case reflect.String:
		return "string" + suffix
	case reflect.Bool:
		return "bool" + suffix
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "int" + suffix
	case reflect.Int64:
		return "int64" + suffix
	case reflect.Float32, reflect.Float64:
		return "float64" + suffix
	case reflect.Slice:
		return schemaSlice(t, optional)
	case reflect.Map:
		return "object" + suffix
	case reflect.Struct:
		if t == timeType {
			return "string" + suffix
		}
		schema := buildSchema(t)
		if optional {
			schema["_optional"] = true
		}
		return schema
	case reflect.Ptr:
		return schemaType(t.Elem(), optional)
	}
	return "any" + suffix
}

// schemaSlice returns the schema for a slice type.
func schemaSlice(t reflect.Type, optional bool) any {
	suffix := ""
	if optional {
		suffix = "?"
	}
	elem := t.Elem()

	// []StructType → nested object with _type: array
	if elem.Kind() == reflect.Struct && elem != timeType {
		schema := buildSchema(elem)
		schema["_type"] = "array"
		if optional {
			schema["_optional"] = true
		}
		return schema
	}

	// []map → "[]object"
	if elem.Kind() == reflect.Map {
		return "[]object" + suffix
	}

	// [][]T → nested arrays
	if elem.Kind() == reflect.Slice {
		inner := schemaSlice(elem, false)
		if innerMap, ok := inner.(map[string]any); ok {
			result := map[string]any{"_type": "array", "_items": innerMap}
			if optional {
				result["_optional"] = true
			}
			return result
		}
		return "[]array" + suffix
	}

	// []basic → "[]type"
	return "[]" + scalarName(elem) + suffix
}

// scalarName returns the schema name for a scalar type.
func scalarName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "int"
	case reflect.Int64:
		return "int64"
	case reflect.Float32, reflect.Float64:
		return "float64"
	}
	return "any"
}
