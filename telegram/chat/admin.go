package chat

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// PromoteAdmin promotes a user to admin.
//
//nolint:funlen // Function requires complex admin permission handling
func (c *Client) PromoteAdmin(ctx context.Context, params types.PromoteAdminParams) (*types.PromoteAdminResult, error) {
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

	// Resolve user to promote
	userPeer, err := c.ResolvePeer(ctx, params.User)
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

	_, err = c.API.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
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

	// Resolve user to demote
	userPeer, err := c.ResolvePeer(ctx, params.User)
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

	_, err = c.API.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
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
