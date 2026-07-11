package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"agent-telegram/internal/ipc"
	"agent-telegram/internal/observability"
	"agent-telegram/internal/operations"
)

// handleError handles RPC errors with user-friendly messages.
func (r *Runner) handleError(err *ipc.ErrorObject) {
	if r.agentMode {
		r.printErrorEnvelope(err)
		Exit(1)
		return
	}
	switch err.Code {
	case ipc.ErrCodeServerNotRunning:
		fmt.Fprintln(os.Stderr, "Error: Server is not running")
		fmt.Fprintln(os.Stderr, "Run: agent-telegram server ensure")
	case ipc.ErrCodeNotAuthorized:
		fmt.Fprintln(os.Stderr, "Error: Not authorized")
		fmt.Fprintln(os.Stderr, "Please authenticate first: agent-telegram auth")
	case ipc.ErrCodeNotInitialized:
		fmt.Fprintln(os.Stderr, "Error: Client not initialized")
		fmt.Fprintln(os.Stderr, "The server may still be starting up. Please try again in a few seconds.")
	default:
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
	}
	Exit(1)
}

func (r *Runner) ensureErrorToRPC(err error, method string) *ipc.ErrorObject {
	msg := err.Error()
	switch msg {
	case "server is not running; run: agent-telegram server ensure":
		return ipc.NewTypedError(ipc.ErrCodeServerNotRunning, ipc.ErrorTypeServerNotRunning, msg, map[string]any{
			"nextCommand": agentCommand(r.runID, "server", "ensure"),
		})
	case "telegram client is not ready within timeout":
		return ipc.NewTypedError(ipc.ErrCodeNotInitialized, ipc.ErrorTypeNotInitialized, msg, map[string]any{
			"nextCommand": agentCommand(r.runID, "server", "wait-ready"),
		})
	default:
		return ipc.NewTypedError(-32000, ipc.ErrorTypeInternal, fmt.Sprintf("%s failed before %s", method, msg), nil)
	}
}

func (r *Runner) printErrorEnvelope(err *ipc.ErrorObject) {
	r.printErrorEnvelopeWithDetails(err, FailureDetails{})
}

func (r *Runner) printErrorEnvelopeWithDetails(err *ipc.ErrorObject, details FailureDetails) {
	nextActions := nextActionsForError(err, r)
	if details.NextActions != nil {
		nextActions = details.NextActions
	}
	payload := map[string]any{
		"ok":          false,
		"runId":       r.runID,
		"traceId":     r.traceID,
		"command":     r.commandPath,
		"method":      r.lastMethod,
		"safety":      r.lastSafety,
		"error":       errorObjectForOutput(err),
		"diagnosis":   diagnosisForError(err),
		"nextActions": nextActions,
	}
	if details.PartialResult != nil {
		payload["partialResult"] = details.PartialResult
	}
	data, jsonErr := json.MarshalIndent(payload, "", "  ")
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
		return
	}
	fmt.Println(string(data))
}

func errorObjectForOutput(err *ipc.ErrorObject) map[string]any {
	out := map[string]any{
		"code":      err.Code,
		"type":      errorType(err),
		"message":   err.Message,
		"retryable": retryableError(err),
	}
	if err.Data != nil {
		out["data"] = observability.RedactAny(err.Data)
	}
	return out
}

func diagnosisForError(err *ipc.ErrorObject) map[string]any {
	errType := errorType(err)
	switch errType {
	case ipc.ErrorTypeServerNotRunning:
		return map[string]any{
			"category": "server_not_running",
			"summary":  "The local IPC server is not reachable.",
			"retry":    "Start the server, then retry the original command.",
		}
	case ipc.ErrorTypeNotAuthorized:
		return map[string]any{
			"category": "not_authorized",
			"summary":  "Telegram session is missing or expired.",
			"retry":    "Authenticate first, then retry the original command.",
		}
	case ipc.ErrorTypeNotInitialized:
		return map[string]any{
			"category": "not_initialized",
			"summary":  "The server is running but Telegram is not ready yet.",
			"retry":    "Wait for readiness, then retry.",
		}
	case ipc.ErrorTypeFloodWait:
		return map[string]any{
			"category": "rate_limited",
			"summary":  "Telegram returned a flood wait.",
			"retry":    "Wait for retryAfter seconds when present.",
		}
	case ipc.ErrorTypeValidation:
		return map[string]any{
			"category": "validation",
			"summary":  "Command parameters failed validation.",
			"retry":    "Inspect the schema and retry with valid parameters.",
		}
	case ipc.ErrorTypeTimeout:
		return map[string]any{
			"category": "timeout",
			"summary":  "The requested wait did not complete before the deadline.",
			"retry":    "Continue waiting or inspect the trace before repeating a write action.",
		}
	case ipc.ErrorTypePeerNotFound:
		return map[string]any{
			"category": "peer_not_found",
			"summary":  "Telegram could not resolve the requested peer.",
			"retry":    "Verify the username, ID, invite link, or access rights.",
		}
	default:
		return map[string]any{
			"category": "unknown",
			"summary":  "The command failed.",
			"retry":    "Inspect audit/logs with the trace ID.",
		}
	}
}

func nextActionsForError(err *ipc.ErrorObject, r *Runner) []map[string]any {
	errType := errorType(err)
	switch errType {
	case ipc.ErrorTypeServerNotRunning:
		return []map[string]any{{
			"kind":    "start_server",
			"command": agentCommand(r.runID, "server", "ensure"),
			"safety":  "local",
			"reason":  "server is required before RPC-backed commands",
		}}
	case ipc.ErrorTypeNotAuthorized:
		return []map[string]any{{
			"kind":    "authenticate",
			"command": "agent-telegram auth --agent --run-id " + r.runID,
			"safety":  "sensitive",
			"reason":  "Telegram session is required",
		}}
	case ipc.ErrorTypeNotInitialized:
		return []map[string]any{{
			"kind":    "wait_ready",
			"command": agentCommand(r.runID, "server", "wait-ready"),
			"safety":  "local",
			"reason":  "server is still initializing",
		}}
	case ipc.ErrorTypeValidation:
		command := r.commandPath + " --schema"
		if r.commandPath == "" {
			command = "agent-telegram manifest"
		}
		return []map[string]any{{
			"kind":    "inspect_schema",
			"command": command,
			"safety":  "read",
			"reason":  "schema explains required parameters",
		}}
	default:
		return []map[string]any{{
			"kind":    "inspect_trace",
			"command": "agent-telegram trace inspect " + r.traceID + " --agent --run-id " + r.runID,
			"safety":  "read",
			"reason":  "trace bundle can show related audit and log events",
		}}
	}
}

func retryableError(err *ipc.ErrorObject) bool {
	switch errorType(err) {
	case ipc.ErrorTypeServerNotRunning, ipc.ErrorTypeNotInitialized, ipc.ErrorTypeTimeout, ipc.ErrorTypeFloodWait:
		return true
	default:
		return false
	}
}

func agentCommand(runID string, args ...string) string {
	base := "agent-telegram"
	for _, arg := range args {
		base += " " + arg
	}
	base += " --agent"
	if runID != "" {
		base += " --run-id " + runID
	}
	return base
}

func operationSafety(method string) string {
	if op, ok := operations.Get(method); ok {
		return op.Safety
	}
	return ""
}

func errorType(err *ipc.ErrorObject) string {
	if err == nil || err.Data == nil {
		return ""
	}
	if m, ok := err.Data.(map[string]any); ok {
		if value, ok := m["type"].(string); ok {
			return value
		}
	}
	return ""
}
