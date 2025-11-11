package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger(buf *bytes.Buffer) *zap.Logger {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)
	return zap.New(core)
}

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedPrefix string
	}{
		{"успешное создание короткой ссылки", "https://example.com", http.StatusCreated, "http://localhost:8080/"},
		{"пустое тело запроса", "", http.StatusBadRequest, http.StatusText(http.StatusBadRequest)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)

			repo := repository.NewURLRepository("")
			h := &URLHandler{Repo: repo, BaseURL: "http://localhost:8080", logger: logger}

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))

			h.PostHandler(c)

			res := rec.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if !strings.HasPrefix(string(body), tt.expectedPrefix) {
				t.Errorf("expected body prefix %q, got %q", tt.expectedPrefix, string(body))
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := repository.NewURLRepository("")
	repo.CreateWithID("abc123", "https://example.com")
	h := &URLHandler{Repo: repo, BaseURL: "http://localhost:8080", logger: zap.NewNop()}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
		expectedLoc    string
	}{
		{"валидный ID", "/abc123", http.StatusTemporaryRedirect, "", "https://example.com"},
		{"несуществующий ID", "/nope", http.StatusNotFound, http.StatusText(http.StatusNotFound), ""},
		{"пустой ID", "/", http.StatusBadRequest, http.StatusText(http.StatusBadRequest), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)
			r.GET("/:id", h.GetHandler)
			r.GET("/", h.GetHandler)

			c.Request = httptest.NewRequest(http.MethodGet, tt.url, nil)
			r.ServeHTTP(rec, c.Request)

			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if tt.expectedLoc != "" {
				loc := res.Header.Get("Location")
				if loc != tt.expectedLoc {
					t.Errorf("expected Location %q, got %q", tt.expectedLoc, loc)
				}
			}
			if tt.expectedBody != "" {
				body, _ := io.ReadAll(res.Body)
				if !strings.Contains(string(body), tt.expectedBody) {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedPrefix string
	}{
		{"валидный JSON", `{"url":"https://example.com"}`, http.StatusCreated, "http://localhost:8080/"},
		{"пустой JSON", `{}`, http.StatusBadRequest, `{"error":"Bad Request"}`},
		{"невалидный JSON", `invalid`, http.StatusBadRequest, `{"error":"Bad Request"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)
			repo := repository.NewURLRepository("")
			h := &URLHandler{Repo: repo, BaseURL: "http://localhost:8080", logger: logger}

			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)
			r.POST("/api/shorten", h.ShortenHandler)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := string(bodyBytes)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp model.ShortenResponse
				if err := json.Unmarshal(bodyBytes, &resp); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if !strings.HasPrefix(resp.Result, tt.expectedPrefix) {
					t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, resp.Result)
				}
			} else if !strings.HasPrefix(bodyStr, tt.expectedPrefix) {
				t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, bodyStr)
			}
		})
	}
}

type mockRepo struct {
	err error
}

func (m *mockRepo) Ping(ctx context.Context) error {
	return m.err
}

func (m *mockRepo) Create(url string) string     { return "id123" }
func (m *mockRepo) CreateWithID(id, url string)  {}
func (m *mockRepo) Get(id string) (string, bool) { return "", false }

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name           string
		repo           repository.Repository
		expectedStatus int
		expectedBody   string
	}{
		{"DB не настроена", nil, http.StatusInternalServerError, "database not configured"},
		{"DB недоступна", &mockRepo{err: errors.New("ping failed")}, http.StatusInternalServerError, "database unreachable"},
		{"Успешный ping", &mockRepo{err: nil}, http.StatusOK, "pong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			h := &URLHandler{
				Repo:    tt.repo,
				BaseURL: "http://localhost:8080",
				logger:  zap.NewNop(),
			}

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			c.Request = req

			h.PingHandler(c)

			res := rec.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if string(body) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
			}
		})
	}
}

func TestShortenBatchHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedCount  int
		expectedPrefix string
	}{
		{
			name:           "успешное создание коротких ссылок",
			body:           `[{"correlation_id":"1","original_url":"https://example.com"},{"correlation_id":"2","original_url":"https://golang.org"}]`,
			expectedStatus: http.StatusCreated,
			expectedCount:  2,
			expectedPrefix: "http://localhost:8080/",
		},
		{
			name:           "пустой массив",
			body:           `[]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "невалидный JSON",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)

			repo := repository.NewURLRepository("")
			h := &URLHandler{Repo: repo, BaseURL: "http://localhost:8080", logger: logger}

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.body))

			h.ShortenBatchHandler(c)

			res := rec.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if res.StatusCode == http.StatusCreated {
				var resp []struct {
					CorrelationID string `json:"correlation_id"`
					ShortURL      string `json:"short_url"`
				}
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if len(resp) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(resp))
				}

				for _, item := range resp {
					if !strings.HasPrefix(item.ShortURL, tt.expectedPrefix) {
						t.Errorf("expected short URL to start with %q, got %q", tt.expectedPrefix, item.ShortURL)
					}
					if item.CorrelationID == "" {
						t.Errorf("correlation_id should not be empty")
					}
				}
			}
		})
	}
}
