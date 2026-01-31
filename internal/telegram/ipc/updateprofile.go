// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

		"agent-telegram/telegram/types"
)

// UpdateProfileParams represents parameters for update_profile request.
type UpdateProfileParams struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
	Bio       string `json:"bio,omitempty"`
}

// UpdateProfileHandler returns a handler for update_profile requests.
func UpdateProfileHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p UpdateProfileParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.FirstName == "" {
			return nil, fmt.Errorf("firstName is required")
		}

		result, err := client.UpdateProfile(context.Background(), types.UpdateProfileParams{
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Bio:       p.Bio,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}

		return result, nil
	}
}

// UpdateAvatarParams represents parameters for update_avatar request.
type UpdateAvatarParams struct {
	File string `json:"file"`
}

// UpdateAvatarHandler returns a handler for update_avatar requests.
func UpdateAvatarHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p UpdateAvatarParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.File == "" {
			return nil, fmt.Errorf("file is required")
		}

		// Check file exists
		if _, err := os.Stat(p.File); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", p.File)
		}

		result, err := client.UpdateAvatar(context.Background(), types.UpdateAvatarParams{
			File: p.File,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update avatar: %w", err)
		}

		return result, nil
	}
}
