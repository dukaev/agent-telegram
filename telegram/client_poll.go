// Package telegram provides Telegram client poll functionality.
package telegram

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

// SendPoll sends a poll to a peer.
func (c *Client) SendPoll(ctx context.Context, params SendPollParams) (*SendPollResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	api := c.client.API()

	inputPeer, err := resolvePeer(ctx, api, params.Peer)
	if err != nil {
		return nil, err
	}

	// Create poll answers
	answers := make([]tg.PollAnswer, len(params.Options))
	for i, opt := range params.Options {
		optionData := make([]byte, 8)
		if _, err := rand.Read(optionData); err != nil {
			return nil, fmt.Errorf("failed to generate option data: %w", err)
		}

		answers[i] = tg.PollAnswer{
			Text: tg.TextWithEntities{
				Text: opt.Text,
			},
			Option: optionData,
		}
	}

	// Create poll
	poll := tg.Poll{
		Question: tg.TextWithEntities{
			Text: params.Question,
		},
		Answers:      answers,
		PublicVoters: !params.Anonymous,
		Quiz:         params.Quiz,
	}

	// Create media poll
	mediaPoll := &tg.InputMediaPoll{
		Poll: poll,
	}

	// Set correct answer for quiz
	if params.Quiz && params.CorrectIdx >= 0 && params.CorrectIdx < len(answers) {
		mediaPoll.SetCorrectAnswers([][]byte{answers[params.CorrectIdx].Option})
	}

	result, err := api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    mediaPoll,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send poll: %w", err)
	}

	msgID := extractMessageID(result)
	return &SendPollResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Question: params.Question,
	}, nil
}
