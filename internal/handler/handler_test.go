package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger(buf *bytes.Buffer) *zap.Logger {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)
	return zap.New(core)
}

type mockRepo struct {
	err   error
	store map[string]map[string]string // userID -> id -> URL
}

func newMockRepo(err error) *mockRepo {
	return &mockRepo{
		err:   err,
		store: make(map[string]map[string]string),
	}
}

func (m *mockRepo) Ping(ctx context.Context) error {
	return m.err
}

func (m *mockRepo) Create(originalURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	id := "mockID"
	if m.store["__global__"] == nil {
		m.store["__global__"] = make(map[string]string)
	}
	m.store["__global__"][id] = originalURL
	return id, nil
}

func (m *mockRepo) CreateWithID(id, url string) {
	if m.store["__global__"] == nil {
		m.store["__global__"] = make(map[string]string)
	}
	m.store["__global__"][id] = url
}

func (m *mockRepo) Get(id string) (string, bool) {
	for _, urls := range m.store {
		if url, ok := urls[id]; ok {
			return url, true
		}
	}
	return "", false
}

func (m *mockRepo) CreateForUser(userID, originalURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.store[userID] == nil {
		m.store[userID] = make(map[string]string)
	}
	id := "userID" + strconv.Itoa(len(m.store[userID])+1)
	m.store[userID][id] = originalURL
	return id, nil
}

func (m *mockRepo) GetAllForUser(userID string) map[string]string {
	if m.store[userID] == nil {
		return nil
	}
	return m.store[userID]
}

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedPrefix string
	}{
		{
			name:           "успешное создание короткой ссылки",
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedPrefix: "http://localhost:8080/",
		},
		{
			name:           "пустое тело запроса",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedPrefix: http.StatusText(http.StatusBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)
			repo := newMockRepo(nil)

			h := &URLHandler{
				Repo:    repo,
				BaseURL: "http://localhost:8080",
				logger:  logger,
			}

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))

			h.PostHandler(c)

			res := rec.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := string(bodyBytes)

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}
			if tt.expectedStatus == http.StatusCreated {
				if !strings.HasPrefix(bodyStr, tt.expectedPrefix) {
					t.Errorf("expected prefix %q, got body %q", tt.expectedPrefix, bodyStr)
				}
			} else {
				if strings.TrimSpace(bodyStr) != tt.expectedPrefix {
					t.Errorf("expected body %q, got %q", tt.expectedPrefix, bodyStr)
				}
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
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
		expectedJSON   string
		expectedPrefix string
	}{
		{"валидный JSON", `{"url":"https://example.com"}`, http.StatusCreated, "", "http://localhost:8080/"},
		{"пустой JSON", `{}`, http.StatusBadRequest, `{"error":"Bad Request"}`, ""},
		{"невалидный JSON", `invalid`, http.StatusBadRequest, `{"error":"Bad Request"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)
			repo := newMockRepo(nil)

			h := &URLHandler{Repo: repo, BaseURL: "http://localhost:8080", logger: logger}

			rec := httptest.NewRecorder()
			router := gin.Default()
			router.POST("/api/shorten", h.ShortenHandler)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := strings.TrimSpace(string(bodyBytes))

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp model.ShortenResponse
				if err := json.Unmarshal(bodyBytes, &resp); err != nil {
					t.Fatalf("failed to unmarshal JSON: %v", err)
				}
				if !strings.HasPrefix(resp.Result, tt.expectedPrefix) {
					t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, resp.Result)
				}
				return
			}

			if tt.expectedJSON != "" && bodyStr != tt.expectedJSON {
				t.Errorf("expected JSON %q, got %q", tt.expectedJSON, bodyStr)
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name           string
		repo           repository.Repository
		expectedStatus int
		expectedBody   string
	}{
		{"DB не настроена", nil, http.StatusInternalServerError, "database not configured"},
		{"DB недоступна", newMockRepo(errors.New("ping failed")), http.StatusInternalServerError, "database unreachable"},
		{"Успешный ping", newMockRepo(nil), http.StatusOK, "pong"},
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
		expectedJSON   string
		expectedPrefix string
	}{
		{
			name:           "успешное создание коротких ссылок",
			body:           `[{"correlation_id":"1","original_url":"https://example.com"}, {"correlation_id":"2","original_url":"https://golang.org"}]`,
			expectedStatus: http.StatusCreated,
			expectedCount:  2,
			expectedPrefix: "http://localhost:8080/",
		},
		{"пустой массив", `[]`, http.StatusBadRequest, 0, `{"error":"invalid request body"}`, ""},
		{"невалидный JSON", `invalid json`, http.StatusBadRequest, 0, `{"error":"invalid request body"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			logBuf := &bytes.Buffer{}
			logger := newTestLogger(logBuf)
			repo := newMockRepo(nil)

			h := &URLHandler{
				Repo:    repo,
				BaseURL: "http://localhost:8080",
				logger:  logger,
			}

			router := gin.Default()
			router.POST("/api/shorten/batch", h.ShortenBatchHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			bodyBytes, _ := io.ReadAll(res.Body)
			bodyStr := strings.TrimSpace(string(bodyBytes))

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp []struct {
					CorrelationID string `json:"correlation_id"`
					ShortURL      string `json:"short_url"`
				}
				if err := json.Unmarshal(bodyBytes, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(resp))
				}
				for _, item := range resp {
					if item.CorrelationID == "" {
						t.Errorf("correlation_id should not be empty")
					}
					if !strings.HasPrefix(item.ShortURL, tt.expectedPrefix) {
						t.Errorf("expected prefix %q, got %q", tt.expectedPrefix, item.ShortURL)
					}
				}
				return
			}

			if tt.expectedJSON != "" && bodyStr != tt.expectedJSON {
				t.Errorf("expected JSON %q, got %q", tt.expectedJSON, bodyStr)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo(nil)
	h := &URLHandler{
		Repo:    repo,
		BaseURL: "http://localhost:8080",
		logger:  zap.NewNop(),
	}

	cookie := service.GenerateUserCookie()
	userID, _ := service.ValidateUserCookie(&http.Request{Header: http.Header{"Cookie": []string{cookie.String()}}})

	repo.CreateForUser(userID, "https://example.com")
	repo.CreateForUser(userID, "https://golang.org")

	tests := []struct {
		name           string
		cookie         *http.Cookie
		expectedStatus int
		expectedCount  int
	}{
		{"валидная кука с URL", cookie, http.StatusOK, 2},
		{"валидная кука без URL", service.GenerateUserCookie(), http.StatusNoContent, 0},
		{"отсутствие куки", nil, http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)
			r.GET("/api/user/urls", h.GetUserURLs)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			c.Request = req

			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp []map[string]string
				body, _ := io.ReadAll(res.Body)
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(resp))
				}
				for _, item := range resp {
					if item["short_url"] == "" || item["original_url"] == "" {
						t.Errorf("short_url or original_url is empty")
					}
				}
			}
		})
	}
}
