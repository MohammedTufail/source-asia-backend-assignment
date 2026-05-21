package main

import (
	"encoding/json"
	"net/http"
)

// requestBody is the expected JSON payload for POST /request.
type requestBody struct {
	UserID  string          `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

// errorResponse is a standard JSON error envelope.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// acceptedResponse is the success body for POST /request.
type acceptedResponse struct {
	Status  string `json:"status"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// statsResponse wraps per-user stats for GET /stats.
type statsResponse struct {
	Users map[string]UserStats `json:"users"`
	Note  string               `json:"note"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handleRequest processes POST /request.
func handleRequest(rl *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method_not_allowed"})
			return
		}

		var body requestBody
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error:   "invalid_json",
				Message: "Request body must be valid JSON with user_id and payload fields.",
			})
			return
		}

		if body.UserID == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error:   "missing_user_id",
				Message: "user_id is required and must be a non-empty string.",
			})
			return
		}

		if len(body.Payload) == 0 || string(body.Payload) == "null" {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error:   "missing_payload",
				Message: "payload is required and must be a valid JSON value.",
			})
			return
		}

		if !rl.Allow(body.UserID) {
			writeJSON(w, http.StatusTooManyRequests, errorResponse{
				Error:   "rate_limit_exceeded",
				Message: "Rate limit exceeded: maximum 5 requests per user per 1-minute rolling window.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, acceptedResponse{
			Status:  "accepted",
			UserID:  body.UserID,
			Message: "Request accepted successfully.",
		})
	}
}

// handleStats processes GET /stats.
func handleStats(rl *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method_not_allowed"})
			return
		}

		stats := rl.Stats()
		writeJSON(w, http.StatusOK, statsResponse{
			Users: stats,
			Note:  "window_accepted reflects requests in the current 1-minute rolling window. rejected_cumulative is the total rejected count since server start.",
		})
	}
}