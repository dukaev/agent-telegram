// Package message provides commands for managing messages.
package message

import (
	"github.com/spf13/cobra"
	"agent-telegram/cmd/send"
)

// MsgCmd represents the msg command group.
var MsgCmd = &cobra.Command{
	GroupID: "message",
	Use:     "msg",
	Short:   "Message management commands",
	Long:    `Commands for managing Telegram messages - send, delete, forward, pin, react, and more.`,
}

// AddMsgCommand adds the msg command group to the root command.
func AddMsgCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MsgCmd)

	// Call AddXxxCommand functions with MsgCmd to setup flags and Run
	AddDeleteCommand(MsgCmd)
	AddForwardCommand(MsgCmd)
	AddPinMessageCommand(MsgCmd)
	AddInspectButtonsCommand(MsgCmd)
	AddPressButtonCommand(MsgCmd)
	AddReactionCommand(MsgCmd)
	AddInspectKeyboardCommand(MsgCmd)
	AddReadCommand(MsgCmd)
	AddTypingCommand(MsgCmd)
	AddScheduledCommand(MsgCmd)

	// Update Use strings for subcommands
	DeleteCmd.Use = "delete <message_id|id1,id2,...>"
	ForwardCmd.Use = "forward <message_id>"
	PinMessageCmd.Use = "pin <message_id>"
	InspectButtonsCmd.Use = "inspect-buttons <message_id>"
	PressButtonCmd.Use = "press-button <message_id> <button_index>"
	ReactionCmd.Use = "reaction <message_id> <emoji>"
	InspectKeyboardCmd.Use = "inspect-keyboard"

	// Add send command as a top-level command (in "Manage Messages" group)
	send.AddSendCommand(rootCmd)
}
