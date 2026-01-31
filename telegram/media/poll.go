// Package media provides Telegram poll operations.
package media

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendPoll sends a poll to a peer.
func (c *Client) SendPoll(ctx context.Context, params types.SendPollParams) (*types.SendPollResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := c.resolvePeer(ctx, params.Peer)
	if err != nil {
		return nil, err
	}

	// Build poll answers with random option bytes
	answers := make([]tg.PollAnswer, len(params.Options))
	optionBytes := make([][]byte, len(params.Options))
	for i, opt := range params.Options {
		optionBytes[i] = make([]byte, 8)
		if _, err := rand.Read(optionBytes[i]); err != nil {
			return nil, fmt.Errorf("failed to generate option bytes: %w", err)
		}
		answers[i] = tg.PollAnswer{
			Text:   tg.TextWithEntities{Text: opt.Text},
			Option: optionBytes[i],
		}
	}

	// Create the poll media
	poll := &tg.InputMediaPoll{
		Poll: tg.Poll{
			ID:       0,
			Question: tg.TextWithEntities{Text: params.Question},
			Closed:   false,
			Quiz:     params.Quiz,
			Answers:  answers,
		},
	}

	// Set correct answers for quiz
	if params.Quiz && params.CorrectIdx >= 0 && params.CorrectIdx < len(optionBytes) {
		poll.SetCorrectAnswers([][]byte{optionBytes[params.CorrectIdx]})
	}

	// Send the poll
	result, err := c.api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    poll,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send poll: %w", err)
	}

	msgID := extractMessageID(result)
	return &types.SendPollResult{
		ID:       msgID,
		Date:     time.Now().Unix(),
		Peer:     params.Peer,
		Question: params.Question,
	}, nil
}
