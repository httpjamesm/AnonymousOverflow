package middleware

import (
	"anonymousoverflow/config"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ratelimitWindow   = time.Minute
	ratelimitRequests = 30
)

type limiter struct {
	mu          sync.Mutex
	now         func() time.Time
	windowStart time.Time
	counts      map[string]int
}

func newLimiter(now func() time.Time) *limiter {
	return &limiter{now: now, windowStart: now(), counts: map[string]int{}}
}

func (l *limiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now := l.now(); now.Sub(l.windowStart) >= ratelimitWindow {
		l.windowStart = now
		l.counts = map[string]int{}
	}
	l.counts[ip]++
	return l.counts[ip] <= ratelimitRequests
}

func Ratelimit() gin.HandlerFunc {
	return ratelimit(newLimiter(time.Now))
}

func ratelimit(l *limiter) gin.HandlerFunc {
	return func(c *gin.Context) {

		if os.Getenv("DISABLE_RATELIMIT") == "true" {
			c.Next()
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		if !l.allow(c.ClientIP()) {
			c.HTML(429, "home.html", gin.H{
				"errorMessage": "You have exceeded the request limit. Please try again in a minute.",
				"version":      config.Version,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
