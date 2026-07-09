//go:build darwin && cgo

package sessionstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestKeychainRoundTripOptIn(t *testing.T) {
	if os.Getenv("AGENT_TELEGRAM_TEST_KEYCHAIN") != "1" {
		t.Skip("set AGENT_TELEGRAM_TEST_KEYCHAIN=1 to exercise the login Keychain")
	}
	storage, err := Open(KeychainProvider, fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = storage.Delete(context.Background()) }()
	if err := storage.StoreSession(context.Background(), []byte("native-keychain-session")); err != nil {
		t.Fatal(err)
	}
	data, err := storage.LoadSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "native-keychain-session" {
		t.Fatalf("session = %q", data)
	}
}
