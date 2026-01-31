// Package user provides Telegram user info operations.
package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// getAccessHash extracts access hash from the resolved peer.
func getAccessHash(peerClass *tg.ContactsResolvedPeer, id int64) int64 {
	for _, chat := range peerClass.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == id {
				return c.AccessHash
			}
		case *tg.Chat:
			if c.ID == id {
				return 0
			}
		}
	}
	for _, user := range peerClass.Users {
		if u, ok := user.(*tg.User); ok && u.ID == id {
			return u.AccessHash
		}
	}
	return 0
}

// GetUserInfo gets information about a user by username.
func (c *Client) GetUserInfo(ctx context.Context, params types.GetUserInfoParams) (*types.GetUserInfoResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	username := trimUsernamePrefix(params.Username)

	resolvedPeer, err := c.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user @%s: %w", username, err)
	}

	var user *tg.User
	for _, u := range resolvedPeer.Users {
		if u, ok := u.(*tg.User); ok {
			user = u
			break
		}
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get full user info to get bio
	fullUser, err := c.api.UsersGetFullUser(ctx, &tg.InputUser{
		UserID:     user.ID,
		AccessHash: getAccessHash(resolvedPeer, user.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get full user info: %w", err)
	}

	bio := fullUser.FullUser.About

	return &types.GetUserInfoResult{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Bio:       bio,
		Verified:  user.Verified,
		Bot:       user.Bot,
	}, nil
}
