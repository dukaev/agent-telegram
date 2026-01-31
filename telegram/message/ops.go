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
