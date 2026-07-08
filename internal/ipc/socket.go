// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

const (
	defaultSocketPath = "/tmp/agent-telegram.sock"
)

// DefaultSocketPath returns the default socket path.
func DefaultSocketPath() string {
	return defaultSocketPath
}

// SocketServer represents a Unix socket JSON-RPC server.
type SocketServer struct {
	server   *Server
	path     string
	listener net.Listener
	mu       sync.Mutex
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewSocketServer creates a new Unix socket server.
func NewSocketServer(path string) *SocketServer {
	if path == "" {
		path = defaultSocketPath
	}
	return &SocketServer{
		server: NewServer(),
		path:   path,
	}
}

// IsServerRunning checks if a server is already running on the socket.
func IsServerRunning(parent context.Context, socketPath string) bool {
	if socketPath == "" {
		socketPath = defaultSocketPath
	}

	// Check if socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return false
	}

	// Try to connect to the socket
	ctx, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Start starts the socket server.
func (s *SocketServer) Start(ctx context.Context) error {
	serverCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Check if server is already running
	if IsServerRunning(ctx, s.path) {
		return fmt.Errorf("server is already running on %s", s.path)
	}

	// Remove stale socket file if present (server not running but file exists)
	_ = os.Remove(s.path)

	// Create Unix socket listener
	//nolint:noctx // Server has its own context management via Shutdown()
	listener, err := net.Listen("unix", s.path)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.path, err)
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	// Set socket permissions
	if err := os.Chmod(s.path, 0600); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	fmt.Printf("IPC server listening on %s\n", s.path)

	// Accept connections
	s.wg.Go(func() { s.acceptLoop(serverCtx) })

	// Wait for context cancellation
	<-serverCtx.Done()
	return s.Shutdown()
}

// acceptLoop accepts incoming connections.
func (s *SocketServer) acceptLoop(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		s.wg.Go(func() { s.handleConnection(ctx, conn) })
	}
}

// handleConnection handles a single connection.
func (s *SocketServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() { _ = conn.Close() }()

	if err := s.server.Serve(ctx, conn); err != nil {
		slog.Debug("connection error", "error", err)
	}
}

// Shutdown gracefully shuts down the server.
func (s *SocketServer) Shutdown() error {
	if s.cancel != nil {
		s.cancel()
	}

	// Close listener
	s.mu.Lock()
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.mu.Unlock()

	// Wait for all connections to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Remove socket file
		_ = os.Remove(s.path)
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("shutdown timeout")
	}
}

// Register registers a method handler.
func (s *SocketServer) Register(name string, handler Handler) {
	s.server.Register(name, handler)
}

// SetPolicyChecker sets the local policy checker used before method execution.
func (s *SocketServer) SetPolicyChecker(policy PolicyChecker) {
	s.server.SetPolicyChecker(policy)
}
