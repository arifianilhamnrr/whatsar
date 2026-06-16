package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/whatsar/whatsar/internal/httputil"
)

type rateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	limit    int
	window   time.Duration
}

func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  requestsPerMinute,
		window: time.Minute,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.RemoteAddr
			}

			if !rl.allow(key) {
				httputil.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Terlalu banyak request, coba lagi nanti")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	times := rl.hits[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= rl.limit {
		rl.hits[key] = filtered
		return false
	}

	rl.hits[key] = append(filtered, now)
	return true
}