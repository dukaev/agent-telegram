package ipc

import (
	"testing"
	"time"
)

func TestRequestTimeoutUsesEnvOverride(t *testing.T) {
	t.Setenv(EnvRPCTimeout, "45s")

	if got := RequestTimeout(); got != 45*time.Second {
		t.Fatalf("RequestTimeout() = %s, want 45s", got)
	}
	if got := ClientTimeout(); got != 50*time.Second {
		t.Fatalf("ClientTimeout() = %s, want 50s", got)
	}
}

func TestRequestTimeoutIgnoresInvalidEnv(t *testing.T) {
	t.Setenv(EnvRPCTimeout, "nope")

	if got := RequestTimeout(); got != defaultRequestTimeout {
		t.Fatalf("RequestTimeout() = %s, want default %s", got, defaultRequestTimeout)
	}
}
