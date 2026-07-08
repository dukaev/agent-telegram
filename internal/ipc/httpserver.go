package ipc

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	"agent-telegram/internal/operations"
	"agent-telegram/internal/skills"
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
	policy  PolicyChecker
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
	mux.HandleFunc("GET /manifest", s.handleManifest)
	mux.HandleFunc("GET /openapi.json", s.handleOpenAPI)
	mux.HandleFunc("POST /rpc/{method}", s.handleRPC)

	var handler http.Handler = mux
	if cors != "" {
		handler = s.corsMiddleware(handler)
	}
	if secret != "" {
		handler = s.authMiddleware(handler)
	}
	handler = s.loggingMiddleware(handler)

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: ClientTimeout(),
	}
	return s
}

// Register implements MethodRegistrar.
func (s *HTTPServer) Register(name string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = handler
}

// SetPolicyChecker sets the local policy checker used for HTTP calls.
func (s *HTTPServer) SetPolicyChecker(policy PolicyChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policy = policy
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
	names := slices.Sorted(maps.Keys(s.methods))
	s.mu.RUnlock()

	writeJSONResponse(w, http.StatusOK, map[string]any{"ok": true, "methods": names})
}

func (s *HTTPServer) handleManifest(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]any{
		"ok":         true,
		"operations": operations.Manifest(),
		"errorTypes": ErrorTypesManifest(),
		"skills":     skills.Manifest(),
	})
}

func (s *HTTPServer) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, http.StatusOK, operations.OpenAPI("agent-telegram API", "dev"))
}

func (s *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	req := newHTTPRPCRequest(w, r)
	handler, ok := s.lookupHandler(req.method)
	if !ok {
		rpcErr := NewTypedError(ErrCodeMethodNotFound, ErrorTypeMethodNotFound, "method not found", nil)
		s.writeRPCError(w, req, nil, rpcErr, http.StatusNotFound, false)
		return
	}

	body, rpcErr, status := readHTTPRPCBody(r)
	if rpcErr != nil {
		s.writeRPCError(w, req, nil, rpcErr, status, false)
		return
	}

	params, dryRun, validateOnly, rpcErr := parseHTTPRPCParams(r, body)
	if rpcErr != nil {
		s.writeRPCError(w, req, nil, rpcErr, errorToHTTPStatus(rpcErr), false)
		return
	}

	if rpcErr := s.checkHTTPPolicy(r.Context(), req.method, params); rpcErr != nil {
		s.writeRPCError(w, req, params, rpcErr, errorToHTTPStatus(rpcErr), dryRun || validateOnly)
		return
	}
	if dryRun || validateOnly {
		s.handleHTTPDryRun(w, req, params, dryRun, validateOnly)
		return
	}

	result, rpcErr := handler(params)
	if rpcErr != nil {
		s.writeRPCError(w, req, params, rpcErr, errorToHTTPStatus(rpcErr), false)
		return
	}

	s.writeRPCSuccess(w, req, params, result)
}
