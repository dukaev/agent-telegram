//go:build windows

package paths

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns true if lock was acquired, false if already locked by another process.
func (l *LockFile) TryLock() (bool, error) {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return false, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock without blocking using Windows API
	err = windows.LockFileEx(
		windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&windows.Overlapped{},
	)
	if err != nil {
		_ = file.Close()
		if err == windows.ERROR_LOCK_VIOLATION {
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
	err := windows.UnlockFileEx(
		windows.Handle(l.file.Fd()),
		0,
		1,
		0,
		&windows.Overlapped{},
	)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}
	l.file = nil
	return nil
}
