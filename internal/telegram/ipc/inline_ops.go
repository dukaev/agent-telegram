// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// InspectInlineButtonsHandler returns a handler for inspect_inline_buttons requests.
func InspectInlineButtonsHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(
		client.InspectInlineButtons,
		"inspect inline buttons",
	)
}

// PressInlineButtonHandler returns a handler for press_inline_button requests.
func PressInlineButtonHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(
		client.PressInlineButton,
		"press inline button",
	)
}
