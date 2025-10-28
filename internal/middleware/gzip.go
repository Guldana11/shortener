package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
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
			defer gzWriter.Close()

			c.Writer = &gzipResponseWriter{
				ResponseWriter: c.Writer,
				Writer:         gzWriter,
			}
		}

		c.Next()
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
		return w.Writer.Write(data)
	}
	return w.ResponseWriter.Write(data)
}
