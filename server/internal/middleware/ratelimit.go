package middleware

import (
	"net/http"
	"sync"
	"time"
)

// Token bucket rate limiter per IP
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int // tokens per minute
	burst    int
	banLimit int
	banDur   time.Duration
}

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	reqCount  int // requests in current window
	windowStart time.Time
	bannedUntil time.Time
}

func NewRateLimiter(requestsPerMin, burst, banThreshold, banDurationMin int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMin,
		burst:    burst,
		banLimit: banThreshold,
		banDur:   time.Duration(banDurationMin) * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if !rl.allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{tokens: float64(rl.burst), lastSeen: now, windowStart: now}
		rl.visitors[ip] = v
	}

	// Check ban
	if now.Before(v.bannedUntil) {
		return false
	}

	// Refill tokens
	elapsed := now.Sub(v.lastSeen).Minutes()
	v.tokens += elapsed * float64(rl.rate)
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastSeen = now

	// Abuse detection: count requests in 1-minute window
	if now.Sub(v.windowStart) > time.Minute {
		v.reqCount = 0
		v.windowStart = now
	}
	v.reqCount++
	if v.reqCount > rl.banLimit {
		v.bannedUntil = now.Add(rl.banDur)
		return false
	}

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// Periodically remove stale entries
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, v := range rl.visitors {
			if v.lastSeen.Before(cutoff) && time.Now().After(v.bannedUntil) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (client IP when behind trusted proxy)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	// Strip port
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
