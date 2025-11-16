package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mashhkensss/PR-service/internal/http/httperror"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

type bucket struct {
	tokens     int
	lastRefill time.Time
	lastSeen   time.Time
}

type RateLimiter struct {
	limit          int
	interval       time.Duration
	ttl            time.Duration
	trustForwarded bool
	buckets        map[string]*bucket
	mu             sync.Mutex
	nextCleanup    time.Time
}

func NewRateLimiter(limit int, interval time.Duration, trustForwarded bool) *RateLimiter {
	if interval <= 0 {
		interval = time.Second
	}
	ttl := interval * 10
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &RateLimiter{
		limit:          limit,
		interval:       interval,
		ttl:            ttl,
		trustForwarded: trustForwarded,
		buckets:        make(map[string]*bucket),
		nextCleanup:    time.Now().Add(ttl),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.clientIP(r)
		if allowed, retryAfter := rl.allow(key); !allowed {
			if retryAfter > 0 {
				secs := int(retryAfter / time.Second)
				if retryAfter%time.Second != 0 {
					secs++
				}
				if secs <= 0 {
					secs = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(secs))
			}
			status, resp := httperror.RateLimited()
			response.ErrorResponse(w, status, resp)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.After(rl.nextCleanup) {
		rl.cleanup(now)
	}

	b := rl.buckets[key]
	if b == nil {
		b = &bucket{
			tokens:     rl.limit,
			lastRefill: now,
			lastSeen:   now,
		}
		rl.buckets[key] = b
	}

	if since := now.Sub(b.lastRefill); since >= rl.interval {
		refills := int(since / rl.interval)
		b.tokens += refills * rl.limit
		if b.tokens > rl.limit {
			b.tokens = rl.limit
		}
		b.lastRefill = now
	}
	b.lastSeen = now

	if b.tokens <= 0 {
		wait := rl.interval - now.Sub(b.lastRefill)
		if wait <= 0 {
			wait = rl.interval
		}
		return false, wait
	}
	b.tokens--
	return true, 0
}

func (rl *RateLimiter) cleanup(now time.Time) {
	expireBefore := now.Add(-rl.ttl)
	for key, b := range rl.buckets {
		if b.lastSeen.Before(expireBefore) {
			delete(rl.buckets, key)
		}
	}
	rl.nextCleanup = now.Add(rl.ttl)
}

func (rl *RateLimiter) clientIP(r *http.Request) string {
	if rl.trustForwarded {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				if ip := strings.TrimSpace(parts[0]); ip != "" {
					return ip
				}
			}
		}
		if xr := r.Header.Get("X-Real-IP"); xr != "" {
			return strings.TrimSpace(xr)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
