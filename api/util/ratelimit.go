package util

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	tokens          int
	lastRefreshTime time.Time
	mu              sync.Mutex
}

func NewRateLimiter(tokensPerMinute int) *RateLimiter {
	return &RateLimiter{
		tokens:          tokensPerMinute,
		lastRefreshTime: time.Now(),
	}
}

func (r *RateLimiter) MiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		r.mu.Lock()
		defer r.mu.Unlock()

		now := time.Now()
		elapsed := now.Sub(r.lastRefreshTime)
		if elapsed > time.Minute {
			r.tokens = 1
			r.lastRefreshTime = now
		}

		if r.tokens > 0 {
			r.tokens--
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
		}
	}
}
