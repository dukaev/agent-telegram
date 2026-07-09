package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestCreateTelegramClientDefaultsToMemory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(envTelegramSession, "")

	tgClient := createTelegramClient(123, "app-hash", telegramClientOptions{})
	imported, err := tgClient.ImportSession(context.Background(), []byte("memory-session"))
	if err != nil {
		t.Fatal(err)
	}
	if !imported {
		t.Fatal("client should use memory storage by default")
	}
	if got := string(tgClient.ExportSession()); got != "memory-session" {
		t.Fatalf("exported session = %q", got)
	}
}

func TestCreateTelegramClientUsesEnvSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(envTelegramSession, base64.StdEncoding.EncodeToString([]byte("env-session")))

	tgClient := createTelegramClient(123, "app-hash", telegramClientOptions{})
	if got := string(tgClient.ExportSession()); got != "env-session" {
		t.Fatalf("exported session = %q", got)
	}
}

func TestImportSessionForMemoryStorage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(envTelegramSession, "")
	tgClient := createTelegramClient(123, "app-hash", telegramClientOptions{})

	if err := importSessionForMemoryStorage(tgClient, []byte("fresh-auth-session")); err != nil {
		t.Fatal(err)
	}
	if got := string(tgClient.ExportSession()); got != "fresh-auth-session" {
		t.Fatalf("exported session = %q", got)
	}
}

func TestParseReloadSessionData(t *testing.T) {
	raw, err := json.Marshal(map[string]string{
		"session": base64.StdEncoding.EncodeToString([]byte("reload-session")),
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := parseReloadSessionData(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "reload-session" {
		t.Fatalf("decoded session = %q", got)
	}
}

func TestLogoutTelegramClientAllowsNilClient(_ *testing.T) {
	logoutTelegramClient(nil, true)
}
