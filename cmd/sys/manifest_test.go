package sys

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestManifestCommandOutputsOperations(t *testing.T) {
	manifestOpenAPI = false
	output := captureStdout(t, func() {
		ManifestCmd.Run(ManifestCmd, nil)
	})

	var body struct {
		OK         bool `json:"ok"`
		Operations []struct {
			Method string `json:"method"`
			Safety string `json:"safety"`
		} `json:"operations"`
		ErrorTypes []struct {
			Type      string `json:"type"`
			Code      int    `json:"code"`
			Retryable bool   `json:"retryable"`
		} `json:"errorTypes"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if !body.OK {
		t.Fatal("manifest ok should be true")
	}
	if len(body.Operations) == 0 {
		t.Fatal("manifest should include operations")
	}
	if len(body.ErrorTypes) == 0 {
		t.Fatal("manifest should include error types")
	}
}

func TestManifestCommandOutputsOpenAPI(t *testing.T) {
	manifestOpenAPI = true
	defer func() { manifestOpenAPI = false }()

	output := captureStdout(t, func() {
		ManifestCmd.Run(ManifestCmd, nil)
	})

	var body struct {
		OpenAPI string         `json:"openapi"`
		Paths   map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil {
		t.Fatal(err)
	}
	if body.OpenAPI != "3.1.0" {
		t.Fatalf("openapi = %q, want 3.1.0", body.OpenAPI)
	}
	if _, ok := body.Paths["/rpc/send_message"]; !ok {
		t.Fatal("OpenAPI should include /rpc/send_message")
	}
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

	type readResult struct {
		data []byte
		err  error
	}
	readDone := make(chan readResult, 1)
	go func() {
		data, err := io.ReadAll(r)
		readDone <- readResult{data: data, err: err}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	result := <-readDone
	if result.err != nil {
		t.Fatal(result.err)
	}
	return string(result.data)
}
