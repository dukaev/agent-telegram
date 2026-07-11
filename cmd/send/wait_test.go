package send

import (
	"strings"
	"testing"
	"time"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/ipc"
)

type fakeReplyPoller struct {
	results []any
	calls   int
}

func TestReplyTimeoutFailurePreservesActionAndRecoveryCorrelation(t *testing.T) {
	runner := cliutil.NewRunner("", true)
	outcome := WaitOutcome{AfterMessageID: 123, Polls: 4, Timeout: 20 * time.Second}
	action := map[string]any{"id": 123}
	err, details := replyTimeoutFailure(runner, "bot name'$`", action, outcome)
	if err.Code != ipc.ErrCodeTimeout {
		t.Fatalf("error code = %d, want timeout", err.Code)
	}
	data, _ := err.Data.(map[string]any)
	if data["type"] != ipc.ErrorTypeTimeout || data["actionSucceeded"] != true {
		t.Fatalf("error data = %#v", data)
	}
	if details.AuditStatus != "partial" || details.PartialResult == nil {
		t.Fatalf("details = %#v", details)
	}
	if len(details.NextActions) != 2 || details.NextActions[0]["kind"] != "wait_for_reply" || details.NextActions[1]["kind"] != "inspect_trace" {
		t.Fatalf("next actions = %#v", details.NextActions)
	}
	waitCommand, _ := details.NextActions[0]["command"].(string)
	traceCommand, _ := details.NextActions[1]["command"].(string)
	if !strings.Contains(waitCommand, "msg wait 'bot name'\\''$`' --after-id 123 --timeout 20s --agent --run-id "+runner.RunID()) {
		t.Fatalf("unsafe or incomplete wait command: %q", waitCommand)
	}
	if !strings.Contains(traceCommand, "trace inspect "+runner.TraceID()+" --agent --run-id "+runner.RunID()) {
		t.Fatalf("uncorrelated trace command: %q", traceCommand)
	}
}

func TestReplyTimeoutFailureForStandaloneWaitHasNoPartialAction(t *testing.T) {
	runner := cliutil.NewRunner("", true)
	err, details := replyTimeoutFailure(runner, "@bot", nil, WaitOutcome{Timeout: time.Second})
	data, _ := err.Data.(map[string]any)
	if _, ok := data["actionSucceeded"]; ok {
		t.Fatalf("standalone wait data = %#v", data)
	}
	if details.PartialResult != nil || details.AuditStatus != "error" {
		t.Fatalf("standalone details = %#v", details)
	}
}

func (f *fakeReplyPoller) CallInternal(method string, params any) any {
	if method != "get_messages" {
		panic("unexpected method: " + method)
	}
	index := f.calls
	f.calls++
	if index >= len(f.results) {
		return map[string]any{"messages": []any{}}
	}
	return f.results[index]
}

func TestWaitOutcomeReturnsIncomingMessageAfterAction(t *testing.T) {
	restoreWaitClock(t)
	base := time.Unix(1, 0)
	now := base
	waitNow = func() time.Time { return now }
	waitSleep = func(d time.Duration) { now = now.Add(d) }

	poller := &fakeReplyPoller{results: []any{
		map[string]any{"messages": []any{
			map[string]any{"id": float64(124), "out": true},
		}},
		map[string]any{"messages": []any{
			map[string]any{"id": float64(125), "out": false},
		}},
	}}
	outcome := WaitForReply(poller, "@bot", 123, time.Second)
	if !outcome.Completed || outcome.Polls != 2 || outcome.AfterMessageID != 123 || outcome.Timeout != time.Second {
		t.Fatalf("outcome = %+v", outcome)
	}
	reply, _ := outcome.Reply.(map[string]any)
	if got := int64(reply["id"].(float64)); got != 125 {
		t.Fatalf("reply id = %d, want 125", got)
	}
}

func TestWaitOutcomeDeadlineIsDeterministic(t *testing.T) {
	restoreWaitClock(t)
	base := time.Unix(1, 0)
	times := []time.Time{base, base.Add(time.Nanosecond)}
	waitNow = func() time.Time {
		if len(times) == 1 {
			return times[0]
		}
		value := times[0]
		times = times[1:]
		return value
	}
	waitSleep = func(time.Duration) { t.Fatal("deadline should not sleep") }

	outcome := WaitForReply(&fakeReplyPoller{}, "-5424738551", 123, time.Nanosecond)
	want := WaitOutcome{AfterMessageID: 123, Timeout: time.Nanosecond, Completed: false}
	if outcome != want {
		t.Fatalf("outcome = %+v, want %+v", outcome, want)
	}
}

func restoreWaitClock(t *testing.T) {
	t.Helper()
	oldNow, oldSleep := waitNow, waitSleep
	t.Cleanup(func() {
		waitNow = oldNow
		waitSleep = oldSleep
	})
}
