package sessionstore

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/gotd/td/session"
)

func TestMemoryProviderRoundTrip(t *testing.T) {
	storage, err := Open(MemoryProvider, "test-profile")
	if err != nil {
		t.Fatal(err)
	}
	if storage.Persistent() {
		t.Fatal("memory provider must not report persistence")
	}
	if _, err := storage.LoadSession(context.Background()); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("empty load error = %v", err)
	}
	if err := storage.StoreSession(context.Background(), []byte("session")); err != nil {
		t.Fatal(err)
	}
	data, err := storage.LoadSession(context.Background())
	if err != nil || string(data) != "session" {
		t.Fatalf("loaded session = %q, %v", data, err)
	}
	data[0] = 'X'
	if got := string(storage.ExportSession()); got != "session" {
		t.Fatalf("provider did not retain a private copy: %q", got)
	}
	if err := storage.Delete(context.Background()); err != nil {
		t.Fatal(err)
	}
	if data := storage.ExportSession(); data != nil {
		t.Fatalf("deleted session = %q", data)
	}
}

func TestDefaultSelectionUsesEnvironment(t *testing.T) {
	t.Setenv(EnvProvider, MemoryProvider)
	t.Setenv(EnvProfile, "work")
	storage, err := OpenDefault()
	if err != nil {
		t.Fatal(err)
	}
	if selection := storage.Selection(); selection.Provider != MemoryProvider || selection.Profile != "work" {
		t.Fatalf("selection = %+v", selection)
	}
}

func TestProviderRegistryAndProfileValidation(t *testing.T) {
	if !slices.Contains(ProviderNames(), MemoryProvider) {
		t.Fatalf("memory provider is not registered: %v", ProviderNames())
	}
	if _, err := Open("missing", "default"); err == nil {
		t.Fatal("unknown provider should fail")
	}
	if _, err := Open(MemoryProvider, "../escape"); err == nil {
		t.Fatal("unsafe profile should fail")
	}
}
