package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// GetParticipants retrieves participants from a chat or channel.
//
//nolint:funlen // Function requires handling multiple peer types
func (c *Client) GetParticipants(
	ctx context.Context,
	params types.GetParticipantsParams,
) (*types.GetParticipantsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
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
		channelResult, err := c.API.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Filter: &tg.ChannelParticipantsRecent{},
			Limit:  limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get channel participants: %w", err)
		}

		// Extract participants
		if r, ok := channelResult.(*tg.ChannelsChannelParticipants); ok {
			for _, participantClass := range r.Participants {
				result.Participants = append(result.Participants, extractParticipant(participantClass, r.Users))
			}
			result.Count = r.Count
		}

	case *tg.InputPeerChat:
		// For chats, use MessagesGetFullChat
		fullChat, err := c.API.MessagesGetFullChat(ctx, p.ChatID)
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
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	participantsResult, err := c.GetParticipants(ctx, types.GetParticipantsParams(params))
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
		Peer:   params.Peer,
		Admins: admins,
		Count:  len(admins),
	}, nil
}

// GetBanned retrieves banned users from a chat or channel.
func (c *Client) GetBanned(ctx context.Context, params types.GetBannedParams) (*types.GetBannedResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	peer, err := c.ResolvePeer(ctx, params.Peer)
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
	channelResult, err := c.API.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		Filter: &tg.ChannelParticipantsKicked{},
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get banned users: %w", err)
	}

	result := &types.GetBannedResult{
		Peer:   params.Peer,
		Banned: []types.Participant{},
	}

	// Extract banned participants
	if r, ok := channelResult.(*tg.ChannelsChannelParticipants); ok {
		for _, participantClass := range r.Participants {
			result.Banned = append(result.Banned, extractParticipant(participantClass, r.Users))
		}
		result.Count = len(result.Banned)
	}

	return result, nil
}
