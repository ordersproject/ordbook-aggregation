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
			// 如果该 IP 地址之前访问过，检查访问次数和时间间隔
			if store.AccessTimes >= limiter.maxAccessTimes && now.Sub(store.LastAccess) < limiter.interval {
				// 如果访问次数超限，返回 429 Too Many Requests 状态码
				c.JSON(http.StatusTooManyRequests, gin.H{"message": "Too Many Requests"})
				c.Abort()
				return
			}
			store.AccessTimes++
			store.LastAccess = now
		} else {
			// 如果该 IP 地址之前没有访问过，初始化访问记录
			limiter.ips[ip] = &IPStore{
				AccessTimes: 1,
				LastAccess:  now,
			}
		}
		c.Next()
	}
}