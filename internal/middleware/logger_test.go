package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "should log request and response info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			encoderCfg := zap.NewDevelopmentEncoderConfig()
			encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			core := zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderCfg),
				zapcore.AddSync(&buf),
				zap.InfoLevel,
			)
			logger := zap.New(core)
			defer logger.Sync()

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(LoggerMiddleware(logger))

			r.GET("/ping", func(c *gin.Context) {
				c.String(http.StatusOK, "pong")
			})

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			logOut := buf.String()
			if logOut == "" {
				t.Fatal("expected log output, got empty string")
			}

			expectedSubstrings := []string{
				"HTTP request completed",
				`"method": "GET"`,
				`"uri": "/ping"`,
				`"status": 200`,
				`"response_size":`,
				`"duration_ms":`,
			}

			for _, s := range expectedSubstrings {
				if !strings.Contains(logOut, s) {
					t.Errorf("expected log to contain %q, got log:\n%s", s, logOut)
				}
			}
		})
	}
}
