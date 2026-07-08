package authflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateStoreCreateLoadDelete(t *testing.T) {
	dir := t.TempDir()
	store := NewStateStore(dir)

	state, err := store.Create("+15551234567", "hash", 123, "app-hash", []byte("session-data"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if state.ID == "" {
		t.Fatal("state ID should be set")
	}

	info, err := os.Stat(filepath.Join(dir, state.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("state file mode = %v, want 0600", got)
	}

	loaded, err := store.Load(state.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Phone != state.Phone || loaded.PhoneCodeHash != state.PhoneCodeHash {
		t.Fatalf("loaded state mismatch: %+v", loaded)
	}
	sessionData, err := loaded.SessionData()
	if err != nil {
		t.Fatal(err)
	}
	if string(sessionData) != "session-data" {
		t.Fatalf("loaded session = %q", sessionData)
	}

	if err := store.Delete(state.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(state.ID); err == nil {
		t.Fatal("deleted state should not load")
	}
}

func TestStateStoreRejectsExpiredState(t *testing.T) {
	dir := t.TempDir()
	store := NewStateStore(dir)
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	state, err := store.Create("+15551234567", "hash", 123, "app-hash", []byte("session-data"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	store.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := store.Load(state.ID); err == nil {
		t.Fatal("expired state should be rejected")
	}
	if _, err := os.Stat(filepath.Join(dir, state.ID+".json")); !os.IsNotExist(err) {
		t.Fatalf("expired state file should be removed, stat err = %v", err)
	}
}

func TestStateStoreRejectsUnsafeIDs(t *testing.T) {
	store := NewStateStore(t.TempDir())
	for _, id := range []string{"../secret", "short", "bad/id/0000000000000000", "bad$id000000000000"} {
		if _, err := store.Load(id); err == nil {
			t.Fatalf("Load(%q) should reject unsafe id", id)
		}
	}
}
