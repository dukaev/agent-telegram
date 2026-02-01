// Package chat provides Telegram chat operations.
package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// Client provides chat operations.
type Client struct {
	api    *tg.Client
	parent ParentClient
}

// ParentClient is an interface for accessing parent client methods.
type ParentClient interface {
	ResolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error)
}

// NewClient creates a new chat client.
func NewClient(tc ParentClient) *Client {
	return &Client{
		parent: tc,
	}
}

// SetAPI sets the API client (called when the telegram client is initialized).
func (c *Client) SetAPI(api *tg.Client) {
	c.api = api
}

// resolvePeer resolves a peer string to InputPeerClass using the parent client's cache.
func (c *Client) resolvePeer(ctx context.Context, peer string) (tg.InputPeerClass, error) {
	if c.parent == nil {
		return nil, fmt.Errorf("parent client not set")
	}
	return c.parent.ResolvePeer(ctx, peer)
}

// Join joins a chat or channel using an invite link.
func (c *Client) Join(ctx context.Context, inviteLink string) (tg.UpdatesClass, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Extract hash from invite link
	hash, err := extractInviteHash(inviteLink)
	if err != nil {
		return nil, fmt.Errorf("invalid invite link: %w", err)
	}

	// Use messages.ImportChatInvite to join
	result, err := c.api.MessagesImportChatInvite(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to join chat: %w", err)
	}

	return result, nil
}

// extractInviteHash extracts the hash from various invite link formats.
func extractInviteHash(link string) (string, error) {
	// Common patterns:
	// https://t.me/+hash
	// https://t.me/joinchat/hash
	// tg://join?invite=hash
	// +hash
	// hash

	// Trim whitespace
	link = trimSpace(link)
	if link == "" {
		return "", fmt.Errorf("empty invite link")
	}

	// Remove common prefixes
	prefixes := []string{
		"https://t.me/+",
		"https://t.me/joinchat/",
		"tg://join?invite=",
		"+",
	}

	for _, prefix := range prefixes {
		if len(link) > len(prefix) && link[:len(prefix)] == prefix {
			return link[len(prefix):], nil
		}
	}

	// Assume the link is already a hash
	return link, nil
}

func trimSpace(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// JoinChat joins a chat or channel using an invite link.
func (c *Client) JoinChat(ctx context.Context, params types.JoinChatParams) (*types.JoinChatResult, error) {
	_, err := c.Join(ctx, params.InviteLink)
	if err != nil {
		return nil, err
	}

	return &types.JoinChatResult{
		Success: true,
	}, nil
}

// Subscribe subscribes to a channel by username.
func (c *Client) Subscribe(ctx context.Context, channel string) (tg.UpdatesClass, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve peer to get InputChannel
	peer, err := c.resolvePeer(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve channel: %w", err)
	}

	var inputChannel *tg.InputChannel
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		inputChannel = &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, fmt.Errorf("not a channel: %s", channel)
	}

	// Join the channel
	result, err := c.api.ChannelsJoinChannel(ctx, inputChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to channel: %w", err)
	}

	return result, nil
}

// SubscribeChannel subscribes to a channel by username.
func (c *Client) SubscribeChannel(ctx context.Context, params types.SubscribeChannelParams) (*types.SubscribeChannelResult, error) {
	updates, err := c.Subscribe(ctx, params.Channel)
	if err != nil {
		return nil, err
	}

	result := &types.SubscribeChannelResult{
		Success: true,
	}

	// Extract channel info from chats
	switch u := updates.(type) {
	case *tg.Updates:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				result.ChatID = ch.ID
				result.Title = ch.Title
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range u.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				result.ChatID = ch.ID
				result.Title = ch.Title
			}
		}
	}

	return result, nil
}

