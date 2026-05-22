package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Rate limiter unit tests

func TestRateLimiter_AllowsUpToFive(t *testing.T) {
	rl := NewRateLimiter()
	for i := 0; i < maxRequests; i++ {
		if !rl.Allow("user1") {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
	if rl.Allow("user1") {
		t.Fatal("expected 6th request to be rejected")
	}
}

func TestRateLimiter_IndependentUsers(t *testing.T) {
	rl := NewRateLimiter()
	for i := 0; i < maxRequests; i++ {
		rl.Allow("alice")
	}
	// alice is now at limit; bob should still be free
	if !rl.Allow("bob") {
		t.Fatal("bob should not be affected by alice's limit")
	}
}

func TestRateLimiter_ConcurrentSafety(t *testing.T) {
	rl := NewRateLimiter()
	var wg sync.WaitGroup
	accepted := 0
	var mu sync.Mutex

	// Fire 50 concurrent requests for the same user.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("stress") {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if accepted != maxRequests {
		t.Fatalf("expected exactly %d accepted, got %d", maxRequests, accepted)
	}
}

func TestRateLimiter_StatsRejectedCumulative(t *testing.T) {
	rl := NewRateLimiter()
	for i := 0; i < maxRequests+3; i++ {
		rl.Allow("user1")
	}
	stats := rl.Stats()
	if stats["user1"].RejectedTotal != 3 {
		t.Fatalf("expected 3 cumulative rejections, got %d", stats["user1"].RejectedTotal)
	}
}

// HTTP handler tests

func newTestServer() (*httptest.Server, *RateLimiter) {
	rl := NewRateLimiter()
	mux := http.NewServeMux()
	mux.HandleFunc("/request", handleRequest(rl))
	mux.HandleFunc("/stats", handleStats(rl))
	return httptest.NewServer(mux), rl
}

func postRequest(t *testing.T, server *httptest.Server, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(server.URL+"/request", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /request failed: %v", err)
	}
	return resp
}

func TestHandler_AcceptsValidRequest(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	resp := postRequest(t, srv, `{"user_id":"alice","payload":{"key":"value"}}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestHandler_RejectsMissingUserID(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	resp := postRequest(t, srv, `{"payload":42}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_RejectsEmptyUserID(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	resp := postRequest(t, srv, `{"user_id":"","payload":1}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_RejectsMissingPayload(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	resp := postRequest(t, srv, `{"user_id":"alice"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_RejectsInvalidJSON(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	resp := postRequest(t, srv, `not json`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_RateLimit(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	body := `{"user_id":"limited","payload":true}`
	for i := 0; i < maxRequests; i++ {
		resp := postRequest(t, srv, body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("request %d: expected 201, got %d", i+1, resp.StatusCode)
		}
	}

	resp := postRequest(t, srv, body)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 6th request, got %d", resp.StatusCode)
	}
}

func TestHandler_StatsEndpoint(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	body := `{"user_id":"statuser","payload":1}`
	postRequest(t, srv, body)
	postRequest(t, srv, body)

	resp, err := http.Get(srv.URL + "/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode stats: %v", err)
	}

	stats, ok := result.Users["statuser"]
	if !ok {
		t.Fatal("expected statuser in stats response")
	}
	if stats.WindowAccepted != 2 {
		t.Fatalf("expected window_accepted=2, got %d", stats.WindowAccepted)
	}
}

// TestHandler_ConcurrentRequests fires 20 goroutines simultaneously for one
// user and verifies exactly maxRequests (5) are accepted.
func TestHandler_ConcurrentRequests(t *testing.T) {
	srv, _ := newTestServer()
	defer srv.Close()

	body := `{"user_id":"concurrent","payload":"data"}`
	var wg sync.WaitGroup
	var mu sync.Mutex
	accepted := 0

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := postRequest(t, srv, body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if accepted != maxRequests {
		t.Fatalf("expected exactly %d accepted, got %d", maxRequests, accepted)
	}
}

// TestRateLimiter_WindowExpiry verifies that the window resets after 1 minute.
// This test manipulates time by directly inserting old timestamps.
func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter()

	// Fill up the window with old timestamps (2 minutes ago).
	rl.mu.Lock()
	rl.windows["expuser"] = &windowEntry{
		timestamps: []time.Time{
			time.Now().Add(-2 * time.Minute),
			time.Now().Add(-2 * time.Minute),
			time.Now().Add(-2 * time.Minute),
			time.Now().Add(-2 * time.Minute),
			time.Now().Add(-2 * time.Minute),
		},
	}
	rl.mu.Unlock()

	// All 5 slots are "old" — a fresh request should be allowed.
	if !rl.Allow("expuser") {
		t.Fatal("expected request to be allowed after window expiry")
	}
}