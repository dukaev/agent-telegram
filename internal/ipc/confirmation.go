package ipc

import "context"

type confirmationContextKey struct{}

// WithConfirmation records whether the caller explicitly confirmed a
// destructive or paid operation.
func WithConfirmation(ctx context.Context, confirmed bool) context.Context {
	return context.WithValue(ctx, confirmationContextKey{}, confirmed)
}

// IsConfirmed reports whether an operation was explicitly confirmed.
func IsConfirmed(ctx context.Context) bool {
	confirmed, _ := ctx.Value(confirmationContextKey{}).(bool)
	return confirmed
}
