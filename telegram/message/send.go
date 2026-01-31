// Package message provides Telegram message sending operations.
package message

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendMessage sends a message to a peer.
func (c *Client) SendMessage(ctx context.Context, params types.SendMessageParams) (*types.SendMessageResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Clean peer (remove @ prefix)
	peer := strings.TrimPrefix(params.Peer, "@")

	// Resolve username to get input peer
	inputPeer, err := resolvePeer(ctx, c.api, peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	// Send message
	result, err := c.api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  params.Message,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Extract message ID from response
	msgID := extractMessageID(result)

	return &types.SendMessageResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Message: params.Message,
		Peer:    peer,
	}, nil
}

// SendReply sends a reply to a message.
func (c *Client) SendReply(ctx context.Context, params types.SendReplyParams) (*types.SendReplyResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	result, err := c.api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  params.Text,
		ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: int(params.MessageID)},
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send reply: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendReplyResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    params.Peer,
		Text:    params.Text,
		ReplyTo: params.MessageID,
	}, nil
}
