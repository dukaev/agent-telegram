package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/authflow"
	"agent-telegram/internal/config"
	"agent-telegram/internal/types"
)

type fakeAuthBackend struct {
	sessionPath string
	sendResult  *types.SendCodeResult
	signResult  *types.SignInResult
	passResult  *types.SignInResult
	sentPhone   string
	signedCode  string
	password    string
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

func (f *fakeAuthBackend) SessionPath() string {
	return f.sessionPath
}

func TestAuthBeginWritesStateAndJSON(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	backend := &fakeAuthBackend{
		sessionPath: filepath.Join(tmp, ".agent-telegram", "session.json"),
		sendResult:  &types.SendCodeResult{PhoneCodeHash: "hash", Timeout: 30},
	}
	newAuthBackend = func(_ *config.Config) authflow.Backend { return backend }

	output := captureStdout(t, func() {
		runAuthBegin(&cobra.Command{}, nil)
	})

	var body struct {
		OK      bool   `json:"ok"`
		Next    string `json:"next"`
		StateID string `json:"stateId"`
		Phone   string `json:"phone"`
		Timeout int    `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if !body.OK || body.Next != "code" || body.StateID == "" || body.Phone != "***3792" || body.Timeout != 30 {
		t.Fatalf("unexpected begin response: %s", output)
	}
	if backend.sentPhone != "+88806283792" {
		t.Fatalf("sent phone = %q", backend.sentPhone)
	}

	state, err := authflow.NewStateStore(authStateDir).Load(body.StateID)
	if err != nil {
		t.Fatal(err)
	}
	if state.PhoneCodeHash != "hash" {
		t.Fatalf("state hash = %q, want hash", state.PhoneCodeHash)
	}
}

func TestAuthVerifyCompletesSuccessfulLogin(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	authStateID = state.ID
	authCodeStdin = true
	authReload = false

	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
		signResult:  &types.SignInResult{Success: true},
	}
	newAuthBackend = func(_ *config.Config) authflow.Backend { return backend }

	withStdin(t, " 12345\n", func() {
		output := captureStdout(t, func() {
			runAuthVerify(&cobra.Command{}, nil)
		})
		var body struct {
			OK          bool   `json:"ok"`
			Next        string `json:"next"`
			SessionPath string `json:"sessionPath"`
		}
		if err := json.Unmarshal([]byte(output), &body); err != nil {
			t.Fatal(err)
		}
		if !body.OK || body.Next != "done" || body.SessionPath != state.SessionPath {
			t.Fatalf("unexpected verify response: %s", output)
		}
	})

	if backend.signedCode != "12345" {
		t.Fatalf("signed code = %q, want 12345", backend.signedCode)
	}
	if _, err := authflow.NewStateStore(authStateDir).Load(state.ID); err == nil {
		t.Fatal("state should be removed after successful login")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".agent-telegram", "config.json")); err != nil {
		t.Fatalf("config should be saved: %v", err)
	}
}

func TestAuthVerifyReports2FAWithoutDeletingState(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	authStateID = state.ID
	authCodeStdin = true

	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
		signResult:  &types.SignInResult{Requires2FA: true, TwoFactorHint: "hint"},
	}
	newAuthBackend = func(_ *config.Config) authflow.Backend { return backend }

	withStdin(t, "12345\n", func() {
		output := captureStdout(t, func() {
			runAuthVerify(&cobra.Command{}, nil)
		})
		var body struct {
			OK          bool   `json:"ok"`
			Next        string `json:"next"`
			Requires2FA bool   `json:"requires2FA"`
			Hint        string `json:"hint"`
		}
		if err := json.Unmarshal([]byte(output), &body); err != nil {
			t.Fatal(err)
		}
		if !body.OK || body.Next != "password" || !body.Requires2FA || body.Hint != "hint" {
			t.Fatalf("unexpected 2FA response: %s", output)
		}
	})

	loaded, err := authflow.NewStateStore(authStateDir).Load(state.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Requires2FA || loaded.TwoFactorHint != "hint" {
		t.Fatalf("state should retain 2FA metadata: %+v", loaded)
	}
}

func TestAuthPasswordCompletes2FA(t *testing.T) {
	tmp := t.TempDir()
	resetAuthGlobals(t, tmp)
	state := createTestState(t, tmp)
	authStateID = state.ID
	authPassStdin = true
	authReload = false

	backend := &fakeAuthBackend{
		sessionPath: state.SessionPath,
		passResult:  &types.SignInResult{Success: true},
	}
	newAuthBackend = func(_ *config.Config) authflow.Backend { return backend }

	withStdin(t, "password with spaces\n", func() {
		output := captureStdout(t, func() {
			runAuthPassword(&cobra.Command{}, nil)
		})
		var body struct {
			OK   bool   `json:"ok"`
			Next string `json:"next"`
		}
		if err := json.Unmarshal([]byte(output), &body); err != nil {
			t.Fatal(err)
		}
		if !body.OK || body.Next != "done" {
			t.Fatalf("unexpected password response: %s", output)
		}
	})

	if backend.password != "password with spaces" {
		t.Fatalf("password = %q, want trailing newline trimmed only", backend.password)
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

	authAppID = "123"
	authAppHash = "app-hash"
	authPhone = "+88806283792"
	authStateDir = filepath.Join(home, ".agent-telegram", "auth-state")
	authStateTTL = time.Minute
	authStateID = ""
	authCodeStdin = false
	authPassStdin = false
	authReload = false
	authStatusPhone = ""
}

func createTestState(t *testing.T, home string) *authflow.State {
	t.Helper()
	store := authflow.NewStateStore(authStateDir)
	state, err := store.Create(
		"+88806283792",
		"hash",
		123,
		"app-hash",
		filepath.Join(home, ".agent-telegram", "session.json"),
		time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan struct {
		data []byte
		err  error
	}, 1)
	go func() {
		data, err := io.ReadAll(r)
		done <- struct {
			data []byte
			err  error
		}{data: data, err: err}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	result := <-done
	if result.err != nil {
		t.Fatal(result.err)
	}
	return string(result.data)
}

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()

	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	defer func() { os.Stdin = old }()

	if _, err := io.Copy(w, bytes.NewBufferString(input)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	fn()
}
