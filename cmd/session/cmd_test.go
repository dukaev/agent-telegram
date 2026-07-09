package session

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"agent-telegram/internal/sessionstore"
)

func newSessionTestRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TELEGRAM_APP_ID", "")
	t.Setenv("TELEGRAM_APP_HASH", "")
	t.Setenv(sessionstore.EnvProvider, sessionstore.MemoryProvider)
	t.Setenv(sessionstore.EnvProfile, "session-command-test")

	root := &cobra.Command{Use: "agent-telegram", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("confirm", false, "confirm")
	root.PersistentFlags().String("session-provider", sessionstore.MemoryProvider, "provider")
	root.PersistentFlags().String("profile", "session-command-test", "profile")
	parent := &cobra.Command{Use: "session"}
	AddExportCommand(parent)
	AddImportCommand(parent)
	AddStatusCommand(parent)
	AddProvidersCommand(parent)
	AddForgetCommand(parent)
	root.AddCommand(parent)
	root.SetArgs(append([]string{"session"}, args...))
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	return root, out
}

func TestSessionImportStatusExportForget(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("managed-session"))
	root, out := newSessionTestRoot(t, "import", "--confirm")
	root.SetIn(strings.NewReader(encoded))
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"provider": "memory"`) {
		t.Fatalf("import output = %s", out.String())
	}

	root, out = newSessionTestRoot(t, "status")
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if stored, _ := status["stored"].(bool); !stored {
		t.Fatalf("status = %v", status)
	}

	root, out = newSessionTestRoot(t, "export", "--confirm")
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != encoded {
		t.Fatalf("export = %q", out.String())
	}

	root, _ = newSessionTestRoot(t, "forget", "--confirm")
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	root, out = newSessionTestRoot(t, "status")
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"stored": false`) {
		t.Fatalf("status after forget = %s", out.String())
	}
}

func TestSessionExportRequiresConfirmation(t *testing.T) {
	root, _ := newSessionTestRoot(t, "export")
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v", err)
	}
}
