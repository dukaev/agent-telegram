// Package callback provides a callback API server for Telegram updates.
package callback

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State holds the persisted callback configuration.
type State struct {
	CallbackURL      string         `json:"callbackUrl"`
	VerifyCode       string         `json:"verifyCode"`
	Verified         bool           `json:"verified"`
	Subscriptions    []Subscription `json:"subscriptions"`
	NextSubID        int64          `json:"nextSubId"`
	LastErrorDate    *int64         `json:"lastErrorDate"`
	LastErrorMessage *string        `json:"lastErrorMessage"`
}

// Subscription represents a channel subscription.
type Subscription struct {
	ID         int64    `json:"subscriptionId"`
	Type       string   `json:"type"`       // "channel"
	ChannelID  string   `json:"channelId"`  // "@username" or resolved peer
	EventTypes []string `json:"eventTypes"` // ["new_post", "edit_post"]
	CreatedAt  int64    `json:"createdAt"`
}

// Store persists the callback state to a JSON file.
type Store struct {
	mu   sync.RWMutex
	path string
	data State
}

// NewStore creates a Store backed by the given file path.
// It loads existing state if the file exists.
func NewStore(dir string) (*Store, error) {
	s := &Store{
		path: filepath.Join(dir, "callback.json"),
		data: State{NextSubID: 1},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil // fresh state
	}
	if err != nil {
		return fmt.Errorf("read callback state: %w", err)
	}
	return json.Unmarshal(data, &s.data)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal callback state: %w", err)
	}
	return os.WriteFile(s.path, data, 0600)
}

// Get returns a copy of the current state.
func (s *Store) Get() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// SetCallbackURL sets a new (unverified) callback URL and returns the verify code.
func (s *Store) SetCallbackURL(url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, err := generateVerifyCode()
	if err != nil {
		return "", err
	}

	s.data.CallbackURL = url
	s.data.VerifyCode = code
	s.data.Verified = false

	return code, s.save()
}

// MarkVerified marks the current callback URL as verified.
func (s *Store) MarkVerified() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Verified = true
	return s.save()
}

// AddSubscription adds a channel subscription and returns its ID.
func (s *Store) AddSubscription(sub Subscription) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub.ID = s.data.NextSubID
	sub.CreatedAt = time.Now().Unix()
	s.data.NextSubID++
	s.data.Subscriptions = append(s.data.Subscriptions, sub)

	return sub.ID, s.save()
}

// RemoveSubscription removes a subscription by ID. Returns error if not found.
func (s *Store) RemoveSubscription(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.data.Subscriptions {
		if sub.ID == id {
			s.data.Subscriptions = append(s.data.Subscriptions[:i], s.data.Subscriptions[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("subscription not found")
}

// RecordError stores the last delivery error.
func (s *Store) RecordError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	s.data.LastErrorDate = &now
	s.data.LastErrorMessage = &msg
	_ = s.save()
}

// generateVerifyCode generates a random 6-digit suffix for the verify code.
func generateVerifyCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("generate verify code: %w", err)
	}
	return fmt.Sprintf("AGENT_TG_VERIFY_%06d", n.Int64()+100000), nil
}
