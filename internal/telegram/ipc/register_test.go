package ipc

import (
	"context"
	"errors"
	"testing"

	baseipc "agent-telegram/internal/ipc"
	"agent-telegram/telegram/client"
)

func TestClassifyRPCErrorTimeout(t *testing.T) {
	errObj := classifyRPCError(context.DeadlineExceeded)

	if errObj.Code != baseipc.ErrCodeTimeout {
		t.Fatalf("code = %d, want %d", errObj.Code, baseipc.ErrCodeTimeout)
	}
	if got := errObj.Data.(map[string]any)["type"]; got != baseipc.ErrorTypeTimeout {
		t.Fatalf("type = %v, want %s", got, baseipc.ErrorTypeTimeout)
	}
}

func TestClassifyRPCErrorNotInitialized(t *testing.T) {
	errObj := classifyRPCError(client.ErrNotInitialized)

	if errObj.Code != baseipc.ErrCodeNotInitialized {
		t.Fatalf("code = %d, want %d", errObj.Code, baseipc.ErrCodeNotInitialized)
	}
}

func TestClassifyRPCErrorFloodWaitRetryAfter(t *testing.T) {
	errObj := classifyRPCError(errors.New("telegram: FLOOD_WAIT_42"))

	if errObj.Code != baseipc.ErrCodeFloodWait {
		t.Fatalf("code = %d, want %d", errObj.Code, baseipc.ErrCodeFloodWait)
	}
	data := errObj.Data.(map[string]any)
	if data["type"] != baseipc.ErrorTypeFloodWait {
		t.Fatalf("type = %v, want %s", data["type"], baseipc.ErrorTypeFloodWait)
	}
	if data["retryAfter"] != 42 {
		t.Fatalf("retryAfter = %v, want 42", data["retryAfter"])
	}
}

func TestClassifyRPCErrorPeerNotFound(t *testing.T) {
	errObj := classifyRPCError(errors.New("peer not found: @missing"))

	if errObj.Code != baseipc.ErrCodePeerNotFound {
		t.Fatalf("code = %d, want %d", errObj.Code, baseipc.ErrCodePeerNotFound)
	}
}
