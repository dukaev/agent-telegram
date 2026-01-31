// Package reaction provides Telegram reaction operations.
package reaction

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides reaction operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new reaction client.
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

// createReaction creates a Reaction from emoji string.
func createReaction(emoji string) tg.ReactionClass {
	return &tg.ReactionEmoji{
		Emoticon: emoji,
	}
}

// AddReaction adds a reaction to a message.
func (c *Client) AddReaction(ctx context.Context, params types.AddReactionParams) (*types.AddReactionResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	// Create reaction from emoji
	reaction := createReaction(params.Emoji)

	_, err = c.api.MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
		Peer:     inputPeer,
		MsgID:    int(params.MessageID),
		Reaction: []tg.ReactionClass{reaction},
		Big:      params.Big,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	return &types.AddReactionResult{
		Success:   true,
		MessageID: params.MessageID,
		Emoji:     params.Emoji,
	}, nil
}

// RemoveReaction removes reactions from a message.
func (c *Client) RemoveReaction(
	ctx context.Context, params types.RemoveReactionParams,
) (*types.RemoveReactionResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
		Peer:     inputPeer,
		MsgID:    int(params.MessageID),
		Reaction: []tg.ReactionClass{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove reaction: %w", err)
	}

	return &types.RemoveReactionResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// extractReactions extracts reactions from message reactions.
func extractReactions(msgReactions tg.MessageReactions) []types.Reaction {
	//nolint:prealloc // Size is unknown upfront
	var result []types.Reaction

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

		result = append(result, types.Reaction{
			Emoji:  emoji,
			Count:  count,
			FromMe: fromMe,
		})
	}

	return result
}

// ListReactions lists reactions on a message.
func (c *Client) ListReactions(
	ctx context.Context, params types.ListReactionsParams,
) (*types.ListReactionsResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Get messages to find reactions
	messages, err := c.api.MessagesGetMessages(ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: int(params.MessageID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	var reactions []types.Reaction

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

	return &types.ListReactionsResult{
		MessageID: params.MessageID,
		Reactions: reactions,
	}, nil
}
