// Package helpers provides shared utility functions for Telegram operations.
package helpers

import (
	"fmt"

	"github.com/gotd/td/tg"
)

// PeerFormat specifies the format for peer string representation.
type PeerFormat int

const (
	// PeerFormatCompact uses compact format: user123, -456, -100789
	PeerFormatCompact PeerFormat = iota
	// PeerFormatTyped uses typed format: user:123, chat:456, channel:789
	PeerFormatTyped
)

// FormatPeer formats a peer to string using the specified format.
func FormatPeer(peer tg.PeerClass, format PeerFormat) string {
	if peer == nil {
		return ""
	}

	switch p := peer.(type) {
	case *tg.PeerUser:
		if format == PeerFormatTyped {
			return fmt.Sprintf("user:%d", p.UserID)
		}
		return fmt.Sprintf("user%d", p.UserID)
	case *tg.PeerChat:
		if format == PeerFormatTyped {
			return fmt.Sprintf("chat:%d", p.ChatID)
		}
		return fmt.Sprintf("-%d", p.ChatID)
	case *tg.PeerChannel:
		if format == PeerFormatTyped {
			return fmt.Sprintf("channel:%d", p.ChannelID)
		}
		return fmt.Sprintf("-100%d", p.ChannelID)
	default:
		return ""
	}
}

// GetAccessHash extracts access hash from the resolved peer.
func GetAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
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