// GetTopics retrieves forum topics from a channel.
func (c *Client) GetTopics(ctx context.Context, params types.GetTopicsParams) (*types.GetTopicsResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve peer to get InputChannel
	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Verify it's a channel
	switch peer.(type) {
	case *tg.InputPeerChannel:
		// OK, it's a channel
	default:
		return nil, fmt.Errorf("not a channel: %s", params.Peer)
	}

	// Set limit
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	// Get forum topics using MessagesGetForumTopics
	result, err := c.api.MessagesGetForumTopics(ctx, &tg.MessagesGetForumTopicsRequest{
		Peer:  peer,
		Limit: int(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get forum topics: %w", err)
	}

	topicsResult := &types.GetTopicsResult{
		Peer:   params.Peer,
		Topics: []types.ForumTopic{},
	}

	// Extract topics from result
	for _, topicClass := range result.Topics {
		if topic, ok := topicClass.(*tg.ForumTopic); ok {
			topicsResult.Topics = append(topicsResult.Topics, types.ForumTopic{
				ID:        int64(topic.ID),
				Title:     topic.Title,
				IconColor: int32(topic.IconColor),
				Top:       topic.Pinned,
				Closed:    topic.Closed,
			})
		}
	}
	topicsResult.Count = len(topicsResult.Topics)

	return topicsResult, nil
}

// CreateGroup creates a new group chat.
func (c *Client) CreateGroup(ctx context.Context, params types.CreateGroupParams) (*types.CreateGroupResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Resolve members to InputPeerClass
	var users []tg.InputUserClass
	for _, member := range params.Members {
		peer, err := c.resolvePeer(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %s: %w", member, err)
		}

		switch p := peer.(type) {
		case *tg.InputPeerUser:
			users = append(users, &tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			})
		default:
			return nil, fmt.Errorf("member %s is not a user", member)
		}
	}

	// Create group
	result, err := c.api.MessagesCreateChat(ctx, &tg.MessagesCreateChatRequest{
		Users:     users,
		Title:     params.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	groupResult := &types.CreateGroupResult{
		Success: true,
		Title:   params.Title,
	}

	// Extract chat ID from result.Updates
	switch r := result.Updates.(type) {
	case *tg.Updates:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Chat); ok {
				groupResult.ChatID = ch.ID
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Chat); ok {
				groupResult.ChatID = ch.ID
			}
		}
	}

	return groupResult, nil
}

// CreateChannel creates a new channel or supergroup.
func (c *Client) CreateChannel(ctx context.Context, params types.CreateChannelParams) (*types.CreateChannelResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// Create channel
	result, err := c.api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Title:     params.Title,
		About:     params.Description,
		ForImport: false,
		Megagroup: params.Megagroup,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	channelResult := &types.CreateChannelResult{
		Success: true,
		Title:   params.Title,
	}

	// Extract chat ID from result
	var inputChannel *tg.InputChannel
	switch r := result.(type) {
	case *tg.Updates:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				channelResult.ChatID = ch.ID
				inputChannel = &tg.InputChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				}
			}
		}
	case *tg.UpdatesCombined:
		for _, chat := range r.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				channelResult.ChatID = ch.ID
				inputChannel = &tg.InputChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				}
			}
		}
	}

	// Set username if provided
	if params.Username != "" && inputChannel != nil {
		_, err = c.api.ChannelsUpdateUsername(ctx, &tg.ChannelsUpdateUsernameRequest{
			Channel:  inputChannel,
			Username: params.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set username: %w", err)
		}
	}

	return channelResult, nil
}

