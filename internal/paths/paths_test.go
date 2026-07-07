package paths

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInstanceFileNamesKeepDefaultPaths(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() (string, error)
		baseName string
	}{
		{name: "log", fn: LogFilePath, baseName: "server.log"},
		{name: "cli log", fn: CLILogFilePath, baseName: "cli.log"},
		{name: "audit", fn: AuditFilePath, baseName: "audit.jsonl"},
		{name: "pid", fn: PIDFilePath, baseName: "server.pid"},
		{name: "lock", fn: LockFilePath, baseName: "server.lock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.fn()
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(path) != tt.baseName {
				t.Fatalf("base name = %q, want %q", filepath.Base(path), tt.baseName)
			}
		})
	}
}

func TestInstanceFileNamesUseSocketScopedSuffix(t *testing.T) {
	defaultPID, err := PIDFilePath()
	if err != nil {
		t.Fatal(err)
	}
	customPID, err := PIDFilePathForSocket("/tmp/agent-telegram-alt.sock")
	if err != nil {
		t.Fatal(err)
	}

	if customPID == defaultPID {
		t.Fatal("custom socket PID path should differ from default PID path")
	}
	if !strings.HasPrefix(filepath.Base(customPID), "server-") {
		t.Fatalf("custom PID file %q should include a server hash prefix", customPID)
	}
	if filepath.Ext(customPID) != ".pid" {
		t.Fatalf("custom PID file extension = %q, want .pid", filepath.Ext(customPID))
	}
}
