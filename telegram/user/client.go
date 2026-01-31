// Package user provides Telegram user operations.
package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// Client provides user operations.
type Client struct {
	api      *tg.Client
	telegram *telegram.Client
}

// NewClient creates a new user client.
func NewClient(tc *telegram.Client) *Client {
	return &Client{
		telegram: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves peer from username.
func resolvePeer(ctx context.Context, api *tg.Client, peer string) (tg.InputPeerClass, error) {
	peer = strings.TrimPrefix(peer, "@")

	peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: peer})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	switch p := peerClass.Peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: getAccessHash(peerClass, p.UserID),
		}, nil
	case *tg.PeerChat:
		return &tg.InputPeerChat{
			ChatID: p.ChatID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{
			ChannelID:  p.ChannelID,
			AccessHash: getAccessHash(peerClass, p.ChannelID),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported peer type: %T", p)
	}
}

// getAccessHash extracts access hash from the resolved peer.
func getAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
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

// trimUsernamePrefix removes the @ prefix from username.
func trimUsernamePrefix(username string) string {
	if len(username) > 0 && username[0] == '@' {
		return username[1:]
	}
	return username
}
