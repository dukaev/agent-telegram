// Package ipc provides Telegram IPC handlers for chat admin operations.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram/types"
)

// GetAdminsHandler returns a handler for get_admins requests.
func GetAdminsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetAdminsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.GetAdmins(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get admins: %w", err)
		}

		return result, nil
	}
}

// GetBannedHandler returns a handler for get_banned requests.
func GetBannedHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetBannedParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.GetBanned(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get banned users: %w", err)
		}

		return result, nil
	}
}

// PromoteAdminHandler returns a handler for promote_admin requests.
func PromoteAdminHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.PromoteAdminParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.PromoteAdmin(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to promote admin: %w", err)
		}

		return result, nil
	}
}

// DemoteAdminHandler returns a handler for demote_admin requests.
func DemoteAdminHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.DemoteAdminParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.DemoteAdmin(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to demote admin: %w", err)
		}

		return result, nil
	}
}

// GetInviteLinkHandler returns a handler for get_invite_link requests.
func GetInviteLinkHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetInviteLinkParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.GetInviteLink(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get invite link: %w", err)
		}

		return result, nil
	}
}
