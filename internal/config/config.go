// Package config provides configuration management for the agent-telegram application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/session"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
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

const (
	prefix = "AGENT_TELEGRAM_"
	delim  = "."

	// Hardcoded Telegram API credentials
	defaultAppID   = 23027031
	defaultAppHash = "1c6f5d81f2a754e0d2f0e3f06b1cbe17"
)

// Load loads configuration from multiple sources in priority order:
// 1. Config file (if provided)
// 2. Environment variables (AGENT_TELEGRAM_* prefix)
func Load(configPath string) (*Config, error) {
	k := koanf.New(delim)

	// Load from config file if provided
	if configPath != "" {
		if err := loadFromFile(k, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Load from environment variables (overrides config file)
	if err := loadFromEnv(k); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// Map koanf config to Config struct
	return parseConfig(k)
}

// LoadFromEnv loads configuration from environment variables.
// Returns an error if required environment variables are missing.
func LoadFromEnv() (*Config, error) {
	k := koanf.New(delim)
	if err := loadFromEnv(k); err != nil {
		return nil, err
	}
	return parseConfig(k)
}

// LoadFromEnvWithOptionalPhone loads configuration from environment variables.
// Phone number is optional and can be set later.
func LoadFromEnvWithOptionalPhone() (*Config, error) {
	k := koanf.New(delim)
	if err := loadFromEnv(k); err != nil {
		return nil, err
	}
	return parseConfigOptionalPhone(k)
}

// loadFromFile loads configuration from a file (JSON or YAML).
func loadFromFile(k *koanf.Koanf, path string) error {
	var parser koanf.Parser

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		parser = json.Parser()
	case ".yaml", ".yml":
		parser = yaml.Parser()
	default:
		return fmt.Errorf("unsupported config file format: %s", filepath.Ext(path))
	}

	return k.Load(file.Provider(path), parser)
}

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(k *koanf.Koanf) error {
	return k.Load(env.Provider(prefix, delim, func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)
}

// parseConfig maps koanf instance to Config struct.
func parseConfig(k *koanf.Koanf) (*Config, error) {
	appID := k.Int("app_id")
	appHash := k.String("app_hash")
	phone := k.String("phone")
	sessionPath := k.String("session_path")

	// Use hardcoded credentials as defaults
	if appID == 0 {
		appID = defaultAppID
	}
	if appHash == "" {
		appHash = defaultAppHash
	}
	if phone == "" {
		return nil, fmt.Errorf("missing required field: phone")
	}

	// Default session path
	if sessionPath == "" {
		sessionPath = filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	}

	return &Config{
		AppID:       appID,
		AppHash:     appHash,
		Phone:       phone,
		SessionPath: sessionPath,
	}, nil
}

// parseConfigOptionalPhone maps koanf instance to Config struct with optional phone.
func parseConfigOptionalPhone(k *koanf.Koanf) (*Config, error) {
	appID := k.Int("app_id")
	appHash := k.String("app_hash")
	phone := k.String("phone")
	sessionPath := k.String("session_path")

	// Use hardcoded credentials as defaults
	if appID == 0 {
		appID = defaultAppID
	}
	if appHash == "" {
		appHash = defaultAppHash
	}

	// Default session path
	if sessionPath == "" {
		sessionPath = filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	}

	return &Config{
		AppID:       appID,
		AppHash:     appHash,
		Phone:       phone,
		SessionPath: sessionPath,
	}, nil
}

// LoadFromArgs loads configuration from command-line arguments.
func LoadFromArgs(appID int, appHash, phone string, sessionPath string) *Config {
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
