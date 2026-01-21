package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

 
type RateLimiter struct {
	tokens      map[string]float64
	maxTokens   float64
	refillRate  float64  
	lastRefill  map[string]time.Time
	mu          sync.RWMutex
}

 
func NewRateLimiter(maxTokens float64, requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		tokens:     make(map[string]float64),
		maxTokens:  maxTokens,
		refillRate: requestsPerSecond,
		lastRefill: make(map[string]time.Time),
	}
}

 
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	 
	lastRefill, exists := rl.lastRefill[clientID]
	if !exists {
		rl.tokens[clientID] = rl.maxTokens
		rl.lastRefill[clientID] = now
	} else {
		elapsed := now.Sub(lastRefill).Seconds()
		tokensToAdd := elapsed * rl.refillRate
		rl.tokens[clientID] = min(rl.maxTokens, rl.tokens[clientID]+tokensToAdd)
		rl.lastRefill[clientID] = now
	}

	 
	if rl.tokens[clientID] >= 1.0 {
		rl.tokens[clientID]--
		return true
	}

	return false
}

 
func RateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(200, 100)  
	logger, _ := zap.NewDevelopment()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			logger.Warn("Rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "RATE_LIMITED",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

 
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

 
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

 
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
