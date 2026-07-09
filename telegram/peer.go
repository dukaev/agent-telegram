package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/helpers"
)

// ResolvePeer resolves a peer string to InputPeerClass with caching.
// Uses singleflight to deduplicate concurrent resolutions of the same peer.
// Supported formats:
//   - @username - resolves via API
//   - 123456789 - positive number = user ID
//   - -123456789 - negative number = chat ID
//   - -100XXXXXXXXX - channel ID (common Telegram format)
func (c *Client) ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	cacheKey := peer
	if strings.HasPrefix(cacheKey, "@") {
		cacheKey = "@" + strings.ToLower(strings.TrimPrefix(cacheKey, "@"))
	}
	// Check cache first
	if cached, ok := c.peerCache.Load(cacheKey); ok {
		if inputPeer, ok := cached.(tg.InputPeerClass); ok {
			return inputPeer, nil
		}
	}

	// "me", "self", "current_user" resolves to InputPeerSelf (Saved Messages)
	if peer == "me" || peer == "self" || peer == "current_user" {
		inputPeer := &tg.InputPeerSelf{}
		c.peerCache.Store(peer, inputPeer)
		return inputPeer, nil
	}
	// Use singleflight to deduplicate concurrent resolutions of the same peer.
	// This prevents thundering herd when multiple handlers resolve the same peer.
	result, err, _ := c.peerFlight.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache inside singleflight (another goroutine may have populated it)
		if cached, ok := c.peerCache.Load(cacheKey); ok {
			if inputPeer, ok := cached.(tg.InputPeerClass); ok {
				return inputPeer, nil
			}
		}

		var inputPeer tg.InputPeerClass
		var err error

		if len(peer) > 0 && peer[0] == '@' {
			inputPeer, err = c.resolveUsername(ctx, peer[1:])
		} else {
			inputPeer, err = c.resolveNumericPeer(ctx, peer)
		}

		if err != nil {
			return nil, err
		}

		c.peerCache.Store(cacheKey, inputPeer)
		return inputPeer, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(tg.InputPeerClass), nil
}

// ResolvePeerID resolves a peer string and returns its normalized typed format
// (e.g., "channel:123", "user:456", "chat:789").
func (c *Client) ResolvePeerID(ctx context.Context, peer string) (string, error) {
	inputPeer, err := c.ResolvePeer(ctx, peer)
	if err != nil {
		return "", err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		return fmt.Sprintf("channel:%d", p.ChannelID), nil
	case *tg.InputPeerChat:
		return fmt.Sprintf("chat:%d", p.ChatID), nil
	case *tg.InputPeerUser:
		return fmt.Sprintf("user:%d", p.UserID), nil
	default:
		return peer, nil
	}
}

// CachePeer stores a resolved peer in the cache.
// This allows domain clients to populate the cache from API responses
// (e.g., discussion group peers discovered via messages.getDiscussionMessage).
func (c *Client) CachePeer(peer string, inputPeer tg.InputPeerClass) {
	if strings.HasPrefix(peer, "@") {
		peer = "@" + strings.ToLower(strings.TrimPrefix(peer, "@"))
	}
	c.peerCache.Store(peer, inputPeer)
}

// resolveUsername resolves a username via Telegram API.
func (c *Client) resolveUsername(ctx context.Context, username string) (tg.InputPeerClass, error) {
	tgClient := c.currentTelegramClient()
	if tgClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	peerClass, err := tgClient.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
	if err != nil {
		return nil, err
	}

	switch p := peerClass.Peer.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: helpers.GetAccessHash(peerClass, p.UserID),
		}, nil
	case *tg.PeerChat:
		return &tg.InputPeerChat{
			ChatID: p.ChatID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{
			ChannelID:  p.ChannelID,
			AccessHash: helpers.GetAccessHash(peerClass, p.ChannelID),
		}, nil
	default:
		return nil, fmt.Errorf("unknown peer type")
	}
}

// resolveNumericPeer resolves a numeric peer ID by looking up in dialogs.
func (c *Client) resolveNumericPeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	id, err := strconv.ParseInt(peer, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid peer format: %s", peer)
	}

	// Determine peer type based on ID format
	if id > 0 {
		// Positive = user ID, need to find access hash from dialogs
		return c.resolveUserByID(ctx, id)
	}

	// Negative ID
	absID := -id

	// Check if it's a channel (starts with 100)
	if strings.HasPrefix(peer, "-100") && len(peer) > 4 {
		// Channel ID format: -100XXXXXXXXX
		channelID := absID - 1000000000000
		if channelID > 0 {
			return c.resolveChannelByID(ctx, channelID)
		}
	}

	// Regular chat (basic group)
	return &tg.InputPeerChat{ChatID: absID}, nil
}

