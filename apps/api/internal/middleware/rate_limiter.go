package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/sevima/tilik-api/internal/handler"
)

// RateLimiter implements a token bucket rate limiter per IP address
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     int
	lastSeen   time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window duration (e.g., 1 * time.Minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	
	// Cleanup stale visitors every 5 minutes
	go rl.cleanupVisitors()
	
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:   rl.rate,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}
	return v
}

func (v *visitor) allow(rate int, window time.Duration) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen)

	// Refill tokens based on elapsed time
	if elapsed >= window {
		v.tokens = rate
		v.lastSeen = now
	} else {
		// Partial refill based on elapsed time
		tokensToAdd := int(float64(rate) * (elapsed.Seconds() / window.Seconds()))
		v.tokens += tokensToAdd
		if v.tokens > rate {
			v.tokens = rate
		}
		if tokensToAdd > 0 {
			v.lastSeen = now
		}
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an HTTP middleware that applies rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		visitor := rl.getVisitor(ip)

		if !visitor.allow(rl.rate, rl.window) {
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.rate)))
			w.Header().Set("Retry-After", rl.window.String())
			handler.WriteError(w, http.StatusTooManyRequests, "Terlalu banyak permintaan. Silakan coba lagi nanti.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real IP address from the request
// Checks X-Forwarded-For and X-Real-IP headers (for reverse proxy)
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (proxy/load balancer)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list (client IP)
		for i := 0; i < len(forwarded); i++ {
			if forwarded[i] == ',' || forwarded[i] == ' ' {
				return forwarded[:i]
			}
		}
		return forwarded
	}

	// Check X-Real-IP header (nginx)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	// RemoteAddr format: "IP:port"
	ip := r.RemoteAddr
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == ':' {
			return ip[:i]
		}
	}
	return ip
}
