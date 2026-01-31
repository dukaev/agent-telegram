// Package user provides Telegram user block operations.
package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// BlockPeer blocks a peer.
func (c *Client) BlockPeer(ctx context.Context, params types.BlockPeerParams) (*types.BlockPeerResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.ContactsBlock(ctx, &tg.ContactsBlockRequest{
		ID: inputPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to block peer: %w", err)
	}

	return &types.BlockPeerResult{
		Success: true,
		Peer:    params.Peer,
	}, nil
}

// UnblockPeer unblocks a peer.
func (c *Client) UnblockPeer(ctx context.Context, params types.UnblockPeerParams) (*types.UnblockPeerResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = c.api.ContactsUnblock(ctx, &tg.ContactsUnblockRequest{
		ID: inputPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unblock peer: %w", err)
	}

	return &types.UnblockPeerResult{
		Success: true,
		Peer:    params.Peer,
	}, nil
}
