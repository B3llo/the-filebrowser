package fbhttp

import (
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
)

// authRateLimiter is a minimal in-memory rate limiter for auth endpoints.
// 20 requests per minute per IP, burst 20. Resets automatically.
// Good enough to slow credential stuffing without Redis dependency.
type authRateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

var globalAuthLimiter = &authRateLimiter{
	hits:   make(map[string][]time.Time),
	limit:  20,
	window: time.Minute,
}

func (l *authRateLimiter) allow(ip string) bool {
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
