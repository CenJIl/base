package web

import (
	"context"
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
	BurstSize         int           // Maximum burst size
	CleanupInterval   time.Duration // Cleanup interval
}

// IPRateLimiter IP-based rate limiter
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   *RateLimiterConfig
	stop     chan struct{}
	stopOnce sync.Once
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 1
	}
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   &RateLimiterConfig{RequestsPerSecond: rps, BurstSize: burst, CleanupInterval: 5 * time.Minute},
		stop:     make(chan struct{}),
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

// Cleanup starts stale limiter cleanup and returns immediately.
func (rl *IPRateLimiter) Cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.mu.Lock()
				rl.limiters = make(map[string]*rate.Limiter)
				rl.mu.Unlock()
			case <-rl.stop:
				return
			}
		}
	}()
}

// Close stops cleanup and releases limiter state.
func (rl *IPRateLimiter) Close() {
	if rl != nil {
		rl.stopOnce.Do(func() { close(rl.stop) })
	}
}

var (
	globalIPRateLimiter *IPRateLimiter
)

// InitRateLimiter initializes global rate limiter
// InitRateLimiter initializes the global rate limiter.
func InitRateLimiter(rps float64, burst int) {
	if globalIPRateLimiter != nil {
		globalIPRateLimiter.Close()
	}
	globalIPRateLimiter = NewIPRateLimiter(rps, burst)
	globalIPRateLimiter.Cleanup()
	logger.Infof("Rate limiter initialized: %v req/s, burst %d", rps, burst)
}

// CloseRateLimiter stops the global rate limiter cleanup loop.
func CloseRateLimiter() {
	if globalIPRateLimiter != nil {
		globalIPRateLimiter.Close()
		globalIPRateLimiter = nil
	}
}
func RateLimitMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if globalIPRateLimiter == nil {
			c.Next(ctx)
			return
		}

		clientIP := c.ClientIP()
		if !globalIPRateLimiter.Allow(clientIP) {
			logger.Warnf("Rate limit exceeded for IP: %s", clientIP)
			c.Header("Retry-After", "1")
			c.JSON(consts.StatusTooManyRequests, Fail(429, "rate limit exceeded"))
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
