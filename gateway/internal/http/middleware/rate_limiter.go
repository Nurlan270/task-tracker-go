package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/ratelimiter"
)

type RateLimiterMiddleware struct {
	limiter *ratelimiter.TokenBucket
}

func NewRateLimiterMiddleware(limiter *ratelimiter.TokenBucket) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{limiter: limiter}
}

const keyPrefix = "ratelimit:"

func (m *RateLimiterMiddleware) Limit(
	key KeyFn,
	cap int,
	rate float64,
	interval time.Duration,
) gin.HandlerFunc {
	log := logger.Get().SetOp("http.middleware.rate_limiter.Limit")

	return func(c *gin.Context) {
		allowed, remaining, err := m.limiter.Allow(
			c.Request.Context(),
			keyPrefix+key(c),
			cap,
			rate,
			interval.Seconds(),
		)
		if err != nil {
			log.Error("failed to check for rate limit", zap.Error(err))

			httplib.AbortWithErr(c, errs.ErrInternalServer)

			return
		}

		// Add standard rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(cap))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(
			time.Now().Add(interval).Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatFloat(interval.Seconds(), 'f', 0, 64))

			httplib.AbortWithErr(c, errs.ErrTooManyRequests)

			return
		}

		//	Request allowed, continue
		c.Next()
	}
}

// KeyFn used to add key for storing limit entries.
type KeyFn func(*gin.Context) string

// IPKey returns key string containing user's IP.
func IPKey(c *gin.Context) string {
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// UserIDKey returns key string containing user's ID retrieved from context.
// If it does not find user in context, it will fall back to IPKey.
func UserIDKey(c *gin.Context) string {
	user, exists := c.Get(ctxlib.UserCtxKey)
	if !exists {
		return IPKey(c)
	}

	return fmt.Sprintf("user:%d", user.(*entity.User).ID)
}

// GroupKey returns key string containing passed group in combination with user's IP.
func GroupKey(g string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		return fmt.Sprintf("%s:group:%s", IPKey(c), g)
	}
}
