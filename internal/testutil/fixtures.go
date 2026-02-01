// Package testutil provides testing utilities for the agent-telegram project.
package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// FixtureMeta contains metadata about a recorded fixture.
//
//nolint:tagliatelle // JSON tags match fixture file format
type FixtureMeta struct {
	Method        string    `json:"method"`
	RecordedAt    time.Time `json:"recorded_at"`
	TelegramLayer int       `json:"telegram_layer"`
	Notes         string    `json:"notes,omitempty"`
	Sanitized     bool      `json:"sanitized"`
}

// Fixture represents a recorded API request/response pair.
type Fixture struct {
	Meta     FixtureMeta     `json:"meta"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
}

// fixturesDir returns the path to testdata/fixtures directory.
func fixturesDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller info")
	}
	// internal/testutil/fixtures.go -> testdata/fixtures
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "fixtures")
}

// LoadFixture loads a fixture file by relative path.
// Example: LoadFixture(t, "messages/get_history_private_chat.json")
func LoadFixture(t *testing.T, relativePath string) *Fixture {
	t.Helper()

	path := filepath.Join(fixturesDir(), relativePath)
	data, err := os.ReadFile(path) //nolint:gosec // Test fixture path is safe
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", relativePath, err)
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("failed to unmarshal fixture %s: %v", relativePath, err)
	}

	return &fixture
}

// LoadFixtureRaw loads raw fixture file content.
func LoadFixtureRaw(t *testing.T, relativePath string) []byte {
	t.Helper()

	path := filepath.Join(fixturesDir(), relativePath)
	data, err := os.ReadFile(path) //nolint:gosec // Test fixture path is safe
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", relativePath, err)
	}

	return data
}

// MustLoadFixture loads a fixture or panics (for use in init/setup).
func MustLoadFixture(relativePath string) *Fixture {
	path := filepath.Join(fixturesDir(), relativePath)
	data, err := os.ReadFile(path) //nolint:gosec // Test fixture path is safe
	if err != nil {
		panic("failed to read fixture: " + err.Error())
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		panic("failed to unmarshal fixture: " + err.Error())
	}

	return &fixture
}

// UnmarshalResponse unmarshals fixture response into the given type.
func UnmarshalResponse[T any](t *testing.T, f *Fixture) T {
	t.Helper()

	var result T
	if err := json.Unmarshal(f.Response, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	return result
}

// UnmarshalRequest unmarshals fixture request into the given type.
func UnmarshalRequest[T any](t *testing.T, f *Fixture) T {
	t.Helper()

	var result T
	if err := json.Unmarshal(f.Request, &result); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	return result
}

// FixtureExists checks if a fixture file exists.
func FixtureExists(relativePath string) bool {
	path := filepath.Join(fixturesDir(), relativePath)
	_, err := os.Stat(path)
	return err == nil
}

// ListFixtures returns all fixture files in a directory.
func ListFixtures(t *testing.T, dir string) []string {
	t.Helper()

	path := filepath.Join(fixturesDir(), dir)
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("failed to read fixtures dir %s: %v", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}

	return files
}
