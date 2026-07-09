package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const pendingSessionFile = "pending-session.bin"

func pendingSessionPath() (string, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(configPath), pendingSessionFile), nil
}

// SavePendingSession stores a one-time owner-only handoff for a daemon that is
// started after authentication. The daemon deletes it immediately after read.
func SavePendingSession(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("session is empty")
	}
	path, err := pendingSessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".pending-session-*")
	if err != nil {
		return fmt.Errorf("create session handoff: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write session handoff: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close session handoff: %w", err)
	}
	_ = os.Remove(path) // Windows rename does not replace an existing file.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("publish session handoff: %w", err)
	}
	return nil
}

// ConsumePendingSession reads and removes a one-time session handoff.
func ConsumePendingSession() ([]byte, error) {
	path, err := pendingSessionPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read session handoff: %w", err)
	}
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("remove session handoff: %w", err)
	}
	return data, nil
}
