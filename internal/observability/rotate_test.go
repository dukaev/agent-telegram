package observability

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAppendLogRotatesOversizedFile(t *testing.T) {
	t.Setenv(logMaxBytesEnv, "5")
	path := filepath.Join(t.TempDir(), "agent.log")
	if err := os.WriteFile(path, []byte("123456"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := OpenAppendLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path + ".1"); err != nil || string(data) != "123456" {
		t.Fatalf("rotated file = %q, %v", data, err)
	}
}

func TestRotatingWriterRotatesWhileOpen(t *testing.T) {
	t.Setenv(logMaxBytesEnv, "5")
	path := filepath.Join(t.TempDir(), "daemon.log")
	writer, err := NewRotatingWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("1234")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("56")); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path + ".1"); err != nil || string(data) != "1234" {
		t.Fatalf("rotated generation = %q, %v", data, err)
	}
}
