package ipc

import "context"

const (
	SurfaceIPC  = "ipc"
	SurfaceHTTP = "http"
)

type surfaceContextKey struct{}
type fileRootsContextKey struct{}

// WithSurface annotates a request with its transport surface.
func WithSurface(ctx context.Context, surface string) context.Context {
	return context.WithValue(ctx, surfaceContextKey{}, surface)
}

// SurfaceFromContext returns the transport surface for a request.
func SurfaceFromContext(ctx context.Context) string {
	surface, _ := ctx.Value(surfaceContextKey{}).(string)
	return surface
}

// WithFileRoots attaches server-side file roots allowed for an HTTP request.
func WithFileRoots(ctx context.Context, roots []string) context.Context {
	return context.WithValue(ctx, fileRootsContextKey{}, append([]string(nil), roots...))
}

// FileRootsFromContext returns a defensive copy of configured file roots.
func FileRootsFromContext(ctx context.Context) []string {
	roots, _ := ctx.Value(fileRootsContextKey{}).([]string)
	return append([]string(nil), roots...)
}
