// Package ipc provides Telegram IPC handlers for chat operations.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram/types"
)

// CreateGroupHandler returns a handler for create_group requests.
func CreateGroupHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.CreateGroupParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.CreateGroup(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to create group: %w", err)
		}

		return result, nil
	}
}

// CreateChannelHandler returns a handler for create_channel requests.
func CreateChannelHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.CreateChannelParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.CreateChannel(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel: %w", err)
		}

		return result, nil
	}
}

// EditTitleHandler returns a handler for edit_title requests.
func EditTitleHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.EditTitleParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.EditTitle(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to edit title: %w", err)
		}

		return result, nil
	}
}

// SetPhotoHandler returns a handler for set_photo requests.
func SetPhotoHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.SetPhotoParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.SetPhoto(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to set photo: %w", err)
		}

		return result, nil
	}
}

// DeletePhotoHandler returns a handler for delete_photo requests.
func DeletePhotoHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.DeletePhotoParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.DeletePhoto(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to delete photo: %w", err)
		}

		return result, nil
	}
}

// LeaveHandler returns a handler for leave requests.
func LeaveHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.LeaveParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.Leave(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to leave chat: %w", err)
		}

		return result, nil
	}
}

// InviteHandler returns a handler for invite requests.
func InviteHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.InviteParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.Invite(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to invite users: %w", err)
		}

		return result, nil
	}
}

// GetParticipantsHandler returns a handler for get_participants requests.
func GetParticipantsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p types.GetParticipantsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.GetParticipants(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to get participants: %w", err)
		}

		return result, nil
	}
}

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
