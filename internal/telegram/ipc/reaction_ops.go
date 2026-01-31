// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// AddReactionHandler returns a handler for add_reaction requests.
func AddReactionHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.AddReactionParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.AddReaction(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to add reaction: %w", err)
		}

		return result, nil
	}
}

// RemoveReactionHandler returns a handler for remove_reaction requests.
func RemoveReactionHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.RemoveReactionParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.RemoveReaction(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to remove reaction: %w", err)
		}

		return result, nil
	}
}

// ListReactionsHandler returns a handler for list_reactions requests.
func ListReactionsHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.ListReactionsParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.ListReactions(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to list reactions: %w", err)
		}

		return result, nil
	}
}
