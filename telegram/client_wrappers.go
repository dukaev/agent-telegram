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

// ForwardMessage forwards a message to another peer.
func (c *Client) ForwardMessage(
	ctx context.Context, params types.ForwardMessageParams,
) (*types.ForwardMessageResult, error) {
	return c.message.ForwardMessage(ctx, params)
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

// PinChat pins or unpins a chat in the dialog list.
func (c *Client) PinChat(ctx context.Context, params types.PinChatParams) (*types.PinChatResult, error) {
	return c.chat.PinChat(ctx, params)
}

// JoinChat joins a chat or channel using an invite link.
func (c *Client) JoinChat(ctx context.Context, params types.JoinChatParams) (*types.JoinChatResult, error) {
	return c.chat.JoinChat(ctx, params)
}

// SubscribeChannel subscribes to a public channel.
func (c *Client) SubscribeChannel(ctx context.Context, params types.SubscribeChannelParams) (*types.SubscribeChannelResult, error) {
	return c.chat.SubscribeChannel(ctx, params)
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

// GetContacts returns the user's contact list with optional search filter.
func (c *Client) GetContacts(ctx context.Context, params types.GetContactsParams) (*types.GetContactsResult, error) {
	return c.user.GetContacts(ctx, params)
}

// SearchGlobal searches for public chats, channels, and bots globally.
func (c *Client) SearchGlobal(ctx context.Context, params types.SearchGlobalParams) (*types.SearchGlobalResult, error) {
	return c.search.SearchGlobal(ctx, params)
}

// SearchInChat searches for messages within a specific chat.
func (c *Client) SearchInChat(ctx context.Context, params types.SearchInChatParams) (*types.SearchInChatResult, error) {
	return c.search.SearchInChat(ctx, params)
}

// AddContact adds a new contact to the user's contact list.
func (c *Client) AddContact(ctx context.Context, params types.AddContactParams) (*types.AddContactResult, error) {
	return c.user.AddContact(ctx, params)
}

// DeleteContact deletes a contact from the user's contact list.
func (c *Client) DeleteContact(ctx context.Context, params types.DeleteContactParams) (*types.DeleteContactResult, error) {
	return c.user.DeleteContact(ctx, params)
}

// GetTopics retrieves forum topics from a channel.
func (c *Client) GetTopics(ctx context.Context, params types.GetTopicsParams) (*types.GetTopicsResult, error) {
	return c.chat.GetTopics(ctx, params)
}

// CreateGroup creates a new group chat.
func (c *Client) CreateGroup(ctx context.Context, params types.CreateGroupParams) (*types.CreateGroupResult, error) {
	return c.chat.CreateGroup(ctx, params)
}

// CreateChannel creates a new channel or supergroup.
func (c *Client) CreateChannel(ctx context.Context, params types.CreateChannelParams) (*types.CreateChannelResult, error) {
	return c.chat.CreateChannel(ctx, params)
}

// EditTitle edits the title of a chat or channel.
func (c *Client) EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error) {
	return c.chat.EditTitle(ctx, params)
}

// SetPhoto sets the photo for a chat or channel.
func (c *Client) SetPhoto(ctx context.Context, params types.SetPhotoParams) (*types.SetPhotoResult, error) {
	return c.chat.SetPhoto(ctx, params)
}

// DeletePhoto deletes the photo from a chat or channel.
func (c *Client) DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error) {
	return c.chat.DeletePhoto(ctx, params)
}

// Leave leaves a chat or channel.
func (c *Client) Leave(ctx context.Context, params types.LeaveParams) (*types.LeaveResult, error) {
	return c.chat.Leave(ctx, params)
}

// Invite invites users to a chat or channel.
func (c *Client) Invite(ctx context.Context, params types.InviteParams) (*types.InviteResult, error) {
	return c.chat.Invite(ctx, params)
}

// GetParticipants retrieves participants from a chat or channel.
func (c *Client) GetParticipants(ctx context.Context, params types.GetParticipantsParams) (*types.GetParticipantsResult, error) {
	return c.chat.GetParticipants(ctx, params)
}

// GetAdmins retrieves admins from a chat or channel.
func (c *Client) GetAdmins(ctx context.Context, params types.GetAdminsParams) (*types.GetAdminsResult, error) {
	return c.chat.GetAdmins(ctx, params)
}

// GetBanned retrieves banned users from a chat or channel.
func (c *Client) GetBanned(ctx context.Context, params types.GetBannedParams) (*types.GetBannedResult, error) {
	return c.chat.GetBanned(ctx, params)
}

// PromoteAdmin promotes a user to admin.
func (c *Client) PromoteAdmin(ctx context.Context, params types.PromoteAdminParams) (*types.PromoteAdminResult, error) {
	return c.chat.PromoteAdmin(ctx, params)
}

// DemoteAdmin demotes an admin to regular user.
func (c *Client) DemoteAdmin(ctx context.Context, params types.DemoteAdminParams) (*types.DemoteAdminResult, error) {
	return c.chat.DemoteAdmin(ctx, params)
}

// GetInviteLink gets or creates an invite link for a chat or channel.
func (c *Client) GetInviteLink(ctx context.Context, params types.GetInviteLinkParams) (*types.GetInviteLinkResult, error) {
	return c.chat.GetInviteLink(ctx, params)
}
