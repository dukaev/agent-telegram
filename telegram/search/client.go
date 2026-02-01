// Package search provides Telegram search operations.
package search

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides search operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new search client.
func NewClient(tc ParentClient) *Client {
	return &Client{
		parent: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (c *Client) resolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if c.parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return c.parent.ResolvePeer(ctx, peer)
}

// SearchGlobal searches for public chats, channels, and bots globally.
func (c *Client) SearchGlobal(ctx context.Context, params types.SearchGlobalParams) (*types.SearchGlobalResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Limit results
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Search globally - use ContactsSearch for finding users/channels/bots globally
	result, err := c.api.ContactsSearch(ctx, &tg.ContactsSearchRequest{
		Q:     params.Query,
		Limit: int(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search globally: %w", err)
	}

	var results []types.SearchResult

	// Process found results - ContactsFound contains results from peers
	// result is directly *tg.ContactsFound
	userMap := make(map[int64]*tg.User)
	for _, u := range result.Users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	chatMap := make(map[int64]*tg.Chat)
	channelMap := make(map[int64]*tg.Channel)
	for _, ch := range result.Chats {
		if chat, ok := ch.(*tg.Chat); ok {
			chatMap[chat.ID] = chat
		}
		if channel, ok := ch.(*tg.Channel); ok {
			channelMap[channel.ID] = channel
		}
	}

	// Extract results from my results (personalized)
	for _, peerClass := range result.MyResults {
		results = append(results, buildSearchResult(peerClass, userMap, chatMap, channelMap))
	}

	// Extract results from peers list
	for _, peerClass := range result.Results {
		results = append(results, buildSearchResult(peerClass, userMap, chatMap, channelMap))
	}

	return &types.SearchGlobalResult{
		Query:   params.Query,
		Type:    params.Type,
		Results: results,
		Count:   len(results),
	}, nil
}

// SearchInChat searches for messages within a specific chat.
func (c *Client) SearchInChat(ctx context.Context, params types.SearchInChatParams) (*types.SearchInChatResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve peer
	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Limit results
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Offset for pagination - use offset ID for MessagesSearch
	offsetID := params.Offset
	if offsetID < 0 {
		offsetID = 0
	}

	// Search in chat - without filter for now, will search all messages
	result, err := c.api.MessagesSearch(ctx, &tg.MessagesSearchRequest{
		Peer:     inputPeer,
		Q:        params.Query,
		TopMsgID: 0,
		Limit:    int(limit),
		OffsetID: int(offsetID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search in chat: %w", err)
	}

	// Extract results from different result types
	var messages []types.MessageResult
	var totalCount int

	switch r := result.(type) {
	case *tg.MessagesMessages:
		messages = extractMessages(r.Messages, r.Users, r.Chats)
		totalCount = len(messages)
	case *tg.MessagesMessagesSlice:
		messages = extractMessages(r.Messages, r.Users, r.Chats)
		totalCount = r.Count
	case *tg.MessagesChannelMessages:
		messages = extractMessages(r.Messages, r.Users, r.Chats)
		totalCount = r.Count
	}

	return &types.SearchInChatResult{
		Peer:     params.Peer,
		Query:    params.Query,
		Type:     params.Type,
		Messages: messages,
		Count:    len(messages),
		Total:    totalCount,
		Limit:    limit,
		Offset:   offsetID,
	}, nil
}

// buildSearchResult builds a search result from a peer class.
func buildSearchResult(peerClass tg.PeerClass, userMap map[int64]*tg.User, chatMap map[int64]*tg.Chat, channelMap map[int64]*tg.Channel) types.SearchResult {
	result := types.SearchResult{
		Peer: formatPeer(peerClass),
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
func extractMessages(messages []tg.MessageClass, users []tg.UserClass, chats []tg.ChatClass) []types.MessageResult {
	userMap := make(map[int64]*tg.User)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	var results []types.MessageResult
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
			PeerID:   formatPeer(m.PeerID),
			FromID:   formatFromID(m.FromID),
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

// formatPeer formats a peer class to string.
func formatPeer(peer tg.PeerClass) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("user:%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("channel:%d", p.ChannelID)
	}
	return ""
}

// formatFromID formats a from ID peer class to string.
func formatFromID(fromID tg.PeerClass) string {
	if fromID == nil {
		return ""
	}
	switch p := fromID.(type) {
	case *tg.PeerUser:
		return fmt.Sprintf("user:%d", p.UserID)
	case *tg.PeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		return fmt.Sprintf("channel:%d", p.ChannelID)
	}
	return ""
}

// getFromName extracts the sender name from peer.
func getFromName(fromID tg.PeerClass, userMap map[int64]*tg.User) string {
	if fromID == nil {
		return ""
	}
	if p, ok := fromID.(*tg.PeerUser); ok {
		if user, ok := userMap[p.UserID]; ok {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			if name == "" && user.Username != "" {
				name = user.Username
			}
			return name
		}
	}
	return ""
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
