// Package ipc provides Telegram IPC handlers.
package ipc

// GetContactsHandler returns a handler for get_contacts requests.
func GetContactsHandler(client Client) HandlerFunc {
	return Handler(client.User().GetContacts, "get contacts")
}

// AddContactHandler returns a handler for add_contact requests.
func AddContactHandler(client Client) HandlerFunc {
	return Handler(client.User().AddContact, "add contact")
}

// DeleteContactHandler returns a handler for delete_contact requests.
func DeleteContactHandler(client Client) HandlerFunc {
	return Handler(client.User().DeleteContact, "delete contact")
}
