// Package ipc provides inter-process communication via JSON-RPC.
package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client represents a JSON-RPC client.
type Client struct {
	path    string
	timeout time.Duration
}

// NewClient creates a new JSON-RPC client.
func NewClient(path string) *Client {
	if path == "" {
		path = defaultSocketPath
	}
	return &Client{
		path:    path,
		timeout: 5 * time.Second,
	}
}

// Call calls a JSON-RPC method.
func (c *Client) Call(method string, params interface{}) (interface{}, *ErrorObject) {
	// Create request
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		ID:      1,
	}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, ErrInternalError
		}
		req.Params = data
	}

	// Connect to socket
	dialer := &net.Dialer{Timeout: c.timeout}
	conn, err := dialer.Dial("unix", c.path)
	if err != nil {
		return nil, &ErrorObject{
			Code:    -32000,
			Message: fmt.Sprintf("Failed to connect: %v", err),
		}
	}
	defer func() { _ = conn.Close() }()

	// Set deadline
	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	// Send request
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return nil, ErrInternalError
	}

	// Receive response
	decoder := json.NewDecoder(conn)
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return nil, ErrInternalError
	}

	return resp.Result, resp.Error
}

// Ping sends a ping request.
func (c *Client) Ping(message string) (*PingResult, error) {
	result, rpcErr := c.Call("ping", map[string]string{"message": message})
	if rpcErr != nil {
		return nil, fmt.Errorf("ping failed: %s", rpcErr.Message)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var pingResult PingResult
	if err := json.Unmarshal(data, &pingResult); err != nil {
		return nil, err
	}

	return &pingResult, nil
}

// Echo sends an echo request.
func (c *Client) Echo(message string) (string, error) {
	result, rpcErr := c.Call("echo", map[string]string{"message": message})
	if rpcErr != nil {
		return "", fmt.Errorf("echo failed: %s", rpcErr.Message)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response type")
	}

	echo, ok := m["echo"].(string)
	if !ok {
		return "", fmt.Errorf("no echo in response")
	}

	return echo, nil
}

// Status gets the server status.
func (c *Client) Status() (map[string]interface{}, error) {
	result, rpcErr := c.Call("status", nil)
	if rpcErr != nil {
		return nil, fmt.Errorf("status failed: %s", rpcErr.Message)
	}

	status, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	return status, nil
}
