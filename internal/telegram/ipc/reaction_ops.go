// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// AddReactionHandler returns a handler for add_reaction requests.
func AddReactionHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.AddReaction, "add reaction")
}

// RemoveReactionHandler returns a handler for remove_reaction requests.
func RemoveReactionHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.RemoveReaction, "remove reaction")
}

// ListReactionsHandler returns a handler for list_reactions requests.
func ListReactionsHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.ListReactions, "list reactions")
}
