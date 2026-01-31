// Package config provides configuration management for the agent-telegram application.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
)

// Config holds the application configuration.
type Config struct {
	// Telegram API credentials
	AppID   int
	AppHash string

	// User phone number
	Phone string

	// Session storage path
	SessionPath string
}

// New creates a new Config with the given parameters.
func New(appID int, appHash, phone, sessionPath string) *Config {
	if sessionPath == "" {
		sessionPath = filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	}

	return &Config{
		AppID:       appID,
		AppHash:     appHash,
		Phone:       phone,
		SessionPath: sessionPath,
	}
}

// SessionStorage returns a session storage for the given user ID.
func (c *Config) SessionStorage(userID int) session.Storage {
	sessionDir := filepath.Join(c.SessionPath, fmt.Sprintf("user_%d", userID))
	return &session.FileStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.AppID == 0 {
		return fmt.Errorf("app_id is required")
	}
	if c.AppHash == "" {
		return fmt.Errorf("app_hash is required")
	}
	if c.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	return nil
}

// ValidateOptional validates the configuration with optional phone.
func (c *Config) ValidateOptional() error {
	if c.AppID == 0 {
		return fmt.Errorf("app_id is required")
	}
	if c.AppHash == "" {
		return fmt.Errorf("app_hash is required")
	}
	return nil
}

// WithPhone returns a new config with the phone number set.
func (c *Config) WithPhone(phone string) *Config {
	return &Config{
		AppID:       c.AppID,
		AppHash:     c.AppHash,
		Phone:       phone,
		SessionPath: c.SessionPath,
	}
}

// Clone returns a deep copy of the config.
func (c *Config) Clone() *Config {
	return &Config{
		AppID:       c.AppID,
		AppHash:     c.AppHash,
		Phone:       c.Phone,
		SessionPath: c.SessionPath,
	}
}
