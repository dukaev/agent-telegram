// Package ipc provides inter-process communication via JSON-RPC.
package ipc

// MethodCaller defines the interface for RPC method calls.
// This interface allows for mocking in tests.
type MethodCaller interface {
	// Call invokes a JSON-RPC method with the given parameters.
	// Returns the result or an error object.
	Call(method string, params any) (any, *ErrorObject)
}

// Pinger defines ping functionality.
type Pinger interface {
	Ping(message string) (*PingResult, error)
}

// RPCClient is a composite interface combining common client operations.
// This is the main interface that should be used for dependency injection.
type RPCClient interface {
	MethodCaller
	Pinger
}

// Ensure *ipc.Client implements RPCClient.
var (
	_ MethodCaller = (*Client)(nil)
	_ Pinger       = (*Client)(nil)
	_ RPCClient    = (*Client)(nil)
)

// MockClient is a mock implementation for testing.
type MockClient struct {
	// CallFunc is the mock implementation for Call.
	CallFunc func(method string, params any) (any, *ErrorObject)
	// PingFunc is the mock implementation for Ping.
	PingFunc func(message string) (*PingResult, error)
}

// Call implements MethodCaller for MockClient.
func (m *MockClient) Call(method string, params any) (any, *ErrorObject) {
	if m.CallFunc != nil {
		return m.CallFunc(method, params)
	}
	return nil, ErrInternalError
}

// Ping implements Pinger for MockClient.
func (m *MockClient) Ping(message string) (*PingResult, error) {
	if m.PingFunc != nil {
		return m.PingFunc(message)
	}
	return &PingResult{
		Message: "pong: " + message,
		Pong:    true,
	}, nil
}
