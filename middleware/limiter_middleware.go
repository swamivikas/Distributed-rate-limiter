package middleware

import (
	"net/http"
	"strings"

	"math/rand"
	"strconv"
	"time"

	"rate-limiter/limiter"
)

// IPExtractor extracts client IP from request (behind proxies X-Forwarded-For first).
func IPExtractor(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip := r.RemoteAddr; ip != "" {
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			return ip[:idx]
		}
		return ip
	}
	// fallback: generate a pseudo-random ID to avoid key collision in tests
	rand.Seed(time.Now().UnixNano())
	return "anon-" + strconv.FormatInt(rand.Int63(), 36)
}

// New constructs the middleware with given limiter and key prefix.
func New(l *limiter.Limiter, keyPrefix string) func(http.Handler) http.Handler {
	if keyPrefix == "" {
		keyPrefix = "ratelimit:"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyPrefix + IPExtractor(r)
			allowed, err := l.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
