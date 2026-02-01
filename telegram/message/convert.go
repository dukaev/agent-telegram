package message

import (
	"fmt"

	"github.com/gotd/td/tg"
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
//
//nolint:gocognit,funlen // Function requires extracting multiple message fields
func convertMessagesToResult(messages []tg.MessageClass, userMap map[int64]tg.UserClass) []types.MessageResult {
	result := make([]types.MessageResult, 0, len(messages))
	for _, msgClass := range messages {
		msg, ok := msgClass.(*tg.Message)
		if !ok {
			continue
		}

		msgResult := types.MessageResult{
			ID:        int64(msg.ID),
			Date:      int64(msg.Date),
			Out:       msg.Out,
			Pinned:    msg.Pinned,
			Mentioned: msg.Mentioned,
			Silent:    msg.Silent,
			Post:      msg.Post,
		}

		// Extract text
		if msg.Message != "" {
			msgResult.Text = msg.Message
		}

		// Extract sender info
		if msg.FromID != nil {
			if fromUser, ok := msg.FromID.(*tg.PeerUser); ok {
				msgResult.FromID = fmt.Sprintf("user%d", fromUser.UserID)
				if user, ok := userMap[fromUser.UserID].(*tg.User); ok {
					msgResult.FromName = buildUserDisplayName(user)
				}
			}
		}

		// Extract peer ID (chat where message was sent)
		if msg.PeerID != nil {
			msgResult.PeerID = formatPeer(msg.PeerID)
		}

		// Extract edit date
		if msg.EditDate != 0 {
			msgResult.EditDate = int64(msg.EditDate)
		}

		// Extract media
		if msg.Media != nil {
			msgResult.Media = convertMedia(msg.Media)
		}

		// Extract views
		if msg.Views != 0 {
			msgResult.Views = msg.Views
		}

		// Extract forwards
		if msg.Forwards != 0 {
			msgResult.Forwards = msg.Forwards
		}

		// Extract reply to
		if msg.ReplyTo != nil {
			msgResult.ReplyTo = convertReplyHeader(msg.ReplyTo)
		}

		// Extract forward info
		if !msg.FwdFrom.Zero() {
			msgResult.FwdFrom = convertFwdHeader(msg.FwdFrom)
		}

		// Extract reactions
		if !msg.Reactions.Zero() {
			msgResult.Reactions = convertReactions(msg.Reactions)
		}

		// Extract entities (formatting)
		if len(msg.Entities) > 0 {
			msgResult.Entities = convertEntities(msg.Entities)
		}

		// Extract via bot ID
		if msg.ViaBotID != 0 {
			msgResult.ViaBotID = msg.ViaBotID
		}

		// Extract post author
		if msg.PostAuthor != "" {
			msgResult.PostAuthor = msg.PostAuthor
		}

		// Extract grouped ID (albums)
		if msg.GroupedID != 0 {
			msgResult.GroupedID = msg.GroupedID
		}

		// Extract TTL period
		if msg.TTLPeriod != 0 {
			msgResult.TTLPeriod = msg.TTLPeriod
		}

		// Extract inline buttons
		if msg.ReplyMarkup != nil {
			msgResult.Buttons = extractButtons(msg.ReplyMarkup)
		}

		result = append(result, msgResult)
	}
	return result
}

// formatPeer formats a peer to string.
func formatPeer(peer tg.PeerClass) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("user%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("-%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("-100%d", p.ChannelID)
	default:
		return ""
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
	result["from_id"] = formatPeer(fwdHeader.FromID)
	if fwdHeader.Date != 0 {
		result["date"] = fwdHeader.Date
	}
	if fwdHeader.FromName != "" {
		result["from_name"] = fwdHeader.FromName
	}
	return result
}
