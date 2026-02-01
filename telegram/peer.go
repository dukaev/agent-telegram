package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
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
				AccessHash: getAccessHashFromPeerClass(peerClass, p.UserID),
			}
		case *tg.PeerChat:
			inputPeer = &tg.InputPeerChat{
				ChatID: p.ChatID,
			}
		case *tg.PeerChannel:
			inputPeer = &tg.InputPeerChannel{
				ChannelID:  p.ChannelID,
				AccessHash: getAccessHashFromPeerClass(peerClass, p.ChannelID),
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

// getAccessHashFromPeerClass extracts access hash from the resolved peer.
func getAccessHashFromPeerClass(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
	}
	return 0
}
