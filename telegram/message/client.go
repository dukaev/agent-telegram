// Package message provides Telegram message operations.
package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides message operations.
type Client struct {
	api *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new message client.
func NewClient(tc ParentClient) *Client {
	return &Client{
		parent: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// MessageResult represents a single message result.
type MessageResult struct { // revive:disable:exported // Used internally
	ID       int64  `json:"id"`
	Date     int64  `json:"date"`
	Text     string `json:"text,omitempty"`
	FromID   string `json:"fromId,omitempty"`
	FromName string `json:"fromName,omitempty"`
	Out      bool   `json:"out"`
}

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

// resolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (c *Client) resolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if c.parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return c.parent.ResolvePeer(ctx, peer)
}

// getAccessHash extracts access hash from the resolved peer.
func getAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
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
	switch r := replyHeader.(type) {
	case *tg.MessageReplyHeader:
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

// convertReactions converts reactions to map slice.
func convertReactions(reactions tg.MessageReactions) []map[string]any {
	result := make([]map[string]any, 0, len(reactions.Results))
	for _, r := range reactions.Results {
		reaction := map[string]any{
			"count": r.Count,
		}
		if r.Reaction != nil {
			switch react := r.Reaction.(type) {
			case *tg.ReactionEmoji:
				reaction["emoticon"] = react.Emoticon
			case *tg.ReactionCustomEmoji:
				reaction["document_id"] = react.DocumentID
			}
		}
		if r.ChosenOrder != 0 {
			reaction["chosen_order"] = r.ChosenOrder
		}
		result = append(result, reaction)
	}
	return result
}

// convertEntities converts message entities to map slice.
func convertEntities(entities []tg.MessageEntityClass) []map[string]any {
	result := make([]map[string]any, 0, len(entities))
	for _, e := range entities {
		entity := map[string]any{
			"offset": e.GetOffset(),
			"length": e.GetLength(),
		}
		switch ent := e.(type) {
		case *tg.MessageEntityTextURL:
			entity["type"] = "text_url"
			entity["url"] = ent.URL
		case *tg.MessageEntityURL:
			entity["type"] = "url"
		case *tg.MessageEntityEmail:
			entity["type"] = "email"
		case *tg.MessageEntityHashtag:
			entity["type"] = "hashtag"
		case *tg.MessageEntityCashtag:
			entity["type"] = "cashtag"
		case *tg.MessageEntityMention:
			entity["type"] = "mention"
		case *tg.MessageEntityMentionName:
			entity["type"] = "mention_name"
			entity["user_id"] = ent.UserID
		case *tg.MessageEntityBotCommand:
			entity["type"] = "bot_command"
		case *tg.MessageEntityBold:
			entity["type"] = "bold"
		case *tg.MessageEntityItalic:
			entity["type"] = "italic"
		case *tg.MessageEntityUnderline:
			entity["type"] = "underline"
		case *tg.MessageEntityStrike:
			entity["type"] = "strike"
		case *tg.MessageEntityCode:
			entity["type"] = "code"
		case *tg.MessageEntityPre:
			entity["type"] = "pre"
			if ent.Language != "" {
				entity["language"] = ent.Language
			}
		case *tg.MessageEntityBlockquote:
			entity["type"] = "blockquote"
		}
		result = append(result, entity)
	}
	return result
}
