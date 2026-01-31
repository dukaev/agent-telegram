// Package message provides Telegram message operations.
package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// UpdateMessage edits a message.
func (c *Client) UpdateMessage(
	ctx context.Context, params types.UpdateMessageParams,
) (*types.UpdateMessageResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
		Peer:    inputPeer,
		ID:      int(params.MessageID),
		Message: params.Text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return &types.UpdateMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(
	ctx context.Context, params types.DeleteMessageParams,
) (*types.DeleteMessageResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	_, err := c.api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     []int{int(params.MessageID)},
		Revoke: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return &types.DeleteMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// ForwardMessage forwards a message to another peer.
func (c *Client) ForwardMessage(
	ctx context.Context, params types.ForwardMessageParams,
) (*types.ForwardMessageResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	fromPeer, err := resolvePeer(ctx, c.api, params.FromPeer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve fromPeer: %w", err)
	}

	toPeer, err := resolvePeer(ctx, c.api, params.ToPeer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve toPeer: %w", err)
	}

	result, err := c.api.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
		FromPeer: fromPeer,
		ID:       []int{int(params.MessageID)},
		ToPeer:   toPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to forward message: %w", err)
	}

	newMsgID := extractMessageID(result)

	return &types.ForwardMessageResult{
		Success:   true,
		MessageID: newMsgID,
	}, nil
}
