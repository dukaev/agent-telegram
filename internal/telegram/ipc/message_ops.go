// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// SendReplyHandler returns a handler for send_reply requests.
func SendReplyHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.SendReply, "send reply")
}

// UpdateMessageHandler returns a handler for update_message requests.
func UpdateMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.UpdateMessage, "update message")
}

// DeleteMessageHandler returns a handler for delete_message requests.
func DeleteMessageHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.DeleteMessage, "delete message")
}
