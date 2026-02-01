package message

import (
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
)

// buildUserDisplayName builds a display name from a user.
func buildUserDisplayName(user *tg.User) string {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if name == "" {
		name = user.Username
	}
	if name == "" {
		name = ""
	}
	return name
}

// extractMessageID extracts message ID from response.
func extractMessageID(result tg.UpdatesClass) int64 {
	switch r := result.(type) {
	case *tg.Updates:
		if len(r.Updates) > 0 {
			if msg, ok := r.Updates[0].(*tg.UpdateMessageID); ok {
				return int64(msg.ID)
			}
		}
	case *tg.UpdateShortSentMessage:
		return int64(r.ID)
	}
	return 0
}

// extractMessagesData extracts messages and users from the response.
func extractMessagesData(messagesClass tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.UserClass) {
	switch m := messagesClass.(type) {
	case *tg.MessagesMessages:
		return m.Messages, m.Users
	case *tg.MessagesMessagesSlice:
		return m.Messages, m.Users
	case *tg.MessagesChannelMessages:
		return m.Messages, m.Users
	default:
		return nil, nil
	}
}

// convertMessagesToResult converts messages to the result format.
func convertMessagesToResult(messages []tg.MessageClass, userMap map[int64]tg.UserClass) []types.MessageResult {
	result := make([]types.MessageResult, 0, len(messages))
	for _, msgClass := range messages {
		msg, ok := msgClass.(*tg.Message)
		if !ok {
			continue
		}
		result = append(result, convertMessage(msg, userMap))
	}
	return result
}

// convertMessage converts a single message to MessageResult.
func convertMessage(msg *tg.Message, userMap map[int64]tg.UserClass) types.MessageResult {
	r := types.MessageResult{
		ID:        int64(msg.ID),
		Date:      int64(msg.Date),
		Text:      msg.Message,
		Out:       msg.Out,
		Pinned:    msg.Pinned,
		Mentioned: msg.Mentioned,
		Silent:    msg.Silent,
		Post:      msg.Post,
	}

	extractSenderInfo(msg, userMap, &r)
	extractMessageMeta(msg, &r)
	extractMessageContent(msg, &r)

	return r
}

// extractSenderInfo extracts sender information from message.
func extractSenderInfo(msg *tg.Message, userMap map[int64]tg.UserClass, r *types.MessageResult) {
	if msg.FromID == nil {
		return
	}
	fromUser, ok := msg.FromID.(*tg.PeerUser)
	if !ok {
		return
	}
	r.FromID = fmt.Sprintf("user%d", fromUser.UserID)
	if user, ok := userMap[fromUser.UserID].(*tg.User); ok {
		r.FromName = buildUserDisplayName(user)
	}
}

// extractMessageMeta extracts message metadata.
func extractMessageMeta(msg *tg.Message, r *types.MessageResult) {
	if msg.PeerID != nil {
		r.PeerID = helpers.FormatPeer(msg.PeerID, helpers.PeerFormatCompact)
	}
	if msg.EditDate != 0 {
		r.EditDate = int64(msg.EditDate)
	}
	if msg.Views != 0 {
		r.Views = msg.Views
	}
	if msg.Forwards != 0 {
		r.Forwards = msg.Forwards
	}
	if msg.ViaBotID != 0 {
		r.ViaBotID = msg.ViaBotID
	}
	if msg.PostAuthor != "" {
		r.PostAuthor = msg.PostAuthor
	}
	if msg.GroupedID != 0 {
		r.GroupedID = msg.GroupedID
	}
	if msg.TTLPeriod != 0 {
		r.TTLPeriod = msg.TTLPeriod
	}
}

// extractMessageContent extracts message content (media, reactions, etc).
func extractMessageContent(msg *tg.Message, r *types.MessageResult) {
	if msg.Media != nil {
		r.Media = convertMedia(msg.Media)
	}
	if msg.ReplyTo != nil {
		r.ReplyTo = convertReplyHeader(msg.ReplyTo)
	}
	if !msg.FwdFrom.Zero() {
		r.FwdFrom = convertFwdHeader(msg.FwdFrom)
	}
	if !msg.Reactions.Zero() {
		r.Reactions = convertReactions(msg.Reactions)
	}
	if len(msg.Entities) > 0 {
		r.Entities = convertEntities(msg.Entities)
	}
	if msg.ReplyMarkup != nil {
		r.Buttons = extractButtons(msg.ReplyMarkup)
	}
}

// convertMedia converts media to map.
func convertMedia(media tg.MessageMediaClass) map[string]any {
	result := make(map[string]any)
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		result["type"] = "photo"
		if m.Photo != nil {
			if photo, ok := m.Photo.(*tg.Photo); ok {
				result["photo_id"] = photo.ID
			}
		}
	case *tg.MessageMediaDocument:
		result["type"] = "document"
		if m.Document != nil {
			if doc, ok := m.Document.(*tg.Document); ok {
				result["document_id"] = doc.ID
			}
		}
	case *tg.MessageMediaWebPage:
		result["type"] = "webpage"
		if m.Webpage != nil {
			if wp, ok := m.Webpage.(*tg.WebPage); ok {
				result["url"] = wp.URL
				result["display_url"] = wp.DisplayURL
			}
		}
	case *tg.MessageMediaGeo:
		result["type"] = "geo"
	case *tg.MessageMediaContact:
		result["type"] = "contact"
	case *tg.MessageMediaPoll:
		result["type"] = "poll"
	default:
		result["type"] = "unknown"
	}
	return result
}

// convertReplyHeader converts reply header to map.
func convertReplyHeader(replyHeader tg.MessageReplyHeaderClass) map[string]any {
	result := make(map[string]any)
	if r, ok := replyHeader.(*tg.MessageReplyHeader); ok {
		result["reply_to_msg_id"] = r.ReplyToMsgID
		if r.ForumTopic {
			result["forum_topic"] = true
		}
	}
	return result
}

// convertFwdHeader converts forward header to map.
func convertFwdHeader(fwdHeader tg.MessageFwdHeader) map[string]any {
	result := make(map[string]any)
	result["from_id"] = helpers.FormatPeer(fwdHeader.FromID, helpers.PeerFormatCompact)
	if fwdHeader.Date != 0 {
		result["date"] = fwdHeader.Date
	}
	if fwdHeader.FromName != "" {
		result["from_name"] = fwdHeader.FromName
	}
	return result
}
