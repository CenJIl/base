package web

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"golang.org/x/time/rate"
)

// RateLimiterConfig rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Requests per second
	BurstSize        int           // Maximum burst size
	CleanupInterval  time.Duration // Cleanup interval
}

// IPRateLimiter IP-based rate limiter
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   *RateLimiterConfig
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config: &RateLimiterConfig{
			RequestsPerSecond: rps,
			BurstSize:        burst,
			CleanupInterval:  5 * time.Minute,
		},
	}
}

// Allow checks if the request from given IP is allowed
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize)
		rl.limiters[ip] = limiter
	}

	return limiter.Allow()
}

// Cleanup removes stale limiters
func (rl *IPRateLimiter) Cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			rl.limiters = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
			logger.Debugf("Rate limiter cleanup completed")
		}
	}()
}

var (
	globalIPRateLimiter *IPRateLimiter
)

// InitRateLimiter initializes global rate limiter
func InitRateLimiter(rps float64, burst int) {
	globalIPRateLimiter = NewIPRateLimiter(rps, burst)
	globalIPRateLimiter.Cleanup()
	logger.Infof("Rate limiter initialized: %v req/s, burst %d", rps, burst)
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if globalIPRateLimiter == nil {
			c.Next(ctx)
			return
		}

		clientIP := c.ClientIP()
		if !globalIPRateLimiter.Allow(clientIP) {
			logger.Warnf("Rate limit exceeded for IP: %s", clientIP)
			c.JSON(consts.StatusTooManyRequests, map[string]any{
				"code": 429,
				"message": "Rate limit exceeded",
				"data": map[string]any{
					"limit": fmt.Sprintf("%.0f req/s", globalIPRateLimiter.config.RequestsPerSecond),
				},
			})
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}
