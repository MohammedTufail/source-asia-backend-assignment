package main

import (
	"sync"
	"time"
)

const (
	windowDuration = 1 * time.Minute
	maxRequests    = 5
)

// windowEntry tracks request timestamps within the rolling window.
type windowEntry struct {
	timestamps []time.Time
}

// UserStats holds the stats for a single user.
type UserStats struct {
	Accepted        int `json:"accepted"`
	RejectedTotal   int `json:"rejected_cumulative"` // cumulative across all windows
	WindowAccepted  int `json:"window_accepted"`     // accepted in current 1-min window
}

// RateLimiter is a concurrent-safe, per-user rolling-window rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string]*windowEntry
	// cumulative rejected counts per user (never reset)
	rejected map[string]int
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		windows:  make(map[string]*windowEntry),
		rejected: make(map[string]int),
	}
}

// Allow returns true if the request is within the rate limit for the user.
// It also updates internal counters. This is the single lock-holding operation
// so concurrent callers for the same user_id are serialised here.
func (rl *RateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-windowDuration)

	entry, ok := rl.windows[userID]
	if !ok {
		entry = &windowEntry{}
		rl.windows[userID] = entry
	}

	// Evict timestamps outside the rolling window.
	valid := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	entry.timestamps = valid

	if len(entry.timestamps) >= maxRequests {
		rl.rejected[userID]++
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

// Stats returns per-user statistics. Caller must not hold the mutex.
func (rl *RateLimiter) Stats() map[string]UserStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-windowDuration)

	result := make(map[string]UserStats, len(rl.windows))

	// Collect all user IDs (windows + users who may only have rejections).
	allUsers := make(map[string]struct{})
	for id := range rl.windows {
		allUsers[id] = struct{}{}
	}
	for id := range rl.rejected {
		allUsers[id] = struct{}{}
	}

	for id := range allUsers {
		windowAccepted := 0
		totalAccepted := 0
		if entry, ok := rl.windows[id]; ok {
			for _, ts := range entry.timestamps {
				totalAccepted++ // every stored timestamp = accepted request
				if ts.After(cutoff) {
					windowAccepted++
				}
			}
		}
		result[id] = UserStats{
			Accepted:       totalAccepted,
			RejectedTotal:  rl.rejected[id],
			WindowAccepted: windowAccepted,
		}
	}
	return result
}