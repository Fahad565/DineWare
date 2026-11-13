package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dineware-backend/db"
	"dineware-backend/models"
)

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	store db.Store
}

// NewHandler constructs a new Handler instance.
func NewHandler(store db.Store) *Handler {
	return &Handler{store: store}
}

// WriteJSON sends a JSON response with status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError sends a formatted JSON error response.
func WriteError(w http.ResponseWriter, status int, message string, details string) {
	WriteJSON(w, status, models.ErrorResponse{
		Error:   message,
		Details: details,
	})
}

// ExtractID extracts the path parameter or falls back to URL parsing.
func ExtractID(r *http.Request, prefix, suffix string) string {
	// First attempt Go 1.22+ PathValue
	if val := r.PathValue("id"); val != "" {
		return val
	}

	// Fallback to manual path extraction
	path := r.URL.Path
	if prefix != "" && strings.HasPrefix(path, prefix) {
		path = strings.TrimPrefix(path, prefix)
	}
	if suffix != "" && strings.HasSuffix(path, suffix) {
		path = strings.TrimSuffix(path, suffix)
	}
	return strings.Trim(path, "/")
}

// CORSMiddleware wraps an http.Handler with cross-origin headers and handles OPTIONS preflights.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, ApiKey")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
