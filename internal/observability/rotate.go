package observability

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

// RotatingWriter rotates while a long-running daemon is still active.
type RotatingWriter struct {
	mu   sync.Mutex
	path string
	file *os.File
}

var _ io.Writer = (*RotatingWriter)(nil)

// NewRotatingWriter opens a writer with size-based retention.
func NewRotatingWriter(path string) (*RotatingWriter, error) {
	file, err := OpenAppendLog(path)
	if err != nil {
		return nil, err
	}
	return &RotatingWriter{path: path, file: file}, nil
}

// Write implements io.Writer.
func (w *RotatingWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, os.ErrClosed
	}
	if info, err := w.file.Stat(); err == nil && info.Size()+int64(len(data)) >= configuredLogMaxBytes() {
		if err := w.file.Close(); err != nil {
			return 0, err
		}
		if err := rotateIfNeeded(w.path, 1, defaultLogBackups); err != nil {
			return 0, err
		}
		file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return 0, err
		}
		w.file = file
	}
	return w.file.Write(data)
}

const (
	defaultLogMaxBytes = int64(10 << 20)
	defaultLogBackups  = 3
	logMaxBytesEnv     = "AGENT_TELEGRAM_LOG_MAX_BYTES"
)

// OpenAppendLog rotates an oversized file and opens the active generation.
func OpenAppendLog(path string) (*os.File, error) {
	if err := rotateIfNeeded(path, configuredLogMaxBytes(), defaultLogBackups); err != nil {
		return nil, err
	}
	//nolint:gosec // callers provide paths under the owner-only config directory.
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

func rotateIfNeeded(path string, maxBytes int64, backups int) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if maxBytes <= 0 || info.Size() < maxBytes {
		return nil
	}
	for generation := backups - 1; generation >= 1; generation-- {
		from := fmt.Sprintf("%s.%d", path, generation)
		to := fmt.Sprintf("%s.%d", path, generation+1)
		_ = os.Rename(from, to)
	}
	return os.Rename(path, path+".1")
}

func configuredLogMaxBytes() int64 {
	if raw := os.Getenv(logMaxBytesEnv); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			return value
		}
	}
	return defaultLogMaxBytes
}
