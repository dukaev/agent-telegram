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
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Convert int64 slice to int slice for API
	ids := make([]int, len(params.MessageIDs))
	for i, id := range params.MessageIDs {
		ids[i] = int(id)
	}

	_, err := c.API.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
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
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.API.MessagesDeleteHistory(ctx, &tg.MessagesDeleteHistoryRequest{
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

// PinChat pins or unpins a chat in the dialog list.
func (c *Client) PinChat(ctx context.Context, params types.PinChatParams) (*types.PinChatResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	inputPeer, err := c.ResolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	// Create InputDialogPeer from InputPeer
	dialogPeer := &tg.InputDialogPeer{
		Peer: inputPeer,
	}

	_, err = c.API.MessagesToggleDialogPin(ctx, &tg.MessagesToggleDialogPinRequest{
		Pinned: !params.Disable,
		Peer:   dialogPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to toggle dialog pin: %w", err)
	}

	return &types.PinChatResult{
		Success: true,
		Peer:    params.Peer,
		Pinned:  !params.Disable,
	}, nil
}
