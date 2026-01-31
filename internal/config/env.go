// Package config provides configuration management for the agent-telegram application.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/env"
)

const (
	// Prefix is the environment variable prefix.
	prefix = "AGENT_TELEGRAM_"
	// Delimiter is the delimiter for nested config keys.
	delim = "."
)

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

// loadFromEnv loads configuration from environment variables into a koanf instance.
func loadFromEnv(k *koanf.Koanf) error {
	return k.Load(env.Provider(prefix, delim, func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)
}

// LoadFromArgs loads configuration from command-line arguments.
func LoadFromArgs(appID int, appHash, phone, sessionPath string) *Config {
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

// GetEnv returns the first non-empty environment variable from the given keys.
func GetEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// ParseAppID parses app ID from string.
func ParseAppID(appIDStr string) (int, error) {
	if appIDStr == "" {
		return 0, fmt.Errorf("app_id cannot be empty")
	}
	var appID int
	_, err := fmt.Sscanf(appIDStr, "%d", &appID)
	if err != nil {
		return 0, fmt.Errorf("invalid app_id format: %w", err)
	}
	return appID, nil
}
