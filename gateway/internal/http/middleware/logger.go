package middleware

import (
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Get().WithOptions(zap.WithCaller(false)).
			Info("Incoming request",
				zap.Int("status_code", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.Duration("duration", time.Since(start).Truncate(time.Millisecond)),
			)
	}
}
