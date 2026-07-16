package ipc

import "errors"

const (
	// Domain-level errors.
	ErrCodeTimeout          = -32004
	ErrCodePeerNotFound     = -32010
	ErrCodeForbidden        = -32011
	ErrCodeFloodWait        = -32012
	ErrCodeTopicsNotEnabled = -32013
	ErrCodePolicyDenied     = -32020
)

const (
	ErrorTypeServerNotRunning = "SERVER_NOT_RUNNING"
	ErrorTypeMethodNotFound   = "METHOD_NOT_FOUND"
	ErrorTypeNotAuthorized    = "NOT_AUTHORIZED"
	ErrorTypeNotInitialized   = "NOT_INITIALIZED"
	ErrorTypeTimeout          = "TIMEOUT"
	ErrorTypePeerNotFound     = "PEER_NOT_FOUND"
	ErrorTypeForbidden        = "FORBIDDEN"
	ErrorTypeFloodWait        = "FLOOD_WAIT"
	ErrorTypeTopicsNotEnabled = "TOPICS_NOT_ENABLED"
	ErrorTypePolicyDenied     = "POLICY_DENIED"
	ErrorTypeValidation       = "VALIDATION"
	ErrorTypeInternal         = "INTERNAL"
)

// NewTypedError creates a JSON-RPC error with a stable machine-readable type.
func NewTypedError(code int, errType, message string, data map[string]any) *ErrorObject {
	if data == nil {
		data = make(map[string]any, 1)
	}
	data["type"] = errType
	return &ErrorObject{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// RPCError is implemented by errors that can provide a JSON-RPC error object.
type RPCError interface {
	RPCError() *ErrorObject
}

// ErrorObjectFromError returns a JSON-RPC error object embedded in err.
func ErrorObjectFromError(err error) *ErrorObject {
	var rpcErr RPCError
	if errors.As(err, &rpcErr) {
		return rpcErr.RPCError()
	}
	return nil
}

// PolicyDeniedError reports that a configured local policy blocked an operation.
type PolicyDeniedError struct {
	Method string
	Reason string
}

// NewPolicyDeniedError creates a policy denial error.
func NewPolicyDeniedError(method, reason string) *PolicyDeniedError {
	return &PolicyDeniedError{Method: method, Reason: reason}
}

// Error implements error.
func (e *PolicyDeniedError) Error() string {
	if e.Method == "" {
		return "operation denied by local policy: " + e.Reason
	}
	return "operation " + e.Method + " denied by local policy: " + e.Reason
}

// RPCError implements RPCError.
func (e *PolicyDeniedError) RPCError() *ErrorObject {
	data := map[string]any{
		"method": e.Method,
		"reason": e.Reason,
	}
	return NewTypedError(ErrCodePolicyDenied, ErrorTypePolicyDenied, e.Error(), data)
}
