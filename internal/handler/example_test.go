package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ExampleURLHandler_PostHandler демонстрирует создание короткого URL через POST /
func ExampleURLHandler_PostHandler() {
	gin.SetMode(gin.TestMode)

	repo := repository.NewURLRepository("")
	publisher := audit.NewPublisher()
	svc := &service.URLService{Repo: repo, BaseURL: "http://localhost:8080", Publisher: publisher, Logger: zap.NewNop()}
	h, _ := NewURLHandler("http://localhost:8080", repo, nil, svc, "")

	r := gin.New()
	r.POST("/", h.PostHandler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Output:
}

// ExampleURLHandler_ShortenHandler демонстрирует POST /api/shorten
func ExampleURLHandler_ShortenHandler() {
	gin.SetMode(gin.TestMode)

	repo := repository.NewURLRepository("")
	publisher := audit.NewPublisher()
	svc := &service.URLService{Repo: repo, BaseURL: "http://localhost:8080", Publisher: publisher, Logger: zap.NewNop()}
	h, _ := NewURLHandler("http://localhost:8080", repo, nil, svc, "")

	r := gin.New()
	r.POST("/api/shorten", h.ShortenHandler)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Output:
}

// ExampleURLHandler_PingHandler демонстрирует GET /ping
func ExampleURLHandler_PingHandler() {
	gin.SetMode(gin.TestMode)

	repo := repository.NewURLRepository("")
	publisher := audit.NewPublisher()
	svc := &service.URLService{Repo: repo, BaseURL: "http://localhost:8080", Publisher: publisher, Logger: zap.NewNop()}
	h, _ := NewURLHandler("http://localhost:8080", repo, nil, svc, "")

	r := gin.New()
	r.GET("/ping", h.PingHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Output:
}
