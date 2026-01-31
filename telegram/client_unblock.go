// Package telegram provides Telegram client unblock functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// UnblockPeer unblocks a peer.
func (c *Client) UnblockPeer(ctx context.Context, params UnblockPeerParams) (*UnblockPeerResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.ContactsUnblock(ctx, &tg.ContactsUnblockRequest{
		ID: inputPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unblock peer: %w", err)
	}

	return &UnblockPeerResult{
		Success: true,
		Peer:    params.Peer,
	}, nil
}