// EditTitle edits the title of a chat or channel.
func (c *Client) EditTitle(ctx context.Context, params types.EditTitleParams) (*types.EditTitleResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.api.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Title: params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit channel title: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.api.MessagesEditChatTitle(ctx, &tg.MessagesEditChatTitleRequest{
			ChatID: p.ChatID,
			Title:  params.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to edit chat title: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.EditTitleResult{
		Success: true,
		Title:   params.Title,
	}, nil
}

// SetPhoto sets the photo for a chat or channel.
func (c *Client) SetPhoto(ctx context.Context, params types.SetPhotoParams) (*types.SetPhotoResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	// This requires file upload functionality
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("set_photo requires file upload - use send_photo command or implement file upload")
}

// DeletePhoto deletes the photo from a chat or channel.
func (c *Client) DeletePhoto(ctx context.Context, params types.DeletePhotoParams) (*types.DeletePhotoResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.api.ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Photo: &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete channel photo: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.api.MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
			ChatID: p.ChatID,
			Photo:  &tg.InputChatPhotoEmpty{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete chat photo: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.DeletePhotoResult{Success: true}, nil
}

// Leave leaves a chat or channel.
func (c *Client) Leave(ctx context.Context, params types.LeaveParams) (*types.LeaveResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		_, err := c.api.ChannelsLeaveChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to leave channel: %w", err)
		}
	case *tg.InputPeerChat:
		_, err := c.api.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID: p.ChatID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to leave chat: %w", err)
		}
	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &types.LeaveResult{Success: true}, nil
}

// Invite invites users to a chat or channel.
func (c *Client) Invite(ctx context.Context, params types.InviteParams) (*types.InviteResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	// Resolve all members to add
	for _, member := range params.Members {
		userPeer, err := c.resolvePeer(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve member %s: %w", member, err)
		}

		var inputUser *tg.InputUser
		switch p := userPeer.(type) {
		case *tg.InputPeerUser:
			inputUser = &tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			}
		default:
			return nil, fmt.Errorf("member %s is not a user", member)
		}

		switch peer.(type) {
		case *tg.InputPeerChannel:
			_, err := c.api.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
				Channel: &tg.InputChannel{
					ChannelID:  peer.(*tg.InputPeerChannel).ChannelID,
					AccessHash: peer.(*tg.InputPeerChannel).AccessHash,
				},
				Users: []tg.InputUserClass{inputUser},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to invite to channel: %w", err)
			}
		case *tg.InputPeerChat:
			_, err := c.api.MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{
				ChatID: peer.(*tg.InputPeerChat).ChatID,
				UserID: inputUser,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to invite to chat: %w", err)
			}
		}
	}

	return &types.InviteResult{Success: true}, nil
}

// GetParticipants retrieves participants from a chat or channel.
func (c *Client) GetParticipants(ctx context.Context, params types.GetParticipantsParams) (*types.GetParticipantsResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	limit := params.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	var result types.GetParticipantsResult
	result.Peer = params.Peer
	result.Participants = []types.Participant{}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		// For channels, use ChannelsGetParticipants
		channelResult, err := c.api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Limit: int(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get channel participants: %w", err)
		}

		// Extract participants
		switch r := channelResult.(type) {
		case *tg.ChannelsChannelParticipants:
			for _, participantClass := range r.Participants {
				result.Participants = append(result.Participants, extractParticipant(participantClass, r.Users))
			}
			result.Count = r.Count
		}

	case *tg.InputPeerChat:
		// For chats, use MessagesGetFullChat
		fullChat, err := c.api.MessagesGetFullChat(ctx, p.ChatID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat info: %w", err)
		}

		if chatFull, ok := fullChat.FullChat.(*tg.ChatFull); ok {
			if participants, ok := chatFull.Participants.(*tg.ChatParticipants); ok {
				for _, participantClass := range participants.Participants {
					result.Participants = append(result.Participants, extractChatParticipant(participantClass, fullChat.Users))
				}
				result.Count = len(result.Participants)
			}
		}

	default:
		return nil, fmt.Errorf("peer must be a chat or channel")
	}

	return &result, nil
}

