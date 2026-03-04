package message

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/tg"

	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
)

// resolveDiscussionPeer resolves the discussion group peer and thread ID for a channel post.
func (c *Client) resolveDiscussionPeer(
	ctx context.Context, channelPeer tg.InputPeerClass, msgID int64,
) (discussionPeer tg.InputPeerClass, threadID int64, err error) {
	disc, err := c.API.MessagesGetDiscussionMessage(ctx, &tg.MessagesGetDiscussionMessageRequest{
		Peer:  channelPeer,
		MsgID: int(msgID),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get discussion message: %w", err)
	}

	if len(disc.Messages) == 0 {
		return nil, 0, fmt.Errorf("no discussion thread found for message %d", msgID)
	}

	topMsg, ok := disc.Messages[0].(*tg.Message)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected message type in discussion")
	}
	threadID = int64(topMsg.ID)

	if topMsg.PeerID != nil {
		switch p := topMsg.PeerID.(type) {
		case *tg.PeerChannel:
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
		return nil, 0, fmt.Errorf("could not resolve discussion group peer")
	}

	// Cache the discussion peer so subsequent commands (send, reaction, etc.)
	// can resolve it by numeric ID without needing dialogs lookup.
	if ch, ok := discussionPeer.(*tg.InputPeerChannel); ok {
		peerStr := "-100" + strconv.FormatInt(ch.ChannelID, 10)
		c.CachePeer(peerStr, discussionPeer)
	} else if chat, ok := discussionPeer.(*tg.InputPeerChat); ok {
		peerStr := "-" + strconv.FormatInt(chat.ChatID, 10)
		c.CachePeer(peerStr, discussionPeer)
	}

	return discussionPeer, threadID, nil
}

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

	discussionPeer, threadID, err := c.resolveDiscussionPeer(ctx, inputPeer, params.MessageID)
	if err != nil {
		return nil, err
	}

	// Empty discussion case
	if discussionPeer == nil {
		return &types.GetRepliesResult{
			Messages: []types.MessageResult{},
			Peer:     peer,
		}, nil
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

// ReplyToComment replies to a specific comment in a channel post's discussion thread.
func (c *Client) ReplyToComment(ctx context.Context, params types.ReplyToCommentParams) (*types.ReplyToCommentResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
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

	// Get discussion group peer + thread ID
	discussionPeer, threadID, err := c.resolveDiscussionPeer(ctx, inputPeer, params.MessageID)
	if err != nil {
		return nil, err
	}

	// Send reply in the discussion thread
	parsed, entities := helpers.ParseCustomEmojis(params.Text)
	req := &tg.MessagesSendMessageRequest{
		Peer:    discussionPeer,
		Message: parsed,
		ReplyTo: &tg.InputReplyToMessage{
			ReplyToMsgID: int(params.CommentID),
			TopMsgID:     int(threadID),
		},
		RandomID: time.Now().UnixNano(),
	}
	if len(entities) > 0 {
		req.SetEntities(entities)
	}
	result, err := c.API.MessagesSendMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to reply to comment: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.ReplyToCommentResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     peer,
		Text:     params.Text,
		ReplyTo:  params.CommentID,
		ThreadID: threadID,
	}, nil
}
