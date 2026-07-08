// Package bot provides high-level bot-flow commands for agents.
package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/cmd/send"
	"agent-telegram/internal/cliutil"
)

var (
	stepTo      cliutil.Recipient
	stepSend    string
	stepWait    bool
	stepTimeout time.Duration
	stepLimit   int

	pressTo      cliutil.Recipient
	pressWait    bool
	pressTimeout time.Duration
)

// BotCmd groups high-level bot-flow helpers.
var BotCmd = &cobra.Command{
	GroupID: "message",
	Use:     "bot",
	Short:   "High-level bot flow commands",
	Long: `High-level bot flow commands for agentic automation.

These commands combine message send/read, inline button inspection, reply
keyboard inspection, and wait handling into compact JSON responses.`,
}

// StepCmd sends optional text and returns the current bot state.
var StepCmd = &cobra.Command{
	Use:   "step [peer]",
	Short: "Send optional text and return bot state",
	Args:  cobra.MaximumNArgs(1),
}

// PressCmd presses an inline button and optionally waits for the next bot message.
var PressCmd = &cobra.Command{
	Use:   "press [peer] <message_id> <button_index>",
	Short: "Press an inline button in a bot flow",
	Args:  cobra.RangeArgs(2, 3),
}

// AddBotCommand adds bot commands to the root command.
func AddBotCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BotCmd)
	BotCmd.AddCommand(StepCmd, PressCmd)

	StepCmd.Flags().VarP(&stepTo, "to", "t", "Bot peer")
	StepCmd.Flags().StringVar(&stepSend, "send", "", "Text to send before reading state")
	StepCmd.Flags().BoolVar(&stepWait, "wait-reply", true, "Wait for a reply after --send")
	StepCmd.Flags().DurationVar(&stepTimeout, "timeout", 20*time.Second, "Maximum wait time")
	StepCmd.Flags().IntVar(&stepLimit, "limit", 10, "Messages to inspect when reading state")

	PressCmd.Flags().VarP(&pressTo, "to", "t", "Bot peer")
	PressCmd.Flags().BoolVar(&pressWait, "wait-reply", true, "Wait for a reply after pressing")
	PressCmd.Flags().DurationVar(&pressTimeout, "timeout", 20*time.Second, "Maximum wait time")

	StepCmd.Run = runStep
	PressCmd.Run = runPress
}

func runStep(_ *cobra.Command, args []string) {
	if len(args) > 0 {
		_ = stepTo.Set(args[0])
	}
	runner := cliutil.NewRunnerFromCmd(StepCmd, true)
	requirePeer(runner, stepTo.Peer())
	var action any
	var message map[string]any
	var waitMeta map[string]any

	if stepSend != "" {
		action = runner.Call("send_message", map[string]any{
			"peer":    stepTo.Peer(),
			"message": stepSend,
		})
		afterID := extractMessageID(action)
		if stepWait {
			reply, polls, err := send.WaitForReply(runner, stepTo.Peer(), afterID, stepTimeout)
			if err != nil {
				runner.Fatal(err.Error())
			}
			message, _ = reply.(map[string]any)
			waitMeta = map[string]any{
				"afterMessageId": afterID,
				"polls":          polls,
				"timeout":        stepTimeout.String(),
			}
		}
	}
	if message == nil {
		message = latestMessage(runner, stepTo.Peer(), stepLimit)
	}

	state := buildBotState(runner, stepTo.Peer(), message)
	if action != nil {
		state["action"] = action
	}
	if waitMeta != nil {
		state["wait"] = waitMeta
	}
	runner.PrintResult(state, nil)
}

func runPress(_ *cobra.Command, args []string) {
	var messageIDArg, buttonIndexArg string
	switch len(args) {
	case 3:
		_ = pressTo.Set(args[0])
		messageIDArg = args[1]
		buttonIndexArg = args[2]
	case 2:
		messageIDArg = args[0]
		buttonIndexArg = args[1]
	}
	runner := cliutil.NewRunnerFromCmd(PressCmd, true)
	requirePeer(runner, pressTo.Peer())
	messageID := runner.MustParseInt64(messageIDArg)
	action := runner.Call("press_inline_button", map[string]any{
		"peer":        pressTo.Peer(),
		"messageId":   messageID,
		"buttonIndex": runner.MustParseInt(buttonIndexArg),
	})

	var message map[string]any
	var waitMeta map[string]any
	if pressWait {
		reply, polls, err := send.WaitForReply(runner, pressTo.Peer(), messageID, pressTimeout)
		if err != nil {
			runner.Fatal(err.Error())
		}
		message, _ = reply.(map[string]any)
		waitMeta = map[string]any{
			"afterMessageId": messageID,
			"polls":          polls,
			"timeout":        pressTimeout.String(),
		}
	} else {
		message = latestMessage(runner, pressTo.Peer(), stepLimit)
	}

	state := buildBotState(runner, pressTo.Peer(), message)
	state["action"] = action
	if waitMeta != nil {
		state["wait"] = waitMeta
	}
	runner.PrintResult(state, nil)
}

