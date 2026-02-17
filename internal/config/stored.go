// Package config provides configuration management.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// StoredConfig represents saved configuration in config.json.
type StoredConfig struct {
	AppID   int    `json:"app_id"`
	AppHash string `json:"app_hash"`
}

// ConfigPath returns the path to config.json.
func ConfigPath() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "", fmt.Errorf("HOME environment variable not set")
	}
	return filepath.Join(homeDir, ".agent-telegram", "config.json"), nil
}

// SaveConfig saves appID and appHash to config.json.
func SaveConfig(appID int, appHash string) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := StoredConfig{
		AppID:   appID,
		AppHash: appHash,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// LoadStoredConfig loads configuration from config.json.
// Falls back to TELEGRAM_APP_ID and TELEGRAM_APP_HASH env vars (for Docker/Coolify).
func LoadStoredConfig() (*StoredConfig, error) {
	// Try env vars first (for Docker/stateless deployments)
	if cfg, ok := loadConfigFromEnv(); ok {
		return cfg, nil
	}

	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"config not found - run 'agent-telegram login' or set TELEGRAM_APP_ID and TELEGRAM_APP_HASH",
			)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg StoredConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.AppID == 0 || cfg.AppHash == "" {
		return nil, fmt.Errorf("invalid config - please run 'agent-telegram login' first")
	}

	return &cfg, nil
}

// loadConfigFromEnv tries to load app credentials from environment variables.
func loadConfigFromEnv() (*StoredConfig, bool) {
	appIDStr := os.Getenv("TELEGRAM_APP_ID")
	appHash := os.Getenv("TELEGRAM_APP_HASH")
	if appIDStr == "" || appHash == "" {
		return nil, false
	}

	appID, err := strconv.Atoi(appIDStr)
	if err != nil || appID == 0 {
		return nil, false
	}

	return &StoredConfig{AppID: appID, AppHash: appHash}, true
}
