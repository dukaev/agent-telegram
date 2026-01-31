// Package ipc provides Telegram IPC handlers.
//nolint:dupl // Handler pattern has expected similarity
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendContactParams represents parameters for send_contact request.
type SendContactParams struct {
	Peer      string `json:"peer"`
	Phone     string `json:"phone"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
}

// SendContactHandler returns a handler for send_contact requests.
func SendContactHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendContactParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.Phone == "" {
			return nil, fmt.Errorf("phone is required")
		}
		if p.FirstName == "" {
			return nil, fmt.Errorf("firstName is required")
		}

		result, err := client.SendContact(context.Background(), telegram.SendContactParams{
			Peer:      p.Peer,
			Phone:     p.Phone,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send contact: %w", err)
		}

		return result, nil
	}
}
