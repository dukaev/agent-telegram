// Package media provides Telegram poll operations.
package media

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// SendPoll sends a poll to a peer.
func (c *Client) SendPoll(ctx context.Context, params types.SendPollParams) (*types.SendPollResult, error) {
	if c.api == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	inputPeer, err := resolvePeer(ctx, c.api, params.Peer)
	if err != nil {
		return nil, err
	}

	sender := message.NewSender(c.api)

	var result tg.UpdatesClass
	if params.Quiz {
		// Create quiz (poll with correct answer)
		if len(params.Options) < 2 {
			return nil, fmt.Errorf("at least 2 options are required")
		}

		// Build answer options
		answerOpts := make([]message.PollAnswerOption, len(params.Options))
		for i, opt := range params.Options {
			if i == params.CorrectIdx {
				answerOpts[i] = message.CorrectPollAnswer(opt.Text)
			} else {
				answerOpts[i] = message.PollAnswer(opt.Text)
			}
		}

		result, err = sender.To(inputPeer).Media(ctx, message.Poll(
			params.Question,
			answerOpts[0],
			answerOpts[1],
			answerOpts[2:]...,
		))
	} else {
		// Create regular poll
		if len(params.Options) < 2 {
			return nil, fmt.Errorf("at least 2 options are required")
		}

		// Build answer options
		answerOpts := make([]message.PollAnswerOption, len(params.Options))
		for i, opt := range params.Options {
			answerOpts[i] = message.PollAnswer(opt.Text)
		}

		result, err = sender.To(inputPeer).Media(ctx, message.Poll(
			params.Question,
			answerOpts[0],
			answerOpts[1],
			answerOpts[2:]...,
		))
	}

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
