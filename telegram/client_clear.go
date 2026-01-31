// Package telegram provides Telegram client clear functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// ClearMessages clears specific messages.
func (c *Client) ClearMessages(ctx context.Context, params ClearMessagesParams) (*ClearMessagesResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	// Convert int64 slice to int slice for API
	ids := make([]int, len(params.MessageIDs))
	for i, id := range params.MessageIDs {
		ids[i] = int(id)
	}

	_, err := api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     ids,
		Revoke: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clear messages: %w", err)
	}

	return &ClearMessagesResult{
		Success: true,
		Cleared: len(ids),
		Peer:    params.Peer,
	}, nil
}

// ClearHistory clears all chat history for a peer.
func (c *Client) ClearHistory(ctx context.Context, params ClearHistoryParams) (*ClearHistoryResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.MessagesDeleteHistory(ctx, &tg.MessagesDeleteHistoryRequest{
		Peer:   inputPeer,
		Revoke: params.Revoke,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clear history: %w", err)
	}

	return &ClearHistoryResult{
		Success: true,
		Peer:    params.Peer,
		Revoke:  params.Revoke,
	}, nil
}
