package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

func TestNewURLHandler(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "Успешное создание хэндлера с базовым URL",
			baseURL: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewURLRepository("test_data.json")
			got := NewURLHandler(tt.baseURL, repo)

			if got == nil {
				t.Fatal("NewURLHandler() вернул nil, ожидался валидный объект")
			}

			if got.repo == nil {
				t.Error("repo должен быть инициализирован, но равен nil")
			}

			if got.repo != repo {
				t.Error("репозиторий в хендлере отличается от переданного")
			}

			if got.BaseURL != tt.baseURL {
				t.Errorf("BaseURL не совпадает: got %v, want %v", got.BaseURL, tt.baseURL)
			}
		})
	}
}

func TestURLHandler_GetHandler(t *testing.T) {
	tests := []struct {
		name             string
		setupRepo        func() *repository.URLRepository
		url              string
		expectedStatus   int
		expectedBody     string
		expectedLocation string
	}{
		{
			name: "Существующий короткий URL",
			setupRepo: func() *repository.URLRepository {
				r := repository.NewURLRepository()
				r.CreateWithID("id", "https://example.com")
				return r
			},
			url:              "/id",
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name: "Несуществующий короткий URL",
			setupRepo: func() *repository.URLRepository {
				return repository.NewURLRepository()
			},
			url:            "/id",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found",
		},
		{
			name: "Пустой id в URL",
			setupRepo: func() *repository.URLRepository {
				return repository.NewURLRepository()
			},
			url:            "/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)

			h := NewURLHandler("http://localhost:8080", tt.setupRepo())

			r.GET("/:id", h.GetHandler)
			r.GET("/", h.GetHandler)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			c.Request = req

			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedLocation != "" {
				loc := res.Header.Get("Location")
				if loc != tt.expectedLocation {
					t.Errorf("expected Location header %q, got %q", tt.expectedLocation, loc)
				}
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(res.Body)
				if !strings.Contains(string(body), tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestURLHandler_PostHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "действительный запрос POST",
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "пустое тело",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)

			repo := repository.NewURLRepository()
			h := NewURLHandler("http://localhost:8080", repo)

			r.POST("/", h.PostHandler)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			c.Request = req

			r.ServeHTTP(rec, req)

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

func TestURLHandler_ShortenHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedPrefix string
	}{
		{
			name:           "действительный JSON-запрос",
			body:           `{"url":"https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedPrefix: `http://localhost:8080/`,
		},
		{
			name:           "пустой JSON",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedPrefix: `{"error":"invalid request"}`,
		},
		{
			name:           "невалидный JSON",
			body:           `invalid`,
			expectedStatus: http.StatusBadRequest,
			expectedPrefix: `{"error":"invalid request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, r := gin.CreateTestContext(rec)

			repo := repository.NewURLRepository()
			h := NewURLHandler("http://localhost:8080", repo)

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
					t.Fatalf("failed to unmarshal JSON: %v", err)
				}
				if !strings.HasPrefix(resp.Result, tt.expectedPrefix) {
					t.Errorf("expected result to start with %q, got %q", tt.expectedPrefix, resp.Result)
				}

				if ct := res.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}
			} else {
				if !strings.HasPrefix(bodyStr, tt.expectedPrefix) {
					t.Errorf("expected body to start with %q, got %q", tt.expectedPrefix, bodyStr)
				}
			}
		})
	}
}
