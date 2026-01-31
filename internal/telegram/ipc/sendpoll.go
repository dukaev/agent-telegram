// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"

		"agent-telegram/telegram/types"
)

// SendPollHandler returns a handler for send_poll requests.
func SendPollHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SendPoll, "send poll")
}

// SendChecklistHandler returns a handler for send_checklist (quiz) requests.
func SendChecklistHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(func(ctx context.Context, p types.SendPollParams) (*types.SendPollResult, error) {
		p.Quiz = true
		p.Anonymous = false
		return client.SendPoll(ctx, p)
	}, "send checklist")
}
