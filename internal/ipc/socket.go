// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"context"
	"fmt"
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
	ctx      context.Context
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

// Start starts the socket server.
func (s *SocketServer) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Remove existing socket file if present
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
	s.wg.Add(1)
	go s.acceptLoop()

	// Wait for context cancellation
	<-s.ctx.Done()
	return s.Shutdown()
}

// acceptLoop accepts incoming connections.
func (s *SocketServer) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		fmt.Printf("📥 New connection from %s\n", conn.RemoteAddr())

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single connection.
func (s *SocketServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		_ = conn.Close()
		fmt.Printf("📤 Connection closed\n")
	}()

	if err := s.server.Serve(conn); err != nil {
		fmt.Printf("Connection error: %v\n", err)
	}
}

// Shutdown gracefully shuts down the server.
func (s *SocketServer) Shutdown() error {
	s.cancel()

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
