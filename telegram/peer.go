package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/helpers"
)

// ResolvePeer resolves a peer string to InputPeerClass with caching.
// This method is shared across all domain clients to avoid duplicate API calls.
func (c *Client) ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	// If peer starts with @, it's a username - resolve it with cache
	if len(peer) > 0 && peer[0] == '@' {
		// Check cache first
		if cached, ok := c.peerCache.Load(peer); ok {
			if inputPeer, ok := cached.(tg.InputPeerClass); ok {
				return inputPeer, nil
			}
		}

		// Not in cache, resolve from API
		peerClass, err := c.client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer[1:]})
		if err != nil {
			return nil, err
		}

		var inputPeer tg.InputPeerClass
		switch p := peerClass.Peer.(type) {
		case *tg.PeerUser:
			inputPeer = &tg.InputPeerUser{
				UserID:     p.UserID,
				AccessHash: helpers.GetAccessHash(peerClass, p.UserID),
			}
		case *tg.PeerChat:
			inputPeer = &tg.InputPeerChat{
				ChatID: p.ChatID,
			}
		case *tg.PeerChannel:
			inputPeer = &tg.InputPeerChannel{
				ChannelID:  p.ChannelID,
				AccessHash: helpers.GetAccessHash(peerClass, p.ChannelID),
			}
		default:
			return nil, fmt.Errorf("unknown peer type")
		}

		// Store in cache
		c.peerCache.Store(peer, inputPeer)
		return inputPeer, nil
	}

	// Try to parse as user ID
	// For now, just return empty peer (will be expanded later)
	return &tg.InputPeerEmpty{}, nil
}

