package bot

import (
	"encoding/json"
	"time"

	"agent-telegram/cmd/send"
	"agent-telegram/internal/cliutil"
)

const botWaitPollInterval = 500 * time.Millisecond

var (
	botWaitNow   = time.Now
	botWaitSleep = time.Sleep
)

type botPoller interface {
	CallInternal(method string, params any) any
}

type messageSnapshot struct {
	ID       int64
	ThreadID int64
	EditDate int64
	Text     string
	Buttons  string
}

type botWaitOutcome struct {
	Message   map[string]any
	Event     string
	Polls     int
	Completed bool
}

func newMessageSnapshot(message map[string]any) messageSnapshot {
	buttons, err := json.Marshal(message["buttons"])
	if err != nil {
		buttons = []byte("null")
	}
	text, _ := message["text"].(string)
	return messageSnapshot{
		ID:       cliutil.ExtractInt64(message, "id"),
		ThreadID: cliutil.ExtractInt64(message, "threadId"),
		EditDate: cliutil.ExtractInt64(message, "editDate"),
		Text:     text,
		Buttons:  string(buttons),
	}
}

func snapshotChanged(before messageSnapshot, after map[string]any) bool {
	current := newMessageSnapshot(after)
	if before.ID == 0 || current.ID != before.ID {
		return false
	}
	return current.EditDate != before.EditDate || current.Text != before.Text || current.Buttons != before.Buttons
}

func getMessage(poller botPoller, peer string, messageID int64) map[string]any {
	result := poller.CallInternal("get_message", map[string]any{
		"peer": peer, "messageId": messageID,
	})
	envelope, ok := result.(map[string]any)
	if !ok {
		return nil
	}
	message, _ := envelope["message"].(map[string]any)
	return message
}

func waitForBotEvent(
	poller botPoller,
	peer string,
	threadID, afterMessageID int64,
	before messageSnapshot,
	timeout time.Duration,
) botWaitOutcome {
	deadline := botWaitNow().Add(timeout)
	outcome := botWaitOutcome{}
	for botWaitNow().Before(deadline) {
		params := map[string]any{"username": peer, "limit": 10}
		if threadID != 0 {
			params["threadId"] = threadID
		}
		outcome.Polls++
		if message := send.FindReply(poller.CallInternal("get_messages", params), afterMessageID, threadID); message != nil {
			outcome.Message = message
			outcome.Event = "new_message"
			outcome.Completed = true
			return outcome
		}
		if current := getMessage(poller, peer, afterMessageID); snapshotChanged(before, current) {
			outcome.Message = current
			outcome.Event = "message_edited"
			outcome.Completed = true
			return outcome
		}

		remaining := deadline.Sub(botWaitNow())
		if remaining <= 0 {
			break
		}
		if remaining < botWaitPollInterval {
			botWaitSleep(remaining)
		} else {
			botWaitSleep(botWaitPollInterval)
		}
	}
	return outcome
}
