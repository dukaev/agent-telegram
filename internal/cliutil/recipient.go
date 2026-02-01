// Package cliutil provides shared CLI utilities.
package cliutil

import (
	"fmt"
	"strings"
)

// Recipient represents a Telegram peer (user, chat, or channel).
// Accepts: @username, username (without @), or chat ID (numeric).
// Normalizes all inputs to the format expected by the Telegram API.
type Recipient struct {
	value string
}

// String returns the raw value.
func (r *Recipient) String() string {
	return r.value
}

// Set implements pflag.Value interface.
func (r *Recipient) Set(s string) error {
	if s == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	r.value = s
	return nil
}

// Type implements pflag.Value interface.
func (r *Recipient) Type() string {
	return "recipient"
}

// Peer returns normalized peer for API.
// @user → @user
// username → @username
// 123456789 → 123456789 (user ID)
// -123456789 → -123456789 (chat/channel ID)
func (r *Recipient) Peer() string {
	if r.value == "" {
		return ""
	}
	if strings.HasPrefix(r.value, "@") {
		return r.value
	}
	// Check if it's a numeric ID (positive or negative)
	if r.value[0] >= '0' && r.value[0] <= '9' {
		return r.value
	}
	if r.value[0] == '-' && len(r.value) > 1 && r.value[1] >= '0' && r.value[1] <= '9' {
		return r.value
	}
	return "@" + r.value
}

// AddToParams adds normalized peer to parameters.
func (r *Recipient) AddToParams(params map[string]any) {
	params["peer"] = r.Peer()
}
