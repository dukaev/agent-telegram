// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendPollHandler returns a handler for send_poll requests.
func SendPollHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.SendPollParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		result, err := client.SendPoll(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to send poll: %w", err)
		}

		return result, nil
	}
}

// SendChecklistHandler returns a handler for send_checklist (quiz) requests.
func SendChecklistHandler(client Client) func(json.RawMessage) (any, error) {
	return func(params json.RawMessage) (any, error) {
		var p telegram.SendPollParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if err := p.ValidateForQuiz(); err != nil {
			return nil, err
		}

		p.Quiz = true
		p.Anonymous = false

		result, err := client.SendPoll(context.Background(), p)
		if err != nil {
			return nil, fmt.Errorf("failed to send checklist: %w", err)
		}

		return result, nil
	}
}
