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
	Repo    repository.Repository
	BaseURL string
	logger  *zap.Logger
}

func NewURLHandler(baseURL string, repo repository.Repository) *URLHandler {
	logger, _ := zap.NewProduction()
	return &URLHandler{
		Repo:    repo,
		BaseURL: baseURL,
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
	id := h.Repo.Create(originalURL)

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

	original, ok := h.Repo.Get(id)
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

	id := h.Repo.Create(req.URL)

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
	if h.Repo == nil {
		c.String(http.StatusInternalServerError, "database not configured")
		return
	}

	if err := h.Repo.Ping(c); err != nil {
		c.String(http.StatusInternalServerError, "database unreachable")
		return
	}
	c.String(http.StatusOK, "pong")
}

// POST /api/shorten/batch
func (h *URLHandler) ShortenBatchHandler(c *gin.Context) {
	var req []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || len(req) == 0 {
		h.logger.Error("failed to decode batch request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	for i, item := range req {
		if item.OriginalURL == "" || item.CorrelationID == "" {
			h.logger.Warn("empty URL or correlation_id in batch", zap.Int("index", i))
			c.JSON(http.StatusBadRequest, gin.H{"error": "each item must have correlation_id and original_url"})
			return
		}
	}

	type responseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	urls := make([]string, len(req))
	for i, item := range req {
		urls[i] = item.OriginalURL
	}

	var ids []string
	if batchRepo, ok := h.Repo.(*repository.URLRepository); ok {
		ids = batchRepo.BatchCreate(urls)
	} else {
		ids = make([]string, len(urls))
		for i, u := range urls {
			ids[i] = h.Repo.Create(u)
		}
	}

	responses := make([]responseItem, len(req))
	for i, item := range req {
		shortURL, _ := url.JoinPath(h.BaseURL, ids[i])
		responses[i] = responseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		}
	}

	c.JSON(http.StatusCreated, responses)
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}
