package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func LoggerMiddleware() gin.HandlerFunc {
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

		log.WithFields(log.Fields{
			"method":        method,
			"uri":           uri,
			"status":        status,
			"response_size": size,
			"duration_ms":   duration.Milliseconds(),
		}).Info("HTTP request completed")
	}
}
