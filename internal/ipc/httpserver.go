package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxAPIBodySize   = 1 << 20 // 1 MB
	httpShutdownWait = 5 * time.Second
)

// HTTPServer serves registered IPC handlers over HTTP.
// It implements MethodRegistrar so RegisterHandlers works identically to SocketServer.
type HTTPServer struct {
	methods map[string]Handler
	mu      sync.RWMutex
	secret  string
	cors    string
	srv     *http.Server
}

// NewHTTPServer creates a new HTTP API server.
func NewHTTPServer(port int, secret, cors string) *HTTPServer {
	s := &HTTPServer{
		methods: make(map[string]Handler),
		secret:  secret,
		cors:    cors,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /methods", s.handleMethods)
	mux.HandleFunc("POST /rpc/{method}", s.handleRPC)

	var handler http.Handler = mux
	handler = s.corsMiddleware(handler)
	if secret != "" {
		handler = s.authMiddleware(handler)
	}
	handler = s.loggingMiddleware(handler)

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return s
}

// Register implements MethodRegistrar.
func (s *HTTPServer) Register(name string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = handler
}

// Start starts the HTTP server. Blocks until ctx is cancelled.
func (s *HTTPServer) Start(ctx context.Context) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.srv.Addr, err)
	}

	slog.Info("HTTP API listening", "addr", s.srv.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), httpShutdownWait)
	defer shutCancel()

	return s.srv.Shutdown(shutCtx) //nolint:contextcheck // parent ctx already cancelled
}

// --- Handlers ---

func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *HTTPServer) handleMethods(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	names := make([]string, 0, len(s.methods))
	for name := range s.methods {
		names = append(names, name)
	}
	s.mu.RUnlock()

	sort.Strings(names)
	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "methods": names})
}

func (s *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	method := r.PathValue("method")

	s.mu.RLock()
	handler, ok := s.methods[method]
	s.mu.RUnlock()

	if !ok {
		writeJSONResponse(w, http.StatusNotFound, map[string]any{
			"ok":    false,
			"error": map[string]any{"code": ErrCodeMethodNotFound, "message": "method not found"},
		})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxAPIBodySize))
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": map[string]any{"code": ErrCodeParseError, "message": "failed to read request body"},
		})
		return
	}

	// Default to empty object if no body
	params := json.RawMessage(body)
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	result, rpcErr := handler(params)
	if rpcErr != nil {
		status := errorToHTTPStatus(rpcErr)
		writeJSONResponse(w, status, map[string]any{
			"ok":    false,
			"error": map[string]any{"code": rpcErr.Code, "message": rpcErr.Message},
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "result": result})
}

// --- Middleware ---

func (s *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" || token != s.secret {
			writeJSONResponse(w, http.StatusUnauthorized, map[string]any{
				"ok":    false,
				"error": map[string]any{"code": ErrCodeNotAuthorized, "message": "unauthorized"},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.cors
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// --- Helpers ---

func errorToHTTPStatus(err *ErrorObject) int {
	switch err.Code {
	case ErrCodeParseError, ErrCodeInvalidRequest, ErrCodeInvalidParams:
		return http.StatusBadRequest
	case ErrCodeMethodNotFound:
		return http.StatusNotFound
	case ErrCodeNotAuthorized:
		return http.StatusForbidden
	case ErrCodeNotInitialized:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errchkjson // best-effort response write
	_ = json.NewEncoder(w).Encode(v)
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
