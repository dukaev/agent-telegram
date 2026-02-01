//go:build !windows

package paths

import (
	"fmt"
	"os"
	"syscall"
)

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
