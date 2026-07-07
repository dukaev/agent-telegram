package message

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/cmd/send"
	"agent-telegram/internal/cliutil"
)

var (
	waitTo      cliutil.Recipient
	waitAfterID int64
	waitTimeout time.Duration
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

	WaitCmd.Flags().VarP(&waitTo, "to", "t", "Peer to wait on")
	WaitCmd.Flags().Int64Var(&waitAfterID, "after-id", 0, "Only return incoming messages after this message ID")
	WaitCmd.Flags().DurationVar(&waitTimeout, "timeout", 30*time.Second, "Maximum time to wait")

	WaitCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = waitTo.Set(args[0])
		}
		if waitTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(WaitCmd, true)
		reply, polls, err := send.WaitForReply(runner, waitTo.Peer(), waitAfterID, waitTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		runner.PrintResult(map[string]any{
			"reply": reply,
			"wait": map[string]any{
				"afterMessageId": waitAfterID,
				"polls":          polls,
				"timeout":        waitTimeout.String(),
			},
		}, nil)
	}
}
