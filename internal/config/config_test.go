package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigValidationAndClone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := New(1, "hash", "+1", "")
	if !strings.HasSuffix(cfg.SessionPath, ".agent-telegram") {
		t.Fatalf("session path = %q", cfg.SessionPath)
	}
	if cfg.SessionStorage() == nil {
		t.Fatal("session storage should be present")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid config = %v", err)
	}
	if err := (&Config{}).Validate(); err == nil {
		t.Fatal("missing app id should fail")
	}
	if err := (&Config{AppID: 1}).Validate(); err == nil {
		t.Fatal("missing app hash should fail")
	}
	if err := (&Config{AppID: 1, AppHash: "hash"}).Validate(); err == nil {
		t.Fatal("missing phone should fail")
	}
	if err := (&Config{AppID: 1, AppHash: "hash"}).ValidateOptional(); err != nil {
		t.Fatalf("optional phone validation = %v", err)
	}
	withPhone := cfg.WithPhone("+2")
	if withPhone.Phone != "+2" || cfg.Phone != "+1" {
		t.Fatalf("WithPhone mutated config: old=%+v new=%+v", cfg, withPhone)
	}
	clone := cfg.Clone()
	if clone == cfg || clone.AppHash != cfg.AppHash {
		t.Fatalf("clone = %+v", clone)
	}
}

func TestLoadersAndEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("AGENT_TELEGRAM_APP_ID", "123")
	t.Setenv("AGENT_TELEGRAM_APP_HASH", "hash")
	t.Setenv("AGENT_TELEGRAM_PHONE", "+1")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppID != 123 || cfg.AppHash != "hash" || cfg.Phone != "+1" {
		t.Fatalf("env config = %+v", cfg)
	}
	optional, err := LoadFromEnvWithOptionalPhone()
	if err != nil || optional.Phone != "+1" {
		t.Fatalf("optional env = %+v, %v", optional, err)
	}
	args := LoadFromArgs(2, "hash2", "+2", "")
	if args.AppID != 2 || args.Phone != "+2" || !strings.HasSuffix(args.SessionPath, ".agent-telegram") {
		t.Fatalf("args config = %+v", args)
	}
	if got := GetEnv("NO_SUCH_KEY", "AGENT_TELEGRAM_PHONE"); got != "+1" {
		t.Fatalf("GetEnv = %q", got)
	}
	if id, err := ParseAppID("42"); err != nil || id != 42 {
		t.Fatalf("ParseAppID = %d, %v", id, err)
	}
	for _, value := range []string{"", "abc"} {
		if _, err := ParseAppID(value); err == nil {
			t.Fatalf("ParseAppID(%q) should fail", value)
		}
	}
}

func TestFileAndStoredConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if path, err := ConfigPath(); err != nil || path != filepath.Join(home, ".agent-telegram", "config.json") {
		t.Fatalf("ConfigPath = %q, %v", path, err)
	}
	for _, path := range []string{"a.json", "a.yaml", "a.yml"} {
		if !IsSupportedFile(path) {
			t.Fatalf("%s should be supported", path)
		}
	}
	if IsSupportedFile("a.toml") {
		t.Fatal("toml should not be supported")
	}
	if _, err := parserForFile("a.toml"); err == nil {
		t.Fatal("unsupported parser should fail")
	}

	jsonPath := filepath.Join(home, "config.json")
	if err := os.WriteFile(jsonPath, []byte(`{"app_id":7,"app_hash":"file","phone":"+7"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppID != 7 || cfg.Phone != "+7" {
		t.Fatalf("file config = %+v", cfg)
	}

	if err := SaveConfig(9, "stored"); err != nil {
		t.Fatal(err)
	}
	stored, err := LoadStoredConfig()
	if err != nil {
		t.Fatal(err)
	}
	if stored.AppID != 9 || stored.AppHash != "stored" {
		t.Fatalf("stored = %+v", stored)
	}

	t.Setenv("TELEGRAM_APP_ID", "10")
	t.Setenv("TELEGRAM_APP_HASH", "envhash")
	stored, err = LoadStoredConfig()
	if err != nil || stored.AppID != 10 || stored.AppHash != "envhash" {
		t.Fatalf("env stored = %+v, %v", stored, err)
	}
	t.Setenv("HOME", "")
	if _, err := ConfigPath(); err == nil {
		t.Fatal("missing HOME should fail")
	}
}
