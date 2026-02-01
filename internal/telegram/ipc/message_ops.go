// Package ipc provides Telegram IPC handlers.
package ipc

// SendReplyHandler returns a handler for send_reply requests.
func SendReplyHandler(client Client) HandlerFunc {
	return Handler(client.Message().SendReply, "send reply")
}

// UpdateMessageHandler returns a handler for update_message requests.
func UpdateMessageHandler(client Client) HandlerFunc {
	return Handler(client.Message().UpdateMessage, "update message")
}

// DeleteMessageHandler returns a handler for delete_message requests.
func DeleteMessageHandler(client Client) HandlerFunc {
	return Handler(client.Message().DeleteMessage, "delete message")
}

// ForwardMessageHandler returns a handler for forward_message requests.
func ForwardMessageHandler(client Client) HandlerFunc {
	return Handler(client.Message().ForwardMessage, "forward message")
}
