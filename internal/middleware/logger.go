// Package middleware содержит промежуточное ПО для HTTP-сервера на Gin,
// включая логирование и сжатие gzip.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware возвращает middleware для Gin, который логирует HTTP-запросы.
//
// Параметры:
// - logger: экземпляр *zap.Logger для записи логов.
//
// Особенности работы:
// 1. Замеряет время обработки запроса.
// 2. Логирует метод запроса, URI, HTTP-статус, размер ответа и продолжительность в миллисекундах.
// 3. URI берётся через c.FullPath(); если он пустой, используется c.Request.RequestURI.
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
