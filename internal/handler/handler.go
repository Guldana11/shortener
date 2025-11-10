package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type URLHandler struct {
	repo    *repository.URLRepository
	BaseURL string
	DB      Pinger
	logger  *zap.Logger
}

func NewURLHandler(baseURL string, repo *repository.URLRepository, db Pinger) *URLHandler {
	logger, _ := zap.NewProduction()
	return &URLHandler{
		repo:    repo,
		BaseURL: baseURL,
		DB:      db,
		logger:  logger,
	}
}

// POST /
func (h *URLHandler) PostHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if len(body) == 0 {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	originalURL := string(body)
	id := h.repo.Create(originalURL)

	shortURL, err := url.JoinPath(h.BaseURL, id)
	if err != nil {
		h.logger.Error("failed to join URL path", zap.Error(err))
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.String(http.StatusCreated, shortURL)
}

// GET /:id
func (h *URLHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" || containsSlash(id) {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	original, ok := h.repo.Get(id)
	if !ok {
		c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, original)
}

// POST /api/shorten
func (h *URLHandler) ShortenHandler(c *gin.Context) {
	var req model.ShortenRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.URL == "" {
		h.logger.Error("failed to decode JSON or missing URL", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	id := h.repo.Create(req.URL)

	shortURL, err := url.JoinPath(h.BaseURL, id)
	if err != nil {
		h.logger.Error("failed to join URL path", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	resp := model.ShortenResponse{Result: shortURL}
	c.JSON(http.StatusCreated, resp)
}

// GET /ping
func (h *URLHandler) PingHandler(c *gin.Context) {
	if h.DB == nil {
		c.String(http.StatusInternalServerError, "database not configured")
		return
	}

	if err := h.DB.Ping(c); err != nil {
		c.String(http.StatusInternalServerError, "database unreachable")
		return
	}

	c.String(http.StatusOK, "pong")
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
