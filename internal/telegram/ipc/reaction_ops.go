// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// AddReactionParams represents parameters for add_reaction request.
type AddReactionParams struct {
	Peer      string `json:"peer"`
	MessageID int64  `json:"messageId"`
	Emoji     string `json:"emoji"`
	Big       bool   `json:"big,omitempty"`
}

// AddReactionHandler returns a handler for add_reaction requests.
func AddReactionHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p AddReactionParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.MessageID == 0 {
			return nil, fmt.Errorf("messageId is required")
		}
		if p.Emoji == "" {
			return nil, fmt.Errorf("emoji is required")
		}

		result, err := client.AddReaction(context.Background(), telegram.AddReactionParams{
			Peer:      p.Peer,
			MessageID: p.MessageID,
			Emoji:     p.Emoji,
			Big:       p.Big,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add reaction: %w", err)
		}

		return result, nil
	}
}

// RemoveReactionParams represents parameters for remove_reaction request.
type RemoveReactionParams struct {
	Peer      string `json:"peer"`
	MessageID int64  `json:"messageId"`
}

// RemoveReactionHandler returns a handler for remove_reaction requests.
func RemoveReactionHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p RemoveReactionParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.MessageID == 0 {
			return nil, fmt.Errorf("messageId is required")
		}

		result, err := client.RemoveReaction(context.Background(), telegram.RemoveReactionParams{
			Peer:      p.Peer,
			MessageID: p.MessageID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to remove reaction: %w", err)
		}

		return result, nil
	}
}

// ListReactionsParams represents parameters for list_reactions request.
type ListReactionsParams struct {
	Peer      string `json:"peer"`
	MessageID int64  `json:"messageId"`
	Limit     int    `json:"limit,omitempty"`
}

// ListReactionsHandler returns a handler for list_reactions requests.
func ListReactionsHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p ListReactionsParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.MessageID == 0 {
			return nil, fmt.Errorf("messageId is required")
		}

		result, err := client.ListReactions(context.Background(), telegram.ListReactionsParams{
			Peer:      p.Peer,
			MessageID: p.MessageID,
			Limit:     p.Limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list reactions: %w", err)
		}

		return result, nil
	}
}
