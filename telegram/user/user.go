// Package user provides Telegram user info operations.
package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/helpers"
	"agent-telegram/telegram/types"
)

// resolvedUser holds a resolved user with its access hash.
type resolvedUser struct {
	user       *tg.User
	accessHash int64
}

// GetUserInfo gets information about a user by username or numeric ID.
func (c *Client) GetUserInfo(ctx context.Context, params types.GetUserInfoParams) (*types.GetUserInfoResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	var resolved resolvedUser
	var err error

	if params.UserID != 0 {
		resolved, err = c.resolveUserByNumericID(ctx, params.UserID)
	} else {
		resolved, err = c.resolveUserByUsername(ctx, params.Username)
	}
	if err != nil {
		return nil, err
	}

	// Get full user info to get bio
	fullUser, err := c.API.UsersGetFullUser(ctx, &tg.InputUser{
		UserID:     resolved.user.ID,
		AccessHash: resolved.accessHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get full user info: %w", err)
	}

	return &types.GetUserInfoResult{
		ID:        resolved.user.ID,
		Username:  resolved.user.Username,
		FirstName: resolved.user.FirstName,
		LastName:  resolved.user.LastName,
		Phone:     resolved.user.Phone,
		Bio:       fullUser.FullUser.About,
		Verified:  resolved.user.Verified,
		Bot:       resolved.user.Bot,
	}, nil
}

// resolveUserByNumericID resolves a user by numeric ID via dialogs + UsersGetUsers.
func (c *Client) resolveUserByNumericID(ctx context.Context, userID int64) (resolvedUser, error) {
	inputPeer, err := c.ResolvePeer(ctx, fmt.Sprintf("%d", userID))
	if err != nil {
		return resolvedUser{}, fmt.Errorf("failed to resolve user %d: %w", userID, err)
	}
	peerUser, ok := inputPeer.(*tg.InputPeerUser)
	if !ok {
		return resolvedUser{}, fmt.Errorf("peer %d is not a user", userID)
	}

	users, err := c.API.UsersGetUsers(ctx, []tg.InputUserClass{
		&tg.InputUser{UserID: userID, AccessHash: peerUser.AccessHash},
	})
	if err != nil {
		return resolvedUser{}, fmt.Errorf("failed to get user %d: %w", userID, err)
	}
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			return resolvedUser{user: user, accessHash: peerUser.AccessHash}, nil
		}
	}
	return resolvedUser{}, fmt.Errorf("user %d not found", userID)
}

// resolveUserByUsername resolves a user by username via ContactsResolveUsername.
func (c *Client) resolveUserByUsername(ctx context.Context, username string) (resolvedUser, error) {
	username = trimUsernamePrefix(username)
	resolvedPeer, err := c.API.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return resolvedUser{}, fmt.Errorf("failed to resolve user @%s: %w", username, err)
	}

	for _, u := range resolvedPeer.Users {
		if user, ok := u.(*tg.User); ok {
			hash := helpers.GetAccessHash(resolvedPeer, user.ID)
			return resolvedUser{user: user, accessHash: hash}, nil
		}
	}
	return resolvedUser{}, fmt.Errorf("user @%s not found", username)
}
