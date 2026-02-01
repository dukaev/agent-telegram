// Package types provides common types for Telegram client contact operations.
package types // revive:disable:var-naming

import "fmt"

// Contact represents a contact in the user's contact list.
type Contact struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Username  string `json:"username,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Mutual    bool   `json:"mutual,omitempty"`
	Verified  bool   `json:"verified,omitempty"`
	Bot       bool   `json:"bot,omitempty"`
	Peer      string `json:"peer,omitempty"`
}

// GetContactsParams holds parameters for GetContacts.
type GetContactsParams struct {
	Query string `json:"query,omitempty"` // Search query to filter contacts
	Limit int    `json:"limit,omitempty"` // Maximum number of contacts to return
}

// Validate validates GetContactsParams (all fields are optional).
func (p GetContactsParams) Validate() error {
	return nil
}

// GetContactsResult is the result of GetContacts.
type GetContactsResult struct {
	Contacts []Contact `json:"contacts"`
	Count    int       `json:"count"`
	Query    string    `json:"query,omitempty"`
}

// AddContactParams holds parameters for AddContact.
type AddContactParams struct {
	Phone     string `json:"phone" validate:"required"`     // Phone number (with country code, e.g. +1234567890)
	FirstName string `json:"firstName" validate:"required"` // First name
	LastName  string `json:"lastName,omitempty"`            // Last name (optional)
}

// Validate validates AddContactParams.
func (p AddContactParams) Validate() error {
	return ValidateStruct(p)
}

// AddContactResult is the result of AddContact.
type AddContactResult struct {
	Success bool    `json:"success"`
	Contact Contact `json:"contact,omitempty"`
}

// DeleteContactParams holds parameters for DeleteContact.
type DeleteContactParams struct {
	Username string `json:"username,omitempty"` // Username to delete (e.g. "username" or "@username")
	UserID   int64  `json:"userId,omitempty"`   // User ID to delete
}

// Validate validates DeleteContactParams (username OR userId required).
func (p DeleteContactParams) Validate() error {
	if p.Username == "" && p.UserID == 0 {
		return fmt.Errorf("username or userId is required")
	}
	return nil
}

// DeleteContactResult is the result of DeleteContact.
type DeleteContactResult struct {
	Success bool `json:"success"`
}
