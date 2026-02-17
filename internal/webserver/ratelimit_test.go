package webserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware(t *testing.T) {
	e := echo.New()
	
	// Create a handler that uses the middleware
	// Limit: 2 requests per second, Burst: 2
	handler := RateLimitMiddleware(rate.Limit(2), 2)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Helper to make requests
	makeRequest := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Mock RealIP since middleware uses it
		req.Header.Set(echo.HeaderXRealIP, "127.0.0.1")
		
		_ = handler(c)
		return rec.Code
	}

	// 1. First request should succeed
	assert.Equal(t, http.StatusOK, makeRequest())

	// 2. Second request should succeed (burst 2)
	assert.Equal(t, http.StatusOK, makeRequest())

	// 3. Third request should fail (limit exceeded)
	assert.Equal(t, http.StatusTooManyRequests, makeRequest())

	// 4. Wait for token refili (0.5s for 1 token if rate is 2/s)
	// safe wait 600ms
	time.Sleep(600 * time.Millisecond)

	// 5. Request should succeed again
	assert.Equal(t, http.StatusOK, makeRequest())
}

func TestRateLimitMultiIP(t *testing.T) {
	e := echo.New()
	limit := rate.Limit(1) // 1 req/s
	burst := 1
	handler := RateLimitMiddleware(limit, burst)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	makeRequest := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		req.Header.Set(echo.HeaderXRealIP, ip)
		_ = handler(c)
		return rec.Code
	}

	// IP 1: Success
	assert.Equal(t, http.StatusOK, makeRequest("1.1.1.1"))
	// IP 1: Fail immediately
	assert.Equal(t, http.StatusTooManyRequests, makeRequest("1.1.1.1"))

	// IP 2: Should Success (independent limit)
	assert.Equal(t, http.StatusOK, makeRequest("2.2.2.2"))
}
