// Package config provides configuration management for the agent-telegram application.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/v2"
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

// parseConfig maps koanf instance to Config struct.
func parseConfig(k *koanf.Koanf) (*Config, error) {
	appID := k.Int("app_id")
	appHash := k.String("app_hash")
	phone := k.String("phone")
	sessionPath := k.String("session_path")

	// Default session path
	if sessionPath == "" {
		sessionPath = filepath.Join(os.Getenv("HOME"), ".agent-telegram")
	}

	if phone == "" {
		return nil, fmt.Errorf("missing required field: phone")
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
