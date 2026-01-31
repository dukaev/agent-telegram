// Package telegram provides wrapper methods for IPC compatibility.
package telegram

import (
	"context"

	"agent-telegram/telegram/types"
)

// GetChats returns the list of dialogs/chats with pagination.
func (c *Client) GetChats(ctx context.Context, limit, offset int) ([]map[string]any, error) {
	return c.chat.GetChats(ctx, limit, offset)
}

// GetMessages returns messages from a dialog.
func (c *Client) GetMessages(ctx context.Context, params types.GetMessagesParams) (*types.GetMessagesResult, error) {
	return c.message.GetMessages(ctx, params)
}

// GetUserInfo gets information about a user.
func (c *Client) GetUserInfo(ctx context.Context, params types.GetUserInfoParams) (*types.GetUserInfoResult, error) {
	return c.user.GetUserInfo(ctx, params)
}

// SendMessage sends a message to a peer.
func (c *Client) SendMessage(ctx context.Context, params types.SendMessageParams) (*types.SendMessageResult, error) {
	return c.message.SendMessage(ctx, params)
}

// SendReply sends a reply to a message.
func (c *Client) SendReply(ctx context.Context, params types.SendReplyParams) (*types.SendReplyResult, error) {
	return c.message.SendReply(ctx, params)
}

// UpdateMessage edits a message.
func (c *Client) UpdateMessage(
	ctx context.Context, params types.UpdateMessageParams,
) (*types.UpdateMessageResult, error) {
	return c.message.UpdateMessage(ctx, params)
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(
	ctx context.Context, params types.DeleteMessageParams,
) (*types.DeleteMessageResult, error) {
	return c.message.DeleteMessage(ctx, params)
}

// PinMessage pins a message.
func (c *Client) PinMessage(ctx context.Context, params types.PinMessageParams) (*types.PinMessageResult, error) {
	return c.pin.PinMessage(ctx, params)
}

// UnpinMessage unpins a message.
func (c *Client) UnpinMessage(ctx context.Context, params types.UnpinMessageParams) (*types.UnpinMessageResult, error) {
	return c.pin.UnpinMessage(ctx, params)
}

// InspectInlineButtons inspects inline buttons in a message.
func (c *Client) InspectInlineButtons(
	ctx context.Context, params types.InspectInlineButtonsParams,
) (*types.InspectInlineButtonsResult, error) {
	return c.message.InspectInlineButtons(ctx, params)
}

// PressInlineButton presses an inline button.
func (c *Client) PressInlineButton(
	ctx context.Context, params types.PressInlineButtonParams,
) (*types.PressInlineButtonResult, error) {
	return c.message.PressInlineButton(ctx, params)
}

// AddReaction adds a reaction to a message.
func (c *Client) AddReaction(ctx context.Context, params types.AddReactionParams) (*types.AddReactionResult, error) {
	return c.reaction.AddReaction(ctx, params)
}

// RemoveReaction removes reactions from a message.
func (c *Client) RemoveReaction(
	ctx context.Context, params types.RemoveReactionParams,
) (*types.RemoveReactionResult, error) {
	return c.reaction.RemoveReaction(ctx, params)
}

// ListReactions lists reactions on a message.
func (c *Client) ListReactions(
	ctx context.Context, params types.ListReactionsParams,
) (*types.ListReactionsResult, error) {
	return c.reaction.ListReactions(ctx, params)
}

// UpdateProfile updates the user's profile.
func (c *Client) UpdateProfile(
	ctx context.Context, params types.UpdateProfileParams,
) (*types.UpdateProfileResult, error) {
	return c.user.UpdateProfile(ctx, params)
}

// UpdateAvatar updates the user's avatar.
func (c *Client) UpdateAvatar(ctx context.Context, params types.UpdateAvatarParams) (*types.UpdateAvatarResult, error) {
	return c.user.UpdateAvatar(ctx, params)
}

// ClearMessages clears specific messages.
func (c *Client) ClearMessages(
	ctx context.Context, params types.ClearMessagesParams,
) (*types.ClearMessagesResult, error) {
	return c.chat.ClearMessages(ctx, params)
}

// ClearHistory clears all chat history for a peer.
func (c *Client) ClearHistory(ctx context.Context, params types.ClearHistoryParams) (*types.ClearHistoryResult, error) {
	return c.chat.ClearHistory(ctx, params)
}

// BlockPeer blocks a peer.
func (c *Client) BlockPeer(ctx context.Context, params types.BlockPeerParams) (*types.BlockPeerResult, error) {
	return c.user.BlockPeer(ctx, params)
}

// UnblockPeer unblocks a peer.
func (c *Client) UnblockPeer(ctx context.Context, params types.UnblockPeerParams) (*types.UnblockPeerResult, error) {
	return c.user.UnblockPeer(ctx, params)
}

// SendPhoto sends a photo to a peer.
func (c *Client) SendPhoto(ctx context.Context, params types.SendPhotoParams) (*types.SendPhotoResult, error) {
	return c.media.SendPhoto(ctx, params)
}

// SendVideo sends a video to a peer.
func (c *Client) SendVideo(ctx context.Context, params types.SendVideoParams) (*types.SendVideoResult, error) {
	return c.media.SendVideo(ctx, params)
}

// SendFile sends a file to a peer.
func (c *Client) SendFile(ctx context.Context, params types.SendFileParams) (*types.SendFileResult, error) {
	return c.media.SendFile(ctx, params)
}

// SendContact sends a contact to a peer.
func (c *Client) SendContact(ctx context.Context, params types.SendContactParams) (*types.SendContactResult, error) {
	return c.media.SendContact(ctx, params)
}

// SendLocation sends a location to a peer.
func (c *Client) SendLocation(ctx context.Context, params types.SendLocationParams) (*types.SendLocationResult, error) {
	return c.media.SendLocation(ctx, params)
}

// SendPoll sends a poll to a peer.
func (c *Client) SendPoll(ctx context.Context, params types.SendPollParams) (*types.SendPollResult, error) {
	return c.media.SendPoll(ctx, params)
}