func latestMessage(runner *cliutil.Runner, peer string, limit int) map[string]any {
	if limit <= 0 {
		limit = 10
	}
	result := runner.Call("get_messages", map[string]any{"username": peer, "limit": limit})
	m, ok := result.(map[string]any)
	if !ok {
		return nil
	}
	messages, ok := m["messages"].([]any)
	if !ok || len(messages) == 0 {
		return nil
	}
	message, _ := messages[0].(map[string]any)
	return message
}

func buildBotState(runner *cliutil.Runner, peer string, message map[string]any) map[string]any {
	state := map[string]any{
		"peer":    peer,
		"message": message,
	}
	if message == nil {
		next := nextActions(peer, 0, state)
		state["nextActions"] = next
		state["suggestedActions"] = actionNames(next)
		return state
	}

	messageID := extractMessageID(message)
	if messageID > 0 {
		buttons := runner.CallInternal("inspect_inline_buttons", map[string]any{
			"peer":      peer,
			"messageId": messageID,
		})
		if m, ok := buttons.(map[string]any); ok {
			state["inlineButtons"] = m["buttons"]
		}
	}
	keyboard := runner.CallInternal("inspect_reply_keyboard", map[string]any{"peer": peer})
	if m, ok := keyboard.(map[string]any); ok {
		state["replyKeyboard"] = m["keyboard"]
	}
	next := nextActions(peer, messageID, state)
	state["nextActions"] = next
	state["suggestedActions"] = actionNames(next)
	return state
}

func nextActions(peer string, messageID int64, state map[string]any) []map[string]any {
	var actions []map[string]any
	if buttons, ok := state["inlineButtons"].([]any); ok && len(buttons) > 0 {
		actions = append(actions, map[string]any{
			"kind":    "press_inline_button",
			"command": fmt.Sprintf("agent-telegram bot press %s %d <button_index> --agent", shellArg(peer), messageID),
			"safety":  "write",
			"reason":  "inline buttons are available on the latest bot message",
			"params": map[string]any{
				"peer":      peer,
				"messageId": messageID,
			},
		})
	}
	if keyboard, ok := state["replyKeyboard"].(map[string]any); ok && keyboard != nil {
		if rows, ok := keyboard["rows"].([]any); ok && len(rows) > 0 {
			actions = append(actions, map[string]any{
				"kind":    "press_reply_keyboard",
				"command": fmt.Sprintf("agent-telegram msg press-keyboard %s <button_text_or_index> --wait-reply --agent", shellArg(peer)),
				"safety":  "write",
				"reason":  "reply keyboard buttons are available",
				"params": map[string]any{
					"peer": peer,
				},
			})
		}
	}
	actions = append(actions, map[string]any{
		"kind":    "send_text",
		"command": fmt.Sprintf("agent-telegram bot step %s --send <text> --agent", shellArg(peer)),
		"safety":  "write",
		"reason":  "send text to continue the bot flow",
		"params": map[string]any{
			"peer": peer,
		},
	})
	return actions
}

func actionNames(actions []map[string]any) []string {
	names := make([]string, 0, len(actions))
	for _, action := range actions {
		if kind, _ := action["kind"].(string); kind != "" {
			names = append(names, kind)
		}
	}
	return names
}

func extractMessageID(value any) int64 {
	m, ok := value.(map[string]any)
	if !ok {
		return 0
	}
	return cliutil.ExtractInt64(m, "id")
}

func shellArg(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n\"'\\$`") {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
	return value
}

func requirePeer(runner *cliutil.Runner, peer string) {
	if peer == "" {
		runner.Fatal("peer is required (positional or --to)")
	}
}
