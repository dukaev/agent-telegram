// Package ipc provides Telegram IPC handlers.
package ipc

// InspectInlineButtonsHandler returns a handler for inspect_inline_buttons requests.
func InspectInlineButtonsHandler(client Client) HandlerFunc {
	return Handler(client.Message().InspectInlineButtons, "inspect inline buttons")
}

// PressInlineButtonHandler returns a handler for press_inline_button requests.
func PressInlineButtonHandler(client Client) HandlerFunc {
	return Handler(client.Message().PressInlineButton, "press inline button")
}
