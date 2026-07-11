package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

type fileFingerprint struct {
	exists bool
	digest [sha256.Size]byte
}

// ReloadingEnforcer applies valid policy file edits before each policy check.
// Its current Enforcer is an immutable snapshot, and invalid replacements never
// displace the last valid snapshot.
type ReloadingEnforcer struct {
	path        string
	resolver    PeerResolver
	mu          sync.Mutex
	current     *Enforcer
	attempted   fileFingerprint
	hasAttempt  bool
	lastFailure string
}

// NewReloadingEnforcer creates an enforcer backed by a reloadable policy file.
func NewReloadingEnforcer(path string, resolver PeerResolver) *ReloadingEnforcer {
	e := &ReloadingEnforcer{
		path:     path,
		resolver: resolver,
		current:  NewEnforcer(Default(), resolver),
	}
	e.refresh()
	return e
}

// Check refreshes the policy snapshot, then validates the request against it.
func (e *ReloadingEnforcer) Check(ctx context.Context, method string, params json.RawMessage) error {
	if e == nil {
		return nil
	}
	e.refresh()
	e.mu.Lock()
	current := e.current
	e.mu.Unlock()
	return current.Check(ctx, method, params)
}

func (e *ReloadingEnforcer) refresh() {
	// Serialize the read and publication so an older concurrent read cannot be
	// published after a newer snapshot.
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := os.ReadFile(e.path)
	if err != nil && !os.IsNotExist(err) {
		e.logReadFailure(err)
		return
	}

	fingerprint := fileFingerprint{exists: err == nil}
	if err == nil {
		fingerprint.digest = sha256.Sum256(data)
	}
	// A readable state ends any transient read-failure state, even when the
	// content fingerprint itself has not changed.
	e.lastFailure = ""
	if e.hasAttempt && e.attempted == fingerprint {
		return
	}
	e.attempted = fingerprint
	e.hasAttempt = true

	p := Default()
	if fingerprint.exists {
		p, err = Parse(data)
		if err != nil {
			e.lastFailure = err.Error()
			slog.Warn("policy reload failed",
				"exists", true,
				"digest", fingerprint.digestString(),
				"error", err,
			)
			return
		}
	}

	e.current = NewEnforcer(p, e.resolver)
	e.lastFailure = ""
	slog.Info("policy reloaded",
		"exists", fingerprint.exists,
		"digest", fingerprint.digestString(),
	)
}

func (e *ReloadingEnforcer) logReadFailure(err error) {
	failure := fmt.Sprintf("read policy: %v", err)
	if e.lastFailure == failure {
		return
	}
	e.lastFailure = failure
	slog.Warn("policy reload failed", "error", failure)
}

func (f fileFingerprint) digestString() string {
	if !f.exists {
		return ""
	}
	return hex.EncodeToString(f.digest[:])
}
