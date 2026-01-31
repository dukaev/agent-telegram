// Package telegram provides Telegram client user functionality.
package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// GetUserInfo gets information about a user by username.
func (c *Client) GetUserInfo(ctx context.Context, params GetUserInfoParams) (*GetUserInfoResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	username := params.Username
	username = trimUsernamePrefix(username)

	resolvedPeer, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
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
	fullUser, err := api.UsersGetFullUser(ctx, &tg.InputUser{
		UserID:     user.ID,
		AccessHash: getAccessHash(resolvedPeer, user.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get full user info: %w", err)
	}

	bio := fullUser.FullUser.About

	return &GetUserInfoResult{
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

// trimUsernamePrefix removes the @ prefix from username.
func trimUsernamePrefix(username string) string {
	if len(username) > 0 && username[0] == '@' {
		return username[1:]
	}
	return username
}
