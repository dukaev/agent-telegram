// Package message provides scheduled message operations.
package message

import (
	"context"
	"fmt"

	"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// GetScheduledMessages retrieves scheduled messages for a chat.
func (c *Client) GetScheduledMessages(
	ctx context.Context,
	params types.GetScheduledMessagesParams,
) (*types.GetScheduledMessagesResult, error) {
	inputPeer, err := c.InitAndResolve(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	result, err := c.API().MessagesGetScheduledHistory(ctx, &tg.MessagesGetScheduledHistoryRequest{
		Peer: inputPeer,
		Hash: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled messages: %w", err)
	}

	messages := make([]types.ScheduledMessage, 0)

	switch r := result.(type) {
	case *tg.MessagesMessages:
		for _, msg := range r.Messages {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, types.ScheduledMessage{
					ID:      int64(m.ID),
					Date:    int64(m.Date),
					Message: m.Message,
					Peer:    params.Peer,
				})
			}
		}
	case *tg.MessagesMessagesSlice:
		for _, msg := range r.Messages {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, types.ScheduledMessage{
					ID:      int64(m.ID),
					Date:    int64(m.Date),
					Message: m.Message,
					Peer:    params.Peer,
				})
			}
		}
	case *tg.MessagesChannelMessages:
		for _, msg := range r.Messages {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, types.ScheduledMessage{
					ID:      int64(m.ID),
					Date:    int64(m.Date),
					Message: m.Message,
					Peer:    params.Peer,
				})
			}
		}
	}

	return &types.GetScheduledMessagesResult{
		Messages: messages,
		Count:    len(messages),
	}, nil
}
