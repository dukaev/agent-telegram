package ipc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"agent-telegram/internal/observability"
)

type denyPolicyChecker struct{}

func (denyPolicyChecker) Check(context.Context, string, json.RawMessage) error {
	return NewPolicyDeniedError("send_message", "write operations are disabled")
}

func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "agent-telegram-ipc-test-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv("HOME", home)
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

func TestHTTPServerRequiresBearerForRPC(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("ok", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		return map[string]any{"ok": true}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/ok", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodPost, "/rpc/ok", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHTTPServerDryRunValidatesWithoutExecuting(t *testing.T) {
	called := false
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		called = true
		return map[string]any{"ok": true}, nil
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/rpc/send_message?dryRun=true",
		strings.NewReader(`{"peer":"@user","message":"hello"}`),
	)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("dry-run status = %d, want %d", rec.Code, http.StatusOK)
	}
	if called {
		t.Fatal("handler should not execute during dry-run")
	}
}

func TestHTTPServerDryRunEnvelope(t *testing.T) {
	called := false
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		called = true
		return map[string]any{"ok": true}, nil
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/rpc/send_message",
		strings.NewReader(`{"dryRun":true,"params":{"peer":"@user","message":"hello"}}`),
	)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("dry-run envelope status = %d, want %d", rec.Code, http.StatusOK)
	}
	if called {
		t.Fatal("handler should not execute during dry-run envelope")
	}

	var body struct {
		DryRun bool   `json:"dryRun"`
		Method string `json:"method"`
		Safety string `json:"safety"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.DryRun || body.Method != "send_message" || body.Safety != "write" {
		t.Fatalf("unexpected dry-run body: %s", rec.Body.String())
	}
}

func TestHTTPServerDryRunReturnsValidationError(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		t.Fatal("handler should not execute on validation failure")
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/send_message?dryRun=true", strings.NewReader(`{"message":"hello"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("validation status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body struct {
		Error struct {
			Data map[string]any `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Data["type"] != ErrorTypeValidation {
		t.Fatalf("error type = %v, want %s", body.Error.Data["type"], ErrorTypeValidation)
	}
}

func TestHTTPServerDryRunHonorsPolicy(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.SetPolicyChecker(denyPolicyChecker{})
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		t.Fatal("handler should not execute when policy denies")
		return nil, nil
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/rpc/send_message?dryRun=true",
		strings.NewReader(`{"peer":"@user","message":"hello"}`),
	)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("policy status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	var body struct {
		Error struct {
			Data map[string]any `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Data["type"] != ErrorTypePolicyDenied {
		t.Fatalf("error type = %v, want %s", body.Error.Data["type"], ErrorTypePolicyDenied)
	}
}

func TestHTTPServerRejectsInvalidJSON(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		t.Fatal("handler should not execute on invalid JSON")
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/send_message", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHTTPServerUnknownMethodTypedError(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")

	req := httptest.NewRequest(http.MethodPost, "/rpc/missing", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body struct {
		Error struct {
			Data map[string]any `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Data["type"] != ErrorTypeMethodNotFound {
		t.Fatalf("error type = %v, want %s", body.Error.Data["type"], ErrorTypeMethodNotFound)
	}
}

func TestHTTPServerRejectsOversizedBody(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		t.Fatal("handler should not execute on oversized body")
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/send_message", strings.NewReader(strings.Repeat("x", maxAPIBodySize+1)))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestHTTPServerValidateOnlyQuery(t *testing.T) {
	called := false
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		called = true
		return map[string]any{"ok": true}, nil
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/rpc/send_message?validateOnly=yes",
		strings.NewReader(`{"peer":"@user","message":"hello"}`),
	)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("validate-only status = %d, want %d", rec.Code, http.StatusOK)
	}
	if called {
		t.Fatal("handler should not execute during validate-only")
	}

	var body struct {
		DryRun       bool   `json:"dryRun"`
		ValidateOnly bool   `json:"validateOnly"`
		Method       string `json:"method"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.DryRun || !body.ValidateOnly || body.Method != "send_message" {
		t.Fatalf("unexpected validate-only body: %s", rec.Body.String())
	}
}

func TestHTTPServerIncludesTypedErrorData(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		return nil, NewTypedError(ErrCodePeerNotFound, ErrorTypePeerNotFound, "not found", nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/send_message", strings.NewReader(`{"peer":"@user","message":"hello"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("typed error status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body struct {
		Error struct {
			Data map[string]any `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Data["type"] != ErrorTypePeerNotFound {
		t.Fatalf("error data type = %v, want %s", body.Error.Data["type"], ErrorTypePeerNotFound)
	}
}

func TestHTTPServerTraceIDAndAudit(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")
	srv.Register("send_message", func(_ json.RawMessage) (interface{}, *ErrorObject) {
		return map[string]any{"messageId": 123, "message": "hello"}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/rpc/send_message", strings.NewReader(`{"peer":"@user","message":"hello"}`))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-Trace-Id", "trace-test")
	req.Header.Set("X-Run-Id", "run-test")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("X-Trace-Id") != "trace-test" {
		t.Fatalf("trace header = %q", rec.Header().Get("X-Trace-Id"))
	}
	if rec.Header().Get("X-Run-Id") != "run-test" {
		t.Fatalf("run header = %q", rec.Header().Get("X-Run-Id"))
	}
	var body struct {
		RunID   string `json:"runId"`
		TraceID string `json:"traceId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.RunID != "run-test" || body.TraceID != "trace-test" {
		t.Fatalf("ids = %q/%q, want run-test/trace-test", body.RunID, body.TraceID)
	}

	events, err := observability.ReadAudit("", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected audit event")
	}
	last := events[len(events)-1]
	if last.RunID != "run-test" || last.TraceID != "trace-test" || last.Method != "send_message" || last.Status != "ok" {
		t.Fatalf("unexpected audit event: %+v", last)
	}
}

func TestHTTPServerManifestAndOpenAPI(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")

	req := httptest.NewRequest(http.MethodGet, "/manifest", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/manifest status = %d, want %d", rec.Code, http.StatusOK)
	}
	var manifest struct {
		OK         bool `json:"ok"`
		Operations []struct {
			Method      string         `json:"method"`
			Safety      string         `json:"safety"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"operations"`
		ErrorTypes []struct {
			Type string `json:"type"`
		} `json:"errorTypes"`
		Skills []struct {
			Name           string `json:"name"`
			InstallCommand string `json:"installCommand"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if !manifest.OK || len(manifest.Operations) == 0 || len(manifest.ErrorTypes) == 0 {
		t.Fatalf("unexpected manifest body: %s", rec.Body.String())
	}
	if len(manifest.Skills) == 0 || manifest.Skills[0].Name != "agent-telegram" || manifest.Skills[0].InstallCommand == "" {
		t.Fatalf("manifest should include installable skills: %s", rec.Body.String())
	}
	if !manifestHasOperation(manifest.Operations, "send_message") {
		t.Fatal("manifest should include send_message")
	}

	req = httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/openapi.json status = %d, want %d", rec.Code, http.StatusOK)
	}
	var openAPI struct {
		OpenAPI string         `json:"openapi"`
		Paths   map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &openAPI); err != nil {
		t.Fatal(err)
	}
	if openAPI.OpenAPI != "3.1.0" {
		t.Fatalf("openapi = %q, want 3.1.0", openAPI.OpenAPI)
	}
	if _, ok := openAPI.Paths["/rpc/send_message"]; !ok {
		t.Fatal("OpenAPI should include /rpc/send_message")
	}
}

func manifestHasOperation(operations []struct {
	Method      string         `json:"method"`
	Safety      string         `json:"safety"`
	InputSchema map[string]any `json:"inputSchema"`
}, method string) bool {
	for _, op := range operations {
		if op.Method == method && op.Safety != "" && len(op.InputSchema) > 0 {
			return true
		}
	}
	return false
}

func TestHTTPServerLeavesHealthPublic(t *testing.T) {
	srv := NewHTTPServer(0, "secret", "")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAllowedCORSOrigin(t *testing.T) {
	if got := allowedCORSOrigin("https://a.test,https://b.test", "https://b.test"); got != "https://b.test" {
		t.Fatalf("allowedCORSOrigin() = %q, want https://b.test", got)
	}
	if got := allowedCORSOrigin("https://a.test", "https://b.test"); got != "" {
		t.Fatalf("disallowed origin = %q, want empty", got)
	}
	if got := allowedCORSOrigin("*", "https://b.test"); got != "*" {
		t.Fatalf("wildcard origin = %q, want *", got)
	}
}
