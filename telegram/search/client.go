// Package search provides Telegram search operations.
package search

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

// Client provides search operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new search client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}

// SearchGlobal searches for public chats, channels, and bots globally.
func (c *Client) SearchGlobal(ctx context.Context, params types.SearchGlobalParams) (*types.SearchGlobalResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Limit results
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Search globally - use ContactsSearch for finding users/channels/bots globally
	result, err := c.API.ContactsSearch(ctx, &tg.ContactsSearchRequest{
		Q:     params.Query,
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search globally: %w", err)
	}

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

	// Pre-allocate results slice
	results := make([]types.SearchResult, 0, len(result.MyResults)+len(result.Results))

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
	if c.API == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve peer
	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
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
	result, err := c.API.MessagesSearch(ctx, &tg.MessagesSearchRequest{
		Peer:     inputPeer,
		Q:        params.Query,
		TopMsgID: 0,
		Limit:    limit,
		OffsetID: offsetID,
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
