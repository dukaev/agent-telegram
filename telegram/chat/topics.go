package chat

import (
	"context"
	"fmt"

	"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// GetTopics retrieves forum topics from a resolved Telegram peer.
func (c *Client) GetTopics(ctx context.Context, params types.GetTopicsParams) (*types.GetTopicsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Resolve peer to get InputChannel
	peer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Set limit
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	// Get forum topics using MessagesGetForumTopics
	result, err := c.API().MessagesGetForumTopics(ctx, &tg.MessagesGetForumTopicsRequest{
		Peer:  peer,
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get forum topics: %w", err)
	}

	topicsResult := &types.GetTopicsResult{
		Peer:   params.Peer,
		Topics: []types.ForumTopic{},
	}

	// Extract topics from result
	for _, topicClass := range result.Topics {
		if topic, ok := topicClass.(*tg.ForumTopic); ok {
			topicsResult.Topics = append(topicsResult.Topics, types.ForumTopic{
				ID:        int64(topic.ID),
				Title:     topic.Title,
				IconColor: int32(topic.IconColor), //nolint:gosec // IconColor is always within int32 range
				Top:       topic.Pinned,
				Closed:    topic.Closed,
			})
		}
	}
	topicsResult.Count = len(topicsResult.Topics)

	return topicsResult, nil
}
