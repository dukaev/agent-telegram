// Package telegram provides Telegram client block functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// BlockPeer blocks a peer.
func (c *Client) BlockPeer(ctx context.Context, params BlockPeerParams) (*BlockPeerResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.ContactsBlock(ctx, &tg.ContactsBlockRequest{
		ID: inputPeer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to block peer: %w", err)
	}

	return &BlockPeerResult{
		Success: true,
		Peer:    params.Peer,
	}, nil
}