// extractParticipant extracts participant info from ChannelParticipantClass.
func extractParticipant(p tg.ChannelParticipantClass, users []tg.UserClass) types.Participant {
	var userID int64
	var isCreator, isAdmin bool

	switch participant := p.(type) {
	case *tg.ChannelParticipantCreator:
		userID = participant.UserID
		isCreator = true
		isAdmin = true
	case *tg.ChannelParticipantAdmin:
		userID = participant.UserID
		isAdmin = true
	case *tg.ChannelParticipant:
		userID = participant.UserID
	case *tg.ChannelParticipantSelf:
		userID = participant.UserID
	}

	// Find user info
	for _, userClass := range users {
		if user, ok := userClass.(*tg.User); ok && user.ID == userID {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			return types.Participant{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
				Bot:       user.Bot,
				Creator:   isCreator,
				Admin:     isAdmin,
				Peer:      formatUserPeer(user.ID, user.Username),
			}
		}
	}

	return types.Participant{ID: userID}
}

// formatUserPeer formats a user peer string.
func formatUserPeer(id int64, username string) string {
	if username != "" {
		return "@" + username
	}
	return fmt.Sprintf("user:%d", id)
}

// extractChatParticipant extracts participant info from ChatParticipantClass.
func extractChatParticipant(p tg.ChatParticipantClass, users []tg.UserClass) types.Participant {
	var userID int64
	var isCreator, isAdmin bool

	switch participant := p.(type) {
	case *tg.ChatParticipantCreator:
		userID = participant.UserID
		isCreator = true
		isAdmin = true
	case *tg.ChatParticipantAdmin:
		userID = participant.UserID
		isAdmin = true
	case *tg.ChatParticipant:
		userID = participant.UserID
	}

	// Find user info
	for _, userClass := range users {
		if user, ok := userClass.(*tg.User); ok && user.ID == userID {
			name := user.FirstName
			if user.LastName != "" {
				name += " " + user.LastName
			}
			return types.Participant{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
				Bot:       user.Bot,
				Creator:   isCreator,
				Admin:     isAdmin,
				Peer:      formatUserPeer(user.ID, user.Username),
			}
		}
	}

	return types.Participant{ID: userID}
}

// GetAdmins retrieves admins from a chat or channel.
func (c *Client) GetAdmins(ctx context.Context, params types.GetAdminsParams) (*types.GetAdminsResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	participantsResult, err := c.GetParticipants(ctx, types.GetParticipantsParams{
		Peer:  params.Peer,
		Limit: params.Limit,
	})
	if err != nil {
		return nil, err
	}

	var admins []types.Participant
	for _, p := range participantsResult.Participants {
		if p.Admin || p.Creator {
			admins = append(admins, p)
		}
	}

	return &types.GetAdminsResult{
		Peer:  params.Peer,
		Admins: admins,
		Count: len(admins),
	}, nil
}

// GetBanned retrieves banned users from a chat or channel.
func (c *Client) GetBanned(ctx context.Context, params types.GetBannedParams) (*types.GetBannedResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	inputChannel, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("peer must be a channel")
	}

	limit := params.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	// Use ChannelsGetParticipants with ChannelParticipantsKicked filter
	channelResult, err := c.api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		Filter: &tg.ChannelParticipantsKicked{},
		Limit:  int(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get banned users: %w", err)
	}

	result := &types.GetBannedResult{
		Peer:   params.Peer,
		Banned: []types.Participant{},
	}

	// Extract banned participants
	switch r := channelResult.(type) {
	case *tg.ChannelsChannelParticipants:
		for _, participantClass := range r.Participants {
			result.Banned = append(result.Banned, extractParticipant(participantClass, r.Users))
		}
		result.Count = len(result.Banned)
	}

	return result, nil
}

