package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"product_service/products/internal/usecase/ports"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
	logger   ports.Logger
	stopChan chan struct{}
	stopOnce sync.Once
}

func NewRateLimiter(limit int, window time.Duration, logger ports.Logger) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		logger:   logger,
		stopChan: make(chan struct{}),
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.stopChan)
	})
}

func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()

		if !rl.allow(clientID) {
			rl.logger.Warn("Rate limit exceeded",
				ports.NewField("client_id", clientID),
				ports.NewField("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[clientID]
	if requests == nil {
		rl.requests[clientID] = []time.Time{now}
		return true
	}

	validCount := 0
	for i := 0; i < len(requests); i++ {
		if requests[i].After(windowStart) {
			if validCount != i {
				requests[validCount] = requests[i]
			}
			validCount++
		}
	}

	requests = requests[:validCount]

	if len(requests) >= rl.limit {
		rl.requests[clientID] = requests
		return false
	}

	requests = append(requests, now)
	rl.requests[clientID] = requests

	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopChan:
			rl.mu.Lock()
			now := time.Now()
			windowStart := now.Add(-rl.window)

			for clientID, requests := range rl.requests {
				validCount := 0
				for i := 0; i < len(requests); i++ {
					if requests[i].After(windowStart) {
						if validCount != i {
							requests[validCount] = requests[i]
						}
						validCount++
					}
				}
				requests = requests[:validCount]

				if len(requests) == 0 {
					delete(rl.requests, clientID)
				} else {
					rl.requests[clientID] = requests
				}
			}
			rl.mu.Unlock()
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			windowStart := now.Add(-rl.window)

			for clientID, requests := range rl.requests {
				validCount := 0
				for i := 0; i < len(requests); i++ {
					if requests[i].After(windowStart) {
						if validCount != i {
							requests[validCount] = requests[i]
						}
						validCount++
					}
				}
				requests = requests[:validCount]

				if len(requests) == 0 {
					delete(rl.requests, clientID)
				} else {
					rl.requests[clientID] = requests
				}
			}
			rl.mu.Unlock()
		}
	}
}

