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
	if err := c.setBlocked(ctx, params.Peer, true); err != nil {
		return nil, err
	}
	return &types.BlockPeerResult{Success: true, Peer: params.Peer}, nil
}

// UnblockPeer unblocks a peer.
func (c *Client) UnblockPeer(ctx context.Context, params types.UnblockPeerParams) (*types.UnblockPeerResult, error) {
	if err := c.setBlocked(ctx, params.Peer, false); err != nil {
		return nil, err
	}
	return &types.UnblockPeerResult{Success: true, Peer: params.Peer}, nil
}

// setBlocked is the shared implementation for block/unblock operations.
func (c *Client) setBlocked(ctx context.Context, peer string, block bool) error {
	if err := c.CheckInitialized(); err != nil {
		return err
	}

	inputPeer, err := c.ResolvePeer(ctx, peer)
	if err != nil {
		return err
	}

	if block {
		_, err = c.API.ContactsBlock(ctx, &tg.ContactsBlockRequest{ID: inputPeer})
	} else {
		_, err = c.API.ContactsUnblock(ctx, &tg.ContactsUnblockRequest{ID: inputPeer})
	}
	if err != nil {
		action := "block"
		if !block {
			action = "unblock"
		}
		return fmt.Errorf("failed to %s peer: %w", action, err)
	}
	return nil
}
