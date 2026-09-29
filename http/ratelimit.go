package fbhttp

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tomasen/realip"
)

const defaultAuthRateLimit = 20

// authRateLimiter is a minimal in-memory rate limiter for auth endpoints.
// 20 requests per minute per IP, burst 20. Resets automatically.
// Good enough to slow credential stuffing without Redis dependency.
// A limit <= 0 disables the limiter.
type authRateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func newAuthRateLimiter(limit int) *authRateLimiter {
	return &authRateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: time.Minute,
	}
}

// authRateLimitFromEnv lets operators (and the Playwright e2e harness, which
// logs in dozens of times per minute from 127.0.0.1) raise or disable the
// budget with FB_AUTH_RATE_LIMIT. Invalid values keep the secure default.
func authRateLimitFromEnv() int {
	raw := os.Getenv("FB_AUTH_RATE_LIMIT")
	if raw == "" {
		return defaultAuthRateLimit
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 0 {
		return defaultAuthRateLimit
	}
	return limit
}

var globalAuthLimiter = newAuthRateLimiter(authRateLimitFromEnv())

func (l *authRateLimiter) allow(ip string) bool {
	if l.limit <= 0 {
		return true
	}

	now := time.Now()
	cutoff := now.Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()
	entries := l.hits[ip]
	kept := entries[:0]
	for _, t := range entries {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}

func withAuthRateLimit(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !globalAuthLimiter.allow(realip.FromRequest(r)) {
			return http.StatusTooManyRequests, nil
		}
		return fn(w, r, d)
	}
}
