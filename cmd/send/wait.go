// Package send provides commands for sending messages and media.
package send

import (
	"fmt"
	"os"
	"time"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/ipc"
	"agent-telegram/internal/observability"
)

const (
	waitPollInterval = 500 * time.Millisecond
	waitPollLimit    = 10
)

var (
	waitNow   = time.Now
	waitSleep = time.Sleep
)

// ReplyPoller is the narrow polling surface used by reply waits.
type ReplyPoller interface {
	CallInternal(method string, params any) any
}

// WaitOutcome describes either a completed reply wait or its deadline.
type WaitOutcome struct {
	Reply          any
	ThreadID       int64
	AfterMessageID int64
	Polls          int
	Timeout        time.Duration
	Completed      bool
}

// WaitForReply polls get_messages until a reply appears after afterMsgID, or timeout.
// Its outcome preserves polling and deadline metadata without classifying the
// timeout as a validation error.
func WaitForReply(poller ReplyPoller, peer string, threadID, afterMsgID int64, timeout time.Duration) WaitOutcome {
	outcome := WaitOutcome{ThreadID: threadID, AfterMessageID: afterMsgID, Timeout: timeout}
	deadline := waitNow().Add(timeout)
	polls := 0

	for waitNow().Before(deadline) {
		params := map[string]any{
			"username": peer,
			"limit":    waitPollLimit,
		}
		if threadID != 0 {
			params["threadId"] = threadID
		}

		polls++
		result := poller.CallInternal("get_messages", params)

		if reply := findReply(result, afterMsgID, threadID); reply != nil {
			outcome.Reply = reply
			outcome.Polls = polls
			outcome.Completed = true
			return outcome
		}

		remaining := deadline.Sub(waitNow())
		if remaining <= 0 {
			break
		}
		if remaining < waitPollInterval {
			waitSleep(remaining)
		} else {
			waitSleep(waitPollInterval)
		}
	}

	outcome.Polls = polls
	return outcome
}

// FailReplyTimeout reports a reply deadline without losing evidence of a
// successful action that preceded the wait.
func FailReplyTimeout(runner *cliutil.Runner, peer string, action any, outcome WaitOutcome) {
	err, details := replyTimeoutFailure(runner, peer, action, outcome)
	runner.FailTyped(err, details)
}

func replyTimeoutFailure(runner *cliutil.Runner, peer string, action any, outcome WaitOutcome) (*ipc.ErrorObject, cliutil.FailureDetails) {
	wait := map[string]any{
		"afterMessageId": outcome.AfterMessageID,
		"polls":          outcome.Polls,
		"timeout":        outcome.Timeout.String(),
		"completed":      false,
	}
	if outcome.ThreadID != 0 {
		wait["threadId"] = outcome.ThreadID
	}
	data := map[string]any{"wait": wait}
	details := cliutil.FailureDetails{
		AuditStatus:  "error",
		AuditSummary: map[string]any{"wait": wait},
	}
	if action != nil {
		data["actionSucceeded"] = true
		details.PartialResult = map[string]any{
			"action": action,
			"wait":   wait,
		}
		details.AuditStatus = "partial"
		details.AuditSummary = map[string]any{
			"action": observability.SummarizeResult(action),
			"wait":   wait,
		}
	}
	threadArg := ""
	if outcome.ThreadID != 0 {
		threadArg = fmt.Sprintf(" --thread-id %d", outcome.ThreadID)
	}
	details.NextActions = []map[string]any{
		{
			"kind":    "wait_for_reply",
			"command": fmt.Sprintf("agent-telegram msg wait %s --after-id %d%s --timeout %s --agent --run-id %s", cliutil.ShellArg(peer), outcome.AfterMessageID, threadArg, outcome.Timeout, cliutil.ShellArg(runner.RunID())),
			"safety":  "read",
			"reason":  "continue waiting without repeating the write action",
		},
		{
			"kind":    "inspect_trace",
			"command": fmt.Sprintf("agent-telegram trace inspect %s --agent --run-id %s", cliutil.ShellArg(runner.TraceID()), cliutil.ShellArg(runner.RunID())),
			"safety":  "read",
			"reason":  "inspect correlated audit and log events",
		},
	}
	err := ipc.NewTypedError(ipc.ErrCodeTimeout, ipc.ErrorTypeTimeout, fmt.Sprintf("no reply within %s", outcome.Timeout), data)
	return err, details
}

// findReply searches messages for an incoming message with ID > afterMsgID.
func findReply(result any, afterMsgID, threadID int64) map[string]any {
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
		if msgID <= afterMsgID {
			continue
		}
		if threadID != 0 && cliutil.ExtractInt64(msg, "threadId") != threadID {
			continue
		}
		return msg
	}

	return nil
}

// HandleWaitReply performs the wait-reply flow after a send.
// sendResult is the result from the send call. Exits on timeout.
func HandleWaitReply(runner *cliutil.Runner, peer string, threadID int64, sendResult any, timeout time.Duration) {
	sentID := extractSentID(sendResult)
	if sentID == 0 {
		fmt.Fprintln(os.Stderr, "Warning: could not extract sent message ID, waiting for any new message")
	}
	HandleWaitReplyAfter(runner, peer, threadID, sentID, sendResult, timeout)
}

// HandleWaitReplyAfter waits for a reply after a known message ID and prints a
// combined result with both the triggering action and the reply.
func HandleWaitReplyAfter(runner *cliutil.Runner, peer string, threadID, afterMsgID int64, actionResult any, timeout time.Duration) {
	fmt.Fprintf(os.Stderr, "Waiting for reply (timeout: %s)...\n", timeout)

	outcome := WaitForReply(runner, peer, threadID, afterMsgID, timeout)
	if !outcome.Completed {
		FailReplyTimeout(runner, peer, actionResult, outcome)
		return
	}

	wait := map[string]any{
		"afterMessageId": afterMsgID,
		"polls":          outcome.Polls,
		"timeout":        timeout.String(),
		"completed":      true,
	}
	if threadID != 0 {
		wait["threadId"] = threadID
	}
	runner.PrintResult(map[string]any{
		"action": actionResult,
		"reply":  outcome.Reply,
		"wait":   wait,
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
