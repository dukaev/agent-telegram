// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// PollOption represents a poll option.
type PollOption struct {
	Text string `json:"text"`
}

// SendPollParams represents parameters for send_poll request.
type SendPollParams struct {
	Peer       string       `json:"peer,omitempty"`
	Username   string       `json:"username,omitempty"`
	Question   string       `json:"question"`
	Options    []PollOption `json:"options"`
	Anonymous  bool         `json:"anonymous,omitempty"`
	Quiz       bool         `json:"quiz,omitempty"`
	CorrectIdx int          `json:"correctIdx,omitempty"`
}

// SendPollHandler returns a handler for send_poll requests.
func SendPollHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendPollParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.Question == "" {
			return nil, fmt.Errorf("question is required")
		}
		if len(p.Options) < 2 {
			return nil, fmt.Errorf("at least 2 options are required")
		}
		if len(p.Options) > 10 {
			return nil, fmt.Errorf("maximum 10 options allowed")
		}

		// Convert options
		options := make([]telegram.PollOption, len(p.Options))
		for i, opt := range p.Options {
			options[i] = telegram.PollOption{Text: opt.Text}
		}

		result, err := client.SendPoll(context.Background(), telegram.SendPollParams{
			Peer:       p.Peer,
			Username:   p.Username,
			Question:   p.Question,
			Options:    options,
			Anonymous:  p.Anonymous,
			Quiz:       p.Quiz,
			CorrectIdx: p.CorrectIdx,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send poll: %w", err)
		}

		return result, nil
	}
}

// SendChecklistHandler returns a handler for send_checklist (quiz) requests.
func SendChecklistHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendPollParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" && p.Username == "" {
			return nil, fmt.Errorf("peer or username is required")
		}
		if p.Question == "" {
			return nil, fmt.Errorf("question is required")
		}
		if len(p.Options) < 2 {
			return nil, fmt.Errorf("at least 2 options are required")
		}
		if len(p.Options) > 10 {
			return nil, fmt.Errorf("maximum 10 options allowed")
		}
		if p.CorrectIdx < 0 || p.CorrectIdx >= len(p.Options) {
			return nil, fmt.Errorf("correctIdx must be between 0 and %d", len(p.Options)-1)
		}

		// Convert options
		options := make([]telegram.PollOption, len(p.Options))
		for i, opt := range p.Options {
			options[i] = telegram.PollOption{Text: opt.Text}
		}

		result, err := client.SendPoll(context.Background(), telegram.SendPollParams{
			Peer:       p.Peer,
			Username:   p.Username,
			Question:   p.Question,
			Options:    options,
			Anonymous:  false,
			Quiz:       true,
			CorrectIdx: p.CorrectIdx,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send checklist: %w", err)
		}

		return result, nil
	}
}
