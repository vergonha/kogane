package auth

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	LoginAttemptLimit  = 10
	LoginAttemptWindow = 15 * time.Minute
)

type Limiter struct {
	attempts map[string]attempt
	window   time.Duration
	limit    int
	mu       sync.Mutex
}

type attempt struct {
	resetAt time.Time
	count   int
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		attempts: make(map[string]attempt),
		window:   window,
		limit:    limit,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	for k, a := range l.attempts {
		if now.After(a.resetAt) {
			delete(l.attempts, k)
		}
	}

	current, ok := l.attempts[key]
	if !ok {
		current = attempt{resetAt: now.Add(l.window)}
	}

	current.count++
	l.attempts[key] = current

	return current.count <= l.limit
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, key)
}

func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
