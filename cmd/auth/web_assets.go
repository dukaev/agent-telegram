package auth

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

func serveAuthIndex(w http.ResponseWriter, r *http.Request) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, err := authWebAssets.ReadFile("web_dist/index.html")
	if err != nil {
		http.Error(w, "auth assets are missing", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(data))
}

func handleAuthAsset(w http.ResponseWriter, r *http.Request) {
	setAuthHeaders(w)

	assetPath := strings.TrimPrefix(r.URL.Path, "/auth/")
	assetPath = path.Clean(assetPath)
	if !strings.HasPrefix(assetPath, "assets/") {
		http.NotFound(w, r)
		return
	}

	data, err := authWebAssets.ReadFile("web_dist/" + assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if contentType := mime.TypeByExtension(path.Ext(assetPath)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(w, r, path.Base(assetPath), time.Time{}, bytes.NewReader(data))
}

func writeAuthResponse(w http.ResponseWriter, r *http.Request, status int, state authClientState) {
	if wantsJSON(r) {
		writeAuthState(w, status, state)
		return
	}
	http.Error(w, state.Error, status)
}

func writeAuthState(w http.ResponseWriter, status int, state authClientState) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		slog.Debug("failed to write auth state", "error", err)
	}
}

func writeAuthPeers(w http.ResponseWriter, status int, state authPeersState) {
	setAuthHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		slog.Debug("failed to write auth peers", "error", err)
	}
}

func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || isJSONRequest(r)
}

func isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func setAuthHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", strings.Join([]string{
		"default-src 'none'",
		"img-src data: 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		"connect-src 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"base-uri 'none'",
	}, "; "))
}
