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
		expectedBody   string
	}{
		{
			name:           "успешное создание короткой ссылки",
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "пустое тело запроса",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   http.StatusText(http.StatusBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			var logBuf bytes.Buffer
			logger := newTestLogger(&logBuf)

			repo := repository.NewURLRepository("")
			h := &URLHandler{
				repo:    repo,
				BaseURL: "http://localhost:8080",
				logger:  logger,
			}

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			ctx.Request = req

			h.PostHandler(ctx)

			res := rec.Result()
			defer res.Body.Close()

			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := string(bodyBytes)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if !strings.HasPrefix(bodyStr, tt.expectedBody) {
				t.Errorf("expected body to start with %q, got %q", tt.expectedBody, bodyStr)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := repository.NewURLRepository("")
	repo.CreateWithID("abc123", "https://example.com")

	h := &URLHandler{
		repo:    repo,
		BaseURL: "http://localhost:8080",
		logger:  zap.NewNop(),
	}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
		expectedLoc    string
	}{
		{
			name:           "валидный короткий ID",
			url:            "/abc123",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedLoc:    "https://example.com",
		},
		{
			name:           "несуществующий ID",
			url:            "/nope",
			expectedStatus: http.StatusNotFound,
			expectedBody:   http.StatusText(http.StatusNotFound),
		},
		{
			name:           "пустой ID",
			url:            "/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   http.StatusText(http.StatusBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)

			r.GET("/:id", h.GetHandler)
			r.GET("/", h.GetHandler)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			c.Request = req
			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, res.StatusCode)
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
		{
			name:           "валидный JSON",
			body:           `{"url":"https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedPrefix: "http://localhost:8080/",
		},
		{
			name:           "пустой JSON",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedPrefix: `{"error":"` + http.StatusText(http.StatusBadRequest) + `"}`,
		},
		{
			name:           "невалидный JSON",
			body:           `invalid`,
			expectedStatus: http.StatusBadRequest,
			expectedPrefix: `{"error":"` + http.StatusText(http.StatusBadRequest) + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			var logBuf bytes.Buffer
			logger := newTestLogger(&logBuf)

			repo := repository.NewURLRepository("")
			h := &URLHandler{
				repo:    repo,
				BaseURL: "http://localhost:8080",
				logger:  logger,
			}

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
			} else {
				if !strings.HasPrefix(bodyStr, tt.expectedPrefix) {
					t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, bodyStr)
				}
			}

			if tt.expectedStatus >= 500 && !strings.Contains(logBuf.String(), "Error") {
				t.Error("expected error to be logged, but log is empty")
			}
		})
	}
}

type Pingable interface {
	Ping(context.Context) error
}

type mockDB struct {
	err error
}

func (m *mockDB) Ping(ctx context.Context) error {
	_ = ctx
	return m.err
}

func TestPingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		db             Pingable
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "DB не настроена",
			db:             nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database not configured",
		},
		{
			name:           "DB недоступна",
			db:             &mockDB{err: errors.New("ping failed")},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database unreachable",
		},
		{
			name:           "Успешный ping",
			db:             &mockDB{err: nil},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			h := &URLHandler{
				DB:     tt.db,
				logger: zap.NewNop(),
			}

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			c.Request = req

			h.PingHandler(c)

			res := rec.Result()
			defer res.Body.Close()

			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := string(bodyBytes)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if bodyStr != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, bodyStr)
			}
		})
	}
}
