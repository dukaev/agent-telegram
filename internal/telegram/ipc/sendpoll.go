// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"

	"agent-telegram/telegram/types"
)

// SendPollHandler returns a handler for send_poll requests.
func SendPollHandler(client Client) HandlerFunc {
	return Handler(client.Media().SendPoll, "send poll")
}

// SendChecklistHandler returns a handler for send_checklist (quiz) requests.
func SendChecklistHandler(client Client) HandlerFunc {
	return Handler(func(ctx context.Context, p types.SendPollParams) (*types.SendPollResult, error) {
		p.Quiz = true
		p.Anonymous = false
		return client.Media().SendPoll(ctx, p)
	}, "send checklist")
}
