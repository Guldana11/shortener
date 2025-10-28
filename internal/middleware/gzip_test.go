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
				json.NewEncoder(c.Writer).Encode(resp)
			})

			var reqBody io.Reader
			if tt.compressRequest {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, err := gz.Write([]byte(tt.requestBody))
				if err != nil {
					t.Fatalf("failed to gzip request body: %v", err)
				}
				gz.Close()
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
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)

			if tt.acceptGzip {
				gzReader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				bodyBytes, _ = io.ReadAll(gzReader)
				gzReader.Close()
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