// PromoteAdmin promotes a user to admin.
func (c *Client) PromoteAdmin(ctx context.Context, params types.PromoteAdminParams) (*types.PromoteAdminResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	inputChannel, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("peer must be a channel")
	}

	// Resolve user to promote
	userPeer, err := c.resolvePeer(ctx, params.User)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	var inputUser *tg.InputUser
	switch p := userPeer.(type) {
	case *tg.InputPeerUser:
		inputUser = &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, fmt.Errorf("user must be a user")
	}

	// Build admin rights
	rights := &tg.ChatAdminRights{}

	if params.CanChangeInfo {
		rights.ChangeInfo = true
	}
	if params.CanPostMessages {
		rights.PostMessages = true
	}
	if params.CanEditMessages {
		rights.EditMessages = true
	}
	if params.CanDeleteMessages {
		rights.DeleteMessages = true
	}
	if params.CanBanUsers {
		rights.BanUsers = true
	}
	if params.CanInviteUsers {
		rights.InviteUsers = true
	}
	if params.CanPinMessages {
		rights.PinMessages = true
	}
	if params.CanAddAdmins {
		rights.AddAdmins = true
	}
	if params.Anonymous {
		rights.Anonymous = true
	}

	_, err = c.api.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		UserID:      inputUser,
		AdminRights: *rights,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to promote admin: %w", err)
	}

	return &types.PromoteAdminResult{Success: true}, nil
}

// DemoteAdmin demotes an admin to regular user.
func (c *Client) DemoteAdmin(ctx context.Context, params types.DemoteAdminParams) (*types.DemoteAdminResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	inputChannel, ok := peer.(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("peer must be a channel")
	}

	// Resolve user to demote
	userPeer, err := c.resolvePeer(ctx, params.User)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	var inputUser *tg.InputUser
	switch p := userPeer.(type) {
	case *tg.InputPeerUser:
		inputUser = &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, fmt.Errorf("user must be a user")
	}

	// Set empty admin rights (demote)
	rights := &tg.ChatAdminRights{}

	_, err = c.api.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		UserID:      inputUser,
		AdminRights: *rights,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to demote admin: %w", err)
	}

	return &types.DemoteAdminResult{Success: true}, nil
}

// GetInviteLink gets or creates an invite link for a chat or channel.
func (c *Client) GetInviteLink(ctx context.Context, params types.GetInviteLinkParams) (*types.GetInviteLinkResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("api client not set")
	}

	peer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		var result tg.ExportedChatInviteClass
		if params.CreateNew {
			// Create new invite link
			result, err = c.api.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
				Peer: &tg.InputPeerChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
			})
		} else {
			// Try to get existing invite link
			fullInfo, err := c.api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get channel info: %w", err)
			}

			if fullCh, ok := fullInfo.FullChat.(*tg.ChannelFull); ok {
				if fullCh.ExportedInvite != nil {
					// Check if it's a ChatInviteExported type
					if invite, ok := fullCh.ExportedInvite.(*tg.ChatInviteExported); ok {
						return &types.GetInviteLinkResult{
							Link:           invite.Link,
							Usage:          invite.Usage,
							UsageLimit:     invite.UsageLimit,
							RequestNeeded:  invite.RequestNeeded,
							Expired:        invite.Revoked,
						}, nil
					}
				}
			}
			// No existing link, create one
			result, err = c.api.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
				Peer: &tg.InputPeerChannel{
					ChannelID:  p.ChannelID,
					AccessHash: p.AccessHash,
				},
			})
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get invite link: %w", err)
		}

		switch r := result.(type) {
		case *tg.ChatInviteExported:
			return &types.GetInviteLinkResult{
				Link:           r.Link,
				Usage:          r.Usage,
				UsageLimit:     r.UsageLimit,
				RequestNeeded:  r.RequestNeeded,
				Expired:        r.Revoked,
			}, nil
		}
	case *tg.InputPeerChat:
		// For chats, use MessagesExportChatInvite
		result, err := c.api.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
			Peer: p,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get invite link: %w", err)
		}

		switch r := result.(type) {
		case *tg.ChatInviteExported:
			return &types.GetInviteLinkResult{
				Link:       r.Link,
				Usage:      r.Usage,
				UsageLimit: r.UsageLimit,
				Expired:    false,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to get invite link")
}
