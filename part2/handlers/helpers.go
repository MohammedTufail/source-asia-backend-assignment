// Shared utilities used across all Part 2 HTTP handlers.
// writeJSON is the single exit-point for all JSON responses, ensuring
// the Content-Type header is always set correctly. errResp constructs
// the standard error envelope {"error":"...", "message":"..."} used
// consistently across every endpoint for predictable client error handling.

package handlers

import (
	"encoding/json"
	"net/http"
)

// apiError is the standard error envelope returned by all endpoints.
type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// errResp constructs an apiError value ready to be passed to writeJSON.
func errResp(code, message string) apiError {
	return apiError{Error: code, Message: message}
}

// writeJSON encodes v as JSON and writes it to w with the given HTTP status.
// It always sets Content-Type: application/json before writing the header.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}