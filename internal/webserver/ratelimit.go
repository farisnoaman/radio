package webserver

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// IPRateLimiter stores rate limiters for each IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IPRateLimiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// cleanup routine to remove old entries to prevent memory leak
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			i.mu.Lock()
			// Simple cleanup: Clear all. For production, track last seen time.
			// Re-creating map is cheap enough for this use case if map gets too big
			if len(i.ips) > 10000 {
				i.ips = make(map[string]*rate.Limiter)
			}
			i.mu.Unlock()
		}
	}()

	return i
}

// GetLimiter returns the rate limiter for the provided IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware creates a middleware for rate limiting
// limit: requests per second
// burst: max burst
func RateLimitMiddleware(limit rate.Limit, burst int) echo.MiddlewareFunc {
	ipLimiter := NewIPRateLimiter(limit, burst)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := ipLimiter.GetLimiter(ip)
			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error":   "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests, please try again later.",
				})
			}
			return next(c)
		}
	}
}
