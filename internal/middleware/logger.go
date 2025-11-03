package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		method := c.Request.Method
		uri := c.FullPath()
		if uri == "" {
			uri = c.Request.RequestURI
		}

		logger.Info("HTTP request completed",
			zap.String("method", method),
			zap.String("uri", uri),
			zap.Int("status", status),
			zap.Int("response_size", size),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
	}
}
