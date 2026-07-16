// Package bot provides high-level bot-flow commands for agents.
package bot

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/cmd/send"
	"agent-telegram/internal/cliutil"
)

var (
	stepTo      cliutil.Recipient
	stepSend    string
	stepText    string
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
	Example: `  agent-telegram bot step <peer> --send <text>
  agent-telegram bot step -5424738551 --send /start`,
	Args: cobra.MaximumNArgs(1),
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
	cliutil.MarkFirstArgPeer(StepCmd)
	cliutil.MarkFirstArgPeer(PressCmd)

	StepCmd.Flags().VarP(&stepTo, "to", "t", "Bot peer")
	StepCmd.Flags().StringVar(&stepSend, "send", "", "Text to send before reading state")
	StepCmd.Flags().StringVar(&stepText, "text", "", "Alias for --send")
	StepCmd.Flags().BoolVar(&stepWait, "wait-reply", true, "Wait for a reply after --send")
	StepCmd.Flags().DurationVar(&stepTimeout, "timeout", 20*time.Second, "Maximum wait time")
	StepCmd.Flags().IntVar(&stepLimit, "limit", 10, "Messages to inspect when reading state")

	PressCmd.Flags().VarP(&pressTo, "to", "t", "Bot peer")
	PressCmd.Flags().BoolVar(&pressWait, "wait-reply", true, "Wait for a reply after pressing")
	PressCmd.Flags().DurationVar(&pressTimeout, "timeout", 20*time.Second, "Maximum wait time")

	StepCmd.Run = runStep
	PressCmd.Run = runPress
}

func runStep(cmd *cobra.Command, args []string) {
	runner := cliutil.NewRunnerFromCmd(cmd, true)
	messageText, err := resolveStepText(cmd, stepSend, stepText)
	if err != nil {
		runner.Fatal(err.Error())
	}
	if len(args) > 0 {
		_ = stepTo.Set(args[0])
	}
	requirePeer(runner, stepTo.Peer())
	var action any
	var message map[string]any
	var waitMeta map[string]any

	if messageText != "" {
		action = runner.Call("send_message", map[string]any{
			"peer":    stepTo.Peer(),
			"message": messageText,
		})
		afterID := extractMessageID(action)
		if stepWait {
			outcome := send.WaitForReply(runner, stepTo.Peer(), 0, afterID, stepTimeout)
			if !outcome.Completed {
				send.FailReplyTimeout(runner, stepTo.Peer(), action, outcome)
				return
			}
			message, _ = outcome.Reply.(map[string]any)
			waitMeta = map[string]any{
				"afterMessageId": afterID,
				"polls":          outcome.Polls,
				"timeout":        stepTimeout.String(),
				"completed":      true,
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

func resolveStepText(cmd *cobra.Command, sendValue, textValue string) (string, error) {
	sendChanged := cmd.Flags().Changed("send")
	textChanged := cmd.Flags().Changed("text")
	if sendChanged && textChanged && sendValue != textValue {
		return "", fmt.Errorf("use only --send or --text when values differ; example: agent-telegram bot step <peer> --send <text>")
	}
	if textChanged {
		return textValue, nil
	}
	return sendValue, nil
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
		outcome := send.WaitForReply(runner, pressTo.Peer(), 0, messageID, pressTimeout)
		if !outcome.Completed {
			send.FailReplyTimeout(runner, pressTo.Peer(), action, outcome)
			return
		}
		message, _ = outcome.Reply.(map[string]any)
		waitMeta = map[string]any{
			"afterMessageId": messageID,
			"polls":          outcome.Polls,
			"timeout":        pressTimeout.String(),
			"completed":      true,
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
			"command": fmt.Sprintf("agent-telegram bot press %s %d <button_index> --agent", cliutil.ShellArg(peer), messageID),
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
				"command": fmt.Sprintf("agent-telegram msg press-keyboard %s <button_text_or_index> --wait-reply --agent", cliutil.ShellArg(peer)),
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
		"command": fmt.Sprintf("agent-telegram bot step %s --send <text> --agent", cliutil.ShellArg(peer)),
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

func requirePeer(runner *cliutil.Runner, peer string) {
	if peer == "" {
		runner.Fatal("peer is required (positional or --to)")
	}
}
