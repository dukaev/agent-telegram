// Package ipc provides Telegram IPC handlers.
package ipc

import "encoding/json"

// GetContactsHandler returns a handler for get_contacts requests.
func GetContactsHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.GetContacts, "get contacts")
}

// AddContactHandler returns a handler for add_contact requests.
func AddContactHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.AddContact, "add contact")
}

// DeleteContactHandler returns a handler for delete_contact requests.
func DeleteContactHandler(client Client) func(json.RawMessage) (any, error) {
	return Handler(client.DeleteContact, "delete contact")
}
