// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"
	"time"

	"agent-telegram/internal/cliutil"
)

const (
	waitPollInterval = 500 * time.Millisecond
	waitPollLimit    = 10
)

// WaitForReply polls get_messages until a reply appears after afterMsgID, or timeout.
// Returns the first incoming message (not sent by us) with ID > afterMsgID.
func WaitForReply(runner *cliutil.Runner, peer string, afterMsgID int64, timeout time.Duration) (any, int, error) {
	deadline := time.Now().Add(timeout)
	polls := 0

	for time.Now().Before(deadline) {
		params := map[string]any{
			"username": peer,
			"limit":    waitPollLimit,
		}

		polls++
		result := runner.CallInternal("get_messages", params)

		if reply := findReply(result, afterMsgID); reply != nil {
			return reply, polls, nil
		}

		remaining := time.Until(deadline)
		if remaining < waitPollInterval {
			time.Sleep(remaining)
		} else {
			time.Sleep(waitPollInterval)
		}
	}

	return nil, polls, fmt.Errorf("no reply within %s", timeout)
}

// findReply searches messages for an incoming message with ID > afterMsgID.
func findReply(result any, afterMsgID int64) map[string]any {
	r, ok := result.(map[string]any)
	if !ok {
		return nil
	}

	messages, ok := r["messages"].([]any)
	if !ok {
		return nil
	}

	for _, item := range messages {
		msg, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Skip our own messages
		if out, _ := msg["out"].(bool); out {
			continue
		}

		// Check message ID > afterMsgID
		msgID := cliutil.ExtractInt64(msg, "id")
		if msgID > afterMsgID {
			return msg
		}
	}

	return nil
}

// HandleWaitReply performs the wait-reply flow after a send.
// sendResult is the result from the send call. Exits on timeout.
func HandleWaitReply(runner *cliutil.Runner, peer string, sendResult any, timeout time.Duration) {
	sentID := extractSentID(sendResult)
	if sentID == 0 {
		fmt.Fprintln(os.Stderr, "Warning: could not extract sent message ID, waiting for any new message")
	}
	HandleWaitReplyAfter(runner, peer, sentID, sendResult, timeout)
}

// HandleWaitReplyAfter waits for a reply after a known message ID and prints a
// combined result with both the triggering action and the reply.
func HandleWaitReplyAfter(runner *cliutil.Runner, peer string, afterMsgID int64, actionResult any, timeout time.Duration) {
	fmt.Fprintf(os.Stderr, "Waiting for reply (timeout: %s)...\n", timeout)

	reply, polls, err := WaitForReply(runner, peer, afterMsgID, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	runner.PrintResult(map[string]any{
		"action": actionResult,
		"reply":  reply,
		"wait": map[string]any{
			"afterMessageId": afterMsgID,
			"polls":          polls,
			"timeout":        timeout.String(),
		},
	}, nil)
}

// extractSentID extracts the message ID from a send result.
func extractSentID(result any) int64 {
	r, ok := result.(map[string]any)
	if !ok {
		return 0
	}
	return cliutil.ExtractInt64(r, "id")
}
