// Package telegram provides Telegram client reaction functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// AddReaction adds a reaction to a message.
func (c *Client) AddReaction(ctx context.Context, params AddReactionParams) (*AddReactionResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	// Create reaction from emoji
	reaction := createReaction(params.Emoji)

	_, err = api.MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
		Peer:     inputPeer,
		MsgID:    int(params.MessageID),
		Reaction: []tg.ReactionClass{reaction},
		Big:      params.Big,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	return &AddReactionResult{
		Success:   true,
		MessageID: params.MessageID,
		Emoji:     params.Emoji,
	}, nil
}

// createReaction creates a Reaction from emoji string.
func createReaction(emoji string) tg.ReactionClass {
	return &tg.ReactionEmoji{
		Emoticon: emoji,
	}
}

// RemoveReaction removes reactions from a message.
func (c *Client) RemoveReaction(ctx context.Context, params RemoveReactionParams) (*RemoveReactionResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
		Peer:     inputPeer,
		MsgID:    int(params.MessageID),
		Reaction: []tg.ReactionClass{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove reaction: %w", err)
	}

	return &RemoveReactionResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// ListReactions lists reactions on a message.
func (c *Client) ListReactions(ctx context.Context, params ListReactionsParams) (*ListReactionsResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Get messages to find reactions
	messages, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: int(params.MessageID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var reactions []Reaction

	// Extract reactions from the message
	switch m := messages.(type) {
	case *tg.MessagesMessages:
		for _, msg := range m.Messages {
			if userMsg, ok := msg.(*tg.Message); ok {
				reactions = extractReactions(userMsg.Reactions)
			}
		}
	case *tg.MessagesMessagesSlice:
		for _, msg := range m.Messages {
			if userMsg, ok := msg.(*tg.Message); ok {
				reactions = extractReactions(userMsg.Reactions)
			}
		}
	}

	return &ListReactionsResult{
		MessageID: params.MessageID,
		Reactions: reactions,
	}, nil
}

// extractReactions extracts reactions from message reactions.
func extractReactions(msgReactions tg.MessageReactions) []Reaction {
	//nolint:prealloc // Size is unknown upfront
	var result []Reaction

	for _, r := range msgReactions.Results {
		count := r.Count
		fromMe := r.ChosenOrder > 0

		// Get emoji
		emoji := ""
		if r.Reaction != nil {
			switch react := r.Reaction.(type) {
			case *tg.ReactionEmoji:
				emoji = react.Emoticon
			case *tg.ReactionCustomEmoji:
				emoji = fmt.Sprintf("custom:%d", react.DocumentID)
			}
		}

		result = append(result, Reaction{
			Emoji:  emoji,
			Count:  count,
			FromMe: fromMe,
		})
	}

	return result
}
