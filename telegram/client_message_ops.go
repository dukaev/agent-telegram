// Package telegram provides Telegram client message operations.
package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// SendReply sends a reply to a message.
func (c *Client) SendReply(ctx context.Context, params SendReplyParams) (*SendReplyResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	result, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     inputPeer,
		Message:  params.Text,
		ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: int(params.MessageID)},
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send reply: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendReplyResult{
		ID:      msgID,
		Date:    time.Now().Unix(),
		Peer:    params.Peer,
		Text:    params.Text,
		ReplyTo: params.MessageID,
	}, nil
}

// UpdateMessage edits a message.
func (c *Client) UpdateMessage(ctx context.Context, params UpdateMessageParams) (*UpdateMessageResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	_, err = api.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
		Peer:      inputPeer,
		ID:        int(params.MessageID),
		Message:   params.Text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return &UpdateMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(ctx context.Context, params DeleteMessageParams) (*DeleteMessageResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	_, err := api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     []int{int(params.MessageID)},
		Revoke: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return &DeleteMessageResult{
		Success:   true,
		MessageID: params.MessageID,
	}, nil
}
