package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/sessionstore"
	"agent-telegram/internal/types"
)

type fakeAuthBackend struct {
	sessionData     []byte
	importedSession []byte
	sendResult      *types.SendCodeResult
	signResult      *types.SignInResult
	passResult      *types.SignInResult
	sentPhone       string
	signedCode      string
	password        string
}

type recordingSessionStore struct {
	data map[string][]byte
}

func (s *recordingSessionStore) Provider() string { return "auth-test-persistent" }
func (s *recordingSessionStore) Persistent() bool { return true }
func (s *recordingSessionStore) Load(_ context.Context, profile string) ([]byte, error) {
	data := s.data[profile]
	if len(data) == 0 {
		return nil, sessionstore.ErrNotFound
	}
	return append([]byte(nil), data...), nil
}
func (s *recordingSessionStore) Save(_ context.Context, profile string, data []byte) error {
	s.data[profile] = append([]byte(nil), data...)
	return nil
}
func (s *recordingSessionStore) Delete(_ context.Context, profile string) error {
	delete(s.data, profile)
	return nil
}

var authTestPersistentStore = &recordingSessionStore{data: make(map[string][]byte)}

func init() {
	sessionstore.RegisterProvider("auth-test-persistent", func() (sessionstore.Store, error) {
		return authTestPersistentStore, nil
	})
}

func (f *fakeAuthBackend) SendCode(_ context.Context, phone string) (*types.SendCodeResult, error) {
	f.sentPhone = phone
	return f.sendResult, nil
}

func (f *fakeAuthBackend) SignIn(_ context.Context, _ string, code, _ string) (*types.SignInResult, error) {
	f.signedCode = code
	return f.signResult, nil
}

func (f *fakeAuthBackend) SignInWith2FA(_ context.Context, _ string, password string) (*types.SignInResult, error) {
	f.password = password
	return f.passResult, nil
}

func (f *fakeAuthBackend) SignInWithQR(
	_ context.Context,
	onToken func(tokenURL string, expiresAt time.Time) error,
) (*types.SignInResult, error) {
	if onToken != nil {
		if err := onToken("tg://login?token=a", time.Time{}); err != nil {
			return nil, err
		}
	}
	return f.signResult, nil
}

func (f *fakeAuthBackend) ImportSession(_ context.Context, data []byte) error {
	f.importedSession = append([]byte(nil), data...)
	return nil
}

func (f *fakeAuthBackend) ExportSession(_ context.Context) ([]byte, error) {
	if len(f.sessionData) == 0 {
		return []byte("fake-session"), nil
	}
	return append([]byte(nil), f.sessionData...), nil
}

func TestFinishAuthPersistsSelectedProvider(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	t.Setenv(sessionstore.EnvProvider, "auth-test-persistent")
	t.Setenv(sessionstore.EnvProfile, "work")
	delete(authTestPersistentStore.data, "work")
	state := createTestState(t)

	body, err := finishAuth(&cobra.Command{}, authRuntimeFromGlobals(), state)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(authTestPersistentStore.data["work"]); got != "state-session" {
		t.Fatalf("persisted session = %q", got)
	}
	if body["sessionStorage"] != "auth-test-persistent" || body["sessionPersistent"] != true {
		t.Fatalf("finish body = %+v", body)
	}
	stored, err := config.LoadStoredConfig()
	if err != nil {
		t.Fatal(err)
	}
	if stored.SessionProvider != "auth-test-persistent" || stored.SessionProfile != "work" {
		t.Fatalf("stored selection = %+v", stored)
	}
}

func TestAuthRuntimeConfigBuildsTelegramConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("TELEGRAM_APP_ID", "999")
	t.Setenv("TELEGRAM_APP_HASH", "env-hash")
	runtime := authRuntimeConfig{
		AppID:   "456",
		AppHash: "runtime-hash",
		Phone:   "+100200300",
	}

	cfg, err := runtime.authConfig(runtime.Phone)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppID != 456 || cfg.AppHash != "runtime-hash" || cfg.Phone != "+100200300" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.SessionPath != filepath.Join(tmp, ".agent-telegram") {
		t.Fatalf("session path = %q", cfg.SessionPath)
	}
}

func resetAuthGlobals(t *testing.T, home string) {
	t.Helper()
	oldBackend := newAuthBackend
	t.Cleanup(func() {
		newAuthBackend = oldBackend
	})

	t.Setenv("HOME", home)
	t.Setenv("TELEGRAM_APP_ID", "")
	t.Setenv("TELEGRAM_APP_HASH", "")
	t.Setenv("AGENT_TELEGRAM_PHONE", "")
	t.Setenv(sessionstore.EnvProvider, sessionstore.MemoryProvider)
	t.Setenv(sessionstore.EnvProfile, sessionstore.DefaultProfile)

	authAppID = "123"
	authAppHash = "app-hash"
	authPhone = "+88806283792"
	authStateDir = filepath.Join(home, ".agent-telegram", "auth-state")
	authStateTTL = time.Minute
	authReload = false
	authWebQR = true
	authWebMock = false
	authWebMockSaved = false
}

func createTestState(t *testing.T) *authflow.State {
	t.Helper()
	store := authflow.NewStateStore(authStateDir)
	state, err := store.Create(
		"+88806283792",
		"hash",
		123,
		"app-hash",
		[]byte("state-session"),
		time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	return state
}
