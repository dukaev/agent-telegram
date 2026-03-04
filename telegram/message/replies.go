package message

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/types"
)

// GetReplies returns replies (comments) to a channel post.
func (c *Client) GetReplies(ctx context.Context, params types.GetRepliesParams) (*types.GetRepliesResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Defaults
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	peer := params.Peer
	if peer == "" {
		peer = params.Username
	}

	// Resolve channel peer
	inputPeer, err := c.ResolvePeer(ctx, normalizePeer(peer))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer %s: %w", peer, err)
	}

	// Get discussion message to find the linked discussion group + thread
	disc, err := c.API.MessagesGetDiscussionMessage(ctx, &tg.MessagesGetDiscussionMessageRequest{
		Peer:  inputPeer,
		MsgID: int(params.MessageID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get discussion message: %w", err)
	}

	if len(disc.Messages) == 0 {
		return &types.GetRepliesResult{
			Messages: []types.MessageResult{},
			Peer:     peer,
		}, nil
	}

	// The first message in disc.Messages is the discussion thread's top message.
	// Extract the discussion group peer from Chats and the top message ID.
	topMsg, ok := disc.Messages[0].(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected message type in discussion")
	}
	threadID := int64(topMsg.ID)

	// Find the discussion group peer from Chats
	var discussionPeer tg.InputPeerClass
	if topMsg.PeerID != nil {
		switch p := topMsg.PeerID.(type) {
		case *tg.PeerChannel:
			// Find the channel's access hash from disc.Chats
			for _, chat := range disc.Chats {
				if ch, ok := chat.(*tg.Channel); ok && ch.ID == p.ChannelID {
					discussionPeer = &tg.InputPeerChannel{
						ChannelID:  ch.ID,
						AccessHash: ch.AccessHash,
					}
					break
				}
			}
		case *tg.PeerChat:
			discussionPeer = &tg.InputPeerChat{ChatID: p.ChatID}
		}
	}
	if discussionPeer == nil {
		return nil, fmt.Errorf("could not resolve discussion group peer")
	}

	// Fetch replies in the thread
	repliesClass, err := c.API.MessagesGetReplies(ctx, &tg.MessagesGetRepliesRequest{
		Peer:     discussionPeer,
		MsgID:    int(threadID),
		Limit:    params.Limit,
		OffsetID: params.OffsetID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}

	messages, users := extractMessagesData(repliesClass)

	userMap := make(map[int64]tg.UserClass)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}

	messageResults := convertMessagesToResult(messages, userMap)

	return &types.GetRepliesResult{
		Messages: messageResults,
		Count:    len(messageResults),
		Peer:     peer,
		ThreadID: threadID,
	}, nil
}
