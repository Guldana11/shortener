package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      string
		compressRequest  bool
		acceptGzip       bool
		expectedResponse string
		expectedStatus   int
	}{
		{
			name:             "обычный запрос без gzip",
			requestBody:      `{"url":"https://example.com"}`,
			compressRequest:  false,
			acceptGzip:       false,
			expectedResponse: `{"result":"http://localhost:8080/`,
			expectedStatus:   http.StatusCreated,
		},
		{
			name:             "gzip-запрос и gzip-ответ",
			requestBody:      `{"url":"https://example.com"}`,
			compressRequest:  true,
			acceptGzip:       true,
			expectedResponse: `{"result":"http://localhost:8080/`,
			expectedStatus:   http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)

			r.Use(GzipMiddleware())

			r.POST("/test", func(c *gin.Context) {
				var req map[string]string
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
					return
				}

				resp := map[string]string{"result": "http://localhost:8080/12345"}
				c.Header("Content-Type", "application/json")
				c.Status(http.StatusCreated)
				if err := json.NewEncoder(c.Writer).Encode(resp); err != nil {
					t.Errorf("failed to encode response: %v", err)
				}
			})

			var reqBody io.Reader
			if tt.compressRequest {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				if _, err := gz.Write([]byte(tt.requestBody)); err != nil {
					t.Fatalf("failed to gzip request body: %v", err)
				}
				if err := gz.Close(); err != nil {
					t.Fatalf("failed to close gzip writer: %v", err)
				}
				reqBody = &buf
			} else {
				reqBody = bytes.NewBufferString(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/test", reqBody)
			req.Header.Set("Content-Type", "application/json")
			if tt.compressRequest {
				req.Header.Set("Content-Encoding", "gzip")
			}
			if tt.acceptGzip {
				req.Header.Set("Accept-Encoding", "gzip")
			}

			c.Request = req
			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if tt.acceptGzip {
				gzReader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				bodyBytes, err = io.ReadAll(gzReader)
				if err != nil {
					t.Fatalf("failed to read gzip response: %v", err)
				}
				if err := gzReader.Close(); err != nil {
					t.Errorf("failed to close gzip reader: %v", err)
				}
			}

			bodyStr := string(bodyBytes)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if !strings.HasPrefix(bodyStr, tt.expectedResponse) {
				t.Errorf("expected body to start with %q, got %q", tt.expectedResponse, bodyStr)
			}

			if tt.acceptGzip {
				if ct := res.Header.Get("Content-Encoding"); ct != "gzip" {
					t.Errorf("expected Content-Encoding gzip, got %s", ct)
				}
			}
		})
	}
}
