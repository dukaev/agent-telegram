package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSenderDeliveryRetryAndDrain(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content-type = %q", r.Header.Get("Content-Type"))
		}
		if attempts == 1 {
			http.Error(w, "try again", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	var errors []string
	sender := New(server.URL, WithRetries(1), WithRetryDelay(time.Millisecond), WithOnError(func(msg string) {
		errors = append(errors, msg)
	}))
	sender.deliver(context.Background(), []byte(`{"ok":true}`))
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if len(errors) != 0 {
		t.Fatalf("unexpected errors = %#v", errors)
	}

	sender.Send([]byte(`{"ok":true}`))
	if sender.QueueLen() != 1 || sender.DroppedCount() != 0 {
		t.Fatalf("queue=%d dropped=%d", sender.QueueLen(), sender.DroppedCount())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sender.Run(ctx)
	select {
	case <-sender.Done():
	default:
		t.Fatal("done should be closed")
	}
}

func TestSenderErrorAndDrop(t *testing.T) {
	var errorMsg string
	sender := New("://bad-url", WithRetries(0), WithOnError(func(msg string) {
		errorMsg = msg
	}))
	sender.deliver(context.Background(), []byte("{}"))
	if !strings.Contains(errorMsg, "giving up") {
		t.Fatalf("error msg = %q", errorMsg)
	}

	sender = New("http://example.invalid")
	for i := 0; i < defaultBufferSize+1; i++ {
		sender.Send([]byte("{}"))
	}
	if sender.DroppedCount() != 1 {
		t.Fatalf("dropped = %d", sender.DroppedCount())
	}
}
