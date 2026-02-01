// Package ipc provides Telegram IPC handlers.
package ipc

// AddReactionHandler returns a handler for add_reaction requests.
func AddReactionHandler(client Client) HandlerFunc {
	return Handler(client.Reaction().AddReaction, "add reaction")
}

// RemoveReactionHandler returns a handler for remove_reaction requests.
func RemoveReactionHandler(client Client) HandlerFunc {
	return Handler(client.Reaction().RemoveReaction, "remove reaction")
}

// ListReactionsHandler returns a handler for list_reactions requests.
func ListReactionsHandler(client Client) HandlerFunc {
	return Handler(client.Reaction().ListReactions, "list reactions")
}
