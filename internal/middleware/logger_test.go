package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should log request and response info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			log.SetLevel(log.InfoLevel)
			log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(LoggerMiddleware())

			r.GET("/ping", func(c *gin.Context) {
				c.String(http.StatusOK, "pong")
			})

			req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			logOut := buf.String()
			expectedFields := []string{"method=GET", "uri=/ping", "status=200", "response_size=", "duration_ms="}
			for _, field := range expectedFields {
				if !strings.Contains(logOut, field) {
					t.Errorf("expected log to contain %q, got %q", field, logOut)
				}
			}
		})
	}
}
