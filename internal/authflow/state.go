package authflow

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultStateTTL is the lifetime of an unfinished auth flow.
	DefaultStateTTL = 15 * time.Minute
	stateVersion    = 1
)

// State contains the temporary data needed between auth steps.
type State struct {
	Version       int       `json:"version"`
	ID            string    `json:"id"`
	Phone         string    `json:"phone"`
	PhoneCodeHash string    `json:"phoneCodeHash"`
	AppID         int       `json:"appId"`
	AppHash       string    `json:"appHash"`
	SessionPath   string    `json:"sessionPath"`
	Requires2FA   bool      `json:"requires2FA,omitempty"`
	TwoFactorHint string    `json:"twoFactorHint,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// Expired reports whether the state has expired at now.
func (s State) Expired(now time.Time) bool {
	return !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt)
}

// StateStore persists temporary auth state to disk.
type StateStore struct {
	dir string
	now func() time.Time
}

// NewStateStore creates a state store in dir.
func NewStateStore(dir string) *StateStore {
	return &StateStore{dir: dir, now: time.Now}
}

// DefaultStateDir returns the default auth state directory.
func DefaultStateDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "."
	}
	return filepath.Join(home, ".agent-telegram", "auth-state")
}

// Create saves a new state and returns it.
func (s *StateStore) Create(phone, codeHash string, appID int, appHash, sessionPath string, ttl time.Duration) (*State, error) {
	if ttl <= 0 {
		ttl = DefaultStateTTL
	}
	id, err := newStateID()
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	state := &State{
		Version:       stateVersion,
		ID:            id,
		Phone:         phone,
		PhoneCodeHash: codeHash,
		AppID:         appID,
		AppHash:       appHash,
		SessionPath:   sessionPath,
		CreatedAt:     now,
		ExpiresAt:     now.Add(ttl),
	}
	if err := s.Save(state); err != nil {
		return nil, err
	}
	return state, nil
}

// Save writes state to disk with owner-only permissions.
func (s *StateStore) Save(state *State) error {
	if err := validateStateID(state.ID); err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("create auth state dir: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal auth state: %w", err)
	}
	if err := os.WriteFile(s.path(state.ID), data, 0600); err != nil {
		return fmt.Errorf("write auth state: %w", err)
	}
	return nil
}

// Load reads state from disk and rejects expired state.
func (s *StateStore) Load(id string) (*State, error) {
	if err := validateStateID(id); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("auth state not found")
		}
		return nil, fmt.Errorf("read auth state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse auth state: %w", err)
	}
	if state.Version != stateVersion {
		return nil, fmt.Errorf("unsupported auth state version: %d", state.Version)
	}
	if state.ID != id {
		return nil, fmt.Errorf("auth state id mismatch")
	}
	if state.Expired(s.now().UTC()) {
		_ = s.Delete(id)
		return nil, fmt.Errorf("auth state expired")
	}
	return &state, nil
}

// Delete removes a state file.
func (s *StateStore) Delete(id string) error {
	if err := validateStateID(id); err != nil {
		return err
	}
	if err := os.Remove(s.path(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete auth state: %w", err)
	}
	return nil
}

func (s *StateStore) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func newStateID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate auth state id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func validateStateID(id string) error {
	if len(id) < 16 || len(id) > 64 {
		return fmt.Errorf("invalid auth state id")
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("invalid auth state id")
	}
	for _, r := range id {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf("invalid auth state id")
	}
	return nil
}
