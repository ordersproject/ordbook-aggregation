package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

type IPStore struct {
	AccessTimes int
	LastAccess  time.Time
}

type IPRateLimiter struct {
	ips map[string]*IPStore
	interval time.Duration
	maxAccessTimes int
	mu  sync.Mutex
}

func NewIPRateLimiter(interval time.Duration, maxAccessTimes int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*IPStore),
		interval: interval,
		maxAccessTimes: maxAccessTimes,
		mu:  sync.Mutex{},
	}
}

func IPRateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter.mu.Lock()
		defer limiter.mu.Unlock()
		store, ok := limiter.ips[ip]
		now := time.Now()
		if ok {
			if store.AccessTimes >= limiter.maxAccessTimes && now.Sub(store.LastAccess) < limiter.interval {
				c.JSON(http.StatusTooManyRequests, gin.H{"message": "Too Many Requests"})
				c.Abort()
				return
			}
			store.AccessTimes++
			store.LastAccess = now
		} else {
			limiter.ips[ip] = &IPStore{
				AccessTimes: 1,
				LastAccess:  now,
			}
		}
		c.Next()
	}
}