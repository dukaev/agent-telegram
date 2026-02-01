// Package user provides privacy settings operations.
package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// privacyKeyMap maps string keys to InputPrivacyKey types.
var privacyKeyMap = map[string]tg.InputPrivacyKeyClass{
	"status_timestamp": &tg.InputPrivacyKeyStatusTimestamp{},
	"chat_invite":      &tg.InputPrivacyKeyChatInvite{},
	"phone_call":       &tg.InputPrivacyKeyPhoneCall{},
	"phone_p2p":        &tg.InputPrivacyKeyPhoneP2P{},
	"forwards":         &tg.InputPrivacyKeyForwards{},
	"profile_photo":    &tg.InputPrivacyKeyProfilePhoto{},
	"phone_number":     &tg.InputPrivacyKeyPhoneNumber{},
	"added_by_phone":   &tg.InputPrivacyKeyAddedByPhone{},
	"voice_messages":   &tg.InputPrivacyKeyVoiceMessages{},
	"about":            &tg.InputPrivacyKeyAbout{},
}

// GetPrivacy retrieves privacy settings for a specific key.
func (c *Client) GetPrivacy(ctx context.Context, params types.GetPrivacyParams) (*types.GetPrivacyResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	key, ok := privacyKeyMap[params.Key]
	if !ok {
		return nil, fmt.Errorf("unknown privacy key: %s", params.Key)
	}

	result, err := c.API.AccountGetPrivacy(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy: %w", err)
	}

	rules := make([]types.PrivacyRule, 0)
	for _, rule := range result.Rules {
		r := types.PrivacyRule{}
		switch v := rule.(type) {
		case *tg.PrivacyValueAllowAll:
			r.Type = "allow_all"
		case *tg.PrivacyValueAllowContacts:
			r.Type = "allow_contacts"
		case *tg.PrivacyValueAllowUsers:
			r.Type = "allow_users"
			r.Users = v.Users
		case *tg.PrivacyValueAllowChatParticipants:
			r.Type = "allow_chats"
			r.Chats = v.Chats
		case *tg.PrivacyValueDisallowAll:
			r.Type = "disallow_all"
		case *tg.PrivacyValueDisallowContacts:
			r.Type = "disallow_contacts"
		case *tg.PrivacyValueDisallowUsers:
			r.Type = "disallow_users"
			r.Users = v.Users
		case *tg.PrivacyValueDisallowChatParticipants:
			r.Type = "disallow_chats"
			r.Chats = v.Chats
		case *tg.PrivacyValueAllowCloseFriends:
			r.Type = "allow_close_friends"
		default:
			continue
		}
		rules = append(rules, r)
	}

	return &types.GetPrivacyResult{
		Key:   params.Key,
		Rules: rules,
	}, nil
}

// SetPrivacy sets privacy settings for a specific key.
func (c *Client) SetPrivacy(ctx context.Context, params types.SetPrivacyParams) (*types.SetPrivacyResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	key, ok := privacyKeyMap[params.Key]
	if !ok {
		return nil, fmt.Errorf("unknown privacy key: %s", params.Key)
	}

	rules := make([]tg.InputPrivacyRuleClass, 0)
	for _, r := range params.Rules {
		var rule tg.InputPrivacyRuleClass
		switch r.Type {
		case "allow_all":
			rule = &tg.InputPrivacyValueAllowAll{}
		case "allow_contacts":
			rule = &tg.InputPrivacyValueAllowContacts{}
		case "allow_users":
			users := make([]tg.InputUserClass, 0)
			for _, uid := range r.Users {
				users = append(users, &tg.InputUser{UserID: uid})
			}
			rule = &tg.InputPrivacyValueAllowUsers{Users: users}
		case "allow_chats":
			rule = &tg.InputPrivacyValueAllowChatParticipants{Chats: r.Chats}
		case "disallow_all":
			rule = &tg.InputPrivacyValueDisallowAll{}
		case "disallow_contacts":
			rule = &tg.InputPrivacyValueDisallowContacts{}
		case "disallow_users":
			users := make([]tg.InputUserClass, 0)
			for _, uid := range r.Users {
				users = append(users, &tg.InputUser{UserID: uid})
			}
			rule = &tg.InputPrivacyValueDisallowUsers{Users: users}
		case "disallow_chats":
			rule = &tg.InputPrivacyValueDisallowChatParticipants{Chats: r.Chats}
		case "allow_close_friends":
			rule = &tg.InputPrivacyValueAllowCloseFriends{}
		default:
			return nil, fmt.Errorf("unknown rule type: %s", r.Type)
		}
		rules = append(rules, rule)
	}

	_, err := c.API.AccountSetPrivacy(ctx, &tg.AccountSetPrivacyRequest{
		Key:   key,
		Rules: rules,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set privacy: %w", err)
	}

	return &types.SetPrivacyResult{Success: true}, nil
}
