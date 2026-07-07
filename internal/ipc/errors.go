package ipc

const (
	// Domain-level errors.
	ErrCodeTimeout      = -32004
	ErrCodePeerNotFound = -32010
	ErrCodeForbidden    = -32011
	ErrCodeFloodWait    = -32012
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
