// Package paths provides common path utilities for agent-telegram.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const (
	// DefaultSocketPath is the default Unix socket path.
	DefaultSocketPath = "/tmp/agent-telegram.sock"

	// ConfigDirName is the name of the config directory.
	ConfigDirName = ".agent-telegram"
)

// ConfigDir returns the path to the config directory (~/.agent-telegram).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	return dir, nil
}

// LogFilePath returns the path to the log file.
func LogFilePath() (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "server.log"), nil
}

// PIDFilePath returns the path to the PID file.
func PIDFilePath() (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "server.pid"), nil
}

// LockFilePath returns the path to the lock file.
func LockFilePath() (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "server.lock"), nil
}

// LockFile represents a file-based lock using flock.
type LockFile struct {
	path string
	file *os.File
}

// NewLockFile creates a new lock file instance.
func NewLockFile(path string) *LockFile {
	return &LockFile{path: path}
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns true if lock was acquired, false if already locked by another process.
func (l *LockFile) TryLock() (bool, error) {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return false, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock without blocking
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		_ = file.Close()
		if err == syscall.EWOULDBLOCK {
			return false, nil // Already locked
		}
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.file = file
	return true, nil
}

// Unlock releases the lock.
func (l *LockFile) Unlock() error {
	if l.file == nil {
		return nil
	}
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}
	l.file = nil
	return nil
}

// WritePID writes the current process PID to a file.
func WritePID(path string) error {
	return os.WriteFile(path, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0600)
}

// ReadPID reads the PID from a file.
func ReadPID(path string) (int, error) {
	//nolint:gosec // path is from trusted PIDFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, fmt.Errorf("invalid PID file content: %w", err)
	}
	return pid, nil
}

// RemovePID removes the PID file.
func RemovePID(path string) error {
	return os.Remove(path)
}
