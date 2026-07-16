package message

import (
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/cmd/send"
	"agent-telegram/internal/cliutil"
)

var (
	waitTo       cliutil.Recipient
	waitAfterID  int64
	waitThreadID int64
	waitTimeout  time.Duration
)

// WaitCmd waits for the next incoming message from a peer.
var WaitCmd = &cobra.Command{
	Use:   "wait [peer]",
	Short: "Wait for the next incoming message",
	Long: `Wait for the next incoming message from a peer after a known message ID.

This is the high-level agentic wait primitive. Internally it polls Telegram
messages, but polling steps are not written as separate CLI audit events.`,
	Args: cobra.MaximumNArgs(1),
}

// AddWaitCommand adds the wait command to the parent command.
func AddWaitCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(WaitCmd)
	cliutil.MarkFirstArgPeer(WaitCmd)

	WaitCmd.Flags().VarP(&waitTo, "to", "t", "Peer to wait on")
	WaitCmd.Flags().Int64Var(&waitAfterID, "after-id", 0, "Only return incoming messages after this message ID")
	WaitCmd.Flags().Int64Var(&waitThreadID, "thread-id", 0, "Forum topic root message ID")
	WaitCmd.Flags().DurationVar(&waitTimeout, "timeout", 30*time.Second, "Maximum time to wait")

	WaitCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = waitTo.Set(args[0])
		}
		runner := cliutil.NewRunnerFromCmd(WaitCmd, true)
		runner.SetAction("wait_for_reply")
		if waitTo.Peer() == "" {
			runner.Fatal("peer is required (positional or --to)")
		}

		if waitThreadID < 0 {
			runner.Fatal("thread-id must be >= 0")
		}

		outcome := send.WaitForReply(runner, waitTo.Peer(), waitThreadID, waitAfterID, waitTimeout)
		if !outcome.Completed {
			send.FailReplyTimeout(runner, waitTo.Peer(), nil, outcome)
			return
		}
		wait := map[string]any{
			"afterMessageId": waitAfterID,
			"polls":          outcome.Polls,
			"timeout":        waitTimeout.String(),
			"completed":      true,
		}
		if waitThreadID != 0 {
			wait["threadId"] = waitThreadID
		}
		runner.PrintResult(map[string]any{
			"reply": outcome.Reply,
			"wait":  wait,
		}, nil)
	}
}
