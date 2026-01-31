// Package chat provides Telegram clear operations.
package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// ClearMessages clears specific messages.
func (c *Client) ClearMessages(
	ctx context.Context, params types.ClearMessagesParams,
) (*types.ClearMessagesResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Convert int64 slice to int slice for API
	ids := make([]int, len(params.MessageIDs))
	for i, id := range params.MessageIDs {
		ids[i] = int(id)
	}

	_, err := c.api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     ids,
		Revoke: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clear messages: %w", err)
	}

	return &types.ClearMessagesResult{
		Success: true,
		Cleared: len(ids),
		Peer:    params.Peer,
	}, nil
}

// ClearHistory clears all chat history for a peer.
func (c *Client) ClearHistory(ctx context.Context, params types.ClearHistoryParams) (*types.ClearHistoryResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.MessagesDeleteHistory(ctx, &tg.MessagesDeleteHistoryRequest{
		Peer:   inputPeer,
		Revoke: params.Revoke,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clear history: %w", err)
	}

	return &types.ClearHistoryResult{
		Success: true,
		Peer:    params.Peer,
		Revoke:  params.Revoke,
	}, nil
}
