package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex // protects clients
	clients map[string]*client // Take client's IP as key
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {

	rl := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}

	// cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {

	// lock
	rl.mu.Lock()

	// unlock when done
	defer rl.mu.Unlock()

	// get client's ip
	c, exists := rl.clients[ip]

	// client baru atau window expired
	if !exists || time.Since(c.lastSeen) > rl.window {

		rl.clients[ip] = &client{
			count:    1,
			lastSeen: time.Now(),
		}

		return true
	}

	c.count++
	c.lastSeen = time.Now()

	return c.count <= rl.limit
}

func (rl *RateLimiter) cleanup() {

	// every 1 minute
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {

		// lock
		rl.mu.Lock()

		for ip, c := range rl.clients {

			// Delete when already expired
			if time.Since(c.lastSeen) > rl.window*2 {
				delete(rl.clients, ip)
			}
		}

		// unlock
		rl.mu.Unlock()
	}
}

func RateLimitMiddleware(
	rl *RateLimiter,
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Take IP and ignore port
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			ip = r.RemoteAddr
		}

		// If 'Allow' returns false
		if !rl.Allow(ip) {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}