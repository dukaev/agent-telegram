// Package user provides Telegram user operations.
package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

// Client provides user operations.
type Client struct {
	*client.BaseClient
}

// NewClient creates a new user client.
func NewClient(tc client.ParentClient) *Client {
	return &Client{
		BaseClient: &client.BaseClient{Parent: tc},
	}
}

// trimUsernamePrefix removes the @ prefix from username.
func trimUsernamePrefix(username string) string {
	if len(username) > 0 && username[0] == '@' {
		return username[1:]
	}
	return username
}

// GetContacts returns the user's contact list with optional search filter.
func (c *Client) GetContacts(ctx context.Context, params types.GetContactsParams) (*types.GetContactsResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Set defaults
	if params.Limit <= 0 {
		params.Limit = 100
	}

	// Get all contacts (hash 0 means get all)
	contactsClass, err := c.API.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	var users []*tg.User
	switch c := contactsClass.(type) {
	case *tg.ContactsContacts:
		// Extract users from the response
		for _, u := range c.Users {
			if user, ok := u.(*tg.User); ok {
				users = append(users, user)
			}
		}
	case *tg.ContactsContactsNotModified:
		return nil, fmt.Errorf("contacts not modified")
	default:
		return nil, fmt.Errorf("unexpected contacts type: %T", c)
	}

	// Build result from contacts
	result := &types.GetContactsResult{
		Contacts: make([]types.Contact, 0),
		Query:    params.Query,
	}

	// Filter and convert contacts
	for _, user := range users {
		contact := convertUserToContact(user)

		// Apply search filter if query is provided
		if params.Query != "" && !matchesQuery(contact, params.Query) {
			continue
		}

		result.Contacts = append(result.Contacts, contact)

		// Stop if we've reached the limit
		if len(result.Contacts) >= params.Limit {
			break
		}
	}

	result.Count = len(result.Contacts)
	return result, nil
}

// convertUserToContact converts a tg.User to a types.Contact.
func convertUserToContact(user *tg.User) types.Contact {
	contact := types.Contact{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Phone:     user.Phone,
		Verified:  user.Verified,
		Bot:       user.Bot,
	}

	// Build peer string
	if user.Username != "" {
		contact.Peer = "@" + user.Username
	} else {
		contact.Peer = fmt.Sprintf("user%d", user.ID)
	}

	return contact
}

// matchesQuery checks if a contact matches the search query.
func matchesQuery(contact types.Contact, query string) bool {
	query = strings.ToLower(query)

	// Check first name
	if strings.Contains(strings.ToLower(contact.FirstName), query) {
		return true
	}

	// Check last name
	if strings.Contains(strings.ToLower(contact.LastName), query) {
		return true
	}

	// Check username
	if strings.Contains(strings.ToLower(contact.Username), query) {
		return true
	}

	// Check phone
	if strings.Contains(contact.Phone, query) {
		return true
	}

	return false
}

// AddContact adds a new contact to the user's contact list.
func (c *Client) AddContact(ctx context.Context, params types.AddContactParams) (*types.AddContactResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	// Import contact using phone number
	contactClass, err := c.API.ContactsImportContacts(ctx, []tg.InputPhoneContact{
		{
			ClientID:  0,
			Phone:     params.Phone,
			FirstName: params.FirstName,
			LastName:  params.LastName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add contact: %w", err)
	}

	result := &types.AddContactResult{Success: true}

	// Extract the imported user
	if len(contactClass.Users) > 0 {
		if user, ok := contactClass.Users[0].(*tg.User); ok {
			result.Contact = convertUserToContact(user)
		}
	}

	return result, nil
}

// DeleteContact deletes a contact from the user's contact list.
func (c *Client) DeleteContact(
	ctx context.Context,
	params types.DeleteContactParams,
) (*types.DeleteContactResult, error) {
	if err := c.CheckInitialized(); err != nil {
		return nil, err
	}

	var userID int64

	if params.UserID != 0 {
		userID = params.UserID
	} else if params.Username != "" {
		// Resolve username to get user ID
		username := trimUsernamePrefix(params.Username)
		peerClass, err := c.API.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve username: %w", err)
		}

		switch p := peerClass.Peer.(type) {
		case *tg.PeerUser:
			userID = p.UserID
		default:
			return nil, fmt.Errorf("not a user: %s", params.Username)
		}
	}

	// Delete the contact using the ID
	_, err := c.API.ContactsDeleteByPhones(ctx, []string{fmt.Sprintf("%d", userID)})
	if err != nil {
		return nil, fmt.Errorf("failed to delete contact: %w", err)
	}

	return &types.DeleteContactResult{Success: true}, nil
}
