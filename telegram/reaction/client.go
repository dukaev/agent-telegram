// Package reaction provides Telegram reaction operations.
package reaction

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides reaction operations.
type Client struct {
	api      *tg.Client
	telegram *telegram.Client
}

// NewClient creates a new reaction client.
func NewClient(tc *telegram.Client) *Client {
	return &Client{
		telegram: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves peer from username.
func resolvePeer(ctx context.Context, api *tg.Client, peer string) (tg.InputPeerClass, error) {
	peer = strings.TrimPrefix(peer, "@")

	peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	switch p := peerClass.Peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: getAccessHash(peerClass, p.UserID),
		}, nil
	case *tg.PeerChat:
		return &tg.InputPeerChat{
			ChatID: p.ChatID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{
			ChannelID:  p.ChannelID,
			AccessHash: getAccessHash(peerClass, p.ChannelID),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported peer type: %T", p)
	}
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

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
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

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
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
