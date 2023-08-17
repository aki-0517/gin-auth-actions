package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	limiter := NewRateLimiter(1) // 1 request per minute
	r.Use(limiter.MiddleWare())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// First request should pass
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Second request within a minute should be blocked
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusTooManyRequests, resp.Code)

	// After waiting for more than a minute, the next request should pass
	limiter.lastRefreshTime = limiter.lastRefreshTime.Add(-time.Minute)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
