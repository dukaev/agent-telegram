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
func WaitForReply(runner *cliutil.Runner, peer string, afterMsgID int64, timeout time.Duration) (any, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		params := map[string]any{
			"username": peer,
			"limit":    waitPollLimit,
		}

		result := runner.Call("get_messages", params)

		if reply := findReply(result, afterMsgID); reply != nil {
			return reply, nil
		}

		remaining := time.Until(deadline)
		if remaining < waitPollInterval {
			time.Sleep(remaining)
		} else {
			time.Sleep(waitPollInterval)
		}
	}

	return nil, fmt.Errorf("no reply within %s", timeout)
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

	fmt.Fprintf(os.Stderr, "Waiting for reply (timeout: %s)...\n", timeout)

	reply, err := WaitForReply(runner, peer, sentID, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	runner.PrintResult(reply, nil)
}

// extractSentID extracts the message ID from a send result.
func extractSentID(result any) int64 {
	r, ok := result.(map[string]any)
	if !ok {
		return 0
	}
	return cliutil.ExtractInt64(r, "id")
}
