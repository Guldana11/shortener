// Package middleware содержит промежуточное ПО для HTTP-сервера на Gin,
// включая сжатие и распаковку gzip.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware возвращает middleware для Gin, который поддерживает сжатие и распаковку gzip.
//
// Особенности работы:
// 1. Если запрос содержит заголовок "Content-Encoding: gzip", тело запроса автоматически распаковывается.
// 2. Если клиент указывает "Accept-Encoding: gzip", ответ сервера сжимается gzip.
// 3. Для ответа используется обёртка gzipResponseWriter, которая реализует интерфейс gin.ResponseWriter.
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid gzip body"})
				return
			}
			defer gzReader.Close()
			c.Request.Body = io.NopCloser(gzReader)
		}

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Writer.Header().Set("Content-Encoding", "gzip")
			gzWriter := gzip.NewWriter(c.Writer)
			gzw := &gzipResponseWriter{
				ResponseWriter: c.Writer,
				Writer:         gzWriter,
			}
			c.Writer = gzw

			c.Next()

			gzWriter.Close()
			return
		}

		c.Next()
	}
}

// gzipResponseWriter оборачивает gin.ResponseWriter и Writer для сжатия gzip.
type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

// Write пишет данные через gzip.Writer.
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}
