// Package ipc provides Telegram IPC handlers.
package ipc

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-telegram/telegram"
)

// SendLocationParams represents parameters for send_location request.
type SendLocationParams struct {
	Peer     string  `json:"peer"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SendLocationHandler returns a handler for send_location requests.
func SendLocationHandler(client Client) func(json.RawMessage) (interface{}, error) {
	return func(params json.RawMessage) (interface{}, error) {
		var p SendLocationParams
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}
		}

		// Validate
		if p.Peer == "" {
			return nil, fmt.Errorf("peer is required")
		}
		if p.Latitude < -90 || p.Latitude > 90 {
			return nil, fmt.Errorf("latitude must be between -90 and 90")
		}
		if p.Longitude < -180 || p.Longitude > 180 {
			return nil, fmt.Errorf("longitude must be between -180 and 180")
		}

		result, err := client.SendLocation(context.Background(), telegram.SendLocationParams{
			Peer:     p.Peer,
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send location: %w", err)
		}

		return result, nil
	}
}