// resolveUserByID finds a user by ID from dialogs.
func (c *Client) resolveUserByID(ctx context.Context, userID int64) (tg.InputPeerClass, error) {
	if peer, err := c.findPeerInDialogs(ctx, func(page dialogPage) tg.InputPeerClass {
		for _, item := range page.users {
			if user, ok := item.(*tg.User); ok && user.ID == userID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	} else if peer != nil {
		return peer, nil
	}
	return nil, fmt.Errorf("user %d not found in dialogs", userID)
}

// resolveChannelByID finds a channel by ID from dialogs.
func (c *Client) resolveChannelByID(ctx context.Context, channelID int64) (tg.InputPeerClass, error) {
	if peer, err := c.findPeerInDialogs(ctx, func(page dialogPage) tg.InputPeerClass {
		for _, item := range page.chats {
			if channel, ok := item.(*tg.Channel); ok && channel.ID == channelID {
				return &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	} else if peer != nil {
		return peer, nil
	}
	return nil, fmt.Errorf("channel %d not found in dialogs", channelID)
}

type dialogPage struct {
	dialogs  []tg.DialogClass
	messages []tg.MessageClass
	chats    []tg.ChatClass
	users    []tg.UserClass
	complete bool
}

func (c *Client) findPeerInDialogs(
	ctx context.Context,
	match func(dialogPage) tg.InputPeerClass,
) (tg.InputPeerClass, error) {
	request := &tg.MessagesGetDialogsRequest{Limit: 100, OffsetPeer: &tg.InputPeerEmpty{}}
	tgClient := c.currentTelegramClient()
	if tgClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	for range 100 {
		result, err := tgClient.API().MessagesGetDialogs(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to get dialogs: %w", err)
		}
		page := unpackDialogPage(result)
		c.cacheDialogPage(page)
		if peer := match(page); peer != nil {
			return peer, nil
		}
		if page.complete || len(page.dialogs) < request.Limit {
			return nil, nil
		}
		last, ok := page.dialogs[len(page.dialogs)-1].(*tg.Dialog)
		if !ok {
			return nil, nil
		}
		offsetPeer := inputPeerFromPage(last.Peer, page)
		if offsetPeer == nil || last.TopMessage == request.OffsetID {
			return nil, nil
		}
		request.OffsetID = last.TopMessage
		request.OffsetDate = messageDate(page.messages, last.TopMessage)
		request.OffsetPeer = offsetPeer
	}
	return nil, fmt.Errorf("dialog lookup exceeded pagination limit")
}

func unpackDialogPage(result tg.MessagesDialogsClass) dialogPage {
	switch page := result.(type) {
	case *tg.MessagesDialogs:
		return dialogPage{page.Dialogs, page.Messages, page.Chats, page.Users, true}
	case *tg.MessagesDialogsSlice:
		return dialogPage{page.Dialogs, page.Messages, page.Chats, page.Users, false}
	default:
		return dialogPage{complete: true}
	}
}

func inputPeerFromPage(peer tg.PeerClass, page dialogPage) tg.InputPeerClass {
	switch value := peer.(type) {
	case *tg.PeerUser:
		for _, item := range page.users {
			if user, ok := item.(*tg.User); ok && user.ID == value.UserID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
			}
		}
	case *tg.PeerChat:
		return &tg.InputPeerChat{ChatID: value.ChatID}
	case *tg.PeerChannel:
		for _, item := range page.chats {
			if channel, ok := item.(*tg.Channel); ok && channel.ID == value.ChannelID {
				return &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			}
		}
	}
	return nil
}

func messageDate(messages []tg.MessageClass, id int) int {
	for _, message := range messages {
		if message.GetID() != id {
			continue
		}
		if value, ok := message.AsNotEmpty(); ok {
			return value.GetDate()
		}
	}
	return 0
}

func (c *Client) cacheDialogPage(page dialogPage) {
	for _, item := range page.users {
		if user, ok := item.(*tg.User); ok {
			peer := &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
			c.peerCache.Store(strconv.FormatInt(user.ID, 10), peer)
			if user.Username != "" {
				c.peerCache.Store("@"+strings.ToLower(user.Username), peer)
			}
		}
	}
	for _, item := range page.chats {
		if channel, ok := item.(*tg.Channel); ok {
			peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			c.peerCache.Store(fmt.Sprintf("-100%d", channel.ID), peer)
			if channel.Username != "" {
				c.peerCache.Store("@"+strings.ToLower(channel.Username), peer)
			}
		}
	}
}
