package search

import (
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
)

// buildSearchResult builds a search result from a peer class.
func buildSearchResult(
	peerClass tg.PeerClass,
	userMap map[int64]*tg.User,
	chatMap map[int64]*tg.Chat,
	channelMap map[int64]*tg.Channel,
) types.SearchResult {
	result := types.SearchResult{
		Peer: helpers.FormatPeer(peerClass, helpers.PeerFormatTyped),
	}

	// Add names/IDs based on peer type
	switch p := peerClass.(type) {
	case *tg.PeerUser:
		if user, ok := userMap[p.UserID]; ok {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			if name == "" && user.Username != "" {
				name = user.Username
			}
			result.FromName = name
			if user.Bot {
				result.Peer = "bot:" + user.Username
			}
		}
	case *tg.PeerChat:
		if chat, ok := chatMap[p.ChatID]; ok {
			result.FromName = chat.Title
		}
	case *tg.PeerChannel:
		if channel, ok := channelMap[p.ChannelID]; ok {
			result.FromName = channel.Title
		}
	}

	return result
}

// extractMessages extracts message results from messages.
func extractMessages(messages []tg.MessageClass, users []tg.UserClass, _ []tg.ChatClass) []types.MessageResult {
	userMap := make(map[int64]*tg.User)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	results := make([]types.MessageResult, 0, len(messages))
	for _, msg := range messages {
		m, ok := msg.(*tg.Message)
		if !ok {
			continue
		}

		result := types.MessageResult{
			ID:       int64(m.ID),
			Date:     int64(m.Date),
			Text:     m.Message,
			Out:      m.Out,
			PeerID:   helpers.FormatPeer(m.PeerID, helpers.PeerFormatTyped),
			FromID:   helpers.FormatPeer(m.FromID, helpers.PeerFormatTyped),
			FromName: getFromName(m.FromID, userMap),
		}

		// Add media info if present
		if m.Media != nil {
			result.Media = extractMediaInfo(m.Media)
		}

		results = append(results, result)
	}

	return results
}

// getFromName extracts the sender name from peer.
func getFromName(fromID tg.PeerClass, userMap map[int64]*tg.User) string {
	if fromID == nil {
		return ""
	}

	p, ok := fromID.(*tg.PeerUser)
	if !ok {
		return ""
	}

	user, ok := userMap[p.UserID]
	if !ok {
		return ""
	}

	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if name == "" && user.Username != "" {
		name = user.Username
	}
	return name
}

// extractMediaInfo extracts media information.
func extractMediaInfo(media tg.MessageMediaClass) map[string]any {
	info := make(map[string]any)
	switch media.(type) {
	case *tg.MessageMediaPhoto:
		info["type"] = "photo"
	case *tg.MessageMediaDocument:
		info["type"] = "document"
	case *tg.MessageMediaWebPage:
		info["type"] = "webpage"
	case *tg.MessageMediaGeo:
		info["type"] = "geo"
	case *tg.MessageMediaContact:
		info["type"] = "contact"
	case *tg.MessageMediaPoll:
		info["type"] = "poll"
	}
	return info
}
