// Package paths provides common path utilities for agent-telegram.
package paths

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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
	return LogFilePathForSocket("")
}

// LogFilePathForSocket returns the server log path for a socket instance.
func LogFilePathForSocket(socketPath string) (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceFileName("server", socketPath, "log")), nil
}

// CLILogFilePath returns the path to the CLI log file.
func CLILogFilePath() (string, error) {
	return CLILogFilePathForSocket("")
}

// CLILogFilePathForSocket returns the CLI log path for a socket instance.
func CLILogFilePathForSocket(socketPath string) (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceFileName("cli", socketPath, "log")), nil
}

// AuditFilePath returns the path to the audit journal.
func AuditFilePath() (string, error) {
	return AuditFilePathForSocket("")
}

// AuditFilePathForSocket returns the audit journal path for a socket instance.
func AuditFilePathForSocket(socketPath string) (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceFileName("audit", socketPath, "jsonl")), nil
}

// PIDFilePath returns the path to the PID file.
func PIDFilePath() (string, error) {
	return PIDFilePathForSocket("")
}

// PIDFilePathForSocket returns the PID file path for a socket instance.
func PIDFilePathForSocket(socketPath string) (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceFileName("server", socketPath, "pid")), nil
}

// LockFilePath returns the path to the lock file.
func LockFilePath() (string, error) {
	return LockFilePathForSocket("")
}

// LockFilePathForSocket returns the lock file path for a socket instance.
func LockFilePathForSocket(socketPath string) (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceFileName("server", socketPath, "lock")), nil
}

func instanceFileName(prefix, socketPath, ext string) string {
	if isDefaultSocket(socketPath) {
		return fmt.Sprintf("%s.%s", prefix, ext)
	}
	sum := sha256.Sum256([]byte(socketPath))
	key := hex.EncodeToString(sum[:])[:12]
	return fmt.Sprintf("%s-%s.%s", prefix, key, ext)
}

func isDefaultSocket(socketPath string) bool {
	return socketPath == "" || socketPath == DefaultSocketPath
}

// LockFile represents a file-based lock.
type LockFile struct {
	path string
	file *os.File
}

// NewLockFile creates a new lock file instance.
func NewLockFile(path string) *LockFile {
	return &LockFile{path: path}
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
